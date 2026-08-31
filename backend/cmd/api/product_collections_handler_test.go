package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
	"github.com/zdenaforero/svg-piggies/backend/internal/productcollections"
)

const testProductCollectionCollectionID = "ac40666e-f053-4344-888e-b5bbd3389b77"

func TestProductCollectionRoutes(t *testing.T) {
	service := &productCollectionServiceStub{
		list: func(_ context.Context, productID string) ([]productcollections.ProductCollection, error) {
			if productID != testProductID {
				t.Fatalf("GetProductCollections() productID = %q", productID)
			}
			return []productcollections.ProductCollection{{
				ProductID:    productID,
				CollectionID: testProductCollectionCollectionID,
			}}, nil
		},
		create: func(
			_ context.Context,
			input productcollections.CreateProductCollectionInput,
		) (productcollections.ProductCollection, error) {
			if input.ProductID != testProductID ||
				input.CollectionID != testProductCollectionCollectionID {
				t.Fatalf("CreateProductCollection() input = %#v", input)
			}
			return productcollections.ProductCollection(input), nil
		},
		delete: func(_ context.Context, productID string, collectionID string) error {
			if productID != testProductID || collectionID != testProductCollectionCollectionID {
				t.Fatalf("DeleteProductCollection() IDs = %q, %q", productID, collectionID)
			}
			return nil
		},
	}
	handler := newProductCollectionTestHandler(service)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "list",
			method:     http.MethodGet,
			path:       "/api/admin/products/" + testProductID + "/collections",
			wantStatus: http.StatusOK,
		},
		{
			name:       "create",
			method:     http.MethodPost,
			path:       "/api/admin/products/" + testProductID + "/collections",
			body:       `{"collectionId":"` + testProductCollectionCollectionID + `"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:   "delete",
			method: http.MethodDelete,
			path: "/api/admin/products/" + testProductID + "/collections/" +
				testProductCollectionCollectionID,
			wantStatus: http.StatusNoContent,
		},
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
		})
	}
}

func TestProductCollectionRouteRejectsInvalidUUID(t *testing.T) {
	t.Parallel()

	handler := newProductCollectionTestHandler(&productCollectionServiceStub{})
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/admin/products/not-a-uuid/collections",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	assertErrorCode(t, response, "invalid_request")
}

func TestProductCollectionRouteMapsConflict(t *testing.T) {
	t.Parallel()

	handler := newProductCollectionTestHandler(&productCollectionServiceStub{
		create: func(
			context.Context,
			productcollections.CreateProductCollectionInput,
		) (productcollections.ProductCollection, error) {
			return productcollections.ProductCollection{}, productcollections.ErrConflict
		},
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/products/"+testProductID+"/collections",
		bytes.NewBufferString(`{"collectionId":"`+testProductCollectionCollectionID+`"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusConflict, response.Body)
	}
	assertErrorCode(t, response, "conflict")
}

func TestProductCollectionRouteMapsMissingReference(t *testing.T) {
	t.Parallel()

	handler := newProductCollectionTestHandler(&productCollectionServiceStub{
		list: func(context.Context, string) ([]productcollections.ProductCollection, error) {
			return nil, productcollections.ErrReferenceNotFound
		},
	})
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/admin/products/"+testProductID+"/collections",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	assertErrorCode(t, response, "not_found")
}

func newProductCollectionTestHandler(service *productCollectionServiceStub) http.Handler {
	handler, err := newHandler(config.Config{
		Environment:       "test",
		AuthMode:          auth.ModeDevelopment,
		CORSAllowedOrigin: "http://localhost:4321",
	}, dependencies{productCollections: service}, nil)
	if err != nil {
		panic(err)
	}
	return handler
}

type productCollectionServiceStub struct {
	list   func(context.Context, string) ([]productcollections.ProductCollection, error)
	create func(context.Context, productcollections.CreateProductCollectionInput) (productcollections.ProductCollection, error)
	delete func(context.Context, string, string) error
}

func (s *productCollectionServiceStub) GetProductCollections(
	ctx context.Context,
	productID string,
) ([]productcollections.ProductCollection, error) {
	if s.list == nil {
		return nil, errors.New("unexpected GetProductCollections call")
	}
	return s.list(ctx, productID)
}

func (s *productCollectionServiceStub) CreateProductCollection(
	ctx context.Context,
	input productcollections.CreateProductCollectionInput,
) (productcollections.ProductCollection, error) {
	if s.create == nil {
		return productcollections.ProductCollection{}, errors.New(
			"unexpected CreateProductCollection call",
		)
	}
	return s.create(ctx, input)
}

func (s *productCollectionServiceStub) DeleteProductCollection(
	ctx context.Context,
	productID string,
	collectionID string,
) error {
	if s.delete == nil {
		return errors.New("unexpected DeleteProductCollection call")
	}
	return s.delete(ctx, productID, collectionID)
}
