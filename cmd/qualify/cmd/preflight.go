// SPDX-FileCopyrightText: 2026 Playground Logic LLC
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/provabl/qualify/internal/preflight"
)

func init() {
	rootCmd.AddCommand(preflightCmd)
	preflightCmd.Flags().String("region", "us-east-1", "AWS region")
}

var preflightCmd = &cobra.Command{
	Use:   "preflight",
	Short: "Verify the calling principal holds the IAM permissions qualify needs",
	Long: `Check that the calling AWS principal can perform qualify's AWS-touching action
(iam:TagRole, to write attest:* training/identity tags) via read-only
iam:SimulatePrincipalPolicy against the caller — it evaluates, it does not act.
A denied action prints a remediation and the command exits non-zero. See the
suite's docs/required-permissions.md.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		region, _ := cmd.Flags().GetString("region")
		results := preflight.CheckCallerPermissions(cmd.Context(), region)
		failures := 0
		for _, r := range results {
			if r.Status {
				fmt.Printf("  ✓ %s\n", r.Name)
				continue
			}
			failures++
			fmt.Printf("  ✗ %s: %s\n", r.Name, r.Detail)
			if r.Remediation != "" {
				fmt.Printf("      Remediation: %s\n", r.Remediation)
			}
		}
		fmt.Println()
		if failures > 0 {
			return fmt.Errorf("preflight failed: %d required permission(s) missing", failures)
		}
		fmt.Println("✓ All required permissions present")
		return nil
	},
}
