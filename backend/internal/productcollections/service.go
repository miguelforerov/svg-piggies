package productcollections

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
		return nil, errors.New("product collections repository is required")
	}
	return &Service{repository: repository}, nil
}

func (s *Service) GetProductCollections(
	ctx context.Context,
	productID string,
) ([]ProductCollection, error) {
	productID, err := validateID("product id", productID)
	if err != nil {
		return nil, err
	}

	productCollections, err := s.repository.ListByProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("get collections for product %s: %w", productID, err)
	}
	if productCollections == nil {
		return []ProductCollection{}, nil
	}
	return productCollections, nil
}

func (s *Service) CreateProductCollection(
	ctx context.Context,
	input CreateProductCollectionInput,
) (ProductCollection, error) {
	productID, err := validateID("product id", input.ProductID)
	if err != nil {
		return ProductCollection{}, err
	}
	collectionID, err := validateID("collection id", input.CollectionID)
	if err != nil {
		return ProductCollection{}, err
	}
	input = CreateProductCollectionInput{
		ProductID:    productID,
		CollectionID: collectionID,
	}

	productCollection, err := s.repository.Create(ctx, input)
	if err != nil {
		return ProductCollection{}, fmt.Errorf("create product collection: %w", err)
	}
	return productCollection, nil
}

func (s *Service) DeleteProductCollection(
	ctx context.Context,
	productID string,
	collectionID string,
) error {
	productID, err := validateID("product id", productID)
	if err != nil {
		return err
	}
	collectionID, err = validateID("collection id", collectionID)
	if err != nil {
		return err
	}

	if err := s.repository.Delete(ctx, productID, collectionID); err != nil {
		return fmt.Errorf(
			"delete collection %s from product %s: %w",
			collectionID,
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
