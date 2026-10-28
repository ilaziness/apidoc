package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ilaziness/apidoc/internal/config"
)

func TestSendGETWithVarsAndAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/42" {
			t.Errorf("path=%s", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "x" {
			t.Errorf("query=%s", r.URL.RawQuery)
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("auth=%s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Trace") != "1" {
			t.Errorf("header missing")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer srv.Close()

	spec := DefaultSpec()
	spec.Method = "GET"
	spec.URL = "{{base}}/users/{id}"
	spec.PathParams = []KV{{Key: "id", Value: "42", Enabled: true}}
	spec.Query = []KV{
		{Key: "q", Value: "x", Enabled: true},
		{Key: "skip", Value: "1", Enabled: false},
	}
	spec.Headers = []KV{{Key: "X-Trace", Value: "1", Enabled: true}}
	spec.Auth = &Auth{Kind: "bearer", Token: "{{token}}"}

	h := New()
	res, err := h.Send(context.Background(), spec, SendOptions{
		Env:   map[string]string{"base": srv.URL, "token": "secret"},
		Proxy: config.ProxyConfig{Mode: config.ProxyOff},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("status=%d", res.StatusCode)
	}
	if !strings.Contains(string(res.Body), `"ok"`) {
		t.Fatalf("body=%s", res.Body)
	}
	if res.Duration <= 0 {
		t.Fatal("duration")
	}
}

func TestSendCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		_, _ = io.WriteString(w, "late")
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	spec := DefaultSpec()
	spec.URL = srv.URL
	spec.TimeoutSec = 5
	_, err := New().Send(ctx, spec, SendOptions{Proxy: config.ProxyConfig{Mode: config.ProxyOff}})
	if err == nil {
		t.Fatal("expected cancel/timeout error")
	}
}

func TestMaybePrettyJSON(t *testing.T) {
	out := maybePrettyJSON([]byte(`{"a":1}`))
	if !strings.Contains(string(out), "\n") {
		t.Fatalf("expected pretty: %s", out)
	}
}
