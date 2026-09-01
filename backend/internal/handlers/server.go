package handlers

import (
	"context"

	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/collections"
	"github.com/zdenaforero/svg-piggies/backend/internal/productcollections"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
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
	GetProductTypeBySlug(ctx context.Context, slug string) (producttypes.ProductType, error)
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

type ProductImageService interface {
	AddProductImages(
		ctx context.Context,
		input productimages.AddProductImagesInput,
	) ([]productimages.ProductImage, error)
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
	productImages        ProductImageService
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
	productImageService ProductImageService,
	productProductTypeService ProductProductTypeService,
	productRelationshipService ProductRelationshipService,
	productService ProductService,
	productTypeService ProductTypeService,
) *Server {
	return &Server{
		environment:          environment,
		collections:          collectionService,
		productCollections:   productCollectionService,
		productImages:        productImageService,
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

func notFoundError(resource string) generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeNotFound,
		Message: resource + " not found",
	}
}

func conflictError() generated.Error {
	return generated.Error{
		Code:    generated.ErrorCodeConflict,
		Message: "a collection with this slug already exists",
	}
}
