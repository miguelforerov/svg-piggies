package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
)

func (r *Repository) UpdateAltText(ctx context.Context, input productimages.AltTextUpdate) error {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(connection)
	tx, err := connection.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin alternative text update: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	var productID string
	err = tx.QueryRow(ctx, `SELECT id::text FROM products WHERE id = $1 FOR UPDATE`, input.ProductID).Scan(&productID)
	if errors.Is(err, pgx.ErrNoRows) {
		return productimages.ErrReferenceNotFound
	}
	if err != nil {
		return fmt.Errorf("find product for alternative text update: %w", err)
	}
	var imageID string
	err = tx.QueryRow(ctx, `SELECT id::text FROM product_images WHERE id = $1 AND product_id = $2`, input.ImageID, input.ProductID).Scan(&imageID)
	if errors.Is(err, pgx.ErrNoRows) {
		return productimages.ErrImageNotFound
	}
	if err != nil {
		return fmt.Errorf("find image for alternative text update: %w", err)
	}
	if input.Present {
		var value any
		if input.AltText != nil {
			value = *input.AltText
		}
		result, err := tx.Exec(ctx, `UPDATE product_images SET alt_text = $3 WHERE id = $1 AND product_id = $2`, input.ImageID, input.ProductID, value)
		if err != nil {
			return fmt.Errorf("write alternative text: %w", err)
		}
		if result.RowsAffected() != 1 {
			return productimages.ErrImageNotFound
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit alternative text update: %w", err)
	}
	return nil
}
