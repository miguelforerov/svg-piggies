package main

import (
	"fmt"

	"github.com/zdenaforero/svg-piggies/backend/internal/carts"
	cartspostgres "github.com/zdenaforero/svg-piggies/backend/internal/carts/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/collections"
	collectionspostgres "github.com/zdenaforero/svg-piggies/backend/internal/collections/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/customers"
	customerspostgres "github.com/zdenaforero/svg-piggies/backend/internal/customers/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/database"
	"github.com/zdenaforero/svg-piggies/backend/internal/handlers"
	"github.com/zdenaforero/svg-piggies/backend/internal/orders"
	orderspostgres "github.com/zdenaforero/svg-piggies/backend/internal/orders/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/productcollections"
	productcollectionspostgres "github.com/zdenaforero/svg-piggies/backend/internal/productcollections/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
	productimagespostgres "github.com/zdenaforero/svg-piggies/backend/internal/productimages/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/productproducttypes"
	productproducttypespostgres "github.com/zdenaforero/svg-piggies/backend/internal/productproducttypes/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/productrelationships"
	productrelationshipspostgres "github.com/zdenaforero/svg-piggies/backend/internal/productrelationships/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	productspostgres "github.com/zdenaforero/svg-piggies/backend/internal/products/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/producttypes"
	producttypespostgres "github.com/zdenaforero/svg-piggies/backend/internal/producttypes/postgres"
)

type dependencies struct {
	carts                handlers.CartService
	collections          handlers.CollectionService
	customers            handlers.CustomerService
	orders               handlers.OrderService
	productCollections   handlers.ProductCollectionService
	productImages        handlers.ProductImageService
	productProductTypes  handlers.ProductProductTypeService
	productRelationships handlers.ProductRelationshipService
	products             handlers.ProductService
	productTypes         handlers.ProductTypeService
}

func buildDependencies(provider database.Provider) (dependencies, error) {
	cartRepository, err := cartspostgres.NewRepository(provider)
	if err != nil {
		return dependencies{}, fmt.Errorf("create carts repository: %w", err)
	}

	cartService, err := carts.NewService(cartRepository)
	if err != nil {
		return dependencies{}, fmt.Errorf("create carts service: %w", err)
	}

	customerRepository, err := customerspostgres.NewRepository(provider)
	if err != nil {
		return dependencies{}, fmt.Errorf("create customers repository: %w", err)
	}
	customerService, err := customers.NewService(customerRepository)
	if err != nil {
		return dependencies{}, fmt.Errorf("create customers service: %w", err)
	}

	orderRepository, err := orderspostgres.NewRepository(provider)
	if err != nil {
		return dependencies{}, fmt.Errorf("create orders repository: %w", err)
	}
	orderService, err := orders.NewService(orderRepository)
	if err != nil {
		return dependencies{}, fmt.Errorf("create orders service: %w", err)
	}

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

	productImageRepository, err := productimagespostgres.NewRepository(provider)
	if err != nil {
		return dependencies{}, fmt.Errorf("create product images repository: %w", err)
	}

	productImageService, err := productimages.NewService(productImageRepository)
	if err != nil {
		return dependencies{}, fmt.Errorf("create product images service: %w", err)
	}

	productProductTypeRepository, err := productproducttypespostgres.NewRepository(provider)
	if err != nil {
		return dependencies{}, fmt.Errorf("create product product types repository: %w", err)
	}

	productProductTypeService, err := productproducttypes.NewService(productProductTypeRepository)
	if err != nil {
		return dependencies{}, fmt.Errorf("create product product types service: %w", err)
	}

	productRelationshipRepository, err := productrelationshipspostgres.NewRepository(provider)
	if err != nil {
		return dependencies{}, fmt.Errorf("create product relationships repository: %w", err)
	}

	productRelationshipService, err := productrelationships.NewService(productRelationshipRepository)
	if err != nil {
		return dependencies{}, fmt.Errorf("create product relationships service: %w", err)
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
		carts:                cartService,
		collections:          service,
		customers:            customerService,
		orders:               orderService,
		productCollections:   productCollectionService,
		productImages:        productImageService,
		productProductTypes:  productProductTypeService,
		productRelationships: productRelationshipService,
		products:             productService,
		productTypes:         productTypeService,
	}, nil
}
