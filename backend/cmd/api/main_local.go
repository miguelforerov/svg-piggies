//go:build !js

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zdenaforero/svg-piggies/backend/internal/config"
	"github.com/zdenaforero/svg-piggies/backend/internal/database"
)

func main() {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	if err := cfg.ValidateDatabaseURL(); err != nil {
		log.Fatal(err)
	}
	authVerifier, err := newAuthVerifier(cfg)
	if err != nil {
		log.Fatal(err)
	}

	connectionContext, cancelConnection := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelConnection()

	databaseProvider, err := database.NewPoolProvider(connectionContext, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer databaseProvider.Close()

	dependencies, err := buildDependencies(databaseProvider)
	if err != nil {
		log.Fatal(err)
	}
	handler, err := newHandler(cfg, dependencies, authVerifier)
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8787"
	}

	log.Printf("API listening on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
