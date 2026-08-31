package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
)

const testProductID = "b2f18e56-dff5-4b1d-a454-60481472179d"

func TestProductRoutes(t *testing.T) {
	timestamp := time.Date(2026, time.August, 31, 10, 0, 0, 0, time.UTC)
	service := &productServiceStub{
		list: func(context.Context) ([]products.Product, error) {
			return []products.Product{{
				ID:        testProductID,
				Title:     "Party Pig",
				Slug:      "party-pig",
				Price:     "19.95",
				Status:    products.StatusActive,
				CreatedAt: timestamp,
				UpdatedAt: timestamp,
			}}, nil
		},
		get: func(_ context.Context, id string) (products.Product, error) {
			if id != testProductID {
				t.Fatalf("GetProduct() id = %q", id)
			}
			return products.Product{
				ID: id, Title: "Party Pig", Slug: "party-pig", Price: "19.95",
				Status: products.StatusActive, CreatedAt: timestamp, UpdatedAt: timestamp,
			}, nil
		},
		create: func(
			_ context.Context,
			input products.CreateProductInput,
		) (products.Product, error) {
			want := products.CreateProductInput{
				Title:       "Party Pig",
				Slug:        "party-pig",
				Description: "Printable party pig",
				Price:       "19.95",
				Status:      products.StatusActive,
			}
			if !reflect.DeepEqual(input, want) {
				t.Fatalf("CreateProduct() input = %#v, want %#v", input, want)
			}
			return products.Product{
				ID: testProductID, Title: input.Title, Slug: input.Slug,
				Description: input.Description, Price: input.Price, Status: input.Status,
				CreatedAt: timestamp, UpdatedAt: timestamp,
			}, nil
		},
		update: func(
			_ context.Context,
			id string,
			input products.UpdateProductInput,
		) (products.Product, error) {
			if id != testProductID {
				t.Fatalf("UpdateProduct() id = %q", id)
			}
			return products.Product{
				ID: id, Title: input.Title, Slug: input.Slug,
				Description: input.Description, Price: input.Price, Status: input.Status,
				CreatedAt: timestamp, UpdatedAt: timestamp,
			}, nil
		},
		delete: func(_ context.Context, id string) error {
			if id != testProductID {
				t.Fatalf("DeleteProduct() id = %q", id)
			}
			return nil
		},
	}
	handler := newProductTestHandler(service)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{name: "list", method: http.MethodGet, path: "/api/admin/products", wantStatus: http.StatusOK},
		{name: "get", method: http.MethodGet, path: "/api/admin/products/" + testProductID, wantStatus: http.StatusOK},
		{
			name:   "create",
			method: http.MethodPost,
			path:   "/api/admin/products",
			body: `{"title":"Party Pig","slug":"party-pig","description":"Printable party pig",` +
				`"price":"19.95","status":"active"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:   "update",
			method: http.MethodPut,
			path:   "/api/admin/products/" + testProductID,
			body: `{"title":"Celebration Pig","slug":"celebration-pig",` +
				`"price":"24.50","status":"archived"}`,
			wantStatus: http.StatusOK,
		},
		{name: "delete", method: http.MethodDelete, path: "/api/admin/products/" + testProductID, wantStatus: http.StatusNoContent},
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

func TestProductCreatePassesOmittedStatusToService(t *testing.T) {
	t.Parallel()

	handler := newProductTestHandler(&productServiceStub{
		create: func(
			_ context.Context,
			input products.CreateProductInput,
		) (products.Product, error) {
			if input.Status != "" {
				t.Fatalf("CreateProduct() status = %q, want empty for service default", input.Status)
			}
			now := time.Now()
			return products.Product{
				ID: testProductID, Title: input.Title, Slug: input.Slug, Price: input.Price,
				Status: products.StatusDraft, CreatedAt: now, UpdatedAt: now,
			}, nil
		},
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/products",
		bytes.NewBufferString(`{"title":"Party Pig","slug":"party-pig","price":"19.95"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusCreated, response.Body)
	}
}

func TestProductRouteRejectsInvalidUUID(t *testing.T) {
	t.Parallel()

	handler := newProductTestHandler(&productServiceStub{})
	request := httptest.NewRequest(http.MethodGet, "/api/admin/products/not-a-uuid", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	assertErrorCode(t, response, "invalid_request")
}

func TestProductRouteMapsNotFound(t *testing.T) {
	t.Parallel()

	handler := newProductTestHandler(&productServiceStub{
		get: func(context.Context, string) (products.Product, error) {
			return products.Product{}, products.ErrNotFound
		},
	})
	request := httptest.NewRequest(http.MethodGet, "/api/admin/products/"+testProductID, nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	assertErrorCode(t, response, "not_found")
}

func newProductTestHandler(service *productServiceStub) http.Handler {
	handler, err := newHandler(config.Config{
		Environment:       "test",
		AuthMode:          auth.ModeDevelopment,
		CORSAllowedOrigin: "http://localhost:4321",
	}, dependencies{products: service}, nil)
	if err != nil {
		panic(err)
	}
	return handler
}

type productServiceStub struct {
	list   func(context.Context) ([]products.Product, error)
	get    func(context.Context, string) (products.Product, error)
	create func(context.Context, products.CreateProductInput) (products.Product, error)
	update func(context.Context, string, products.UpdateProductInput) (products.Product, error)
	delete func(context.Context, string) error
}

func (s *productServiceStub) GetProducts(ctx context.Context) ([]products.Product, error) {
	if s.list == nil {
		return nil, errors.New("unexpected GetProducts call")
	}
	return s.list(ctx)
}

func (s *productServiceStub) GetProduct(
	ctx context.Context,
	id string,
) (products.Product, error) {
	if s.get == nil {
		return products.Product{}, errors.New("unexpected GetProduct call")
	}
	return s.get(ctx, id)
}

func (s *productServiceStub) CreateProduct(
	ctx context.Context,
	input products.CreateProductInput,
) (products.Product, error) {
	if s.create == nil {
		return products.Product{}, errors.New("unexpected CreateProduct call")
	}
	return s.create(ctx, input)
}

func (s *productServiceStub) UpdateProduct(
	ctx context.Context,
	id string,
	input products.UpdateProductInput,
) (products.Product, error) {
	if s.update == nil {
		return products.Product{}, errors.New("unexpected UpdateProduct call")
	}
	return s.update(ctx, id, input)
}

func (s *productServiceStub) DeleteProduct(ctx context.Context, id string) error {
	if s.delete == nil {
		return errors.New("unexpected DeleteProduct call")
	}
	return s.delete(ctx, id)
}
