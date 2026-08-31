package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
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

	var productExists bool
	if err := connection.QueryRow(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)`,
		productID,
	).Scan(&productExists); err != nil {
		return nil, mapError(err)
	}
	if !productExists {
		return nil, productproducttypes.ErrReferenceNotFound
	}

	rows, err := connection.Query(ctx, `
		SELECT product_id::text, product_type_id::text
		FROM product_product_types
		WHERE product_id = $1
		ORDER BY product_type_id
	`, productID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	result := make([]productproducttypes.ProductProductType, 0)
	for rows.Next() {
		productProductType, scanErr := scanProductProductType(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, productProductType)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(err)
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

	return scanProductProductType(connection.QueryRow(ctx, `
		INSERT INTO product_product_types (product_id, product_type_id)
		VALUES ($1, $2)
		RETURNING product_id::text, product_type_id::text
	`, input.ProductID, input.ProductTypeID))
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

	commandTag, err := connection.Exec(ctx, `
		DELETE FROM product_product_types
		WHERE product_id = $1 AND product_type_id = $2
	`, productID, productTypeID)
	if err != nil {
		return mapError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return productproducttypes.ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProductProductType(row scanner) (productproducttypes.ProductProductType, error) {
	var productProductType productproducttypes.ProductProductType
	if err := row.Scan(
		&productProductType.ProductID,
		&productProductType.ProductTypeID,
	); err != nil {
		return productproducttypes.ProductProductType{}, mapError(err)
	}
	return productProductType, nil
}

func mapError(err error) error {
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
