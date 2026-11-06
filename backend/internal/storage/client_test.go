package storage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
)

func validConfig() config.Config {
	return config.Config{
		R2AccessKeyID: "example-access-key", R2SecretAccessKey: "example-secret-key",
		R2Endpoint: "https://example.r2.cloudflarestorage.com", R2Region: "auto",
		ProductFilesBucket: "private-files", ProductAssetsBucket: "public-assets",
	}
}

func TestValidateConfig(t *testing.T) {
	if err := validateConfig(validConfig()); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, want string
		change     func(*config.Config)
	}{
		{"access key", "R2_ACCESS_KEY_ID", func(c *config.Config) { c.R2AccessKeyID = " " }},
		{"secret key", "R2_SECRET_ACCESS_KEY", func(c *config.Config) { c.R2SecretAccessKey = "" }},
		{"endpoint", "R2_ENDPOINT", func(c *config.Config) { c.R2Endpoint = "" }},
		{"region", "R2_REGION", func(c *config.Config) { c.R2Region = "" }},
		{"wrong region", "R2_REGION", func(c *config.Config) { c.R2Region = "us-west-2" }},
		{"files bucket", "PRODUCT_FILES_BUCKET", func(c *config.Config) { c.ProductFilesBucket = "" }},
		{"assets bucket", "PRODUCT_ASSETS_BUCKET", func(c *config.Config) { c.ProductAssetsBucket = "" }},
		{"same buckets", "must be different", func(c *config.Config) { c.ProductAssetsBucket = c.ProductFilesBucket }},
		{"insecure endpoint", "R2_ENDPOINT", func(c *config.Config) { c.R2Endpoint = "http://example.com" }},
		{"endpoint path", "R2_ENDPOINT", func(c *config.Config) { c.R2Endpoint += "/bucket" }},
		{"endpoint credentials", "R2_ENDPOINT", func(c *config.Config) { c.R2Endpoint = "https://user:pass@example.com" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.change(&cfg)
			if err := validateConfig(cfg); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("validateConfig() = %v, want %s error", err, tt.want)
			}
		})
	}
}

func TestNewClientReusesClients(t *testing.T) {
	c, err := NewClient(validConfig())
	if err != nil {
		t.Fatal(err)
	}
	if c.s3Client == nil || c.presignClient == nil {
		t.Fatal("expected S3 and presign clients")
	}
	if _, err := newClient(validConfig(), nil); err == nil {
		t.Fatal("missing HTTP client accepted")
	}
}

