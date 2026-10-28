// Package service holds application use cases: collections, tree, environments,
// OpenAPI import, and send. The UI must call this layer; it must not run SQL
// or protocol clients directly.
package service

import (
	"github.com/ilaziness/apidoc/internal/protocol"
	"github.com/ilaziness/apidoc/internal/store"
)

// Deps are collaborators constructed by main and injected once.
type Deps struct {
	Store     *store.DB
	Protocols *protocol.Registry
}

// Service is the application facade used by the UI.
type Service struct {
	store     *store.DB
	protocols *protocol.Registry
}

// New builds a Service from deps.
func New(deps Deps) *Service {
	return &Service{
		store:     deps.Store,
		protocols: deps.Protocols,
	}
}
