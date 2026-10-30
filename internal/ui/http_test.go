package ui

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ilaziness/apidoc/internal/protocol/httpapi"
)

func TestFormatHTTPResponse(t *testing.T) {
	got := formatHTTPResponse(&httpapi.Response{
		Status:    "200 OK",
		Duration:  time.Millisecond,
		Headers:   http.Header{"X-A": []string{"1"}, "Content-Type": []string{"application/json"}},
		Body:      []byte(`{"ok":true}`),
		Truncated: true,
	})
	if !strings.Contains(got, "200 OK") {
		t.Fatalf("status missing: %s", got)
	}
	if !strings.Contains(got, "Content-Type: application/json") {
		t.Fatalf("header missing: %s", got)
	}
	if !strings.Contains(got, "[truncated]") {
		t.Fatalf("truncated missing: %s", got)
	}
	if !strings.Contains(got, `{"ok":true}`) {
		t.Fatalf("body missing: %s", got)
	}
}
