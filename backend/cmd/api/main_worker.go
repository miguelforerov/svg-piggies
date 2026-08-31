//go:build js && wasm

package main

import (
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
	"github.com/zdenaforero/svg-piggies/backend/internal/database"
)

func main() {
	cfg, err := config.Load(cloudflare.Getenv)
	if err != nil {
		panic(err)
	}
	authVerifier, err := newAuthVerifier(cfg)
	if err != nil {
		panic(err)
	}

	databaseProvider := database.NewHyperdriveProvider(config.HyperdriveBinding, cfg.DatabaseName)
	dependencies, err := buildDependencies(databaseProvider)
	if err != nil {
		panic(err)
	}

	handler, err := newHandler(cfg, dependencies, authVerifier)
	if err != nil {
		panic(err)
	}

	workers.Serve(handler)
}
