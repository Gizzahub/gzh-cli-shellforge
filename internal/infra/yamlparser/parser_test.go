package yamlparser

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		errMsg   string
		validate func(*testing.T, *domain.Manifest)
	}{
		{
			name: "valid manifest",
			content: `modules:
  - name: test-module
    file: test.sh
    requires: []
    os: [Mac]
    description: Test module
`,
			wantErr: false,
			validate: func(t *testing.T, m *domain.Manifest) {
				require.Len(t, m.Modules, 1)
				assert.Equal(t, "test-module", m.Modules[0].Name)
				assert.Equal(t, "test.sh", m.Modules[0].File)
				assert.Equal(t, []string{"Mac"}, m.Modules[0].OS)
				assert.Equal(t, "Test module", m.Modules[0].Description)
			},
		},
		{
			name: "multiple modules with dependencies",
			content: `modules:
  - name: base
    file: base.sh
    requires: []
    os: [Mac, Linux]
  - name: dependent
    file: dependent.sh
    requires: [base]
    os: [Mac]
`,
			wantErr: false,
			validate: func(t *testing.T, m *domain.Manifest) {
				require.Len(t, m.Modules, 2)
				assert.Equal(t, "base", m.Modules[0].Name)
				assert.Equal(t, "dependent", m.Modules[1].Name)
				assert.Equal(t, []string{"base"}, m.Modules[1].Requires)
			},
		},
		{
			name: "optional fields omitted",
			content: `modules:
  - name: minimal
    file: minimal.sh
`,
			wantErr: false,
			validate: func(t *testing.T, m *domain.Manifest) {
				require.Len(t, m.Modules, 1)
				assert.Equal(t, "minimal", m.Modules[0].Name)
				assert.Empty(t, m.Modules[0].Requires)
				assert.Empty(t, m.Modules[0].OS)
				assert.Empty(t, m.Modules[0].Description)
			},
		},
		{
			name:    "invalid YAML syntax",
			content: "modules:\n  - name: test\n    invalid yaml: [",
			wantErr: true,
			errMsg:  "failed to parse YAML",
		},
		{
			name:    "empty file",
			content: "",
			wantErr: false,
			validate: func(t *testing.T, m *domain.Manifest) {
				assert.Empty(t, m.Modules)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create in-memory filesystem
			fs := afero.NewMemMapFs()
			afero.WriteFile(fs, "manifest.yaml", []byte(tt.content), 0o644)

			// Parse
			parser := New(fs)
			manifest, err := parser.Parse("manifest.yaml")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, manifest)
				if tt.validate != nil {
					tt.validate(t, manifest)
				}
			}
		})
	}
}

func TestParser_Parse_FileErrors(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		errMsg string
	}{
		{
			name:   "file not found",
			path:   "/nonexistent/manifest.yaml",
			errMsg: "failed to read manifest file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			parser := New(fs)
			manifest, err := parser.Parse(tt.path)

			assert.Error(t, err)
			assert.Nil(t, manifest)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// TestParser_Parse_RealExample tests parsing the actual example manifest
func TestParser_Parse_RealExample(t *testing.T) {
	// This test assumes examples/manifest.yaml exists
	examplePath := "../../../examples/manifest.yaml"

	// Check if file exists (skip test if running outside project root)
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Skip("Example manifest not found, skipping")
	}

	// Use OS filesystem for real file
	fs := afero.NewOsFs()
	parser := New(fs)
	manifest, err := parser.Parse(examplePath)

	require.NoError(t, err)
	require.NotNil(t, manifest)

	// Verify some expected modules from examples/manifest.yaml
	assert.Greater(t, len(manifest.Modules), 5, "should have multiple modules")

	// Check for specific modules
	hasOSDetection := false
	hasBrewPath := false
	for _, mod := range manifest.Modules {
		if mod.Name == "os-detection" {
			hasOSDetection = true
			assert.Contains(t, mod.OS, "Mac")
			assert.Contains(t, mod.OS, "Linux")
		}
		if mod.Name == "brew-path" {
			hasBrewPath = true
			assert.Contains(t, mod.Requires, "os-detection")
			assert.Contains(t, mod.OS, "Mac")
		}
	}

	assert.True(t, hasOSDetection, "should have os-detection module")
	assert.True(t, hasBrewPath, "should have brew-path module")
}
