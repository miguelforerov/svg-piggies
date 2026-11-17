package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
	"github.com/zdenaforero/svg-piggies/backend/internal/productfiles"
)

type fileServiceStub struct {
	file  productfiles.ProductFile
	err   error
	calls int
}

func (s *fileServiceStub) RequestUpload(context.Context, productfiles.RequestUploadInput) (productfiles.UploadAuthorization, error) {
	s.calls++
	return productfiles.UploadAuthorization{UploadToken: "token", UploadURL: "https://r2.test/upload", Headers: map[string][]string{"If-None-Match": {"*"}}, ExpiresAt: time.Unix(1700000000, 0)}, s.err
}
func (s *fileServiceStub) ConfirmUpload(context.Context, productfiles.ConfirmUploadInput) (productfiles.ProductFile, error) {
	s.calls++
	return s.file, s.err
}
func (s *fileServiceStub) List(context.Context, string) ([]productfiles.ProductFile, error) {
	s.calls++
	return []productfiles.ProductFile{s.file}, s.err
}
func (s *fileServiceStub) UpdateMetadata(context.Context, productfiles.UpdateMetadataInput) (productfiles.ProductFile, error) {
	s.calls++
	return s.file, s.err
}
func (s *fileServiceStub) Delete(context.Context, string, string) error    { s.calls++; return s.err }
func (s *fileServiceStub) Reorder(context.Context, string, []string) error { s.calls++; return s.err }

const fileProductID = "b6140bec-35f1-4e68-b54c-d4c8070bb987"
const fileID = "a6140bec-35f1-4e68-b54c-d4c8070bb987"

func fileTestHandler(t *testing.T, service *fileServiceStub) http.Handler {
	t.Helper()
	h, err := newHandler(config.Config{Environment: "test", AuthMode: auth.ModeDevelopment, CORSAllowedOrigin: "http://localhost:4321"}, dependencies{productFiles: service}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestProductFileRoutesAndPrivateResponse(t *testing.T) {
	service := &fileServiceStub{file: productfiles.ProductFile{ID: fileID, ProductID: fileProductID, FileKey: "private/secret", FileName: "pig.svg", DisplayName: "Pig", Format: "svg", ContentType: "image/svg+xml", FileSizeBytes: 123, AvailabilityStatus: "active", ConfirmationFingerprint: strings.Repeat("a", 64), FingerprintKeyVersion: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}}
	h := fileTestHandler(t, service)
	base := "/api/admin/products/" + fileProductID + "/files"
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{{"GET", base, "", 200}, {"POST", base + "/upload-authorization", `{"fileName":"pig.svg","format":"svg","contentType":"image/svg+xml","fileSizeBytes":123}`, 201}, {"POST", base, `{"uploadToken":"token"}`, 201}, {"PATCH", base + "/" + fileID, `{"displayName":"Pig","displayOrder":0}`, 200}, {"PUT", base + "/order", `{"fileIds":["` + fileID + `"]}`, 204}, {"DELETE", base + "/" + fileID, "", 204}} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		h.ServeHTTP(response, request)
		if response.Code != tc.status {
			t.Fatalf("%s %s: status %d, body %s", tc.method, tc.path, response.Code, response.Body.String())
		}
		if tc.method == "GET" && (strings.Contains(response.Body.String(), "private/secret") || strings.Contains(response.Body.String(), service.file.ConfirmationFingerprint)) {
			t.Fatal("private fields leaked")
		}
	}
	if service.calls != 6 {
		t.Fatalf("service calls = %d", service.calls)
	}
}

func TestProductFileRouteErrorMapping(t *testing.T) {
	base := "/api/admin/products/" + fileProductID + "/files"
	for _, tc := range []struct {
		err        error
		path, body string
		status     int
	}{{productfiles.ErrInvalidUpload, base + "/upload-authorization", `{"fileName":"pig.svg","format":"svg","contentType":"image/svg+xml","fileSizeBytes":123}`, 400}, {productfiles.ErrReferenceNotFound, base + "/upload-authorization", `{"fileName":"pig.svg","format":"svg","contentType":"image/svg+xml","fileSizeBytes":123}`, 404}, {productfiles.ErrStorageFailure, base + "/upload-authorization", `{"fileName":"pig.svg","format":"svg","contentType":"image/svg+xml","fileSizeBytes":123}`, 502}, {productfiles.ErrUploadUnavailable, base + "/upload-authorization", `{"fileName":"pig.svg","format":"svg","contentType":"image/svg+xml","fileSizeBytes":123}`, 503}, {productfiles.ErrExpiredUpload, base, `{"uploadToken":"token"}`, 410}, {productfiles.ErrConflict, base, `{"uploadToken":"token"}`, 409}} {
		service := &fileServiceStub{err: tc.err}
		h := fileTestHandler(t, service)
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
		request.Header.Set("Content-Type", "application/json")
		h.ServeHTTP(response, request)
		if response.Code != tc.status {
			t.Fatalf("%v: status %d, body %s", tc.err, response.Code, response.Body.String())
		}
	}
}

func TestProductFileRoutesRequireAdmin(t *testing.T) {
	cfg := config.Config{Environment: "production", AuthMode: auth.ModeCloudflareAccess, CloudflareAccessTeamDomain: "https://example.cloudflareaccess.com", CloudflareAccessAudience: "test-audience", AdminEmails: []string{"admin@example.com"}}
	service := &fileServiceStub{}
	h, err := newHandler(cfg, dependencies{productFiles: service}, allowImageUploadVerifier{})
	if err != nil {
		t.Fatal(err)
	}
	base := "/api/admin/products/" + fileProductID + "/files"
	for _, tc := range []struct{ method, path string }{{http.MethodGet, base}, {http.MethodPost, base}, {http.MethodPost, base + "/upload-authorization"}, {http.MethodPut, base + "/order"}, {http.MethodPatch, base + "/" + fileID}, {http.MethodDelete, base + "/" + fileID}} {
		response := httptest.NewRecorder()
		h.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, nil))
		if response.Code != http.StatusUnauthorized || service.calls != 0 {
			t.Fatalf("%s %s: status %d, calls %d", tc.method, tc.path, response.Code, service.calls)
		}
	}
}
