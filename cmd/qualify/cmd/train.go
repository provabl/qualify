// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	rootCmd.AddCommand(trainCmd())
}

// trainCmd is the parent command for training operations.
func trainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "train",
		Short: "Training module commands",
		Long:  "Manage and run compliance training modules.",
	}
	cmd.AddCommand(trainRequiredCmd())
	cmd.AddCommand(trainStatusCmd())
	cmd.AddCommand(trainListCmd())
	return cmd
}

// trainRequiredCmd returns the training modules required for the active attest frameworks.
func trainRequiredCmd() *cobra.Command {
	var attDir string
	var frameworkIDs []string

	cmd := &cobra.Command{
		Use:   "required",
		Short: "Show training modules required for your SRE's active frameworks",
		Long: `Reads .attest/sre.yaml (or --attest-dir) to determine which compliance
frameworks are active, then shows which qualify training modules are required.

Examples:
  # Show required training for current SRE
  qualify train required

  # Check requirements for specific frameworks
  qualify train required --framework nist-800-171-r2 --framework hipaa

  # Use a different attest directory
  qualify train required --attest-dir /path/to/.attest`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrainRequired(attDir, frameworkIDs)
		},
	}
	cmd.Flags().StringVar(&attDir, "attest-dir", ".attest", "path to attest directory (.attest/sre.yaml)")
	cmd.Flags().StringArrayVar(&frameworkIDs, "framework", nil, "specific framework IDs (overrides sre.yaml)")
	return cmd
}

// frameworkModuleMap maps framework IDs to the training modules they require.
// This mirrors the database migration 000007 but is available offline (no DB needed).
var frameworkModuleMap = map[string][]string{
	"nist-800-171-r2":       {"cui-fundamentals", "security-awareness", "data-classification"},
	"cmmc-level-1":          {"security-awareness", "cui-fundamentals"},
	"cmmc-level-2":          {"cui-fundamentals", "security-awareness", "data-classification"},
	"cmmc-level-3":          {"cui-fundamentals", "security-awareness", "data-classification"},
	"hipaa":                 {"hipaa-privacy-security", "security-awareness", "data-classification"},
	"ferpa":                 {"ferpa-basics", "security-awareness"},
	"itar":                  {"itar-export-control", "countries-of-concern-awareness"},
	"nih-gds":               {"nih-research-security", "countries-of-concern-awareness", "cui-fundamentals"},
	"nih-research-security": {"nih-research-security"},
	"fedramp-moderate":      {"security-awareness", "data-classification"},
	"fedramp-high":          {"security-awareness", "data-classification"},
	"gdpr":                  {"security-awareness", "data-classification"},
	"fisma-moderate":        {"security-awareness", "data-classification"},
	"cjis":                  {"security-awareness", "data-classification"},
}

// moduleDescriptions provides human-readable names for training modules.
var moduleDescriptions = map[string]string{
	"cui-fundamentals":               "Controlled Unclassified Information (CUI)",
	"hipaa-privacy-security":         "HIPAA Security & Privacy",
	"security-awareness":             "Security Awareness (annual)",
	"ferpa-basics":                   "FERPA",
	"itar-export-control":            "ITAR / Export Controls",
	"data-classification":            "Data Classification",
	"nih-research-security":          "NIH Research Security (NOT-OD-26-017)",
	"countries-of-concern-awareness": "Countries-of-Concern Awareness (NOT-OD-25-083)",
}

// sreYAML is a minimal struct to read active frameworks from .attest/sre.yaml.
type sreYAML struct {
	OrgID      string `yaml:"org_id"`
	Name       string `yaml:"name"`
	Frameworks []struct {
		ID string `yaml:"id"`
	} `yaml:"frameworks"`
}

