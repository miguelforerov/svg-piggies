package handlers

import (
	"context"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
)

func (s *Server) GetPublicCatalog(ctx context.Context, _ generated.GetPublicCatalogRequestObject) (generated.GetPublicCatalogResponseObject, error) {
	productsList, err := s.products.GetProducts(ctx)
	if err != nil {
		return nil, err
	}
	collectionsList, err := s.collections.GetCollections(ctx)
	if err != nil {
		return nil, err
	}
	productTypesList, err := s.productTypes.GetProductTypes(ctx)
	if err != nil {
		return nil, err
	}

	response := generated.PublicCatalog{
		Products:     make([]generated.PublicCatalogProduct, 0),
		Collections:  make([]generated.Collection, 0, len(collectionsList)),
		ProductTypes: make([]generated.ProductType, 0, len(productTypesList)),
	}
	for _, collection := range collectionsList {
		item, err := toAPICollection(collection)
		if err != nil {
			return nil, err
		}
		response.Collections = append(response.Collections, item)
	}
	for _, productType := range productTypesList {
		item, err := toAPIProductType(productType)
		if err != nil {
			return nil, err
		}
		response.ProductTypes = append(response.ProductTypes, item)
	}
	for _, product := range productsList {
		if product.Status != products.StatusActive {
			continue
		}
		apiProduct, err := toAPIProduct(product)
		if err != nil {
			return nil, err
		}
		collections, err := s.productCollections.GetProductCollections(ctx, product.ID)
		if err != nil {
			return nil, err
		}
		types, err := s.productProductTypes.GetProductProductTypes(ctx, product.ID)
		if err != nil {
			return nil, err
		}
		item := generated.PublicCatalogProduct{Product: apiProduct,
			CollectionIds:  make([]uuid.UUID, 0, len(collections)),
			ProductTypeIds: make([]uuid.UUID, 0, len(types)),
		}
		for _, assignment := range collections {
			id, err := uuid.Parse(assignment.CollectionID)
			if err != nil {
				return nil, err
			}
			item.CollectionIds = append(item.CollectionIds, id)
		}
		for _, assignment := range types {
			id, err := uuid.Parse(assignment.ProductTypeID)
			if err != nil {
				return nil, err
			}
			item.ProductTypeIds = append(item.ProductTypeIds, id)
		}
		response.Products = append(response.Products, item)
	}
	return generated.GetPublicCatalog200JSONResponse(response), nil
}
