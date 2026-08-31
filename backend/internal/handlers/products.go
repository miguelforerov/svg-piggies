package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
)

func (s *Server) GetProducts(
	ctx context.Context,
	_ generated.GetProductsRequestObject,
) (generated.GetProductsResponseObject, error) {
	result, err := s.products.GetProducts(ctx)
	if err != nil {
		return nil, err
	}

	response := make(generated.GetProducts200JSONResponse, 0, len(result))
	for _, product := range result {
		apiProduct, err := toAPIProduct(product)
		if err != nil {
			return nil, err
		}
		response = append(response, apiProduct)
	}
	return response, nil
}

func (s *Server) GetProduct(
	ctx context.Context,
	request generated.GetProductRequestObject,
) (generated.GetProductResponseObject, error) {
	product, err := s.products.GetProduct(ctx, request.ProductId.String())
	if err != nil {
		switch {
		case errors.Is(err, products.ErrInvalidInput):
			return generated.GetProduct400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, products.ErrNotFound):
			return generated.GetProduct404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productNotFoundError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPIProduct(product)
	if err != nil {
		return nil, err
	}
	return generated.GetProduct200JSONResponse(response), nil
}

func (s *Server) GetProductBySlug(
	ctx context.Context,
	request generated.GetProductBySlugRequestObject,
) (generated.GetProductBySlugResponseObject, error) {
	product, err := s.products.GetProductBySlug(ctx, request.Slug)
	if err != nil {
		if errors.Is(err, products.ErrNotFound) {
			return generated.GetProductBySlug404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productNotFoundError()),
			}, nil
		}
		return nil, err
	}

	response, err := toAPIProduct(product)
	if err != nil {
		return nil, err
	}

	return generated.GetProductBySlug200JSONResponse(response), nil
}

func (s *Server) CreateProduct(
	ctx context.Context,
	request generated.CreateProductRequestObject,
) (generated.CreateProductResponseObject, error) {
	if request.Body == nil {
		return generated.CreateProduct400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	var status products.Status
	if request.Body.Status != nil {
		status = products.Status(*request.Body.Status)
	}
	product, err := s.products.CreateProduct(ctx, products.CreateProductInput{
		Title:       request.Body.Title,
		Slug:        request.Body.Slug,
		Description: optionalString(request.Body.Description),
		Price:       request.Body.Price,
		Status:      status,
	})
	if err != nil {
		switch {
		case errors.Is(err, products.ErrInvalidInput):
			return generated.CreateProduct400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, products.ErrConflict):
			return generated.CreateProduct409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(productConflictError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPIProduct(product)
	if err != nil {
		return nil, err
	}
	return generated.CreateProduct201JSONResponse(response), nil
}

func (s *Server) UpdateProduct(
	ctx context.Context,
	request generated.UpdateProductRequestObject,
) (generated.UpdateProductResponseObject, error) {
	if request.Body == nil {
		return generated.UpdateProduct400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	product, err := s.products.UpdateProduct(
		ctx,
		request.ProductId.String(),
		products.UpdateProductInput{
			Title:       request.Body.Title,
			Slug:        request.Body.Slug,
			Description: optionalString(request.Body.Description),
			Price:       request.Body.Price,
			Status:      products.Status(request.Body.Status),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, products.ErrInvalidInput):
			return generated.UpdateProduct400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, products.ErrNotFound):
			return generated.UpdateProduct404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productNotFoundError()),
			}, nil
		case errors.Is(err, products.ErrConflict):
			return generated.UpdateProduct409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(productConflictError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPIProduct(product)
	if err != nil {
		return nil, err
	}
	return generated.UpdateProduct200JSONResponse(response), nil
}

func (s *Server) DeleteProduct(
	ctx context.Context,
	request generated.DeleteProductRequestObject,
) (generated.DeleteProductResponseObject, error) {
	err := s.products.DeleteProduct(ctx, request.ProductId.String())
	if err != nil {
		switch {
		case errors.Is(err, products.ErrInvalidInput):
			return generated.DeleteProduct400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, products.ErrNotFound):
			return generated.DeleteProduct404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productNotFoundError()),
			}, nil
		default:
			return nil, err
		}
	}

	return generated.DeleteProduct204Response{}, nil
}

func toAPIProduct(product products.Product) (generated.Product, error) {
	id, err := uuid.Parse(product.ID)
	if err != nil {
		return generated.Product{}, fmt.Errorf("parse product id: %w", err)
	}
	status := generated.ProductStatus(product.Status)
	if !status.Valid() {
		return generated.Product{}, fmt.Errorf("invalid product status %q", product.Status)
	}

	return generated.Product{
		Id:          id,
		Title:       product.Title,
		Slug:        product.Slug,
		Description: product.Description,
		Price:       product.Price,
		Status:      status,
		CreatedAt:   &product.CreatedAt,
		UpdatedAt:   &product.UpdatedAt,
	}, nil
}

func productNotFoundError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeNotFound,
		Message: "product not found",
	}
}

func productConflictError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeConflict,
		Message: "a product with this slug already exists",
	}
}
