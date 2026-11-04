package orders

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) (*Service, error) {
	if repository == nil {
		return nil, errors.New("orders repository is required")
	}
	return &Service{repository: repository}, nil
}

func (s *Service) GetOrders(ctx context.Context) ([]Order, error) {
	result, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("get orders: %w", err)
	}
	if result == nil {
		return []Order{}, nil
	}
	return result, nil
}

func (s *Service) GetOrder(ctx context.Context, id string) (Order, error) {
	id, err := validateID(id)
	if err != nil {
		return Order{}, err
	}
	result, err := s.repository.Get(ctx, id)
	if err != nil {
		return Order{}, fmt.Errorf("get order %s: %w", id, err)
	}
	return normalizeOrder(result), nil
}

func (s *Service) CreateCheckoutOrder(
	ctx context.Context,
	cartToken string,
	customer CheckoutCustomerInput,
) (Order, error) {
	cartToken = strings.TrimSpace(cartToken)
	if len(cartToken) < 20 {
		return Order{}, fmt.Errorf("%w: cart token is required", ErrInvalidInput)
	}

	validatedCustomer, err := validateCustomer(customer)
	if err != nil {
		return Order{}, err
	}
	digest := sha256.Sum256([]byte(cartToken))
	result, err := s.repository.CreateFromCart(ctx, CreateCheckoutOrderInput{
		CartTokenHash: hex.EncodeToString(digest[:]),
		Customer:      validatedCustomer,
	})
	if err != nil {
		return Order{}, fmt.Errorf("create checkout order: %w", err)
	}
	return normalizeOrder(result), nil
}

func (s *Service) UpdateOrder(
	ctx context.Context,
	id string,
	status Status,
) (Order, error) {
	id, err := validateID(id)
	if err != nil {
		return Order{}, err
	}
	if !validStatus(status) {
		return Order{}, fmt.Errorf("%w: unsupported order status", ErrInvalidInput)
	}
	result, err := s.repository.UpdateStatus(ctx, id, status)
	if err != nil {
		return Order{}, fmt.Errorf("update order %s: %w", id, err)
	}
	return normalizeOrder(result), nil
}

func validateID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if _, err := uuid.Parse(id); err != nil {
		return "", fmt.Errorf("%w: order id must be a UUID", ErrInvalidInput)
	}
	return id, nil
}

func validateCustomer(input CheckoutCustomerInput) (CheckoutCustomerInput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return CheckoutCustomerInput{}, fmt.Errorf("%w: a valid email is required", ErrInvalidInput)
	}
	return CheckoutCustomerInput{
		Email:     email,
		FirstName: strings.TrimSpace(input.FirstName),
		LastName:  strings.TrimSpace(input.LastName),
		Phone:     strings.TrimSpace(input.Phone),
	}, nil
}

func validStatus(status Status) bool {
	switch status {
	case StatusPending, StatusFulfilled, StatusCancelled:
		return true
	default:
		return false
	}
}

func normalizeOrder(order Order) Order {
	if order.Items == nil {
		order.Items = []OrderItem{}
	}
	return order
}
