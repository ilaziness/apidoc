package store

import "context"

// migrate runs Bun migrations embedded beside this package.
// Prefer Bun's migrate package over ad-hoc CREATE TABLE IF NOT EXISTS.
func migrate(ctx context.Context, db *DB) error {
	_ = ctx
	_ = db
	// TODO: bun migrate.NewMigrations() + Discover/MustRegister, then Migrator.Up.
	return nil
}
