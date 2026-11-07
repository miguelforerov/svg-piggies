package productimages

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	"github.com/zdenaforero/svg-piggies/backend/internal/storage"
)

const ImageURLTTL = 5 * time.Minute

type ProductImageReader interface {
	GetByProductID(context.Context, string) ([]ProductImage, error)
}

type ImageURLSigner interface {
	PresignGet(context.Context, storage.Bucket, string, time.Duration) (storage.PresignedRequest, error)
}

type ProductImageWithURL struct {
	ProductImage
	ImageURL string
}

type ReadService struct {
	products ProductLookup
	images   ProductImageReader
	signer   ImageURLSigner
}

func NewReadService(products ProductLookup, images ProductImageReader, signer ImageURLSigner) (*ReadService, error) {
	if products == nil || images == nil || signer == nil {
		return nil, errors.New("product lookup, image reader, and URL signer are required")
	}
	return &ReadService{products: products, images: images, signer: signer}, nil
}

func (s *ReadService) GetProductImages(ctx context.Context, productID string) ([]ProductImageWithURL, error) {
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return nil, fmt.Errorf("%w: product id is required", ErrInvalidInput)
	}
	if _, err := s.products.GetProduct(ctx, productID); err != nil {
		switch {
		case errors.Is(err, products.ErrNotFound):
			return nil, ErrReferenceNotFound
		case errors.Is(err, products.ErrInvalidInput):
			return nil, ErrInvalidInput
		default:
			return nil, fmt.Errorf("find product for image read: %w", err)
		}
	}

	images, err := s.images.GetByProductID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("get product images: %w", err)
	}
	result := make([]ProductImageWithURL, 0, len(images))
	for _, image := range images {
		request, err := s.signer.PresignGet(ctx, storage.AssetsBucket, image.R2ObjectKey, ImageURLTTL)
		if err != nil {
			return nil, fmt.Errorf("%w: presign product image: %w", ErrStorageFailure, err)
		}
		result = append(result, ProductImageWithURL{ProductImage: image, ImageURL: request.URL})
	}
	return result, nil
}
