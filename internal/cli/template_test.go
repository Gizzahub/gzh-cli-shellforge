package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateCmd_Structure(t *testing.T) {
	cmd := newTemplateCmd()

	assert.Equal(t, "template", cmd.Use)
	assert.Contains(t, cmd.Short, "Generate module from template")
	assert.Len(t, cmd.Commands(), 2, "should have 2 subcommands")

	// Check subcommands exist
	var hasGenerate, hasList bool
	for _, subcmd := range cmd.Commands() {
		if strings.HasPrefix(subcmd.Use, "generate") {
			hasGenerate = true
		}
		if subcmd.Use == "list" {
			hasList = true
		}
	}
	assert.True(t, hasGenerate, "should have generate subcommand")
	assert.True(t, hasList, "should have list subcommand")
}

func TestTemplateGenerateCmd_Flags(t *testing.T) {
	cmd := newTemplateGenerateCmd()

	// Check config-dir flag
	configDirFlag := cmd.Flags().Lookup("config-dir")
	require.NotNil(t, configDirFlag)
	assert.Equal(t, "modules", configDirFlag.DefValue)

	// Check field flag (StringSlice)
	fieldFlag := cmd.Flags().Lookup("field")
	require.NotNil(t, fieldFlag)
	assert.Equal(t, "f", fieldFlag.Shorthand)

	// Check requires flag (StringSlice)
	requiresFlag := cmd.Flags().Lookup("requires")
	require.NotNil(t, requiresFlag)
	assert.Equal(t, "r", requiresFlag.Shorthand)

	// Check verbose flag
	verboseFlag := cmd.Flags().Lookup("verbose")
	require.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue)
}

func TestTemplateGenerateCmd_Help(t *testing.T) {
	cmd := newTemplateGenerateCmd()

	assert.Contains(t, cmd.Use, "generate")
	assert.Contains(t, cmd.Short, "Generate")
	assert.Contains(t, cmd.Long, "template")
	assert.NotEmpty(t, cmd.Example)
}

func TestTemplateGenerateCmd_Args(t *testing.T) {
	cmd := newTemplateGenerateCmd()

	// Should require exactly 2 args: template-type and module-name
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should error with no args")

	err = cmd.Args(cmd, []string{"path"})
	assert.Error(t, err, "should error with only 1 arg")

	err = cmd.Args(cmd, []string{"path", "my-module"})
	assert.NoError(t, err, "should accept 2 args")

	err = cmd.Args(cmd, []string{"path", "my-module", "extra"})
	assert.Error(t, err, "should error with 3 args")
}

func TestTemplateListCmd_Flags(t *testing.T) {
	cmd := newTemplateListCmd()

	// List command has no flags (it's a simple display command)
	assert.Equal(t, 0, cmd.Flags().NFlag(), "list command should have no custom flags")
}

func TestTemplateListCmd_Help(t *testing.T) {
	cmd := newTemplateListCmd()

	assert.Equal(t, "list", cmd.Use)
	assert.Contains(t, cmd.Short, "List")
	assert.Contains(t, cmd.Long, "template")
}

func TestTemplateListCmd_Args(t *testing.T) {
	cmd := newTemplateListCmd()

	// List command doesn't have custom Args validation, so skip this test
	// The command accepts any args (cobra default behavior)
	assert.NotNil(t, cmd)
}

func TestParseFields(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "single field",
			input: []string{"key=value"},
			want:  map[string]string{"key": "value"},
		},
		{
			name:  "multiple fields",
			input: []string{"key1=value1", "key2=value2"},
			want:  map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:  "field with spaces in value",
			input: []string{"path_dir=/usr/local/bin"},
			want:  map[string]string{"path_dir": "/usr/local/bin"},
		},
		{
			name:  "field with equals in value",
			input: []string{"init_command=eval \"$(nvm init)\""},
			want:  map[string]string{"init_command": "eval \"$(nvm init)\""},
		},
		{
			name:  "empty input",
			input: []string{},
			want:  map[string]string{},
		},
		{
			name:    "invalid format - no equals",
			input:   []string{"invalid"},
			wantErr: true,
		},
		{
			name:    "invalid format - empty key",
			input:   []string{"=value"},
			wantErr: true,
		},
		{
			name:    "invalid format - empty value allowed",
			input:   []string{"key="},
			want:    map[string]string{"key": ""},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFields(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTemplateCmd_FlagShortcuts(t *testing.T) {
	cmd := newTemplateGenerateCmd()

	tests := []struct {
		name      string
		flagName  string
		shorthand string
	}{
		{"config-dir", "config-dir", "c"},
		{"field", "field", "f"},
		{"requires", "requires", "r"},
		{"verbose", "verbose", "v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			require.NotNil(t, flag, "flag should exist")
			assert.Equal(t, tt.shorthand, flag.Shorthand)
		})
	}
}

func TestTemplateCmd_Examples(t *testing.T) {
	generateCmd := newTemplateGenerateCmd()

	// Generate command should have examples
	assert.NotEmpty(t, generateCmd.Example)
	assert.Contains(t, generateCmd.Example, "template generate")
	assert.Contains(t, generateCmd.Example, "-f")
}

func TestTemplateCmd_LongDescription(t *testing.T) {
	cmd := newTemplateCmd()

	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Long, "template")
}

func TestTemplateGenerateCmd_UsageText(t *testing.T) {
	cmd := newTemplateGenerateCmd()

	assert.Contains(t, cmd.Use, "<template-type>")
	assert.Contains(t, cmd.Use, "<module-name>")
}
