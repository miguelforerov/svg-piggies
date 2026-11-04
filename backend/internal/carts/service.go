package carts

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const tokenByteLength = 32

type Service struct {
	repository Repository
}

func NewService(repository Repository) (*Service, error) {
	if repository == nil {
		return nil, errors.New("carts repository is required")
	}
	return &Service{repository: repository}, nil
}

func (s *Service) CreateCart(ctx context.Context) (CreatedCart, error) {
	tokenBytes := make([]byte, tokenByteLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return CreatedCart{}, fmt.Errorf("generate cart token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	cart, err := s.repository.CreateGuest(ctx, hashToken(token))
	if err != nil {
		return CreatedCart{}, fmt.Errorf("create cart: %w", err)
	}
	return CreatedCart{Token: token, Cart: normalizeCart(cart)}, nil
}

func (s *Service) GetCart(ctx context.Context, token string) (Cart, error) {
	token, err := validateToken(token)
	if err != nil {
		return Cart{}, err
	}

	cart, err := s.repository.GetByTokenHash(ctx, hashToken(token))
	if err != nil {
		return Cart{}, fmt.Errorf("get cart: %w", err)
	}
	return normalizeCart(cart), nil
}

func (s *Service) AddItem(
	ctx context.Context,
	token string,
	input AddItemInput,
) (Cart, error) {
	token, err := validateToken(token)
	if err != nil {
		return Cart{}, err
	}
	productID, err := validateProductID(input.ProductID)
	if err != nil {
		return Cart{}, err
	}
	quantity, err := validateQuantity(input.Quantity)
	if err != nil {
		return Cart{}, err
	}

	cart, err := s.repository.AddItem(ctx, hashToken(token), productID, quantity)
	if err != nil {
		return Cart{}, fmt.Errorf("add cart item: %w", err)
	}
	return normalizeCart(cart), nil
}

func (s *Service) UpdateItem(
	ctx context.Context,
	token string,
	productID string,
	quantity int,
) (Cart, error) {
	token, err := validateToken(token)
	if err != nil {
		return Cart{}, err
	}
	productID, err = validateProductID(productID)
	if err != nil {
		return Cart{}, err
	}
	quantity, err = validateQuantity(quantity)
	if err != nil {
		return Cart{}, err
	}

	cart, err := s.repository.UpdateItem(ctx, hashToken(token), productID, quantity)
	if err != nil {
		return Cart{}, fmt.Errorf("update cart item: %w", err)
	}
	return normalizeCart(cart), nil
}

func (s *Service) DeleteItem(
	ctx context.Context,
	token string,
	productID string,
) (Cart, error) {
	token, err := validateToken(token)
	if err != nil {
		return Cart{}, err
	}
	productID, err = validateProductID(productID)
	if err != nil {
		return Cart{}, err
	}

	cart, err := s.repository.DeleteItem(ctx, hashToken(token), productID)
	if err != nil {
		return Cart{}, fmt.Errorf("delete cart item: %w", err)
	}
	return normalizeCart(cart), nil
}

func (s *Service) ClearCart(ctx context.Context, token string) error {
	token, err := validateToken(token)
	if err != nil {
		return err
	}
	if err := s.repository.Clear(ctx, hashToken(token)); err != nil {
		return fmt.Errorf("clear cart: %w", err)
	}
	return nil
}

func validateToken(token string) (string, error) {
	token = strings.TrimSpace(token)
	if len(token) < 20 {
		return "", fmt.Errorf("%w: cart token is required", ErrInvalidInput)
	}
	return token, nil
}

func validateProductID(productID string) (string, error) {
	productID = strings.TrimSpace(productID)
	if _, err := uuid.Parse(productID); err != nil {
		return "", fmt.Errorf("%w: product id must be a UUID", ErrInvalidInput)
	}
	return productID, nil
}

func validateQuantity(quantity int) (int, error) {
	if quantity < 1 || quantity > MaxQuantity {
		return 0, fmt.Errorf(
			"%w: quantity must be between 1 and %d",
			ErrInvalidInput,
			MaxQuantity,
		)
	}
	return quantity, nil
}

func hashToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func normalizeCart(cart Cart) Cart {
	if cart.Items == nil {
		cart.Items = []CartItem{}
	}
	return cart
}
