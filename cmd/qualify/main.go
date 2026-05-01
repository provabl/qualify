// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/provabl/qualify/cmd/qualify/cmd"
)

var (
	version   = "0.1.2"
	commitSHA = "unknown"
	buildDate = "unknown"
)

func main() {
	// Set version information
	cmd.Version = version
	cmd.CommitSHA = commitSHA
	cmd.BuildDate = buildDate

	// Execute CLI
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
