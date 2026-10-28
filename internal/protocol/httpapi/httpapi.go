// Package httpapi implements the HTTP protocol handler.
//
// Named httpapi (not http) so imports do not collide with net/http.
// All net/http client usage belongs here, not in UI or service packages.
package httpapi

import (
	"context"
	"fmt"

	"github.com/ilaziness/apidoc/internal/protocol"
)

// Handler sends HTTP requests.
type Handler struct{}

// New returns an HTTP handler.
func New() *Handler {
	return &Handler{}
}

// Kind implements protocol.Handler.
func (h *Handler) Kind() protocol.Kind {
	return protocol.HTTP
}

// Request is one HTTP send input (expanded as the editor model lands).
type Request struct {
	Method  string
	URL     string
	Headers []Header
	Body    []byte
	// Timeout, FollowRedirects, InsecureSkipVerify come from the saved node + global proxy.
}

// Header is a single request header line.
type Header struct {
	Key     string
	Value   string
	Enabled bool
}

// Response is one HTTP send result for the UI.
type Response struct {
	StatusCode int
	Headers    []Header
	Body       []byte
	// Truncated is set when Body exceeds the PRD 10MB display limit.
	Truncated bool
}

// Send executes the request. Not yet implemented.
func (h *Handler) Send(ctx context.Context, req *Request) (*Response, error) {
	_ = ctx
	_ = req
	return nil, fmt.Errorf("httpapi: send not implemented")
}
