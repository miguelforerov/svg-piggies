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

type imageOrderServiceStub struct {
	ids   []string
	calls int
	err   error
}

func (s *imageOrderServiceStub) ReorderImages(_ context.Context, _ string, ids []string) error {
	s.calls++
	s.ids = ids
	return s.err
}
func orderHandler(t *testing.T, service *imageOrderServiceStub, mode auth.Mode) http.Handler {
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
	h, err := newHandler(cfg, dependencies{productImageOrder: service}, verifier)
	if err != nil {
		t.Fatal(err)
	}
	return h
}
func orderPath() string { return "/api/admin/products/" + testProductID + "/images/order" }
func TestReorderProductImagesRoute(t *testing.T) {
	s := &imageOrderServiceStub{}
	h := orderHandler(t, s, auth.ModeDevelopment)
	for _, body := range []string{`{"imageIds":["` + testProductImageID + `"],"objectKey":"untrusted"}`, `{"imageIds":[]}`} {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest(http.MethodPut, orderPath(), strings.NewReader(body)))
		if r.Code != 204 {
			t.Fatalf("status=%d body=%s", r.Code, r.Body.String())
		}
	}
	if s.calls != 2 || len(s.ids) != 0 {
		t.Fatalf("calls=%d ids=%v", s.calls, s.ids)
	}
}
func TestReorderProductImagesRouteErrors(t *testing.T) {
	for _, tc := range []struct {
		body   string
		err    error
		status int
	}{
		{`{}`, nil, 400}, {`{"imageIds":["` + testProductImageID + `","` + testProductImageID + `"]}`, productimages.ErrInvalidInput, 400},
		{`{"imageIds":[]}`, productimages.ErrReferenceNotFound, 404}, {`{"imageIds":[]}`, errors.New("secret database detail"), 500},
	} {
		s := &imageOrderServiceStub{err: tc.err}
		r := httptest.NewRecorder()
		orderHandler(t, s, auth.ModeDevelopment).ServeHTTP(r, httptest.NewRequest(http.MethodPut, orderPath(), strings.NewReader(tc.body)))
		if r.Code != tc.status || strings.Contains(r.Body.String(), "secret") {
			t.Fatalf("status=%d body=%s", r.Code, r.Body.String())
		}
	}
}
func TestReorderProductImagesRequiresAdmin(t *testing.T) {
	s := &imageOrderServiceStub{}
	r := httptest.NewRecorder()
	orderHandler(t, s, auth.ModeCloudflareAccess).ServeHTTP(r, httptest.NewRequest(http.MethodPut, orderPath(), strings.NewReader(`{"imageIds":[]}`)))
	if r.Code != 401 || s.calls != 0 {
		t.Fatalf("status=%d calls=%d", r.Code, s.calls)
	}
}
