package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.3.0"
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shellforge",
		Short: "Build tool for modular shell configurations",
		Long: `Shellforge is a build tool that assembles modular shell configurations
with dependency resolution and OS-specific filtering.

It reads a manifest file defining shell modules and their dependencies,
resolves the load order using topological sorting, and generates a
single shell configuration file.`,
		Version:      version,
		SilenceUsage: true,
	}

	// Add subcommands
	cmd.AddCommand(newBuildCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newBackupCmd())
	cmd.AddCommand(newRestoreCmd())
	cmd.AddCommand(newCleanupCmd())
	cmd.AddCommand(newTemplateCmd())

	return cmd
}

// Execute runs the root command
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
