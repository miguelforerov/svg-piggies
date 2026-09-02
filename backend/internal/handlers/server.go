package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/collections"
	"github.com/zdenaforero/svg-piggies/backend/internal/productcollections"
	"github.com/zdenaforero/svg-piggies/backend/internal/productproducttypes"
	"github.com/zdenaforero/svg-piggies/backend/internal/productrelationships"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	"github.com/zdenaforero/svg-piggies/backend/internal/producttypes"
)

type CollectionService interface {
	GetCollections(ctx context.Context) ([]collections.Collection, error)
	GetCollection(ctx context.Context, id string) (collections.Collection, error)
	GetCollectionBySlug(ctx context.Context, slug string) (collections.Collection, error)
	CreateCollection(
		ctx context.Context,
		input collections.CreateCollectionInput,
	) (collections.Collection, error)
	UpdateCollection(
		ctx context.Context,
		id string,
		input collections.UpdateCollectionInput,
	) (collections.Collection, error)
	DeleteCollection(ctx context.Context, id string) error
}

type ProductTypeService interface {
	GetProductTypes(ctx context.Context) ([]producttypes.ProductType, error)
	GetProductType(ctx context.Context, id string) (producttypes.ProductType, error)
	CreateProductType(
		ctx context.Context,
		input producttypes.CreateProductTypeInput,
	) (producttypes.ProductType, error)
	UpdateProductType(
		ctx context.Context,
		id string,
		input producttypes.UpdateProductTypeInput,
	) (producttypes.ProductType, error)
	DeleteProductType(ctx context.Context, id string) error
}

