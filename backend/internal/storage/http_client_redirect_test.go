package storage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type redirectRoundTripFunc func(*http.Request) (*http.Response, error)

func (f redirectRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type closeTrackingBody struct {
	io.Reader
	closed bool
}

func (b *closeTrackingBody) Close() error {
	b.closed = true
	return nil
}

func TestR2RedirectTransportRejectsEveryRedirect(t *testing.T) {
	for _, status := range []int{301, 302, 307, 308} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			body := &closeTrackingBody{Reader: strings.NewReader("redirect body")}
			requests := 0
			client := &http.Client{Transport: rejectRedirectTransport{base: redirectRoundTripFunc(func(request *http.Request) (*http.Response, error) {
				requests++
				if request.URL.Host != "example.r2.cloudflarestorage.com" || request.Header.Get("Authorization") != "signed-request" {
					t.Errorf("signed request or destination changed")
				}
				return &http.Response{
					StatusCode: status,
					Header:     http.Header{"Location": {"https://other.example/object"}},
					Body:       body,
					Request:    request,
				}, nil
			})}}
			request, err := http.NewRequest(http.MethodHead, "https://example.r2.cloudflarestorage.com/bucket/object", nil)
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Authorization", "signed-request")
			response, err := client.Do(request)
			if response != nil || !errors.Is(err, errR2Redirect) || requests != 1 || !body.closed {
				t.Fatalf("response = %v, error = %v, requests = %d, body closed = %t", response, err, requests, body.closed)
			}
		})
	}
}

func TestR2RedirectTransportPreservesNormalResponses(t *testing.T) {
	for _, status := range []int{200, 204, 404, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			body := &closeTrackingBody{Reader: strings.NewReader("response body")}
			response := &http.Response{StatusCode: status, Body: body}
			transport := rejectRedirectTransport{base: redirectRoundTripFunc(func(*http.Request) (*http.Response, error) {
				return response, nil
			})}
			request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test/object", nil)
			if err != nil {
				t.Fatal(err)
			}
			got, err := transport.RoundTrip(request)
			if err != nil || got != response || body.closed {
				t.Fatalf("response changed: got = %v, error = %v, body closed = %t", got, err, body.closed)
			}
		})
	}
}

func TestS3HeadRejectsRedirectAsError(t *testing.T) {
	requests := 0
	httpClient := &http.Client{Transport: rejectRedirectTransport{base: redirectRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.URL.Host != "example.r2.cloudflarestorage.com" {
			t.Errorf("SDK request changed destination to %q", request.URL.Host)
		}
		return &http.Response{
			StatusCode: http.StatusMovedPermanently,
			Header:     http.Header{"Location": {"https://other.example/object"}},
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    request,
		}, nil
	})}}
	client, err := newClient(validConfig(), httpClient)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.HeadObject(context.Background(), FilesBucket, "products/test/file.svg"); !errors.Is(err, errR2Redirect) || requests == 0 {
		t.Fatalf("SDK HEAD error = %v, requests = %d; want redirect error", err, requests)
	}
}
