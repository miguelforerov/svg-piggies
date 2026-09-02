package productrelationships

import "context"

type Repository interface {
	ListByProduct(ctx context.Context, productID string) ([]ProductRelationship, error)
	Get(ctx context.Context, productID string, relationshipID string) (ProductRelationship, error)
	Create(ctx context.Context, input CreateProductRelationshipInput) (ProductRelationship, error)
	CreateMany(
		ctx context.Context,
		input CreateProductRelationshipsInput,
	) (ProductWithRelationships, error)
	Update(
		ctx context.Context,
		productID string,
		relationshipID string,
		input UpdateProductRelationshipInput,
	) (ProductRelationship, error)
	Replace(
		ctx context.Context,
		input ReplaceProductRelationshipsInput,
	) (ProductWithRelationships, error)
	Delete(ctx context.Context, productID string, relationshipID string) error
}
