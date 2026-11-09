package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
)

func (r *Repository) GetByIDAndProduct(ctx context.Context, productID, imageID string) (productimages.ProductImage, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return productimages.ProductImage{}, err
	}
	defer release(connection)
	var image productimages.ProductImage
	err = connection.QueryRow(ctx, `SELECT id::text, product_id::text, r2_object_key
		FROM product_images WHERE id = $1 AND product_id = $2`, imageID, productID).
		Scan(&image.ID, &image.ProductID, &image.R2ObjectKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return productimages.ProductImage{}, productimages.ErrImageNotFound
	}
	if err != nil {
		return productimages.ProductImage{}, mapError(err)
	}
	return image, nil
}

func (r *Repository) DeleteByIDAndProduct(ctx context.Context, productID, imageID string) error {
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

	// Serialize primary-image changes with image creation and other deletions.
	var lockedID string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM products WHERE id = $1 FOR UPDATE`, productID).Scan(&lockedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return productimages.ErrReferenceNotFound
		}
		return mapError(err)
	}
	var wasPrimary bool
	err = tx.QueryRow(ctx, `DELETE FROM product_images WHERE id = $1 AND product_id = $2 RETURNING is_primary`, imageID, productID).Scan(&wasPrimary)
	if errors.Is(err, pgx.ErrNoRows) {
		return productimages.ErrImageNotFound
	}
	if err != nil {
		return mapError(err)
	}
	if wasPrimary {
		_, err = tx.Exec(ctx, `UPDATE product_images SET is_primary = true WHERE id = (
			SELECT id FROM product_images WHERE product_id = $1
			ORDER BY display_order, created_at, id LIMIT 1
		)`, productID)
		if err != nil {
			return mapError(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit product image deletion: %w", err)
	}
	return nil
}
