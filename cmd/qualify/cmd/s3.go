// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(s3Cmd)
	s3Cmd.AddCommand(s3CreateBucketCmd)

	// Flags for create-bucket command
	s3CreateBucketCmd.Flags().String("region", "us-east-1", "AWS region for bucket")
	s3CreateBucketCmd.Flags().String("encryption", "AES256", "Encryption type: AES256 or aws:kms")
	s3CreateBucketCmd.Flags().String("kms-key-id", "", "KMS key ID (required if encryption is aws:kms)")
	s3CreateBucketCmd.Flags().Bool("versioning", false, "Enable bucket versioning")
	s3CreateBucketCmd.Flags().String("profile", "default", "AWS credential profile to use")
}

var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Manage AWS S3 resources",
	Long: `Create and manage AWS S3 buckets and objects.

All operations require AWS credentials to be configured. Use 'ark credentials set'
to store your AWS credentials.`,
}

var s3CreateBucketCmd = &cobra.Command{
	Use:   "create-bucket <bucket-name>",
	Short: "Create an S3 bucket",
	Long: `Create a new S3 bucket with the specified configuration.

Bucket names must be globally unique across all AWS accounts and must follow
AWS naming rules:
  - Between 3 and 63 characters long
  - Contain only lowercase letters, numbers, hyphens, and periods
  - Start and end with a letter or number
  - Not formatted as an IP address (e.g., 192.168.1.1)

Examples:
  # Create bucket with default settings (AES256 encryption, us-east-1)
  ark s3 create-bucket my-research-data

  # Create bucket in specific region with versioning
  ark s3 create-bucket my-bucket --region us-west-2 --versioning

  # Create bucket with KMS encryption
  ark s3 create-bucket my-secure-bucket --encryption aws:kms --kms-key-id <key-id>

  # Use non-default credential profile
  ark s3 create-bucket my-bucket --profile production

Note: Some operations may require training completion. If blocked, complete
the required training modules and try again.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bucketName := args[0]

		// Validate bucket name
		if err := validateBucketName(bucketName); err != nil {
			ExitWithError(err)
		}

		// Get flags
		region, _ := cmd.Flags().GetString("region")
		encryption, _ := cmd.Flags().GetString("encryption")
		kmsKeyID, _ := cmd.Flags().GetString("kms-key-id")
		versioning, _ := cmd.Flags().GetBool("versioning")
		profile, _ := cmd.Flags().GetString("profile")

		// Validate encryption settings
		if encryption != "AES256" && encryption != "aws:kms" {
			ExitWithError(fmt.Errorf("encryption must be 'AES256' or 'aws:kms', got: %s", encryption))
		}
		if encryption == "aws:kms" && kmsKeyID == "" {
			ExitWithError(fmt.Errorf("--kms-key-id is required when using aws:kms encryption"))
		}

		// Ensure agent is running
		if err := EnsureAgentRunning(); err != nil {
			ExitWithError(fmt.Errorf("agent not available: %w", err))
		}

		// Prepare request
		reqBody := map[string]interface{}{
			"bucket_name": bucketName,
			"region":      region,
			"encryption": map[string]string{
				"type":       encryption,
				"kms_key_id": kmsKeyID,
			},
			"versioning_enabled": versioning,
			"profile":            profile,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			ExitWithError(fmt.Errorf("failed to marshal request: %w", err))
		}

		// Send request to agent
		url := "http://127.0.0.1:8737/api/s3/buckets"
		resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			ExitWithError(fmt.Errorf("failed to create bucket: %w", err))
		}
		defer resp.Body.Close()

		// Parse response
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			ExitWithError(fmt.Errorf("failed to parse response: %w", err))
		}

		// Handle response based on status code
		switch resp.StatusCode {
		case http.StatusOK, http.StatusCreated:
			// Success
			fmt.Println("✓ S3 bucket created successfully")
			fmt.Println()
			fmt.Printf("  Name:      %s\n", result["bucket_name"])
			fmt.Printf("  Region:    %s\n", result["region"])
			if location, ok := result["location"].(string); ok {
				fmt.Printf("  Location:  %s\n", location)
			}
			if createdAt, ok := result["created_at"].(string); ok {
				if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
					fmt.Printf("  Created:   %s\n", t.Format("2006-01-02 15:04:05 UTC"))
				}
			}

		case http.StatusForbidden:
			// Training gate blocked
			status, _ := result["status"].(string)
			if status == "blocked" {
				fmt.Println("✗ Training required before creating S3 buckets")
				fmt.Println()
				fmt.Println("You must complete the following training modules:")
				fmt.Println()

				if modules, ok := result["required_modules"].([]interface{}); ok && len(modules) > 0 {
					for i, mod := range modules {
						if m, ok := mod.(map[string]interface{}); ok {
							title := m["title"].(string)
							name := m["name"].(string)
							minutes := int(m["estimated_minutes"].(float64))

							fmt.Printf("  %d. %s (%d minutes)\n", i+1, title, minutes)
							fmt.Printf("     Start training: ark training start %s\n", name)
							fmt.Printf("     Or visit: http://localhost:8080/training/%s\n", name)
							fmt.Println()
						}
					}
				}

				fmt.Println("After completing training, run your command again.")
				os.Exit(1)
			}

		default:
			// Error
			if errMsg, ok := result["error"].(string); ok {
				ExitWithError(fmt.Errorf("failed to create bucket: %s", errMsg))
			} else {
				ExitWithError(fmt.Errorf("failed to create bucket: unexpected response (status %d)", resp.StatusCode))
			}
		}
	},
}

// validateBucketName validates an S3 bucket name according to AWS rules
func validateBucketName(name string) error {
	if len(name) < 3 || len(name) > 63 {
		return fmt.Errorf("bucket name must be between 3 and 63 characters long")
	}

	if name[0] == '-' || name[len(name)-1] == '-' {
		return fmt.Errorf("bucket name cannot start or end with a hyphen")
	}

	if name[0] == '.' || name[len(name)-1] == '.' {
		return fmt.Errorf("bucket name cannot start or end with a period")
	}

	// Check for invalid characters
	for i, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '.') {
			return fmt.Errorf("bucket name can only contain lowercase letters, numbers, hyphens, and periods")
		}

		// Check for consecutive periods
		if c == '.' && i > 0 && name[i-1] == '.' {
			return fmt.Errorf("bucket name cannot contain consecutive periods")
		}
	}

	// Check if it looks like an IP address
	parts := strings.Split(name, ".")
	if len(parts) == 4 {
		allNumeric := true
		for _, part := range parts {
			if len(part) == 0 || len(part) > 3 {
				allNumeric = false
				break
			}
			for _, c := range part {
				if c < '0' || c > '9' {
					allNumeric = false
					break
				}
			}
		}
		if allNumeric {
			return fmt.Errorf("bucket name cannot be formatted as an IP address")
		}
	}

	return nil
}
