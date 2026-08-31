// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	clierrors "github.com/gizzahub/gzh-cli-shellforge/internal/cli/errors"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/factory"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

const listCommandName = "list"

type listFlags struct {
	manifest  string
	configDir string
	verbose   bool
	filterOS  string
}

func newListCmd() *cobra.Command {
	flags := &listFlags{}

	cmd := &cobra.Command{
		Use:   listCommandName,
		Short: "List all modules from manifest",
		Long: `List all modules defined in the manifest file with their metadata.

This command reads the manifest and displays module information including:
- Module name and description
- File path
- Dependencies
- OS compatibility

Use --filter to show only modules for a specific OS.
Use --verbose to show detailed information including full file paths.`,
		Example: `  # List all modules
  shellforge list

  # List with custom manifest
  shellforge list --manifest custom.yaml --config-dir modules

  # List only Mac-compatible modules
  shellforge list --filter Mac

  # List with verbose output
  shellforge list --verbose

  # List Linux modules with verbose output
  shellforge list --filter Linux --verbose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.manifest, "manifest", "m", "manifest.yaml", "Path to manifest file")
	cmd.Flags().StringVarP(&flags.configDir, "config-dir", "c", "modules", "Directory containing module files")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed output")
	cmd.Flags().StringVarP(&flags.filterOS, "filter", "F", "", "Filter modules by OS (Mac, Linux)")

	return cmd
}

func runList(cmd *cobra.Command, flags *listFlags) error {
	// Create services and parse manifest
	services := factory.NewServices()
	manifest, err := services.Parser.Parse(flags.manifest)
	if err != nil {
		return clierrors.WrapError("manifest parsing", err)
	}

	printManifestValidationErrors(cmd, manifest.Validate())
	modules := filterModules(manifest.Modules, flags.filterOS)
	printListHeader(cmd, flags, len(modules))

	reader := services.Reader
	for i, module := range modules {
		printModule(cmd, reader, module, i, len(modules), flags)
	}

	return nil
}

func printManifestValidationErrors(cmd *cobra.Command, validationErrors []error) {
	if len(validationErrors) == 0 {
		return
	}
	cmd.PrintErrln("⚠️  Manifest has validation errors:")
	for _, verr := range validationErrors {
		cmd.PrintErrf("  - %s\n", verr.Error())
	}
	cmd.PrintErrln()
}

func filterModules(modules []domain.Module, filterOS string) []domain.Module {
	if filterOS == "" {
		return modules
	}
	filtered := make([]domain.Module, 0, len(modules))
	for _, module := range modules {
		if module.AppliesTo(filterOS) {
			filtered = append(filtered, module)
		}
	}
	return filtered
}

func printListHeader(cmd *cobra.Command, flags *listFlags, moduleCount int) {
	if flags.filterOS != "" {
		cmd.Printf("Modules (%d) - Filtered by OS: %s\n", moduleCount, flags.filterOS)
	} else {
		cmd.Printf("Modules (%d)\n", moduleCount)
	}
	cmd.Printf("Manifest: %s\n\n", flags.manifest)
}

func printModule(cmd *cobra.Command, reader interface{ FileExists(string) bool }, module domain.Module, index, total int, flags *listFlags) {
	osInfo := " [all]"
	if len(module.OS) > 0 {
		osInfo = fmt.Sprintf(" [%s]", strings.Join(module.OS, ", "))
	}
	cmd.Printf("%d. %s%s\n", index+1, module.Name, osInfo)
	if module.Description != "" {
		cmd.Printf("   %s\n", module.Description)
	}
	if flags.verbose {
		existsMarker := "✓"
		if !reader.FileExists(filepath.Join(flags.configDir, module.File)) {
			existsMarker = "✗"
		}
		cmd.Printf("   File: %s %s\n", module.File, existsMarker)
	}
	if len(module.Requires) > 0 {
		prefix := "   →"
		if flags.verbose {
			prefix = "   Requires:"
		}
		cmd.Printf("%s %s\n", prefix, strings.Join(module.Requires, ", "))
	}
	if index < total-1 {
		cmd.Println()
	}
}
