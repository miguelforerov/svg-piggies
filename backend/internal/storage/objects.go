package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type objectClient interface {
	HeadObject(context.Context, *s3.HeadObjectInput, ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// ReadObject downloads at most maxBytes plus one byte to detect an oversized
// response even when the server omits or misreports Content-Length.
func (c *Client) ReadObject(ctx context.Context, bucket Bucket, key, expectedETag string, maxBytes int64) ([]byte, error) {
	bucketName, err := c.bucketName(bucket)
	if err != nil {
		return nil, err
	}
	if err := validateKey(key); err != nil {
		return nil, err
	}
	if maxBytes <= 0 {
		return nil, errors.New("R2 read limit must be positive")
	}
	input := &s3.GetObjectInput{Bucket: aws.String(bucketName), Key: aws.String(key)}
	if expectedETag != "" {
		input.IfMatch = aws.String(expectedETag)
	}
	output, err := c.s3Client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("read R2 object: %w", err)
	}
	if output == nil || output.Body == nil {
		return nil, errors.New("read R2 object: empty response")
	}
	defer output.Body.Close()
	if output.ContentLength != nil && *output.ContentLength > maxBytes {
		return nil, errors.New("read R2 object: size exceeds limit")
	}
	data, err := io.ReadAll(io.LimitReader(output.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read R2 object body: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, errors.New("read R2 object: size exceeds limit")
	}
	return data, nil
}

type ObjectMetadata struct {
	ContentLength int64
	ContentType   string
	ETag          string
	LastModified  time.Time
}

func (c *Client) HeadObject(ctx context.Context, bucket Bucket, key string) (ObjectMetadata, error) {
	bucketName, err := c.bucketName(bucket)
	if err != nil {
		return ObjectMetadata{}, err
	}
	if err := validateKey(key); err != nil {
		return ObjectMetadata{}, err
	}
	output, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName), Key: aws.String(key),
	})
	if err != nil {
		return ObjectMetadata{}, fmt.Errorf("head R2 object: %w", err)
	}
	if output == nil {
		return ObjectMetadata{}, errors.New("head R2 object: empty response")
	}
	return ObjectMetadata{
		ContentLength: aws.ToInt64(output.ContentLength),
		ContentType:   aws.ToString(output.ContentType),
		ETag:          aws.ToString(output.ETag),
		LastModified:  aws.ToTime(output.LastModified),
	}, nil
}

func (c *Client) DeleteObject(ctx context.Context, bucket Bucket, key string) error {
	bucketName, err := c.bucketName(bucket)
	if err != nil {
		return err
	}
	if err := validateKey(key); err != nil {
		return err
	}
	_, err = c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName), Key: aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete R2 object: %w", err)
	}
	return nil
}
