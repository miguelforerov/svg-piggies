package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
)

func (r *Repository) Reorder(ctx context.Context, productID string, imageIDs []string) error {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(connection)
	tx, err := connection.Begin(ctx)
	if err != nil {
		return mapError(err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	var lockedID string
	err = tx.QueryRow(ctx, `SELECT id::text FROM products WHERE id = $1 FOR UPDATE`, productID).Scan(&lockedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return productimages.ErrReferenceNotFound
	}
	if err != nil {
		return mapError(err)
	}
	rows, err := tx.Query(ctx, `SELECT id::text FROM product_images WHERE product_id = $1`, productID)
	if err != nil {
		return mapError(err)
	}
	current := make(map[string]struct{})
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return mapError(err)
		}
		current[id] = struct{}{}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return mapError(err)
	}
	if len(current) != len(imageIDs) {
		return productimages.ErrInvalidInput
	}
	for _, id := range imageIDs {
		if _, ok := current[id]; !ok {
			return productimages.ErrInvalidInput
		}
	}
	for position, id := range imageIDs {
		_, err := tx.Exec(ctx, `UPDATE product_images SET display_order = $1 WHERE id = $2 AND product_id = $3 AND display_order IS DISTINCT FROM $1`, position, id, productID)
		if err != nil {
			return mapError(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit image order: %w", err)
	}
	return nil
}
