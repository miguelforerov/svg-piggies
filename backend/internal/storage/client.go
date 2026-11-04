package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
)

type Client struct {
	s3Client      *s3.Client
	presignClient *s3.PresignClient

	filesBucket  string
	assetsBucket string
}

func NewClient(cfg config.Config) (*Client, error) {
	if strings.TrimSpace(cfg.R2AccessKeyID) == "" {
		return nil, errors.New("R2_ACCESS_KEY_ID is required")
	}
	if strings.TrimSpace(cfg.R2SecretAccessKey) == "" {
		return nil, errors.New("R2_SECRET_ACCESS_KEY is required")
	}
	if strings.TrimSpace(cfg.R2Endpoint) == "" {
		return nil, errors.New("R2_ENDPOINT is required")
	}
	if strings.TrimSpace(cfg.R2Region) == "" {
		return nil, errors.New("R2_REGION is required")
	}
	if strings.TrimSpace(cfg.ProductFilesBucket) == "" {
		return nil, errors.New("PRODUCT_FILES_BUCKET is required")
	}
	if strings.TrimSpace(cfg.ProductAssetsBucket) == "" {
		return nil, errors.New("PRODUCT_ASSETS_BUCKET is required")
	}

	if !strings.HasPrefix(cfg.R2Endpoint, "https://") {
		return nil, errors.New("R2_ENDPOINT must use HTTPS")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(cfg.R2Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.R2AccessKeyID,
				cfg.R2SecretAccessKey,
				"",
			),
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
