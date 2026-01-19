package domain

import "strings"

// Module represents a shell module with dependencies and OS filtering.
type Module struct {
	Name        string   `yaml:"name"`
	File        string   `yaml:"file"`
	Requires    []string `yaml:"requires,omitempty"`
	OS          []string `yaml:"os,omitempty"`
	Description string   `yaml:"description,omitempty"`

	// Target specifies the destination RC file (e.g., zshrc, zprofile, bashrc).
	// Defaults to "zshrc" if not specified.
	Target string `yaml:"target,omitempty"`

	// Priority determines the order within a target file (0-100, lower = earlier).
	// Defaults to 50 if not specified.
	Priority int `yaml:"priority,omitempty"`
}

// GetTarget returns the target RC file, defaulting to "zshrc".
func (m *Module) GetTarget() string {
	if m.Target == "" {
		return "zshrc"
	}
	return m.Target
}

// GetPriority returns the priority, defaulting to 50.
func (m *Module) GetPriority() int {
	if m.Priority == 0 {
		return 50
	}
	return m.Priority
}

// AppliesTo checks if this module applies to the target OS.
// If OS field is empty, module applies to all operating systems.
func (m *Module) AppliesTo(targetOS string) bool {
	if len(m.OS) == 0 {
		return true // No OS restriction
	}

	targetOS = strings.ToLower(targetOS)
	for _, os := range m.OS {
		if strings.EqualFold(os, targetOS) {
			return true
		}
	}
	return false
}

// Validate checks if the module has required fields.
func (m *Module) Validate() error {
	if m.Name == "" {
		return NewValidationError("module missing 'name' field")
	}
	if m.File == "" {
		return NewValidationError("module '%s' missing 'file' field", m.Name)
	}
	return nil
}
