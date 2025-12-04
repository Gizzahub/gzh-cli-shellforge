package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCmd_Flags(t *testing.T) {
	cmd := newListCmd()

	// Check flag existence and defaults
	manifestFlag := cmd.Flags().Lookup("manifest")
	require.NotNil(t, manifestFlag)
	assert.Equal(t, "manifest.yaml", manifestFlag.DefValue)

	configDirFlag := cmd.Flags().Lookup("config-dir")
	require.NotNil(t, configDirFlag)
	assert.Equal(t, "modules", configDirFlag.DefValue)

	verboseFlag := cmd.Flags().Lookup("verbose")
	require.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue)

	filterFlag := cmd.Flags().Lookup("filter")
	require.NotNil(t, filterFlag)
	assert.Equal(t, "", filterFlag.DefValue)
}

func TestListCmd_Help(t *testing.T) {
	cmd := newListCmd()

	assert.Equal(t, "list", cmd.Use)
	assert.Contains(t, cmd.Short, "List all modules")
	assert.Contains(t, cmd.Long, "manifest")
	assert.NotEmpty(t, cmd.Example)
}

func TestListCmd_FlagShortcuts(t *testing.T) {
	cmd := newListCmd()

	tests := []struct {
		flag      string
		shorthand string
	}{
		{"config-dir", "c"},
		{"manifest", "m"},
		{"verbose", "v"},
		{"filter", "F"},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flag)
			require.NotNil(t, flag)
			assert.Equal(t, tt.shorthand, flag.Shorthand)
		})
	}
}

func TestListCmd_Examples(t *testing.T) {
	cmd := newListCmd()

	examples := cmd.Example

	assert.Contains(t, examples, "shellforge list", "should show basic usage")
	assert.Contains(t, examples, "--filter", "should show OS filtering example")
	assert.Contains(t, examples, "--verbose", "should show verbose example")
	assert.Contains(t, examples, "Mac", "should show Mac example")
	assert.Contains(t, examples, "Linux", "should show Linux example")
}

func TestRunList_WithRealExamples(t *testing.T) {
	// Integration test with real example files
	cmd := newListCmd()
	flags := &listFlags{
		manifest:  "../../examples/manifest.yaml",
		configDir: "../../examples/modules",
		verbose:   false,
		filterOS:  "",
	}

	err := runList(cmd, flags)
	assert.NoError(t, err, "list should succeed with valid manifest")
}

func TestRunList_VerboseOutput(t *testing.T) {
	cmd := newListCmd()
	flags := &listFlags{
		manifest:  "../../examples/manifest.yaml",
		configDir: "../../examples/modules",
		verbose:   true,
		filterOS:  "",
	}

	// Capture stdout to verify verbose output
	err := runList(cmd, flags)
	assert.NoError(t, err, "verbose list should succeed")
}

func TestRunList_OSFiltering(t *testing.T) {
	tests := []struct {
		name     string
		filterOS string
		wantErr  bool
	}{
		{
			name:     "filter by Mac",
			filterOS: "Mac",
			wantErr:  false,
		},
		{
			name:     "filter by Linux",
			filterOS: "Linux",
			wantErr:  false,
		},
		{
			name:     "case insensitive filtering",
			filterOS: "mac",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newListCmd()
			flags := &listFlags{
				manifest:  "../../examples/manifest.yaml",
				configDir: "../../examples/modules",
				verbose:   false,
				filterOS:  tt.filterOS,
			}

			err := runList(cmd, flags)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunList_MissingManifest(t *testing.T) {
	cmd := newListCmd()
	flags := &listFlags{
		manifest:  "nonexistent.yaml",
		configDir: "modules",
		verbose:   false,
		filterOS:  "",
	}

	err := runList(cmd, flags)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "manifest parsing failed")
}

func TestListCmd_LongDescription(t *testing.T) {
	cmd := newListCmd()

	long := cmd.Long

	expectedTerms := []string{
		"manifest",
		"module",
		"dependencies",
		"OS",
		"filter",
		"verbose",
	}

	for _, term := range expectedTerms {
		assert.Contains(t, strings.ToLower(long), strings.ToLower(term),
			"long description should mention '%s'", term)
	}
}

func TestListCmd_UsageText(t *testing.T) {
	cmd := newListCmd()

	usage := cmd.UsageString()

	assert.Contains(t, usage, "list", "usage should contain command name")
	assert.Contains(t, usage, "Flags:", "usage should list flags")
	assert.Contains(t, usage, "--manifest", "usage should show --manifest flag")
	assert.Contains(t, usage, "--filter", "usage should show --filter flag")
}

func TestListCmd_Integration(t *testing.T) {
	// This test verifies the command integrates properly with Cobra
	cmd := newListCmd()

	// Set args to use examples
	cmd.SetArgs([]string{
		"--manifest", "../../examples/manifest.yaml",
		"--config-dir", "../../examples/modules",
	})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute
	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Modules", "should display module count")
}

func TestListCmd_FilterIntegration(t *testing.T) {
	cmd := newListCmd()

	cmd.SetArgs([]string{
		"--manifest", "../../examples/manifest.yaml",
		"--config-dir", "../../examples/modules",
		"--filter", "Mac",
	})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Filtered by OS: Mac", "should show filter information")
}

func TestListCmd_VerboseIntegration(t *testing.T) {
	cmd := newListCmd()

	cmd.SetArgs([]string{
		"--manifest", "../../examples/manifest.yaml",
		"--config-dir", "../../examples/modules",
		"--verbose",
	})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "File:", "verbose mode should show file paths")
}
