// Package protocol defines protocol kinds and a handler registry.
//
// Only HTTP is runnable in v1. Other kinds are reserved so imported or future
// nodes can be stored without schema churn. Send logic for each protocol lives
// in a dedicated subpackage (e.g. httpapi), not in UI views.
package protocol

// Kind identifies the wire protocol of an API node (persisted on tree nodes).
type Kind string

const (
	HTTP      Kind = "http"
	GRPC      Kind = "grpc"
	TCP       Kind = "tcp"
	UDP       Kind = "udp"
	WebSocket Kind = "websocket"
)

// Handler is a registered protocol implementation.
// Protocol-specific Send methods stay on concrete types (see httpapi).
type Handler interface {
	Kind() Kind
}

// Registry maps kinds to handlers for "is this protocol supported?" checks.
type Registry struct {
	handlers map[Kind]Handler
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{handlers: make(map[Kind]Handler)}
}

// Register adds a handler. Re-registering the same Kind replaces the previous one.
func (r *Registry) Register(h Handler) {
	r.handlers[h.Kind()] = h
}

// Get returns a handler or false if the kind is unsupported in this build.
func (r *Registry) Get(k Kind) (Handler, bool) {
	h, ok := r.handlers[k]
	return h, ok
}
