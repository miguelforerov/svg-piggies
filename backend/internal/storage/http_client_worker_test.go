//go:build js && wasm

package storage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"syscall/js"
	"testing"

	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/syumai/workers/cloudflare/fetch"
)

// This imitates the reader syumai/workers creates for a null JS Response.body.
// Any attempt to read it reproduces the failure seen in workerd.
type nullFetchBody struct {
	closed bool
}

func (*nullFetchBody) Read([]byte) (int, error) {
	panic("attempt to read a null JavaScript body")
}

func (b *nullFetchBody) Close() error {
	b.closed = true
	return nil
}

func (*nullFetchBody) GetRawJSBody() js.Value { return js.Null() }

func TestWorkerR2ClientUsesSupportedManualRedirectMode(t *testing.T) {
	var selected fetch.RedirectMode
	base := redirectRoundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("unexpected HTTP request")
		return nil, nil
	})
	client := newWorkerR2HTTPClient(func(mode fetch.RedirectMode) *http.Client {
		selected = mode
		return &http.Client{Transport: base}
	})
	if selected != fetch.RedirectModeManual || selected.String() != "manual" {
		t.Fatalf("Worker fetch redirect mode = %q, want manual", selected)
	}
	redirects, ok := client.Transport.(rejectRedirectTransport)
	if !ok {
		t.Fatalf("Worker transport = %T, want redirect rejecting wrapper", client.Transport)
	}
	if _, ok := redirects.base.(workerResponseTransport); !ok {
		t.Fatalf("Worker inner transport = %T, want response normalization", redirects.base)
	}
	if client.Timeout != r2HTTPTimeout || client.CheckRedirect == nil {
		t.Fatal("Worker timeout or redirect policy changed")
	}
}

func TestWorkerHEADNullBodyPreservesResponseMetadata(t *testing.T) {
	body := &nullFetchBody{}
	request, err := http.NewRequest(http.MethodHead, "https://example.test/object", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := &http.Response{
		Status: "200 OK", StatusCode: http.StatusOK,
		Header: http.Header{"Content-Length": {"110"}, "Etag": {"\"etag\""}, "X-Amz-Request-Id": {"request-id"}},
		Body:   body, ContentLength: 110, Request: request,
	}
	transport := workerResponseTransport{base: redirectRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return response, nil
	})}
	got, err := transport.RoundTrip(request)
	if err != nil || got != response || got.Body != http.NoBody || !body.closed {
		t.Fatalf("HEAD response = %v, error = %v, null body closed = %t", got, err, body.closed)
	}
	if got.Status != "200 OK" || got.StatusCode != 200 || got.ContentLength != 110 ||
		got.Header.Get("ETag") != "\"etag\"" || got.Header.Get("X-Amz-Request-Id") != "request-id" || got.Request != request {
		t.Fatal("HEAD response metadata changed")
	}
	if _, err := io.ReadAll(got.Body); err != nil {
		t.Fatalf("reading normalized HEAD body: %v", err)
	}
}

func TestWorkerGETBodyAndHTTPErrorRemainReadable(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNotFound, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			body := io.NopCloser(strings.NewReader("GET response body"))
			response := &http.Response{StatusCode: status, Status: http.StatusText(status), Body: body}
			client := newWorkerR2HTTPClient(func(fetch.RedirectMode) *http.Client {
				return &http.Client{Transport: redirectRoundTripFunc(func(*http.Request) (*http.Response, error) { return response, nil })}
			})
			request, err := http.NewRequest(http.MethodGet, "https://example.test/object", nil)
			if err != nil {
				t.Fatal(err)
			}
			got, err := client.Do(request)
			if err != nil || got.StatusCode != status {
				t.Fatalf("GET response = %v, error = %v", got, err)
			}
			data, err := io.ReadAll(got.Body)
			_ = got.Body.Close()
			if err != nil || string(data) != "GET response body" {
				t.Fatalf("GET body = %q, error = %v", data, err)
			}
		})
	}
}

func TestWorkerHEADRedirectsRemainRejected(t *testing.T) {
	for _, status := range []int{301, 302, 307, 308} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			client := newWorkerR2HTTPClient(func(mode fetch.RedirectMode) *http.Client {
				if mode != fetch.RedirectModeManual {
					t.Fatal("fetch redirect mode changed")
				}
				return &http.Client{Transport: redirectRoundTripFunc(func(request *http.Request) (*http.Response, error) {
					calls++
					if request.URL.Host != "example.test" {
						t.Fatal("signed request destination changed")
					}
					return &http.Response{
						StatusCode: status, Header: http.Header{"Location": {"https://other.example/object"}},
						Body: &nullFetchBody{}, Request: request,
					}, nil
				})}
			})
			request, err := http.NewRequest(http.MethodHead, "https://example.test/object", nil)
			if err != nil {
				t.Fatal(err)
			}
			response, err := client.Do(request)
			if response != nil || !errors.Is(err, errR2Redirect) || calls != 1 {
				t.Fatalf("response = %v, error = %v, calls = %d", response, err, calls)
			}
		})
	}
}

