package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/filesystem"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/yamlparser"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type listFlags struct {
	manifest  string
	configDir string
	verbose   bool
	filterOS  string
}

func newListCmd() *cobra.Command {
	flags := &listFlags{}

	cmd := &cobra.Command{
		Use:   "list",
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
	cmd.Flags().StringVarP(&flags.filterOS, "filter", "f", "", "Filter modules by OS (Mac, Linux)")

	return cmd
}

func runList(cmd *cobra.Command, flags *listFlags) error {
	// Parse manifest
	fs := afero.NewOsFs()
	parser := yamlparser.New(fs)
	manifest, err := parser.Parse(flags.manifest)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Validate manifest
	if validationErrors := manifest.Validate(); len(validationErrors) > 0 {
		cmd.PrintErrln("⚠️  Manifest has validation errors:")
		for _, verr := range validationErrors {
			cmd.PrintErrf("  - %s\n", verr.Error())
		}
		cmd.PrintErrln()
	}

	// Filter modules by OS if specified
	modules := manifest.Modules
	if flags.filterOS != "" {
		var filtered []domain.Module
		for _, module := range modules {
			if module.AppliesTo(flags.filterOS) {
				filtered = append(filtered, module)
			}
		}
		modules = filtered
	}

	// Display header
	if flags.filterOS != "" {
		cmd.Printf("Modules (%d) - Filtered by OS: %s\n", len(modules), flags.filterOS)
	} else {
		cmd.Printf("Modules (%d)\n", len(modules))
	}
	cmd.Printf("Manifest: %s\n\n", flags.manifest)

	// Check if module files exist
	reader := filesystem.NewReader(fs)

	// Display modules
	for i, module := range modules {
		// Module name and OS compatibility
		osInfo := ""
		if len(module.OS) > 0 {
			osInfo = fmt.Sprintf(" [%s]", strings.Join(module.OS, ", "))
		} else {
			osInfo = " [all]"
		}

		cmd.Printf("%d. %s%s\n", i+1, module.Name, osInfo)

		// Description
		if module.Description != "" {
			cmd.Printf("   %s\n", module.Description)
		}

		// File path (verbose mode)
		if flags.verbose {
			fullPath := filepath.Join(flags.configDir, module.File)
			fileExists := reader.FileExists(fullPath)
			existsMarker := "✓"
			if !fileExists {
				existsMarker = "✗"
			}
			cmd.Printf("   File: %s %s\n", module.File, existsMarker)
		}

		// Dependencies
		if len(module.Requires) > 0 {
			if flags.verbose {
				cmd.Printf("   Requires: %s\n", strings.Join(module.Requires, ", "))
			} else {
				cmd.Printf("   → %s\n", strings.Join(module.Requires, ", "))
			}
		}

		// Spacing between modules
		if i < len(modules)-1 {
			cmd.Println()
		}
	}

	return nil
}
