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
	"github.com/stephenafamo/bob/dialect/psql/um"
	"github.com/stephenafamo/scan"
	"github.com/zdenaforero/svg-piggies/backend/internal/database"
	"github.com/zdenaforero/svg-piggies/backend/internal/producttypes"
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

func (r *Repository) List(ctx context.Context) ([]producttypes.ProductType, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release(connection)

	result, err := bob.All(
		ctx,
		connection.BobTransactor(),
		listProductTypesQuery(),
		scan.StructMapper[producttypes.ProductType](),
	)
	if err != nil {
		return nil, mapError(err)
	}
	return result, nil
}

func (r *Repository) Get(ctx context.Context, id string) (producttypes.ProductType, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return producttypes.ProductType{}, err
	}
	defer release(connection)

	productType, err := bob.One(
		ctx,
		connection.BobTransactor(),
		getProductTypeQuery("id", id),
		scan.StructMapper[producttypes.ProductType](),
	)
	if err != nil {
		return producttypes.ProductType{}, mapError(err)
	}
	return productType, nil
}

func (r *Repository) GetBySlug(
	ctx context.Context,
	slug string,
) (producttypes.ProductType, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return producttypes.ProductType{}, err
	}
	defer release(connection)

	productType, err := bob.One(
		ctx,
		connection.BobTransactor(),
		getProductTypeQuery("slug", slug),
		scan.StructMapper[producttypes.ProductType](),
	)
	if err != nil {
		return producttypes.ProductType{}, mapError(err)
	}
	return productType, nil
}

func (r *Repository) Create(
	ctx context.Context,
	input producttypes.CreateProductTypeInput,
) (producttypes.ProductType, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return producttypes.ProductType{}, err
	}
	defer release(connection)

	productType, err := bob.One(
		ctx,
		connection.BobTransactor(),
		createProductTypeQuery(input),
		scan.StructMapper[producttypes.ProductType](),
	)
	if err != nil {
		return producttypes.ProductType{}, mapError(err)
	}
	return productType, nil
}

func (r *Repository) Update(
	ctx context.Context,
	id string,
	input producttypes.UpdateProductTypeInput,
) (producttypes.ProductType, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return producttypes.ProductType{}, err
	}
	defer release(connection)

	productType, err := bob.One(
		ctx,
		connection.BobTransactor(),
		updateProductTypeQuery(id, input),
		scan.StructMapper[producttypes.ProductType](),
	)
	if err != nil {
		return producttypes.ProductType{}, mapError(err)
	}
	return productType, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(connection)

	result, err := bob.Exec(ctx, connection.BobTransactor(), deleteProductTypeQuery(id))
	if err != nil {
		return mapError(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return mapError(err)
	}
	if rowsAffected == 0 {
		return producttypes.ErrNotFound
	}
	return nil
}

func productTypeColumns() []any {
	return []any{
		psql.Cast(psql.Quote("id"), "text"),
		psql.Quote("name"),
		psql.Quote("slug"),
		psql.Quote("description"),
	}
}

func listProductTypesQuery() bob.Query {
	return psql.Select(
		sm.Columns(productTypeColumns()...),
		sm.From("product_types"),
		sm.OrderBy(psql.Quote("name")),
		sm.OrderBy(psql.Quote("id")),
	)
}

func getProductTypeQuery(column string, value string) bob.Query {
	return psql.Select(
		sm.Columns(productTypeColumns()...),
		sm.From("product_types"),
		sm.Where(psql.Quote(column).EQ(psql.Arg(value))),
	)
}

func createProductTypeQuery(input producttypes.CreateProductTypeInput) bob.Query {
	return psql.Insert(
		im.Into("product_types", "name", "slug", "description"),
		im.Values(psql.Arg(input.Name, input.Slug, input.Description)),
		im.Returning(productTypeColumns()...),
	)
}

func updateProductTypeQuery(id string, input producttypes.UpdateProductTypeInput) bob.Query {
	return psql.Update(
		um.Table("product_types"),
		um.SetCol("name").ToArg(input.Name),
		um.SetCol("slug").ToArg(input.Slug),
		um.SetCol("description").ToArg(input.Description),
		um.Where(psql.Quote("id").EQ(psql.Arg(id))),
		um.Returning(productTypeColumns()...),
	)
}

func deleteProductTypeQuery(id string) bob.Query {
	return psql.Delete(
		dm.From("product_types"),
		dm.Where(psql.Quote("id").EQ(psql.Arg(id))),
	)
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return producttypes.ErrNotFound
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return producttypes.ErrConflict
	}
	return fmt.Errorf("product types database operation: %w", err)
}

func release(connection *database.Connection) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = connection.Release(ctx)
}
