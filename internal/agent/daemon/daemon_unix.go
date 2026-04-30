// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

//go:build unix || darwin

package daemon

import (
	"syscall"
)

// getUnixProcAttr returns process attributes for Unix-like systems
// This makes the process detach from the parent and run as a daemon
func getUnixProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		// Create new process group (works on all Unix systems including macOS)
		Setpgid: true,
		Pgid:    0,
		// Note: Setsid is not used as it can cause "operation not permitted" on macOS
		// The process will still run in background via exec.Command
	}
}

// getWindowsProcAttr is a no-op on Unix
func getWindowsProcAttr() *syscall.SysProcAttr {
	return nil
}
