// Package protocol defines protocol kinds.
//
// Only HTTP is runnable in v1. Other kinds are reserved so imported or future
// nodes can be stored without schema churn. Send logic lives in subpackages
// (e.g. httpapi), not in UI views.
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
