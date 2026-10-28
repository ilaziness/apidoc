// Package ui is the go-gui presentation layer.
//
// It must not issue SQL or call net/http directly; use internal/service.
package ui

import (
	"fmt"

	"github.com/ilaziness/apidoc/internal/config"
	"github.com/ilaziness/apidoc/internal/service"
)

// Run starts the desktop window. go-gui wiring lands here.
func Run(cfg *config.Config, svc *service.Service) error {
	_ = cfg
	_ = svc
	return fmt.Errorf("ui: not implemented (go-gui)")
}
