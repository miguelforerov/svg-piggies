package productimages

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) (*Service, error) {
	if repository == nil {
		return nil, errors.New("product images repository is required")
	}
	return &Service{repository: repository}, nil
}

func (s *Service) AddProductImages(
	ctx context.Context,
	input AddProductImagesInput,
) ([]ProductImage, error) {
	productID := strings.TrimSpace(input.ProductID)
	if productID == "" {
		return nil, fmt.Errorf("%w: product id is required", ErrInvalidInput)
	}
	if len(input.Images) == 0 {
		return nil, fmt.Errorf("%w: at least one image is required", ErrInvalidInput)
	}

	images := make([]NewProductImage, 0, len(input.Images))
	objectKeys := make(map[string]struct{}, len(input.Images))
	primaryCount := 0
	for index, image := range input.Images {
		image.R2ObjectKey = strings.TrimSpace(image.R2ObjectKey)
		image.OriginalFilename = strings.TrimSpace(image.OriginalFilename)
		image.ContentType = strings.TrimSpace(image.ContentType)
		image.R2ETag = trimOptional(image.R2ETag)

		if err := validateImage(index, image); err != nil {
			return nil, err
		}
		if _, exists := objectKeys[image.R2ObjectKey]; exists {
			return nil, fmt.Errorf(
				"%w: images[%d].r2ObjectKey is duplicated",
				ErrInvalidInput,
				index,
			)
		}
		objectKeys[image.R2ObjectKey] = struct{}{}
		if image.IsPrimary {
			primaryCount++
		}
		images = append(images, image)
	}
	if primaryCount > 1 {
		return nil, fmt.Errorf("%w: only one image can be primary", ErrInvalidInput)
	}

	result, err := s.repository.CreateMany(ctx, AddProductImagesInput{
		ProductID: productID,
		Images:    images,
	})
	if err != nil {
		return nil, fmt.Errorf("add images to product %s: %w", productID, err)
	}
	if result == nil {
		return []ProductImage{}, nil
	}
	return result, nil
}

func validateImage(index int, image NewProductImage) error {
	fieldError := func(message string) error {
		return fmt.Errorf("%w: images[%d].%s", ErrInvalidInput, index, message)
	}

	if image.R2ObjectKey == "" {
		return fieldError("r2ObjectKey is required")
	}
	if image.OriginalFilename == "" {
		return fieldError("originalFilename is required")
	}
	if image.ContentType == "" {
		return fieldError("contentType is required")
	}
	if !strings.HasPrefix(strings.ToLower(image.ContentType), "image/") {
		return fieldError("contentType must be an image media type")
	}
	if image.FileSizeBytes < 0 {
		return fieldError("fileSizeBytes cannot be negative")
	}
	if image.WidthPX != nil && *image.WidthPX <= 0 {
		return fieldError("widthPx must be greater than zero")
	}
	if image.HeightPX != nil && *image.HeightPX <= 0 {
		return fieldError("heightPx must be greater than zero")
	}
	if image.DisplayOrder < 0 {
		return fieldError("displayOrder cannot be negative")
	}
	return nil
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
