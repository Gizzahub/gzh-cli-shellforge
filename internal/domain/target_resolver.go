// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package domain

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	shellZsh  = "zsh"
	shellBash = "bash"
	shellFish = "fish"

	targetZshrc    = "zshrc"
	targetZprofile = "zprofile"
	targetZshenv   = "zshenv"
	targetZlogin   = "zlogin"
	targetZlogout  = "zlogout"
	targetProfile  = "profile"

	targetBashrc      = "bashrc"
	targetBashProfile = "bash_profile"
	targetBashLogin   = "bash_login"
	targetBashLogout  = "bash_logout"

	targetConfig = "config"
	targetConfD  = "conf.d"

	targetEtcProfile = "etc-profile"
	targetEtcZshrc   = "etc-zshrc"
	targetEtcZshenv  = "etc-zshenv"
)

// systemTargetPaths maps the documented system-wide target names to their absolute paths.
// These targets bypass home-relative resolution and require elevated privileges to write.
// Arbitrary absolute targets are intentionally excluded to limit blast radius.
var systemTargetPaths = map[string]string{
	targetEtcProfile: "/etc/profile",
	targetEtcZshrc:   "/etc/zshrc",
	targetEtcZshenv:  "/etc/zsh/zshenv",
}

// IsSystemTarget reports whether name refers to a system-wide configuration file.
// System targets resolve to absolute paths and require privileged writes.
func IsSystemTarget(name string) bool {
	_, ok := systemTargetPaths[strings.ToLower(name)]
	return ok
}

// TargetResolver resolves target names to actual file paths based on shell type.
type TargetResolver struct {
	shellType string
	homeDir   string
	pathMaps  map[string]map[string]string // shell -> target -> path
}

// NewTargetResolver creates a new resolver for the given shell type and home directory.
func NewTargetResolver(shellType, homeDir string) *TargetResolver {
	r := &TargetResolver{
		shellType: strings.ToLower(shellType),
		homeDir:   homeDir,
	}
	r.initPathMaps()
	return r
}

// initPathMaps initializes the target to path mappings for each shell type.
func (r *TargetResolver) initPathMaps() {
	// XDG_CONFIG_HOME support for Fish shell (XDG Base Directory Specification)
	// Default: ~/.config, can be overridden by $XDG_CONFIG_HOME
	fishConfigBase := r.resolveFishConfigBase()

	r.pathMaps = map[string]map[string]string{
		shellZsh: {
			targetZshrc:    ".zshrc",
			targetZprofile: ".zprofile",
			targetZshenv:   ".zshenv",
			targetZlogin:   ".zlogin",
			targetZlogout:  ".zlogout",
			targetProfile:  ".profile",
		},
		shellBash: {
			targetBashrc:      ".bashrc",
			targetBashProfile: ".bash_profile",
			targetProfile:     ".profile",
			targetBashLogin:   ".bash_login",
			targetBashLogout:  ".bash_logout",
		},
		shellFish: {
			targetConfig: filepath.Join(fishConfigBase, shellFish, "config.fish"),
			targetConfD:  filepath.Join(fishConfigBase, shellFish, targetConfD),
		},
	}
}

// resolveFishConfigBase returns the base config directory for Fish shell.
// Respects XDG_CONFIG_HOME environment variable (XDG Base Directory Specification).
func (r *TargetResolver) resolveFishConfigBase() string {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		return ".config"
	}

	// Handle absolute paths: convert to relative path from homeDir if possible
	if filepath.IsAbs(xdgConfigHome) {
		if r.homeDir != "" && strings.HasPrefix(xdgConfigHome, r.homeDir) {
			relPath, err := filepath.Rel(r.homeDir, xdgConfigHome)
			if err == nil {
				return relPath
			}
		}
		// If absolute path is outside homeDir, use default
		// (current architecture uses homeDir-relative paths)
		return ".config"
	}

	// Relative path: use as-is
	return xdgConfigHome
}

// Resolve returns the full file path for a target name.
// System targets (etc-profile, etc-zshrc, etc-zshenv) return absolute paths directly.
// Returns an error if the target is not valid for the current shell type.
func (r *TargetResolver) Resolve(target string) (string, error) {
	target = strings.ToLower(target)

	if absPath, ok := systemTargetPaths[target]; ok {
		return absPath, nil
	}

	shellMap, ok := r.pathMaps[r.shellType]
	if !ok {
		return "", NewValidationError("unsupported shell type: %s", r.shellType)
	}

	relPath, ok := shellMap[target]
	if !ok {
		return "", NewValidationError("invalid target '%s' for shell type '%s'", target, r.shellType)
	}

	return filepath.Join(r.homeDir, relPath), nil
}

