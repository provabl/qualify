// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/provabl/qualify/internal/database"
	"github.com/provabl/qualify/internal/localaudit"
	"github.com/provabl/qualify/internal/training"
)

// localAuditLogger returns a localaudit.Logger or nil if unavailable.
func localAuditLogger() (*localaudit.Logger, error) {
	return localaudit.New()
}

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
	cmd.AddCommand(labRecordCheckCmd())
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

func labRecordCheckCmd() *cobra.Command {
	var userID, countryCode, performedBy, region string

	cmd := &cobra.Command{
		Use:   "record-check",
		Short: "Record a countries-of-concern compliance check for a researcher",
		Long: `Record that a compliance officer has performed the countries-of-concern
check required by NIH NOT-OD-25-083 for a researcher.

Writes three IAM role tags:
  attest:country              = <ISO-2 country code>
  attest:coc-check-current    = true
  attest:coc-check-expiry     = <1 year from now>

Also stores the check details in the qualify database for audit purposes.
attest's Cedar PDP evaluates attest:country against the countries-of-concern
list for NIH GDS access control and ITAR deemed-export enforcement.

Examples:
  # Record that alice's institution is in the US
  qualify lab record-check --user alice@mru.edu --country US --performed-by compliance@mru.edu

  # Record for a researcher affiliated with a designated country
  qualify lab record-check --user bob@intl.edu --country CN --performed-by compliance@mru.edu`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLabRecordCheck(userID, countryCode, performedBy, region)
		},
	}
	cmd.Flags().StringVar(&userID, "user", "", "researcher user ID or email (required)")
	cmd.Flags().StringVar(&countryCode, "country", "", "ISO 3166-1 alpha-2 country code of institutional affiliation (required)")
	cmd.Flags().StringVar(&performedBy, "performed-by", "", "compliance officer performing the check (required)")
	cmd.Flags().StringVar(&region, "region", "us-east-1", "AWS region for IAM tag writes")
	_ = cmd.MarkFlagRequired("user")
	_ = cmd.MarkFlagRequired("country")
	_ = cmd.MarkFlagRequired("performed-by")
	return cmd
}

func runLabRecordCheck(userID, countryCode, performedBy, region string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	alog, _ := localAuditLogger()

	svc := training.NewServiceWithIAM(context.Background(), db, region)
	if err := svc.RecordCountryCheck(context.Background(), userID, countryCode, performedBy); err != nil {
		return fmt.Errorf("record country check: %w", err)
	}

	upper := strings.ToUpper(countryCode)
	fmt.Printf("✓ attest:country              = %s\n", upper)
	fmt.Printf("✓ attest:coc-check-current    = true\n")
	fmt.Printf("✓ attest:coc-check-expiry     = <1 year from now>\n")
	fmt.Printf("  Recorded for: %s  (performed by: %s)\n", userID, performedBy)
	fmt.Printf("  Database: institutional_affiliation_country updated\n")

	if alog != nil {
		alog.Log("country_check_recorded", userID, "", map[string]any{
			"country":      upper,
			"performed_by": performedBy,
		})
	}
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
		User:     getEnvDefault("DB_USER", "qualify"),
		Password: getEnvDefault("DB_PASSWORD", ""),
		DBName:   getEnvDefault("DB_NAME", "qualify"),
		SSLMode:  getEnvDefault("DB_SSLMODE", "disable"),
	})
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
