package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	clierrors "github.com/gizzahub/gzh-cli-shellforge/internal/cli/errors"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/factory"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/helpers"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/rcparser"
)

type migrateFlags struct {
	outputDir    string
	manifestPath string
	dryRun       bool
	verbose      bool
}

func newMigrateCmd() *cobra.Command {
	flags := &migrateFlags{}

	cmd := &cobra.Command{
		Use:   "migrate <rc-file>",
		Short: "Migrate monolithic RC file to modular structure",
		Long: `Migrate converts a monolithic shell configuration file (.zshrc, .bashrc)
into a modular structure with automatic section detection and categorization.

The command analyzes your RC file, detects sections using header patterns,
and generates individual module files with proper categorization:
  - init.d/     PATH and early initialization
  - rc_pre.d/   Tool initialization (nvm, rbenv, etc.)
  - rc_post.d/  Aliases, functions, and customizations

A manifest.yaml file is also generated with detected dependencies and OS support.`,
		Example: `  # Analyze migration (dry-run)
  gz-shellforge migrate ~/.zshrc --dry-run

  # Migrate to modular structure
  gz-shellforge migrate ~/.zshrc --output-dir modules --manifest manifest.yaml

  # Verbose output
  gz-shellforge migrate ~/.bashrc --output-dir modules -v`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rcFilePath := args[0]
			return runMigrate(rcFilePath, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.outputDir, "output-dir", "o", "modules", "Output directory for module files")
	cmd.Flags().StringVarP(&flags.manifestPath, "manifest", "m", "manifest.yaml", "Manifest file path")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Analyze only, do not create files")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed output")

	return cmd
}

func runMigrate(rcFilePath string, flags *migrateFlags) error {
	// Expand home directory
	var err error
	rcFilePath, err = helpers.ExpandHomePath(rcFilePath)
	if err != nil {
		return clierrors.InvalidPath("RC file", err)
	}

	// Check if RC file exists
	if _, err := os.Stat(rcFilePath); os.IsNotExist(err) {
		return clierrors.FileNotFound(rcFilePath)
	}

	// Initialize services
	services := factory.NewServices()
	parser := rcparser.New(services.Fs)
	service := app.NewMigrationService(parser, services.Reader, services.Writer)

	// Dry-run mode: analyze only
	if flags.dryRun {
		result, err := service.Analyze(rcFilePath)
		if err != nil {
			return clierrors.WrapError("migration analysis", err)
		}

		printAnalysisResult(result, flags.verbose)
		return nil
	}

	// Full migration
	fmt.Printf("Migrating %s...\n", rcFilePath)

	result, err := service.Migrate(rcFilePath, flags.outputDir, flags.manifestPath)
	if err != nil {
		return clierrors.WrapError("migration", err)
	}

	printMigrationResult(result, flags.verbose)
	return nil
}

func printAnalysisResult(result *app.MigrateResult, verbose bool) {
	fmt.Println("=== Migration Analysis ===")
	fmt.Printf("Source: %s\n", result.SourceFile)
	fmt.Printf("Detected sections: %d\n", len(result.Sections))
	fmt.Printf("Modules to create: %d\n\n", result.ModulesCreated)

	if len(result.Sections) == 0 {
		fmt.Println("⚠️  No sections detected. File may be unsegmented.")
		fmt.Println("   Tip: Add section headers like '# --- Section Name ---'")
		return
	}

	fmt.Println("Sections:")
	for i, section := range result.Sections {
		fmt.Printf("  %d. %s\n", i+1, section.Name)
		fmt.Printf("     Category: %s/\n", section.Category)
		fmt.Printf("     Lines: %d-%d\n", section.LineStart, section.LineEnd)

		if verbose {
			fmt.Printf("     Content preview:\n")
			lines := splitLines(section.Content, 3)
			for _, line := range lines {
				fmt.Printf("       %s\n", line)
			}
		}
		fmt.Println()
	}

	if len(result.Warnings) > 0 {
		fmt.Println("⚠️  Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
		fmt.Println()
	}

	fmt.Println("Run without --dry-run to create module files.")
}

func printMigrationResult(result *app.MigrateResult, verbose bool) {
	fmt.Println("=== Migration Complete ===")
	fmt.Printf("✓ Created %d module files\n", result.ModulesCreated)
	fmt.Printf("✓ Generated manifest: %s\n\n", result.ManifestPath)

	if verbose {
		fmt.Println("Module files created:")
		for _, path := range result.ModuleFilesPaths {
			fmt.Printf("  ✓ %s\n", path)
		}
		fmt.Println()
	}

	if len(result.Warnings) > 0 {
		fmt.Println("⚠️  Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
		fmt.Println()
	}

	fmt.Println("Next steps:")
	fmt.Println("  1. Review generated module files")
	fmt.Println("  2. Edit manifest.yaml if needed")
	fmt.Println("  3. Test with: gz-shellforge build --manifest manifest.yaml --os $(uname -s) --dry-run")
	fmt.Println("  4. Deploy: gz-shellforge build --manifest manifest.yaml --os $(uname -s) --output ~/.zshrc")
}

func splitLines(content string, maxLines int) []string {
	lines := []string{}
	current := ""
	count := 0

	for _, c := range content {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
			count++
			if count >= maxLines {
				if len(content) > len(current) {
					lines = append(lines, "...")
				}
				break
			}
		} else {
			current += string(c)
		}
	}

	if current != "" && count < maxLines {
		lines = append(lines, current)
	}

	return lines
}
