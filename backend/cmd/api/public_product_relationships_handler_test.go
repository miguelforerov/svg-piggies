package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/zdenaforero/svg-piggies/backend/internal/productrelationships"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
)

func TestPublicProductRelationshipsPreserveOrderAndFilterInactiveProducts(t *testing.T) {
	const draftID = "a82e7744-d710-425d-bdd1-a396d09a6814"
	const secondActiveID = "95b64568-d176-4d90-a10d-aed34fca3f9d"
	service := &productRelationshipServiceStub{list: func(_ context.Context, productID string) ([]productrelationships.ProductRelationship, error) {
		if productID != testRelationshipProductID {
			t.Fatalf("product ID = %q", productID)
		}
		return []productrelationships.ProductRelationship{
			{RelatedProductID: secondActiveID, DisplayOrder: 0},
			{RelatedProductID: draftID, DisplayOrder: 1},
			{RelatedProductID: testRelatedProductID, DisplayOrder: 2},
		}, nil
	}}
	handler := publicTestHandler(t, dependencies{
		productRelationships: service,
		products: &productServiceStub{
			getProductBySlug: func(_ context.Context, slug string) (products.Product, error) {
				if slug != "party-pig" {
					t.Fatalf("slug = %q", slug)
				}
				return products.Product{ID: testRelationshipProductID, Status: products.StatusActive}, nil
			},
			get: func(_ context.Context, id string) (products.Product, error) {
				status := products.StatusActive
				if id == draftID {
					status = products.StatusDraft
				}
				return products.Product{ID: id, Status: status}, nil
			},
		},
	})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/products/party-pig/relationships", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body)
	}
	var ids []string
	if err := json.Unmarshal(response.Body.Bytes(), &ids); err != nil {
		t.Fatal(err)
	}
	if want := []string{secondActiveID, testRelatedProductID}; !reflect.DeepEqual(ids, want) {
		t.Fatalf("IDs = %v, want %v", ids, want)
	}
}

func TestPublicProductRelationshipsHideInactiveSource(t *testing.T) {
	handler := publicTestHandler(t, dependencies{
		products: &productServiceStub{getProductBySlug: func(context.Context, string) (products.Product, error) {
			return products.Product{ID: testRelationshipProductID, Status: products.StatusDraft}, nil
		}},
	})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/products/party-pig/relationships", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d: %s", response.Code, response.Body)
	}
}
