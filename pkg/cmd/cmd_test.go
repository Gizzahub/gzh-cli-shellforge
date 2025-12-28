// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestNewRootCmd tests the root command factory.
func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	if cmd == nil {
		t.Fatal("NewRootCmd() returned nil")
	}

	if !isCobraCommand(cmd) {
		t.Error("NewRootCmd() did not return a valid Cobra command")
	}
}

// TestRootCmdStructure tests basic structure of root command.
func TestRootCmdStructure(t *testing.T) {
	cmd := NewRootCmd()

	if cmd == nil {
		t.Fatal("NewRootCmd() returned nil")
	}

	// Verify it's a valid command with a Use field
	if cmd.Use == "" {
		t.Error("Root command should have a Use field")
	}

	// Verify it has Short description or commands
	// (It's acceptable to be a command group without Run/RunE)
	if cmd.Short == "" && len(cmd.Commands()) == 0 {
		t.Error("Root command should have Short description or subcommands")
	}
}

// isCobraCommand is a helper to verify something is a Cobra command.
func isCobraCommand(cmd *cobra.Command) bool {
	return cmd != nil && (cmd.Use != "" || cmd.Short != "")
}

// TestNewRootCmdIsEmbeddable tests that the command can be embedded in another CLI.
func TestNewRootCmdIsEmbeddable(t *testing.T) {
	rootCmd := NewRootCmd()

	if rootCmd == nil {
		t.Fatal("NewRootCmd() returned nil")
	}

	// Create a parent command to test embedding
	parentCmd := &cobra.Command{
		Use:   "parent",
		Short: "Parent command",
	}

	// Add the root command as a subcommand
	rootCmd.Use = "shellforge"
	parentCmd.AddCommand(rootCmd)

	// Verify the command was added
	found := false
	for _, cmd := range parentCmd.Commands() {
		if cmd.Use == "shellforge" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Root command could not be embedded as subcommand")
	}
}
