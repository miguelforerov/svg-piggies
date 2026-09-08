package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
)

const testProductImageID = "b6140bec-35f1-4e68-b54c-d4c8070bb987"

func TestAddProductImagesRoute(t *testing.T) {
	timestamp := time.Date(2026, time.September, 3, 10, 0, 0, 0, time.UTC)
	service := &productImageServiceStub{
		add: func(
			_ context.Context,
			input productimages.AddProductImagesInput,
		) ([]productimages.ProductImage, error) {
			if input.ProductID != testProductID || len(input.Images) != 1 {
				t.Fatalf("AddProductImages() input = %#v", input)
			}
			image := input.Images[0]
			if image.R2ObjectKey != "products/party-pig.webp" ||
				image.OriginalFilename != "party-pig.webp" ||
				image.ContentType != "image/webp" ||
				image.FileSizeBytes != 4096 ||
				image.DisplayOrder != 2 || !image.IsPrimary {
				t.Fatalf("AddProductImages() image = %#v", image)
			}
			return []productimages.ProductImage{{
				ID:               testProductImageID,
				ProductID:        input.ProductID,
				R2ObjectKey:      image.R2ObjectKey,
				OriginalFilename: image.OriginalFilename,
				ContentType:      image.ContentType,
				FileSizeBytes:    image.FileSizeBytes,
				DisplayOrder:     image.DisplayOrder,
				IsPrimary:        image.IsPrimary,
				CreatedAt:        timestamp,
				UpdatedAt:        timestamp,
			}}, nil
		},
	}
	handler := newProductImageTestHandler(service)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/products/"+testProductID+"/images",
		bytes.NewBufferString(`{"images":[{"r2ObjectKey":"products/party-pig.webp",`+
			`"originalFilename":"party-pig.webp","contentType":"image/webp",`+
			`"fileSizeBytes":4096,"displayOrder":2,"isPrimary":true}]}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusCreated, response.Body)
	}
	var body []struct {
		ID           string `json:"id"`
		ProductID    string `json:"productId"`
		R2ObjectKey  string `json:"r2ObjectKey"`
		DisplayOrder int    `json:"displayOrder"`
		IsPrimary    bool   `json:"isPrimary"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body) != 1 || body[0].ID != testProductImageID ||
		body[0].ProductID != testProductID || body[0].R2ObjectKey != "products/party-pig.webp" ||
		body[0].DisplayOrder != 2 || !body[0].IsPrimary {
		t.Fatalf("response = %#v", body)
	}
}

func TestAddProductImagesRouteMapsErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
		wantCode   string
	}{
		{name: "invalid", serviceErr: productimages.ErrInvalidInput, wantStatus: 400, wantCode: "invalid_request"},
		{name: "missing product", serviceErr: productimages.ErrReferenceNotFound, wantStatus: 404, wantCode: "not_found"},
		{name: "duplicate key", serviceErr: productimages.ErrConflict, wantStatus: 409, wantCode: "conflict"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := newProductImageTestHandler(&productImageServiceStub{
				add: func(
					context.Context,
					productimages.AddProductImagesInput,
				) ([]productimages.ProductImage, error) {
					return nil, test.serviceErr
				},
			})
			request := httptest.NewRequest(
				http.MethodPost,
				"/api/admin/products/"+testProductID+"/images",
				bytes.NewBufferString(`{"images":[{"r2ObjectKey":"key","originalFilename":"image.webp",`+
					`"contentType":"image/webp","fileSizeBytes":1}]}`),
			)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body)
			}
			assertErrorCode(t, response, test.wantCode)
		})
	}
}

func TestAddProductImagesRouteRejectsMissingBody(t *testing.T) {
	t.Parallel()

	handler := newProductImageTestHandler(&productImageServiceStub{})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/products/"+testProductID+"/images",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func newProductImageTestHandler(service *productImageServiceStub) http.Handler {
	handler, err := newHandler(config.Config{
		Environment:       "test",
		AuthMode:          auth.ModeDevelopment,
		CORSAllowedOrigin: "http://localhost:4321",
	}, dependencies{productImages: service}, nil)
	if err != nil {
		panic(err)
	}
	return handler
}

type productImageServiceStub struct {
	add func(context.Context, productimages.AddProductImagesInput) ([]productimages.ProductImage, error)
}

func (s *productImageServiceStub) AddProductImages(
	ctx context.Context,
	input productimages.AddProductImagesInput,
) ([]productimages.ProductImage, error) {
	if s.add == nil {
		return nil, errors.New("unexpected AddProductImages call")
	}
	return s.add(ctx, input)
}
