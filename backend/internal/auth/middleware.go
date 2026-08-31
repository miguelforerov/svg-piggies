package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

const (
	AccessJWTHeader = "Cf-Access-Jwt-Assertion"
	adminPath       = "/api/admin"
)

type Mode string

const (
	ModeDevelopment      Mode = "development"
	ModeCloudflareAccess Mode = "cloudflare_access"
)

func ParseMode(value string) (Mode, error) {
	mode := Mode(strings.ToLower(strings.TrimSpace(value)))
	switch mode {
	case ModeDevelopment, ModeCloudflareAccess:
		return mode, nil
	default:
		return "", errors.New("AUTH_MODE must be development or cloudflare_access")
	}
}

type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) error
}

type AdminMiddleware struct {
	mode     Mode
	verifier TokenVerifier
}

func NewAdminMiddleware(mode Mode, verifier TokenVerifier) (*AdminMiddleware, error) {
	switch mode {
	case ModeDevelopment:
		return &AdminMiddleware{mode: mode}, nil
	case ModeCloudflareAccess:
		if verifier == nil {
			return nil, errors.New("Cloudflare Access token verifier is required")
		}
		return &AdminMiddleware{mode: mode, verifier: verifier}, nil
	default:
		return nil, errors.New("unsupported admin authentication mode")
	}
}

func (m *AdminMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAdminPath(r.URL.Path) || m.mode == ModeDevelopment {
			next.ServeHTTP(w, r)
			return
		}

		rawToken := strings.TrimSpace(r.Header.Get(AccessJWTHeader))
		if rawToken == "" {
			writeUnauthorized(w)
			return
		}
		if err := m.verifier.Verify(r.Context(), rawToken); err != nil {
			writeUnauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAdminPath(path string) bool {
	return path == adminPath || strings.HasPrefix(path, adminPath+"/")
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    "unauthorized",
		"message": "admin authentication required",
	})
}
