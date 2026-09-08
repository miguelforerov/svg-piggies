package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
	"github.com/stephenafamo/scan"
	"github.com/zdenaforero/svg-piggies/backend/internal/database"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
)

type Repository struct {
	provider database.Provider
}

func NewRepository(provider database.Provider) (*Repository, error) {
	if provider == nil {
		return nil, errors.New("database provider is required")
	}
	return &Repository{provider: provider}, nil
}

func (r *Repository) CreateMany(
	ctx context.Context,
	input productimages.AddProductImagesInput,
) ([]productimages.ProductImage, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release(connection)

	var created []productimages.ProductImage
	err = database.RunInBobTransaction(
		ctx,
		connection.BobTransactor(),
		func(ctx context.Context, transaction bob.Transaction) error {
			if _, err := bob.One(
				ctx,
				transaction,
				lockProductQuery(input.ProductID),
				scan.SingleColumnMapper[string],
			); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return productimages.ErrReferenceNotFound
				}
				return mapError(err)
			}

			if containsPrimaryImage(input.Images) {
				if _, err := bob.Exec(
					ctx,
					transaction,
					clearPrimaryImageQuery(input.ProductID),
				); err != nil {
					return mapError(err)
				}
			}

			created = make([]productimages.ProductImage, 0, len(input.Images))
			for _, inputImage := range input.Images {
				image, err := bob.One(
					ctx,
					transaction,
					insertProductImageQuery(input.ProductID, inputImage),
					scan.StructMapper[productimages.ProductImage](),
				)
				if err != nil {
					return mapError(err)
				}
				created = append(created, image)
			}
			return nil
		},
	)
	if err != nil {
		return nil, mapError(err)
	}
	return created, nil
}

func lockProductQuery(productID string) bob.Query {
	return psql.Select(
		sm.Columns(psql.Cast(psql.Quote("id"), "text")),
		sm.From("products"),
		sm.Where(psql.Quote("id").EQ(psql.Arg(productID))),
		sm.ForUpdate(),
	)
}

func clearPrimaryImageQuery(productID string) bob.Query {
	return psql.Update(
		um.Table("product_images"),
		um.SetCol("is_primary").ToArg(false),
		um.Where(psql.And(
			psql.Quote("product_id").EQ(psql.Arg(productID)),
			psql.Quote("is_primary").EQ(psql.Arg(true)),
		)),
	)
}

func insertProductImageQuery(
	productID string,
	image productimages.NewProductImage,
) bob.Query {
	return psql.Insert(
		im.Into(
			"product_images",
			"product_id",
			"r2_object_key",
			"original_filename",
			"content_type",
			"file_size_bytes",
			"width_px",
			"height_px",
			"alt_text",
			"display_order",
			"is_primary",
			"r2_etag",
		),
		im.Values(psql.Arg(
			productID,
			image.R2ObjectKey,
			image.OriginalFilename,
			image.ContentType,
			image.FileSizeBytes,
			image.WidthPX,
			image.HeightPX,
			image.AltText,
			image.DisplayOrder,
			image.IsPrimary,
			image.R2ETag,
		)),
		im.Returning(
			psql.Cast(psql.Quote("id"), "text"),
			psql.Cast(psql.Quote("product_id"), "text"),
			psql.Quote("r2_object_key"),
			psql.Quote("original_filename"),
			psql.Quote("content_type"),
			psql.Quote("file_size_bytes"),
			psql.Quote("width_px"),
			psql.Quote("height_px"),
			psql.Quote("alt_text"),
			psql.Quote("display_order"),
			psql.Quote("is_primary"),
			psql.Quote("r2_etag"),
			psql.Quote("created_at"),
			psql.Quote("updated_at"),
		),
	)
}

func containsPrimaryImage(images []productimages.NewProductImage) bool {
	for _, image := range images {
		if image.IsPrimary {
			return true
		}
	}
	return false
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return productimages.ErrReferenceNotFound
	}
	if errors.Is(err, productimages.ErrInvalidInput) ||
		errors.Is(err, productimages.ErrReferenceNotFound) ||
		errors.Is(err, productimages.ErrConflict) {
		return err
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23503":
			return productimages.ErrReferenceNotFound
		case "23505":
			return productimages.ErrConflict
		case "23514", "23502", "22P02":
			return productimages.ErrInvalidInput
		}
	}
	return fmt.Errorf("product images database operation: %w", err)
}

func release(connection *database.Connection) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = connection.Release(ctx)
}
