package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateCmd_Structure(t *testing.T) {
	cmd := newMigrateCmd()

	assert.Contains(t, cmd.Use, "migrate")
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
	assert.NotNil(t, cmd.RunE)

	// Verify it requires exactly 1 argument
	assert.NotNil(t, cmd.Args)
}

func TestMigrateCmd_Flags(t *testing.T) {
	cmd := newMigrateCmd()

	tests := []struct {
		name         string
		flagName     string
		shorthand    string
		defaultValue string
	}{
		{
			name:         "output-dir flag",
			flagName:     "output-dir",
			shorthand:    "o",
			defaultValue: "modules",
		},
		{
			name:         "manifest flag",
			flagName:     "manifest",
			shorthand:    "m",
			defaultValue: "manifest.yaml",
		},
		{
			name:      "dry-run flag",
			flagName:  "dry-run",
			shorthand: "",
		},
		{
			name:      "verbose flag",
			flagName:  "verbose",
			shorthand: "v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			require.NotNil(t, flag, "flag %s should exist", tt.flagName)

			if tt.shorthand != "" {
				assert.Equal(t, tt.shorthand, flag.Shorthand)
			}

			if tt.defaultValue != "" {
				assert.Equal(t, tt.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestMigrateCmd_Help(t *testing.T) {
	cmd := newMigrateCmd()

	// Verify help text contains key information
	assert.Contains(t, cmd.Long, "monolithic")
	assert.Contains(t, cmd.Long, "modular structure")
	assert.Contains(t, cmd.Long, "init.d")
	assert.Contains(t, cmd.Long, "rc_pre.d")
	assert.Contains(t, cmd.Long, "rc_post.d")

	// Verify examples are present
	assert.Contains(t, cmd.Example, "--dry-run")
	assert.Contains(t, cmd.Example, "--output-dir")
	assert.Contains(t, cmd.Example, "manifest")
}

func TestMigrateCmd_FlagShortcuts(t *testing.T) {
	cmd := newMigrateCmd()

	tests := []struct {
		name      string
		flagName  string
		shorthand string
	}{
		{"output-dir", "output-dir", "o"},
		{"manifest", "manifest", "m"},
		{"verbose", "verbose", "v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			require.NotNil(t, flag)
			assert.Equal(t, tt.shorthand, flag.Shorthand)
		})
	}
}

func TestMigrateCmd_Examples(t *testing.T) {
	cmd := newMigrateCmd()

	// Verify examples demonstrate key features
	examples := cmd.Example
	assert.Contains(t, examples, "gz-shellforge migrate")
	assert.Contains(t, examples, ".zshrc")
	assert.Contains(t, examples, ".bashrc")
	assert.Contains(t, examples, "--dry-run")
	assert.Contains(t, examples, "--output-dir")
	assert.Contains(t, examples, "-v")
}

func TestMigrateCmd_LongDescription(t *testing.T) {
	cmd := newMigrateCmd()

	longDesc := cmd.Long

	// Check for key concepts
	assert.Contains(t, longDesc, "monolithic")
	assert.Contains(t, longDesc, "modular")
	assert.Contains(t, longDesc, "section detection")
	assert.Contains(t, longDesc, "categorization")

	// Check for directory descriptions
	assert.Contains(t, longDesc, "init.d")
	assert.Contains(t, longDesc, "rc_pre.d")
	assert.Contains(t, longDesc, "rc_post.d")

	// Check for functionality descriptions
	assert.Contains(t, longDesc, "PATH")
	assert.Contains(t, longDesc, "Tool initialization")
	assert.Contains(t, longDesc, "Aliases")
	assert.Contains(t, longDesc, "manifest.yaml")
}

func TestMigrateCmd_UsageText(t *testing.T) {
	cmd := newMigrateCmd()

	usage := cmd.Use
	assert.Equal(t, "migrate <rc-file>", usage)
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		maxLines int
		want     []string
	}{
		{
			name:     "single line",
			content:  "line1",
			maxLines: 3,
			want:     []string{"line1"},
		},
		{
			name:     "multiple lines under limit",
			content:  "line1\nline2\nline3",
			maxLines: 5,
			want:     []string{"line1", "line2", "line3"},
		},
		{
			name:     "truncate at max lines",
			content:  "line1\nline2\nline3\nline4\nline5",
			maxLines: 3,
			want:     []string{"line1", "line2", "line3", "..."},
		},
		{
			name:     "empty content",
			content:  "",
			maxLines: 3,
			want:     []string{},
		},
		{
			name:     "exact max lines",
			content:  "line1\nline2\nline3",
			maxLines: 3,
			want:     []string{"line1", "line2", "line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(tt.content, tt.maxLines)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMigrateCmd_Integration(t *testing.T) {
	// Verify the command is properly structured for integration
	cmd := newMigrateCmd()

	// Should have RunE function
	assert.NotNil(t, cmd.RunE)

	// Should accept exactly 1 argument
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require at least 1 argument")

	err = cmd.Args(cmd, []string{"file1"})
	assert.NoError(t, err, "should accept 1 argument")

	err = cmd.Args(cmd, []string{"file1", "file2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestMigrateCmd_RootIntegration(t *testing.T) {
	// Verify migrate command is registered with root
	rootCmd := NewRootCmd()

	var migrateCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "migrate" {
			migrateCmd = cmd
			break
		}
	}

	require.NotNil(t, migrateCmd, "migrate command should be registered with root")
	assert.Equal(t, "migrate", migrateCmd.Name())
}
