package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
)

func (s *Server) ReorderProductImages(ctx context.Context, request generated.ReorderProductImagesRequestObject) (generated.ReorderProductImagesResponseObject, error) {
	if request.Body == nil || request.Body.ImageIds == nil {
		return generated.ReorderProductImages400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(generated.Error{Code: generated.ErrorCodeInvalidRequest, Message: "imageIds is required"})}, nil
	}
	ids := make([]string, len(request.Body.ImageIds))
	for i, id := range request.Body.ImageIds {
		ids[i] = id.String()
	}
	err := s.productImageOrder.ReorderImages(ctx, request.ProductId.String(), ids)
	if err != nil {
		if errors.Is(err, productimages.ErrInvalidInput) {
			return generated.ReorderProductImages400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(generated.Error{Code: generated.ErrorCodeInvalidRequest, Message: "imageIds must contain every current image exactly once"})}, nil
		}
		if errors.Is(err, productimages.ErrReferenceNotFound) {
			return generated.ReorderProductImages404JSONResponse{NotFoundJSONResponse: generated.NotFoundJSONResponse(productImageProductNotFoundError())}, nil
		}
		return nil, err
	}
	return generated.ReorderProductImages204Response{}, nil
}

func (s *Server) UpdateProductImageAltText(
	ctx context.Context,
	request generated.UpdateProductImageAltTextRequestObject,
) (generated.UpdateProductImageAltTextResponseObject, error) {
	if request.Body == nil {
		return generated.UpdateProductImageAltText400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(generated.Error{
			Code: generated.ErrorCodeInvalidRequest, Message: "request body is required",
		})}, nil
	}
	input := productimages.AltTextUpdate{ProductID: request.ProductId.String(), ImageID: request.ImageId.String()}
	raw := request.Body.AltText
	if len(raw) > 0 {
		input.Present = true
		if !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			var value string
			if err := json.Unmarshal(raw, &value); err != nil {
				return generated.UpdateProductImageAltText400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(generated.Error{
					Code: generated.ErrorCodeInvalidRequest, Message: "altText must be a string or null",
				})}, nil
			}
			input.AltText = &value
		}
	}
	if err := s.productImageAltText.UpdateAltText(ctx, input); err != nil {
		switch {
		case errors.Is(err, productimages.ErrInvalidInput):
			return generated.UpdateProductImageAltText400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(generated.Error{
				Code: generated.ErrorCodeInvalidRequest, Message: "alternative text is invalid or exceeds 500 characters",
			})}, nil
		case errors.Is(err, productimages.ErrReferenceNotFound), errors.Is(err, productimages.ErrImageNotFound):
			return generated.UpdateProductImageAltText404JSONResponse{NotFoundJSONResponse: generated.NotFoundJSONResponse(generated.Error{
				Code: generated.ErrorCodeNotFound, Message: "product image not found",
			})}, nil
		default:
			return nil, err
		}
	}
	return generated.UpdateProductImageAltText204Response{}, nil
}

func (s *Server) GetProductImages(
	ctx context.Context,
	request generated.GetProductImagesRequestObject,
) (generated.GetProductImagesResponseObject, error) {
	images, err := s.productImageReads.GetProductImages(ctx, request.ProductId.String())
	if err != nil {
		switch {
		case errors.Is(err, productimages.ErrInvalidInput):
			return generated.GetProductImages400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productimages.ErrReferenceNotFound):
			return generated.GetProductImages404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productImageProductNotFoundError()),
			}, nil
		case errors.Is(err, productimages.ErrStorageFailure):
			return generated.GetProductImages502JSONResponse{
				BadGatewayJSONResponse: generated.BadGatewayJSONResponse(storageUnavailableError()),
			}, nil
		default:
			return nil, err
		}
	}
	response := make(generated.GetProductImages200JSONResponse, 0, len(images))
	for _, image := range images {
		apiImage, err := toAPIAdminProductImage(image)
		if err != nil {
			return nil, err
		}
		response = append(response, apiImage)
	}
	return response, nil
}

func (s *Server) DeleteProductImage(
	ctx context.Context,
	request generated.DeleteProductImageRequestObject,
) (generated.DeleteProductImageResponseObject, error) {
	err := s.productImageDeletes.DeleteProductImage(ctx, request.ProductId.String(), request.ImageId.String())
	if err != nil {
		switch {
		case errors.Is(err, productimages.ErrInvalidInput):
			return generated.DeleteProductImage400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(generated.Error{
				Code: generated.ErrorCodeInvalidRequest, Message: "invalid product image id",
			})}, nil
		case errors.Is(err, productimages.ErrReferenceNotFound), errors.Is(err, productimages.ErrImageNotFound):
			return generated.DeleteProductImage404JSONResponse{NotFoundJSONResponse: generated.NotFoundJSONResponse(generated.Error{
				Code: generated.ErrorCodeNotFound, Message: "product image not found",
			})}, nil
		case errors.Is(err, productimages.ErrStorageFailure):
			return generated.DeleteProductImage502JSONResponse{BadGatewayJSONResponse: generated.BadGatewayJSONResponse(storageUnavailableError())}, nil
		default:
			return nil, err
		}
	}
	return generated.DeleteProductImage204Response{}, nil
}

func (s *Server) SetPrimaryProductImage(
	ctx context.Context,
	request generated.SetPrimaryProductImageRequestObject,
) (generated.SetPrimaryProductImageResponseObject, error) {
	err := s.productImagePrimary.SetPrimaryImage(ctx, request.ProductId.String(), request.ImageId.String())
	if err != nil {
		switch {
		case errors.Is(err, productimages.ErrInvalidInput):
			return generated.SetPrimaryProductImage400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(generated.Error{
				Code: generated.ErrorCodeInvalidRequest, Message: "invalid product image id",
			})}, nil
		case errors.Is(err, productimages.ErrReferenceNotFound), errors.Is(err, productimages.ErrImageNotFound):
			return generated.SetPrimaryProductImage404JSONResponse{NotFoundJSONResponse: generated.NotFoundJSONResponse(generated.Error{
				Code: generated.ErrorCodeNotFound, Message: "product image not found",
			})}, nil
		default:
			return nil, err
		}
	}
	return generated.SetPrimaryProductImage204Response{}, nil
}

func toAPIAdminProductImage(image productimages.ProductImageWithURL) (generated.AdminProductImage, error) {
	id, err := uuid.Parse(image.ID)
	if err != nil {
		return generated.AdminProductImage{}, fmt.Errorf("parse product image id: %w", err)
	}
	productID, err := uuid.Parse(image.ProductID)
	if err != nil {
		return generated.AdminProductImage{}, fmt.Errorf("parse product id: %w", err)
	}
	return generated.AdminProductImage{
		Id: id, ProductId: productID, OriginalFilename: image.OriginalFilename,
		ContentType: image.ContentType, FileSizeBytes: image.FileSizeBytes,
		WidthPx: image.WidthPX, HeightPx: image.HeightPX, AltText: image.AltText,
		DisplayOrder: image.DisplayOrder, IsPrimary: image.IsPrimary, R2Etag: image.R2ETag,
		ImageUrl: image.ImageURL, CreatedAt: &image.CreatedAt, UpdatedAt: &image.UpdatedAt,
	}, nil
}

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
