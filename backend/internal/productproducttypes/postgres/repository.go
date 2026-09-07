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
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/scan"
	"github.com/zdenaforero/svg-piggies/backend/internal/database"
	"github.com/zdenaforero/svg-piggies/backend/internal/productproducttypes"
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

func (r *Repository) ListByProduct(
	ctx context.Context,
	productID string,
) ([]productproducttypes.ProductProductType, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release(connection)

	executor := connection.BobTransactor()
	if _, err := bob.One(
		ctx,
		executor,
		productExistsQuery(productID),
		scan.SingleColumnMapper[string],
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, productproducttypes.ErrReferenceNotFound
		}
		return nil, mapError(err)
	}

	result, err := bob.All(
		ctx,
		executor,
		listProductProductTypesQuery(productID),
		scan.StructMapper[productproducttypes.ProductProductType](),
	)
	if err != nil {
		return nil, mapError(err)
	}
	if result == nil {
		return []productproducttypes.ProductProductType{}, nil
	}
	return result, nil
}

func (r *Repository) Create(
	ctx context.Context,
	input productproducttypes.CreateProductProductTypeInput,
) (productproducttypes.ProductProductType, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return productproducttypes.ProductProductType{}, err
	}
	defer release(connection)

	productProductType, err := bob.One(
		ctx,
		connection.BobTransactor(),
		createProductProductTypeQuery(input),
		scan.StructMapper[productproducttypes.ProductProductType](),
	)
	if err != nil {
		return productproducttypes.ProductProductType{}, mapError(err)
	}
	return productProductType, nil
}

func (r *Repository) Delete(
	ctx context.Context,
	productID string,
	productTypeID string,
) error {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(connection)

	result, err := bob.Exec(
		ctx,
		connection.BobTransactor(),
		deleteProductProductTypeQuery(productID, productTypeID),
	)
	if err != nil {
		return mapError(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return mapError(err)
	}
	if rowsAffected == 0 {
		return productproducttypes.ErrNotFound
	}
	return nil
}

func productProductTypeColumns() []any {
	return []any{
		psql.Cast(psql.Quote("product_id"), "text"),
		psql.Cast(psql.Quote("product_type_id"), "text"),
	}
}

func productExistsQuery(productID string) bob.Query {
	return psql.Select(
		sm.Columns(psql.Cast(psql.Quote("id"), "text")),
		sm.From("products"),
		sm.Where(psql.Quote("id").EQ(psql.Arg(productID))),
	)
}

func listProductProductTypesQuery(productID string) bob.Query {
	return psql.Select(
		sm.Columns(productProductTypeColumns()...),
		sm.From("product_product_types"),
		sm.Where(psql.Quote("product_id").EQ(psql.Arg(productID))),
		sm.OrderBy(psql.Quote("product_type_id")),
	)
}

func createProductProductTypeQuery(
	input productproducttypes.CreateProductProductTypeInput,
) bob.Query {
	return psql.Insert(
		im.Into("product_product_types", "product_id", "product_type_id"),
		im.Values(psql.Arg(input.ProductID, input.ProductTypeID)),
		im.Returning(productProductTypeColumns()...),
	)
}

func deleteProductProductTypeQuery(productID string, productTypeID string) bob.Query {
	return psql.Delete(
		dm.From("product_product_types"),
		dm.Where(psql.And(
			psql.Quote("product_id").EQ(psql.Arg(productID)),
			psql.Quote("product_type_id").EQ(psql.Arg(productTypeID)),
		)),
	)
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return productproducttypes.ErrNotFound
	}
	if errors.Is(err, productproducttypes.ErrInvalidInput) ||
		errors.Is(err, productproducttypes.ErrNotFound) ||
		errors.Is(err, productproducttypes.ErrReferenceNotFound) ||
		errors.Is(err, productproducttypes.ErrConflict) {
		return err
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23503":
			return productproducttypes.ErrReferenceNotFound
		case "23505":
			return productproducttypes.ErrConflict
		}
	}
	return fmt.Errorf("product product types database operation: %w", err)
}

func release(connection *database.Connection) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = connection.Release(ctx)
}
