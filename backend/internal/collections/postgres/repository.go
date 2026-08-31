package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zdenaforero/svg-piggies/backend/internal/collections"
	"github.com/zdenaforero/svg-piggies/backend/internal/database"
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

func (r *Repository) List(ctx context.Context) ([]collections.Collection, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release(connection)

	rows, err := connection.Query(ctx, `
		SELECT id::text, name, slug, description
		FROM collections
		ORDER BY name, id
	`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	result := make([]collections.Collection, 0)
	for rows.Next() {
		collection, scanErr := scanCollection(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, collection)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(err)
	}
	return result, nil
}

func (r *Repository) Get(ctx context.Context, id string) (collections.Collection, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return collections.Collection{}, err
	}
	defer release(connection)

	return scanCollection(connection.QueryRow(ctx, `
		SELECT id::text, name, slug, description
		FROM collections
		WHERE id = $1
	`, id))
}

func (r *Repository) Create(
	ctx context.Context,
	input collections.CreateCollectionInput,
) (collections.Collection, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return collections.Collection{}, err
	}
	defer release(connection)

	return scanCollection(connection.QueryRow(ctx, `
		INSERT INTO collections (name, slug, description)
		VALUES ($1, $2, $3)
		RETURNING id::text, name, slug, description
	`, input.Name, input.Slug, input.Description))
}

func (r *Repository) Update(
	ctx context.Context,
	id string,
	input collections.UpdateCollectionInput,
) (collections.Collection, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return collections.Collection{}, err
	}
	defer release(connection)

	return scanCollection(connection.QueryRow(ctx, `
		UPDATE collections
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

	commandTag, err := connection.Exec(ctx, `DELETE FROM collections WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return collections.ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanCollection(row scanner) (collections.Collection, error) {
	var collection collections.Collection
	if err := row.Scan(
		&collection.ID,
		&collection.Name,
		&collection.Slug,
		&collection.Description,
	); err != nil {
		return collections.Collection{}, mapError(err)
	}
	return collection, nil
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return collections.ErrNotFound
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return collections.ErrConflict
	}
	return fmt.Errorf("collections database operation: %w", err)
}

func release(connection *database.Connection) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = connection.Release(ctx)
}
