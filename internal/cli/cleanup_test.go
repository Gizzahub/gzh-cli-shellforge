package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanupCmd_Structure(t *testing.T) {
	cmd := newCleanupCmd()
	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "cleanup")
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestCleanupCmd_Flags(t *testing.T) {
	cmd := newCleanupCmd()

	tests := []struct {
		name      string
		flagName  string
		flagType  string
		shorthand string
		required  bool
	}{
		{"file flag exists", "file", "string", "f", true},
		{"keep-count flag exists", "keep-count", "int", "", false},
		{"keep-days flag exists", "keep-days", "int", "", false},
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

func TestCleanupCmd_RequiredFlags(t *testing.T) {
	cmd := newCleanupCmd()

	// Test that file flag is required
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err, "should error when missing required file flag")
	assert.Contains(t, err.Error(), "required flag")
}

func TestCleanupCmd_Help(t *testing.T) {
	cmd := newCleanupCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "cleanup")
	assert.Contains(t, output, "Cleanup removes")
	assert.Contains(t, output, "file")
	assert.Contains(t, output, "keep-count")
	assert.Contains(t, output, "keep-days")
	assert.Contains(t, output, "backup-dir")
	assert.Contains(t, output, "no-git")
	assert.Contains(t, output, "dry-run")
	assert.Contains(t, output, "verbose")
	assert.Contains(t, output, "Retention Policy")
}

func TestCleanupCmd_FlagShortcuts(t *testing.T) {
	cmd := newCleanupCmd()

	tests := []struct {
		name  string
		short string
		long  string
	}{
		{"file", "-f", "--file"},
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

func TestCleanupCmd_Examples(t *testing.T) {
	cmd := newCleanupCmd()

	assert.NotEmpty(t, cmd.Example, "command should have examples")
	assert.Contains(t, cmd.Example, "gz-shellforge cleanup")
	assert.Contains(t, cmd.Example, "--file")
	assert.Contains(t, cmd.Example, "--keep-count")
	assert.Contains(t, cmd.Example, "--keep-days")
	assert.Contains(t, cmd.Example, "--dry-run")
	assert.Contains(t, cmd.Example, "--backup-dir")
	assert.Contains(t, cmd.Example, "--no-git")
}

func TestCleanupCmd_LongDescription(t *testing.T) {
	cmd := newCleanupCmd()

	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Long, "Cleanup")
	assert.Contains(t, cmd.Long, "retention")
	assert.Contains(t, cmd.Long, "policy")
	assert.Contains(t, cmd.Long, "Keep snapshots")
	assert.Contains(t, cmd.Long, "dry-run")
}

func TestCleanupCmd_UsageText(t *testing.T) {
	cmd := newCleanupCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	cmd.Execute()

	output := buf.String()
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "cleanup")
}

func TestCleanupCmd_DefaultValues(t *testing.T) {
	cmd := newCleanupCmd()

	keepCountFlag := cmd.Flags().Lookup("keep-count")
	require.NotNil(t, keepCountFlag)
	assert.Equal(t, "10", keepCountFlag.DefValue, "keep-count should default to 10")

	keepDaysFlag := cmd.Flags().Lookup("keep-days")
	require.NotNil(t, keepDaysFlag)
	assert.Equal(t, "30", keepDaysFlag.DefValue, "keep-days should default to 30")

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

func TestCleanupCmd_Integration(t *testing.T) {
	// Create root command and add cleanup as subcommand
	root := &cobra.Command{Use: "shellforge"}
	root.AddCommand(newCleanupCmd())

	// Test command is properly registered
	cmd, _, err := root.Find([]string{"cleanup"})
	require.NoError(t, err)
	assert.Equal(t, "cleanup", cmd.Name())
}

func TestCleanupCmd_RetentionPolicyDescription(t *testing.T) {
	cmd := newCleanupCmd()

	// Verify the retention policy is well documented in help text
	assert.Contains(t, cmd.Long, "Keep snapshots by count")
	assert.Contains(t, cmd.Long, "Keep snapshots by age")
	assert.Contains(t, cmd.Long, "union")
	assert.Contains(t, cmd.Long, "Safety")
	assert.Contains(t, cmd.Long, "at least one snapshot")
}
