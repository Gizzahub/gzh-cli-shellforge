package domain

import (
	"fmt"
	"path/filepath"
	"strings"
)

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
	r.pathMaps = map[string]map[string]string{
		"zsh": {
			"zshrc":    ".zshrc",
			"zprofile": ".zprofile",
			"zshenv":   ".zshenv",
			"zlogin":   ".zlogin",
			"zlogout":  ".zlogout",
			"profile":  ".profile",
		},
		"bash": {
			"bashrc":       ".bashrc",
			"bash_profile": ".bash_profile",
			"profile":      ".profile",
			"bash_login":   ".bash_login",
			"bash_logout":  ".bash_logout",
		},
		"fish": {
			"config": filepath.Join(".config", "fish", "config.fish"),
		},
	}
}

// Resolve returns the full file path for a target name.
// Returns an error if the target is not valid for the current shell type.
func (r *TargetResolver) Resolve(target string) (string, error) {
	target = strings.ToLower(target)

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

// GetValidTargets returns a list of valid target names for the current shell type.
func (r *TargetResolver) GetValidTargets() []string {
	shellMap, ok := r.pathMaps[r.shellType]
	if !ok {
		return nil
	}

	targets := make([]string, 0, len(shellMap))
	for target := range shellMap {
		targets = append(targets, target)
	}
	return targets
}

// IsValidTarget checks if a target is valid for the current shell type.
func (r *TargetResolver) IsValidTarget(target string) bool {
	target = strings.ToLower(target)
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
				mod.Name, target, r.shellType)
		}
	}
	return nil
}

// GetShellType returns the shell type.
func (r *TargetResolver) GetShellType() string {
	return r.shellType
}

// GetDefaultTarget returns the default target for the current shell type.
func (r *TargetResolver) GetDefaultTarget() string {
	switch r.shellType {
	case "zsh":
		return "zshrc"
	case "bash":
		return "bashrc"
	case "fish":
		return "config"
	default:
		return ""
	}
}

// GetTargetDescription returns a human-readable description of what a target file does.
func GetTargetDescription(target string) string {
	descriptions := map[string]string{
		"zshrc":        "Interactive shell configuration (aliases, functions, completions)",
		"zprofile":     "Login shell configuration (PATH, environment setup)",
		"zshenv":       "All shells (environment variables read by every zsh instance)",
		"zlogin":       "Login shell startup (after zshrc)",
		"zlogout":      "Login shell exit",
		"bashrc":       "Interactive non-login shell configuration",
		"bash_profile": "Login shell configuration (PATH, environment setup)",
		"profile":      "Login shell (sh-compatible, read by many shells)",
		"bash_login":   "Login shell startup (fallback if bash_profile missing)",
		"bash_logout":  "Login shell exit",
		"config":       "Fish shell configuration",
	}
	if desc, ok := descriptions[strings.ToLower(target)]; ok {
		return desc
	}
	return fmt.Sprintf("Target file: %s", target)
}
