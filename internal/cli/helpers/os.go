// Package helpers provides common utility functions for CLI commands.
package helpers

import "runtime"

// DetectOS returns the detected OS name normalized for shellforge.
// Maps runtime.GOOS values to user-friendly names that match manifest OS values.
func DetectOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "Mac"
	case "linux":
		return "Linux"
	case "freebsd":
		return "FreeBSD"
	case "openbsd":
		return "OpenBSD"
	case "netbsd":
		return "NetBSD"
	case "windows":
		return "Windows"
	default:
		return runtime.GOOS
	}
}
