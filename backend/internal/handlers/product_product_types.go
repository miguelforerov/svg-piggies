package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/productproducttypes"
)

func (s *Server) GetProductProductTypes(
	ctx context.Context,
	request generated.GetProductProductTypesRequestObject,
) (generated.GetProductProductTypesResponseObject, error) {
	result, err := s.productProductTypes.GetProductProductTypes(
		ctx,
		request.ProductId.String(),
	)
	if err != nil {
		switch {
		case errors.Is(err, productproducttypes.ErrInvalidInput):
			return generated.GetProductProductTypes400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productproducttypes.ErrReferenceNotFound):
			return generated.GetProductProductTypes404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(
					productProductTypeReferenceNotFoundError(),
				),
			}, nil
		default:
			return nil, err
		}
	}

	response := make(generated.GetProductProductTypes200JSONResponse, 0, len(result))
	for _, productProductType := range result {
		apiProductProductType, err := toAPIProductProductType(productProductType)
		if err != nil {
			return nil, err
		}
		response = append(response, apiProductProductType)
	}
	return response, nil
}

func (s *Server) CreateProductProductType(
	ctx context.Context,
	request generated.CreateProductProductTypeRequestObject,
) (generated.CreateProductProductTypeResponseObject, error) {
	if request.Body == nil {
		return generated.CreateProductProductType400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	productProductType, err := s.productProductTypes.CreateProductProductType(
		ctx,
		productproducttypes.CreateProductProductTypeInput{
			ProductID:     request.ProductId.String(),
			ProductTypeID: request.Body.ProductTypeId.String(),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, productproducttypes.ErrInvalidInput):
			return generated.CreateProductProductType400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productproducttypes.ErrReferenceNotFound):
			return generated.CreateProductProductType404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(
					productProductTypeReferenceNotFoundError(),
				),
			}, nil
		case errors.Is(err, productproducttypes.ErrConflict):
			return generated.CreateProductProductType409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(productProductTypeConflictError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPIProductProductType(productProductType)
	if err != nil {
		return nil, err
	}
	return generated.CreateProductProductType201JSONResponse(response), nil
}

func (s *Server) DeleteProductProductType(
	ctx context.Context,
	request generated.DeleteProductProductTypeRequestObject,
) (generated.DeleteProductProductTypeResponseObject, error) {
	err := s.productProductTypes.DeleteProductProductType(
		ctx,
		request.ProductId.String(),
		request.ProductTypeId.String(),
	)
	if err != nil {
		switch {
		case errors.Is(err, productproducttypes.ErrInvalidInput):
			return generated.DeleteProductProductType400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productproducttypes.ErrNotFound),
			errors.Is(err, productproducttypes.ErrReferenceNotFound):
			return generated.DeleteProductProductType404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productProductTypeNotFoundError()),
			}, nil
		default:
			return nil, err
		}
	}

	return generated.DeleteProductProductType204Response{}, nil
}

func toAPIProductProductType(
	productProductType productproducttypes.ProductProductType,
) (generated.ProductProductType, error) {
	productID, err := uuid.Parse(productProductType.ProductID)
	if err != nil {
		return generated.ProductProductType{}, fmt.Errorf("parse product id: %w", err)
	}
	productTypeID, err := uuid.Parse(productProductType.ProductTypeID)
	if err != nil {
		return generated.ProductProductType{}, fmt.Errorf("parse product type id: %w", err)
	}
	return generated.ProductProductType{
		ProductId:     productID,
		ProductTypeId: productTypeID,
	}, nil
}

func productProductTypeReferenceNotFoundError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeNotFound,
		Message: "product or product type not found",
	}
}

func productProductTypeNotFoundError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeNotFound,
		Message: "product type assignment not found",
	}
}

func productProductTypeConflictError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeConflict,
		Message: "product already has this product type",
	}
}
