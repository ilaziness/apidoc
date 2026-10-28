package openapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseOpenAPI3(t *testing.T) {
	raw := []byte(`{
  "openapi": "3.0.3",
  "info": {"title": "Sample"},
  "paths": {
    "/pets": {
      "get": {"operationId": "listPets", "tags": ["pets"], "summary": "List"}
    },
    "/pets/{id}": {
      "get": {"summary": "Get pet"}
    }
  }
}`)
	doc, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != "3.0" || doc.Title != "Sample" {
		t.Fatalf("%+v", doc)
	}
	if len(doc.Operations) != 2 {
		t.Fatalf("ops=%d", len(doc.Operations))
	}
}

func TestParseSwagger2YAML(t *testing.T) {
	raw := []byte(`
swagger: "2.0"
info:
  title: Old
paths:
  /health:
    get:
      operationId: health
`)
	doc, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != "2.0" || len(doc.Operations) != 1 {
		t.Fatalf("%+v", doc)
	}
}

func TestImportFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "api.json")
	content := `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/x":{"get":{"operationId":"x"}}}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := Import(t.Context(), Source{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != "3.1" || len(doc.Operations) != 1 {
		t.Fatalf("%+v", doc)
	}
}
