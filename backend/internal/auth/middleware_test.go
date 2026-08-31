package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	testIssuer   = "https://svg-piggies.cloudflareaccess.com"
	testAudience = "svg-piggies-admin-audience"
)

func TestDevelopmentModeAllowsAdminRequest(t *testing.T) {
	t.Parallel()

	middleware, err := NewAdminMiddleware(ModeDevelopment, nil)
	if err != nil {
		t.Fatalf("NewAdminMiddleware() error = %v", err)
	}
	handler := middleware.Wrap(successHandler())

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/admin/collections", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestCloudflareAccessModeRejectsMissingJWT(t *testing.T) {
	t.Parallel()

	handler := newCloudflareTestHandler(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/admin/collections", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestCloudflareAccessModeRejectsInvalidJWT(t *testing.T) {
	t.Parallel()

	handler := newCloudflareTestHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/api/admin/collections", nil)
	request.Header.Set(AccessJWTHeader, "not-a-jwt")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestCloudflareAccessModeAcceptsValidJWT(t *testing.T) {
	t.Parallel()

	privateKey := newRSAKey(t)
	handler := newCloudflareTestHandlerWithKey(t, &privateKey.PublicKey)
	token := signToken(t, privateKey, tokenClaims{
		Issuer:   testIssuer,
		Audience: []string{testAudience},
		Expires:  time.Now().Add(time.Hour).Unix(),
	})

	request := httptest.NewRequest(http.MethodGet, "/api/admin/collections", nil)
	request.Header.Set(AccessJWTHeader, token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body)
	}
}

func TestCloudflareAccessValidatesSignatureAndClaims(t *testing.T) {
	t.Parallel()

	trustedKey := newRSAKey(t)
	untrustedKey := newRSAKey(t)
	handler := newCloudflareTestHandlerWithKey(t, &trustedKey.PublicKey)
	now := time.Now()

	tests := []struct {
		name   string
		key    *rsa.PrivateKey
		claims tokenClaims
	}{
		{
			name: "signature",
			key:  untrustedKey,
			claims: tokenClaims{
				Issuer: testIssuer, Audience: []string{testAudience}, Expires: now.Add(time.Hour).Unix(),
			},
		},
		{
			name: "issuer",
			key:  trustedKey,
			claims: tokenClaims{
				Issuer:   "https://other.cloudflareaccess.com",
				Audience: []string{testAudience}, Expires: now.Add(time.Hour).Unix(),
			},
		},
		{
			name: "audience",
			key:  trustedKey,
			claims: tokenClaims{
				Issuer: testIssuer, Audience: []string{"other-audience"}, Expires: now.Add(time.Hour).Unix(),
			},
		},
		{
			name: "expiration",
			key:  trustedKey,
			claims: tokenClaims{
				Issuer: testIssuer, Audience: []string{testAudience}, Expires: now.Add(-time.Minute).Unix(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest(http.MethodGet, "/api/admin/collections", nil)
			request.Header.Set(AccessJWTHeader, signToken(t, test.key, test.claims))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestCloudflareAccessModeLeavesHealthPublic(t *testing.T) {
	t.Parallel()

	handler := newCloudflareTestHandler(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/health", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func newCloudflareTestHandler(t *testing.T) http.Handler {
	t.Helper()
	key := newRSAKey(t)
	return newCloudflareTestHandlerWithKey(t, &key.PublicKey)
}

func newCloudflareTestHandlerWithKey(t *testing.T, publicKey *rsa.PublicKey) http.Handler {
	t.Helper()

	verifier := newCloudflareAccessVerifier(
		testIssuer,
		testAudience,
		&oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{publicKey}},
		nil,
	)
	middleware, err := NewAdminMiddleware(ModeCloudflareAccess, verifier)
	if err != nil {
		t.Fatalf("NewAdminMiddleware() error = %v", err)
	}
	return middleware.Wrap(successHandler())
}

func successHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func newRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() error = %v", err)
	}
	return key
}

type tokenClaims struct {
	Issuer   string   `json:"iss"`
	Audience []string `json:"aud"`
	Expires  int64    `json:"exp"`
	IssuedAt int64    `json:"iat"`
	Subject  string   `json:"sub"`
}

func signToken(t *testing.T, key *rsa.PrivateKey, claims tokenClaims) string {
	t.Helper()
	if claims.IssuedAt == 0 {
		claims.IssuedAt = time.Now().Add(-time.Minute).Unix()
	}
	if claims.Subject == "" {
		claims.Subject = "admin@example.com"
	}

	header := encodeJSON(t, map[string]string{"alg": "RS256", "typ": "JWT"})
	payload := encodeJSON(t, claims)
	signedContent := header + "." + payload
	digest := sha256.Sum256([]byte(signedContent))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signedContent + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func encodeJSON(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(encoded)
}

func TestUnauthorizedResponseDoesNotLeakVerifierErrors(t *testing.T) {
	t.Parallel()

	middleware, err := NewAdminMiddleware(ModeCloudflareAccess, verifierFunc(
		func(context.Context, string) error { return fmt.Errorf("sensitive verifier detail") },
	))
	if err != nil {
		t.Fatalf("NewAdminMiddleware() error = %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/admin/collections", nil)
	request.Header.Set(AccessJWTHeader, "token")
	response := httptest.NewRecorder()
	middleware.Wrap(successHandler()).ServeHTTP(response, request)

	if strings.Contains(response.Body.String(), "sensitive") {
		t.Fatalf("response leaked verifier error: %s", response.Body)
	}
}

type verifierFunc func(context.Context, string) error

func (f verifierFunc) Verify(ctx context.Context, token string) error {
	return f(ctx, token)
}
