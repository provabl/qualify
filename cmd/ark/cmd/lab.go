// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/provabl/qualify/internal/database"
	"github.com/provabl/qualify/internal/training"
)

func init() {
	rootCmd.AddCommand(labCmd())
}

func labCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lab",
		Short: "Manage lab membership and identity tags",
		Long:  "Set lab membership, admin level, and IAM role registration for a researcher.",
	}
	cmd.AddCommand(labSetupCmd())
	cmd.AddCommand(labRegisterRoleCmd())
	return cmd
}

func labSetupCmd() *cobra.Command {
	var userID, labID, adminLevel, region string

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Assign a researcher to a lab and write attest:lab-id and attest:admin-level IAM tags",
		Long: `Assign a researcher to a lab environment.

Writes two IAM role tags that attest's Cedar PDP evaluates:
  attest:lab-id       = <lab-id>
  attest:admin-level  = none | env | sre

The researcher must have their IAM role ARN registered first.
Run 'qualify lab register-role' if not already done.

Examples:
  qualify lab setup --user alice@mru.edu --lab-id chen-quantum-lab
  qualify lab setup --user bob@mru.edu   --lab-id chen-quantum-lab --admin-level env`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLabSetup(userID, labID, adminLevel, region)
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "user ID or email (required)")
	cmd.Flags().StringVar(&labID, "lab-id", "", "lab identifier (required)")
	cmd.Flags().StringVar(&adminLevel, "admin-level", "none", "admin level: none | env | sre")
	cmd.Flags().StringVar(&region, "region", "us-east-1", "AWS region")

	_ = cmd.MarkFlagRequired("user")
	_ = cmd.MarkFlagRequired("lab-id")

	return cmd
}

func labRegisterRoleCmd() *cobra.Command {
	var userID, roleARN string

	cmd := &cobra.Command{
		Use:   "register-role",
		Short: "Register an IAM role ARN for a researcher",
		Long: `Register the IAM role ARN associated with a researcher's account.
This ARN is used by qualify when writing attest:* training and identity tags.

Example:
  qualify lab register-role --user alice@mru.edu \
    --role-arn arn:aws:iam::123456789012:role/researcher-alice`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLabRegisterRole(userID, roleARN)
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "user ID or email (required)")
	cmd.Flags().StringVar(&roleARN, "role-arn", "", "IAM role ARN (required)")

	_ = cmd.MarkFlagRequired("user")
	_ = cmd.MarkFlagRequired("role-arn")

	return cmd
}

func runLabSetup(userID, labID, adminLevel, region string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	svc := training.NewServiceWithIAM(context.Background(), db, region)

	if err := svc.SetIdentityTags(context.Background(), userID, labID, adminLevel); err != nil {
		return fmt.Errorf("set identity tags: %w", err)
	}

	fmt.Printf("✓ attest:lab-id       = %s\n", labID)
	fmt.Printf("✓ attest:admin-level  = %s\n", adminLevel)
	fmt.Printf("  Written to IAM role for user %s\n", userID)
	return nil
}

func runLabRegisterRole(userID, roleARN string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	svc := training.NewService(db)
	if err := svc.RegisterRoleARN(context.Background(), userID, roleARN); err != nil {
		return fmt.Errorf("register role ARN: %w", err)
	}

	fmt.Printf("✓ Registered IAM role ARN for %s\n", userID)
	fmt.Printf("  %s\n", roleARN)
	return nil
}

func openDB() (*database.DB, error) {
	port := 5432
	if p := os.Getenv("DB_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	return database.New(database.Config{
		Host:     getEnvDefault("DB_HOST", "localhost"),
		Port:     port,
		User:     getEnvDefault("DB_USER", "ark"),
		Password: getEnvDefault("DB_PASSWORD", ""),
		DBName:   getEnvDefault("DB_NAME", "ark"),
		SSLMode:  getEnvDefault("DB_SSLMODE", "disable"),
	})
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
