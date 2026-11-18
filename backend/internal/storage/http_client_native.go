//go:build !js || !wasm

package storage

import "net/http"

func newR2HTTPClient() *http.Client {
	return &http.Client{
		Timeout:       r2HTTPTimeout,
		CheckRedirect: rejectR2Redirect,
	}
}
