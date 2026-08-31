// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

// Package helpers provides common utility functions for CLI commands.
package helpers

import "runtime"

const (
	macOSName   = "Mac"
	linuxOSName = "Linux"
	freeBSDName = "FreeBSD"
	openBSDName = "OpenBSD"
	netBSDName  = "NetBSD"
	windowsName = "Windows"
)

// DetectOS returns the detected OS name normalized for shellforge.
// Maps runtime.GOOS values to user-friendly names that match manifest OS values.
func DetectOS() string {
	switch runtime.GOOS {
	case "darwin":
		return macOSName
	case "linux":
		return linuxOSName
	case "freebsd":
		return freeBSDName
	case "openbsd":
		return openBSDName
	case "netbsd":
		return netBSDName
	case "windows":
		return windowsName
	default:
		return runtime.GOOS
	}
}
