package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/scan"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
)

// CheckUploadSchema executes a zero-row read of the column required by both
// confirmation queries. PostgreSQL validates the column without reading data.
func (r *Repository) CheckUploadSchema(ctx context.Context) error {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(connection)

	rows, err := connection.Query(ctx, `SELECT confirmation_fingerprint FROM product_images LIMIT 0`)
	if err != nil {
		return err
	}
	rows.Close()
	return rows.Err()
}

func (r *Repository) GetByObjectKey(ctx context.Context, key string) (productimages.ProductImage, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return productimages.ProductImage{}, err
	}
	defer release(connection)

	image, err := bob.One(ctx, connection.BobTransactor(), getByObjectKeyQuery(key), scan.StructMapper[productimages.ProductImage]())
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
		return productimages.ProductImage{}, productimages.ErrImageNotFound
	}
	if err != nil {
		return productimages.ProductImage{}, mapError(err)
	}
	return image, nil
}

func getByObjectKeyQuery(key string) bob.Query {
	return psql.Select(
		sm.Columns(
			psql.Cast(psql.Quote("id"), "text"),
			psql.Cast(psql.Quote("product_id"), "text"),
			psql.Quote("r2_object_key"), psql.Quote("original_filename"),
			psql.Quote("content_type"), psql.Quote("file_size_bytes"),
			psql.Quote("width_px"), psql.Quote("height_px"),
			psql.Quote("alt_text"), psql.Quote("display_order"),
			psql.Quote("is_primary"), psql.Quote("r2_etag"),
			psql.Quote("confirmation_fingerprint"),
			psql.Quote("created_at"), psql.Quote("updated_at"),
		),
		sm.From("product_images"),
		sm.Where(psql.Quote("r2_object_key").EQ(psql.Arg(key))),
	)
}
