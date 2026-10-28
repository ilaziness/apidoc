package store

import (
	"context"
	"embed"
	"fmt"

	"github.com/uptrace/bun/migrate"
)

//go:embed migrations/*.sql
var sqlMigrations embed.FS

// Migrate applies pending Bun migrations.
func (d *DB) Migrate(ctx context.Context) error {
	migrations := migrate.NewMigrations()
	if err := migrations.Discover(sqlMigrations); err != nil {
		return fmt.Errorf("store: discover migrations: %w", err)
	}
	migrator := migrate.NewMigrator(d.DB, migrations)
	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("store: migrate init: %w", err)
	}
	if err := migrator.Lock(ctx); err != nil {
		return fmt.Errorf("store: migrate lock: %w", err)
	}
	defer migrator.Unlock(ctx) //nolint:errcheck

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("store: migrate: %w", err)
	}
	_ = group
	return nil
}
