package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/productrelationships"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
)

func (s *Server) GetPublicProductRelationships(
	ctx context.Context,
	request generated.GetPublicProductRelationshipsRequestObject,
) (generated.GetPublicProductRelationshipsResponseObject, error) {
	product, err := s.products.GetProductBySlug(ctx, request.Slug)
	if err != nil {
		if errors.Is(err, products.ErrNotFound) {
			return generated.GetPublicProductRelationships404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productNotFoundError()),
			}, nil
		}
		return nil, err
	}
	if product.Status != products.StatusActive {
		return generated.GetPublicProductRelationships404JSONResponse{
			NotFoundJSONResponse: generated.NotFoundJSONResponse(productNotFoundError()),
		}, nil
	}

	relationships, err := s.productRelationships.GetProductRelationships(ctx, product.ID)
	if err != nil {
		if errors.Is(err, productrelationships.ErrReferenceNotFound) {
			return generated.GetPublicProductRelationships404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productNotFoundError()),
			}, nil
		}
		return nil, err
	}

	response := make(generated.GetPublicProductRelationships200JSONResponse, 0, len(relationships))
	for _, relationship := range relationships {
		related, err := s.products.GetProduct(ctx, relationship.RelatedProductID)
		if errors.Is(err, products.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if related.Status != products.StatusActive {
			continue
		}
		id, err := uuid.Parse(related.ID)
		if err != nil {
			return nil, err
		}
		response = append(response, id)
	}
	return response, nil
}
