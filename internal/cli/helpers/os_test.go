package helpers

import (
	"runtime"
	"testing"
)

func TestDetectOS(t *testing.T) {
	// Test that DetectOS returns a non-empty string
	result := DetectOS()
	if result == "" {
		t.Error("DetectOS() returned empty string")
	}

	// Verify against runtime.GOOS mapping
	expected := map[string]string{
		"darwin":  "Mac",
		"linux":   "Linux",
		"freebsd": "BSD",
		"openbsd": "BSD",
		"netbsd":  "BSD",
		"windows": "Windows",
	}

	if expectedOS, ok := expected[runtime.GOOS]; ok {
		if result != expectedOS {
			t.Errorf("DetectOS() = %q, want %q for GOOS=%q", result, expectedOS, runtime.GOOS)
		}
	} else {
		// For unknown OS, should return runtime.GOOS as-is
		if result != runtime.GOOS {
			t.Errorf("DetectOS() = %q, want %q for unknown GOOS", result, runtime.GOOS)
		}
	}
}

func TestDetectOS_ReturnsKnownValue(t *testing.T) {
	result := DetectOS()
	knownValues := []string{"Mac", "Linux", "BSD", "Windows"}

	// Either it's a known value or it's the raw GOOS
	for _, known := range knownValues {
		if result == known {
			return // Test passes
		}
	}

	// If not a known value, should be the raw GOOS
	if result != runtime.GOOS {
		t.Errorf("DetectOS() = %q, expected known value or %q", result, runtime.GOOS)
	}
}
