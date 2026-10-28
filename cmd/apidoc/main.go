// Command apidoc is the ApiDoc desktop application entrypoint.
package main

import (
	"fmt"
	"os"

	"github.com/ilaziness/apidoc/internal/config"
	"github.com/ilaziness/apidoc/internal/protocol/httpapi"
	"github.com/ilaziness/apidoc/internal/service"
	"github.com/ilaziness/apidoc/internal/store"
	"github.com/ilaziness/apidoc/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "apidoc: %v\n", err)
		os.Exit(1)
	}
}

// Startup: load app config → open SQLite (migrates) → UI.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	db, err := store.Open(cfg.DataDir)
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}
	defer db.Close()

	svc := service.New(service.Deps{
		Store:  db,
		Config: cfg,
		HTTP:   httpapi.New(),
	})

	return ui.Run(cfg, svc)
}
