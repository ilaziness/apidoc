// Package openapi imports OpenAPI/Swagger 2.0, 3.0.x, and 3.1.x from file or URL.
// Export is out of scope for v1.
package openapi

import (
	"context"
	"errors"
)

// ErrNotImplemented is returned until the importer is built.
var ErrNotImplemented = errors.New("openapi: import not implemented")

// Source is either a local file path or an HTTP(S) URL (exactly one should be set).
type Source struct {
	FilePath string
	URL      string
}

// Result is a parsed document ready to map into collection tree nodes.
type Result struct {
	Version string
	// Document holds the parsed model once implemented.
	Document any
}

// Import loads and auto-detects OpenAPI version from file or URL.
func Import(ctx context.Context, src Source) (*Result, error) {
	_ = ctx
	_ = src
	return nil, ErrNotImplemented
}
