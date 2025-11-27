package yamlparser

import (
	"fmt"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// Parser implements YAML manifest parsing.
type Parser struct {
	fs afero.Fs
}

// New creates a new YAML parser with the given filesystem.
func New(fs afero.Fs) *Parser {
	return &Parser{fs: fs}
}

// Parse reads and parses a YAML manifest file.
// Returns a Manifest or an error if parsing fails.
func (p *Parser) Parse(path string) (*domain.Manifest, error) {
	// Read file
	data, err := afero.ReadFile(p.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Parse YAML
	var manifest domain.Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &manifest, nil
}
