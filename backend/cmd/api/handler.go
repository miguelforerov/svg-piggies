package main

import (
	"encoding/json"
	"log"
	"net/http"

	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
	"github.com/zdenaforero/svg-piggies/backend/internal/handlers"
)

func newHandler(
	cfg config.Config,
	dependencies dependencies,
	verifier auth.TokenVerifier,
) (http.Handler, error) {
	if err := cfg.ValidateAuth(); err != nil {
		return nil, err
	}
	adminAuth, err := auth.NewAdminMiddleware(cfg.AuthMode, verifier)
	if err != nil {
		return nil, err
	}

	server := handlers.NewServer(
		cfg.Environment,
		dependencies.collections,
		dependencies.productCollections,
		dependencies.productProductTypes,
		dependencies.productRelationships,
		dependencies.products,
		dependencies.productTypes,
	)
	strictHandler := generated.NewStrictHandlerWithOptions(
		server,
		nil,
		generated.StrictHTTPServerOptions{
			RequestErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
				writeAPIError(w, http.StatusBadRequest, generated.ErrorCodeInvalidRequest, err.Error())
			},
			ResponseErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
				log.Printf("API response error: %v", err)
				writeAPIError(
					w,
					http.StatusInternalServerError,
					generated.ErrorCodeInternalError,
					"an internal error occurred",
				)
			},
		},
	)

	router := generated.HandlerWithOptions(strictHandler, generated.StdHTTPServerOptions{
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeAPIError(w, http.StatusBadRequest, generated.ErrorCodeInvalidRequest, err.Error())
		},
	})

	return withCORS(adminAuth.Wrap(router), cfg.CORSAllowedOrigin), nil
}

func withCORS(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Stripe-Signature")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Vary", "Origin")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeAPIError(w http.ResponseWriter, status int, code generated.ErrorCode, message string) {
	writeJSON(w, status, generated.Error{Code: code, Message: message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