func TestWorkerSDKHEADProcessesNullBody(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNotFound} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			client := newWorkerR2HTTPClient(func(fetch.RedirectMode) *http.Client {
				return &http.Client{Transport: redirectRoundTripFunc(func(request *http.Request) (*http.Response, error) {
					calls++
					if request.Method != http.MethodHead || request.Header.Get("Authorization") == "" {
						t.Fatal("SDK HEAD request or SigV4 Authorization changed")
					}
					return &http.Response{
						StatusCode: status, Status: http.StatusText(status),
						Header: http.Header{"Content-Length": {"110"}, "Content-Type": {"image/svg+xml"}, "Etag": {"\"etag\""}},
						Body:   &nullFetchBody{}, ContentLength: 110, Request: request,
					}, nil
				})}
			})
			storageClient, err := newClient(validConfig(), client)
			if err != nil {
				t.Fatal(err)
			}
			metadata, err := storageClient.HeadObject(context.Background(), FilesBucket, "products/test/file.svg")
			if calls != 1 {
				t.Fatalf("SDK HEAD calls = %d, want 1", calls)
			}
			if status == http.StatusOK {
				if err != nil || metadata.ContentLength != 110 || metadata.ETag != "\"etag\"" {
					t.Fatalf("SDK HEAD metadata = %#v, error = %v", metadata, err)
				}
				return
			}
			var responseError *smithyhttp.ResponseError
			if !errors.As(err, &responseError) || responseError.HTTPStatusCode() != status {
				t.Fatalf("SDK HEAD error = %v, want HTTP %d", err, status)
			}
		})
	}
}

func TestWorkerSDKHEADWithActualFetchResponseConversion(t *testing.T) {
	// Use syumai/workers' real fetch response conversion with a local JS
	// binding. The binding returns a native Response whose body is null.
	var calls int
	var method, redirect string
	var bodyWasNull bool
	binding := js.Global().Get("Object").New()
	fetchFunc := js.FuncOf(func(_ js.Value, args []js.Value) any {
		calls++
		method = args[0].Get("method").String()
		redirect = args[1].Get("redirect").String()
		init := js.ValueOf(map[string]any{
			"status": 200,
			"headers": map[string]any{
				"Content-Length": "110",
				"Content-Type":   "image/svg+xml",
				"ETag":           "\"etag\"",
			},
		})
		response := js.Global().Get("Response").New(js.Null(), init)
		bodyWasNull = response.Get("body").IsNull()
		return js.Global().Get("Promise").Call("resolve", response)
	})
	defer fetchFunc.Release()
	binding.Set("fetch", fetchFunc)
	client := newWorkerR2HTTPClient(func(mode fetch.RedirectMode) *http.Client {
		return fetch.NewClient(fetch.WithBinding(binding)).HTTPClient(mode)
	})
	storageClient, err := newClient(validConfig(), client)
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := storageClient.HeadObject(context.Background(), FilesBucket, "products/test/file.svg")
	if err != nil || calls != 1 || method != http.MethodHead || redirect != "manual" || !bodyWasNull ||
		metadata.ContentLength != 110 || metadata.ETag != "\"etag\"" {
		t.Fatalf("SDK HEAD metadata = %#v, error = %v, calls = %d, method = %q, redirect = %q, null body = %t",
			metadata, err, calls, method, redirect, bodyWasNull)
	}
}

func syntheticWorkerResponse(status int, body js.Value, headers js.Value) (js.Value, js.Value, js.Func) {
	init := js.Global().Get("Object").New()
	init.Set("status", status)
	init.Set("headers", headers)
	response := js.Global().Get("Response").New(body, init)
	binding := js.Global().Get("Object").New()
	fetchFunc := js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return js.Global().Get("Promise").Call("resolve", response)
	})
	binding.Set("fetch", fetchFunc)
	return response, binding, fetchFunc
}

