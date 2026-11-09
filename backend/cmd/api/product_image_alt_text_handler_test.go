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

type imageAltTextServiceStub struct {
	input productimages.AltTextUpdate
	calls int
	err   error
}

func (s *imageAltTextServiceStub) UpdateAltText(_ context.Context, input productimages.AltTextUpdate) error {
	s.input = input
	s.calls++
	return s.err
}

func altTextTestHandler(t *testing.T, service *imageAltTextServiceStub, mode auth.Mode) http.Handler {
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
	handler, err := newHandler(cfg, dependencies{productImageAltText: service}, verifier)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func altTextPath() string {
	return "/api/admin/products/" + testProductID + "/images/" + testProductImageID
}

func altTextRequest(body string) *http.Request {
	request := httptest.NewRequest(http.MethodPatch, altTextPath(), strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	return request
}

func TestUpdateProductImageAltTextRoutePreservesPresence(t *testing.T) {
	for _, tc := range []struct {
		name, body      string
		present, isNull bool
		value           string
	}{
		{name: "omitted", body: `{}`},
		{name: "null", body: `{"altText":null}`, present: true, isNull: true},
		{name: "empty", body: `{"altText":""}`, present: true, value: ""},
		{name: "text", body: `{"altText":"Pink pig","objectKey":"ignored"}`, present: true, value: "Pink pig"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &imageAltTextServiceStub{}
			response := httptest.NewRecorder()
			altTextTestHandler(t, service, auth.ModeDevelopment).ServeHTTP(response, altTextRequest(tc.body))
			if response.Code != http.StatusNoContent || response.Body.Len() != 0 {
				t.Fatalf("response = %d %s", response.Code, response.Body.String())
			}
			if service.calls != 1 || service.input.ProductID != testProductID || service.input.ImageID != testProductImageID || service.input.Present != tc.present {
				t.Fatalf("service input = %+v", service.input)
			}
			if tc.present && !tc.isNull && (service.input.AltText == nil || *service.input.AltText != tc.value) {
				t.Fatalf("alt text = %v, want %q", service.input.AltText, tc.value)
			}
			if (!tc.present || tc.isNull) && service.input.AltText != nil {
				t.Fatalf("alt text should be nil: %v", *service.input.AltText)
			}
		})
	}
}

func TestUpdateProductImageAltTextRouteErrors(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		err        error
		status     int
	}{
		{name: "missing body", body: ``, status: 400},
		{name: "wrong type", body: `{"altText":42}`, status: 400},
		{name: "too long", body: `{"altText":"text"}`, err: productimages.ErrInvalidInput, status: 400},
		{name: "missing product", body: `{}`, err: productimages.ErrReferenceNotFound, status: 404},
		{name: "missing or other product image", body: `{}`, err: productimages.ErrImageNotFound, status: 404},
		{name: "database error", body: `{}`, err: errors.New("secret database detail"), status: 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &imageAltTextServiceStub{err: tc.err}
			response := httptest.NewRecorder()
			altTextTestHandler(t, service, auth.ModeDevelopment).ServeHTTP(response, altTextRequest(tc.body))
			if response.Code != tc.status {
				t.Fatalf("response = %d %s, want %d", response.Code, response.Body.String(), tc.status)
			}
			if strings.Contains(response.Body.String(), "secret") {
				t.Fatalf("internal detail leaked: %s", response.Body.String())
			}
		})
	}
}

func TestUpdateProductImageAltTextRouteRequiresAdmin(t *testing.T) {
	service := &imageAltTextServiceStub{}
	response := httptest.NewRecorder()
	altTextTestHandler(t, service, auth.ModeCloudflareAccess).ServeHTTP(response, altTextRequest(`{"altText":"new"}`))
	if response.Code != http.StatusUnauthorized || service.calls != 0 {
		t.Fatalf("unauthenticated response = %d; calls = %d", response.Code, service.calls)
	}
	cfg := config.Config{Environment: "production", AuthMode: auth.ModeCloudflareAccess, CloudflareAccessTeamDomain: "https://example.cloudflareaccess.com", CloudflareAccessAudience: "test-audience", AdminEmails: []string{"admin@example.com"}}
	nonAdmin, err := newHandler(cfg, dependencies{productImageAltText: service}, imageIdentityVerifier{email: "customer@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	request := altTextRequest(`{"altText":"new"}`)
	request.Header.Set(auth.AccessJWTHeader, "verified-token")
	response = httptest.NewRecorder()
	nonAdmin.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || service.calls != 0 {
		t.Fatalf("non-admin response = %d; calls = %d", response.Code, service.calls)
	}
}
