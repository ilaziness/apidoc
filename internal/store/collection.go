package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// ListCollections returns non-deleted collections ordered by sort_order, id.
func (d *DB) ListCollections(ctx context.Context) ([]Collection, error) {
	var rows []Collection
	err := d.NewSelect().Model(&rows).
		OrderExpr("sort_order ASC, id ASC").
		Scan(ctx)
	return rows, err
}

// GetCollection returns a collection by id (including soft-deleted when includeDeleted).
func (d *DB) GetCollection(ctx context.Context, id int64, includeDeleted bool) (*Collection, error) {
	c := new(Collection)
	q := d.NewSelect().Model(c).Where("id = ?", id)
	if !includeDeleted {
		q = q.Where("deleted_at IS NULL")
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

// CreateCollection inserts a collection.
func (d *DB) CreateCollection(ctx context.Context, name string) (*Collection, error) {
	now := time.Now().UTC()
	var maxOrder sql.NullInt64
	_ = d.NewSelect().Model((*Collection)(nil)).
		ColumnExpr("MAX(sort_order)").
		Where("deleted_at IS NULL").
		Scan(ctx, &maxOrder)
	order := 0
	if maxOrder.Valid {
		order = int(maxOrder.Int64) + 1
	}
	c := &Collection{
		Name:      name,
		SortOrder: order,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := d.NewInsert().Model(c).Exec(ctx)
	return c, err
}

// RenameCollection updates the name.
func (d *DB) RenameCollection(ctx context.Context, id int64, name string) error {
	res, err := d.NewUpdate().Model((*Collection)(nil)).
		Set("name = ?", name).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ? AND deleted_at IS NULL", id).
		Exec(ctx)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SoftDeleteCollection marks a collection and its nodes/envs deleted.
func (d *DB) SoftDeleteCollection(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	if _, err := d.ExecContext(ctx,
		`UPDATE nodes SET deleted_at = ?, updated_at = ? WHERE collection_id = ? AND deleted_at IS NULL`,
		now, now, id); err != nil {
		return err
	}
	if _, err := d.ExecContext(ctx,
		`UPDATE environments SET deleted_at = ?, updated_at = ? WHERE collection_id = ? AND deleted_at IS NULL`,
		now, now, id); err != nil {
		return err
	}
	res, err := d.NewDelete().Model((*Collection)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// RestoreCollection clears deleted_at on the collection and its nodes/envs.
func (d *DB) RestoreCollection(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	res, err := d.ExecContext(ctx,
		`UPDATE collections SET deleted_at = NULL, updated_at = ? WHERE id = ? AND deleted_at IS NOT NULL`,
		now, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	if _, err := d.ExecContext(ctx,
		`UPDATE nodes SET deleted_at = NULL, updated_at = ? WHERE collection_id = ? AND deleted_at IS NOT NULL`,
		now, id); err != nil {
		return err
	}
	if _, err := d.ExecContext(ctx,
		`UPDATE environments SET deleted_at = NULL, updated_at = ? WHERE collection_id = ? AND deleted_at IS NOT NULL`,
		now, id); err != nil {
		return err
	}
	return nil
}

// ListTrashedCollections returns soft-deleted collections.
func (d *DB) ListTrashedCollections(ctx context.Context) ([]Collection, error) {
	var rows []Collection
	err := d.NewSelect().Model(&rows).
		WhereDeleted().
		OrderExpr("deleted_at DESC").
		Scan(ctx)
	return rows, err
}

// PurgeExpiredTrash permanently deletes rows soft-deleted more than retention ago.
func (d *DB) PurgeExpiredTrash(ctx context.Context, retention time.Duration) error {
	cutoff := time.Now().UTC().Add(-retention)
	if _, err := d.ExecContext(ctx,
		`DELETE FROM nodes WHERE deleted_at IS NOT NULL AND deleted_at < ?
		 OR collection_id IN (SELECT id FROM collections WHERE deleted_at IS NOT NULL AND deleted_at < ?)`,
		cutoff, cutoff); err != nil {
		return err
	}
	if _, err := d.ExecContext(ctx,
		`DELETE FROM environments WHERE deleted_at IS NOT NULL AND deleted_at < ?
		 OR collection_id IN (SELECT id FROM collections WHERE deleted_at IS NOT NULL AND deleted_at < ?)`,
		cutoff, cutoff); err != nil {
		return err
	}
	if _, err := d.ExecContext(ctx, `DELETE FROM collections WHERE deleted_at IS NOT NULL AND deleted_at < ?`, cutoff); err != nil {
		return err
	}
	return nil
}

// EmptyTrash permanently deletes all soft-deleted rows.
func (d *DB) EmptyTrash(ctx context.Context) error {
	if _, err := d.ExecContext(ctx,
		`DELETE FROM nodes WHERE deleted_at IS NOT NULL
		 OR collection_id IN (SELECT id FROM collections WHERE deleted_at IS NOT NULL)`); err != nil {
		return err
	}
	if _, err := d.ExecContext(ctx,
		`DELETE FROM environments WHERE deleted_at IS NOT NULL
		 OR collection_id IN (SELECT id FROM collections WHERE deleted_at IS NOT NULL)`); err != nil {
		return err
	}
	if _, err := d.ExecContext(ctx, `DELETE FROM collections WHERE deleted_at IS NOT NULL`); err != nil {
		return err
	}
	return nil
}

// UpdateCollectionAuth sets default auth JSON.
func (d *DB) UpdateCollectionAuth(ctx context.Context, id int64, authJSON string) error {
	res, err := d.NewUpdate().Model((*Collection)(nil)).
		Set("default_auth = ?", authJSON).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ? AND deleted_at IS NULL", id).
		Exec(ctx)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("collection not found")
	}
	return nil
}
