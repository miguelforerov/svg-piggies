package collections

import "context"

type Repository interface {
	List(ctx context.Context) ([]Collection, error)
	Get(ctx context.Context, id string) (Collection, error)
	Create(ctx context.Context, input CreateCollectionInput) (Collection, error)
	Update(ctx context.Context, id string, input UpdateCollectionInput) (Collection, error)
	Delete(ctx context.Context, id string) error
}
