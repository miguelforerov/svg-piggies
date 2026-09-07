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
	"github.com/zdenaforero/svg-piggies/backend/internal/productrelationships"
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

func (r *Repository) ListByProduct(
	ctx context.Context,
	productID string,
) ([]productrelationships.ProductRelationship, error) {
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
			return nil, productrelationships.ErrReferenceNotFound
		}
		return nil, mapError(err)
	}

	result, err := bob.All(
		ctx,
		executor,
		listRelationshipsQuery(productID),
		scan.StructMapper[productrelationships.ProductRelationship](),
	)
	if err != nil {
		return nil, mapError(err)
	}
	if result == nil {
		return []productrelationships.ProductRelationship{}, nil
	}
	return result, nil
}

func (r *Repository) Get(
	ctx context.Context,
	productID string,
	relationshipID string,
) (productrelationships.ProductRelationship, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return productrelationships.ProductRelationship{}, err
	}
	defer release(connection)

	relationship, err := bob.One(
		ctx,
		connection.BobTransactor(),
		getRelationshipQuery(productID, relationshipID),
		scan.StructMapper[productrelationships.ProductRelationship](),
	)
	if err != nil {
		return productrelationships.ProductRelationship{}, mapError(err)
	}
	return relationship, nil
}

func (r *Repository) Create(
	ctx context.Context,
	input productrelationships.CreateProductRelationshipInput,
) (productrelationships.ProductRelationship, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return productrelationships.ProductRelationship{}, err
	}
	defer release(connection)

	relationship, err := bob.One(
		ctx,
		connection.BobTransactor(),
		createRelationshipQuery(input),
		scan.StructMapper[productrelationships.ProductRelationship](),
	)
	if err != nil {
		return productrelationships.ProductRelationship{}, mapError(err)
	}
	return relationship, nil
}

func (r *Repository) CreateMany(
	ctx context.Context,
	input productrelationships.CreateProductRelationshipsInput,
) (productrelationships.ProductWithRelationships, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return productrelationships.ProductWithRelationships{}, err
	}
	defer release(connection)

	var result productrelationships.ProductWithRelationships
	err = database.RunInBobTransaction(
		ctx,
		connection.BobTransactor(),
		func(ctx context.Context, transaction bob.Transaction) error {
			product, err := getLockedProduct(ctx, transaction, input.ProductID)
			if err != nil {
				return err
			}

			nextDisplayOrder, err := bob.One(
				ctx,
				transaction,
				nextDisplayOrderQuery(input.ProductID),
				scan.SingleColumnMapper[int],
			)
			if err != nil {
				return mapError(err)
			}

			for _, relatedProductID := range input.RelatedProductIDs {
				insertResult, err := bob.Exec(
					ctx,
					transaction,
					createRelationshipIgnoringConflictQuery(
						input.ProductID,
						relatedProductID,
						nextDisplayOrder,
					),
				)
				if err != nil {
					return mapError(err)
				}
				rowsAffected, err := insertResult.RowsAffected()
				if err != nil {
					return mapError(err)
				}
				if rowsAffected > 0 {
					nextDisplayOrder++
				}
			}

			relationships, err := getPopulatedRelationships(ctx, transaction, input.ProductID)
			if err != nil {
				return err
			}
			result = productrelationships.ProductWithRelationships{
				Product:       product,
				Relationships: relationships,
			}
			return nil
		},
	)
	if err != nil {
		return productrelationships.ProductWithRelationships{}, mapError(err)
	}
	return result, nil
}

func (r *Repository) Update(
	ctx context.Context,
	productID string,
	relationshipID string,
	input productrelationships.UpdateProductRelationshipInput,
) (productrelationships.ProductRelationship, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return productrelationships.ProductRelationship{}, err
	}
	defer release(connection)

	relationship, err := bob.One(
		ctx,
		connection.BobTransactor(),
		updateRelationshipQuery(productID, relationshipID, input),
		scan.StructMapper[productrelationships.ProductRelationship](),
	)
	if err != nil {
		return productrelationships.ProductRelationship{}, mapError(err)
	}
	return relationship, nil
}

