package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestProfilesCommand_Structure(t *testing.T) {
	cmd := newProfilesCmd()

	assert.Equal(t, "profiles", cmd.Use)
	assert.Contains(t, cmd.Short, "shell profile")

	// Check subcommands exist
	subCmds := cmd.Commands()
	assert.GreaterOrEqual(t, len(subCmds), 3, "Should have at least 3 subcommands")

	// Find subcommands by name
	var listCmd, showCmd, checkCmd bool
	for _, sub := range subCmds {
		switch sub.Name() {
		case "list":
			listCmd = true
		case "show":
			showCmd = true
		case "check":
			checkCmd = true
		}
	}

	assert.True(t, listCmd, "Should have 'list' subcommand")
	assert.True(t, showCmd, "Should have 'show' subcommand")
	assert.True(t, checkCmd, "Should have 'check' subcommand")
}

func TestProfilesListCommand_Structure(t *testing.T) {
	cmd := newProfilesListCmd()

	assert.Equal(t, "list [category]", cmd.Use)
	assert.Contains(t, cmd.Short, "List")

	// Check flags
	flags := cmd.Flags()
	assert.NotNil(t, flags.Lookup("data-dir"), "Should have --data-dir flag")
	assert.NotNil(t, flags.Lookup("verbose"), "Should have --verbose flag")

	// Check accepts 0 or 1 arguments
	err := cmd.Args(cmd, []string{})
	assert.NoError(t, err, "Should accept no arguments")

	err = cmd.Args(cmd, []string{"distributions"})
	assert.NoError(t, err, "Should accept one argument")

	err = cmd.Args(cmd, []string{"a", "b"})
	assert.Error(t, err, "Should reject two arguments")
}

func TestProfilesShowCommand_Structure(t *testing.T) {
	cmd := newProfilesShowCmd()

	assert.Equal(t, "show <type> <name>", cmd.Use)
	assert.Contains(t, cmd.Short, "Show")

	// Check flags
	flags := cmd.Flags()
	assert.NotNil(t, flags.Lookup("data-dir"), "Should have --data-dir flag")
	assert.NotNil(t, flags.Lookup("verbose"), "Should have --verbose flag")

	// Check requires exactly 2 arguments
	err := cmd.Args(cmd, []string{"os", "mac"})
	assert.NoError(t, err, "Should accept two arguments")

	err = cmd.Args(cmd, []string{"os"})
	assert.Error(t, err, "Should reject one argument")

	err = cmd.Args(cmd, []string{})
	assert.Error(t, err, "Should reject no arguments")
}

func TestProfilesCheckCommand_Structure(t *testing.T) {
	cmd := newProfilesCheckCmd()

	assert.Equal(t, "check <context>", cmd.Use)
	assert.Contains(t, cmd.Short, "Check")

	// Check flags
	flags := cmd.Flags()
	assert.NotNil(t, flags.Lookup("data-dir"), "Should have --data-dir flag")

	// Check requires exactly 1 argument
	err := cmd.Args(cmd, []string{"cron"})
	assert.NoError(t, err, "Should accept one argument")

	err = cmd.Args(cmd, []string{})
	assert.Error(t, err, "Should reject no arguments")
}

func TestProfilesCommand_Help(t *testing.T) {
	cmd := newProfilesCmd()

	// Capture help output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "profiles", "Help should mention profiles")
	assert.Contains(t, output, "list", "Help should mention list subcommand")
	assert.Contains(t, output, "show", "Help should mention show subcommand")
	assert.Contains(t, output, "check", "Help should mention check subcommand")
}

func TestProfilesListCommand_Help(t *testing.T) {
	cmd := newProfilesListCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "distributions", "Help should mention distributions")
	assert.Contains(t, output, "managers", "Help should mention managers")
	assert.Contains(t, output, "desktops", "Help should mention desktops")
	assert.Contains(t, output, "modes", "Help should mention modes")
	assert.Contains(t, output, "multiplexers", "Help should mention multiplexers")
}

