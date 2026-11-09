package productimages

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	"github.com/zdenaforero/svg-piggies/backend/internal/storage"
)

type ImageDeletionRepository interface {
	GetByIDAndProduct(context.Context, string, string) (ProductImage, error)
	DeleteByIDAndProduct(context.Context, string, string) error
}

type ImageObjectDeleter interface {
	DeleteObject(context.Context, storage.Bucket, string) error
}

type DeleteService struct {
	products ProductLookup
	images   ImageDeletionRepository
	storage  ImageObjectDeleter
}

func NewDeleteService(products ProductLookup, images ImageDeletionRepository, store ImageObjectDeleter) (*DeleteService, error) {
	if products == nil || images == nil || store == nil {
		return nil, errors.New("product lookup, image repository, and storage are required")
	}
	return &DeleteService{products: products, images: images, storage: store}, nil
}

func (s *DeleteService) DeleteProductImage(ctx context.Context, productID, imageID string) error {
	productID, imageID = strings.TrimSpace(productID), strings.TrimSpace(imageID)
	if productID == "" || imageID == "" {
		return fmt.Errorf("%w: product id and image id are required", ErrInvalidInput)
	}
	if _, err := s.products.GetProduct(ctx, productID); err != nil {
		switch {
		case errors.Is(err, products.ErrNotFound):
			return ErrReferenceNotFound
		case errors.Is(err, products.ErrInvalidInput):
			return ErrInvalidInput
		default:
			return fmt.Errorf("find product for image deletion: %w", err)
		}
	}
	image, err := s.images.GetByIDAndProduct(ctx, productID, imageID)
	if err != nil {
		return fmt.Errorf("find product image for deletion: %w", err)
	}
	// PostgreSQL keeps the object key if the remote operation or the following
	// database deletion fails, so the same image ID can be retried safely.
	if err := s.storage.DeleteObject(ctx, storage.AssetsBucket, image.R2ObjectKey); err != nil {
		return fmt.Errorf("%w: delete R2 object: %w", ErrStorageFailure, err)
	}
	if err := s.images.DeleteByIDAndProduct(ctx, productID, imageID); err != nil {
		return fmt.Errorf("delete product image metadata: %w", err)
	}
	return nil
}