func TestWorkerHeaderConversionAndSDKHEADMetadata(t *testing.T) {
	const date = "Wed, 21 Oct 2026 07:28:00 GMT"
	const disposition = `attachment; filename="one,two.svg"`
	const metadataValue = "red,blue"
	headers := js.Global().Get("Headers").New()
	headers.Call("set", "Date", date)
	headers.Call("set", "Last-Modified", date)
	headers.Call("set", "Content-Length", "110")
	headers.Call("set", "Content-Type", "image/svg+xml")
	headers.Call("set", "ETag", `"etag"`)
	headers.Call("set", "Content-Disposition", disposition)
	headers.Call("set", "X-Amz-Meta-Note", metadataValue)
	headers.Call("set", "Cache-Control", "public, max-age=60")
	headers.Call("append", "Set-Cookie", "a=1; Path=/")
	headers.Call("append", "Set-Cookie", "b=2; Path=/")
	jsResponse, binding, fetchFunc := syntheticWorkerResponse(http.StatusOK, js.Null(), headers)
	defer fetchFunc.Release()
	if got := jsResponse.Get("headers").Call("get", "Date").String(); got != date {
		t.Fatalf("original JS Date = %q", got)
	}
	if got := jsResponse.Get("headers").Call("get", "Last-Modified").String(); got != date {
		t.Fatalf("original JS Last-Modified = %q", got)
	}
	base := fetch.NewClient(fetch.WithBinding(binding)).HTTPClient(fetch.RedirectModeManual)
	request, err := http.NewRequest(http.MethodHead, "https://example.test/object", nil)
	if err != nil {
		t.Fatal(err)
	}
	converted, err := base.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	if got := converted.Header.Values("Date"); len(got) != 2 || got[0] != "Wed" || got[1] != " 21 Oct 2026 07:28:00 GMT" {
		t.Fatalf("dependency Date conversion = %q", got)
	}
	if got := converted.Header.Values("Last-Modified"); len(got) != 2 || got[0] != "Wed" || got[1] != " 21 Oct 2026 07:28:00 GMT" {
		t.Fatalf("dependency Last-Modified conversion = %q", got)
	}
	if _, err := http.ParseTime(converted.Header.Get("Date")); err == nil {
		t.Fatal("split Date unexpectedly parsed as a complete HTTP date")
	}
	if _, err := http.ParseTime(converted.Header.Get("Last-Modified")); err == nil {
		t.Fatal("split Last-Modified unexpectedly parsed as a complete HTTP date")
	}
	_ = converted.Body.Close()
	client := newWorkerR2HTTPClient(func(mode fetch.RedirectMode) *http.Client {
		return fetch.NewClient(fetch.WithBinding(binding)).HTTPClient(mode)
	})
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || response.ContentLength != 110 {
		t.Fatalf("HEAD status = %d, content length = %d", response.StatusCode, response.ContentLength)
	}
	if data, err := io.ReadAll(response.Body); err != nil || len(data) != 0 {
		t.Fatalf("normalized HEAD body = %q, error = %v", data, err)
	}
	for name, want := range map[string]string{
		"Date": date, "Last-Modified": date,
		"Content-Disposition": disposition, "X-Amz-Meta-Note": metadataValue,
		"Cache-Control": "public, max-age=60",
	} {
		if got := response.Header.Get(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
		if values := response.Header.Values(name); len(values) != 1 || values[0] != want {
			t.Errorf("%s values = %q, want one complete value", name, values)
		}
	}
	if _, ok := response.Header["Last-Modified"]; !ok {
		t.Fatal("header name was not canonicalized")
	}
	if got := response.Header.Values("Set-Cookie"); len(got) != 2 || got[0] != "a=1; Path=/" || got[1] != "b=2; Path=/" {
		t.Fatalf("separate Set-Cookie values = %q", got)
	}
	storageClient, err := newClient(validConfig(), client)
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := storageClient.HeadObject(context.Background(), FilesBucket, "products/test/file.svg")
	wantModified, parseErr := http.ParseTime(date)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if err != nil || metadata.ContentLength != 110 || metadata.ContentType != "image/svg+xml" ||
		metadata.ETag != `"etag"` || !metadata.LastModified.Equal(wantModified) {
		t.Fatalf("SDK HEAD metadata = %#v, error = %v", metadata, err)
	}
}

func TestWorkerGETWithActualFetchResponsePreservesBody(t *testing.T) {
	const date = "Wed, 21 Oct 2026 07:28:00 GMT"
	const content = "<svg xmlns=\"http://www.w3.org/2000/svg\"></svg>"
	headers := js.Global().Get("Headers").New()
	headers.Call("set", "Date", date)
	headers.Call("set", "X-Amz-Meta-Note", "red,blue")
	_, binding, fetchFunc := syntheticWorkerResponse(http.StatusOK, js.ValueOf(content), headers)
	defer fetchFunc.Release()
	client := newWorkerR2HTTPClient(func(mode fetch.RedirectMode) *http.Client {
		return fetch.NewClient(fetch.WithBinding(binding)).HTTPClient(mode)
	})
	request, err := http.NewRequest(http.MethodGet, "https://example.test/object", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil || string(data) != content || response.Header.Get("Date") != date || response.Header.Get("X-Amz-Meta-Note") != "red,blue" {
		t.Fatalf("GET body = %q, Date = %q, metadata = %q, error = %v", data,
			response.Header.Get("Date"), response.Header.Get("X-Amz-Meta-Note"), err)
	}
}
