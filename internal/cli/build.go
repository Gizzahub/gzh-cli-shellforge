package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	clierrors "github.com/gizzahub/gzh-cli-shellforge/internal/cli/errors"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/factory"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/helpers"
)

type buildFlags struct {
	configDir string
	manifest  string
	targetOS  string
	dryRun    bool
	verbose   bool

	// Multi-target options (v2)
	outputDir string
	shell     string
	targets   []string
	backup    bool

	// Legacy single-output mode
	singleOutput string
}

func newBuildCmd() *cobra.Command {
	flags := &buildFlags{}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build shell configuration from modules",
		Long: `Build generates shell configuration files from modular components.

Multi-Target Mode (v2):
  By default, modules are grouped by their 'target' field (zshrc, zprofile, etc.)
  and written to separate RC files in the output directory.

Legacy Mode:
  Use --single-output to write all modules to a single file (v1 behavior).

The build process:
  1. Reads the manifest file
  2. Resolves module dependencies using topological sorting
  3. Filters modules by target OS
  4. Groups modules by target RC file (v2) or combines them (legacy)
  5. Sorts modules by priority within each target
  6. Writes the output file(s)`,
		Example: `  # Build to default ./build/ directory (OS auto-detected)
  gz-shellforge build

  # Build with explicit OS and output directory
  gz-shellforge build --os Mac --output-dir ~/staging

  # Dry run to preview output (OS auto-detected)
  gz-shellforge build --dry-run

  # Build for bash on Linux
  gz-shellforge build --os Linux --shell bash --output-dir ~

  # Build only specific targets
  gz-shellforge build --os Mac --target zshrc --target zprofile

  # Legacy: single file output
  gz-shellforge build --os Mac --single-output ~/.zshrc

  # With backup of existing files
  gz-shellforge build --output-dir ~ --backup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(flags)
		},
	}

	// OS flag (auto-detected if not specified)
	cmd.Flags().StringVar(&flags.targetOS, "os", "", "Target operating system (auto-detected if not specified)")

	// Multi-target options (v2)
	cmd.Flags().StringVarP(&flags.outputDir, "output-dir", "d", "", "Output directory for RC files (default: ./build)")
	cmd.Flags().StringVarP(&flags.shell, "shell", "s", "", "Shell type (zsh, bash, fish)")
	cmd.Flags().StringArrayVarP(&flags.targets, "target", "t", nil, "Specific targets to build (can be repeated)")
	cmd.Flags().BoolVar(&flags.backup, "backup", false, "Create backup of existing files")

	// Legacy single-output mode
	cmd.Flags().StringVarP(&flags.singleOutput, "single-output", "o", "", "Single output file (legacy mode)")

	// Common options
	cmd.Flags().StringVarP(&flags.configDir, "config-dir", "c", "modules", "Directory containing module files")
	cmd.Flags().StringVarP(&flags.manifest, "manifest", "m", "manifest.yaml", "Path to manifest file")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Preview output without writing files")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed output")

	// Deprecate old --output flag
	cmd.Flags().String("output", "", "Deprecated: use --single-output instead")
	_ = cmd.Flags().MarkDeprecated("output", "use --single-output instead")

	return cmd
}

func runBuild(flags *buildFlags) error {
	// Auto-detect OS if not specified
	if flags.targetOS == "" {
		flags.targetOS = helpers.DetectOS()
		if flags.verbose {
			fmt.Printf("Auto-detected OS: %s\n", flags.targetOS)
		}
	}

	// Determine build mode
	isLegacyMode := flags.singleOutput != ""
	hasOutputDir := flags.outputDir != ""

	// Set default output directory if not specified
	if !flags.dryRun && !isLegacyMode && !hasOutputDir {
		flags.outputDir = "./build"
		hasOutputDir = true
		if flags.verbose {
			fmt.Printf("Using default output directory: %s\n", flags.outputDir)
		}
	}

	if isLegacyMode && hasOutputDir {
		return fmt.Errorf("cannot use both --single-output and --output-dir; choose one mode")
	}

	// Get home directory for path expansion
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}

	// Verbose output
	if flags.verbose {
		printBuildHeader(flags, isLegacyMode)
	}

	// Create services
	services := factory.NewServices()
	builder := services.NewBuilder()

	// Build options
	opts := app.BuildOptions{
		ConfigDir:    flags.configDir,
		Manifest:     flags.manifest,
		OS:           flags.targetOS,
		DryRun:       flags.dryRun,
		Verbose:      flags.verbose,
		OutputDir:    flags.outputDir,
		Shell:        flags.shell,
		Targets:      flags.targets,
		CreateBackup: flags.backup,
		HomeDir:      homeDir,
		Output:       flags.singleOutput, // Legacy mode
	}

	// Expand paths
	if opts.OutputDir != "" {
		expanded, err := helpers.ExpandHomePath(opts.OutputDir)
		if err != nil {
			return clierrors.InvalidPath("output-dir", err)
		}
		opts.OutputDir = expanded
	}
	if opts.Output != "" {
		expanded, err := helpers.ExpandHomePath(opts.Output)
		if err != nil {
			return clierrors.InvalidPath("single-output", err)
		}
		opts.Output = expanded
	}

	// Execute build
	result, err := builder.Build(opts)
	if err != nil {
		return clierrors.WrapError("build", err)
	}

	// Display results
	if flags.dryRun {
		printDryRunResult(flags, result)
	} else {
		printBuildResult(flags, result, isLegacyMode)
	}

	return nil
}

func printBuildHeader(flags *buildFlags, isLegacyMode bool) {
	fmt.Printf("Building shell configuration...\n")
	fmt.Printf("  Manifest: %s\n", flags.manifest)
	fmt.Printf("  Config dir: %s\n", flags.configDir)
	fmt.Printf("  Target OS: %s\n", flags.targetOS)
	if flags.shell != "" {
		fmt.Printf("  Shell: %s\n", flags.shell)
	}
	if isLegacyMode {
		fmt.Printf("  Mode: Legacy (single output)\n")
		fmt.Printf("  Output: %s\n", flags.singleOutput)
	} else {
		fmt.Printf("  Mode: Multi-target\n")
		if flags.outputDir != "" {
			fmt.Printf("  Output dir: %s\n", flags.outputDir)
		}
		if len(flags.targets) > 0 {
			fmt.Printf("  Targets: %v\n", flags.targets)
		}
	}
	if flags.dryRun {
		fmt.Printf("  Dry run: yes (no files will be written)\n")
	}
	if flags.backup {
		fmt.Printf("  Backup: enabled\n")
	}
	fmt.Println()
}

func printDryRunResult(flags *buildFlags, result *app.BuildResult) {
	if flags.verbose {
		fmt.Printf("✓ Build preview completed\n")
		fmt.Printf("  Shell: %s\n", result.ShellType)
		fmt.Printf("  OS: %s\n", result.TargetOS)
		fmt.Printf("  Total modules: %d\n", result.TotalModuleCount)
		fmt.Printf("  Targets: %d\n", len(result.Targets))
		fmt.Println()
	}

	if len(result.Targets) == 1 {
		// Single target - just show content
		if flags.verbose {
			fmt.Println("--- Generated Configuration (Dry Run) ---")
		}
		fmt.Println(result.Targets[0].Content)
	} else {
		// Multiple targets - show each with header
		for _, target := range result.Targets {
			fmt.Printf("\n=== %s (%d modules) ===\n", target.Target, target.ModuleCount)
			if flags.verbose {
				fmt.Printf("    File: %s\n", target.FilePath)
				fmt.Printf("    Modules: %v\n", target.ModuleNames)
			}
			fmt.Println(target.Content)
		}
	}
}

func printBuildResult(flags *buildFlags, result *app.BuildResult, isLegacyMode bool) {
	if flags.verbose {
		fmt.Printf("✓ Build completed successfully\n")
		fmt.Printf("  Shell: %s\n", result.ShellType)
		fmt.Printf("  OS: %s\n", result.TargetOS)
		fmt.Printf("  Total modules: %d\n", result.TotalModuleCount)
		fmt.Printf("  Generated at: %s\n", result.GeneratedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	if isLegacyMode {
		// Legacy mode - single file
		target := result.Targets[0]
		fmt.Printf("✓ Configuration written to: %s\n", target.FilePath)
		if flags.verbose {
			fmt.Printf("  Modules: %d (%s)\n", target.ModuleCount, strings.Join(target.ModuleNames, ", "))
			fmt.Printf("  Size: %d bytes\n", len(target.Content))
		}
	} else {
		// Multi-target mode
		fmt.Printf("✓ Generated %d RC files:\n", len(result.Targets))
		for _, target := range result.Targets {
			fmt.Printf("  • %s → %s (%d modules)\n", target.Target, target.FilePath, target.ModuleCount)
			if target.BackupPath != "" {
				fmt.Printf("    Backup: %s\n", target.BackupPath)
			}
			if flags.verbose {
				fmt.Printf("    Modules: %s\n", strings.Join(target.ModuleNames, ", "))
			}
		}
	}
}
