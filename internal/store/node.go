package store

import (
	"context"
	"database/sql"
	"time"
)

// ListNodes returns active nodes for a collection.
func (d *DB) ListNodes(ctx context.Context, collectionID int64) ([]Node, error) {
	var rows []Node
	err := d.NewSelect().Model(&rows).
		Where("collection_id = ?", collectionID).
		OrderExpr("sort_order ASC, id ASC").
		Scan(ctx)
	return rows, err
}

// GetNode returns a node by id.
func (d *DB) GetNode(ctx context.Context, id int64) (*Node, error) {
	n := new(Node)
	if err := d.NewSelect().Model(n).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}
	return n, nil
}

// CreateNode inserts a node under parent (nil = root).
func (d *DB) CreateNode(ctx context.Context, n *Node) error {
	now := time.Now().UTC()
	var maxOrder sql.NullInt64
	q := d.NewSelect().Model((*Node)(nil)).ColumnExpr("MAX(sort_order)").
		Where("collection_id = ?", n.CollectionID)
	if n.ParentID == nil {
		q = q.Where("parent_id IS NULL")
	} else {
		q = q.Where("parent_id = ?", *n.ParentID)
	}
	_ = q.Scan(ctx, &maxOrder)
	if maxOrder.Valid {
		n.SortOrder = int(maxOrder.Int64) + 1
	}
	n.CreatedAt = now
	n.UpdatedAt = now
	_, err := d.NewInsert().Model(n).Exec(ctx)
	return err
}

// UpdateNode saves mutable fields.
func (d *DB) UpdateNode(ctx context.Context, n *Node) error {
	n.UpdatedAt = time.Now().UTC()
	res, err := d.NewUpdate().Model(n).
		Column("name", "parent_id", "sort_order", "markdown", "http", "protocol", "updated_at").
		WherePK().
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SoftDeleteNode marks a node and its descendants deleted.
func (d *DB) SoftDeleteNode(ctx context.Context, id int64) error {
	kids, err := d.descendantNodeIDs(ctx, id)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	res, err := d.ExecContext(ctx,
		`UPDATE nodes SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`,
		now, now, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	for _, nid := range kids {
		if _, err := d.ExecContext(ctx,
			`UPDATE nodes SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`,
			now, now, nid); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) descendantNodeIDs(ctx context.Context, rootID int64) ([]int64, error) {
	var out []int64
	frontier := []int64{rootID}
	for len(frontier) > 0 {
		pid := frontier[0]
		frontier = frontier[1:]
		var kids []Node
		if err := d.NewSelect().Model(&kids).
			Column("id").
			Where("parent_id = ?", pid).
			Scan(ctx); err != nil {
			return nil, err
		}
		for _, k := range kids {
			out = append(out, k.ID)
			frontier = append(frontier, k.ID)
		}
	}
	return out, nil
}

// RestoreNode clears deleted_at on the node and soft-deleted descendants.
func (d *DB) RestoreNode(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	res, err := d.ExecContext(ctx,
		`UPDATE nodes SET deleted_at = NULL, updated_at = ? WHERE id = ? AND deleted_at IS NOT NULL`,
		now, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return d.restoreDescendantNodes(ctx, id, now)
}

func (d *DB) restoreDescendantNodes(ctx context.Context, parentID int64, now time.Time) error {
	var kids []Node
	// Include soft-deleted children (Bun soft-delete filter would hide them).
	if err := d.NewSelect().Model(&kids).
		Column("id").
		Where("parent_id = ?", parentID).
		WhereDeleted().
		Scan(ctx); err != nil {
		return err
	}
	for _, k := range kids {
		if _, err := d.ExecContext(ctx,
			`UPDATE nodes SET deleted_at = NULL, updated_at = ? WHERE id = ? AND deleted_at IS NOT NULL`,
			now, k.ID); err != nil {
			return err
		}
		if err := d.restoreDescendantNodes(ctx, k.ID, now); err != nil {
			return err
		}
	}
	return nil
}

// ListTrashedNodes returns soft-deleted nodes whose collection is still live
// (collection-level deletes are restored via RestoreCollection).
func (d *DB) ListTrashedNodes(ctx context.Context) ([]Node, error) {
	var rows []Node
	err := d.NewSelect().Model(&rows).
		WhereDeleted().
		Where("collection_id IN (SELECT id FROM collections WHERE deleted_at IS NULL)").
		OrderExpr("deleted_at DESC").
		Scan(ctx)
	return rows, err
}

// FindNodeByName finds a live node by collection, parent, kind, name.
func (d *DB) FindNodeByName(ctx context.Context, collectionID int64, parentID *int64, kind, name string) (*Node, error) {
	n := new(Node)
	q := d.NewSelect().Model(n).
		Where("collection_id = ? AND kind = ? AND name = ?", collectionID, kind, name)
	if parentID == nil {
		q = q.Where("parent_id IS NULL")
	} else {
		q = q.Where("parent_id = ?", *parentID)
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	return n, nil
}
