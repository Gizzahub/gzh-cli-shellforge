// Package helpers provides common utility functions for CLI commands.
package helpers

import (
	"fmt"
	"os"
	"path/filepath"
)

// ExpandHomePath expands ~ to the user's home directory.
// Returns the path unchanged if it doesn't start with ~.
func ExpandHomePath(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if len(path) == 1 {
		return home, nil
	}

	return filepath.Join(home, path[1:]), nil
}

// ResolveBackupDir resolves the backup directory path.
// If specified is empty, returns the default backup directory (~/.backup/shellforge).
// Otherwise, expands ~ in the specified path.
func ResolveBackupDir(specified string) (string, error) {
	if specified == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(home, ".backup", "shellforge"), nil
	}

	expanded, err := ExpandHomePath(specified)
	if err != nil {
		return "", fmt.Errorf("invalid backup directory: %w", err)
	}
	return expanded, nil
}

// DefaultBackupDir returns the default backup directory path.
func DefaultBackupDir() (string, error) {
	return ResolveBackupDir("")
}
