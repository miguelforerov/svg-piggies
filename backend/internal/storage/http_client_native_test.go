//go:build !js || !wasm

package storage

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNativeR2HTTPClientDoesNotFollowRedirect(t *testing.T) {
	client := newR2HTTPClient()
	if client.Transport != nil {
		t.Fatal("native R2 client must use the standard Go HTTP transport")
	}
	requests := 0
	client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		if req.URL.Host != "example.r2.cloudflarestorage.com" || req.Header.Get("Authorization") != "signed-request" {
			t.Errorf("signed request changed: %s %s", req.Method, req.URL)
		}
		return &http.Response{
			StatusCode: http.StatusTemporaryRedirect,
			Header:     http.Header{"Location": {"https://other.example/object"}},
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    req,
		}, nil
	})

	req, err := http.NewRequest(http.MethodHead, "https://example.r2.cloudflarestorage.com/bucket/object", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "signed-request")
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusTemporaryRedirect || requests != 1 {
		t.Fatalf("redirect response = %d, requests = %d", response.StatusCode, requests)
	}
}
