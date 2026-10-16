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

func TestNormalizeKnownOS(t *testing.T) {
	tests := map[string]string{
		"mac":     "Mac",
		"LINUX":   "Linux",
		"FreeBSD": "FreeBSD",
		"openbsd": "OpenBSD",
		"NETBSD":  "NetBSD",
		"windows": "Windows",
	}
	for input, want := range tests {
		got, ok := NormalizeKnownOS(input)
		if !ok || got != want {
			t.Errorf("NormalizeKnownOS(%q) = %q, %t; want %q, true", input, got, ok, want)
		}
	}
	if _, ok := NormalizeKnownOS("Mca"); ok {
		t.Error("NormalizeKnownOS accepted an unknown OS")
	}
}
