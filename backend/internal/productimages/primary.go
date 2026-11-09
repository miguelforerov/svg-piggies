package productimages

import (
	"context"
	"fmt"
	"strings"
)

type PrimaryImageRepository interface {
	SetPrimary(context.Context, string, string) error
}

type PrimaryService struct {
	images PrimaryImageRepository
}

func NewPrimaryService(images PrimaryImageRepository) (*PrimaryService, error) {
	if images == nil {
		return nil, fmt.Errorf("primary image repository is required")
	}
	return &PrimaryService{images: images}, nil
}

func (s *PrimaryService) SetPrimaryImage(ctx context.Context, productID, imageID string) error {
	productID, imageID = strings.TrimSpace(productID), strings.TrimSpace(imageID)
	if productID == "" || imageID == "" {
		return fmt.Errorf("%w: product id and image id are required", ErrInvalidInput)
	}
	if err := s.images.SetPrimary(ctx, productID, imageID); err != nil {
		return fmt.Errorf("set primary product image: %w", err)
	}
	return nil
}
