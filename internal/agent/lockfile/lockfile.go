// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package lockfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// LockFile represents a PID-based lock file
type LockFile struct {
	path string
	pid  int
}

// New creates a new lock file at the specified path
func New(path string) *LockFile {
	return &LockFile{
		path: path,
		pid:  os.Getpid(),
	}
}

// Acquire attempts to acquire the lock
// Returns an error if the lock is already held by another running process
func (l *LockFile) Acquire() error {
	// Check if lock file exists
	if _, err := os.Stat(l.path); err == nil {
		// Lock file exists - check if process is still running
		data, err := os.ReadFile(l.path)
		if err != nil {
			return fmt.Errorf("read existing lock file: %w", err)
		}

		pidStr := strings.TrimSpace(string(data))
		existingPID, err := strconv.Atoi(pidStr)
		if err != nil {
			// Invalid PID in lock file - consider it stale
			if err := os.Remove(l.path); err != nil {
				return fmt.Errorf("remove stale lock file: %w", err)
			}
		} else {
			// Check if process exists
			if processExists(existingPID) {
				return fmt.Errorf("agent already running with PID %d", existingPID)
			}
			// Process doesn't exist - remove stale lock
			if err := os.Remove(l.path); err != nil {
				return fmt.Errorf("remove stale lock file: %w", err)
			}
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(l.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create lock directory: %w", err)
	}

	// Write our PID to the lock file
	pidStr := fmt.Sprintf("%d\n", l.pid)
	if err := os.WriteFile(l.path, []byte(pidStr), 0600); err != nil {
		return fmt.Errorf("write lock file: %w", err)
	}

	return nil
}

// Release removes the lock file
func (l *LockFile) Release() error {
	// Read the lock file to verify it's our lock
	data, err := os.ReadFile(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			// Lock file doesn't exist - nothing to release
			return nil
		}
		return fmt.Errorf("read lock file: %w", err)
	}

	pidStr := strings.TrimSpace(string(data))
	lockPID, err := strconv.Atoi(pidStr)
	if err != nil {
		// Invalid PID - remove the file anyway
		return os.Remove(l.path)
	}

	// Only remove if it's our lock
	if lockPID != l.pid {
		return fmt.Errorf("lock file belongs to PID %d, not %d", lockPID, l.pid)
	}

	return os.Remove(l.path)
}

// IsLocked checks if a valid lock is currently held
func IsLocked(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return false
	}

	return processExists(pid)
}

// GetLockedPID returns the PID that holds the lock, or 0 if not locked
func GetLockedPID(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0
	}

	if processExists(pid) {
		return pid
	}

	return 0
}

// processExists checks if a process with the given PID exists
func processExists(pid int) bool {
	// Send signal 0 to check if process exists
	// This doesn't actually send a signal, just checks permissions
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, FindProcess always succeeds, so we need to send a signal to check
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