func TestProfilesShowCommand_Help(t *testing.T) {
	cmd := newProfilesShowCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "os", "Help should mention os type")
	assert.Contains(t, output, "manager", "Help should mention manager type")
	assert.Contains(t, output, "desktop", "Help should mention desktop type")
	assert.Contains(t, output, "mode", "Help should mention mode type")
	assert.Contains(t, output, "multiplexer", "Help should mention multiplexer type")
}

func TestProfilesCheckCommand_Help(t *testing.T) {
	cmd := newProfilesCheckCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "cron", "Help should mention cron context")
	assert.Contains(t, output, "docker-exec", "Help should mention docker-exec context")
	assert.Contains(t, output, "github-actions", "Help should mention github-actions context")
}

func TestProfilesCommand_Examples(t *testing.T) {
	tests := []struct {
		name   string
		cmd    func() *cobra.Command
		wantEx []string
	}{
		{
			name:   "list command examples",
			cmd:    newProfilesListCmd,
			wantEx: []string{"profiles list", "distributions", "managers"},
		},
		{
			name:   "show command examples",
			cmd:    newProfilesShowCmd,
			wantEx: []string{"profiles show", "os mac", "manager rbenv", "mode login"},
		},
		{
			name:   "check command examples",
			cmd:    newProfilesCheckCmd,
			wantEx: []string{"profiles check", "cron", "github-actions"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd()
			for _, ex := range tt.wantEx {
				assert.Contains(t, cmd.Example, ex, "Example should contain: %s", ex)
			}
		})
	}
}

func TestProfilesCommand_IntegrationWithRoot(t *testing.T) {
	rootCmd := NewRootCmd()

	// Find profiles command
	var profilesCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "profiles" {
			profilesCmd = cmd
			break
		}
	}

	assert.NotNil(t, profilesCmd, "Root command should have profiles subcommand")
	assert.Equal(t, "profiles", profilesCmd.Name())
}

func TestProfilesCommand_FlagDefaults(t *testing.T) {
	tests := []struct {
		name     string
		cmd      func() *cobra.Command
		flagName string
		wantVal  string
	}{
		{
			name:     "list data-dir default",
			cmd:      newProfilesListCmd,
			flagName: "data-dir",
			wantVal:  "",
		},
		{
			name:     "show data-dir default",
			cmd:      newProfilesShowCmd,
			flagName: "data-dir",
			wantVal:  "",
		},
		{
			name:     "check data-dir default",
			cmd:      newProfilesCheckCmd,
			flagName: "data-dir",
			wantVal:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd()
			flag := cmd.Flags().Lookup(tt.flagName)
			assert.NotNil(t, flag)
			assert.Equal(t, tt.wantVal, flag.DefValue)
		})
	}
}

func TestProfilesCommand_VerboseFlag(t *testing.T) {
	tests := []struct {
		name string
		cmd  func() *cobra.Command
	}{
		{"list verbose", newProfilesListCmd},
		{"show verbose", newProfilesShowCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd()
			flag := cmd.Flags().Lookup("verbose")
			assert.NotNil(t, flag, "Should have verbose flag")
			assert.Equal(t, "v", flag.Shorthand, "Verbose shorthand should be -v")
		})
	}
}

// Helper to check command descriptions are well-formed
func TestProfilesCommand_Descriptions(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"profiles", newProfilesCmd()},
		{"list", newProfilesListCmd()},
		{"show", newProfilesShowCmd()},
		{"check", newProfilesCheckCmd()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.cmd.Short, "Short description should not be empty")
			assert.NotEmpty(t, tt.cmd.Long, "Long description should not be empty")
			assert.Greater(t, len(tt.cmd.Long), len(tt.cmd.Short), "Long description should be longer than short")
			assert.True(t, strings.HasPrefix(tt.cmd.Short, strings.ToUpper(tt.cmd.Short[:1])), "Short description should start with capital letter")
		})
	}
}
