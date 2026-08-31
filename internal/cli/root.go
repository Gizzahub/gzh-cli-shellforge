// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

// Package cli implements the Shellforge command-line interface and command handlers.
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var version = "0.5.1"

const shellforgeCommandName = "shellforge"

func mustMarkFlagRequired(cmd *cobra.Command, names ...string) {
	for _, name := range names {
		if err := cmd.MarkFlagRequired(name); err != nil {
			panic(fmt.Sprintf("mark flag %q required: %v", name, err))
		}
	}
}

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   shellforgeCommandName,
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
	cmd.AddCommand(newPrepareCmd())
	cmd.AddCommand(newBuildCmd())
	cmd.AddCommand(newDeployCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newBackupCmd())
	cmd.AddCommand(newRestoreCmd())
	cmd.AddCommand(newCleanupCmd())
	cmd.AddCommand(newTemplateCmd())
	cmd.AddCommand(newMigrateCmd())
	cmd.AddCommand(newDiffCmd())
	cmd.AddCommand(newProfilesCmd())

	return cmd
}

// Execute runs the root command.
func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := NewRootCmd().ExecuteContext(ctx)
	stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
