package productproducttypes

import "context"

type Repository interface {
	ListByProduct(ctx context.Context, productID string) ([]ProductProductType, error)
	Create(ctx context.Context, input CreateProductProductTypeInput) (ProductProductType, error)
	Delete(ctx context.Context, productID string, productTypeID string) error
}
