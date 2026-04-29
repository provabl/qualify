package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/provabl/qualify/internal/agent/store"
)

// Client wraps AWS SDK clients and configuration
type Client struct {
	S3Client  *s3.Client
	STSClient *sts.Client
	Config    aws.Config
	Region    string
}

// NewClientFromCredentials creates an AWS client from stored credentials
func NewClientFromCredentials(ctx context.Context, creds *store.Credentials, region string) (*Client, error) {
	if creds == nil {
		return nil, fmt.Errorf("credentials are nil")
	}

	// Use region from credentials if not explicitly provided
	if region == "" {
		region = creds.Region
	}
	if region == "" {
		region = "us-east-1" // Default region
	}

	// Create static credentials provider
	credsProvider := credentials.NewStaticCredentialsProvider(
		creds.AccessKeyID,
		creds.SecretAccessKey,
		creds.SessionToken,
	)

	// Build AWS config
	cfg := aws.Config{
		Region:      region,
		Credentials: credsProvider,
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Create STS client for credential validation
	stsClient := sts.NewFromConfig(cfg)

	return &Client{
		S3Client:  s3Client,
		STSClient: stsClient,
		Config:    cfg,
		Region:    region,
	}, nil
}

// ValidateCredentials checks if the credentials are valid by calling STS GetCallerIdentity
func (c *Client) ValidateCredentials(ctx context.Context) error {
	_, err := c.STSClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}
	return nil
}
