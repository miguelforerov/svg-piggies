package storage

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const maxPresignDuration = 7 * 24 * time.Hour

// Public product assets accept raster formats only. SVG and HTML can contain
// active content and need a separate review before being served publicly.
var allowedAssetMediaTypes = map[string]bool{
	"image/png": true, "image/jpeg": true, "image/webp": true, "image/avif": true,
}

type presigner interface {
	PresignPutObject(context.Context, *s3.PutObjectInput, ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
	PresignGetObject(context.Context, *s3.GetObjectInput, ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

// PresignedRequest includes headers that the browser must send with the URL.
// Treat the URL as a temporary credential and disclose it only to an authorized user.
type PresignedRequest struct {
	URL     string
	Headers http.Header
}

func validateDuration(duration time.Duration) error {
	if duration < time.Second || duration > maxPresignDuration || duration%time.Second != 0 {
		return fmt.Errorf("R2 presigned URL duration must be whole seconds from 1 second to 7 days; got %s", duration)
	}
	return nil
}

// PresignPut grants a temporary, reusable upload for a server-selected key.
// Before calling it, an endpoint must authorize the operation, choose the bucket
// from trusted server configuration, and generate the key on the server. The
// browser must not choose either bucket or key. Verify size with HEAD afterward.
func (c *Client) PresignPut(ctx context.Context, bucket Bucket, key, contentType string, duration time.Duration) (PresignedRequest, error) {
	bucketName, err := c.bucketName(bucket)
	if err != nil {
		return PresignedRequest{}, err
	}
	if err := validateKey(key); err != nil {
		return PresignedRequest{}, err
	}
	if err := validateDuration(duration); err != nil {
		return PresignedRequest{}, err
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return PresignedRequest{}, fmt.Errorf("invalid R2 upload content type: %w", err)
	}
	if bucket == AssetsBucket && !allowedAssetMediaTypes[mediaType] {
		return PresignedRequest{}, fmt.Errorf("R2 asset content type %q is not allowed", mediaType)
	}

	request, err := c.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName), Key: aws.String(key), ContentType: aws.String(contentType),
	}, func(options *s3.PresignOptions) { options.Expires = duration })
	if err != nil {
		return PresignedRequest{}, fmt.Errorf("presign R2 PUT: %w", err)
	}
	if request == nil {
		return PresignedRequest{}, errors.New("presign R2 PUT: empty response")
	}
	return PresignedRequest{URL: request.URL, Headers: request.SignedHeader.Clone()}, nil
}

// PresignGet grants a temporary download for a server-selected key. An endpoint
// must authorize the user and derive the bucket and key from server-side data.
func (c *Client) PresignGet(ctx context.Context, bucket Bucket, key string, duration time.Duration) (PresignedRequest, error) {
	bucketName, err := c.bucketName(bucket)
	if err != nil {
		return PresignedRequest{}, err
	}
	if err := validateKey(key); err != nil {
		return PresignedRequest{}, err
	}
	if err := validateDuration(duration); err != nil {
		return PresignedRequest{}, err
	}

	request, err := c.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName), Key: aws.String(key),
	}, func(options *s3.PresignOptions) { options.Expires = duration })
	if err != nil {
		return PresignedRequest{}, fmt.Errorf("presign R2 GET: %w", err)
	}
	if request == nil {
		return PresignedRequest{}, errors.New("presign R2 GET: empty response")
	}
	return PresignedRequest{URL: request.URL, Headers: request.SignedHeader.Clone()}, nil
}
