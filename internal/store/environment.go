package store

import (
	"context"
	"database/sql"
	"time"
)

// ListEnvironments returns environments for a collection.
func (d *DB) ListEnvironments(ctx context.Context, collectionID int64) ([]Environment, error) {
	var rows []Environment
	err := d.NewSelect().Model(&rows).
		Where("collection_id = ?", collectionID).
		OrderExpr("sort_order ASC, id ASC").
		Scan(ctx)
	return rows, err
}

// GetEnvironment returns an environment by id.
func (d *DB) GetEnvironment(ctx context.Context, id int64) (*Environment, error) {
	e := new(Environment)
	if err := d.NewSelect().Model(e).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}
	return e, nil
}

// CreateEnvironment inserts an environment.
func (d *DB) CreateEnvironment(ctx context.Context, collectionID int64, name, variablesJSON string) (*Environment, error) {
	now := time.Now().UTC()
	var maxOrder sql.NullInt64
	_ = d.NewSelect().Model((*Environment)(nil)).
		ColumnExpr("MAX(sort_order)").
		Where("collection_id = ?", collectionID).
		Scan(ctx, &maxOrder)
	order := 0
	if maxOrder.Valid {
		order = int(maxOrder.Int64) + 1
	}
	e := &Environment{
		CollectionID: collectionID,
		Name:         name,
		Variables:    variablesJSON,
		SortOrder:    order,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, err := d.NewInsert().Model(e).Exec(ctx)
	return e, err
}

// UpdateEnvironment updates name and variables.
func (d *DB) UpdateEnvironment(ctx context.Context, e *Environment) error {
	e.UpdatedAt = time.Now().UTC()
	res, err := d.NewUpdate().Model(e).
		Column("name", "variables", "updated_at").
		WherePK().
		Where("deleted_at IS NULL").
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

// SoftDeleteEnvironment marks an environment deleted.
func (d *DB) SoftDeleteEnvironment(ctx context.Context, id int64) error {
	res, err := d.NewDelete().Model((*Environment)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// RestoreEnvironment clears deleted_at.
func (d *DB) RestoreEnvironment(ctx context.Context, id int64) error {
	res, err := d.ExecContext(ctx,
		`UPDATE environments SET deleted_at = NULL, updated_at = ? WHERE id = ? AND deleted_at IS NOT NULL`,
		time.Now().UTC(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
