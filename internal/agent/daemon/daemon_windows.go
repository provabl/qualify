// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package daemon

import (
	"syscall"
)

// getWindowsProcAttr returns process attributes for Windows
// This starts the process in the background without a console window
func getWindowsProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}

// getUnixProcAttr is a no-op on Windows
func getUnixProcAttr() *syscall.SysProcAttr {
	return nil
}
