// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

// Package helpers provides common utility functions for CLI commands.
package helpers

import (
	"runtime"
	"strings"
)

const (
	macOSName   = "Mac"
	linuxOSName = "Linux"
	freeBSDName = "FreeBSD"
	openBSDName = "OpenBSD"
	netBSDName  = "NetBSD"
	windowsName = "Windows"
)

var knownOS = map[string]string{
	"mac":     macOSName,
	"linux":   linuxOSName,
	"freebsd": freeBSDName,
	"openbsd": openBSDName,
	"netbsd":  netBSDName,
	"windows": windowsName,
}

// NormalizeKnownOS returns the canonical shellforge OS name for a supported
// target, accepting its spelling case-insensitively.
func NormalizeKnownOS(value string) (string, bool) {
	canonical, ok := knownOS[strings.ToLower(value)]
	return canonical, ok
}

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
