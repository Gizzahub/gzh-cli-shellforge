package cli

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/filesystem"
)

func TestValidateCmd_Flags(t *testing.T) {
	cmd := newValidateCmd()

	// Check flags with defaults
	manifestFlag := cmd.Flags().Lookup("manifest")
	require.NotNil(t, manifestFlag)
	assert.Equal(t, "manifest.yaml", manifestFlag.DefValue)

	configDirFlag := cmd.Flags().Lookup("config-dir")
	require.NotNil(t, configDirFlag)
	assert.Equal(t, "modules", configDirFlag.DefValue)

	verboseFlag := cmd.Flags().Lookup("verbose")
	require.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue)
}

func TestValidateCmd_Help(t *testing.T) {
	cmd := newValidateCmd()

	assert.Equal(t, "validate", cmd.Use)
	assert.Contains(t, cmd.Short, "Validate manifest")
	assert.Contains(t, cmd.Long, "circular dependencies")
	assert.NotEmpty(t, cmd.Example)
}

func TestValidateCmd_FlagShortcuts(t *testing.T) {
	cmd := newValidateCmd()

	// Test short flag versions
	tests := []struct {
		flag      string
		shorthand string
	}{
		{"config-dir", "c"},
		{"manifest", "m"},
		{"verbose", "v"},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flag)
			require.NotNil(t, flag)
			assert.Equal(t, tt.shorthand, flag.Shorthand)
		})
	}
}

func TestValidateCmd_Examples(t *testing.T) {
	cmd := newValidateCmd()

	examples := cmd.Example

	assert.Contains(t, examples, "shellforge validate", "should show basic example")
	assert.Contains(t, examples, "--manifest", "should show custom manifest example")
	assert.Contains(t, examples, "--verbose", "should show verbose example")
}

func TestRunValidate_Success(t *testing.T) {
	// Integration test with real examples
	// Skip if examples directory doesn't exist
	manifestPath := "../../examples/manifest.yaml"

	// Check if file exists using OS filesystem
	reader := filesystem.NewReader(afero.NewOsFs())
	if !reader.FileExists(manifestPath) {
		t.Skip("Example files not found, skipping integration test")
	}

	flags := &validateFlags{
		configDir: "../../examples/modules",
		manifest:  manifestPath,
		verbose:   false,
	}

	err := runValidate(flags)
	assert.NoError(t, err, "validation should succeed with valid manifest")
}

func TestRunValidate_VerboseOutput(t *testing.T) {
	// Test verbose mode
	manifestPath := "../../examples/manifest.yaml"

	// Check if file exists using OS filesystem
	reader := filesystem.NewReader(afero.NewOsFs())
	if !reader.FileExists(manifestPath) {
		t.Skip("Example files not found, skipping integration test")
	}

	flags := &validateFlags{
		configDir: "../../examples/modules",
		manifest:  manifestPath,
		verbose:   true,
	}

	err := runValidate(flags)
	assert.NoError(t, err, "validation should succeed in verbose mode")
}

func TestDependencyValidator_CircularDependency(t *testing.T) {
	validator := &dependencyValidator{}

	// Create a manifest with circular dependency
	manifest := &domain.Manifest{
		Modules: []domain.Module{
			{Name: "a", Requires: []string{"b"}},
			{Name: "b", Requires: []string{"a"}},
		},
	}

	err := validator.checkCircularDependencies(manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestDependencyValidator_NoCycle(t *testing.T) {
	validator := &dependencyValidator{}

	// Create a manifest with valid dependencies
	manifest := &domain.Manifest{
		Modules: []domain.Module{
			{Name: "a", Requires: []string{}},
			{Name: "b", Requires: []string{"a"}},
			{Name: "c", Requires: []string{"b"}},
		},
	}

	err := validator.checkCircularDependencies(manifest)
	assert.NoError(t, err)
}

func TestDependencyValidator_ComplexCycle(t *testing.T) {
	validator := &dependencyValidator{}

	// Create a manifest with complex circular dependency
	manifest := &domain.Manifest{
		Modules: []domain.Module{
			{Name: "a", Requires: []string{"b"}},
			{Name: "b", Requires: []string{"c"}},
			{Name: "c", Requires: []string{"a"}},
		},
	}

	err := validator.checkCircularDependencies(manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestDependencyValidator_SelfDependency(t *testing.T) {
	validator := &dependencyValidator{}

	// Create a manifest with self-dependency
	manifest := &domain.Manifest{
		Modules: []domain.Module{
			{Name: "a", Requires: []string{"a"}},
		},
	}

	err := validator.checkCircularDependencies(manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestDependencyValidator_DAG(t *testing.T) {
	validator := &dependencyValidator{}

	// Create a manifest with valid DAG (directed acyclic graph)
	manifest := &domain.Manifest{
		Modules: []domain.Module{
			{Name: "a", Requires: []string{}},
			{Name: "b", Requires: []string{}},
			{Name: "c", Requires: []string{"a", "b"}},
			{Name: "d", Requires: []string{"c"}},
		},
	}

	err := validator.checkCircularDependencies(manifest)
	assert.NoError(t, err)
}

func TestValidateCmd_LongDescription(t *testing.T) {
	cmd := newValidateCmd()

	long := cmd.Long

	// Verify long description contains key concepts
	expectedTerms := []string{
		"manifest",
		"circular dependencies",
		"module files",
		"validation",
	}

	for _, term := range expectedTerms {
		assert.Contains(t, long, term,
			"long description should mention '%s'", term)
	}
}

func TestValidateCmd_UsageText(t *testing.T) {
	cmd := newValidateCmd()

	usage := cmd.UsageString()

	// Verify it contains key information
	assert.Contains(t, usage, "validate", "usage should contain command name")
	assert.Contains(t, usage, "Flags:", "usage should list flags")
	assert.Contains(t, usage, "--manifest", "usage should show --manifest flag")
}
