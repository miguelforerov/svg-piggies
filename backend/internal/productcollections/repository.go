package productcollections

import "context"

type Repository interface {
	ListByProduct(ctx context.Context, productID string) ([]ProductCollection, error)
	Create(ctx context.Context, input CreateProductCollectionInput) (ProductCollection, error)
	Delete(ctx context.Context, productID string, collectionID string) error
}
