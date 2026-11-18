package storage

import (
	"errors"
	"fmt"
	"net/http"
)

var errR2Redirect = errors.New("R2 redirect response rejected")

// rejectRedirectTransport prevents a signed S3 request from being sent to a
// different destination, even when the underlying fetch returns a 3xx response.
type rejectRedirectTransport struct {
	base http.RoundTripper
}

func (t rejectRedirectTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := t.base.RoundTrip(request)
	if err != nil || response == nil {
		return response, err
	}
	if response.StatusCode >= 300 && response.StatusCode < 400 {
		if response.Body != nil {
			_ = response.Body.Close()
		}
		return nil, fmt.Errorf("%w: HTTP %d", errR2Redirect, response.StatusCode)
	}
	return response, nil
}
