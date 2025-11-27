package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Repository represents a git repository for backup operations
type Repository struct {
	dir string
}

// NewRepository creates a new git repository wrapper
func NewRepository(dir string) *Repository {
	return &Repository{dir: dir}
}

// IsGitInstalled checks if git is available on the system
func IsGitInstalled() bool {
	cmd := exec.Command("git", "--version")
	return cmd.Run() == nil
}

// IsInitialized checks if the directory is already a git repository
func (r *Repository) IsInitialized() bool {
	cmd := exec.Command("git", "-C", r.dir, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// Init initializes a new git repository
func (r *Repository) Init() error {
	if !IsGitInstalled() {
		return fmt.Errorf("git is not installed or not in PATH")
	}

	if r.IsInitialized() {
		return nil // Already initialized
	}

	cmd := exec.Command("git", "init", r.dir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git init failed: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// Add stages files for commit
func (r *Repository) Add(paths ...string) error {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	args := append([]string{"-C", r.dir, "add"}, paths...)
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// Commit creates a commit with the given message
func (r *Repository) Commit(message string) error {
	cmd := exec.Command("git", "-C", r.dir, "commit", "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if error is due to nothing to commit
		if strings.Contains(string(output), "nothing to commit") {
			return nil // Not an error
		}
		return fmt.Errorf("git commit failed: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// AddAndCommit is a convenience method that stages and commits changes
func (r *Repository) AddAndCommit(message string, paths ...string) error {
	if err := r.Add(paths...); err != nil {
		return err
	}
	return r.Commit(message)
}

// ConfigUser sets git user configuration for the repository
func (r *Repository) ConfigUser(name, email string) error {
	// Set user name
	cmd := exec.Command("git", "-C", r.dir, "config", "user.name", name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set user.name: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	// Set user email
	cmd = exec.Command("git", "-C", r.dir, "config", "user.email", email)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set user.email: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// GetStatus returns the current git status
func (r *Repository) GetStatus() (string, error) {
	cmd := exec.Command("git", "-C", r.dir, "status", "--short")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// HasChanges checks if there are uncommitted changes
func (r *Repository) HasChanges() (bool, error) {
	status, err := r.GetStatus()
	if err != nil {
		return false, err
	}
	return len(status) > 0, nil
}
