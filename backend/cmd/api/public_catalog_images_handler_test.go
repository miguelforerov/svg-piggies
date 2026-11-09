package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
	"github.com/zdenaforero/svg-piggies/backend/internal/collections"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
	"github.com/zdenaforero/svg-piggies/backend/internal/productcollections"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
	"github.com/zdenaforero/svg-piggies/backend/internal/productproducttypes"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	"github.com/zdenaforero/svg-piggies/backend/internal/producttypes"
)

func publicTestHandler(t *testing.T, deps dependencies) http.Handler {
	t.Helper()
	handler, err := newHandler(config.Config{Environment: "test", AuthMode: auth.ModeDevelopment, CORSAllowedOrigin: "http://localhost:4321"}, deps, nil)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func TestPublicProductImagesOnlyExposeActiveProducts(t *testing.T) {
	status := products.StatusActive
	product := &productServiceStub{getProductBySlug: func(_ context.Context, slug string) (products.Product, error) {
		if slug != "pig" {
			t.Fatalf("slug = %q", slug)
		}
		return products.Product{ID: testProductID, Slug: slug, Status: status}, nil
	}}
	image := productimages.ProductImageWithURL{ProductImage: productimages.ProductImage{ID: testProductImageID, ProductID: testProductID, R2ObjectKey: "secret-key", OriginalFilename: "secret-name.webp", IsPrimary: true}, ImageURL: "https://example.test/signed"}
	handler := publicTestHandler(t, dependencies{products: product, productImageReads: imageReadServiceStub{images: []productimages.ProductImageWithURL{image}}})
	request := httptest.NewRequest(http.MethodGet, "/api/products/pig/images", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 200 {
		t.Fatalf("status = %d: %s", response.Code, response.Body)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("signed response must not be cached")
	}
	if strings.Contains(response.Body.String(), "secret-") {
		t.Fatalf("private metadata leaked: %s", response.Body)
	}
	var body []map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 1 || body[0]["imageUrl"] != "https://example.test/signed" {
		t.Fatalf("body = %s", response.Body)
	}

	status = products.StatusDraft
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 404 {
		t.Fatalf("draft status = %d", response.Code)
	}

	status = products.StatusActive
	storageFailure := publicTestHandler(t, dependencies{products: product, productImageReads: imageReadServiceStub{err: productimages.ErrStorageFailure}})
	response = httptest.NewRecorder()
	storageFailure.ServeHTTP(response, request)
	if response.Code != 502 {
		t.Fatalf("storage failure status = %d: %s", response.Code, response.Body)
	}
}

func TestPublicCatalogFiltersDraftProducts(t *testing.T) {
	handler := publicTestHandler(t, dependencies{
		products: &productServiceStub{list: func(context.Context) ([]products.Product, error) {
			return []products.Product{
				{ID: testProductID, Title: "Pig", Slug: "pig", Price: "1.00", Status: products.StatusActive},
				{ID: "be4e32b8-c79b-410c-a63d-38c153550ffc", Title: "Draft", Slug: "draft", Price: "1.00", Status: products.StatusDraft},
			}, nil
		}},
		collections:  &collectionServiceStub{list: func(context.Context) ([]collections.Collection, error) { return []collections.Collection{}, nil }},
		productTypes: &productTypeServiceStub{list: func(context.Context) ([]producttypes.ProductType, error) { return []producttypes.ProductType{}, nil }},
		productCollections: &productCollectionServiceStub{list: func(context.Context, string) ([]productcollections.ProductCollection, error) {
			return []productcollections.ProductCollection{}, nil
		}},
		productProductTypes: &productProductTypeServiceStub{list: func(context.Context, string) ([]productproducttypes.ProductProductType, error) {
			return []productproducttypes.ProductProductType{}, nil
		}},
	})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/catalog", nil))
	if response.Code != 200 {
		t.Fatalf("status = %d: %s", response.Code, response.Body)
	}
	var body struct {
		Products []struct {
			Product struct {
				Slug string `json:"slug"`
			} `json:"product"`
		} `json:"products"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Products) != 1 || body.Products[0].Product.Slug != "pig" {
		t.Fatalf("body = %s", response.Body)
	}
}
