package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
	"github.com/zdenaforero/svg-piggies/backend/internal/productrelationships"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
)

const (
	testRelationshipProductID = "b2f18e56-dff5-4b1d-a454-60481472179d"
	testRelationshipID        = "43124267-9f20-47f9-890f-dd03b94ac1e6"
	testRelatedProductID      = "511477a7-d02f-4601-b33e-a5a87bc8aee3"
)

func TestProductRelationshipRoutes(t *testing.T) {
	service := &productRelationshipServiceStub{
		list: func(
			_ context.Context,
			productID string,
		) ([]productrelationships.ProductRelationship, error) {
			if productID != testRelationshipProductID {
				t.Fatalf("GetProductRelationships() productID = %q", productID)
			}
			return []productrelationships.ProductRelationship{{
				ID:               testRelationshipID,
				ProductID:        productID,
				RelatedProductID: testRelatedProductID,
				DisplayOrder:     2,
			}}, nil
		},
		get: func(
			_ context.Context,
			productID string,
			relationshipID string,
		) (productrelationships.ProductRelationship, error) {
			if productID != testRelationshipProductID || relationshipID != testRelationshipID {
				t.Fatalf("GetProductRelationship() IDs = %q, %q", productID, relationshipID)
			}
			return productrelationships.ProductRelationship{
				ID: relationshipID, ProductID: productID,
				RelatedProductID: testRelatedProductID, DisplayOrder: 2,
			}, nil
		},
		create: func(
			_ context.Context,
			input productrelationships.CreateProductRelationshipInput,
		) (productrelationships.ProductRelationship, error) {
			want := productrelationships.CreateProductRelationshipInput{
				ProductID:        testRelationshipProductID,
				RelatedProductID: testRelatedProductID,
				DisplayOrder:     2,
			}
			if !reflect.DeepEqual(input, want) {
				t.Fatalf("CreateProductRelationship() input = %#v, want %#v", input, want)
			}
			return productrelationships.ProductRelationship{
				ID: testRelationshipID, ProductID: input.ProductID,
				RelatedProductID: input.RelatedProductID, DisplayOrder: input.DisplayOrder,
			}, nil
		},
		createMany: func(
			_ context.Context,
			input productrelationships.CreateProductRelationshipsInput,
		) (productrelationships.ProductWithRelationships, error) {
			wantIDs := []string{testRelatedProductID}
			if input.ProductID != testRelationshipProductID ||
				!reflect.DeepEqual(input.RelatedProductIDs, wantIDs) {
				t.Fatalf("CreateProductRelationships() input = %#v", input)
			}
			return testProductRelationshipAggregate(), nil
		},
		update: func(
			_ context.Context,
			productID string,
			relationshipID string,
			input productrelationships.UpdateProductRelationshipInput,
		) (productrelationships.ProductRelationship, error) {
			if productID != testRelationshipProductID || relationshipID != testRelationshipID {
				t.Fatalf("UpdateProductRelationship() IDs = %q, %q", productID, relationshipID)
			}
			return productrelationships.ProductRelationship{
				ID: relationshipID, ProductID: productID,
				RelatedProductID: input.RelatedProductID, DisplayOrder: input.DisplayOrder,
			}, nil
		},
		replace: func(
			_ context.Context,
			input productrelationships.ReplaceProductRelationshipsInput,
		) (productrelationships.ProductWithRelationships, error) {
			wantIDs := []string{testRelatedProductID}
			if input.ProductID != testRelationshipProductID ||
				!reflect.DeepEqual(input.RelatedProductIDs, wantIDs) {
				t.Fatalf("ReplaceProductRelationships() input = %#v", input)
			}
			timestamp := time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC)
			return productrelationships.ProductWithRelationships{
				Product: products.Product{
					ID: testRelationshipProductID, Title: "Party Pig", Slug: "party-pig",
					Price: "19.95", Status: products.StatusActive,
					CreatedAt: timestamp, UpdatedAt: timestamp,
				},
				Relationships: []productrelationships.PopulatedProductRelationship{{
					RelationshipID: testRelationshipID,
					DisplayOrder:   0,
					Product: products.Product{
						ID: testRelatedProductID, Title: "Friend Pig", Slug: "friend-pig",
						Price: "12.00", Status: products.StatusActive,
						CreatedAt: timestamp, UpdatedAt: timestamp,
					},
				}},
			}, nil
		},
		delete: func(_ context.Context, productID string, relationshipID string) error {
			if productID != testRelationshipProductID || relationshipID != testRelationshipID {
				t.Fatalf("DeleteProductRelationship() IDs = %q, %q", productID, relationshipID)
			}
			return nil
		},
	}
	handler := newProductRelationshipTestHandler(service)

	basePath := "/api/admin/products/" + testRelationshipProductID + "/relationships"
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{name: "list", method: http.MethodGet, path: basePath, wantStatus: http.StatusOK},
		{name: "get", method: http.MethodGet, path: basePath + "/" + testRelationshipID, wantStatus: http.StatusOK},
		{
			name:   "create",
			method: http.MethodPost,
			path:   basePath,
			body: `{"relatedProductId":"` + testRelatedProductID +
				`","displayOrder":2}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "add many",
			method:     http.MethodPost,
			path:       basePath + "/bulk",
			body:       `{"relatedProductIds":["` + testRelatedProductID + `"]}`,
			wantStatus: http.StatusOK,
		},
		{
			name:   "update",
			method: http.MethodPut,
			path:   basePath + "/" + testRelationshipID,
			body: `{"relatedProductId":"` + testRelatedProductID +
				`","displayOrder":4}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "replace many",
			method:     http.MethodPut,
			path:       basePath,
			body:       `{"relatedProductIds":["` + testRelatedProductID + `"]}`,
			wantStatus: http.StatusOK,
		},
		{name: "delete", method: http.MethodDelete, path: basePath + "/" + testRelationshipID, wantStatus: http.StatusNoContent},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, bytes.NewBufferString(test.body))
			if test.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body)
			}
			if test.name == "add many" || test.name == "replace many" {
				var body struct {
					Product struct {
						ID string `json:"id"`
					} `json:"product"`
					Relationships []struct {
						Product struct {
							ID string `json:"id"`
						} `json:"product"`
					} `json:"relationships"`
				}
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if body.Product.ID != testRelationshipProductID ||
					len(body.Relationships) != 1 ||
					body.Relationships[0].Product.ID != testRelatedProductID {
					t.Fatalf("populated response = %#v", body)
				}
			}
		})
	}
}

func TestProductRelationshipCreateDefaultsDisplayOrder(t *testing.T) {
	t.Parallel()

	handler := newProductRelationshipTestHandler(&productRelationshipServiceStub{
		create: func(
			_ context.Context,
			input productrelationships.CreateProductRelationshipInput,
		) (productrelationships.ProductRelationship, error) {
			if input.DisplayOrder != 0 {
				t.Fatalf("display order = %d, want 0", input.DisplayOrder)
			}
			return productrelationships.ProductRelationship{
				ID: testRelationshipID, ProductID: input.ProductID,
				RelatedProductID: input.RelatedProductID,
			}, nil
		},
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/products/"+testRelationshipProductID+"/relationships",
		bytes.NewBufferString(`{"relatedProductId":"`+testRelatedProductID+`"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusCreated, response.Body)
	}
}

