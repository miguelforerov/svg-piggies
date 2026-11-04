package carts

import "context"

type Repository interface {
	CreateGuest(ctx context.Context, tokenHash string) (Cart, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (Cart, error)
	AddItem(
		ctx context.Context,
		tokenHash string,
		productID string,
		quantity int,
	) (Cart, error)
	UpdateItem(
		ctx context.Context,
		tokenHash string,
		productID string,
		quantity int,
	) (Cart, error)
	DeleteItem(ctx context.Context, tokenHash string, productID string) (Cart, error)
	Clear(ctx context.Context, tokenHash string) error
}
