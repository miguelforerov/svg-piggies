package config

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
)

const (
	HyperdriveBinding   = "HYPERDRIVE"
	ProductFilesBinding = "PRODUCT_FILES"

	ProductionDatabaseName  = "svg_piggies"
	DevelopmentDatabaseName = "svg_piggies_development"
	TestDatabaseName        = "svg_piggies_test"
)

type LookupFunc func(string) string

type Config struct {
	Environment                string
	AuthMode                   auth.Mode
	CloudflareAccessTeamDomain string
	CloudflareAccessAudience   string
	DatabaseURL                string
	DatabaseName               string
	CORSAllowedOrigin          string
	DownloadURLTTL             time.Duration
	StripeSecretKey            string
	StripeWebhookSecret        string
}

func Load(lookup LookupFunc) (Config, error) {
	if lookup == nil {
		return Config{}, errors.New("config lookup function is required")
	}

	ttlSeconds, err := positiveInt(lookup, "DOWNLOAD_URL_TTL_SECONDS", 900)
	if err != nil {
		return Config{}, err
	}

	environment := valueOrDefault(lookup("APP_ENV"), "development")
	authModeValue := lookup("AUTH_MODE")
	if authModeValue == "" {
		if isLocalEnvironment(environment) {
			authModeValue = string(auth.ModeDevelopment)
		} else {
			return Config{}, fmt.Errorf("AUTH_MODE is required when APP_ENV=%s", environment)
		}
	}
	authMode, err := auth.ParseMode(authModeValue)
	if err != nil {
		return Config{}, err
	}

	config := Config{
		Environment: environment,
		AuthMode:    authMode,
		CloudflareAccessTeamDomain: strings.TrimRight(strings.TrimSpace(
			lookup("CLOUDFLARE_ACCESS_TEAM_DOMAIN"),
		), "/"),
		CloudflareAccessAudience: strings.TrimSpace(lookup("CLOUDFLARE_ACCESS_AUD")),
		DatabaseURL:              lookup("DATABASE_URL"),
		DatabaseName:             valueOrDefault(lookup("DATABASE_NAME"), DatabaseNameForEnvironment(environment)),
		CORSAllowedOrigin:        valueOrDefault(lookup("CORS_ALLOWED_ORIGIN"), "http://localhost:4321"),
		DownloadURLTTL:           time.Duration(ttlSeconds) * time.Second,
		StripeSecretKey:          lookup("STRIPE_SECRET_KEY"),
		StripeWebhookSecret:      lookup("STRIPE_WEBHOOK_SECRET"),
	}
	if err := config.ValidateAuth(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (c Config) ValidateAuth() error {
	switch c.AuthMode {
	case auth.ModeDevelopment:
		if !isLocalEnvironment(c.Environment) {
			return fmt.Errorf(
				"AUTH_MODE=development is only allowed when APP_ENV is development or test; got APP_ENV=%s",
				c.Environment,
			)
		}
	case auth.ModeCloudflareAccess:
		if c.CloudflareAccessTeamDomain == "" {
			return errors.New("CLOUDFLARE_ACCESS_TEAM_DOMAIN is required when AUTH_MODE=cloudflare_access")
		}
		teamURL, err := url.Parse(c.CloudflareAccessTeamDomain)
		if err != nil || teamURL.Scheme != "https" || teamURL.Host == "" ||
			teamURL.Path != "" || teamURL.RawQuery != "" || teamURL.Fragment != "" ||
			!strings.HasSuffix(strings.ToLower(teamURL.Hostname()), ".cloudflareaccess.com") {
			return errors.New(
				"CLOUDFLARE_ACCESS_TEAM_DOMAIN must be an HTTPS Cloudflare Access team URL such as https://myteam.cloudflareaccess.com",
			)
		}
		if c.CloudflareAccessAudience == "" {
			return errors.New("CLOUDFLARE_ACCESS_AUD is required when AUTH_MODE=cloudflare_access")
		}
	default:
		return errors.New("unsupported AUTH_MODE")
	}
	return nil
}

func isLocalEnvironment(environment string) bool {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "development", "test":
		return true
	default:
		return false
	}
}

func DatabaseNameForEnvironment(environment string) string {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "production":
		return ProductionDatabaseName
	case "test":
		return TestDatabaseName
	default:
		return DevelopmentDatabaseName
	}
}

// ValidateDatabaseURL validates the direct PostgreSQL URL used by the native
// Go server. Cloudflare Workers receive their URL from the Hyperdrive binding.
func (c Config) ValidateDatabaseURL() error {
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return errors.New("DATABASE_URL is required")
	}

	parsed, err := url.Parse(c.DatabaseURL)
	if err != nil {
		return fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return errors.New("DATABASE_URL must use the postgres or postgresql scheme")
	}

	databaseName := strings.TrimPrefix(parsed.EscapedPath(), "/")
	databaseName, err = url.PathUnescape(databaseName)
	if err != nil {
		return fmt.Errorf("parse database name from DATABASE_URL: %w", err)
	}
	if databaseName == "" {
		return errors.New("DATABASE_URL must include a database name")
	}
	if databaseName != c.DatabaseName {
		return fmt.Errorf(
			"DATABASE_URL selects database %q, but %s expects %q",
			databaseName,
			c.Environment,
			c.DatabaseName,
		)
	}

	return nil
}

func (c Config) ValidatePayments() error {
	if c.StripeSecretKey == "" {
		return errors.New("STRIPE_SECRET_KEY is required")
	}
	if c.StripeWebhookSecret == "" {
		return errors.New("STRIPE_WEBHOOK_SECRET is required")
	}
	return nil
}

func positiveInt(lookup LookupFunc, name string, fallback int) (int, error) {
	raw := lookup(name)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return value, nil
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
