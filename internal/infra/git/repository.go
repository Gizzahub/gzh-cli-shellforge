// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

// Package git provides operations for interacting with Git repositories.
package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Repository represents a git repository for backup operations.
type Repository struct {
	dir string
}

// NewRepository creates a new git repository wrapper.
func NewRepository(dir string) *Repository {
	return &Repository{dir: dir}
}

// runGitCommand executes a git command and returns any error.
func (r *Repository) runGitCommand(subcommand string, args ...string) error {
	allArgs := append([]string{"-C", r.dir, subcommand}, args...)
	cmd := exec.CommandContext(context.Background(), "git", allArgs...) // #nosec G204 -- fixed "git" binary, no shell; subcommand is a literal from this package and r.dir/args are separate argv elements
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s failed: %w (%s)", subcommand, err, strings.TrimSpace(string(output)))
	}
	return nil
}

// runGitCommandWithOutput executes a git command and returns its output.
func (r *Repository) runGitCommandWithOutput(subcommand string, args ...string) (string, error) {
	allArgs := append([]string{"-C", r.dir, subcommand}, args...)
	cmd := exec.CommandContext(context.Background(), "git", allArgs...) // #nosec G204 -- fixed "git" binary, no shell; subcommand is a literal from this package and r.dir/args are separate argv elements
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w (%s)", subcommand, err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

// IsGitInstalled checks if git is available on the system.
func IsGitInstalled() bool {
	cmd := exec.CommandContext(context.Background(), "git", "--version")
	return cmd.Run() == nil
}

// IsInitialized checks if the directory is already a git repository.
func (r *Repository) IsInitialized() bool {
	cmd := exec.CommandContext(context.Background(), "git", "-C", r.dir, "rev-parse", "--git-dir") // #nosec G204 -- fixed git argv; r.dir is the configured backup directory passed as one argument, no shell
	return cmd.Run() == nil
}

// Init initializes a new git repository.
func (r *Repository) Init() error {
	if !IsGitInstalled() {
		return fmt.Errorf("git is not installed or not in PATH")
	}

	if r.IsInitialized() {
		return nil // Already initialized
	}

	cmd := exec.CommandContext(context.Background(), "git", "init", "--", r.dir) // #nosec G204 -- fixed git argv, no shell; "--" keeps a backup directory such as "--bare" from being read as an option
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git init failed: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// Add stages files for commit.
func (r *Repository) Add(paths ...string) error {
	if len(paths) == 0 {
		paths = []string{"."}
	}
	// "--" keeps a path that starts with "-" from being read as a git option.
	return r.runGitCommand("add", append([]string{"--"}, paths...)...)
}

// Commit creates a commit with the given message.
func (r *Repository) Commit(message string) error {
	cmd := exec.CommandContext(context.Background(), "git", "-C", r.dir, "commit", "-m", message) // #nosec G204 -- fixed git argv; the message is the value of -m, so it cannot become an option, and no shell is used
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)
		// Check if error is due to nothing to commit (in various languages)
		if strings.Contains(outputStr, "nothing to commit") ||
			strings.Contains(outputStr, "커밋할 사항 없음") ||
			strings.Contains(outputStr, "nothing added to commit") {
			return nil // Not an error
		}
		return fmt.Errorf("git commit failed: %w (%s)", err, strings.TrimSpace(outputStr))
	}

	return nil
}

// AddAndCommit is a convenience method that stages and commits changes.
func (r *Repository) AddAndCommit(message string, paths ...string) error {
	if err := r.Add(paths...); err != nil {
		return err
	}
	return r.Commit(message)
}

// ConfigUser sets git user configuration for the repository.
func (r *Repository) ConfigUser(name, email string) error {
	if err := r.runGitCommand("config", "user.name", name); err != nil {
		return fmt.Errorf("failed to set user.name: %w", err)
	}
	if err := r.runGitCommand("config", "user.email", email); err != nil {
		return fmt.Errorf("failed to set user.email: %w", err)
	}
	return nil
}

// GetStatus returns the current git status.
func (r *Repository) GetStatus() (string, error) {
	return r.runGitCommandWithOutput("status", "--short")
}

// HasChanges checks if there are uncommitted changes.
func (r *Repository) HasChanges() (bool, error) {
	status, err := r.GetStatus()
	if err != nil {
		return false, err
	}
	return status != "", nil
}
