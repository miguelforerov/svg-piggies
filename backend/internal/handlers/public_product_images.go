package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
)

func (s *Server) GetPublicProductImages(ctx context.Context, request generated.GetPublicProductImagesRequestObject) (generated.GetPublicProductImagesResponseObject, error) {
	product, err := s.products.GetProductBySlug(ctx, request.Slug)
	if err != nil {
		if errors.Is(err, products.ErrNotFound) {
			return generated.GetPublicProductImages404JSONResponse{NotFoundJSONResponse: generated.NotFoundJSONResponse(productNotFoundError())}, nil
		}
		return nil, err
	}
	if product.Status != products.StatusActive {
		return generated.GetPublicProductImages404JSONResponse{NotFoundJSONResponse: generated.NotFoundJSONResponse(productNotFoundError())}, nil
	}
	images, err := s.productImageReads.GetProductImages(ctx, product.ID)
	if err != nil {
		if errors.Is(err, productimages.ErrStorageFailure) {
			return generated.GetPublicProductImages502JSONResponse{BadGatewayJSONResponse: generated.BadGatewayJSONResponse(storageUnavailableError())}, nil
		}
		return nil, err
	}
	response := make(generated.GetPublicProductImages200JSONResponse, 0, len(images))
	for _, image := range images {
		id, err := uuid.Parse(image.ID)
		if err != nil {
			return nil, err
		}
		response = append(response, generated.PublicProductImage{
			Id: id, ImageUrl: image.ImageURL, AltText: image.AltText,
			WidthPx: image.WidthPX, HeightPx: image.HeightPX,
			DisplayOrder: image.DisplayOrder, IsPrimary: image.IsPrimary,
		})
	}
	return response, nil
}
