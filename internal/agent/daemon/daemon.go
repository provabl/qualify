// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Start launches the agent as a background daemon process
// Returns nil if the agent was started successfully
func Start(agentBinaryPath string) error {
	// Get data directory for log file
	dataDir, err := getDataDir()
	if err != nil {
		return fmt.Errorf("get data directory: %w", err)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	// Log file path
	logPath := filepath.Join(dataDir, "agent.log")

	// Open log file for appending
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer logFile.Close()

	// If no binary path provided, try to find it
	if agentBinaryPath == "" {
		agentBinaryPath, err = findAgentBinary()
		if err != nil {
			return fmt.Errorf("find agent binary: %w", err)
		}
	}

	// Create command to start agent
	cmd := exec.Command(agentBinaryPath)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Dir = dataDir

	// Platform-specific process management
	if runtime.GOOS == "windows" {
		// Windows: Start process in background
		cmd.SysProcAttr = getWindowsProcAttr()
	} else {
		// Unix: Fork and detach
		cmd.SysProcAttr = getUnixProcAttr()
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start agent process: %w", err)
	}

	// Don't wait for the process - let it run in background
	// On Unix, the process will be reparented to init
	return nil
}

// findAgentBinary locates the ark-agent binary
func findAgentBinary() (string, error) {
	// Check if ark-agent is in PATH
	path, err := exec.LookPath("ark-agent")
	if err == nil {
		return path, nil
	}

	// Check in the same directory as the current executable
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("get executable path: %w", err)
	}

	exeDir := filepath.Dir(exePath)
	agentPath := filepath.Join(exeDir, "ark-agent")
	if runtime.GOOS == "windows" {
		agentPath += ".exe"
	}

	if _, err := os.Stat(agentPath); err == nil {
		return agentPath, nil
	}

	return "", fmt.Errorf("ark-agent binary not found in PATH or %s", exeDir)
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
