package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
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

func (r *Repository) GetByProductID(ctx context.Context, productID string) ([]productimages.ProductImage, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release(connection)

	images, err := bob.All(ctx, connection.BobTransactor(), getByProductIDQuery(productID), scan.StructMapper[productimages.ProductImage]())
	if err != nil {
		return nil, mapError(err)
	}
	if images == nil {
		return []productimages.ProductImage{}, nil
	}
	return images, nil
}

func getByProductIDQuery(productID string) bob.Query {
	return psql.Select(
		productImageColumns(),
		sm.From("product_images"),
		sm.Where(psql.Quote("product_id").EQ(psql.Arg(productID))),
		sm.OrderBy(psql.Quote("display_order")),
		sm.OrderBy(psql.Quote("created_at")),
		sm.OrderBy(psql.Quote("id")),
	)
}

func getByObjectKeyQuery(key string) bob.Query {
	return psql.Select(
		productImageColumns(),
		sm.From("product_images"),
		sm.Where(psql.Quote("r2_object_key").EQ(psql.Arg(key))),
	)
}

func productImageColumns() bob.Mod[*dialect.SelectQuery] {
	return sm.Columns(
		psql.Cast(psql.Quote("id"), "text"),
		psql.Cast(psql.Quote("product_id"), "text"),
		psql.Quote("r2_object_key"), psql.Quote("original_filename"),
		psql.Quote("content_type"), psql.Quote("file_size_bytes"),
		psql.Quote("width_px"), psql.Quote("height_px"),
		psql.Quote("alt_text"), psql.Quote("display_order"),
		psql.Quote("is_primary"), psql.Quote("r2_etag"),
		psql.Quote("confirmation_fingerprint"),
		psql.Quote("created_at"), psql.Quote("updated_at"),
	)
}
