package domain

// ShellConfig configures shell type for manifest v2.
type ShellConfig struct {
	Type string `yaml:"type"` // zsh, bash, fish
}

// OutputConfig configures output settings for manifest v2.
type OutputConfig struct {
	Directory string `yaml:"directory,omitempty"` // Output directory (defaults to ~)
	Backup    bool   `yaml:"backup,omitempty"`    // Create backup of existing files
}

// Manifest represents a collection of shell modules.
type Manifest struct {
	Version string       `yaml:"version,omitempty"` // Manifest version ("1" or "2")
	Shell   ShellConfig  `yaml:"shell,omitempty"`   // Shell configuration (v2)
	Output  OutputConfig `yaml:"output,omitempty"`  // Output configuration (v2)
	Modules []Module     `yaml:"modules"`
}

// IsLegacy returns true if this is a v1 (legacy) manifest without version or target fields.
func (m *Manifest) IsLegacy() bool {
	if m.Version != "" && m.Version != "1" {
		return false
	}
	// Also check if any module has target set
	for _, mod := range m.Modules {
		if mod.Target != "" {
			return false
		}
	}
	return true
}

// GetShellType returns the shell type, defaulting to "zsh".
func (m *Manifest) GetShellType() string {
	if m.Shell.Type == "" {
		return "zsh"
	}
	return m.Shell.Type
}

// GetOutputDirectory returns the output directory, defaulting to "~".
func (m *Manifest) GetOutputDirectory() string {
	if m.Output.Directory == "" {
		return "~"
	}
	return m.Output.Directory
}

// FindModule finds a module by name.
// Returns the module and true if found, nil and false otherwise.
func (m *Manifest) FindModule(name string) (*Module, bool) {
	for i := range m.Modules {
		if m.Modules[i].Name == name {
			return &m.Modules[i], true
		}
	}
	return nil, false
}

// Validate checks the manifest for errors.
// Returns a slice of all validation errors found.
func (m *Manifest) Validate() []error {
	var errors []error

	// Check for duplicate module names
	seen := make(map[string]bool)
	for _, mod := range m.Modules {
		if seen[mod.Name] {
			errors = append(errors, NewValidationError("duplicate module name: %s", mod.Name))
		}
		seen[mod.Name] = true

		// Validate each module
		if err := mod.Validate(); err != nil {
			errors = append(errors, err)
		}
	}

	// Check that all dependencies reference existing modules
	for _, mod := range m.Modules {
		for _, dep := range mod.Requires {
			if _, found := m.FindModule(dep); !found {
				errors = append(errors, NewValidationError(
					"module '%s' requires non-existent module '%s'", mod.Name, dep))
			}
		}
	}

	return errors
}
