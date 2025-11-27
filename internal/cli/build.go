package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/filesystem"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/yamlparser"
)

type buildFlags struct {
	configDir string
	manifest  string
	output    string
	targetOS  string
	dryRun    bool
	verbose   bool
}

func newBuildCmd() *cobra.Command {
	flags := &buildFlags{}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build shell configuration from modules",
		Long: `Build generates a shell configuration file from modular components.

It reads the manifest file, resolves module dependencies using topological
sorting, filters modules by target OS, and concatenates the module files
in the correct order.`,
		Example: `  # Build for macOS with default manifest
  shellforge build --os Mac

  # Build with custom manifest and output
  shellforge build --manifest custom.yaml --output ~/.zshrc --os Mac

  # Dry run to preview output
  shellforge build --os Linux --dry-run

  # Verbose mode for debugging
  shellforge build --os Mac --verbose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(flags)
		},
	}

	// Define flags
	cmd.Flags().StringVarP(&flags.configDir, "config-dir", "c", "modules", "Directory containing module files")
	cmd.Flags().StringVarP(&flags.manifest, "manifest", "m", "manifest.yaml", "Path to manifest file")
	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (stdout if not specified)")
	cmd.Flags().StringVar(&flags.targetOS, "os", "", "Target operating system (Mac, Linux, etc.)")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Preview output without writing file")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed output")

	// Mark required flags
	cmd.MarkFlagRequired("os")

	return cmd
}

func runBuild(flags *buildFlags) error {
	// Validate flags
	if flags.output == "" && !flags.dryRun {
		return fmt.Errorf("--output is required unless --dry-run is specified")
	}

	if flags.verbose {
		fmt.Printf("Building shell configuration...\n")
		fmt.Printf("  Manifest: %s\n", flags.manifest)
		fmt.Printf("  Config dir: %s\n", flags.configDir)
		fmt.Printf("  Target OS: %s\n", flags.targetOS)
		if flags.dryRun {
			fmt.Printf("  Mode: Dry run (no file will be written)\n")
		} else {
			fmt.Printf("  Output: %s\n", flags.output)
		}
		fmt.Println()
	}

	// Create filesystem abstraction
	fs := afero.NewOsFs()

	// Create services
	parser := yamlparser.New(fs)
	reader := filesystem.NewReader(fs)
	writer := filesystem.NewWriter(fs)
	builder := app.NewBuilderService(parser, reader, writer)

	// Build options
	opts := app.BuildOptions{
		ConfigDir: flags.configDir,
		Manifest:  flags.manifest,
		Output:    flags.output,
		OS:        flags.targetOS,
		DryRun:    flags.dryRun,
		Verbose:   flags.verbose,
	}

	// Execute build
	result, err := builder.Build(opts)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Display results
	if flags.verbose {
		fmt.Printf("✓ Build completed successfully\n")
		fmt.Printf("  Modules loaded: %d\n", result.ModuleCount)
		fmt.Printf("  Load order: %v\n", result.ModuleNames)
		fmt.Printf("  Generated at: %s\n", result.GeneratedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	// Output handling
	if flags.dryRun {
		if flags.verbose {
			fmt.Println("--- Generated Configuration (Dry Run) ---")
		}
		fmt.Println(result.Output)
	} else {
		// Expand home directory in output path
		outputPath := flags.output
		if len(outputPath) > 0 && outputPath[0] == '~' {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to expand home directory: %w", err)
			}
			outputPath = filepath.Join(home, outputPath[1:])
		}

		fmt.Printf("✓ Configuration written to: %s\n", outputPath)
		if flags.verbose {
			fmt.Printf("  Size: %d bytes\n", len(result.Output))
		}
	}

	return nil
}
