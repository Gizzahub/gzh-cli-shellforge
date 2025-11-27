package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	assert.Equal(t, "shellforge", cmd.Use)
	assert.Contains(t, cmd.Short, "Build tool for modular shell configurations")
	assert.NotEmpty(t, cmd.Long)
	assert.Equal(t, version, cmd.Version)
}

func TestRootCmd_Version(t *testing.T) {
	cmd := NewRootCmd()

	assert.Equal(t, "0.1.0", cmd.Version)
}

func TestRootCmd_HasSubcommands(t *testing.T) {
	cmd := NewRootCmd()

	// Check that all expected subcommands exist
	expectedSubcommands := []string{"build", "validate", "list"}

	for _, cmdName := range expectedSubcommands {
		subCmd := findCommand(cmd, cmdName)
		require.NotNil(t, subCmd, "%s subcommand should exist", cmdName)
		assert.Equal(t, cmdName, subCmd.Use, "%s command name should match", cmdName)
	}

	// Verify we have at least the expected number of custom commands
	// (Cobra adds completion and help automatically)
	commands := cmd.Commands()
	assert.GreaterOrEqual(t, len(commands), 3, "should have at least build, validate, and list commands")
}

func TestRootCmd_Help(t *testing.T) {
	cmd := NewRootCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "shellforge")
	assert.Contains(t, output, "build tool")
	assert.Contains(t, output, "Available Commands:")
	assert.Contains(t, output, "build")
	assert.Contains(t, output, "validate")
	assert.Contains(t, output, "list")
}

func TestRootCmd_Version_Flag(t *testing.T) {
	cmd := NewRootCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "0.1.0")
}

func TestRootCmd_LongDescription(t *testing.T) {
	cmd := NewRootCmd()

	long := cmd.Long

	// Verify long description contains key concepts
	expectedTerms := []string{
		"modular",
		"dependencies",
		"topological",
		"OS-specific",
		"manifest",
		"shell configuration",
	}

	lowerLong := strings.ToLower(long)
	for _, term := range expectedTerms {
		assert.Contains(t, lowerLong, strings.ToLower(term),
			"long description should mention '%s'", term)
	}
}

func TestRootCmd_SilenceUsage(t *testing.T) {
	cmd := NewRootCmd()

	// SilenceUsage should be true to avoid showing usage on errors
	assert.True(t, cmd.SilenceUsage, "SilenceUsage should be true")
}

func TestRootCmd_NoArgs(t *testing.T) {
	cmd := NewRootCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	// Running without args should show help
	assert.Contains(t, output, "Available Commands:")
}

func TestRootCmd_InvalidCommand(t *testing.T) {
	cmd := NewRootCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"invalid-command"})

	err := cmd.Execute()
	assert.Error(t, err)

	output := buf.String()
	assert.Contains(t, output, "unknown command")
}

func TestRootCmd_GlobalFlags(t *testing.T) {
	cmd := NewRootCmd()

	// Version should be set on the command
	assert.Equal(t, version, cmd.Version)

	// Verify version output works
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), version)
}

// Helper function to find a command by name
func findCommand(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

func TestRootCmd_SubcommandHelp(t *testing.T) {
	cmd := NewRootCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"build", "--help"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "build")
	assert.Contains(t, output, "manifest")
	assert.Contains(t, output, "--os")
}

func TestRootCmd_CompletionCommand(t *testing.T) {
	cmd := NewRootCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"completion", "--help"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "completion", "completion command should be available")
}

func TestRootCmd_HelpCommand(t *testing.T) {
	cmd := NewRootCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"help"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Available Commands:", "help command should show available commands")
}
