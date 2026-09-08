package productimages

import "context"

type Repository interface {
	CreateMany(
		ctx context.Context,
		input AddProductImagesInput,
	) ([]ProductImage, error)
}
