package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type objectClient interface {
	HeadObject(context.Context, *s3.HeadObjectInput, ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
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
