package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenMigrateCRUD(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	c, err := db.CreateCollection(ctx, "Demo")
	if err != nil {
		t.Fatal(err)
	}
	if c.ID == 0 {
		t.Fatal("expected id")
	}
	list, err := db.ListCollections(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%v err=%v", list, err)
	}

	folder := &Node{CollectionID: c.ID, Kind: NodeFolder, Name: "API"}
	if err := db.CreateNode(ctx, folder); err != nil {
		t.Fatal(err)
	}
	httpNode := &Node{
		CollectionID: c.ID,
		ParentID:     &folder.ID,
		Kind:         NodeHTTP,
		Name:         "Get users",
		Protocol:     "http",
		HTTP:         `{"method":"GET","url":"{{baseUrl}}/users"}`,
	}
	if err := db.CreateNode(ctx, httpNode); err != nil {
		t.Fatal(err)
	}
	nodes, err := db.ListNodes(ctx, c.ID)
	if err != nil || len(nodes) != 2 {
		t.Fatalf("nodes=%d err=%v", len(nodes), err)
	}

	env, err := db.CreateEnvironment(ctx, c.ID, "dev", `{"baseUrl":"https://example.com"}`)
	if err != nil {
		t.Fatal(err)
	}
	envs, err := db.ListEnvironments(ctx, c.ID)
	if err != nil || len(envs) != 1 || envs[0].ID != env.ID {
		t.Fatalf("envs=%v err=%v", envs, err)
	}

	if err := db.SoftDeleteCollection(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	list, err = db.ListCollections(ctx)
	if err != nil || len(list) != 0 {
		t.Fatalf("after delete list=%v", list)
	}
	nodes, err = db.ListNodes(ctx, c.ID)
	if err != nil || len(nodes) != 0 {
		t.Fatalf("cascade delete nodes=%d err=%v", len(nodes), err)
	}
	if err := db.RestoreCollection(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	list, err = db.ListCollections(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("after restore list=%v", list)
	}
	nodes, err = db.ListNodes(ctx, c.ID)
	if err != nil || len(nodes) != 2 {
		t.Fatalf("cascade restore nodes=%d err=%v", len(nodes), err)
	}

	// Folder soft-delete must cascade to children.
	folder2 := &Node{CollectionID: c.ID, Kind: NodeFolder, Name: "Nest"}
	if err := db.CreateNode(ctx, folder2); err != nil {
		t.Fatal(err)
	}
	child := &Node{CollectionID: c.ID, ParentID: &folder2.ID, Kind: NodeMarkdown, Name: "note", Markdown: "x"}
	if err := db.CreateNode(ctx, child); err != nil {
		t.Fatal(err)
	}
	if err := db.SoftDeleteNode(ctx, folder2.ID); err != nil {
		t.Fatal(err)
	}
	nodes, err = db.ListNodes(ctx, c.ID)
	if err != nil || len(nodes) != 2 {
		t.Fatalf("after folder delete want 2 roots, got %d err=%v", len(nodes), err)
	}
	if err := db.RestoreNode(ctx, folder2.ID); err != nil {
		t.Fatal(err)
	}
	nodes, err = db.ListNodes(ctx, c.ID)
	if err != nil || len(nodes) != 4 {
		t.Fatalf("after folder restore want 4, got %d err=%v", len(nodes), err)
	}

	if err := db.SoftDeleteCollection(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	// Soft-delete filters block Model Update; set deleted_at via raw SQL.
	past := time.Now().UTC().Add(-40 * 24 * time.Hour)
	if _, err = db.ExecContext(ctx, `UPDATE collections SET deleted_at = ? WHERE id = ?`, past, c.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.PurgeExpiredTrash(ctx, 30*24*time.Hour); err != nil {
		t.Fatal(err)
	}
	trashed, err := db.ListTrashedCollections(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(trashed) != 0 {
		t.Fatalf("expected purged, got %d", len(trashed))
	}

	if db.Path() != filepath.Join(dir, DefaultDBFile) {
		t.Fatalf("path=%s", db.Path())
	}
}
