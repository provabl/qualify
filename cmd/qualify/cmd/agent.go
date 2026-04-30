// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/provabl/qualify/internal/agent/daemon"
	"github.com/provabl/qualify/internal/agent/lockfile"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(agentStartCmd)
	agentCmd.AddCommand(agentStopCmd)
	agentCmd.AddCommand(agentStatusCmd)
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage the Ark agent",
	Long: `Manage the Ark agent service that runs locally and brokers AWS credentials.

The agent runs on localhost:8737 and must be running for Ark CLI commands to work.`,
}

var agentStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Ark agent",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if agent is already running
		if isAgentRunning() {
			fmt.Println("✓ Agent is already running")
			return
		}

		fmt.Println("Starting Ark agent...")

		// Start agent in background
		if err := daemon.Start(""); err != nil {
			ExitWithError(fmt.Errorf("start agent: %w", err))
		}

		// Wait for agent to become healthy (with timeout)
		fmt.Print("Waiting for agent to start")
		if err := waitForAgent(10 * time.Second); err != nil {
			fmt.Println(" ✗")
			ExitWithError(fmt.Errorf("agent failed to start: %w", err))
		}

		fmt.Println(" ✓")
		fmt.Println("Agent started successfully")

		// Show log location
		dataDir, _ := getDataDir()
		if dataDir != "" {
			logPath := filepath.Join(dataDir, "agent.log")
			fmt.Printf("Logs: %s\n", logPath)
		}
	},
}

var agentStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Ark agent",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if agent is running
		if !isAgentRunning() {
			fmt.Println("Agent is not running")
			return
		}

		// Get lock file path
		dataDir, err := getDataDir()
		if err != nil {
			ExitWithError(fmt.Errorf("get data directory: %w", err))
		}

		lockPath := filepath.Join(dataDir, "agent.lock")

		// Get PID from lock file
		pid := lockfile.GetLockedPID(lockPath)
		if pid == 0 {
			fmt.Println("Agent is running but lock file not found")
			fmt.Println("Try manually stopping the agent process")
			os.Exit(1)
		}

		fmt.Printf("Stopping agent (PID %d)...\n", pid)

		// Find process and send SIGTERM
		process, err := os.FindProcess(pid)
		if err != nil {
			ExitWithError(fmt.Errorf("find process: %w", err))
		}

		// Send termination signal
		if err := process.Signal(os.Interrupt); err != nil {
			ExitWithError(fmt.Errorf("send signal: %w", err))
		}

		// Wait for agent to stop (with timeout)
		fmt.Print("Waiting for agent to stop")
		timeout := time.After(10 * time.Second)
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				fmt.Println(" ✗")
				fmt.Println("Agent did not stop gracefully, may need to be killed manually")
				os.Exit(1)
			case <-ticker.C:
				if !isAgentRunning() {
					fmt.Println(" ✓")
					fmt.Println("Agent stopped successfully")
					return
				}
				fmt.Print(".")
			}
		}
	},
}

var agentStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check agent status",
	Run: func(cmd *cobra.Command, args []string) {
		if isAgentRunning() {
			fmt.Println("✓ Agent is running")

			// Try to get version
			version, err := getAgentVersion()
			if err == nil {
				fmt.Printf("  Version: %s\n", version)
			}
			os.Exit(0)
		} else {
			fmt.Println("✗ Agent is not running")
			fmt.Println("")
			fmt.Println("Start the agent with: ark agent start")
			os.Exit(1)
		}
	},
}

// isAgentRunning checks if the agent is responding to health checks
func isAgentRunning() bool {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get("http://127.0.0.1:8737/api/system/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// getAgentVersion gets the agent version
func getAgentVersion() (string, error) {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get("http://127.0.0.1:8737/api/system/version")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var versionResp struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&versionResp); err != nil {
		return "", err
	}

	return versionResp.Version, nil
}

// waitForAgent waits for the agent to become healthy with a timeout
func waitForAgent(timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			return fmt.Errorf("timeout waiting for agent")
		case <-ticker.C:
			if isAgentRunning() {
				return nil
			}
			fmt.Print(".")
		}
	}
}

// getDataDir returns the agent data directory
func getDataDir() (string, error) {
	// Check environment variable first
	if dir := os.Getenv("ARK_AGENT_DATA"); dir != "" {
		return dir, nil
	}

	// Use ~/.qualify for data storage
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	return filepath.Join(home, ".qualify"), nil
}

// EnsureAgentRunning ensures the agent is running, starting it if necessary
// This is called automatically by commands that require the agent
// Set ARK_NO_AUTO_START=1 to disable auto-start behavior
func EnsureAgentRunning() error {
	// Check if auto-start is disabled
	if os.Getenv("ARK_NO_AUTO_START") != "" {
		if !isAgentRunning() {
			return fmt.Errorf("agent is not running (auto-start disabled)")
		}
		return nil
	}

	// Agent already running - nothing to do
	if isAgentRunning() {
		return nil
	}

	// Start agent in background
	if err := daemon.Start(""); err != nil {
		return fmt.Errorf("auto-start agent: %w", err)
	}

	// Wait for agent to become healthy (shorter timeout for auto-start)
	if err := waitForAgent(5 * time.Second); err != nil {
		return fmt.Errorf("agent failed to start: %w", err)
	}

	return nil
}
