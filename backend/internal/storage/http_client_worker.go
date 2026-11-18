//go:build js && wasm

package storage

import (
	"net/http"
	"strings"
	"syscall/js"

	"github.com/syumai/workers/cloudflare/fetch"
)

func newR2HTTPClient() *http.Client {
	return newWorkerR2HTTPClient(func(mode fetch.RedirectMode) *http.Client {
		return fetch.NewClient().HTTPClient(mode)
	})
}

func newWorkerR2HTTPClient(newFetchClient func(fetch.RedirectMode) *http.Client) *http.Client {
	client := newFetchClient(fetch.RedirectModeManual)
	client.Transport = rejectRedirectTransport{base: workerResponseTransport{base: client.Transport}}
	client.Timeout = r2HTTPTimeout
	client.CheckRedirect = rejectR2Redirect
	return client
}

// syumai/workers v0.33.0 splits every fetch header value at commas and wraps
// a null Response.body in a reader. Reassemble the converted header values and
// replace only a null HEAD body before the response reaches the S3 SDK.
type workerResponseTransport struct {
	base http.RoundTripper
}

func (t workerResponseTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := t.base.RoundTrip(request)
	if err != nil || response == nil {
		return response, err
	}
	body, fromFetch := response.Body.(interface{ GetRawJSBody() js.Value })
	if fromFetch {
		// JS Headers.entries() normally combines repeated list fields into one
		// comma-separated value. The dependency splits that value without
		// regard for commas inside dates, quoted text, or metadata. Joining
		// the fragments restores the original value and its whitespace.
		// Set-Cookie is the exception: separate cookies must stay separate.
		for name, values := range response.Header {
			if len(values) > 1 && !strings.EqualFold(name, "Set-Cookie") {
				response.Header[name] = []string{strings.Join(values, ",")}
			}
		}
	}
	if request.Method != http.MethodHead {
		return response, nil
	}
	if response.Body == nil {
		response.Body = http.NoBody
		return response, nil
	}
	if fromFetch {
		raw := body.GetRawJSBody()
		if raw.IsNull() || raw.IsUndefined() {
			_ = response.Body.Close()
			response.Body = http.NoBody
		}
	}
	return response, nil
}
