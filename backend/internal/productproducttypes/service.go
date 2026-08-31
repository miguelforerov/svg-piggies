package productproducttypes

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
		return nil, errors.New("product product types repository is required")
	}
	return &Service{repository: repository}, nil
}

func (s *Service) GetProductProductTypes(
	ctx context.Context,
	productID string,
) ([]ProductProductType, error) {
	productID, err := validateID("product id", productID)
	if err != nil {
		return nil, err
	}

	productProductTypes, err := s.repository.ListByProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("get product types for product %s: %w", productID, err)
	}
	if productProductTypes == nil {
		return []ProductProductType{}, nil
	}
	return productProductTypes, nil
}

func (s *Service) CreateProductProductType(
	ctx context.Context,
	input CreateProductProductTypeInput,
) (ProductProductType, error) {
	productID, err := validateID("product id", input.ProductID)
	if err != nil {
		return ProductProductType{}, err
	}
	productTypeID, err := validateID("product type id", input.ProductTypeID)
	if err != nil {
		return ProductProductType{}, err
	}
	input = CreateProductProductTypeInput{
		ProductID:     productID,
		ProductTypeID: productTypeID,
	}

	productProductType, err := s.repository.Create(ctx, input)
	if err != nil {
		return ProductProductType{}, fmt.Errorf("create product product type: %w", err)
	}
	return productProductType, nil
}

func (s *Service) DeleteProductProductType(
	ctx context.Context,
	productID string,
	productTypeID string,
) error {
	productID, err := validateID("product id", productID)
	if err != nil {
		return err
	}
	productTypeID, err = validateID("product type id", productTypeID)
	if err != nil {
		return err
	}

	if err := s.repository.Delete(ctx, productID, productTypeID); err != nil {
		return fmt.Errorf(
			"delete product type %s from product %s: %w",
			productTypeID,
			productID,
			err,
		)
	}
	return nil
}

func validateID(name string, id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("%w: %s is required", ErrInvalidInput, name)
	}
	return id, nil
}
