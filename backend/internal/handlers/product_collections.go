package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/productcollections"
)

func (s *Server) GetProductCollections(
	ctx context.Context,
	request generated.GetProductCollectionsRequestObject,
) (generated.GetProductCollectionsResponseObject, error) {
	result, err := s.productCollections.GetProductCollections(ctx, request.ProductId.String())
	if err != nil {
		switch {
		case errors.Is(err, productcollections.ErrInvalidInput):
			return generated.GetProductCollections400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productcollections.ErrReferenceNotFound):
			return generated.GetProductCollections404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productCollectionReferenceNotFoundError()),
			}, nil
		default:
			return nil, err
		}
	}

	response := make(generated.GetProductCollections200JSONResponse, 0, len(result))
	for _, productCollection := range result {
		apiProductCollection, err := toAPIProductCollection(productCollection)
		if err != nil {
			return nil, err
		}
		response = append(response, apiProductCollection)
	}
	return response, nil
}

func (s *Server) CreateProductCollection(
	ctx context.Context,
	request generated.CreateProductCollectionRequestObject,
) (generated.CreateProductCollectionResponseObject, error) {
	if request.Body == nil {
		return generated.CreateProductCollection400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	productCollection, err := s.productCollections.CreateProductCollection(
		ctx,
		productcollections.CreateProductCollectionInput{
			ProductID:    request.ProductId.String(),
			CollectionID: request.Body.CollectionId.String(),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, productcollections.ErrInvalidInput):
			return generated.CreateProductCollection400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productcollections.ErrReferenceNotFound):
			return generated.CreateProductCollection404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productCollectionReferenceNotFoundError()),
			}, nil
		case errors.Is(err, productcollections.ErrConflict):
			return generated.CreateProductCollection409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(productCollectionConflictError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPIProductCollection(productCollection)
	if err != nil {
		return nil, err
	}
	return generated.CreateProductCollection201JSONResponse(response), nil
}

func (s *Server) DeleteProductCollection(
	ctx context.Context,
	request generated.DeleteProductCollectionRequestObject,
) (generated.DeleteProductCollectionResponseObject, error) {
	err := s.productCollections.DeleteProductCollection(
		ctx,
		request.ProductId.String(),
		request.CollectionId.String(),
	)
	if err != nil {
		switch {
		case errors.Is(err, productcollections.ErrInvalidInput):
			return generated.DeleteProductCollection400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, productcollections.ErrNotFound),
			errors.Is(err, productcollections.ErrReferenceNotFound):
			return generated.DeleteProductCollection404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(productCollectionNotFoundError()),
			}, nil
		default:
			return nil, err
		}
	}

	return generated.DeleteProductCollection204Response{}, nil
}

func toAPIProductCollection(
	productCollection productcollections.ProductCollection,
) (generated.ProductCollection, error) {
	productID, err := uuid.Parse(productCollection.ProductID)
	if err != nil {
		return generated.ProductCollection{}, fmt.Errorf("parse product id: %w", err)
	}
	collectionID, err := uuid.Parse(productCollection.CollectionID)
	if err != nil {
		return generated.ProductCollection{}, fmt.Errorf("parse collection id: %w", err)
	}
	return generated.ProductCollection{
		ProductId:    productID,
		CollectionId: collectionID,
	}, nil
}

func productCollectionReferenceNotFoundError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeNotFound,
		Message: "product or collection not found",
	}
}

func productCollectionNotFoundError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeNotFound,
		Message: "product collection assignment not found",
	}
}

func productCollectionConflictError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeConflict,
		Message: "product is already assigned to this collection",
	}
}
