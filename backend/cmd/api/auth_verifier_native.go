//go:build !js

package main

import (
	"net/http"

	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
)

func newAuthVerifier(cfg config.Config) (auth.TokenVerifier, error) {
	if cfg.AuthMode == auth.ModeDevelopment {
		return nil, nil
	}
	return auth.NewCloudflareAccessVerifier(
		cfg.CloudflareAccessTeamDomain,
		cfg.CloudflareAccessAudience,
		http.DefaultClient,
	)
}
