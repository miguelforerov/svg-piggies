package productrelationships

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
		return nil, errors.New("product relationships repository is required")
	}
	return &Service{repository: repository}, nil
}

func (s *Service) GetProductRelationships(
	ctx context.Context,
	productID string,
) ([]ProductRelationship, error) {
	productID, err := validateID("product id", productID)
	if err != nil {
		return nil, err
	}

	relationships, err := s.repository.ListByProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("get relationships for product %s: %w", productID, err)
	}
	if relationships == nil {
		return []ProductRelationship{}, nil
	}
	return relationships, nil
}

func (s *Service) GetProductRelationship(
	ctx context.Context,
	productID string,
	relationshipID string,
) (ProductRelationship, error) {
	productID, err := validateID("product id", productID)
	if err != nil {
		return ProductRelationship{}, err
	}
	relationshipID, err = validateID("relationship id", relationshipID)
	if err != nil {
		return ProductRelationship{}, err
	}

	relationship, err := s.repository.Get(ctx, productID, relationshipID)
	if err != nil {
		return ProductRelationship{}, fmt.Errorf(
			"get relationship %s for product %s: %w",
			relationshipID,
			productID,
			err,
		)
	}
	return relationship, nil
}

func (s *Service) CreateProductRelationship(
	ctx context.Context,
	input CreateProductRelationshipInput,
) (ProductRelationship, error) {
	productID, err := validateID("product id", input.ProductID)
	if err != nil {
		return ProductRelationship{}, err
	}
	relatedProductID, err := validateID("related product id", input.RelatedProductID)
	if err != nil {
		return ProductRelationship{}, err
	}
	if err := validateRelationshipFields(productID, relatedProductID, input.DisplayOrder); err != nil {
		return ProductRelationship{}, err
	}
	input = CreateProductRelationshipInput{
		ProductID:        productID,
		RelatedProductID: relatedProductID,
		DisplayOrder:     input.DisplayOrder,
	}

	relationship, err := s.repository.Create(ctx, input)
	if err != nil {
		return ProductRelationship{}, fmt.Errorf("create product relationship: %w", err)
	}
	return relationship, nil
}

func (s *Service) CreateProductRelationships(
	ctx context.Context,
	input CreateProductRelationshipsInput,
) (ProductWithRelationships, error) {
	productID, err := validateID("product id", input.ProductID)
	if err != nil {
		return ProductWithRelationships{}, err
	}
	if len(input.RelatedProductIDs) == 0 {
		return ProductWithRelationships{}, fmt.Errorf(
			"%w: at least one related product id is required",
			ErrInvalidInput,
		)
	}

	relatedProductIDs := make([]string, 0, len(input.RelatedProductIDs))
	seen := make(map[string]struct{}, len(input.RelatedProductIDs))
	for _, candidate := range input.RelatedProductIDs {
		relatedProductID, err := validateID("related product id", candidate)
		if err != nil {
			return ProductWithRelationships{}, err
		}
		if productID == relatedProductID {
			return ProductWithRelationships{}, fmt.Errorf(
				"%w: a product cannot be related to itself",
				ErrInvalidInput,
			)
		}
		if _, exists := seen[relatedProductID]; exists {
			continue
		}
		seen[relatedProductID] = struct{}{}
		relatedProductIDs = append(relatedProductIDs, relatedProductID)
	}

	result, err := s.repository.CreateMany(ctx, CreateProductRelationshipsInput{
		ProductID:         productID,
		RelatedProductIDs: relatedProductIDs,
	})
	if err != nil {
		return ProductWithRelationships{}, fmt.Errorf(
			"create relationships for product %s: %w",
			productID,
			err,
		)
	}
	if result.Relationships == nil {
		result.Relationships = []PopulatedProductRelationship{}
	}
	return result, nil
}

func (s *Service) UpdateProductRelationship(
	ctx context.Context,
	productID string,
	relationshipID string,
	input UpdateProductRelationshipInput,
) (ProductRelationship, error) {
	productID, err := validateID("product id", productID)
	if err != nil {
		return ProductRelationship{}, err
	}
	relationshipID, err = validateID("relationship id", relationshipID)
	if err != nil {
		return ProductRelationship{}, err
	}
	relatedProductID, err := validateID("related product id", input.RelatedProductID)
	if err != nil {
		return ProductRelationship{}, err
	}
	if err := validateRelationshipFields(productID, relatedProductID, input.DisplayOrder); err != nil {
		return ProductRelationship{}, err
	}
	input = UpdateProductRelationshipInput{
		RelatedProductID: relatedProductID,
		DisplayOrder:     input.DisplayOrder,
	}

	relationship, err := s.repository.Update(ctx, productID, relationshipID, input)
	if err != nil {
		return ProductRelationship{}, fmt.Errorf(
			"update relationship %s for product %s: %w",
			relationshipID,
			productID,
			err,
		)
	}
	return relationship, nil
}

func (s *Service) ReplaceProductRelationships(
	ctx context.Context,
	input ReplaceProductRelationshipsInput,
) (ProductWithRelationships, error) {
	productID, err := validateID("product id", input.ProductID)
	if err != nil {
		return ProductWithRelationships{}, err
	}

	relatedProductIDs := make([]string, 0, len(input.RelatedProductIDs))
	seen := make(map[string]struct{}, len(input.RelatedProductIDs))
	for _, candidate := range input.RelatedProductIDs {
		relatedProductID, err := validateID("related product id", candidate)
		if err != nil {
			return ProductWithRelationships{}, err
		}
		if productID == relatedProductID {
			return ProductWithRelationships{}, fmt.Errorf(
				"%w: a product cannot be related to itself",
				ErrInvalidInput,
			)
		}
		if _, exists := seen[relatedProductID]; exists {
			return ProductWithRelationships{}, fmt.Errorf(
				"%w: related product ids must be unique",
				ErrInvalidInput,
			)
		}
		seen[relatedProductID] = struct{}{}
		relatedProductIDs = append(relatedProductIDs, relatedProductID)
	}

	result, err := s.repository.Replace(ctx, ReplaceProductRelationshipsInput{
		ProductID:         productID,
		RelatedProductIDs: relatedProductIDs,
	})
	if err != nil {
		return ProductWithRelationships{}, fmt.Errorf(
			"replace relationships for product %s: %w",
			productID,
			err,
		)
	}
	if result.Relationships == nil {
		result.Relationships = []PopulatedProductRelationship{}
	}
	return result, nil
}

func (s *Service) DeleteProductRelationship(
	ctx context.Context,
	productID string,
	relationshipID string,
) error {
	productID, err := validateID("product id", productID)
	if err != nil {
		return err
	}
	relationshipID, err = validateID("relationship id", relationshipID)
	if err != nil {
		return err
	}

	if err := s.repository.Delete(ctx, productID, relationshipID); err != nil {
		return fmt.Errorf(
			"delete relationship %s from product %s: %w",
			relationshipID,
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

func validateRelationshipFields(productID string, relatedProductID string, displayOrder int) error {
	if productID == relatedProductID {
		return fmt.Errorf("%w: a product cannot be related to itself", ErrInvalidInput)
	}
	if displayOrder < 0 {
		return fmt.Errorf("%w: display order cannot be negative", ErrInvalidInput)
	}
	return nil
}
