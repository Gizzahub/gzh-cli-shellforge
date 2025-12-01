// Package cmd provides public API for embedding shellforge commands.
package cmd

import (
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli"
	"github.com/spf13/cobra"
)

// NewRootCmd returns the root shellforge command for embedding in other CLIs.
// This allows other projects to integrate shellforge as a subcommand.
//
// Example usage:
//
//	shellforgeCmd := cmd.NewRootCmd()
//	shellforgeCmd.Use = "shellforge"  // Customize the command name
//	rootCmd.AddCommand(shellforgeCmd)
func NewRootCmd() *cobra.Command {
	return cli.NewRootCmd()
}
