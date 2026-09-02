package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/productrelationships"
)

func (s *Server) GetProductRelationships(
	ctx context.Context,
	request generated.GetProductRelationshipsRequestObject,
) (generated.GetProductRelationshipsResponseObject, error) {
	result, err := s.productRelationships.GetProductRelationships(ctx, request.ProductId.String())
	if err != nil {
		switch {
		case errors.Is(err, productrelationships.ErrInvalidInput):
			return generated.GetProductRelationships400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productrelationships.ErrReferenceNotFound):
			return generated.GetProductRelationships404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(
					productRelationshipReferenceNotFoundError(),
				),
			}, nil
		default:
			return nil, err
		}
	}

	response := make(generated.GetProductRelationships200JSONResponse, 0, len(result))
	for _, relationship := range result {
		apiRelationship, err := toAPIProductRelationship(relationship)
		if err != nil {
			return nil, err
		}
		response = append(response, apiRelationship)
	}
	return response, nil
}

func (s *Server) GetProductRelationship(
	ctx context.Context,
	request generated.GetProductRelationshipRequestObject,
) (generated.GetProductRelationshipResponseObject, error) {
	relationship, err := s.productRelationships.GetProductRelationship(
		ctx,
		request.ProductId.String(),
		request.RelationshipId.String(),
	)
	if err != nil {
		switch {
		case errors.Is(err, productrelationships.ErrInvalidInput):
			return generated.GetProductRelationship400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productrelationships.ErrNotFound),
			errors.Is(err, productrelationships.ErrReferenceNotFound):
			return generated.GetProductRelationship404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productRelationshipNotFoundError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPIProductRelationship(relationship)
	if err != nil {
		return nil, err
	}
	return generated.GetProductRelationship200JSONResponse(response), nil
}

func (s *Server) CreateProductRelationship(
	ctx context.Context,
	request generated.CreateProductRelationshipRequestObject,
) (generated.CreateProductRelationshipResponseObject, error) {
	if request.Body == nil {
		return generated.CreateProductRelationship400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	relationship, err := s.productRelationships.CreateProductRelationship(
		ctx,
		productrelationships.CreateProductRelationshipInput{
			ProductID:        request.ProductId.String(),
			RelatedProductID: request.Body.RelatedProductId.String(),
			DisplayOrder:     optionalInt(request.Body.DisplayOrder),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, productrelationships.ErrInvalidInput):
			return generated.CreateProductRelationship400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productrelationships.ErrReferenceNotFound):
			return generated.CreateProductRelationship404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(
					productRelationshipReferenceNotFoundError(),
				),
			}, nil
		case errors.Is(err, productrelationships.ErrConflict):
			return generated.CreateProductRelationship409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(productRelationshipConflictError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPIProductRelationship(relationship)
	if err != nil {
		return nil, err
	}
	return generated.CreateProductRelationship201JSONResponse(response), nil
}

func (s *Server) CreateProductRelationships(
	ctx context.Context,
	request generated.CreateProductRelationshipsRequestObject,
) (generated.CreateProductRelationshipsResponseObject, error) {
	if request.Body == nil {
		return generated.CreateProductRelationships400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	relatedProductIDs := make([]string, 0, len(request.Body.RelatedProductIds))
	for _, relatedProductID := range request.Body.RelatedProductIds {
		relatedProductIDs = append(relatedProductIDs, relatedProductID.String())
	}
	result, err := s.productRelationships.CreateProductRelationships(
		ctx,
		productrelationships.CreateProductRelationshipsInput{
			ProductID:         request.ProductId.String(),
			RelatedProductIDs: relatedProductIDs,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, productrelationships.ErrInvalidInput):
			return generated.CreateProductRelationships400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productrelationships.ErrNotFound),
			errors.Is(err, productrelationships.ErrReferenceNotFound):
			return generated.CreateProductRelationships404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(
					productRelationshipReferenceNotFoundError(),
				),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPIProductWithRelationships(result)
	if err != nil {
		return nil, err
	}
	return generated.CreateProductRelationships200JSONResponse(response), nil
}

func (s *Server) ReplaceProductRelationships(
	ctx context.Context,
	request generated.ReplaceProductRelationshipsRequestObject,
) (generated.ReplaceProductRelationshipsResponseObject, error) {
	if request.Body == nil {
		return generated.ReplaceProductRelationships400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	relatedProductIDs := make([]string, 0, len(request.Body.RelatedProductIds))
	for _, relatedProductID := range request.Body.RelatedProductIds {
		relatedProductIDs = append(relatedProductIDs, relatedProductID.String())
	}
	result, err := s.productRelationships.ReplaceProductRelationships(
		ctx,
		productrelationships.ReplaceProductRelationshipsInput{
			ProductID:         request.ProductId.String(),
			RelatedProductIDs: relatedProductIDs,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, productrelationships.ErrInvalidInput):
			return generated.ReplaceProductRelationships400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productrelationships.ErrNotFound),
			errors.Is(err, productrelationships.ErrReferenceNotFound):
			return generated.ReplaceProductRelationships404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(
					productRelationshipReferenceNotFoundError(),
				),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPIProductWithRelationships(result)
	if err != nil {
		return nil, err
	}
	return generated.ReplaceProductRelationships200JSONResponse(response), nil
}

func (s *Server) UpdateProductRelationship(
	ctx context.Context,
	request generated.UpdateProductRelationshipRequestObject,
) (generated.UpdateProductRelationshipResponseObject, error) {
	if request.Body == nil {
		return generated.UpdateProductRelationship400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	relationship, err := s.productRelationships.UpdateProductRelationship(
		ctx,
		request.ProductId.String(),
		request.RelationshipId.String(),
		productrelationships.UpdateProductRelationshipInput{
			RelatedProductID: request.Body.RelatedProductId.String(),
			DisplayOrder:     request.Body.DisplayOrder,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, productrelationships.ErrInvalidInput):
			return generated.UpdateProductRelationship400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productrelationships.ErrNotFound),
			errors.Is(err, productrelationships.ErrReferenceNotFound):
			return generated.UpdateProductRelationship404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productRelationshipNotFoundError()),
			}, nil
		case errors.Is(err, productrelationships.ErrConflict):
			return generated.UpdateProductRelationship409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(productRelationshipConflictError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPIProductRelationship(relationship)
	if err != nil {
		return nil, err
	}
	return generated.UpdateProductRelationship200JSONResponse(response), nil
}

func (s *Server) DeleteProductRelationship(
	ctx context.Context,
	request generated.DeleteProductRelationshipRequestObject,
) (generated.DeleteProductRelationshipResponseObject, error) {
	err := s.productRelationships.DeleteProductRelationship(
		ctx,
		request.ProductId.String(),
		request.RelationshipId.String(),
	)
	if err != nil {
		switch {
		case errors.Is(err, productrelationships.ErrInvalidInput):
			return generated.DeleteProductRelationship400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productrelationships.ErrNotFound),
			errors.Is(err, productrelationships.ErrReferenceNotFound):
			return generated.DeleteProductRelationship404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productRelationshipNotFoundError()),
			}, nil
		default:
			return nil, err
		}
	}

	return generated.DeleteProductRelationship204Response{}, nil
}

func toAPIProductRelationship(
	relationship productrelationships.ProductRelationship,
) (generated.ProductRelationship, error) {
	id, err := uuid.Parse(relationship.ID)
	if err != nil {
		return generated.ProductRelationship{}, fmt.Errorf("parse relationship id: %w", err)
	}
	productID, err := uuid.Parse(relationship.ProductID)
	if err != nil {
		return generated.ProductRelationship{}, fmt.Errorf("parse product id: %w", err)
	}
	relatedProductID, err := uuid.Parse(relationship.RelatedProductID)
	if err != nil {
		return generated.ProductRelationship{}, fmt.Errorf("parse related product id: %w", err)
	}
	return generated.ProductRelationship{
		Id:               id,
		ProductId:        productID,
		RelatedProductId: relatedProductID,
		DisplayOrder:     relationship.DisplayOrder,
	}, nil
}

func toAPIProductWithRelationships(
	result productrelationships.ProductWithRelationships,
) (generated.ProductWithRelationships, error) {
	product, err := toAPIProduct(result.Product)
	if err != nil {
		return generated.ProductWithRelationships{}, err
	}

	relationships := make([]generated.PopulatedProductRelationship, 0, len(result.Relationships))
	for _, relationship := range result.Relationships {
		relationshipID, err := uuid.Parse(relationship.RelationshipID)
		if err != nil {
			return generated.ProductWithRelationships{}, fmt.Errorf(
				"parse relationship id: %w",
				err,
			)
		}
		relatedProduct, err := toAPIProduct(relationship.Product)
		if err != nil {
			return generated.ProductWithRelationships{}, err
		}
		relationships = append(relationships, generated.PopulatedProductRelationship{
			RelationshipId: relationshipID,
			DisplayOrder:   relationship.DisplayOrder,
			Product:        relatedProduct,
		})
	}

	return generated.ProductWithRelationships{
		Product:       product,
		Relationships: relationships,
	}, nil
}

func optionalInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func productRelationshipReferenceNotFoundError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeNotFound,
		Message: "product or related product not found",
	}
}

func productRelationshipNotFoundError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeNotFound,
		Message: "product relationship not found",
	}
}

func productRelationshipConflictError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeConflict,
		Message: "this product relationship already exists",
	}
}
