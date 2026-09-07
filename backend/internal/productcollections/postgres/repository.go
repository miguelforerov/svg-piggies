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
	"github.com/zdenaforero/svg-piggies/backend/internal/productcollections"
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
) ([]productcollections.ProductCollection, error) {
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
			return nil, productcollections.ErrReferenceNotFound
		}
		return nil, mapError(err)
	}

	result, err := bob.All(
		ctx,
		executor,
		listProductCollectionsQuery(productID),
		scan.StructMapper[productcollections.ProductCollection](),
	)
	if err != nil {
		return nil, mapError(err)
	}
	if result == nil {
		return []productcollections.ProductCollection{}, nil
	}
	return result, nil
}

func (r *Repository) Create(
	ctx context.Context,
	input productcollections.CreateProductCollectionInput,
) (productcollections.ProductCollection, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return productcollections.ProductCollection{}, err
	}
	defer release(connection)

	productCollection, err := bob.One(
		ctx,
		connection.BobTransactor(),
		createProductCollectionQuery(input),
		scan.StructMapper[productcollections.ProductCollection](),
	)
	if err != nil {
		return productcollections.ProductCollection{}, mapError(err)
	}
	return productCollection, nil
}

func (r *Repository) Delete(
	ctx context.Context,
	productID string,
	collectionID string,
) error {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(connection)

	result, err := bob.Exec(
		ctx,
		connection.BobTransactor(),
		deleteProductCollectionQuery(productID, collectionID),
	)
	if err != nil {
		return mapError(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return mapError(err)
	}
	if rowsAffected == 0 {
		return productcollections.ErrNotFound
	}
	return nil
}

func productCollectionColumns() []any {
	return []any{
		psql.Cast(psql.Quote("product_id"), "text"),
		psql.Cast(psql.Quote("collection_id"), "text"),
	}
}

func productExistsQuery(productID string) bob.Query {
	return psql.Select(
		sm.Columns(psql.Cast(psql.Quote("id"), "text")),
		sm.From("products"),
		sm.Where(psql.Quote("id").EQ(psql.Arg(productID))),
	)
}

func listProductCollectionsQuery(productID string) bob.Query {
	return psql.Select(
		sm.Columns(productCollectionColumns()...),
		sm.From("product_collections"),
		sm.Where(psql.Quote("product_id").EQ(psql.Arg(productID))),
		sm.OrderBy(psql.Quote("collection_id")),
	)
}

func createProductCollectionQuery(
	input productcollections.CreateProductCollectionInput,
) bob.Query {
	return psql.Insert(
		im.Into("product_collections", "product_id", "collection_id"),
		im.Values(psql.Arg(input.ProductID, input.CollectionID)),
		im.Returning(productCollectionColumns()...),
	)
}

func deleteProductCollectionQuery(productID string, collectionID string) bob.Query {
	return psql.Delete(
		dm.From("product_collections"),
		dm.Where(psql.And(
			psql.Quote("product_id").EQ(psql.Arg(productID)),
			psql.Quote("collection_id").EQ(psql.Arg(collectionID)),
		)),
	)
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return productcollections.ErrNotFound
	}
	if errors.Is(err, productcollections.ErrInvalidInput) ||
		errors.Is(err, productcollections.ErrNotFound) ||
		errors.Is(err, productcollections.ErrReferenceNotFound) ||
		errors.Is(err, productcollections.ErrConflict) {
		return err
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23503":
			return productcollections.ErrReferenceNotFound
		case "23505":
			return productcollections.ErrConflict
		}
	}
	return fmt.Errorf("product collections database operation: %w", err)
}

func release(connection *database.Connection) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = connection.Release(ctx)
}