func TestSDKPresignsR2PathStyleRequests(t *testing.T) {
	c, err := NewClient(validConfig())
	if err != nil {
		t.Fatal(err)
	}
	put, err := c.PresignPut(context.Background(), AssetsBucket, "products/123/image.webp", "image/webp", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(put.URL)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Host != "example.r2.cloudflarestorage.com" || parsed.EscapedPath() != "/public-assets/products/123/image.webp" || parsed.Query().Get("X-Amz-Expires") != "60" {
		t.Fatalf("unexpected presigned PUT URL: %s", put.URL)
	}
	if signed := parsed.Query().Get("X-Amz-SignedHeaders"); !strings.Contains(signed, "content-type") || !strings.Contains(signed, "if-none-match") {
		t.Fatalf("required PUT headers were not signed: %s", signed)
	}
	if got := put.Headers.Get("Content-Type"); got != "image/webp" {
		t.Fatalf("signed Content-Type = %q, want image/webp", got)
	}
	if got := put.Headers.Get("If-None-Match"); got != "*" {
		t.Fatalf("signed If-None-Match = %q, want *", got)
	}
	if strings.Contains(put.URL, validConfig().R2SecretAccessKey) {
		t.Fatal("presigned URL exposes the secret access key")
	}
	get, err := c.PresignGet(context.Background(), FilesBucket, "orders/123/download.zip", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err = url.Parse(get.URL)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.EscapedPath() != "/private-files/orders/123/download.zip" || parsed.Query().Get("X-Amz-Expires") != "60" {
		t.Fatalf("unexpected presigned GET URL: %s", get.URL)
	}
}

func TestValidateDurationAndKey(t *testing.T) {
	for _, d := range []time.Duration{0, time.Millisecond, 1500 * time.Millisecond, maxPresignDuration + time.Second} {
		if err := validateDuration(d); err == nil {
			t.Errorf("validateDuration(%s) accepted invalid duration", d)
		}
	}
	for _, d := range []time.Duration{time.Second, 15 * time.Minute, maxPresignDuration} {
		if err := validateDuration(d); err != nil {
			t.Errorf("validateDuration(%s): %v", d, err)
		}
	}
	for _, key := range []string{"", "/root", "a//b", "a/../b", "a\\b", "a\nb"} {
		if err := validateKey(key); err == nil {
			t.Errorf("validateKey(%q) accepted invalid key", key)
		}
	}
	if err := validateKey("products/123/image.svg"); err != nil {
		t.Fatal(err)
	}
}

func TestPresignPutMediaTypes(t *testing.T) {
	c := &Client{presignClient: &fakePresigner{}, filesBucket: "private-files", assetsBucket: "public-assets"}
	for _, mediaType := range []string{"image/png", "image/jpeg", "image/webp"} {
		if _, err := c.PresignPut(context.Background(), AssetsBucket, "products/123/image", mediaType, time.Minute); err != nil {
			t.Errorf("asset type %q rejected: %v", mediaType, err)
		}
	}
	for _, mediaType := range []string{"image/avif", "image/svg+xml", "text/html", "application/zip", "invalid\nvalue"} {
		if _, err := c.PresignPut(context.Background(), AssetsBucket, "products/123/image", mediaType, time.Minute); err == nil {
			t.Errorf("asset type %q accepted", mediaType)
		}
	}
	if _, err := c.PresignPut(context.Background(), FilesBucket, "products/123/download.zip", "application/zip", time.Minute); err != nil {
		t.Fatalf("private download type rejected: %v", err)
	}
}

type fakePresigner struct {
	putInput *s3.PutObjectInput
	getInput *s3.GetObjectInput
	expires  time.Duration
	err      error
}

func (f *fakePresigner) PresignPutObject(_ context.Context, input *s3.PutObjectInput, options ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	f.putInput = input
	f.expires = expiration(options)
	if f.err != nil {
		return nil, f.err
	}
	return &v4.PresignedHTTPRequest{URL: "https://example.test/upload", SignedHeader: http.Header{"Content-Type": {"image/webp"}}}, nil
}

func (f *fakePresigner) PresignGetObject(_ context.Context, input *s3.GetObjectInput, options ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	f.getInput = input
	f.expires = expiration(options)
	if f.err != nil {
		return nil, f.err
	}
	return &v4.PresignedHTTPRequest{URL: "https://example.test/download"}, nil
}

func expiration(options []func(*s3.PresignOptions)) time.Duration {
	var result s3.PresignOptions
	for _, option := range options {
		option(&result)
	}
	return result.Expires
}

type fakeObjectClient struct {
	headInput     *s3.HeadObjectInput
	deleteInput   *s3.DeleteObjectInput
	headContext   context.Context
	deleteContext context.Context
	err           error
}

func (f *fakeObjectClient) HeadObject(ctx context.Context, input *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	f.headInput = input
	f.headContext = ctx
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if f.err != nil {
		return nil, f.err
	}
	return &s3.HeadObjectOutput{ContentLength: aws.Int64(42), ContentType: aws.String("image/webp"), ETag: aws.String("etag")}, nil
}

func (f *fakeObjectClient) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	f.deleteInput = input
	f.deleteContext = ctx
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if f.err != nil {
		return nil, f.err
	}
	return &s3.DeleteObjectOutput{}, nil
}

func TestOperationsSelectBucketsAndBuildInputs(t *testing.T) {
	p := &fakePresigner{}
	o := &fakeObjectClient{}
	c := &Client{presignClient: p, s3Client: o, filesBucket: "private-files", assetsBucket: "public-assets"}
	ctx := context.Background()
	put, err := c.PresignPut(ctx, AssetsBucket, "products/123/image.webp", "image/webp", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if aws.ToString(p.putInput.Bucket) != "public-assets" || aws.ToString(p.putInput.Key) != "products/123/image.webp" || aws.ToString(p.putInput.ContentType) != "image/webp" || p.expires != time.Minute {
		t.Fatalf("unexpected PUT input: %+v, expiry %s", p.putInput, p.expires)
	}
	if put.URL == "" || put.Headers.Get("Content-Type") != "image/webp" {
		t.Fatalf("unexpected PUT request: %+v", put)
	}
	put.Headers.Set("Content-Type", "other")

	get, err := c.PresignGet(ctx, FilesBucket, "orders/123/download.zip", 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if get.URL == "" || aws.ToString(p.getInput.Bucket) != "private-files" || p.expires != 2*time.Minute {
		t.Fatalf("unexpected GET: %+v", p.getInput)
	}

	metadata, err := c.HeadObject(ctx, AssetsBucket, "products/123/image.webp")
	if err != nil {
		t.Fatal(err)
	}
	if aws.ToString(o.headInput.Bucket) != "public-assets" || metadata.ContentLength != 42 || metadata.ContentType != "image/webp" || metadata.ETag != "etag" {
		t.Fatalf("unexpected HEAD: %+v, %+v", o.headInput, metadata)
	}
	if err := c.DeleteObject(ctx, FilesBucket, "orders/123/download.zip"); err != nil {
		t.Fatal(err)
	}
	if aws.ToString(o.deleteInput.Bucket) != "private-files" {
		t.Fatalf("unexpected DELETE: %+v", o.deleteInput)
	}
}

func TestOperationsRejectInvalidInputsAndWrapErrors(t *testing.T) {
	p := &fakePresigner{err: errors.New("sign failed")}
	o := &fakeObjectClient{err: errors.New("storage failed")}
	c := &Client{presignClient: p, s3Client: o, filesBucket: "private-files", assetsBucket: "public-assets"}
	ctx := context.Background()
	if _, err := c.PresignPut(ctx, Bucket("unknown"), "key", "image/webp", time.Minute); err == nil {
		t.Fatal("unknown bucket accepted")
	}
	if _, err := c.PresignPut(ctx, AssetsBucket, "key", "invalid\ncontent-type", time.Minute); err == nil {
		t.Fatal("invalid content type accepted")
	}
	if _, err := c.PresignGet(ctx, FilesBucket, "../key", time.Minute); err == nil {
		t.Fatal("invalid key accepted")
	}
	if _, err := c.PresignGet(ctx, FilesBucket, "key", 0); err == nil {
		t.Fatal("invalid duration accepted")
	}
	if _, err := c.PresignPut(ctx, AssetsBucket, "key", "image/webp", time.Minute); !errors.Is(err, p.err) {
		t.Fatalf("PUT error = %v", err)
	}
	if _, err := c.PresignGet(ctx, FilesBucket, "key", time.Minute); !errors.Is(err, p.err) {
		t.Fatalf("GET error = %v", err)
	}
	if _, err := c.HeadObject(ctx, FilesBucket, "key"); !errors.Is(err, o.err) {
		t.Fatalf("HEAD error = %v", err)
	}
	if err := c.DeleteObject(ctx, FilesBucket, "key"); !errors.Is(err, o.err) {
		t.Fatalf("DELETE error = %v", err)
	}
}

func TestHeadAndDeletePropagateCancellation(t *testing.T) {
	o := &fakeObjectClient{}
	c := &Client{s3Client: o, filesBucket: "private-files", assetsBucket: "public-assets"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.HeadObject(ctx, FilesBucket, "orders/123/download.zip"); !errors.Is(err, context.Canceled) {
		t.Fatalf("HEAD cancellation = %v", err)
	}
	if o.headContext != ctx {
		t.Fatal("HEAD did not pass caller context")
	}
	if err := c.DeleteObject(ctx, FilesBucket, "orders/123/download.zip"); !errors.Is(err, context.Canceled) {
		t.Fatalf("DELETE cancellation = %v", err)
	}
	if o.deleteContext != ctx {
		t.Fatal("DELETE did not pass caller context")
	}
}

type recordingHTTPClient struct {
	requests []*http.Request
}

func (c *recordingHTTPClient) Do(request *http.Request) (*http.Response, error) {
	c.requests = append(c.requests, request)
	status := http.StatusOK
	headers := http.Header{"Content-Length": {"42"}, "Content-Type": {"image/webp"}}
	if request.Method == http.MethodDelete {
		status = http.StatusNoContent
		headers = make(http.Header)
	}
	return &http.Response{
		StatusCode: status, Status: http.StatusText(status), Header: headers,
		Body: io.NopCloser(strings.NewReader("")), Request: request,
	}, nil
}

func TestSDKUsesInjectedHTTPClientForObjectOperations(t *testing.T) {
	httpClient := &recordingHTTPClient{}
	c, err := newClient(validConfig(), httpClient)
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := c.HeadObject(context.Background(), FilesBucket, "orders/123/download.zip")
	if err != nil {
		t.Fatal(err)
	}
	if metadata.ContentLength != 42 || metadata.ContentType != "image/webp" {
		t.Fatalf("unexpected HEAD metadata: %+v", metadata)
	}
	if err := c.DeleteObject(context.Background(), AssetsBucket, "products/123/image.webp"); err != nil {
		t.Fatal(err)
	}
	if len(httpClient.requests) != 2 {
		t.Fatalf("HTTP requests = %d, want 2", len(httpClient.requests))
	}
	if req := httpClient.requests[0]; req.Method != http.MethodHead || req.URL.EscapedPath() != "/private-files/orders/123/download.zip" || req.Header.Get("Authorization") == "" {
		t.Fatalf("unexpected HEAD request: %s %s", req.Method, req.URL)
	}
	if req := httpClient.requests[1]; req.Method != http.MethodDelete || req.URL.EscapedPath() != "/public-assets/products/123/image.webp" || req.Header.Get("Authorization") == "" {
		t.Fatalf("unexpected DELETE request: %s %s", req.Method, req.URL)
	}
}
