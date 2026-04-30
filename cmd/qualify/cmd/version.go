// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildDate = "unknown"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  `Display detailed version information including commit SHA and build date.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Ark version %s\n", Version)
		fmt.Printf("Commit:     %s\n", CommitSHA)
		fmt.Printf("Built:      %s\n", BuildDate)
		fmt.Printf("Go version: %s\n", GoVersion())
		fmt.Printf("Platform:   %s\n", Platform())
	},
}