func (r *Repository) Replace(
	ctx context.Context,
	input productrelationships.ReplaceProductRelationshipsInput,
) (productrelationships.ProductWithRelationships, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return productrelationships.ProductWithRelationships{}, err
	}
	defer release(connection)

	var result productrelationships.ProductWithRelationships
	err = database.RunInBobTransaction(
		ctx,
		connection.BobTransactor(),
		func(ctx context.Context, transaction bob.Transaction) error {
			product, err := getLockedProduct(ctx, transaction, input.ProductID)
			if err != nil {
				return err
			}

			if _, err := bob.Exec(
				ctx,
				transaction,
				deleteRelationshipsByProductQuery(input.ProductID),
			); err != nil {
				return mapError(err)
			}

			for displayOrder, relatedProductID := range input.RelatedProductIDs {
				if _, err := bob.Exec(
					ctx,
					transaction,
					insertRelationshipQuery(input.ProductID, relatedProductID, displayOrder),
				); err != nil {
					return mapError(err)
				}
			}

			relationships, err := getPopulatedRelationships(ctx, transaction, input.ProductID)
			if err != nil {
				return err
			}
			result = productrelationships.ProductWithRelationships{
				Product:       product,
				Relationships: relationships,
			}
			return nil
		},
	)
	if err != nil {
		return productrelationships.ProductWithRelationships{}, mapError(err)
	}
	return result, nil
}

