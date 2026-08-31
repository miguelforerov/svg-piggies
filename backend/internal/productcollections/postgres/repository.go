package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
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

	var productExists bool
	if err := connection.QueryRow(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)`,
		productID,
	).Scan(&productExists); err != nil {
		return nil, mapError(err)
	}
	if !productExists {
		return nil, productcollections.ErrReferenceNotFound
	}

	rows, err := connection.Query(ctx, `
		SELECT product_id::text, collection_id::text
		FROM product_collections
		WHERE product_id = $1
		ORDER BY collection_id
	`, productID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	result := make([]productcollections.ProductCollection, 0)
	for rows.Next() {
		productCollection, scanErr := scanProductCollection(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, productCollection)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(err)
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

	return scanProductCollection(connection.QueryRow(ctx, `
		INSERT INTO product_collections (product_id, collection_id)
		VALUES ($1, $2)
		RETURNING product_id::text, collection_id::text
	`, input.ProductID, input.CollectionID))
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

	commandTag, err := connection.Exec(ctx, `
		DELETE FROM product_collections
		WHERE product_id = $1 AND collection_id = $2
	`, productID, collectionID)
	if err != nil {
		return mapError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return productcollections.ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProductCollection(row scanner) (productcollections.ProductCollection, error) {
	var productCollection productcollections.ProductCollection
	if err := row.Scan(
		&productCollection.ProductID,
		&productCollection.CollectionID,
	); err != nil {
		return productcollections.ProductCollection{}, mapError(err)
	}
	return productCollection, nil
}

func mapError(err error) error {
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