func runTrainRequired(attDir string, overrideFrameworks []string) error {
	var activeFrameworks []string

	if len(overrideFrameworks) > 0 {
		activeFrameworks = overrideFrameworks
	} else {
		// Read from .attest/sre.yaml
		srePath := filepath.Join(attDir, "sre.yaml")
		data, err := os.ReadFile(srePath) // #nosec G304 — operator-controlled path
		if os.IsNotExist(err) {
			fmt.Printf("No .attest/sre.yaml found at %s\n\n", srePath)
			fmt.Println("To check requirements for specific frameworks:")
			fmt.Println("  qualify train required --framework nist-800-171-r2 --framework hipaa")
			fmt.Println()
			fmt.Println("To initialize attest first:")
			fmt.Println("  attest init --region us-east-1")
			return nil
		}
		if err != nil {
			return fmt.Errorf("read sre.yaml: %w", err)
		}

		var sre sreYAML
		if err := yaml.Unmarshal(data, &sre); err != nil {
			return fmt.Errorf("parse sre.yaml: %w", err)
		}
		for _, fw := range sre.Frameworks {
			activeFrameworks = append(activeFrameworks, fw.ID)
		}

		if len(activeFrameworks) == 0 {
			fmt.Println("No frameworks active in .attest/sre.yaml")
			fmt.Println("Run 'attest frameworks add <id>' to activate frameworks.")
			return nil
		}
		if sre.Name != "" {
			fmt.Printf("SRE: %s\n", sre.Name)
		}
	}

	fmt.Printf("Active frameworks: %s\n\n", strings.Join(activeFrameworks, ", "))

	// Collect required modules (dedup).
	required := make(map[string]bool)
	fwForModule := make(map[string][]string) // module → frameworks that require it
	for _, fwID := range activeFrameworks {
		modules, ok := frameworkModuleMap[fwID]
		if !ok {
			continue
		}
		for _, m := range modules {
			required[m] = true
			fwForModule[m] = append(fwForModule[m], fwID)
		}
	}

	if len(required) == 0 {
		fmt.Println("No training modules required for active frameworks.")
		fmt.Println("(Framework-specific requirements may not be mapped yet.)")
		return nil
	}

	fmt.Printf("Required training modules (%d):\n\n", len(required))

	// Display in logical order.
	order := []string{
		"security-awareness",
		"data-classification",
		"cui-fundamentals",
		"hipaa-privacy-security",
		"ferpa-basics",
		"itar-export-control",
		"nih-research-security",
		"countries-of-concern-awareness",
	}
	for _, m := range order {
		if !required[m] {
			continue
		}
		desc := moduleDescriptions[m]
		if desc == "" {
			desc = m
		}
		fws := strings.Join(fwForModule[m], ", ")
		fmt.Printf("  %-38s  required by: %s\n", desc, fws)
	}

	fmt.Println()
	fmt.Println("Complete training:")
	fmt.Println("  qualify train start <module-id>")
	fmt.Println()
	fmt.Println("Check your current status:")
	fmt.Println("  qualify train status")

	return nil
}

// trainStatusCmd shows the current training completion status.
func trainStatusCmd() *cobra.Command {
	var userID string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show training completion status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrainStatus(userID)
		},
	}
	cmd.Flags().StringVar(&userID, "user", "", "user ID or email (admin: check any user)")
	return cmd
}

func runTrainStatus(userID string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	query := `
		SELECT m.name, m.title, p.status, p.score, p.completed_at, p.expires_at
		FROM training_modules m
		LEFT JOIN training_progress p ON p.module_id = m.name
		WHERE ($1 = '' OR p.user_id = $1)
		  AND m.required_for_frameworks != '[]'
		ORDER BY m.name`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("query training status: %w", err)
	}
	defer rows.Close()

	fmt.Printf("Training status%s:\n\n", func() string {
		if userID != "" {
			return " for " + userID
		}
		return ""
	}())

	found := false
	for rows.Next() {
		found = true
		var name, title string
		var status, score, completedAt, expiresAt *string
		if scanErr := rows.Scan(&name, &title, &status, &score, &completedAt, &expiresAt); scanErr != nil {
			continue
		}
		if status == nil || *status == "" {
			fmt.Printf("  ✗ %-40s  not started\n", title)
		} else if *status == "completed" {
			expiry := ""
			if expiresAt != nil {
				expiry = "  expires " + *expiresAt
			}
			fmt.Printf("  ✓ %-40s  complete%s\n", title, expiry)
		} else {
			fmt.Printf("  … %-40s  in progress\n", title)
		}
	}
	if !found {
		fmt.Println("  No training records found.")
		fmt.Println("  Run 'qualify train required' to see what's needed.")
	}
	return rows.Err()
}

// trainListCmd lists available training modules.
func trainListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available training modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrainList()
		},
	}
}

func runTrainList() error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.QueryContext(context.Background(),
		`SELECT name, title, category, difficulty, estimated_minutes FROM training_modules ORDER BY category, name`)
	if err != nil {
		return fmt.Errorf("list modules: %w", err)
	}
	defer rows.Close()

	fmt.Println("Available training modules:")

	lastCategory := ""
	for rows.Next() {
		var name, title, category, difficulty string
		var mins int
		if err := rows.Scan(&name, &title, &category, &difficulty, &mins); err != nil {
			continue
		}
		if category != lastCategory {
			fmt.Printf("  %s\n", strings.ToUpper(category))
			lastCategory = category
		}
		fmt.Printf("    %-38s  %s (%d min)\n", name, difficulty, mins)
	}
	return rows.Err()
}


