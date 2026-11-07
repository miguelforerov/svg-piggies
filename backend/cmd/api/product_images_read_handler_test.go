package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
)

type imageReadServiceStub struct {
	images []productimages.ProductImageWithURL
	err    error
}

func (s imageReadServiceStub) GetProductImages(_ context.Context, _ string) ([]productimages.ProductImageWithURL, error) {
	return s.images, s.err
}

func newImageReadTestHandler(t *testing.T, service imageReadServiceStub) http.Handler {
	t.Helper()
	handler, err := newHandler(config.Config{Environment: "test", AuthMode: auth.ModeDevelopment, CORSAllowedOrigin: "http://localhost:4321"}, dependencies{productImageReads: service}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func imageReadRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/api/admin/products/"+testProductID+"/images", nil)
}

func TestGetProductImagesRoute(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	image := productimages.ProductImage{ID: testProductImageID, ProductID: testProductID, R2ObjectKey: "private/r2-key", OriginalFilename: "pig.webp", ContentType: "image/webp", FileSizeBytes: 123, IsPrimary: true, CreatedAt: now, UpdatedAt: now}
	for _, tt := range []struct {
		name   string
		images []productimages.ProductImageWithURL
		count  int
	}{
		{name: "image", images: []productimages.ProductImageWithURL{{ProductImage: image, ImageURL: "https://example.test/signed"}}, count: 1},
		{name: "empty", count: 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			newImageReadTestHandler(t, imageReadServiceStub{images: tt.images}).ServeHTTP(response, imageReadRequest())
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d: %s", response.Code, response.Body)
			}
			if strings.Contains(response.Body.String(), "r2ObjectKey") || strings.Contains(response.Body.String(), "private/r2-key") {
				t.Fatalf("object key leaked: %s", response.Body)
			}
			var body []map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body == nil || len(body) != tt.count {
				t.Fatalf("body = %s", response.Body)
			}
			if tt.count == 1 && (body[0]["imageUrl"] != "https://example.test/signed" || body[0]["id"] != testProductImageID) {
				t.Fatalf("body = %s", response.Body)
			}
		})
	}
}

func TestGetProductImagesRouteErrors(t *testing.T) {
	for _, tt := range []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"invalid", productimages.ErrInvalidInput, 400, "invalid_request"},
		{"missing product", productimages.ErrReferenceNotFound, 404, "not_found"},
		{"storage", fmt.Errorf("%w: secret storage detail", productimages.ErrStorageFailure), 502, "internal_error"},
		{"unexpected", errors.New("secret database detail"), 500, "internal_error"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			newImageReadTestHandler(t, imageReadServiceStub{err: tt.err}).ServeHTTP(response, imageReadRequest())
			if response.Code != tt.status {
				t.Fatalf("status = %d: %s", response.Code, response.Body)
			}
			if strings.Contains(response.Body.String(), "secret") {
				t.Fatalf("internal detail leaked: %s", response.Body)
			}
			var body struct {
				Code string `json:"code"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code != tt.code {
				t.Fatalf("code = %q, want %q", body.Code, tt.code)
			}
		})
	}
}
