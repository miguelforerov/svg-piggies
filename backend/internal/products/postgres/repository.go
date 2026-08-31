package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

	rows, err := connection.Query(ctx, `
		SELECT id::text, title, slug, description, price::text, status::text,
		       created_at, updated_at
		FROM products
		ORDER BY created_at DESC, id
	`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	result := make([]products.Product, 0)
	for rows.Next() {
		product, scanErr := scanProduct(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, product)
	}
	if err := rows.Err(); err != nil {
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

	return scanProduct(connection.QueryRow(ctx, `
		SELECT id::text, title, slug, description, price::text, status::text,
		       created_at, updated_at
		FROM products
		WHERE id = $1
	`, id))
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

	return scanProduct(connection.QueryRow(ctx, `
		INSERT INTO products (title, slug, description, price, status)
		VALUES ($1, $2, $3, $4::numeric, $5::product_status)
		RETURNING id::text, title, slug, description, price::text, status::text,
		          created_at, updated_at
	`, input.Title, input.Slug, input.Description, input.Price, input.Status))
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

	return scanProduct(connection.QueryRow(ctx, `
		UPDATE products
		SET title = $2,
		    slug = $3,
		    description = $4,
		    price = $5::numeric,
		    status = $6::product_status
		WHERE id = $1
		RETURNING id::text, title, slug, description, price::text, status::text,
		          created_at, updated_at
	`, id, input.Title, input.Slug, input.Description, input.Price, input.Status))
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(connection)

	commandTag, err := connection.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return products.ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProduct(row scanner) (products.Product, error) {
	var product products.Product
	if err := row.Scan(
		&product.ID,
		&product.Title,
		&product.Slug,
		&product.Description,
		&product.Price,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt,
	); err != nil {
		return products.Product{}, mapError(err)
	}
	return product, nil
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
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
