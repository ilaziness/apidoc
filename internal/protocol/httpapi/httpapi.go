// Package httpapi implements the HTTP protocol handler.
package httpapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ilaziness/apidoc/internal/config"
	"github.com/ilaziness/apidoc/internal/protocol/vars"
)

const maxResponseBody = 10 * 1024 * 1024 // 10MB (PRD)

// Handler sends HTTP requests.
type Handler struct{}

// New returns an HTTP handler.
func New() *Handler { return &Handler{} }

// SendOptions controls transport outside the saved Spec (global proxy).
type SendOptions struct {
	Proxy          config.ProxyConfig
	Env            map[string]string
	CollectionAuth *Auth // used when Spec.Auth is nil
}

// Response is one HTTP send result for the UI.
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       []byte
	Truncated  bool
	Duration   time.Duration
}

// Send executes the request.
func (h *Handler) Send(ctx context.Context, spec Spec, opt SendOptions) (*Response, error) {
	method := strings.ToUpper(strings.TrimSpace(spec.Method))
	if !allowedMethod(method) {
		return nil, fmt.Errorf("httpapi: unsupported method %q", spec.Method)
	}

	env := opt.Env
	if env == nil {
		env = map[string]string{}
	}

	rawURL := vars.Expand(spec.URL, env)
	rawURL = applyPathParams(rawURL, spec.PathParams, env)
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("httpapi: url: %w", err)
	}
	q := u.Query()
	for _, kv := range spec.Query {
		if !kv.Enabled || kv.Key == "" {
			continue
		}
		q.Add(vars.Expand(kv.Key, env), vars.Expand(kv.Value, env))
	}

	auth := resolveAuth(spec.Auth, opt.CollectionAuth)
	if auth != nil && auth.Kind == "apikey" && strings.EqualFold(auth.In, "query") {
		q.Set(vars.Expand(auth.Key, env), vars.Expand(auth.Value, env))
	}
	u.RawQuery = q.Encode()

	bodyReader, contentType, err := buildBody(spec.Body, env)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}
	for _, kv := range spec.Headers {
		if !kv.Enabled || kv.Key == "" {
			continue
		}
		req.Header.Add(vars.Expand(kv.Key, env), vars.Expand(kv.Value, env))
	}
	if contentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}
	applyAuthHeaders(req, auth, env)

	timeout := time.Duration(spec.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	client, err := newClient(timeout, spec.FollowRedirects, spec.InsecureSkipVerify, opt.Proxy)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	limited := io.LimitReader(res.Body, maxResponseBody+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	truncated := false
	if len(data) > maxResponseBody {
		data = data[:maxResponseBody]
		truncated = true
	}
	data = maybePrettyJSON(data)

	return &Response{
		StatusCode: res.StatusCode,
		Status:     res.Status,
		Headers:    res.Header.Clone(),
		Body:       data,
		Truncated:  truncated,
		Duration:   time.Since(start),
	}, nil
}

func allowedMethod(m string) bool {
	switch m {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}

func applyPathParams(raw string, params []KV, env map[string]string) string {
	for _, kv := range params {
		if !kv.Enabled || kv.Key == "" {
			continue
		}
		token := "{" + kv.Key + "}"
		raw = strings.ReplaceAll(raw, token, url.PathEscape(vars.Expand(kv.Value, env)))
	}
	return raw
}

func resolveAuth(node, collection *Auth) *Auth {
	if node != nil {
		return node
	}
	return collection
}

func applyAuthHeaders(req *http.Request, auth *Auth, env map[string]string) {
	if auth == nil {
		return
	}
	switch auth.Kind {
	case "", "none":
		return
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+vars.Expand(auth.Token, env))
	case "basic":
		req.SetBasicAuth(vars.Expand(auth.User, env), vars.Expand(auth.Pass, env))
	case "apikey":
		if strings.EqualFold(auth.In, "query") {
			return
		}
		req.Header.Set(vars.Expand(auth.Key, env), vars.Expand(auth.Value, env))
	}
}

func buildBody(body Body, env map[string]string) (io.Reader, string, error) {
	switch body.Kind {
	case "", "none":
		return nil, "", nil
	case "json":
		return strings.NewReader(vars.Expand(body.JSON, env)), "application/json", nil
	case "raw":
		ct := body.ContentType
		if ct == "" {
			ct = "text/plain"
		}
		return strings.NewReader(vars.Expand(body.Raw, env)), ct, nil
	case "form", "x-www-form-urlencoded":
		form := url.Values{}
		for _, kv := range body.Form {
			if !kv.Enabled || kv.Key == "" {
				continue
			}
			form.Add(vars.Expand(kv.Key, env), vars.Expand(kv.Value, env))
		}
		return strings.NewReader(form.Encode()), "application/x-www-form-urlencoded", nil
	case "multipart":
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		for _, kv := range body.Multipart {
			if !kv.Enabled || kv.Key == "" {
				continue
			}
			key := vars.Expand(kv.Key, env)
			val := vars.Expand(kv.Value, env)
			if path, ok := strings.CutPrefix(val, "file:"); ok {
				f, err := os.Open(path)
				if err != nil {
					return nil, "", err
				}
				part, err := w.CreateFormFile(key, path)
				if err != nil {
					_ = f.Close()
					return nil, "", err
				}
				_, err = io.Copy(part, f)
				_ = f.Close()
				if err != nil {
					return nil, "", err
				}
				continue
			}
			if err := w.WriteField(key, val); err != nil {
				return nil, "", err
			}
		}
		if err := w.Close(); err != nil {
			return nil, "", err
		}
		return &buf, w.FormDataContentType(), nil
	case "binary":
		if body.BinaryPath == "" {
			return nil, "", fmt.Errorf("httpapi: binary body missing path")
		}
		f, err := os.Open(body.BinaryPath)
		if err != nil {
			return nil, "", err
		}
		data, err := io.ReadAll(f)
		_ = f.Close()
		if err != nil {
			return nil, "", err
		}
		ct := body.ContentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		return bytes.NewReader(data), ct, nil
	default:
		return nil, "", fmt.Errorf("httpapi: unknown body kind %q", body.Kind)
	}
}

func newClient(timeout time.Duration, followRedirects, insecure bool, proxy config.ProxyConfig) (*http.Client, error) {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	switch proxy.Mode {
	case config.ProxyOff:
		transport.Proxy = nil
	case config.ProxyManual:
		if proxy.Manual == "" {
			return nil, fmt.Errorf("httpapi: manual proxy URL empty")
		}
		u, err := url.Parse(proxy.Manual)
		if err != nil {
			return nil, fmt.Errorf("httpapi: proxy url: %w", err)
		}
		transport.Proxy = http.ProxyURL(u)
	case config.ProxySystem, "":
		transport.Proxy = http.ProxyFromEnvironment
	default:
		return nil, fmt.Errorf("httpapi: unknown proxy mode %q", proxy.Mode)
	}
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // PRD allows skip verify
	}
	client := &http.Client{Transport: transport, Timeout: timeout}
	if !followRedirects {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return client, nil
}

func maybePrettyJSON(b []byte) []byte {
	trim := bytes.TrimSpace(b)
	if len(trim) == 0 || (trim[0] != '{' && trim[0] != '[') {
		return b
	}
	var v any
	if err := json.Unmarshal(trim, &v); err != nil {
		return b
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return b
	}
	return out
}

// EncodeSpec marshals Spec to JSON.
func EncodeSpec(s Spec) (string, error) {
	b, err := json.Marshal(s)
	return string(b), err
}

// DecodeSpec unmarshals Spec from JSON.
func DecodeSpec(s string) (Spec, error) {
	if s == "" {
		return DefaultSpec(), nil
	}
	var spec Spec
	if err := json.Unmarshal([]byte(s), &spec); err != nil {
		return Spec{}, err
	}
	if spec.TimeoutSec == 0 {
		spec.TimeoutSec = 60
	}
	return spec, nil
}
