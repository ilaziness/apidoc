// Package store is the Bun/SQLite persistence layer.
//
// Driver must be pure Go (modernc.org/sqlite). Do not use mattn/go-sqlite3,
// and do not rely on Bun sqliteshim when CGO is enabled (it may pick mattn).
package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite" // pure-Go SQLite driver (PRD: no mattn/CGO)
)

// DefaultDBFile is the SQLite filename inside the user data directory.
const DefaultDBFile = "apidoc.db"

// DB wraps bun.DB.
type DB struct {
	*bun.DB
	path string
}

// Open opens (or creates) the SQLite file under dataDir and applies migrations.
func Open(dataDir string) (*DB, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("store: empty data dir")
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("store: mkdir: %w", err)
	}
	path := filepath.Join(dataDir, DefaultDBFile)
	dsn := "file:" + filepath.ToSlash(path) + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	sqldb, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: open: %w", err)
	}
	sqldb.SetMaxOpenConns(1)
	db := &DB{
		DB:   bun.NewDB(sqldb, sqlitedialect.New()),
		path: path,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	if err := db.Migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// Path returns the database file path.
func (d *DB) Path() string {
	return d.path
}

// Close releases database resources.
func (d *DB) Close() error {
	if d == nil || d.DB == nil {
		return nil
	}
	return d.DB.Close()
}