func (r *Repository) Delete(
	ctx context.Context,
	productID string,
	relationshipID string,
) error {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(connection)

	result, err := bob.Exec(
		ctx,
		connection.BobTransactor(),
		deleteRelationshipQuery(productID, relationshipID),
	)
	if err != nil {
		return mapError(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return mapError(err)
	}
	if rowsAffected == 0 {
		return productrelationships.ErrNotFound
	}
	return nil
}

func relationshipColumns() []any {
	return []any{
		psql.Cast(psql.Quote("id"), "text"),
		psql.Cast(psql.Quote("product_id"), "text"),
		psql.Cast(psql.Quote("related_product_id"), "text"),
		psql.Quote("display_order"),
	}
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

func productExistsQuery(productID string) bob.Query {
	return psql.Select(
		sm.Columns(psql.Cast(psql.Quote("id"), "text")),
		sm.From("products"),
		sm.Where(psql.Quote("id").EQ(psql.Arg(productID))),
	)
}

func lockedProductQuery(productID string) bob.Query {
	return psql.Select(
		sm.Columns(productColumns()...),
		sm.From("products"),
		sm.Where(psql.Quote("id").EQ(psql.Arg(productID))),
		sm.ForUpdate(),
	)
}

func listRelationshipsQuery(productID string) bob.Query {
	return psql.Select(
		sm.Columns(relationshipColumns()...),
		sm.From("product_relationships"),
		sm.Where(psql.Quote("product_id").EQ(psql.Arg(productID))),
		sm.OrderBy(psql.Quote("display_order")),
		sm.OrderBy(psql.Quote("id")),
	)
}

func getRelationshipQuery(productID string, relationshipID string) bob.Query {
	return psql.Select(
		sm.Columns(relationshipColumns()...),
		sm.From("product_relationships"),
		sm.Where(psql.And(
			psql.Quote("product_id").EQ(psql.Arg(productID)),
			psql.Quote("id").EQ(psql.Arg(relationshipID)),
		)),
	)
}

func createRelationshipQuery(input productrelationships.CreateProductRelationshipInput) bob.Query {
	return psql.Insert(
		im.Into("product_relationships", "product_id", "related_product_id", "display_order"),
		im.Values(psql.Arg(input.ProductID, input.RelatedProductID, input.DisplayOrder)),
		im.Returning(relationshipColumns()...),
	)
}

func nextDisplayOrderQuery(productID string) bob.Query {
	nextOrder := psql.F("MAX", psql.Quote("display_order"))().Plus(psql.Arg(1))
	return psql.Select(
		sm.Columns(psql.F("COALESCE", nextOrder, psql.Arg(0))),
		sm.From("product_relationships"),
		sm.Where(psql.Quote("product_id").EQ(psql.Arg(productID))),
	)
}

func createRelationshipIgnoringConflictQuery(
	productID string,
	relatedProductID string,
	displayOrder int,
) bob.Query {
	return psql.Insert(
		im.Into("product_relationships", "product_id", "related_product_id", "display_order"),
		im.Values(psql.Arg(productID, relatedProductID, displayOrder)),
		im.OnConflict(
			psql.Quote("product_id"),
			psql.Quote("related_product_id"),
		).DoNothing(),
	)
}

func insertRelationshipQuery(
	productID string,
	relatedProductID string,
	displayOrder int,
) bob.Query {
	return psql.Insert(
		im.Into("product_relationships", "product_id", "related_product_id", "display_order"),
		im.Values(psql.Arg(productID, relatedProductID, displayOrder)),
	)
}

func updateRelationshipQuery(
	productID string,
	relationshipID string,
	input productrelationships.UpdateProductRelationshipInput,
) bob.Query {
	return psql.Update(
		um.Table("product_relationships"),
		um.SetCol("related_product_id").ToArg(input.RelatedProductID),
		um.SetCol("display_order").ToArg(input.DisplayOrder),
		um.Where(psql.And(
			psql.Quote("product_id").EQ(psql.Arg(productID)),
			psql.Quote("id").EQ(psql.Arg(relationshipID)),
		)),
		um.Returning(relationshipColumns()...),
	)
}

func deleteRelationshipsByProductQuery(productID string) bob.Query {
	return psql.Delete(
		dm.From("product_relationships"),
		dm.Where(psql.Quote("product_id").EQ(psql.Arg(productID))),
	)
}

func deleteRelationshipQuery(productID string, relationshipID string) bob.Query {
	return psql.Delete(
		dm.From("product_relationships"),
		dm.Where(psql.And(
			psql.Quote("product_id").EQ(psql.Arg(productID)),
			psql.Quote("id").EQ(psql.Arg(relationshipID)),
		)),
	)
}

type populatedRelationshipRow struct {
	RelationshipID     string          `db:"relationship_id"`
	DisplayOrder       int             `db:"display_order"`
	ProductID          string          `db:"product_id"`
	ProductTitle       string          `db:"product_title"`
	ProductSlug        string          `db:"product_slug"`
	ProductDescription string          `db:"product_description"`
	ProductPrice       string          `db:"product_price"`
	ProductStatus      products.Status `db:"product_status"`
	ProductCreatedAt   time.Time       `db:"product_created_at"`
	ProductUpdatedAt   time.Time       `db:"product_updated_at"`
}

func populatedRelationshipsQuery(productID string) bob.Query {
	return psql.Select(
		sm.Columns(
			psql.Cast(psql.Quote("pr", "id"), "text").As("relationship_id"),
			psql.Quote("pr", "display_order").As("display_order"),
			psql.Cast(psql.Quote("p", "id"), "text").As("product_id"),
			psql.Quote("p", "title").As("product_title"),
			psql.Quote("p", "slug").As("product_slug"),
			psql.Quote("p", "description").As("product_description"),
			psql.Cast(psql.Quote("p", "price"), "text").As("product_price"),
			psql.Cast(psql.Quote("p", "status"), "text").As("product_status"),
			psql.Quote("p", "created_at").As("product_created_at"),
			psql.Quote("p", "updated_at").As("product_updated_at"),
		),
		sm.From("product_relationships").As("pr"),
		sm.InnerJoin("products").As("p").OnEQ(
			psql.Quote("p", "id"),
			psql.Quote("pr", "related_product_id"),
		),
		sm.Where(psql.Quote("pr", "product_id").EQ(psql.Arg(productID))),
		sm.OrderBy(psql.Quote("pr", "display_order")),
		sm.OrderBy(psql.Quote("pr", "id")),
	)
}

func getLockedProduct(
	ctx context.Context,
	transaction bob.Transaction,
	productID string,
) (products.Product, error) {
	product, err := bob.One(
		ctx,
		transaction,
		lockedProductQuery(productID),
		scan.StructMapper[products.Product](),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return products.Product{}, productrelationships.ErrReferenceNotFound
		}
		return products.Product{}, mapError(err)
	}
	return product, nil
}

func getPopulatedRelationships(
	ctx context.Context,
	transaction bob.Transaction,
	productID string,
) ([]productrelationships.PopulatedProductRelationship, error) {
	rows, err := bob.All(
		ctx,
		transaction,
		populatedRelationshipsQuery(productID),
		scan.StructMapper[populatedRelationshipRow](),
	)
	if err != nil {
		return nil, mapError(err)
	}

	result := make([]productrelationships.PopulatedProductRelationship, 0, len(rows))
	for _, row := range rows {
		result = append(result, productrelationships.PopulatedProductRelationship{
			RelationshipID: row.RelationshipID,
			DisplayOrder:   row.DisplayOrder,
			Product: products.Product{
				ID:          row.ProductID,
				Title:       row.ProductTitle,
				Slug:        row.ProductSlug,
				Description: row.ProductDescription,
				Price:       row.ProductPrice,
				Status:      row.ProductStatus,
				CreatedAt:   row.ProductCreatedAt,
				UpdatedAt:   row.ProductUpdatedAt,
			},
		})
	}
	return result, nil
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return productrelationships.ErrNotFound
	}
	if errors.Is(err, productrelationships.ErrInvalidInput) ||
		errors.Is(err, productrelationships.ErrNotFound) ||
		errors.Is(err, productrelationships.ErrReferenceNotFound) ||
		errors.Is(err, productrelationships.ErrConflict) {
		return err
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23503":
			return productrelationships.ErrReferenceNotFound
		case "23505":
			return productrelationships.ErrConflict
		case "23514":
			return productrelationships.ErrInvalidInput
		}
	}
	return fmt.Errorf("product relationships database operation: %w", err)
}

func release(connection *database.Connection) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = connection.Release(ctx)
}
