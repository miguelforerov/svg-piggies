package main

import (
	"bytes"
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

const testProductImageID = "b6140bec-35f1-4e68-b54c-d4c8070bb987"

type imageUploadServiceStub struct {
	request func(context.Context, productimages.RequestUploadInput) (productimages.UploadAuthorization, error)
	confirm func(context.Context, productimages.ConfirmUploadInput) (productimages.ProductImage, error)
}

func (s *imageUploadServiceStub) RequestUpload(ctx context.Context, input productimages.RequestUploadInput) (productimages.UploadAuthorization, error) {
	if s.request == nil {
		return productimages.UploadAuthorization{}, errors.New("unexpected request")
	}
	return s.request(ctx, input)
}
func (s *imageUploadServiceStub) ConfirmUpload(ctx context.Context, input productimages.ConfirmUploadInput) (productimages.ProductImage, error) {
	if s.confirm == nil {
		return productimages.ProductImage{}, errors.New("unexpected confirmation")
	}
	return s.confirm(ctx, input)
}

func newImageUploadTestHandler(t *testing.T, service *imageUploadServiceStub, mode auth.Mode) http.Handler {
	t.Helper()
	cfg := config.Config{Environment: "test", AuthMode: mode, CORSAllowedOrigin: "http://localhost:4321"}
	var verifier auth.TokenVerifier
	if mode == auth.ModeCloudflareAccess {
		cfg.Environment = "production"
		cfg.CloudflareAccessTeamDomain = "https://example.cloudflareaccess.com"
		cfg.CloudflareAccessAudience = "test-audience"
		cfg.AdminEmails = []string{"admin@example.com"}
		verifier = allowImageUploadVerifier{}
	}
	handler, err := newHandler(cfg, dependencies{productImageUploads: service}, verifier)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

type allowImageUploadVerifier struct{}

func (allowImageUploadVerifier) Verify(context.Context, string) (string, error) {
	return "admin@example.com", nil
}

func imageUploadRequest(path, body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func TestImageUploadRoutes(t *testing.T) {
	expires := time.Date(2026, time.November, 4, 12, 10, 0, 0, time.UTC)
	service := &imageUploadServiceStub{
		request: func(_ context.Context, input productimages.RequestUploadInput) (productimages.UploadAuthorization, error) {
			if input.ProductID != testProductID || input.OriginalFilename != "party.webp" || input.ContentType != "image/webp" || input.FileSizeBytes != 4096 {
				t.Fatalf("request input = %+v", input)
			}
			return productimages.UploadAuthorization{UploadToken: "signed-token", ObjectKey: "products/server-key.webp", UploadURL: "https://example.test/upload", Headers: http.Header{"Content-Type": {"image/webp"}, "If-None-Match": {"*"}}, ExpiresAt: expires}, nil
		},
		confirm: func(_ context.Context, input productimages.ConfirmUploadInput) (productimages.ProductImage, error) {
			if input.ProductID != testProductID || input.UploadToken != "signed-token" || input.DisplayOrder != 2 || !input.IsPrimary {
				t.Fatalf("confirm input = %+v", input)
			}
			return productimages.ProductImage{ID: testProductImageID, ProductID: input.ProductID, R2ObjectKey: "products/server-key.webp", OriginalFilename: "party.webp", ContentType: "image/webp", FileSizeBytes: 4096, DisplayOrder: 2, IsPrimary: true, CreatedAt: expires, UpdatedAt: expires}, nil
		},
	}
	handler := newImageUploadTestHandler(t, service, auth.ModeDevelopment)
	start := httptest.NewRecorder()
	handler.ServeHTTP(start, imageUploadRequest("/api/admin/products/"+testProductID+"/images/uploads", `{"originalFilename":"party.webp","contentType":"image/webp","fileSizeBytes":4096}`))
	if start.Code != http.StatusCreated {
		t.Fatalf("start status = %d: %s", start.Code, start.Body)
	}
	if start.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("upload authorization must not be cached")
	}
	var authorization struct {
		UploadToken string      `json:"uploadToken"`
		ObjectKey   string      `json:"objectKey"`
		Headers     http.Header `json:"headers"`
	}
	if err := json.Unmarshal(start.Body.Bytes(), &authorization); err != nil {
		t.Fatal(err)
	}
	if authorization.UploadToken != "signed-token" || authorization.ObjectKey != "products/server-key.webp" || authorization.Headers.Get("If-None-Match") != "*" {
		t.Fatalf("authorization = %+v", authorization)
	}

	confirm := httptest.NewRecorder()
	handler.ServeHTTP(confirm, imageUploadRequest("/api/admin/products/"+testProductID+"/images", `{"uploadToken":"signed-token","displayOrder":2,"isPrimary":true}`))
	if confirm.Code != http.StatusCreated {
		t.Fatalf("confirm status = %d: %s", confirm.Code, confirm.Body)
	}
	var image struct {
		ID            string `json:"id"`
		FileSizeBytes int64  `json:"fileSizeBytes"`
	}
	if err := json.Unmarshal(confirm.Body.Bytes(), &image); err != nil {
		t.Fatal(err)
	}
	if image.ID != testProductImageID || image.FileSizeBytes != 4096 {
		t.Fatalf("image = %+v", image)
	}
}

func TestImageUploadRoutesRequireAdmin(t *testing.T) {
	handler := newImageUploadTestHandler(t, &imageUploadServiceStub{}, auth.ModeCloudflareAccess)
	for _, path := range []string{"/api/admin/products/" + testProductID + "/images/uploads", "/api/admin/products/" + testProductID + "/images"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, imageUploadRequest(path, `{}`))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s status = %d", path, response.Code)
		}
	}
}

