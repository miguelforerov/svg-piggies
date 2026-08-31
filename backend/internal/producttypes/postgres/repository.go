package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

	rows, err := connection.Query(ctx, `
		SELECT id::text, name, slug, description
		FROM product_types
		ORDER BY name, id
	`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	result := make([]producttypes.ProductType, 0)
	for rows.Next() {
		productType, scanErr := scanProductType(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, productType)
	}
	if err := rows.Err(); err != nil {
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

	return scanProductType(connection.QueryRow(ctx, `
		SELECT id::text, name, slug, description
		FROM product_types
		WHERE id = $1
	`, id))
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

	return scanProductType(connection.QueryRow(ctx, `
		INSERT INTO product_types (name, slug, description)
		VALUES ($1, $2, $3)
		RETURNING id::text, name, slug, description
	`, input.Name, input.Slug, input.Description))
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

	return scanProductType(connection.QueryRow(ctx, `
		UPDATE product_types
		SET name = $2, slug = $3, description = $4
		WHERE id = $1
		RETURNING id::text, name, slug, description
	`, id, input.Name, input.Slug, input.Description))
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(connection)

	commandTag, err := connection.Exec(ctx, `DELETE FROM product_types WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return producttypes.ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProductType(row scanner) (producttypes.ProductType, error) {
	var productType producttypes.ProductType
	if err := row.Scan(
		&productType.ID,
		&productType.Name,
		&productType.Slug,
		&productType.Description,
	); err != nil {
		return producttypes.ProductType{}, mapError(err)
	}
	return productType, nil
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
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
