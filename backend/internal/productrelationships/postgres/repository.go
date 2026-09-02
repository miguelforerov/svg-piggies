package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

	var productExists bool
	if err := connection.QueryRow(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)`,
		productID,
	).Scan(&productExists); err != nil {
		return nil, mapError(err)
	}
	if !productExists {
		return nil, productrelationships.ErrReferenceNotFound
	}

	rows, err := connection.Query(ctx, `
		SELECT id::text, product_id::text, related_product_id::text, display_order
		FROM product_relationships
		WHERE product_id = $1
		ORDER BY display_order, id
	`, productID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	result := make([]productrelationships.ProductRelationship, 0)
	for rows.Next() {
		relationship, scanErr := scanProductRelationship(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, relationship)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(err)
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

	return scanProductRelationship(connection.QueryRow(ctx, `
		SELECT id::text, product_id::text, related_product_id::text, display_order
		FROM product_relationships
		WHERE product_id = $1 AND id = $2
	`, productID, relationshipID))
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

	return scanProductRelationship(connection.QueryRow(ctx, `
		INSERT INTO product_relationships (product_id, related_product_id, display_order)
		VALUES ($1, $2, $3)
		RETURNING id::text, product_id::text, related_product_id::text, display_order
	`, input.ProductID, input.RelatedProductID, input.DisplayOrder))
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

	transaction, err := connection.Begin(ctx)
	if err != nil {
		return productrelationships.ProductWithRelationships{}, mapError(err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	product, err := scanProduct(transaction.QueryRow(ctx, `
		SELECT id::text, title, slug, description, price::text, status::text,
		       created_at, updated_at
		FROM products
		WHERE id = $1
		FOR UPDATE
	`, input.ProductID))
	if err != nil {
		if errors.Is(err, productrelationships.ErrNotFound) {
			return productrelationships.ProductWithRelationships{},
				productrelationships.ErrReferenceNotFound
		}
		return productrelationships.ProductWithRelationships{}, err
	}

	var nextDisplayOrder int
	if err := transaction.QueryRow(ctx, `
		SELECT COALESCE(MAX(display_order) + 1, 0)
		FROM product_relationships
		WHERE product_id = $1
	`, input.ProductID).Scan(&nextDisplayOrder); err != nil {
		return productrelationships.ProductWithRelationships{}, mapError(err)
	}

	for _, relatedProductID := range input.RelatedProductIDs {
		commandTag, err := transaction.Exec(ctx, `
			INSERT INTO product_relationships (product_id, related_product_id, display_order)
			VALUES ($1, $2, $3)
			ON CONFLICT (product_id, related_product_id) DO NOTHING
		`, input.ProductID, relatedProductID, nextDisplayOrder)
		if err != nil {
			return productrelationships.ProductWithRelationships{}, mapError(err)
		}
		if commandTag.RowsAffected() > 0 {
			nextDisplayOrder++
		}
	}

	rows, err := transaction.Query(ctx, `
		SELECT pr.id::text, pr.display_order,
		       p.id::text, p.title, p.slug, p.description, p.price::text,
		       p.status::text, p.created_at, p.updated_at
		FROM product_relationships pr
		JOIN products p ON p.id = pr.related_product_id
		WHERE pr.product_id = $1
		ORDER BY pr.display_order, pr.id
	`, input.ProductID)
	if err != nil {
		return productrelationships.ProductWithRelationships{}, mapError(err)
	}

	relationships := make([]productrelationships.PopulatedProductRelationship, 0)
	for rows.Next() {
		var relationship productrelationships.PopulatedProductRelationship
		if err := rows.Scan(
			&relationship.RelationshipID,
			&relationship.DisplayOrder,
			&relationship.Product.ID,
			&relationship.Product.Title,
			&relationship.Product.Slug,
			&relationship.Product.Description,
			&relationship.Product.Price,
			&relationship.Product.Status,
			&relationship.Product.CreatedAt,
			&relationship.Product.UpdatedAt,
		); err != nil {
			rows.Close()
			return productrelationships.ProductWithRelationships{}, mapError(err)
		}
		relationships = append(relationships, relationship)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return productrelationships.ProductWithRelationships{}, mapError(err)
	}
	rows.Close()

	if err := transaction.Commit(ctx); err != nil {
		return productrelationships.ProductWithRelationships{}, mapError(err)
	}
	return productrelationships.ProductWithRelationships{
		Product:       product,
		Relationships: relationships,
	}, nil
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

	return scanProductRelationship(connection.QueryRow(ctx, `
		UPDATE product_relationships
		SET related_product_id = $3, display_order = $4
		WHERE product_id = $1 AND id = $2
		RETURNING id::text, product_id::text, related_product_id::text, display_order
	`, productID, relationshipID, input.RelatedProductID, input.DisplayOrder))
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

	transaction, err := connection.Begin(ctx)
	if err != nil {
		return productrelationships.ProductWithRelationships{}, mapError(err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	product, err := scanProduct(transaction.QueryRow(ctx, `
		SELECT id::text, title, slug, description, price::text, status::text,
		       created_at, updated_at
		FROM products
		WHERE id = $1
		FOR UPDATE
	`, input.ProductID))
	if err != nil {
		if errors.Is(err, productrelationships.ErrNotFound) {
			return productrelationships.ProductWithRelationships{},
				productrelationships.ErrReferenceNotFound
		}
		return productrelationships.ProductWithRelationships{}, err
	}

	if _, err := transaction.Exec(
		ctx,
		`DELETE FROM product_relationships WHERE product_id = $1`,
		input.ProductID,
	); err != nil {
		return productrelationships.ProductWithRelationships{}, mapError(err)
	}

	for displayOrder, relatedProductID := range input.RelatedProductIDs {
		if _, err := transaction.Exec(ctx, `
			INSERT INTO product_relationships (product_id, related_product_id, display_order)
			VALUES ($1, $2, $3)
		`, input.ProductID, relatedProductID, displayOrder); err != nil {
			return productrelationships.ProductWithRelationships{}, mapError(err)
		}
	}

	rows, err := transaction.Query(ctx, `
		SELECT pr.id::text, pr.display_order,
		       p.id::text, p.title, p.slug, p.description, p.price::text,
		       p.status::text, p.created_at, p.updated_at
		FROM product_relationships pr
		JOIN products p ON p.id = pr.related_product_id
		WHERE pr.product_id = $1
		ORDER BY pr.display_order, pr.id
	`, input.ProductID)
	if err != nil {
		return productrelationships.ProductWithRelationships{}, mapError(err)
	}

	relationships := make([]productrelationships.PopulatedProductRelationship, 0)
	for rows.Next() {
		var relationship productrelationships.PopulatedProductRelationship
		if err := rows.Scan(
			&relationship.RelationshipID,
			&relationship.DisplayOrder,
			&relationship.Product.ID,
			&relationship.Product.Title,
			&relationship.Product.Slug,
			&relationship.Product.Description,
			&relationship.Product.Price,
			&relationship.Product.Status,
			&relationship.Product.CreatedAt,
			&relationship.Product.UpdatedAt,
		); err != nil {
			rows.Close()
			return productrelationships.ProductWithRelationships{}, mapError(err)
		}
		relationships = append(relationships, relationship)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return productrelationships.ProductWithRelationships{}, mapError(err)
	}
	rows.Close()

	if err := transaction.Commit(ctx); err != nil {
		return productrelationships.ProductWithRelationships{}, mapError(err)
	}
	return productrelationships.ProductWithRelationships{
		Product:       product,
		Relationships: relationships,
	}, nil
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

	commandTag, err := connection.Exec(ctx, `
		DELETE FROM product_relationships
		WHERE product_id = $1 AND id = $2
	`, productID, relationshipID)
	if err != nil {
		return mapError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return productrelationships.ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProductRelationship(row scanner) (productrelationships.ProductRelationship, error) {
	var relationship productrelationships.ProductRelationship
	if err := row.Scan(
		&relationship.ID,
		&relationship.ProductID,
		&relationship.RelatedProductID,
		&relationship.DisplayOrder,
	); err != nil {
		return productrelationships.ProductRelationship{}, mapError(err)
	}
	return relationship, nil
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
		return productrelationships.ErrNotFound
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
