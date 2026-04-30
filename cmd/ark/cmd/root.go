// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	verbose    bool
	configFile string
)

var rootCmd = &cobra.Command{
	Use:   "qualify",
	Short: "qualify - Researcher training & access gating for academic institutions",
	Long: `qualify provides integrated researcher training and security tooling for research institutions.

The training-as-tool approach embeds security education directly into AWS workflows,
ensuring researchers can use cloud resources safely and compliantly from day one.

qualify consists of three components:
  • CLI      - Command-line interface for scripting and automation
  • Agent    - Local service (localhost:8737) that brokers AWS credentials
  • Backend  - Institutional backend for training, policies, and audit

Getting Started:
  1. Start the agent:    qualify agent start
  2. Configure profile:  qualify config set profile default
  3. Lab setup:          qualify lab setup --lab-id <id>

For more information: https://github.com/provabl/qualify`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", Version, CommitSHA, BuildDate),
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Config file (default: ~/.qualify/config.yml)")

	// Disable default completion command (we'll add our own)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// GoVersion returns the Go version
func GoVersion() string {
	return runtime.Version()
}

// Platform returns the OS and architecture
func Platform() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}

// ExitWithError prints an error and exits
func ExitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
