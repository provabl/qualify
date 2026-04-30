// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package aws

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// CreateBucketInput represents parameters for bucket creation
type CreateBucketInput struct {
	BucketName        string
	Region            string
	EncryptionType    string // "AES256" or "aws:kms"
	KMSKeyID          string // Optional, for aws:kms encryption
	VersioningEnabled bool
}

// CreateBucketOutput represents the result of bucket creation
type CreateBucketOutput struct {
	BucketName string    `json:"bucket_name"`
	Region     string    `json:"region"`
	Location   string    `json:"location"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateBucket creates an S3 bucket with specified configuration
func CreateBucket(ctx context.Context, client *Client, input CreateBucketInput) (*CreateBucketOutput, error) {
	// Validate bucket name
	if err := validateBucketName(input.BucketName); err != nil {
		return nil, err
	}

	// Use client's region if not specified
	region := input.Region
	if region == "" {
		region = client.Region
	}

	// Create bucket input
	createInput := &s3.CreateBucketInput{
		Bucket: aws.String(input.BucketName),
	}

	// For regions other than us-east-1, specify location constraint
	if region != "us-east-1" {
		createInput.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}

	// Create the bucket
	_, err := client.S3Client.CreateBucket(ctx, createInput)
	if err != nil {
		return nil, translateS3Error(err)
	}

	// Configure encryption if requested
	if input.EncryptionType != "" && input.EncryptionType != "none" {
		if err := configureBucketEncryption(ctx, client, input.BucketName, input.EncryptionType, input.KMSKeyID); err != nil {
			return nil, fmt.Errorf("configure encryption: %w", err)
		}
	}

	// Enable versioning if requested
	if input.VersioningEnabled {
		if err := enableBucketVersioning(ctx, client, input.BucketName); err != nil {
			return nil, fmt.Errorf("enable versioning: %w", err)
		}
	}

	// Build response
	output := &CreateBucketOutput{
		BucketName: input.BucketName,
		Region:     region,
		Location:   fmt.Sprintf("http://%s.s3.amazonaws.com/", input.BucketName),
		CreatedAt:  time.Now().UTC(),
	}

	return output, nil
}

// configureBucketEncryption sets encryption configuration for a bucket
func configureBucketEncryption(ctx context.Context, client *Client, bucket, encType, kmsKeyID string) error {
	var rule types.ServerSideEncryptionRule

	if encType == "aws:kms" {
		rule = types.ServerSideEncryptionRule{
			ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
				SSEAlgorithm:   types.ServerSideEncryptionAwsKms,
				KMSMasterKeyID: aws.String(kmsKeyID),
			},
		}
	} else {
		// Default to AES256
		rule = types.ServerSideEncryptionRule{
			ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
				SSEAlgorithm: types.ServerSideEncryptionAes256,
			},
		}
	}

	_, err := client.S3Client.PutBucketEncryption(ctx, &s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucket),
		ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
			Rules: []types.ServerSideEncryptionRule{rule},
		},
	})

	return err
}

// enableBucketVersioning enables versioning for a bucket
func enableBucketVersioning(ctx context.Context, client *Client, bucket string) error {
	_, err := client.S3Client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucket),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled,
		},
	})
	return err
}

// validateBucketName validates S3 bucket naming rules
func validateBucketName(name string) error {
	if len(name) < 3 || len(name) > 63 {
		return fmt.Errorf("bucket name must be between 3 and 63 characters")
	}

	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("bucket name cannot start or end with hyphen")
	}

	if strings.Contains(name, "..") {
		return fmt.Errorf("bucket name cannot contain consecutive periods")
	}

	// Check for uppercase letters
	if strings.ToLower(name) != name {
		return fmt.Errorf("bucket name must be lowercase")
	}

	return nil
}

// translateS3Error translates AWS SDK errors to user-friendly messages
func translateS3Error(err error) error {
	if err == nil {
		return nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()

		switch code {
		case "BucketAlreadyExists":
			return fmt.Errorf("bucket name already taken globally (S3 bucket names must be unique across all AWS accounts)")
		case "BucketAlreadyOwnedByYou":
			return fmt.Errorf("you already own a bucket with this name")
		case "InvalidBucketName":
			return fmt.Errorf("invalid bucket name (check naming rules: 3-63 chars, lowercase, no consecutive periods)")
		case "AccessDenied":
			return fmt.Errorf("permission denied (check your AWS credentials have s3:CreateBucket permission)")
		case "TooManyBuckets":
			return fmt.Errorf("bucket limit reached (AWS allows 100 buckets per account by default)")
		default:
			return fmt.Errorf("AWS error (%s): %s", code, apiErr.ErrorMessage())
		}
	}

	return fmt.Errorf("S3 operation failed: %w", err)
}
