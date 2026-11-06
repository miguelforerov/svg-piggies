package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
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
	Verify(ctx context.Context, rawToken string) (string, error)
}

type AdminMiddleware struct {
	mode     Mode
	verifier TokenVerifier
	admins   map[string]struct{}
}

func NormalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || email == "" || strings.ContainsAny(email, "\r\n") {
		return "", errors.New("invalid administrative email")
	}
	return email, nil
}

func NewAdminMiddleware(mode Mode, verifier TokenVerifier, adminEmails []string) (*AdminMiddleware, error) {
	switch mode {
	case ModeDevelopment:
		return &AdminMiddleware{mode: mode}, nil
	case ModeCloudflareAccess:
		if verifier == nil {
			return nil, errors.New("Cloudflare Access token verifier is required")
		}
		if len(adminEmails) == 0 {
			return nil, errors.New("ADMIN_EMAILS must contain at least one administrator")
		}
		admins := make(map[string]struct{}, len(adminEmails))
		for _, value := range adminEmails {
			email, err := NormalizeEmail(value)
			if err != nil {
				return nil, err
			}
			admins[email] = struct{}{}
		}
		return &AdminMiddleware{mode: mode, verifier: verifier, admins: admins}, nil
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
		email, err := m.verifier.Verify(r.Context(), rawToken)
		if err != nil {
			writeUnauthorized(w)
			return
		}
		email, err = NormalizeEmail(email)
		if err != nil {
			writeUnauthorized(w)
			return
		}
		if _, ok := m.admins[email]; !ok {
			writeForbidden(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code": "forbidden", "message": "admin permission required",
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
