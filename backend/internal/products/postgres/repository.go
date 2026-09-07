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
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
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

func (r *Repository) List(ctx context.Context) ([]products.Product, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release(connection)

	result, err := bob.All(
		ctx,
		connection.BobTransactor(),
		listProductsQuery(),
		scan.StructMapper[products.Product](),
	)
	if err != nil {
		return nil, mapError(err)
	}
	return result, nil
}

func (r *Repository) Get(ctx context.Context, id string) (products.Product, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return products.Product{}, err
	}
	defer release(connection)

	product, err := bob.One(
		ctx,
		connection.BobTransactor(),
		getProductQuery("id", id),
		scan.StructMapper[products.Product](),
	)
	if err != nil {
		return products.Product{}, mapError(err)
	}
	return product, nil
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (products.Product, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return products.Product{}, err
	}
	defer release(connection)

	product, err := bob.One(
		ctx,
		connection.BobTransactor(),
		getProductQuery("slug", slug),
		scan.StructMapper[products.Product](),
	)
	if err != nil {
		return products.Product{}, mapError(err)
	}
	return product, nil
}

func (r *Repository) Create(
	ctx context.Context,
	input products.CreateProductInput,
) (products.Product, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return products.Product{}, err
	}
	defer release(connection)

	product, err := bob.One(
		ctx,
		connection.BobTransactor(),
		createProductQuery(input),
		scan.StructMapper[products.Product](),
	)
	if err != nil {
		return products.Product{}, mapError(err)
	}
	return product, nil
}

func (r *Repository) Update(
	ctx context.Context,
	id string,
	input products.UpdateProductInput,
) (products.Product, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return products.Product{}, err
	}
	defer release(connection)

	product, err := bob.One(
		ctx,
		connection.BobTransactor(),
		updateProductQuery(id, input),
		scan.StructMapper[products.Product](),
	)
	if err != nil {
		return products.Product{}, mapError(err)
	}
	return product, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(connection)

	result, err := bob.Exec(ctx, connection.BobTransactor(), deleteProductQuery(id))
	if err != nil {
		return mapError(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return mapError(err)
	}
	if rowsAffected == 0 {
		return products.ErrNotFound
	}
	return nil
}

func productColumns() []any {
	return []any{
		psql.Cast(psql.Quote("id"), "text"),
		psql.Quote("title"),
		psql.Quote("slug"),
		psql.Quote("description"),
		psql.Cast(psql.Quote("price"), "text"),
		psql.Cast(psql.Quote("status"), "text"),
		psql.Quote("created_at"),
		psql.Quote("updated_at"),
	}
}

func listProductsQuery() bob.Query {
	return psql.Select(
		sm.Columns(productColumns()...),
		sm.From("products"),
		sm.OrderBy(psql.Quote("created_at")).Desc(),
		sm.OrderBy(psql.Quote("id")),
	)
}

func getProductQuery(column string, value string) bob.Query {
	return psql.Select(
		sm.Columns(productColumns()...),
		sm.From("products"),
		sm.Where(psql.Quote(column).EQ(psql.Arg(value))),
	)
}

func createProductQuery(input products.CreateProductInput) bob.Query {
	return psql.Insert(
		im.Into("products", "title", "slug", "description", "price", "status"),
		im.Values(
			psql.Arg(input.Title),
			psql.Arg(input.Slug),
			psql.Arg(input.Description),
			psql.Cast(psql.Arg(input.Price), "numeric"),
			psql.Cast(psql.Arg(input.Status), "product_status"),
		),
		im.Returning(productColumns()...),
	)
}

func updateProductQuery(id string, input products.UpdateProductInput) bob.Query {
	return psql.Update(
		um.Table("products"),
		um.SetCol("title").ToArg(input.Title),
		um.SetCol("slug").ToArg(input.Slug),
		um.SetCol("description").ToArg(input.Description),
		um.SetCol("price").To(psql.Cast(psql.Arg(input.Price), "numeric")),
		um.SetCol("status").To(psql.Cast(psql.Arg(input.Status), "product_status")),
		um.Where(psql.Quote("id").EQ(psql.Arg(id))),
		um.Returning(productColumns()...),
	)
}

func deleteProductQuery(id string) bob.Query {
	return psql.Delete(
		dm.From("products"),
		dm.Where(psql.Quote("id").EQ(psql.Arg(id))),
	)
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return products.ErrNotFound
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return products.ErrConflict
	}
	return fmt.Errorf("products database operation: %w", err)
}

func release(connection *database.Connection) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = connection.Release(ctx)
}
