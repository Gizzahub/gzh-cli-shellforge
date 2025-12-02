package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/helpers"
)

func TestBackupCmd_Structure(t *testing.T) {
	cmd := newBackupCmd()
	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "backup")
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestBackupCmd_Flags(t *testing.T) {
	cmd := newBackupCmd()

	tests := []struct {
		name      string
		flagName  string
		flagType  string
		shorthand string
		required  bool
	}{
		{"file flag exists", "file", "string", "f", true},
		{"message flag exists", "message", "string", "m", false},
		{"backup-dir flag exists", "backup-dir", "string", "", false},
		{"no-git flag exists", "no-git", "bool", "", false},
		{"verbose flag exists", "verbose", "bool", "v", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			require.NotNil(t, flag, "flag %s should exist", tt.flagName)
			assert.Equal(t, tt.flagType, flag.Value.Type(), "flag %s should be %s", tt.flagName, tt.flagType)
			assert.Equal(t, tt.shorthand, flag.Shorthand, "flag %s should have shorthand %s", tt.flagName, tt.shorthand)
		})
	}
}

func TestBackupCmd_RequiredFlags(t *testing.T) {
	cmd := newBackupCmd()

	// Test that file flag is required
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err, "should error when missing required file flag")
	assert.Contains(t, err.Error(), "required flag")
}

func TestBackupCmd_Help(t *testing.T) {
	cmd := newBackupCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "backup")
	assert.Contains(t, output, "Backup creates")
	assert.Contains(t, output, "file")
	assert.Contains(t, output, "message")
	assert.Contains(t, output, "backup-dir")
	assert.Contains(t, output, "no-git")
	assert.Contains(t, output, "verbose")
}

func TestBackupCmd_FlagShortcuts(t *testing.T) {
	cmd := newBackupCmd()

	tests := []struct {
		name  string
		short string
		long  string
	}{
		{"file", "-f", "--file"},
		{"message", "-m", "--message"},
		{"verbose", "-v", "--verbose"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortFlag := cmd.Flags().ShorthandLookup(tt.short[1:])
			longFlag := cmd.Flags().Lookup(tt.long[2:])

			require.NotNil(t, shortFlag, "short flag %s should exist", tt.short)
			require.NotNil(t, longFlag, "long flag %s should exist", tt.long)
			assert.Equal(t, shortFlag, longFlag, "short and long flags should be the same")
		})
	}
}

func TestBackupCmd_Examples(t *testing.T) {
	cmd := newBackupCmd()

	assert.NotEmpty(t, cmd.Example, "command should have examples")
	assert.Contains(t, cmd.Example, "gz-shellforge backup")
	assert.Contains(t, cmd.Example, "--file")
	assert.Contains(t, cmd.Example, "--message")
	assert.Contains(t, cmd.Example, "--no-git")
	assert.Contains(t, cmd.Example, "--backup-dir")
}

func TestBackupCmd_LongDescription(t *testing.T) {
	cmd := newBackupCmd()

	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Long, "Backup")
	assert.Contains(t, cmd.Long, "snapshot")
	assert.Contains(t, cmd.Long, "git")
	assert.Contains(t, cmd.Long, "restore")
}

func TestBackupCmd_UsageText(t *testing.T) {
	cmd := newBackupCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	cmd.Execute()

	output := buf.String()
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "backup")
}

func TestBackupCmd_DefaultValues(t *testing.T) {
	cmd := newBackupCmd()

	messageFlag := cmd.Flags().Lookup("message")
	require.NotNil(t, messageFlag)
	assert.Equal(t, "", messageFlag.DefValue, "message should default to empty string")

	backupDirFlag := cmd.Flags().Lookup("backup-dir")
	require.NotNil(t, backupDirFlag)
	assert.Equal(t, "", backupDirFlag.DefValue, "backup-dir should default to empty string")

	noGitFlag := cmd.Flags().Lookup("no-git")
	require.NotNil(t, noGitFlag)
	assert.Equal(t, "false", noGitFlag.DefValue, "no-git should default to false")

	verboseFlag := cmd.Flags().Lookup("verbose")
	require.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue, "verbose should default to false")
}

func TestBackupCmd_Integration(t *testing.T) {
	// Create root command and add backup as subcommand
	root := &cobra.Command{Use: "shellforge"}
	root.AddCommand(newBackupCmd())

	// Test command is properly registered
	cmd, _, err := root.Find([]string{"backup"})
	require.NoError(t, err)
	assert.Equal(t, "backup", cmd.Name())
}

func TestExpandHomePath(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantStart string
		wantErr   bool
	}{
		{
			name:      "tilde only",
			input:     "~",
			wantStart: "/",
			wantErr:   false,
		},
		{
			name:      "tilde with path",
			input:     "~/.zshrc",
			wantStart: "/",
			wantErr:   false,
		},
		{
			name:      "no tilde",
			input:     "/etc/zshrc",
			wantStart: "/etc/zshrc",
			wantErr:   false,
		},
		{
			name:      "empty string",
			input:     "",
			wantStart: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := helpers.ExpandHomePath(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantStart != "" {
					assert.Contains(t, got, tt.wantStart)
				}
			}
		})
	}
}

func TestGitRepositoryAdapter(t *testing.T) {
	// This is just a structural test to ensure the adapter exists
	// Full functionality should be tested in integration tests
	t.Run("adapter can be created", func(t *testing.T) {
		// We can't actually create a git.Repository without a real directory
		// This test just verifies the types exist
		assert.NotNil(t, newGitRepositoryAdapter)
	})
}
