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
	rootCmd.AddCommand(onboardCmd())
}

// onboardCmd walks a new user through the full qualify setup:
// register IAM role → show required training → lab assignment → next steps.
func onboardCmd() *cobra.Command {
	var email string
	var roleARN string
	var labID string
	var adminLevel string
	var attDir string

	cmd := &cobra.Command{
		Use:   "onboard",
		Short: "Guide a new user through the full qualify setup",
		Long: `Onboard a new user to a Secure Research Environment.
Walks through: IAM role registration, required training modules
(based on active frameworks), and lab assignment.

Examples:
  qualify onboard \
    --email alice@mru.edu \
    --role-arn arn:aws:iam::123456789012:role/researcher-alice \
    --lab-id chen-quantum-lab

  # Dry run — show what would be configured
  qualify onboard --email alice@mru.edu --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			return runOnboard(email, roleARN, labID, adminLevel, attDir, dryRun)
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "user email address (required)")
	cmd.Flags().StringVar(&roleARN, "role-arn", "", "IAM role ARN for the user")
	cmd.Flags().StringVar(&labID, "lab-id", "", "lab or environment identifier")
	cmd.Flags().StringVar(&adminLevel, "admin-level", "none", "admin level: none | env | sre")
	cmd.Flags().StringVar(&attDir, "attest-dir", ".attest", "path to .attest directory")
	cmd.Flags().Bool("dry-run", false, "show what would be configured without making changes")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func runOnboard(email, roleARN, labID, adminLevel, attDir string, dryRun bool) error {
	prefix := ""
	if dryRun {
		prefix = "[dry-run] "
	}

	fmt.Printf("Onboarding %s...\n\n", email)

	// ── Step 1: Register IAM role ──────────────────────────────────────────
	fmt.Println("── Step 1: IAM role registration")
	if roleARN == "" {
		fmt.Println("  ⚠ --role-arn not provided")
		fmt.Printf("  To register later: qualify lab register-role --user %s --role-arn <arn>\n", email)
	} else {
		if dryRun {
			fmt.Printf("  %sWould register: %s → %s\n", prefix, email, roleARN)
		} else {
			db, err := openDB()
			if err != nil {
				fmt.Printf("  ⚠ Database not available: %v\n", err)
				fmt.Printf("  To register manually: qualify lab register-role --user %s --role-arn %s\n", email, roleARN)
			} else {
				defer db.Close()
				// Use context from background — the training service handles this
				// We call the openDB helper directly for the UPDATE
				_, updateErr := db.ExecContext(
					context.Background(),
					`UPDATE users SET metadata = jsonb_set(COALESCE(metadata,'{}'), '{role_arn}', to_jsonb($2::text)) WHERE id = $1`,
					email, roleARN)
				if updateErr != nil {
					fmt.Printf("  ⚠ Could not register role ARN: %v\n", updateErr)
				} else {
					fmt.Printf("  ✓ Role ARN registered: %s\n", shortARN(roleARN))
				}
			}
		}
	}
	fmt.Println()

	// ── Step 2: Required training ──────────────────────────────────────────
	fmt.Println("── Step 2: Required training")

	var activeFrameworks []string
	srePath := filepath.Join(attDir, "sre.yaml")
	if data, err := os.ReadFile(srePath); err == nil { // #nosec G304 — operator-controlled path
		var sre sreYAML
		if yaml.Unmarshal(data, &sre) == nil {
			for _, fw := range sre.Frameworks {
				activeFrameworks = append(activeFrameworks, fw.ID)
			}
		}
	}

	if len(activeFrameworks) == 0 {
		fmt.Println("  No active frameworks found in .attest/sre.yaml")
		fmt.Println("  Run 'attest init' and 'attest frameworks add' to configure frameworks.")
		fmt.Println()
	} else {
		fmt.Printf("  Active frameworks: %s\n\n", strings.Join(activeFrameworks, ", "))

		// Collect required modules
		required := make(map[string]bool)
		fwForModule := make(map[string][]string)
		for _, fw := range activeFrameworks {
			if modules, ok := frameworkModuleMap[fw]; ok {
				for _, m := range modules {
					required[m] = true
					fwForModule[m] = append(fwForModule[m], fw)
				}
			}
		}

		if len(required) > 0 {
			fmt.Printf("  Required training (%d modules):\n\n", len(required))
			order := []string{
				"security-awareness", "data-classification", "cui-fundamentals",
				"hipaa-privacy-security", "ferpa-basics", "itar-export-control",
				"nih-research-security", "countries-of-concern-awareness",
			}
			for _, m := range order {
				if !required[m] {
					continue
				}
				desc := moduleDescriptions[m]
				if desc == "" {
					desc = m
				}
				fmt.Printf("    ✗ %s\n", desc)
			}
			fmt.Println()
			fmt.Println("  Complete training (run in any order):")
			for _, m := range order {
				if required[m] {
					fmt.Printf("    qualify train start %s\n", m)
				}
			}
			fmt.Println()
			fmt.Println("  Each completion writes IAM tags automatically.")
			fmt.Println("  Cedar PDP grants access as each module is completed.")
		}
	}

	// ── Step 3: Lab assignment ─────────────────────────────────────────────
	fmt.Println("── Step 3: Lab assignment")
	if labID == "" {
		fmt.Println("  --lab-id not provided — skipping lab assignment")
		fmt.Printf("  To assign later: qualify lab setup --user %s --lab-id <id>\n", email)
	} else {
		if dryRun {
			fmt.Printf("  %sWould assign: %s → lab %s (admin-level: %s)\n", prefix, email, labID, adminLevel)
		} else {
			fmt.Printf("  qualify lab setup --user %s --lab-id %s --admin-level %s\n", email, labID, adminLevel)
			fmt.Println("  Run the above command to complete lab assignment.")
		}
	}
	fmt.Println()

	// ── Summary ────────────────────────────────────────────────────────────
	fmt.Println("── Onboarding summary")
	fmt.Printf("  User:        %s\n", email)
	if roleARN != "" {
		fmt.Printf("  Role:        %s\n", shortARN(roleARN))
	}
	if labID != "" {
		fmt.Printf("  Lab:         %s\n", labID)
	}
	fmt.Println()
	fmt.Println("  When all training is complete, Cedar PDP will automatically")
	fmt.Println("  grant access based on the attest:* IAM tags.")
	fmt.Println()
	fmt.Printf("  Check status: qualify train status --user %s\n", email)

	return nil
}

func shortARN(arn string) string {
	// Show just the role name portion for readability.
	if idx := strings.LastIndex(arn, "/"); idx >= 0 {
		return ".../" + arn[idx+1:]
	}
	return arn
}

