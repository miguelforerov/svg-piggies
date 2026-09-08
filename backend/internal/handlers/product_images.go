package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
)

func (s *Server) AddProductImages(
	ctx context.Context,
	request generated.AddProductImagesRequestObject,
) (generated.AddProductImagesResponseObject, error) {
	if request.Body == nil {
		return generated.AddProductImages400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	images := make([]productimages.NewProductImage, 0, len(request.Body.Images))
	for _, image := range request.Body.Images {
		images = append(images, productimages.NewProductImage{
			R2ObjectKey:      image.R2ObjectKey,
			OriginalFilename: image.OriginalFilename,
			ContentType:      image.ContentType,
			FileSizeBytes:    image.FileSizeBytes,
			WidthPX:          image.WidthPx,
			HeightPX:         image.HeightPx,
			AltText:          image.AltText,
			DisplayOrder:     optionalInt(image.DisplayOrder),
			IsPrimary:        optionalBool(image.IsPrimary),
			R2ETag:           image.R2Etag,
		})
	}

	result, err := s.productImages.AddProductImages(ctx, productimages.AddProductImagesInput{
		ProductID: request.ProductId.String(),
		Images:    images,
	})
	if err != nil {
		switch {
		case errors.Is(err, productimages.ErrInvalidInput):
			return generated.AddProductImages400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productimages.ErrReferenceNotFound):
			return generated.AddProductImages404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productImageProductNotFoundError()),
			}, nil
		case errors.Is(err, productimages.ErrConflict):
			return generated.AddProductImages409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(productImageConflictError()),
			}, nil
		default:
			return nil, err
		}
	}

	response := make(generated.AddProductImages201JSONResponse, 0, len(result))
	for _, image := range result {
		apiImage, err := toAPIProductImage(image)
		if err != nil {
			return nil, err
		}
		response = append(response, apiImage)
	}
	return response, nil
}

func toAPIProductImage(image productimages.ProductImage) (generated.ProductImage, error) {
	id, err := uuid.Parse(image.ID)
	if err != nil {
		return generated.ProductImage{}, fmt.Errorf("parse product image id: %w", err)
	}
	productID, err := uuid.Parse(image.ProductID)
	if err != nil {
		return generated.ProductImage{}, fmt.Errorf("parse product id: %w", err)
	}
	return generated.ProductImage{
		Id:               id,
		ProductId:        productID,
		R2ObjectKey:      image.R2ObjectKey,
		OriginalFilename: image.OriginalFilename,
		ContentType:      image.ContentType,
		FileSizeBytes:    image.FileSizeBytes,
		WidthPx:          image.WidthPX,
		HeightPx:         image.HeightPX,
		AltText:          image.AltText,
		DisplayOrder:     image.DisplayOrder,
		IsPrimary:        image.IsPrimary,
		R2Etag:           image.R2ETag,
		CreatedAt:        &image.CreatedAt,
		UpdatedAt:        &image.UpdatedAt,
	}, nil
}

func optionalBool(value *bool) bool {
	return value != nil && *value
}

func productImageProductNotFoundError() generated.Error {
	return generated.Error{Code: generated.ErrorCodeNotFound, Message: "product not found"}
}

func productImageConflictError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeConflict,
		Message: "an image with this R2 object key already exists",
	}
}
