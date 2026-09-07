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

	result, err := bob.All(
		ctx,
		connection.BobTransactor(),
		listCollectionsQuery(),
		scan.StructMapper[collections.Collection](),
	)
	if err != nil {
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

	collection, err := bob.One(
		ctx,
		connection.BobTransactor(),
		getCollectionQuery("id", id),
		scan.StructMapper[collections.Collection](),
	)
	if err != nil {
		return collections.Collection{}, mapError(err)
	}
	return collection, nil
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (collections.Collection, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return collections.Collection{}, err
	}
	defer release(connection)

	collection, err := bob.One(
		ctx,
		connection.BobTransactor(),
		getCollectionQuery("slug", slug),
		scan.StructMapper[collections.Collection](),
	)
	if err != nil {
		return collections.Collection{}, mapError(err)
	}
	return collection, nil
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

	collection, err := bob.One(
		ctx,
		connection.BobTransactor(),
		createCollectionQuery(input),
		scan.StructMapper[collections.Collection](),
	)
	if err != nil {
		return collections.Collection{}, mapError(err)
	}
	return collection, nil
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

	collection, err := bob.One(
		ctx,
		connection.BobTransactor(),
		updateCollectionQuery(id, input),
		scan.StructMapper[collections.Collection](),
	)
	if err != nil {
		return collections.Collection{}, mapError(err)
	}
	return collection, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(connection)

	result, err := bob.Exec(ctx, connection.BobTransactor(), deleteCollectionQuery(id))
	if err != nil {
		return mapError(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return mapError(err)
	}
	if rowsAffected == 0 {
		return collections.ErrNotFound
	}
	return nil
}

func collectionColumns() []any {
	return []any{
		psql.Cast(psql.Quote("id"), "text"),
		psql.Quote("name"),
		psql.Quote("slug"),
		psql.Quote("description"),
	}
}

func listCollectionsQuery() bob.Query {
	return psql.Select(
		sm.Columns(collectionColumns()...),
		sm.From("collections"),
		sm.OrderBy(psql.Quote("name")),
		sm.OrderBy(psql.Quote("id")),
	)
}

func getCollectionQuery(column string, value string) bob.Query {
	return psql.Select(
		sm.Columns(collectionColumns()...),
		sm.From("collections"),
		sm.Where(psql.Quote(column).EQ(psql.Arg(value))),
	)
}

func createCollectionQuery(input collections.CreateCollectionInput) bob.Query {
	return psql.Insert(
		im.Into("collections", "name", "slug", "description"),
		im.Values(psql.Arg(input.Name, input.Slug, input.Description)),
		im.Returning(collectionColumns()...),
	)
}

func updateCollectionQuery(id string, input collections.UpdateCollectionInput) bob.Query {
	return psql.Update(
		um.Table("collections"),
		um.SetCol("name").ToArg(input.Name),
		um.SetCol("slug").ToArg(input.Slug),
		um.SetCol("description").ToArg(input.Description),
		um.Where(psql.Quote("id").EQ(psql.Arg(id))),
		um.Returning(collectionColumns()...),
	)
}

func deleteCollectionQuery(id string) bob.Query {
	return psql.Delete(
		dm.From("collections"),
		dm.Where(psql.Quote("id").EQ(psql.Arg(id))),
	)
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
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
