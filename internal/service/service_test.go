package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ilaziness/apidoc/internal/config"
	"github.com/ilaziness/apidoc/internal/openapi"
	"github.com/ilaziness/apidoc/internal/protocol/httpapi"
	"github.com/ilaziness/apidoc/internal/store"
)

func TestCollectionTreeSendImport(t *testing.T) {
	dir := t.TempDir()
	db, err := store.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	cfg := &config.Config{
		DataDir:  dir,
		Language: config.LangEN,
		Theme:    config.ThemeLight,
		Proxy:    config.ProxyConfig{Mode: config.ProxyOff},
	}
	svc := New(Deps{Store: db, Config: cfg, HTTP: httpapi.New()})
	ctx := context.Background()

	col, err := svc.CreateCollection(ctx, "Demo")
	if err != nil {
		t.Fatal(err)
	}
	env, err := svc.CreateEnvironment(ctx, col.ID, "dev", map[string]string{"base": "http://invalid"})
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"hello": "world"})
	}))
	defer srv.Close()
	if err := svc.UpdateEnvironment(ctx, env, map[string]string{"base": srv.URL}); err != nil {
		t.Fatal(err)
	}

	spec := httpapi.DefaultSpec()
	spec.URL = "{{base}}/ping"
	node, err := svc.CreateHTTP(ctx, col.ID, nil, "ping", spec)
	if err != nil {
		t.Fatal(err)
	}
	res, err := svc.SendHTTP(ctx, node, env.ID)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("status=%d body=%s", res.StatusCode, res.Body)
	}

	api := filepath.Join(dir, "api.json")
	content := `{"openapi":"3.0.0","info":{"title":"Imp"},"paths":{"/a":{"get":{"operationId":"getA","tags":["t"]}}}}`
	if err := os.WriteFile(api, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	imported, err := svc.ImportOpenAPI(ctx, ImportOpenAPIOptions{
		Source: openapi.Source{FilePath: api},
	})
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := svc.ListNodes(ctx, imported.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) < 2 {
		t.Fatalf("imported nodes=%d", len(nodes))
	}
}
