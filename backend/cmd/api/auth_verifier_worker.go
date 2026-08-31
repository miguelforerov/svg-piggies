//go:build js && wasm

package main

import (
	"github.com/syumai/workers/cloudflare/fetch"
	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
)

func newAuthVerifier(cfg config.Config) (auth.TokenVerifier, error) {
	if cfg.AuthMode == auth.ModeDevelopment {
		return nil, nil
	}
	client := fetch.NewClient().HTTPClient(fetch.RedirectModeError)
	return auth.NewCloudflareAccessVerifier(
		cfg.CloudflareAccessTeamDomain,
		cfg.CloudflareAccessAudience,
		client,
	)
}