type imageIdentityVerifier struct{ email string }

func (v imageIdentityVerifier) Verify(context.Context, string) (string, error) { return v.email, nil }

func TestImageUploadRoutesRejectAuthenticatedNonAdmin(t *testing.T) {
	cfg := config.Config{
		Environment: "production", AuthMode: auth.ModeCloudflareAccess,
		CloudflareAccessTeamDomain: "https://example.cloudflareaccess.com",
		CloudflareAccessAudience:   "test-audience",
		AdminEmails:                []string{"admin@example.com"},
	}
	handler, err := newHandler(cfg, dependencies{productImageUploads: &imageUploadServiceStub{}}, imageIdentityVerifier{email: "customer@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/admin/products/" + testProductID + "/images/uploads", "/api/admin/products/" + testProductID + "/images"} {
		request := imageUploadRequest(path, `{}`)
		request.Header.Set(auth.AccessJWTHeader, "verified-token")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden {
			t.Fatalf("%s status = %d, want 403", path, response.Code)
		}
	}
}

func TestImageUploadRouteErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{"invalid", productimages.ErrInvalidUpload, 400},
		{"missing product", productimages.ErrReferenceNotFound, 404},
		{"storage", productimages.ErrStorageFailure, 502},
		{"schema unavailable", fmt.Errorf("%w: column confirmation_fingerprint does not exist", productimages.ErrUploadUnavailable), 503},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newImageUploadTestHandler(t, &imageUploadServiceStub{request: func(context.Context, productimages.RequestUploadInput) (productimages.UploadAuthorization, error) {
				return productimages.UploadAuthorization{}, tt.err
			}}, auth.ModeDevelopment)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, imageUploadRequest("/api/admin/products/"+testProductID+"/images/uploads", `{"originalFilename":"a.webp","contentType":"image/webp","fileSizeBytes":1}`))
			if response.Code != tt.status {
				t.Fatalf("status = %d: %s", response.Code, response.Body)
			}
			if strings.Contains(response.Body.String(), "confirmation_fingerprint") {
				t.Fatalf("database detail leaked: %s", response.Body)
			}
		})
	}
	for _, tt := range []struct {
		name    string
		err     error
		status  int
		code    string
		message string
	}{
		{"invalid token", productimages.ErrInvalidUpload, http.StatusBadRequest, "invalid_request", "invalid image upload"},
		{"invalid input", productimages.ErrInvalidInput, http.StatusBadRequest, "invalid_request", "invalid product image input"},
		{"missing product", productimages.ErrReferenceNotFound, http.StatusNotFound, "not_found", "product or uploaded image not found"},
		{"missing object", productimages.ErrObjectNotFound, http.StatusNotFound, "not_found", "product or uploaded image not found"},
		{"object mismatch", productimages.ErrObjectMismatch, http.StatusConflict, "conflict", "image upload could not be confirmed"},
		{"invalid image", productimages.ErrInvalidImage, http.StatusConflict, "conflict", "image upload could not be confirmed"},
		{"real conflict", productimages.ErrConflict, http.StatusConflict, "conflict", "image upload could not be confirmed"},
		{"expired", fmt.Errorf("%w: internal token detail", productimages.ErrExpiredUpload), http.StatusGone, "expired_upload", "image upload authorization has expired"},
		{"storage", fmt.Errorf("%w: internal storage detail", productimages.ErrStorageFailure), http.StatusBadGateway, "internal_error", "image storage is unavailable"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler := newImageUploadTestHandler(t, &imageUploadServiceStub{confirm: func(context.Context, productimages.ConfirmUploadInput) (productimages.ProductImage, error) {
				return productimages.ProductImage{}, tt.err
			}}, auth.ModeDevelopment)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, imageUploadRequest("/api/admin/products/"+testProductID+"/images", `{"uploadToken":"token"}`))
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tt.status, response.Body)
			}
			if response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
			}
			var body struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code != tt.code || body.Message != tt.message {
				t.Fatalf("body = %+v, want code %q and message %q", body, tt.code, tt.message)
			}
			if strings.Contains(response.Body.String(), "internal token detail") || strings.Contains(response.Body.String(), "internal storage detail") {
				t.Fatalf("internal detail leaked: %s", response.Body)
			}
		})
	}
}
