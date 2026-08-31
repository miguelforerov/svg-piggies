package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
	"github.com/zdenaforero/svg-piggies/backend/internal/producttypes"
)

const testProductTypeID = "d9279e25-3a9b-4c76-b822-a0819a23a47a"

func TestProductTypeRoutes(t *testing.T) {
	service := &productTypeServiceStub{
		list: func(context.Context) ([]producttypes.ProductType, error) {
			return []producttypes.ProductType{{
				ID:          testProductTypeID,
				Name:        "Animals",
				Slug:        "animals",
				Description: "Animal illustrations",
			}}, nil
		},
		get: func(_ context.Context, id string) (producttypes.ProductType, error) {
			if id != testProductTypeID {
				t.Fatalf("GetProductType() id = %q", id)
			}
			return producttypes.ProductType{ID: id, Name: "Animals", Slug: "animals"}, nil
		},
		create: func(
			_ context.Context,
			input producttypes.CreateProductTypeInput,
		) (producttypes.ProductType, error) {
			want := producttypes.CreateProductTypeInput{
				Name:        "Animals",
				Slug:        "animals",
				Description: "Animal illustrations",
			}
			if !reflect.DeepEqual(input, want) {
				t.Fatalf("CreateProductType() input = %#v, want %#v", input, want)
			}
			return producttypes.ProductType{
				ID:          testProductTypeID,
				Name:        input.Name,
				Slug:        input.Slug,
				Description: input.Description,
			}, nil
		},
		update: func(
			_ context.Context,
			id string,
			input producttypes.UpdateProductTypeInput,
		) (producttypes.ProductType, error) {
			if id != testProductTypeID {
				t.Fatalf("UpdateProductType() id = %q", id)
			}
			return producttypes.ProductType{
				ID:          id,
				Name:        input.Name,
				Slug:        input.Slug,
				Description: input.Description,
			}, nil
		},
		delete: func(_ context.Context, id string) error {
			if id != testProductTypeID {
				t.Fatalf("DeleteProductType() id = %q", id)
			}
			return nil
		},
	}
	handler := newProductTypeTestHandler(service)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{name: "list", method: http.MethodGet, path: "/api/admin/product-types", wantStatus: http.StatusOK},
		{name: "get", method: http.MethodGet, path: "/api/admin/product-types/" + testProductTypeID, wantStatus: http.StatusOK},
		{
			name:       "create",
			method:     http.MethodPost,
			path:       "/api/admin/product-types",
			body:       `{"name":"Animals","slug":"animals","description":"Animal illustrations"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "update",
			method:     http.MethodPut,
			path:       "/api/admin/product-types/" + testProductTypeID,
			body:       `{"name":"Wildlife","slug":"wildlife"}`,
			wantStatus: http.StatusOK,
		},
		{name: "delete", method: http.MethodDelete, path: "/api/admin/product-types/" + testProductTypeID, wantStatus: http.StatusNoContent},
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

func TestProductTypeRouteRejectsInvalidUUID(t *testing.T) {
	t.Parallel()

	handler := newProductTypeTestHandler(&productTypeServiceStub{})
	request := httptest.NewRequest(http.MethodGet, "/api/admin/product-types/not-a-uuid", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	assertErrorCode(t, response, "invalid_request")
}

func TestProductTypeRouteMapsNotFound(t *testing.T) {
	t.Parallel()

	handler := newProductTypeTestHandler(&productTypeServiceStub{
		get: func(context.Context, string) (producttypes.ProductType, error) {
			return producttypes.ProductType{}, producttypes.ErrNotFound
		},
	})
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/admin/product-types/"+testProductTypeID,
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	assertErrorCode(t, response, "not_found")
}

func newProductTypeTestHandler(service *productTypeServiceStub) http.Handler {
	handler, err := newHandler(config.Config{
		Environment:       "test",
		AuthMode:          auth.ModeDevelopment,
		CORSAllowedOrigin: "http://localhost:4321",
	}, dependencies{productTypes: service}, nil)
	if err != nil {
		panic(err)
	}
	return handler
}

type productTypeServiceStub struct {
	list   func(context.Context) ([]producttypes.ProductType, error)
	get    func(context.Context, string) (producttypes.ProductType, error)
	create func(context.Context, producttypes.CreateProductTypeInput) (producttypes.ProductType, error)
	update func(context.Context, string, producttypes.UpdateProductTypeInput) (producttypes.ProductType, error)
	delete func(context.Context, string) error
}

func (s *productTypeServiceStub) GetProductTypes(
	ctx context.Context,
) ([]producttypes.ProductType, error) {
	if s.list == nil {
		return nil, errors.New("unexpected GetProductTypes call")
	}
	return s.list(ctx)
}

func (s *productTypeServiceStub) GetProductType(
	ctx context.Context,
	id string,
) (producttypes.ProductType, error) {
	if s.get == nil {
		return producttypes.ProductType{}, errors.New("unexpected GetProductType call")
	}
	return s.get(ctx, id)
}

func (s *productTypeServiceStub) CreateProductType(
	ctx context.Context,
	input producttypes.CreateProductTypeInput,
) (producttypes.ProductType, error) {
	if s.create == nil {
		return producttypes.ProductType{}, errors.New("unexpected CreateProductType call")
	}
	return s.create(ctx, input)
}

func (s *productTypeServiceStub) UpdateProductType(
	ctx context.Context,
	id string,
	input producttypes.UpdateProductTypeInput,
) (producttypes.ProductType, error) {
	if s.update == nil {
		return producttypes.ProductType{}, errors.New("unexpected UpdateProductType call")
	}
	return s.update(ctx, id, input)
}

func (s *productTypeServiceStub) DeleteProductType(ctx context.Context, id string) error {
	if s.delete == nil {
		return errors.New("unexpected DeleteProductType call")
	}
	return s.delete(ctx, id)
}
