// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	rootCmd.AddCommand(credentialsCmd)
	credentialsCmd.AddCommand(credentialsSetCmd)
	credentialsCmd.AddCommand(credentialsListCmd)
	credentialsCmd.AddCommand(credentialsDeleteCmd)

	// Flags for set command
	credentialsSetCmd.Flags().String("access-key-id", "", "AWS access key ID")
	credentialsSetCmd.Flags().String("secret-access-key", "", "AWS secret access key")
	credentialsSetCmd.Flags().String("session-token", "", "AWS session token (for temporary credentials)")
	credentialsSetCmd.Flags().String("region", "us-east-1", "Default AWS region")
}

var credentialsCmd = &cobra.Command{
	Use:   "credentials",
	Short: "Manage AWS credentials",
	Long: `Store and manage AWS credentials for use with Ark.

Credentials are stored securely in the local agent database and never leave your machine.`,
}

var credentialsSetCmd = &cobra.Command{
	Use:   "set <profile-name>",
	Short: "Store AWS credentials for a profile",
	Long: `Store AWS credentials for a named profile.

The credentials are stored in the agent's local database at ~/.ark/agent.db.
They are used automatically when you run AWS operations.

Examples:
  # Interactive prompts for credentials
  ark credentials set default

  # Provide credentials via flags
  ark credentials set prod --access-key-id AKIA... --secret-access-key ...

Note: Credentials are stored locally and not encrypted. In production
environments, consider using IAM roles or AWS SSO instead.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profile := args[0]

		// Ensure agent is running
		if err := EnsureAgentRunning(); err != nil {
			ExitWithError(fmt.Errorf("agent not available: %w", err))
		}

		// Get credentials from flags or prompt
		accessKeyID, _ := cmd.Flags().GetString("access-key-id")
		secretAccessKey, _ := cmd.Flags().GetString("secret-access-key")
		sessionToken, _ := cmd.Flags().GetString("session-token")
		region, _ := cmd.Flags().GetString("region")

		// Prompt for missing credentials
		if accessKeyID == "" {
			fmt.Print("AWS Access Key ID: ")
			fmt.Scanln(&accessKeyID)
		}

		if secretAccessKey == "" {
			fmt.Print("AWS Secret Access Key: ")
			secretBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				ExitWithError(fmt.Errorf("failed to read password: %w", err))
			}
			fmt.Println() // New line after password input
			secretAccessKey = string(secretBytes)
		}

		// Validate required fields
		if accessKeyID == "" || secretAccessKey == "" {
			ExitWithError(fmt.Errorf("access-key-id and secret-access-key are required"))
		}

		// Create request payload
		payload := map[string]interface{}{
			"profile":           profile,
			"access_key_id":     accessKeyID,
			"secret_access_key": secretAccessKey,
			"region":            region,
		}
		if sessionToken != "" {
			payload["session_token"] = sessionToken
		}

		// Send to agent
		jsonData, err := json.Marshal(payload)
		if err != nil {
			ExitWithError(fmt.Errorf("marshal credentials: %w", err))
		}

		resp, err := http.Post(
			"http://127.0.0.1:8737/api/credentials",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			ExitWithError(fmt.Errorf("send to agent: %w", err))
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errResp struct {
				Error string `json:"error"`
			}
			json.NewDecoder(resp.Body).Decode(&errResp)
			ExitWithError(fmt.Errorf("agent error: %s", errResp.Error))
		}

		fmt.Printf("✓ Credentials stored for profile '%s'\n", profile)
		fmt.Println()
		fmt.Println("Note: Credentials are stored locally and not encrypted.")
		fmt.Println("In production, consider using IAM roles or AWS SSO.")
	},
}

var credentialsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored credential profiles",
	Long:  `Display all credential profiles stored in the agent.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure agent is running
		if err := EnsureAgentRunning(); err != nil {
			ExitWithError(fmt.Errorf("agent not available: %w", err))
		}

		// Query agent
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("http://127.0.0.1:8737/api/credentials")
		if err != nil {
			ExitWithError(fmt.Errorf("query agent: %w", err))
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ExitWithError(fmt.Errorf("agent returned status %d", resp.StatusCode))
		}

		var profiles []struct {
			Profile string `json:"profile"`
			Region  string `json:"region"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
			ExitWithError(fmt.Errorf("decode response: %w", err))
		}

		if len(profiles) == 0 {
			fmt.Println("No credentials stored.")
			fmt.Println()
			fmt.Println("Add credentials with: ark credentials set <profile>")
			return
		}

		fmt.Println("Stored credential profiles:")
		fmt.Println()
		for _, p := range profiles {
			fmt.Printf("  %s  (region: %s)\n", p.Profile, p.Region)
		}
	},
}

var credentialsDeleteCmd = &cobra.Command{
	Use:   "delete <profile-name>",
	Short: "Delete stored credentials",
	Long:  `Remove credentials for a profile from the agent.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profile := args[0]

		// Ensure agent is running
		if err := EnsureAgentRunning(); err != nil {
			ExitWithError(fmt.Errorf("agent not available: %w", err))
		}

		// Send delete request to agent
		url := fmt.Sprintf("http://127.0.0.1:8737/api/credentials/%s", profile)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			ExitWithError(fmt.Errorf("create request: %w", err))
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			ExitWithError(fmt.Errorf("send to agent: %w", err))
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			fmt.Printf("Profile '%s' not found\n", profile)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			ExitWithError(fmt.Errorf("agent returned status %d", resp.StatusCode))
		}

		fmt.Printf("✓ Deleted credentials for profile '%s'\n", profile)
	},
}
