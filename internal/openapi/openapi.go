// Package openapi imports OpenAPI/Swagger 2.0, 3.0.x, and 3.1.x from file or URL.
package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Source is either a local file path or an HTTP(S) URL (exactly one should be set).
type Source struct {
	FilePath string
	URL      string
}

// Operation is one path+method endpoint to import.
type Operation struct {
	Tags        []string
	Name        string
	Method      string
	Path        string
	Description string
}

// Document is a normalized import result.
type Document struct {
	Version    string
	Title      string
	Operations []Operation
}

// Import loads and auto-detects OpenAPI version from file or URL.
func Import(ctx context.Context, src Source) (*Document, error) {
	raw, err := load(ctx, src)
	if err != nil {
		return nil, err
	}
	return Parse(raw)
}

func load(ctx context.Context, src Source) ([]byte, error) {
	switch {
	case src.FilePath != "" && src.URL == "":
		return os.ReadFile(src.FilePath)
	case src.URL != "" && src.FilePath == "":
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.URL, nil)
		if err != nil {
			return nil, err
		}
		client := &http.Client{Timeout: 60 * time.Second}
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return nil, fmt.Errorf("openapi: GET %s: %s", src.URL, res.Status)
		}
		return io.ReadAll(res.Body)
	default:
		return nil, fmt.Errorf("openapi: set exactly one of FilePath or URL")
	}
}

// Parse auto-detects JSON/YAML and OpenAPI version.
func Parse(raw []byte) (*Document, error) {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		root = nil
		if err2 := yaml.Unmarshal(raw, &root); err2 != nil {
			return nil, fmt.Errorf("openapi: not JSON (%w) or YAML (%w)", err, err2)
		}
	}
	if root == nil {
		return nil, fmt.Errorf("openapi: empty document")
	}

	doc := &Document{}
	if v, ok := root["swagger"].(string); ok && strings.HasPrefix(v, "2.") {
		doc.Version = "2.0"
		return parseSwagger2(root, doc)
	}
	if v, ok := root["openapi"].(string); ok {
		switch {
		case strings.HasPrefix(v, "3.0"):
			doc.Version = "3.0"
		case strings.HasPrefix(v, "3.1"):
			doc.Version = "3.1"
		default:
			doc.Version = v
		}
		return parseOpenAPI3(root, doc)
	}
	return nil, fmt.Errorf("openapi: unrecognized document (need swagger 2.x or openapi 3.x)")
}

func parseSwagger2(root map[string]any, doc *Document) (*Document, error) {
	if info, ok := root["info"].(map[string]any); ok {
		doc.Title, _ = info["title"].(string)
	}
	paths, _ := root["paths"].(map[string]any)
	for path, item := range paths {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for method, opAny := range obj {
			m := strings.ToUpper(method)
			if !isHTTPMethod(m) {
				continue
			}
			op, _ := opAny.(map[string]any)
			doc.Operations = append(doc.Operations, opFromMap(op, m, path))
		}
	}
	return doc, nil
}

func parseOpenAPI3(root map[string]any, doc *Document) (*Document, error) {
	if info, ok := root["info"].(map[string]any); ok {
		doc.Title, _ = info["title"].(string)
	}
	paths, _ := root["paths"].(map[string]any)
	for path, item := range paths {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for method, opAny := range obj {
			m := strings.ToUpper(method)
			if !isHTTPMethod(m) {
				continue
			}
			op, _ := opAny.(map[string]any)
			doc.Operations = append(doc.Operations, opFromMap(op, m, path))
		}
	}
	return doc, nil
}

func opFromMap(op map[string]any, method, path string) Operation {
	out := Operation{Method: method, Path: path}
	if op == nil {
		out.Name = method + " " + path
		return out
	}
	if s, ok := op["operationId"].(string); ok && s != "" {
		out.Name = s
	} else if s, ok := op["summary"].(string); ok && s != "" {
		out.Name = s
	} else {
		out.Name = method + " " + path
	}
	if s, ok := op["description"].(string); ok {
		out.Description = s
	}
	if tags, ok := op["tags"].([]any); ok {
		for _, t := range tags {
			if s, ok := t.(string); ok {
				out.Tags = append(out.Tags, s)
			}
		}
	}
	return out
}

func isHTTPMethod(m string) bool {
	switch m {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}
