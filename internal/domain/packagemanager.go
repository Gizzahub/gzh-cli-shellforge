package domain

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// PackageManager checks and installs packages for one manifest "packages:"
// key (e.g. "brew", "cask", "apt"). Implementations must be safe to call
// IsInstalled repeatedly and should treat "not found" as (false, nil) rather
// than an error.
type PackageManager interface {
	// Name is the manifest key this manager handles.
	Name() string
	// IsInstalled reports whether pkg is already installed.
	IsInstalled(pkg string) (bool, error)
	// Install installs pkg. PrepareService only calls Install for packages
	// IsInstalled reported as missing, so implementations do not need their
	// own idempotency check.
	Install(pkg string) error
}

// CommandRunner executes external commands. Kept as a small interface so
// package managers can be unit-tested without invoking brew/apt/dpkg.
type CommandRunner interface {
	// Run executes name with args and returns combined stdout+stderr output.
	Run(name string, args ...string) ([]byte, error)
}

// OsCommandRunner is the production CommandRunner backed by os/exec.
type OsCommandRunner struct{}

func (OsCommandRunner) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput() //nolint:gosec // args come from the manifest, not user input at runtime
}

// BrewFormulaManager manages Homebrew formulae (manifest key "brew").
type BrewFormulaManager struct{ Runner CommandRunner }

// NewBrewFormulaManager creates a BrewFormulaManager backed by the real brew CLI.
func NewBrewFormulaManager() *BrewFormulaManager {
	return &BrewFormulaManager{Runner: OsCommandRunner{}}
}

func (m *BrewFormulaManager) Name() string { return "brew" }

func (m *BrewFormulaManager) IsInstalled(pkg string) (bool, error) {
	_, err := m.Runner.Run("brew", "list", "--formula", "--versions", pkg)
	return err == nil, nil
}

func (m *BrewFormulaManager) Install(pkg string) error {
	out, err := m.Runner.Run("brew", "install", pkg)
	if err != nil {
		return fmt.Errorf("brew install %s: %w: %s", pkg, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// BrewCaskManager manages Homebrew casks (manifest key "cask").
type BrewCaskManager struct{ Runner CommandRunner }

// NewBrewCaskManager creates a BrewCaskManager backed by the real brew CLI.
func NewBrewCaskManager() *BrewCaskManager {
	return &BrewCaskManager{Runner: OsCommandRunner{}}
}

func (m *BrewCaskManager) Name() string { return "cask" }

func (m *BrewCaskManager) IsInstalled(pkg string) (bool, error) {
	_, err := m.Runner.Run("brew", "list", "--cask", "--versions", pkg)
	return err == nil, nil
}

func (m *BrewCaskManager) Install(pkg string) error {
	out, err := m.Runner.Run("brew", "install", "--cask", pkg)
	if err != nil {
		return fmt.Errorf("brew install --cask %s: %w: %s", pkg, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// AptManager manages Debian/Ubuntu packages (manifest key "apt").
type AptManager struct{ Runner CommandRunner }

// NewAptManager creates an AptManager backed by the real dpkg-query/apt-get CLIs.
func NewAptManager() *AptManager {
	return &AptManager{Runner: OsCommandRunner{}}
}

func (m *AptManager) Name() string { return "apt" }

func (m *AptManager) IsInstalled(pkg string) (bool, error) {
	out, err := m.Runner.Run("dpkg-query", "-W", "-f=${Status}", pkg)
	if err != nil {
		var exitErr interface{ ExitCode() int }
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil
		}

		output := strings.TrimSpace(string(out))
		if output == "" {
			return false, fmt.Errorf("dpkg-query status for %q: %w", pkg, err)
		}
		return false, fmt.Errorf("dpkg-query status for %q: %w: %s", pkg, err, output)
	}
	return strings.Contains(string(out), "install ok installed"), nil
}

func (m *AptManager) Install(pkg string) error {
	out, err := m.Runner.Run("apt-get", "install", "-y", pkg)
	if err != nil {
		return fmt.Errorf("apt-get install -y %s: %w: %s", pkg, err, strings.TrimSpace(string(out)))
	}
	return nil
}
