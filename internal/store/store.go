// Package store is the Bun/SQLite persistence layer.
//
// Driver must be pure Go (modernc.org/sqlite). Do not use mattn/go-sqlite3,
// and do not rely on Bun sqliteshim when CGO is enabled (it may pick mattn).
package store

import (
	"context"
	"fmt"
	"path/filepath"
)

// DefaultDBFile is the SQLite filename inside the user data directory.
const DefaultDBFile = "apidoc.db"

// DB wraps the database connection. Bun wiring lands here later.
type DB struct {
	path string
}

// Open opens (or creates) the SQLite file under dataDir.
func Open(dataDir string) (*DB, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("store: empty data dir")
	}
	return &DB{path: filepath.Join(dataDir, DefaultDBFile)}, nil
}

// Path returns the database file path.
func (d *DB) Path() string {
	return d.path
}

// Close releases database resources.
func (d *DB) Close() error {
	return nil
}

// Ping checks connectivity once Bun is wired.
func (d *DB) Ping(ctx context.Context) error {
	_ = ctx
	return nil
}

// Migrate applies pending Bun migrations. See migrations.go.
func (d *DB) Migrate(ctx context.Context) error {
	return migrate(ctx, d)
}
