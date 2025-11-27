package domain

import "strings"

// Module represents a shell module with dependencies and OS filtering.
type Module struct {
	Name        string   `yaml:"name"`
	File        string   `yaml:"file"`
	Requires    []string `yaml:"requires,omitempty"`
	OS          []string `yaml:"os,omitempty"`
	Description string   `yaml:"description,omitempty"`
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
