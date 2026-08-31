package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/producttypes"
)

func (s *Server) GetProductTypes(
	ctx context.Context,
	_ generated.GetProductTypesRequestObject,
) (generated.GetProductTypesResponseObject, error) {
	result, err := s.productTypes.GetProductTypes(ctx)
	if err != nil {
		return nil, err
	}

	response := make(generated.GetProductTypes200JSONResponse, 0, len(result))
	for _, productType := range result {
		apiProductType, err := toAPIProductType(productType)
		if err != nil {
			return nil, err
		}
		response = append(response, apiProductType)
	}
	return response, nil
}

func (s *Server) GetProductType(
	ctx context.Context,
	request generated.GetProductTypeRequestObject,
) (generated.GetProductTypeResponseObject, error) {
	productType, err := s.productTypes.GetProductType(ctx, request.ProductTypeId.String())
	if err != nil {
		switch {
		case errors.Is(err, producttypes.ErrInvalidInput):
			return generated.GetProductType400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, producttypes.ErrNotFound):
			return generated.GetProductType404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productTypeNotFoundError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPIProductType(productType)
	if err != nil {
		return nil, err
	}
	return generated.GetProductType200JSONResponse(response), nil
}

func (s *Server) CreateProductType(
	ctx context.Context,
	request generated.CreateProductTypeRequestObject,
) (generated.CreateProductTypeResponseObject, error) {
	if request.Body == nil {
		return generated.CreateProductType400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	productType, err := s.productTypes.CreateProductType(ctx, producttypes.CreateProductTypeInput{
		Name:        request.Body.Name,
		Slug:        request.Body.Slug,
		Description: optionalString(request.Body.Description),
	})
	if err != nil {
		switch {
		case errors.Is(err, producttypes.ErrInvalidInput):
			return generated.CreateProductType400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, producttypes.ErrConflict):
			return generated.CreateProductType409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(productTypeConflictError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPIProductType(productType)
	if err != nil {
		return nil, err
	}
	return generated.CreateProductType201JSONResponse(response), nil
}

func (s *Server) UpdateProductType(
	ctx context.Context,
	request generated.UpdateProductTypeRequestObject,
) (generated.UpdateProductTypeResponseObject, error) {
	if request.Body == nil {
		return generated.UpdateProductType400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	productType, err := s.productTypes.UpdateProductType(
		ctx,
		request.ProductTypeId.String(),
		producttypes.UpdateProductTypeInput{
			Name:        request.Body.Name,
			Slug:        request.Body.Slug,
			Description: optionalString(request.Body.Description),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, producttypes.ErrInvalidInput):
			return generated.UpdateProductType400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, producttypes.ErrNotFound):
			return generated.UpdateProductType404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productTypeNotFoundError()),
			}, nil
		case errors.Is(err, producttypes.ErrConflict):
			return generated.UpdateProductType409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(productTypeConflictError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPIProductType(productType)
	if err != nil {
		return nil, err
	}
	return generated.UpdateProductType200JSONResponse(response), nil
}

func (s *Server) DeleteProductType(
	ctx context.Context,
	request generated.DeleteProductTypeRequestObject,
) (generated.DeleteProductTypeResponseObject, error) {
	err := s.productTypes.DeleteProductType(ctx, request.ProductTypeId.String())
	if err != nil {
		switch {
		case errors.Is(err, producttypes.ErrInvalidInput):
			return generated.DeleteProductType400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, producttypes.ErrNotFound):
			return generated.DeleteProductType404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productTypeNotFoundError()),
			}, nil
		default:
			return nil, err
		}
	}

	return generated.DeleteProductType204Response{}, nil
}

func toAPIProductType(productType producttypes.ProductType) (generated.ProductType, error) {
	id, err := uuid.Parse(productType.ID)
	if err != nil {
		return generated.ProductType{}, fmt.Errorf("parse product type id: %w", err)
	}
	return generated.ProductType{
		Id:          id,
		Name:        productType.Name,
		Slug:        productType.Slug,
		Description: productType.Description,
	}, nil
}

func productTypeNotFoundError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeNotFound,
		Message: "product type not found",
	}
}

func productTypeConflictError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeConflict,
		Message: "a product type with this slug already exists",
	}
}