// GetValidTargets returns a list of valid target names for the current shell type,
// including the shared system-wide targets (etc-profile, etc-zshrc, etc-zshenv).
func (r *TargetResolver) GetValidTargets() []string {
	shellMap, ok := r.pathMaps[r.shellType]
	if !ok {
		return nil
	}

	targets := make([]string, 0, len(shellMap)+len(systemTargetPaths))
	for target := range shellMap {
		targets = append(targets, target)
	}
	for target := range systemTargetPaths {
		targets = append(targets, target)
	}
	return targets
}

// IsValidTarget checks if a target is valid for the current shell type or is a system target.
func (r *TargetResolver) IsValidTarget(target string) bool {
	target = strings.ToLower(target)
	if _, ok := systemTargetPaths[target]; ok {
		return true
	}
	shellMap, ok := r.pathMaps[r.shellType]
	if !ok {
		return false
	}
	_, ok = shellMap[target]
	return ok
}

// ValidateTargets checks if all modules have valid targets for the current shell type.
func (r *TargetResolver) ValidateTargets(modules []Module) error {
	for _, mod := range modules {
		target := mod.GetTarget()
		if !r.IsValidTarget(target) {
			return NewValidationError(
				"module '%s' has invalid target '%s' for shell type '%s'",
				mod.Name, target, r.shellType,
			)
		}
	}
	return nil
}

// GetShellType returns the shell type.
func (r *TargetResolver) GetShellType() string {
	return r.shellType
}

// IsDirectoryTarget returns true if the target is a directory (e.g., conf.d).
func (r *TargetResolver) IsDirectoryTarget(target string) bool {
	target = strings.ToLower(target)
	// Directory targets that generate multiple files
	directoryTargets := map[string]bool{
		targetConfD: true,
	}
	return directoryTargets[target]
}

// GetDefaultTarget returns the default target for the current shell type.
func (r *TargetResolver) GetDefaultTarget() string {
	switch r.shellType {
	case shellZsh:
		return targetZshrc
	case shellBash:
		return targetBashrc
	case shellFish:
		return targetConfig
	default:
		return ""
	}
}

// GetRelativePath returns the deployment path for a target.
// For user targets it is home-relative (e.g. ".zshrc").
// For system targets it is the absolute path (e.g. "/etc/profile"); callers must
// use filepath.IsAbs to distinguish and must NOT join the result with HomeDir.
func (r *TargetResolver) GetRelativePath(target string) (string, error) {
	target = strings.ToLower(target)

	if absPath, ok := systemTargetPaths[target]; ok {
		return absPath, nil
	}

	shellMap, ok := r.pathMaps[r.shellType]
	if !ok {
		return "", NewValidationError("unsupported shell type: %s", r.shellType)
	}

	relPath, ok := shellMap[target]
	if !ok {
		return "", NewValidationError("invalid target '%s' for shell type '%s'", target, r.shellType)
	}

	return relPath, nil
}

// GetTargetDescription returns a human-readable description of what a target file does.
func GetTargetDescription(target string) string {
	descriptions := map[string]string{
		targetZshrc:       "Interactive shell configuration (aliases, functions, completions)",
		targetZprofile:    "Login shell configuration (PATH, environment setup)",
		targetZshenv:      "All shells (environment variables read by every zsh instance)",
		targetZlogin:      "Login shell startup (after zshrc)",
		targetZlogout:     "Login shell exit",
		targetBashrc:      "Interactive non-login shell configuration",
		targetBashProfile: "Login shell configuration (PATH, environment setup)",
		targetProfile:     "Login shell (sh-compatible, read by many shells)",
		targetBashLogin:   "Login shell startup (fallback if bash_profile missing)",
		targetBashLogout:  "Login shell exit",
		targetConfig:      "Fish shell configuration",
		targetConfD:       "Fish modular configs (auto-sourced .fish files in conf.d/)",
		// System-wide targets (require elevated privileges)
		targetEtcProfile: "System-wide login shell config (/etc/profile) — affects all users, requires sudo",
		targetEtcZshrc:   "System-wide zsh interactive config (/etc/zshrc) — affects all users, requires sudo",
		targetEtcZshenv:  "System-wide zsh environment (/etc/zsh/zshenv) — affects all users, requires sudo",
	}
	if desc, ok := descriptions[strings.ToLower(target)]; ok {
		return desc
	}
	return fmt.Sprintf("Target file: %s", target)
}
