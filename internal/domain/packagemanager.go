// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package domain

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

const (
	aptPackageManagerID  = "apt"
	brewCommand          = "brew"
	caskPackageManagerID = "cask"
)

// PackageManager checks and installs packages for one manifest "packages:"
// key (e.g. "brew", "cask", "apt"). Implementations must be safe to call
// IsInstalled repeatedly and should treat "not found" as (false, nil) rather
// than an error.
type PackageManager interface {
	// Name is the manifest key this manager handles.
	Name() string
	// IsInstalled reports whether pkg is already installed.
	IsInstalled(ctx context.Context, pkg string) (bool, error)
	// Install installs pkg. PrepareService only calls Install for packages
	// IsInstalled reported as missing, so implementations do not need their
	// own idempotency check.
	Install(ctx context.Context, pkg string) error
}

// CommandRunner executes external commands. Kept as a small interface so
// package managers can be unit-tested without invoking brew/apt/dpkg.
type CommandRunner interface {
	// Run executes name with args and returns combined stdout+stderr output.
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// StreamingCommandRunner can route a command's stdout and stderr to a writer
// while the command is still running. Package managers use it for installs so
// a user can observe a successful long-running installation as it progresses.
type StreamingCommandRunner interface {
	CommandRunner
	RunWithOutput(ctx context.Context, output io.Writer, name string, args ...string) error
}

// OsCommandRunner is the production CommandRunner backed by os/exec.
type OsCommandRunner struct{}

// Run executes name with args through the operating system.
func (OsCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput() //nolint:gosec // args come from the manifest, not user input at runtime
	if ctxErr := ctx.Err(); ctxErr != nil {
		return out, ctxErr
	}
	return out, err
}

// RunWithOutput executes name with its stdout and stderr streamed to output.
func (OsCommandRunner) RunWithOutput(ctx context.Context, output io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // args come from the manifest, not user input at runtime
	cmd.Stdout = output
	cmd.Stderr = output
	err := cmd.Run()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	return err
}

// BrewFormulaManager manages Homebrew formulae (manifest key "brew").
type BrewFormulaManager struct {
	Runner CommandRunner
	Output io.Writer
}

// NewBrewFormulaManager creates a BrewFormulaManager backed by the real brew CLI.
func NewBrewFormulaManager() *BrewFormulaManager {
	return &BrewFormulaManager{Runner: OsCommandRunner{}}
}

// Name returns the manifest package key handled by this manager.
func (m *BrewFormulaManager) Name() string { return brewCommand }

// IsInstalled reports whether the Homebrew formula is installed.
func (m *BrewFormulaManager) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	_, err := m.Runner.Run(ctx, brewCommand, "list", "--formula", "--versions", pkg)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return false, ctxErr
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false, err
	}
	return err == nil, nil
}

// Install installs the Homebrew formula.
func (m *BrewFormulaManager) Install(ctx context.Context, pkg string) error {
	err := runInstall(ctx, m.Runner, m.Output, brewCommand, "install", pkg)
	if err != nil {
		return fmt.Errorf("brew install %s: %w", pkg, err)
	}
	return nil
}

// BrewCaskManager manages Homebrew casks (manifest key "cask").
type BrewCaskManager struct {
	Runner CommandRunner
	Output io.Writer
}

// NewBrewCaskManager creates a BrewCaskManager backed by the real brew CLI.
func NewBrewCaskManager() *BrewCaskManager {
	return &BrewCaskManager{Runner: OsCommandRunner{}}
}

// Name returns the manifest package key handled by this manager.
func (m *BrewCaskManager) Name() string { return caskPackageManagerID }

// IsInstalled reports whether the Homebrew cask is installed.
func (m *BrewCaskManager) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	_, err := m.Runner.Run(ctx, brewCommand, "list", "--cask", "--versions", pkg)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return false, ctxErr
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false, err
	}
	return err == nil, nil
}

// Install installs the Homebrew cask.
func (m *BrewCaskManager) Install(ctx context.Context, pkg string) error {
	err := runInstall(ctx, m.Runner, m.Output, brewCommand, "install", "--cask", pkg)
	if err != nil {
		return fmt.Errorf("brew install --cask %s: %w", pkg, err)
	}
	return nil
}

// AptManager manages Debian/Ubuntu packages (manifest key "apt").
type AptManager struct {
	Runner CommandRunner
	Output io.Writer
}

// NewAptManager creates an AptManager backed by the real dpkg-query/apt-get CLIs.
func NewAptManager() *AptManager {
	return &AptManager{Runner: OsCommandRunner{}}
}

// Name returns the manifest package key handled by this manager.
func (m *AptManager) Name() string { return aptPackageManagerID }

// IsInstalled reports whether the Debian or Ubuntu package is installed.
func (m *AptManager) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	out, err := m.Runner.Run(ctx, "dpkg-query", "-W", "-f=${Status}", pkg)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return false, ctxErr
	}
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

// Install installs the Debian or Ubuntu package.
func (m *AptManager) Install(ctx context.Context, pkg string) error {
	err := runInstall(ctx, m.Runner, m.Output, "apt-get", "install", "-y", pkg)
	if err != nil {
		return fmt.Errorf("apt-get install -y %s: %w", pkg, err)
	}
	return nil
}

func runInstall(ctx context.Context, runner CommandRunner, output io.Writer, name string, args ...string) error {
	if output != nil {
		if streaming, ok := runner.(StreamingCommandRunner); ok {
			return streaming.RunWithOutput(ctx, output, name, args...)
		}
	}

	out, err := runner.Run(ctx, name, args...)
	if output != nil && len(out) > 0 {
		if _, writeErr := output.Write(out); writeErr != nil && err == nil {
			return writeErr
		}
	}
	if err != nil {
		if message := strings.TrimSpace(string(out)); message != "" {
			return fmt.Errorf("%w: %s", err, message)
		}
		return err
	}
	return nil
}