func TestProductRelationshipRouteRejectsInvalidUUID(t *testing.T) {
	t.Parallel()

	handler := newProductRelationshipTestHandler(&productRelationshipServiceStub{})
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/admin/products/not-a-uuid/relationships",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	assertErrorCode(t, response, "invalid_request")
}

func TestProductRelationshipRouteMapsConflict(t *testing.T) {
	t.Parallel()

	handler := newProductRelationshipTestHandler(&productRelationshipServiceStub{
		create: func(
			context.Context,
			productrelationships.CreateProductRelationshipInput,
		) (productrelationships.ProductRelationship, error) {
			return productrelationships.ProductRelationship{}, productrelationships.ErrConflict
		},
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/products/"+testRelationshipProductID+"/relationships",
		bytes.NewBufferString(`{"relatedProductId":"`+testRelatedProductID+`"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusConflict, response.Body)
	}
	assertErrorCode(t, response, "conflict")
}

func TestProductRelationshipRouteMapsMissingReference(t *testing.T) {
	t.Parallel()

	handler := newProductRelationshipTestHandler(&productRelationshipServiceStub{
		list: func(context.Context, string) ([]productrelationships.ProductRelationship, error) {
			return nil, productrelationships.ErrReferenceNotFound
		},
	})
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/admin/products/"+testRelationshipProductID+"/relationships",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	assertErrorCode(t, response, "not_found")
}

func newProductRelationshipTestHandler(service *productRelationshipServiceStub) http.Handler {
	handler, err := newHandler(config.Config{
		Environment:       "test",
		AuthMode:          auth.ModeDevelopment,
		CORSAllowedOrigin: "http://localhost:4321",
	}, dependencies{productRelationships: service}, nil)
	if err != nil {
		panic(err)
	}
	return handler
}

type productRelationshipServiceStub struct {
	list       func(context.Context, string) ([]productrelationships.ProductRelationship, error)
	get        func(context.Context, string, string) (productrelationships.ProductRelationship, error)
	create     func(context.Context, productrelationships.CreateProductRelationshipInput) (productrelationships.ProductRelationship, error)
	createMany func(context.Context, productrelationships.CreateProductRelationshipsInput) (productrelationships.ProductWithRelationships, error)
	update     func(context.Context, string, string, productrelationships.UpdateProductRelationshipInput) (productrelationships.ProductRelationship, error)
	replace    func(context.Context, productrelationships.ReplaceProductRelationshipsInput) (productrelationships.ProductWithRelationships, error)
	delete     func(context.Context, string, string) error
}

func (s *productRelationshipServiceStub) GetProductRelationships(
	ctx context.Context,
	productID string,
) ([]productrelationships.ProductRelationship, error) {
	if s.list == nil {
		return nil, errors.New("unexpected GetProductRelationships call")
	}
	return s.list(ctx, productID)
}

func (s *productRelationshipServiceStub) GetProductRelationship(
	ctx context.Context,
	productID string,
	relationshipID string,
) (productrelationships.ProductRelationship, error) {
	if s.get == nil {
		return productrelationships.ProductRelationship{}, errors.New(
			"unexpected GetProductRelationship call",
		)
	}
	return s.get(ctx, productID, relationshipID)
}

func (s *productRelationshipServiceStub) CreateProductRelationship(
	ctx context.Context,
	input productrelationships.CreateProductRelationshipInput,
) (productrelationships.ProductRelationship, error) {
	if s.create == nil {
		return productrelationships.ProductRelationship{}, errors.New(
			"unexpected CreateProductRelationship call",
		)
	}
	return s.create(ctx, input)
}

func (s *productRelationshipServiceStub) CreateProductRelationships(
	ctx context.Context,
	input productrelationships.CreateProductRelationshipsInput,
) (productrelationships.ProductWithRelationships, error) {
	if s.createMany == nil {
		return productrelationships.ProductWithRelationships{}, errors.New(
			"unexpected CreateProductRelationships call",
		)
	}
	return s.createMany(ctx, input)
}

func (s *productRelationshipServiceStub) UpdateProductRelationship(
	ctx context.Context,
	productID string,
	relationshipID string,
	input productrelationships.UpdateProductRelationshipInput,
) (productrelationships.ProductRelationship, error) {
	if s.update == nil {
		return productrelationships.ProductRelationship{}, errors.New(
			"unexpected UpdateProductRelationship call",
		)
	}
	return s.update(ctx, productID, relationshipID, input)
}

func (s *productRelationshipServiceStub) ReplaceProductRelationships(
	ctx context.Context,
	input productrelationships.ReplaceProductRelationshipsInput,
) (productrelationships.ProductWithRelationships, error) {
	if s.replace == nil {
		return productrelationships.ProductWithRelationships{}, errors.New(
			"unexpected ReplaceProductRelationships call",
		)
	}
	return s.replace(ctx, input)
}

func (s *productRelationshipServiceStub) DeleteProductRelationship(
	ctx context.Context,
	productID string,
	relationshipID string,
) error {
	if s.delete == nil {
		return errors.New("unexpected DeleteProductRelationship call")
	}
	return s.delete(ctx, productID, relationshipID)
}

func testProductRelationshipAggregate() productrelationships.ProductWithRelationships {
	timestamp := time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC)
	return productrelationships.ProductWithRelationships{
		Product: products.Product{
			ID: testRelationshipProductID, Title: "Party Pig", Slug: "party-pig",
			Price: "19.95", Status: products.StatusActive,
			CreatedAt: timestamp, UpdatedAt: timestamp,
		},
		Relationships: []productrelationships.PopulatedProductRelationship{{
			RelationshipID: testRelationshipID,
			DisplayOrder:   0,
			Product: products.Product{
				ID: testRelatedProductID, Title: "Friend Pig", Slug: "friend-pig",
				Price: "12.00", Status: products.StatusActive,
				CreatedAt: timestamp, UpdatedAt: timestamp,
			},
		}},
	}
}
