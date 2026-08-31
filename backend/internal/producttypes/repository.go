package producttypes

import "context"

type Repository interface {
	List(ctx context.Context) ([]ProductType, error)
	Get(ctx context.Context, id string) (ProductType, error)
	Create(ctx context.Context, input CreateProductTypeInput) (ProductType, error)
	Update(ctx context.Context, id string, input UpdateProductTypeInput) (ProductType, error)
	Delete(ctx context.Context, id string) error
}
