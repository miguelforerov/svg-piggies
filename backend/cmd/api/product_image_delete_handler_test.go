package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
)

type imageDeleteServiceStub struct {
	productID string
	imageID   string
	calls     int
	err       error
}

func (s *imageDeleteServiceStub) DeleteProductImage(_ context.Context, productID, imageID string) error {
	s.productID, s.imageID = productID, imageID
	s.calls++
	return s.err
}

func deleteImageTestHandler(t *testing.T, service *imageDeleteServiceStub, mode auth.Mode) http.Handler {
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
	handler, err := newHandler(cfg, dependencies{productImageDeletes: service}, verifier)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func TestDeleteProductImageRoute(t *testing.T) {
	service := &imageDeleteServiceStub{}
	handler := deleteImageTestHandler(t, service, auth.ModeDevelopment)
	request := httptest.NewRequest(http.MethodDelete, "/api/admin/products/"+testProductID+"/images/"+testProductImageID, strings.NewReader(`{"objectKey":"products/another-product/secret"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || response.Body.Len() != 0 {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
	if service.calls != 1 || service.productID != testProductID || service.imageID != testProductImageID {
		t.Fatalf("delete call = %+v", service)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("admin deletion must not be cached")
	}
}

func TestDeleteProductImageRouteErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{"missing product", productimages.ErrReferenceNotFound, 404},
		{"missing image or other product", productimages.ErrImageNotFound, 404},
		{"R2 failed", errors.Join(productimages.ErrStorageFailure, errors.New("secret R2 detail")), 502},
		{"DB failed after R2", errors.New("secret DB detail"), 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := deleteImageTestHandler(t, &imageDeleteServiceStub{err: tc.err}, auth.ModeDevelopment)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/api/admin/products/"+testProductID+"/images/"+testProductImageID, nil))
			if response.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tc.status, response.Body.String())
			}
			if strings.Contains(response.Body.String(), "secret") {
				t.Fatalf("internal detail leaked: %s", response.Body.String())
			}
		})
	}
}

func TestDeleteProductImageRouteRequiresAdmin(t *testing.T) {
	service := &imageDeleteServiceStub{}
	handler := deleteImageTestHandler(t, service, auth.ModeCloudflareAccess)
	path := "/api/admin/products/" + testProductID + "/images/" + testProductImageID
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, path, nil))
	if response.Code != http.StatusUnauthorized || service.calls != 0 {
		t.Fatalf("unauthenticated response = %d; calls = %d", response.Code, service.calls)
	}
	cfg := config.Config{
		Environment: "production", AuthMode: auth.ModeCloudflareAccess,
		CloudflareAccessTeamDomain: "https://example.cloudflareaccess.com",
		CloudflareAccessAudience:   "test-audience",
		AdminEmails:                []string{"admin@example.com"},
	}
	nonAdminHandler, err := newHandler(cfg, dependencies{productImageDeletes: service}, imageIdentityVerifier{email: "customer@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodDelete, path, nil)
	request.Header.Set(auth.AccessJWTHeader, "verified-token")
	response = httptest.NewRecorder()
	nonAdminHandler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || service.calls != 0 {
		t.Fatalf("non-admin response = %d; calls = %d", response.Code, service.calls)
	}
}
