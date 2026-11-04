package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
)

type Client struct {
	s3Client      objectClient
	presignClient presigner

	filesBucket  string
	assetsBucket string
}

func NewClient(cfg config.Config) (*Client, error) {
	// Unlike the SDK's BuildableClient, the standard client leaves DialContext
	// unset in js/wasm, allowing Go's net/http to use the Worker fetch API.
	return newClient(cfg, &http.Client{Timeout: 30 * time.Second})
}

func newClient(cfg config.Config, httpClient aws.HTTPClient) (*Client, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	if httpClient == nil {
		return nil, errors.New("R2 HTTP client is required")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(cfg.R2Region),
		awsconfig.WithHTTPClient(httpClient),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.R2AccessKeyID, cfg.R2SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load R2 configuration: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.R2Endpoint)
		options.UsePathStyle = true
	})

	return &Client{
		s3Client:      s3Client,
		presignClient: s3.NewPresignClient(s3Client),
		filesBucket:   cfg.ProductFilesBucket,
		assetsBucket:  cfg.ProductAssetsBucket,
	}, nil
}

func validateConfig(cfg config.Config) error {
	if strings.TrimSpace(cfg.R2AccessKeyID) == "" {
		return errors.New("R2_ACCESS_KEY_ID is required")
	}
	if strings.TrimSpace(cfg.R2SecretAccessKey) == "" {
		return errors.New("R2_SECRET_ACCESS_KEY is required")
	}
	if strings.TrimSpace(cfg.R2Endpoint) == "" {
		return errors.New("R2_ENDPOINT is required")
	}
	if cfg.R2Region != "auto" {
		return errors.New("R2_REGION must be auto for Cloudflare R2")
	}
	if strings.TrimSpace(cfg.ProductFilesBucket) == "" {
		return errors.New("PRODUCT_FILES_BUCKET is required")
	}
	if strings.TrimSpace(cfg.ProductAssetsBucket) == "" {
		return errors.New("PRODUCT_ASSETS_BUCKET is required")
	}
	if cfg.ProductFilesBucket == cfg.ProductAssetsBucket {
		return errors.New("PRODUCT_FILES_BUCKET and PRODUCT_ASSETS_BUCKET must be different")
	}

	endpoint, err := url.Parse(cfg.R2Endpoint)
	if err != nil || endpoint.Scheme != "https" || endpoint.Hostname() == "" || endpoint.User != nil ||
		endpoint.Path != "" || endpoint.RawQuery != "" || endpoint.Fragment != "" || endpoint.Opaque != "" {
		return errors.New("R2_ENDPOINT must be an HTTPS origin without a path, query, fragment, or credentials")
	}
	return nil
}

// Bucket identifies one of the two configured R2 buckets.
type Bucket string

const (
	FilesBucket  Bucket = "files"
	AssetsBucket Bucket = "assets"
)

func (c *Client) bucketName(bucket Bucket) (string, error) {
	switch bucket {
	case FilesBucket:
		return c.filesBucket, nil
	case AssetsBucket:
		return c.assetsBucket, nil
	default:
		return "", fmt.Errorf("unknown R2 bucket %q", bucket)
	}
}

// validateKey only checks the shape of a key. Callers must construct keys from
// trusted server-side identifiers after authorizing the requested operation.
func validateKey(key string) error {
	if key == "" || strings.TrimSpace(key) != key || strings.HasPrefix(key, "/") || strings.Contains(key, "\\") {
		return errors.New("R2 object key must be a non-empty relative path")
	}
	for _, segment := range strings.Split(key, "/") {
		if segment == "" || segment == "." || segment == ".." || strings.IndexFunc(segment, func(r rune) bool { return r < 0x20 || r == 0x7f }) >= 0 {
			return errors.New("R2 object key contains an invalid path segment")
		}
	}
	return nil
}
