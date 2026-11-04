package orders

import "context"

type Repository interface {
	List(ctx context.Context) ([]Order, error)
	Get(ctx context.Context, id string) (Order, error)
	CreateFromCart(ctx context.Context, input CreateCheckoutOrderInput) (Order, error)
	UpdateStatus(ctx context.Context, id string, status Status) (Order, error)
}
