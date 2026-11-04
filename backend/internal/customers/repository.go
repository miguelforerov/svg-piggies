package customers

import "context"

type Repository interface {
	List(ctx context.Context) ([]Customer, error)
	Get(ctx context.Context, id string) (Customer, error)
	Create(ctx context.Context, input CreateCustomerInput) (Customer, error)
	Update(ctx context.Context, id string, input UpdateCustomerInput) (Customer, error)
	Delete(ctx context.Context, id string) error
}