type ProductService interface {
	GetProducts(ctx context.Context) ([]products.Product, error)
	GetProduct(ctx context.Context, id string) (products.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (products.Product, error)
	CreateProduct(
		ctx context.Context,
		input products.CreateProductInput,
	) (products.Product, error)
	UpdateProduct(
		ctx context.Context,
		id string,
		input products.UpdateProductInput,
	) (products.Product, error)
	DeleteProduct(ctx context.Context, id string) error
}

type ProductCollectionService interface {
	GetProductCollections(
		ctx context.Context,
		productID string,
	) ([]productcollections.ProductCollection, error)
	CreateProductCollection(
		ctx context.Context,
		input productcollections.CreateProductCollectionInput,
	) (productcollections.ProductCollection, error)
	DeleteProductCollection(ctx context.Context, productID string, collectionID string) error
}

type ProductProductTypeService interface {
	GetProductProductTypes(
		ctx context.Context,
		productID string,
	) ([]productproducttypes.ProductProductType, error)
	CreateProductProductType(
		ctx context.Context,
		input productproducttypes.CreateProductProductTypeInput,
	) (productproducttypes.ProductProductType, error)
	DeleteProductProductType(ctx context.Context, productID string, productTypeID string) error
}

type ProductRelationshipService interface {
	GetProductRelationships(
		ctx context.Context,
		productID string,
	) ([]productrelationships.ProductRelationship, error)
	GetProductRelationship(
		ctx context.Context,
		productID string,
		relationshipID string,
	) (productrelationships.ProductRelationship, error)
	CreateProductRelationship(
		ctx context.Context,
		input productrelationships.CreateProductRelationshipInput,
	) (productrelationships.ProductRelationship, error)
	CreateProductRelationships(
		ctx context.Context,
		input productrelationships.CreateProductRelationshipsInput,
	) (productrelationships.ProductWithRelationships, error)
	UpdateProductRelationship(
		ctx context.Context,
		productID string,
		relationshipID string,
		input productrelationships.UpdateProductRelationshipInput,
	) (productrelationships.ProductRelationship, error)
	ReplaceProductRelationships(
		ctx context.Context,
		input productrelationships.ReplaceProductRelationshipsInput,
	) (productrelationships.ProductWithRelationships, error)
	DeleteProductRelationship(ctx context.Context, productID string, relationshipID string) error
}

type Server struct {
	environment          string
	collections          CollectionService
	productCollections   ProductCollectionService
	productProductTypes  ProductProductTypeService
	productRelationships ProductRelationshipService
	products             ProductService
	productTypes         ProductTypeService
}

var _ generated.StrictServerInterface = (*Server)(nil)

func NewServer(
	environment string,
	collectionService CollectionService,
	productCollectionService ProductCollectionService,
	productProductTypeService ProductProductTypeService,
	productRelationshipService ProductRelationshipService,
	productService ProductService,
	productTypeService ProductTypeService,
) *Server {
	return &Server{
		environment:          environment,
		collections:          collectionService,
		productCollections:   productCollectionService,
		productProductTypes:  productProductTypeService,
		productRelationships: productRelationshipService,
		products:             productService,
		productTypes:         productTypeService,
	}
}

func (s *Server) GetHealth(
	context.Context,
	generated.GetHealthRequestObject,
) (generated.GetHealthResponseObject, error) {
	return generated.GetHealth200JSONResponse{
		Environment: s.environment,
		Status:      generated.Ok,
	}, nil
}

func (s *Server) GetCollections(
	ctx context.Context,
	_ generated.GetCollectionsRequestObject,
) (generated.GetCollectionsResponseObject, error) {
	result, err := s.collections.GetCollections(ctx)
	if err != nil {
		return nil, err
	}

	response := make(generated.GetCollections200JSONResponse, 0, len(result))
	for _, collection := range result {
		apiCollection, err := toAPICollection(collection)
		if err != nil {
			return nil, err
		}
		response = append(response, apiCollection)
	}
	return response, nil
}

func (s *Server) GetCollection(
	ctx context.Context,
	request generated.GetCollectionRequestObject,
) (generated.GetCollectionResponseObject, error) {
	collection, err := s.collections.GetCollection(ctx, request.CollectionId.String())
	if err != nil {
		switch {
		case errors.Is(err, collections.ErrInvalidInput):
			return generated.GetCollection400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, collections.ErrNotFound):
			return generated.GetCollection404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPICollection(collection)
	if err != nil {
		return nil, err
	}
	return generated.GetCollection200JSONResponse(response), nil
}

func (s *Server) GetCollectionBySlug(
	ctx context.Context,
	request generated.GetCollectionBySlugRequestObject,
) (generated.GetCollectionBySlugResponseObject, error) {
	collection, err := s.collections.GetCollectionBySlug(ctx, request.Slug)
	if err != nil {
		switch {
		case errors.Is(err, collections.ErrNotFound):
			return generated.GetCollectionBySlug404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPICollection(collection)
	if err != nil {
		return nil, err
	}
	return generated.GetCollectionBySlug200JSONResponse(response), nil
}

func (s *Server) CreateCollection(
	ctx context.Context,
	request generated.CreateCollectionRequestObject,
) (generated.CreateCollectionResponseObject, error) {
	if request.Body == nil {
		return generated.CreateCollection400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	collection, err := s.collections.CreateCollection(ctx, collections.CreateCollectionInput{
		Name:        request.Body.Name,
		Slug:        request.Body.Slug,
		Description: optionalString(request.Body.Description),
	})
	if err != nil {
		switch {
		case errors.Is(err, collections.ErrInvalidInput):
			return generated.CreateCollection400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, collections.ErrConflict):
			return generated.CreateCollection409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(conflictError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPICollection(collection)
	if err != nil {
		return nil, err
	}
	return generated.CreateCollection201JSONResponse(response), nil
}

func (s *Server) UpdateCollection(
	ctx context.Context,
	request generated.UpdateCollectionRequestObject,
) (generated.UpdateCollectionResponseObject, error) {
	if request.Body == nil {
		return generated.UpdateCollection400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	collection, err := s.collections.UpdateCollection(
		ctx,
		request.CollectionId.String(),
		collections.UpdateCollectionInput{
			Name:        request.Body.Name,
			Slug:        request.Body.Slug,
			Description: optionalString(request.Body.Description),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, collections.ErrInvalidInput):
			return generated.UpdateCollection400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, collections.ErrNotFound):
			return generated.UpdateCollection404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError()),
			}, nil
		case errors.Is(err, collections.ErrConflict):
			return generated.UpdateCollection409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(conflictError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPICollection(collection)
	if err != nil {
		return nil, err
	}
	return generated.UpdateCollection200JSONResponse(response), nil
}

func (s *Server) DeleteCollection(
	ctx context.Context,
	request generated.DeleteCollectionRequestObject,
) (generated.DeleteCollectionResponseObject, error) {
	err := s.collections.DeleteCollection(ctx, request.CollectionId.String())
	if err != nil {
		switch {
		case errors.Is(err, collections.ErrInvalidInput):
			return generated.DeleteCollection400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, collections.ErrNotFound):
			return generated.DeleteCollection404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError()),
			}, nil
		default:
			return nil, err
		}
	}

	return generated.DeleteCollection204Response{}, nil
}

func toAPICollection(collection collections.Collection) (generated.Collection, error) {
	id, err := uuid.Parse(collection.ID)
	if err != nil {
		return generated.Collection{}, fmt.Errorf("parse collection id: %w", err)
	}
	return generated.Collection{
		Id:          id,
		Name:        collection.Name,
		Slug:        collection.Slug,
		Description: collection.Description,
	}, nil
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func invalidRequestError(err error) generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeInvalidRequest,
		Message: err.Error(),
	}
}

func notFoundError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeNotFound,
		Message: "collection not found",
	}
}

func conflictError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeConflict,
		Message: "a collection with this slug already exists",
	}
}
