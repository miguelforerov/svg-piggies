package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
)

// SetPrimary serializes with image creation and deletion by locking the parent
// product. The single UPDATE clears any previous primary and marks only the
// selected image, or the transaction rolls back without changing any rows.
func (r *Repository) SetPrimary(ctx context.Context, productID, imageID string) error {
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
	var selectedID string
	err = tx.QueryRow(ctx, `SELECT id::text FROM product_images WHERE id = $1 AND product_id = $2`, imageID, productID).Scan(&selectedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return productimages.ErrImageNotFound
	}
	if err != nil {
		return mapError(err)
	}
	_, err = tx.Exec(ctx, `UPDATE product_images
		SET is_primary = (id = $2)
		WHERE product_id = $1 AND is_primary IS DISTINCT FROM (id = $2)`, productID, imageID)
	if err != nil {
		return fmt.Errorf("update primary image: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit primary image change: %w", err)
	}
	return nil
}
