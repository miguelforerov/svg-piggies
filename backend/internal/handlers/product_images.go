package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
)

func (s *Server) RequestProductImageUpload(
	ctx context.Context,
	request generated.RequestProductImageUploadRequestObject,
) (generated.RequestProductImageUploadResponseObject, error) {
	if request.Body == nil {
		return generated.RequestProductImageUpload400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}
	authorization, err := s.productImageUploads.RequestUpload(ctx, productimages.RequestUploadInput{
		ProductID: request.ProductId.String(), OriginalFilename: request.Body.OriginalFilename,
		ContentType: string(request.Body.ContentType), FileSizeBytes: request.Body.FileSizeBytes,
	})
	if err != nil {
		switch {
		case errors.Is(err, productimages.ErrInvalidUpload):
			return generated.RequestProductImageUpload400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productimages.ErrReferenceNotFound):
			return generated.RequestProductImageUpload404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productImageProductNotFoundError()),
			}, nil
		case errors.Is(err, productimages.ErrStorageFailure):
			return generated.RequestProductImageUpload502JSONResponse{
				BadGatewayJSONResponse: generated.BadGatewayJSONResponse(storageUnavailableError()),
			}, nil
		case errors.Is(err, productimages.ErrUploadUnavailable):
			return generated.RequestProductImageUpload503JSONResponse{
				ServiceUnavailableJSONResponse: generated.ServiceUnavailableJSONResponse(uploadUnavailableError()),
			}, nil
		default:
			return nil, err
		}
	}
	return generated.RequestProductImageUpload201JSONResponse{
		UploadToken: authorization.UploadToken, ObjectKey: authorization.ObjectKey,
		UploadUrl: authorization.UploadURL, Headers: authorization.Headers,
		ExpiresAt: authorization.ExpiresAt,
	}, nil
}

func (s *Server) ConfirmProductImageUpload(
	ctx context.Context,
	request generated.ConfirmProductImageUploadRequestObject,
) (generated.ConfirmProductImageUploadResponseObject, error) {
	if request.Body == nil {
		return generated.ConfirmProductImageUpload400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(errors.New("request body is required"))),
		}, nil
	}
	result, err := s.productImageUploads.ConfirmUpload(ctx, productimages.ConfirmUploadInput{
		ProductID: request.ProductId.String(), UploadToken: request.Body.UploadToken,
		AltText: request.Body.AltText, DisplayOrder: optionalInt(request.Body.DisplayOrder),
		IsPrimary: optionalBool(request.Body.IsPrimary),
	})
	if err != nil {
		switch {
		case errors.Is(err, productimages.ErrInvalidUpload), errors.Is(err, productimages.ErrInvalidInput):
			return generated.ConfirmProductImageUpload400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productimages.ErrReferenceNotFound), errors.Is(err, productimages.ErrObjectNotFound):
			return generated.ConfirmProductImageUpload404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(imageUploadNotFoundError()),
			}, nil
		case errors.Is(err, productimages.ErrExpiredUpload):
			return generated.ConfirmProductImageUpload410JSONResponse{
				GoneJSONResponse: generated.GoneJSONResponse(generated.Error{
					Code: generated.ErrorCodeExpiredUpload, Message: "image upload authorization has expired",
				}),
			}, nil
		case errors.Is(err, productimages.ErrObjectMismatch), errors.Is(err, productimages.ErrInvalidImage), errors.Is(err, productimages.ErrConflict):
			return generated.ConfirmProductImageUpload409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(productImageConflictError()),
			}, nil
		case errors.Is(err, productimages.ErrStorageFailure):
			return generated.ConfirmProductImageUpload502JSONResponse{
				BadGatewayJSONResponse: generated.BadGatewayJSONResponse(storageUnavailableError()),
			}, nil
		default:
			return nil, err
		}
	}

	apiImage, err := toAPIProductImage(result)
	if err != nil {
		return nil, err
	}
	return generated.ConfirmProductImageUpload201JSONResponse(apiImage), nil
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
		Message: "image upload could not be confirmed",
	}
}

func storageUnavailableError() generated.Error {
	return generated.Error{Code: generated.ErrorCodeInternalError, Message: "image storage is unavailable"}
}

func uploadUnavailableError() generated.Error {
	return generated.Error{Code: generated.ErrorCodeInternalError, Message: "image uploads are temporarily unavailable"}
}

func imageUploadNotFoundError() generated.Error {
	return generated.Error{Code: generated.ErrorCodeNotFound, Message: "product or uploaded image not found"}
}
