package domain

// Manifest represents a collection of shell modules.
type Manifest struct {
	Modules []Module `yaml:"modules"`
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
