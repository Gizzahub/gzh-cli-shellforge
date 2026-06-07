package helpers

import (
	"runtime"
	"slices"
	"testing"
)

func TestDetectOS(t *testing.T) {
	result := DetectOS()
	if result == "" {
		t.Error("DetectOS() returned empty string")
	}

	expected := map[string]string{
		"darwin":  "Mac",
		"linux":   "Linux",
		"freebsd": "FreeBSD",
		"openbsd": "OpenBSD",
		"netbsd":  "NetBSD",
		"windows": "Windows",
	}

	if expectedOS, ok := expected[runtime.GOOS]; ok {
		if result != expectedOS {
			t.Errorf("DetectOS() = %q, want %q for GOOS=%q", result, expectedOS, runtime.GOOS)
		}
	} else {
		if result != runtime.GOOS {
			t.Errorf("DetectOS() = %q, want %q for unknown GOOS", result, runtime.GOOS)
		}
	}
}

func TestDetectOS_ReturnsKnownValue(t *testing.T) {
	result := DetectOS()
	knownValues := []string{"Mac", "Linux", "FreeBSD", "OpenBSD", "NetBSD", "Windows"}

	if slices.Contains(knownValues, result) {
		return
	}

	if result != runtime.GOOS {
		t.Errorf("DetectOS() = %q, expected known value or %q", result, runtime.GOOS)
	}
}
