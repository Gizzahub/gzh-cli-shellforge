package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestoreCmd_Structure(t *testing.T) {
	cmd := newRestoreCmd()
	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "restore")
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestRestoreCmd_Flags(t *testing.T) {
	cmd := newRestoreCmd()

	tests := []struct {
		name      string
		flagName  string
		flagType  string
		shorthand string
		required  bool
	}{
		{"file flag exists", "file", "string", "f", true},
		{"snapshot flag exists", "snapshot", "string", "s", true},
		{"backup-dir flag exists", "backup-dir", "string", "", false},
		{"no-git flag exists", "no-git", "bool", "", false},
		{"dry-run flag exists", "dry-run", "bool", "", false},
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

func TestRestoreCmd_RequiredFlags(t *testing.T) {
	cmd := newRestoreCmd()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errText string
	}{
		{
			name:    "missing both required flags",
			args:    []string{},
			wantErr: true,
			errText: "required flag",
		},
		{
			name:    "missing snapshot flag",
			args:    []string{"--file", "test.sh"},
			wantErr: true,
			errText: "required flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err, "should error when missing required flags")
				assert.Contains(t, err.Error(), tt.errText)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRestoreCmd_Help(t *testing.T) {
	cmd := newRestoreCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "restore")
	assert.Contains(t, output, "Restore")
	assert.Contains(t, output, "file")
	assert.Contains(t, output, "snapshot")
	assert.Contains(t, output, "backup-dir")
	assert.Contains(t, output, "no-git")
	assert.Contains(t, output, "dry-run")
	assert.Contains(t, output, "verbose")
}

func TestRestoreCmd_FlagShortcuts(t *testing.T) {
	cmd := newRestoreCmd()

	tests := []struct {
		name  string
		short string
		long  string
	}{
		{"file", "-f", "--file"},
		{"snapshot", "-s", "--snapshot"},
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

func TestRestoreCmd_Examples(t *testing.T) {
	cmd := newRestoreCmd()

	assert.NotEmpty(t, cmd.Example, "command should have examples")
	assert.Contains(t, cmd.Example, "gz-shellforge restore")
	assert.Contains(t, cmd.Example, "--file")
	assert.Contains(t, cmd.Example, "--snapshot")
	assert.Contains(t, cmd.Example, "--dry-run")
	assert.Contains(t, cmd.Example, "--backup-dir")
	assert.Contains(t, cmd.Example, "--no-git")
}

func TestRestoreCmd_LongDescription(t *testing.T) {
	cmd := newRestoreCmd()

	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Long, "Restore")
	assert.Contains(t, cmd.Long, "snapshot")
	assert.Contains(t, cmd.Long, "safety backup")
	assert.Contains(t, cmd.Long, "dry-run")
}

func TestRestoreCmd_UsageText(t *testing.T) {
	cmd := newRestoreCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	cmd.Execute()

	output := buf.String()
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "restore")
}

func TestRestoreCmd_DefaultValues(t *testing.T) {
	cmd := newRestoreCmd()

	backupDirFlag := cmd.Flags().Lookup("backup-dir")
	require.NotNil(t, backupDirFlag)
	assert.Equal(t, "", backupDirFlag.DefValue, "backup-dir should default to empty string")

	noGitFlag := cmd.Flags().Lookup("no-git")
	require.NotNil(t, noGitFlag)
	assert.Equal(t, "false", noGitFlag.DefValue, "no-git should default to false")

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	require.NotNil(t, dryRunFlag)
	assert.Equal(t, "false", dryRunFlag.DefValue, "dry-run should default to false")

	verboseFlag := cmd.Flags().Lookup("verbose")
	require.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue, "verbose should default to false")
}

func TestRestoreCmd_Integration(t *testing.T) {
	// Create root command and add restore as subcommand
	root := &cobra.Command{Use: "shellforge"}
	root.AddCommand(newRestoreCmd())

	// Test command is properly registered
	cmd, _, err := root.Find([]string{"restore"})
	require.NoError(t, err)
	assert.Equal(t, "restore", cmd.Name())
}
