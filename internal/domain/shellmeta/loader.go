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
	core, err := l.loadCore(filepath.Join(dir, "core.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to load core.yaml: %w", err)
	}
	profiles.Core = core

	// Load contexts.yaml
	contexts, err := l.loadContexts(filepath.Join(dir, "contexts.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to load contexts.yaml: %w", err)
	}
	profiles.Contexts = contexts

	// Load dev.yaml
	dev, err := l.loadDev(filepath.Join(dir, "dev.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to load dev.yaml: %w", err)
	}
	profiles.Dev = dev

	// Load automation.yaml
	automation, err := l.loadAutomation(filepath.Join(dir, "automation.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to load automation.yaml: %w", err)
	}
	profiles.Automation = automation

	return profiles, nil
}

// LoadCore loads only the core.yaml file.
func (l *Loader) LoadCore(path string) (*CoreProfiles, error) {
	return l.loadCore(path)
}

// LoadContexts loads only the contexts.yaml file.
func (l *Loader) LoadContexts(path string) (*ContextProfiles, error) {
	return l.loadContexts(path)
}

// LoadDev loads only the dev.yaml file.
func (l *Loader) LoadDev(path string) (*DevProfiles, error) {
	return l.loadDev(path)
}

// LoadAutomation loads only the automation.yaml file.
func (l *Loader) LoadAutomation(path string) (*AutomationProfiles, error) {
	return l.loadAutomation(path)
}

func (l *Loader) loadCore(path string) (*CoreProfiles, error) {
	data, err := afero.ReadFile(l.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var core CoreProfiles
	if err := yaml.Unmarshal(data, &core); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &core, nil
}

func (l *Loader) loadContexts(path string) (*ContextProfiles, error) {
	data, err := afero.ReadFile(l.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var contexts ContextProfiles
	if err := yaml.Unmarshal(data, &contexts); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &contexts, nil
}

func (l *Loader) loadDev(path string) (*DevProfiles, error) {
	data, err := afero.ReadFile(l.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var dev DevProfiles
	if err := yaml.Unmarshal(data, &dev); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &dev, nil
}

func (l *Loader) loadAutomation(path string) (*AutomationProfiles, error) {
	data, err := afero.ReadFile(l.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var automation AutomationProfiles
	if err := yaml.Unmarshal(data, &automation); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &automation, nil
}
