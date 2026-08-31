package main

import (
	"fmt"

	"github.com/zdenaforero/svg-piggies/backend/internal/collections"
	collectionspostgres "github.com/zdenaforero/svg-piggies/backend/internal/collections/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/database"
	"github.com/zdenaforero/svg-piggies/backend/internal/handlers"
	"github.com/zdenaforero/svg-piggies/backend/internal/productcollections"
	productcollectionspostgres "github.com/zdenaforero/svg-piggies/backend/internal/productcollections/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/productproducttypes"
	productproducttypespostgres "github.com/zdenaforero/svg-piggies/backend/internal/productproducttypes/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	productspostgres "github.com/zdenaforero/svg-piggies/backend/internal/products/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/producttypes"
	producttypespostgres "github.com/zdenaforero/svg-piggies/backend/internal/producttypes/postgres"
)

type dependencies struct {
	collections         handlers.CollectionService
	productCollections  handlers.ProductCollectionService
	productProductTypes handlers.ProductProductTypeService
	products            handlers.ProductService
	productTypes        handlers.ProductTypeService
}

func buildDependencies(provider database.Provider) (dependencies, error) {
	repository, err := collectionspostgres.NewRepository(provider)
	if err != nil {
		return dependencies{}, fmt.Errorf("create collections repository: %w", err)
	}

	service, err := collections.NewService(repository)
	if err != nil {
		return dependencies{}, fmt.Errorf("create collections service: %w", err)
	}

	productCollectionRepository, err := productcollectionspostgres.NewRepository(provider)
	if err != nil {
		return dependencies{}, fmt.Errorf("create product collections repository: %w", err)
	}

	productCollectionService, err := productcollections.NewService(productCollectionRepository)
	if err != nil {
		return dependencies{}, fmt.Errorf("create product collections service: %w", err)
	}

	productProductTypeRepository, err := productproducttypespostgres.NewRepository(provider)
	if err != nil {
		return dependencies{}, fmt.Errorf("create product product types repository: %w", err)
	}

	productProductTypeService, err := productproducttypes.NewService(productProductTypeRepository)
	if err != nil {
		return dependencies{}, fmt.Errorf("create product product types service: %w", err)
	}

	productRepository, err := productspostgres.NewRepository(provider)
	if err != nil {
		return dependencies{}, fmt.Errorf("create products repository: %w", err)
	}

	productService, err := products.NewService(productRepository)
	if err != nil {
		return dependencies{}, fmt.Errorf("create products service: %w", err)
	}

	productTypeRepository, err := producttypespostgres.NewRepository(provider)
	if err != nil {
		return dependencies{}, fmt.Errorf("create product types repository: %w", err)
	}

	productTypeService, err := producttypes.NewService(productTypeRepository)
	if err != nil {
		return dependencies{}, fmt.Errorf("create product types service: %w", err)
	}

	return dependencies{
		collections:         service,
		productCollections:  productCollectionService,
		productProductTypes: productProductTypeService,
		products:            productService,
		productTypes:        productTypeService,
	}, nil
}
