package productimages

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const MaxAltTextRunes = 500

type AltTextUpdate struct {
	ProductID string
	ImageID   string
	Present   bool
	AltText   *string
}

type AltTextRepository interface {
	UpdateAltText(context.Context, AltTextUpdate) error
}

type AltTextService struct {
	images AltTextRepository
}

func NewAltTextService(images AltTextRepository) (*AltTextService, error) {
	if images == nil {
		return nil, errors.New("alternative text repository is required")
	}
	return &AltTextService{images: images}, nil
}

func (s *AltTextService) UpdateAltText(ctx context.Context, input AltTextUpdate) error {
	input.ProductID = strings.TrimSpace(input.ProductID)
	input.ImageID = strings.TrimSpace(input.ImageID)
	if input.ProductID == "" || input.ImageID == "" {
		return fmt.Errorf("%w: product id and image id are required", ErrInvalidInput)
	}
	if input.Present && input.AltText != nil {
		text := strings.TrimSpace(*input.AltText)
		if utf8.RuneCountInString(text) > MaxAltTextRunes {
			return fmt.Errorf("%w: alternative text exceeds %d characters", ErrInvalidInput, MaxAltTextRunes)
		}
		input.AltText = &text
	}
	if err := s.images.UpdateAltText(ctx, input); err != nil {
		return fmt.Errorf("update product image alternative text: %w", err)
	}
	return nil
}
