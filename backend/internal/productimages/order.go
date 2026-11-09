package productimages

import (
	"context"
	"fmt"
	"strings"
)

type OrderRepository interface {
	Reorder(context.Context, string, []string) error
}

type OrderService struct{ images OrderRepository }

func NewOrderService(images OrderRepository) (*OrderService, error) {
	if images == nil {
		return nil, fmt.Errorf("order repository is required")
	}
	return &OrderService{images: images}, nil
}

func (s *OrderService) ReorderImages(ctx context.Context, productID string, imageIDs []string) error {
	productID = strings.TrimSpace(productID)
	if productID == "" || imageIDs == nil {
		return fmt.Errorf("%w: product id and imageIds are required", ErrInvalidInput)
	}
	seen := make(map[string]struct{}, len(imageIDs))
	for _, id := range imageIDs {
		if id == "" {
			return fmt.Errorf("%w: image id is required", ErrInvalidInput)
		}
		if _, ok := seen[id]; ok {
			return fmt.Errorf("%w: duplicate image id", ErrInvalidInput)
		}
		seen[id] = struct{}{}
	}
	if err := s.images.Reorder(ctx, productID, imageIDs); err != nil {
		return fmt.Errorf("reorder product images: %w", err)
	}
	return nil
}
