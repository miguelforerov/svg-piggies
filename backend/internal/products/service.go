package products

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var pricePattern = regexp.MustCompile(`^(0|[0-9]{1,10})(\.[0-9]{1,2})?$`)

type Service struct {
	repository Repository
}

func NewService(repository Repository) (*Service, error) {
	if repository == nil {
		return nil, errors.New("products repository is required")
	}
	return &Service{repository: repository}, nil
}

func (s *Service) GetProducts(ctx context.Context) ([]Product, error) {
	products, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("get products: %w", err)
	}
	if products == nil {
		return []Product{}, nil
	}
	return products, nil
}

func (s *Service) GetProduct(ctx context.Context, id string) (Product, error) {
	id, err := validateID(id)
	if err != nil {
		return Product{}, err
	}

	product, err := s.repository.Get(ctx, id)
	if err != nil {
		return Product{}, fmt.Errorf("get product %s: %w", id, err)
	}
	return product, nil
}

func (s *Service) CreateProduct(
	ctx context.Context,
	input CreateProductInput,
) (Product, error) {
	input, err := validateCreateInput(input)
	if err != nil {
		return Product{}, err
	}

	product, err := s.repository.Create(ctx, input)
	if err != nil {
		return Product{}, fmt.Errorf("create product: %w", err)
	}
	return product, nil
}

func (s *Service) UpdateProduct(
	ctx context.Context,
	id string,
	input UpdateProductInput,
) (Product, error) {
	id, err := validateID(id)
	if err != nil {
		return Product{}, err
	}
	input, err = validateUpdateInput(input)
	if err != nil {
		return Product{}, err
	}

	product, err := s.repository.Update(ctx, id, input)
	if err != nil {
		return Product{}, fmt.Errorf("update product %s: %w", id, err)
	}
	return product, nil
}

func (s *Service) DeleteProduct(ctx context.Context, id string) error {
	id, err := validateID(id)
	if err != nil {
		return err
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete product %s: %w", id, err)
	}
	return nil
}

func validateID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	return id, nil
}

func validateCreateInput(input CreateProductInput) (CreateProductInput, error) {
	status := input.Status
	if status == "" {
		status = StatusDraft
	}

	title, slug, description, price, status, err := validateFields(
		input.Title,
		input.Slug,
		input.Description,
		input.Price,
		status,
	)
	if err != nil {
		return CreateProductInput{}, err
	}
	return CreateProductInput{
		Title:       title,
		Slug:        slug,
		Description: description,
		Price:       price,
		Status:      status,
	}, nil
}

func validateUpdateInput(input UpdateProductInput) (UpdateProductInput, error) {
	title, slug, description, price, status, err := validateFields(
		input.Title,
		input.Slug,
		input.Description,
		input.Price,
		input.Status,
	)
	if err != nil {
		return UpdateProductInput{}, err
	}
	return UpdateProductInput{
		Title:       title,
		Slug:        slug,
		Description: description,
		Price:       price,
		Status:      status,
	}, nil
}

func validateFields(
	title string,
	slug string,
	description string,
	price string,
	status Status,
) (string, string, string, string, Status, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", "", "", "", "", fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "", "", "", "", "", fmt.Errorf("%w: slug is required", ErrInvalidInput)
	}

	price = strings.TrimSpace(price)
	if !pricePattern.MatchString(price) {
		return "", "", "", "", "", fmt.Errorf(
			"%w: price must be a non-negative decimal with at most 10 integer and 2 fractional digits",
			ErrInvalidInput,
		)
	}

	if !status.Valid() {
		return "", "", "", "", "", fmt.Errorf(
			"%w: status must be draft, active, or archived",
			ErrInvalidInput,
		)
	}

	return title, slug, strings.TrimSpace(description), price, status, nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusDraft, StatusActive, StatusArchived:
		return true
	default:
		return false
	}
}
