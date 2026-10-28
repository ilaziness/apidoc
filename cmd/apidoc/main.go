// Command apidoc is the ApiDoc desktop application entrypoint.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ilaziness/apidoc/internal/config"
	"github.com/ilaziness/apidoc/internal/protocol"
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

// Startup order (docs/PRD.md §10):
// load app config → open SQLite in data dir → migrate → UI.
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

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	protocols := protocol.NewRegistry()
	protocols.Register(httpapi.New())

	svc := service.New(service.Deps{
		Store:     db,
		Protocols: protocols,
	})

	return ui.Run(cfg, svc)
}
