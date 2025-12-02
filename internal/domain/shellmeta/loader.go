package shellmeta

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// Loader loads shell profile metadata from YAML files.
type Loader struct {
	fs afero.Fs
}

// NewLoader creates a new Loader with the given filesystem.
func NewLoader(fs afero.Fs) *Loader {
	return &Loader{fs: fs}
}

// Load loads all shell profile metadata from the specified directory.
// The directory should contain: core.yaml, contexts.yaml, dev.yaml, automation.yaml
func (l *Loader) Load(dir string) (*ShellProfiles, error) {
	profiles := &ShellProfiles{}

	// Load core.yaml
	if err := l.loadYAML(filepath.Join(dir, "core.yaml"), &profiles.Core); err != nil {
		return nil, fmt.Errorf("failed to load core.yaml: %w", err)
	}

	// Load contexts.yaml
	if err := l.loadYAML(filepath.Join(dir, "contexts.yaml"), &profiles.Contexts); err != nil {
		return nil, fmt.Errorf("failed to load contexts.yaml: %w", err)
	}

	// Load dev.yaml
	if err := l.loadYAML(filepath.Join(dir, "dev.yaml"), &profiles.Dev); err != nil {
		return nil, fmt.Errorf("failed to load dev.yaml: %w", err)
	}

	// Load automation.yaml
	if err := l.loadYAML(filepath.Join(dir, "automation.yaml"), &profiles.Automation); err != nil {
		return nil, fmt.Errorf("failed to load automation.yaml: %w", err)
	}

	return profiles, nil
}

// loadYAML is a generic helper that reads and unmarshals a YAML file into the given target.
// The target must be a pointer to the destination struct.
func (l *Loader) loadYAML(path string, target interface{}) error {
	data, err := afero.ReadFile(l.fs, path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filepath.Base(path), err)
	}

	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	return nil
}

// LoadCore loads only the core.yaml file.
func (l *Loader) LoadCore(path string) (*CoreProfiles, error) {
	var core CoreProfiles
	if err := l.loadYAML(path, &core); err != nil {
		return nil, err
	}
	return &core, nil
}

// LoadContexts loads only the contexts.yaml file.
func (l *Loader) LoadContexts(path string) (*ContextProfiles, error) {
	var contexts ContextProfiles
	if err := l.loadYAML(path, &contexts); err != nil {
		return nil, err
	}
	return &contexts, nil
}

// LoadDev loads only the dev.yaml file.
func (l *Loader) LoadDev(path string) (*DevProfiles, error) {
	var dev DevProfiles
	if err := l.loadYAML(path, &dev); err != nil {
		return nil, err
	}
	return &dev, nil
}

// LoadAutomation loads only the automation.yaml file.
func (l *Loader) LoadAutomation(path string) (*AutomationProfiles, error) {
	var automation AutomationProfiles
	if err := l.loadYAML(path, &automation); err != nil {
		return nil, err
	}
	return &automation, nil
}
