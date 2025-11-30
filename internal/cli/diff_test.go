package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffCmd_Structure(t *testing.T) {
	cmd := newDiffCmd()
	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "diff")
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestDiffCmd_Flags(t *testing.T) {
	cmd := newDiffCmd()

	tests := []struct {
		name      string
		flagName  string
		flagType  string
		shorthand string
		hasFlag   bool
	}{
		{"format flag exists", "format", "string", "f", true},
		{"verbose flag exists", "verbose", "bool", "v", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if tt.hasFlag {
				require.NotNil(t, flag, "flag %s should exist", tt.flagName)
				assert.Equal(t, tt.flagType, flag.Value.Type(), "flag %s should be %s", tt.flagName, tt.flagType)
				assert.Equal(t, tt.shorthand, flag.Shorthand, "flag %s should have shorthand %s", tt.flagName, tt.shorthand)
			} else {
				assert.Nil(t, flag, "flag %s should not exist", tt.flagName)
			}
		})
	}
}

func TestDiffCmd_Help(t *testing.T) {
	cmd := newDiffCmd()

	// Test help output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "diff")
	assert.Contains(t, output, "Compare")
	assert.Contains(t, output, "format")
	assert.Contains(t, output, "verbose")
}

func TestDiffCmd_Args(t *testing.T) {
	cmd := newDiffCmd()

	// Test missing required arguments
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err, "should error when missing arguments")

	// Test with only one argument
	cmd.SetArgs([]string{"file1.sh"})
	err = cmd.Execute()
	assert.Error(t, err, "should error when missing second argument")
}

func TestDiffCmd_FlagShortcuts(t *testing.T) {
	cmd := newDiffCmd()

	tests := []struct {
		name  string
		short string
		long  string
	}{
		{"format", "-f", "--format"},
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

func TestDiffCmd_Examples(t *testing.T) {
	cmd := newDiffCmd()

	assert.NotEmpty(t, cmd.Example, "command should have examples")
	assert.Contains(t, cmd.Example, "gz-shellforge diff")
	assert.Contains(t, cmd.Example, "--format")
}

func TestDiffCmd_LongDescription(t *testing.T) {
	cmd := newDiffCmd()

	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Long, "Compare")
	assert.Contains(t, cmd.Long, "summary")
	assert.Contains(t, cmd.Long, "unified")
	assert.Contains(t, cmd.Long, "context")
	assert.Contains(t, cmd.Long, "side-by-side")
}

func TestDiffCmd_UsageText(t *testing.T) {
	cmd := newDiffCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	cmd.Execute()

	output := buf.String()
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "diff <original-file> <generated-file>")
}

func TestDiffCmd_DefaultValues(t *testing.T) {
	cmd := newDiffCmd()

	formatFlag := cmd.Flags().Lookup("format")
	require.NotNil(t, formatFlag)
	assert.Equal(t, "summary", formatFlag.DefValue, "format should default to 'summary'")

	verboseFlag := cmd.Flags().Lookup("verbose")
	require.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue, "verbose should default to false")
}

func TestDiffCmd_Integration(t *testing.T) {
	// Create root command and add diff as subcommand
	root := &cobra.Command{Use: "shellforge"}
	root.AddCommand(newDiffCmd())

	// Test command is properly registered
	cmd, _, err := root.Find([]string{"diff"})
	require.NoError(t, err)
	assert.Equal(t, "diff", cmd.Name())
}

func TestDiffCmd_FormatValidation(t *testing.T) {
	cmd := newDiffCmd()

	formatFlag := cmd.Flags().Lookup("format")
	require.NotNil(t, formatFlag)

	// Test valid formats in help text
	assert.Contains(t, formatFlag.Usage, "summary")
	assert.Contains(t, formatFlag.Usage, "unified")
	assert.Contains(t, formatFlag.Usage, "context")
	assert.Contains(t, formatFlag.Usage, "side-by-side")
}
