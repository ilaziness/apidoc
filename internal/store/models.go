package store

import (
	"time"

	"github.com/uptrace/bun"
)

// Collection is a top-level API project (文档集).
type Collection struct {
	bun.BaseModel `bun:"table:collections"`

	ID          int64      `bun:",pk,autoincrement"`
	Name        string     `bun:",notnull"`
	DefaultAuth string     `bun:",nullzero,type:text"` // JSON AuthSpec
	SortOrder   int        `bun:",notnull"`
	CreatedAt   time.Time  `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt   time.Time  `bun:",nullzero,notnull,default:current_timestamp"`
	DeletedAt   *time.Time `bun:",nullzero,soft_delete"`
}

// Node is a tree entry: folder | markdown | http.
type Node struct {
	bun.BaseModel `bun:"table:nodes"`

	ID           int64      `bun:",pk,autoincrement"`
	CollectionID int64      `bun:",notnull"`
	ParentID     *int64     `bun:",nullzero"`
	Kind         string     `bun:",notnull"` // folder | markdown | http
	Name         string     `bun:",notnull"`
	Protocol     string     `bun:",nullzero"` // http | grpc | ...
	SortOrder    int        `bun:",notnull"`
	Markdown     string     `bun:",nullzero,type:text"`
	HTTP         string     `bun:",nullzero,type:text"` // JSON HTTPSpec
	CreatedAt    time.Time  `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt    time.Time  `bun:",nullzero,notnull,default:current_timestamp"`
	DeletedAt    *time.Time `bun:",nullzero,soft_delete"`
}

// Environment is a named variable set inside a collection.
type Environment struct {
	bun.BaseModel `bun:"table:environments"`

	ID           int64      `bun:",pk,autoincrement"`
	CollectionID int64      `bun:",notnull"`
	Name         string     `bun:",notnull"`
	Variables    string     `bun:",nullzero,type:text"` // JSON map[string]string
	SortOrder    int        `bun:",notnull"`
	CreatedAt    time.Time  `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt    time.Time  `bun:",nullzero,notnull,default:current_timestamp"`
	DeletedAt    *time.Time `bun:",nullzero,soft_delete"`
}

const (
	NodeFolder   = "folder"
	NodeMarkdown = "markdown"
	NodeHTTP     = "http"
)
