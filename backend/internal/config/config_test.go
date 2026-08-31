package config

import (
	"strings"
	"testing"
	"time"

	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
)

func TestLoadDefaults(t *testing.T) {
	t.Parallel()

	got, err := Load(func(string) string { return "" })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got.Environment != "development" {
		t.Errorf("Environment = %q, want development", got.Environment)
	}
	if got.AuthMode != auth.ModeDevelopment {
		t.Errorf("AuthMode = %q, want %q", got.AuthMode, auth.ModeDevelopment)
	}
	if got.DatabaseName != DevelopmentDatabaseName {
		t.Errorf("DatabaseName = %q, want %q", got.DatabaseName, DevelopmentDatabaseName)
	}
	if got.CORSAllowedOrigin != "http://localhost:4321" {
		t.Errorf("CORSAllowedOrigin = %q", got.CORSAllowedOrigin)
	}
	if got.DownloadURLTTL != 15*time.Minute {
		t.Errorf("DownloadURLTTL = %s, want 15m", got.DownloadURLTTL)
	}
}

func TestLoadRejectsDevelopmentAuthInProduction(t *testing.T) {
	t.Parallel()

	_, err := Load(func(name string) string {
		switch name {
		case "APP_ENV":
			return "production"
		case "AUTH_MODE":
			return "development"
		default:
			return ""
		}
	})
	if err == nil {
		t.Fatal("Load() error = nil, want unsafe authentication configuration error")
	}
	if !strings.Contains(err.Error(), "AUTH_MODE=development") {
		t.Fatalf("Load() error = %q, want clear AUTH_MODE error", err)
	}
}

func TestLoadCloudflareAccessAuth(t *testing.T) {
	t.Parallel()

	got, err := Load(func(name string) string {
		switch name {
		case "APP_ENV":
			return "production"
		case "AUTH_MODE":
			return "cloudflare_access"
		case "CLOUDFLARE_ACCESS_TEAM_DOMAIN":
			return "https://svg-piggies.cloudflareaccess.com/"
		case "CLOUDFLARE_ACCESS_AUD":
			return "access-audience"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.CloudflareAccessTeamDomain != "https://svg-piggies.cloudflareaccess.com" {
		t.Errorf("CloudflareAccessTeamDomain = %q", got.CloudflareAccessTeamDomain)
	}
}

func TestDatabaseNameForEnvironment(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"development": DevelopmentDatabaseName,
		"test":        TestDatabaseName,
		"production":  ProductionDatabaseName,
	}
	for environment, want := range tests {
		if got := DatabaseNameForEnvironment(environment); got != want {
			t.Errorf("DatabaseNameForEnvironment(%q) = %q, want %q", environment, got, want)
		}
	}
}

func TestValidateDatabaseURL(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Environment:  "test",
		DatabaseName: TestDatabaseName,
		DatabaseURL:  "postgresql://postgres:postgres@localhost:5432/svg_piggies_test?sslmode=disable",
	}
	if err := cfg.ValidateDatabaseURL(); err != nil {
		t.Fatalf("ValidateDatabaseURL() error = %v", err)
	}

	cfg.DatabaseURL = "postgresql://postgres:postgres@localhost:5432/svg_piggies_development"
	if err := cfg.ValidateDatabaseURL(); err == nil {
		t.Fatal("ValidateDatabaseURL() error = nil, want database name mismatch")
	}
}

func TestLoadRejectsInvalidDownloadTTL(t *testing.T) {
	t.Parallel()

	_, err := Load(func(name string) string {
		if name == "DOWNLOAD_URL_TTL_SECONDS" {
			return "0"
		}
		return ""
	})
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
}

func TestValidatePayments(t *testing.T) {
	t.Parallel()

	config := Config{
		StripeSecretKey:     "sk_test_value",
		StripeWebhookSecret: "whsec_value",
	}
	if err := config.ValidatePayments(); err != nil {
		t.Fatalf("ValidatePayments() error = %v", err)
	}
}
