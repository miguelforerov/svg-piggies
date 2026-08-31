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
	"github.com/zdenaforero/svg-piggies/backend/internal/productproducttypes"
)

const testProductProductTypeID = "64a85c9b-a0b8-4d83-bec3-f97615c6f32b"

func TestProductProductTypeRoutes(t *testing.T) {
	service := &productProductTypeServiceStub{
		list: func(
			_ context.Context,
			productID string,
		) ([]productproducttypes.ProductProductType, error) {
			if productID != testProductID {
				t.Fatalf("GetProductProductTypes() productID = %q", productID)
			}
			return []productproducttypes.ProductProductType{{
				ProductID:     productID,
				ProductTypeID: testProductProductTypeID,
			}}, nil
		},
		create: func(
			_ context.Context,
			input productproducttypes.CreateProductProductTypeInput,
		) (productproducttypes.ProductProductType, error) {
			if input.ProductID != testProductID || input.ProductTypeID != testProductProductTypeID {
				t.Fatalf("CreateProductProductType() input = %#v", input)
			}
			return productproducttypes.ProductProductType(input), nil
		},
		delete: func(_ context.Context, productID string, productTypeID string) error {
			if productID != testProductID || productTypeID != testProductProductTypeID {
				t.Fatalf("DeleteProductProductType() IDs = %q, %q", productID, productTypeID)
			}
			return nil
		},
	}
	handler := newProductProductTypeTestHandler(service)

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
			path:       "/api/admin/products/" + testProductID + "/product-types",
			wantStatus: http.StatusOK,
		},
		{
			name:       "create",
			method:     http.MethodPost,
			path:       "/api/admin/products/" + testProductID + "/product-types",
			body:       `{"productTypeId":"` + testProductProductTypeID + `"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:   "delete",
			method: http.MethodDelete,
			path: "/api/admin/products/" + testProductID + "/product-types/" +
				testProductProductTypeID,
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

func TestProductProductTypeRouteRejectsInvalidUUID(t *testing.T) {
	t.Parallel()

	handler := newProductProductTypeTestHandler(&productProductTypeServiceStub{})
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/admin/products/not-a-uuid/product-types",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	assertErrorCode(t, response, "invalid_request")
}

func TestProductProductTypeRouteMapsConflict(t *testing.T) {
	t.Parallel()

	handler := newProductProductTypeTestHandler(&productProductTypeServiceStub{
		create: func(
			context.Context,
			productproducttypes.CreateProductProductTypeInput,
		) (productproducttypes.ProductProductType, error) {
			return productproducttypes.ProductProductType{}, productproducttypes.ErrConflict
		},
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/products/"+testProductID+"/product-types",
		bytes.NewBufferString(`{"productTypeId":"`+testProductProductTypeID+`"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusConflict, response.Body)
	}
	assertErrorCode(t, response, "conflict")
}

func TestProductProductTypeRouteMapsMissingReference(t *testing.T) {
	t.Parallel()

	handler := newProductProductTypeTestHandler(&productProductTypeServiceStub{
		list: func(
			context.Context,
			string,
		) ([]productproducttypes.ProductProductType, error) {
			return nil, productproducttypes.ErrReferenceNotFound
		},
	})
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/admin/products/"+testProductID+"/product-types",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	assertErrorCode(t, response, "not_found")
}

func newProductProductTypeTestHandler(service *productProductTypeServiceStub) http.Handler {
	handler, err := newHandler(config.Config{
		Environment:       "test",
		AuthMode:          auth.ModeDevelopment,
		CORSAllowedOrigin: "http://localhost:4321",
	}, dependencies{productProductTypes: service}, nil)
	if err != nil {
		panic(err)
	}
	return handler
}

type productProductTypeServiceStub struct {
	list   func(context.Context, string) ([]productproducttypes.ProductProductType, error)
	create func(context.Context, productproducttypes.CreateProductProductTypeInput) (productproducttypes.ProductProductType, error)
	delete func(context.Context, string, string) error
}

func (s *productProductTypeServiceStub) GetProductProductTypes(
	ctx context.Context,
	productID string,
) ([]productproducttypes.ProductProductType, error) {
	if s.list == nil {
		return nil, errors.New("unexpected GetProductProductTypes call")
	}
	return s.list(ctx, productID)
}

func (s *productProductTypeServiceStub) CreateProductProductType(
	ctx context.Context,
	input productproducttypes.CreateProductProductTypeInput,
) (productproducttypes.ProductProductType, error) {
	if s.create == nil {
		return productproducttypes.ProductProductType{}, errors.New(
			"unexpected CreateProductProductType call",
		)
	}
	return s.create(ctx, input)
}

func (s *productProductTypeServiceStub) DeleteProductProductType(
	ctx context.Context,
	productID string,
	productTypeID string,
) error {
	if s.delete == nil {
		return errors.New("unexpected DeleteProductProductType call")
	}
	return s.delete(ctx, productID, productTypeID)
}
