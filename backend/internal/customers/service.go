package customers

import (
	"context"
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
		return nil, errors.New("customers repository is required")
	}
	return &Service{repository: repository}, nil
}

func (s *Service) GetCustomers(ctx context.Context) ([]Customer, error) {
	result, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("get customers: %w", err)
	}
	if result == nil {
		return []Customer{}, nil
	}
	return result, nil
}

func (s *Service) GetCustomer(ctx context.Context, id string) (Customer, error) {
	id, err := validateID(id)
	if err != nil {
		return Customer{}, err
	}
	result, err := s.repository.Get(ctx, id)
	if err != nil {
		return Customer{}, fmt.Errorf("get customer %s: %w", id, err)
	}
	return result, nil
}

func (s *Service) CreateCustomer(
	ctx context.Context,
	input CreateCustomerInput,
) (Customer, error) {
	input, err := validateInput(input)
	if err != nil {
		return Customer{}, err
	}
	result, err := s.repository.Create(ctx, input)
	if err != nil {
		return Customer{}, fmt.Errorf("create customer: %w", err)
	}
	return result, nil
}

func (s *Service) UpdateCustomer(
	ctx context.Context,
	id string,
	input UpdateCustomerInput,
) (Customer, error) {
	id, err := validateID(id)
	if err != nil {
		return Customer{}, err
	}
	input, err = validateInput(input)
	if err != nil {
		return Customer{}, err
	}
	result, err := s.repository.Update(ctx, id, input)
	if err != nil {
		return Customer{}, fmt.Errorf("update customer %s: %w", id, err)
	}
	return result, nil
}

func (s *Service) DeleteCustomer(ctx context.Context, id string) error {
	id, err := validateID(id)
	if err != nil {
		return err
	}
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete customer %s: %w", id, err)
	}
	return nil
}

func validateID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if _, err := uuid.Parse(id); err != nil {
		return "", fmt.Errorf("%w: customer id must be a UUID", ErrInvalidInput)
	}
	return id, nil
}

func validateInput(input CreateCustomerInput) (CreateCustomerInput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return CreateCustomerInput{}, fmt.Errorf("%w: a valid email is required", ErrInvalidInput)
	}
	return CreateCustomerInput{
		Email:            email,
		FirstName:        strings.TrimSpace(input.FirstName),
		LastName:         strings.TrimSpace(input.LastName),
		Phone:            strings.TrimSpace(input.Phone),
		MarketingConsent: input.MarketingConsent,
	}, nil
}
