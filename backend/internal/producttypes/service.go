package producttypes

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
		return nil, errors.New("product types repository is required")
	}
	return &Service{repository: repository}, nil
}

func (s *Service) GetProductTypes(ctx context.Context) ([]ProductType, error) {
	productTypes, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("get product types: %w", err)
	}
	if productTypes == nil {
		return []ProductType{}, nil
	}
	return productTypes, nil
}

func (s *Service) GetProductType(ctx context.Context, id string) (ProductType, error) {
	id, err := validateID(id)
	if err != nil {
		return ProductType{}, err
	}

	productType, err := s.repository.Get(ctx, id)
	if err != nil {
		return ProductType{}, fmt.Errorf("get product type %s: %w", id, err)
	}
	return productType, nil
}

func (s *Service) CreateProductType(
	ctx context.Context,
	input CreateProductTypeInput,
) (ProductType, error) {
	input, err := validateCreateInput(input)
	if err != nil {
		return ProductType{}, err
	}

	productType, err := s.repository.Create(ctx, input)
	if err != nil {
		return ProductType{}, fmt.Errorf("create product type: %w", err)
	}
	return productType, nil
}

func (s *Service) UpdateProductType(
	ctx context.Context,
	id string,
	input UpdateProductTypeInput,
) (ProductType, error) {
	id, err := validateID(id)
	if err != nil {
		return ProductType{}, err
	}
	input, err = validateUpdateInput(input)
	if err != nil {
		return ProductType{}, err
	}

	productType, err := s.repository.Update(ctx, id, input)
	if err != nil {
		return ProductType{}, fmt.Errorf("update product type %s: %w", id, err)
	}
	return productType, nil
}

func (s *Service) DeleteProductType(ctx context.Context, id string) error {
	id, err := validateID(id)
	if err != nil {
		return err
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete product type %s: %w", id, err)
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

func validateCreateInput(input CreateProductTypeInput) (CreateProductTypeInput, error) {
	name, slug, description, err := validateFields(input.Name, input.Slug, input.Description)
	if err != nil {
		return CreateProductTypeInput{}, err
	}
	return CreateProductTypeInput{
		Name:        name,
		Slug:        slug,
		Description: description,
	}, nil
}

func validateUpdateInput(input UpdateProductTypeInput) (UpdateProductTypeInput, error) {
	name, slug, description, err := validateFields(input.Name, input.Slug, input.Description)
	if err != nil {
		return UpdateProductTypeInput{}, err
	}
	return UpdateProductTypeInput{
		Name:        name,
		Slug:        slug,
		Description: description,
	}, nil
}

func validateFields(name, slug, description string) (string, string, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", "", "", fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "", "", "", fmt.Errorf("%w: slug is required", ErrInvalidInput)
	}

	return name, slug, strings.TrimSpace(description), nil
}
