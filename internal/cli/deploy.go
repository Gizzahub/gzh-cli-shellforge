package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	clierrors "github.com/gizzahub/gzh-cli-shellforge/internal/cli/errors"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/factory"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/helpers"
)

type deployFlags struct {
	buildDir string
	dryRun   bool
	backup   bool
	verbose  bool
}

func newDeployCmd() *cobra.Command {
	flags := &deployFlags{}

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy built configuration files to their actual paths",
		Long: `Deploy copies the built configuration files from the build directory
to their actual destination paths (e.g., ~/.zshrc).

The deploy process:
  1. Reads files from the build directory (default: ./build)
  2. Determines destination paths based on file names
  3. Optionally backs up existing files
  4. Copies files to their destinations

Typical workflow:
  1. Build: gz-shellforge build           # Generates files in ./build/
  2. Review: ls -la ./build/              # Check generated files
  3. Deploy: gz-shellforge deploy --backup # Deploy with backup`,
		Example: `  # Deploy from default build directory
  gz-shellforge deploy

  # Preview without deploying
  gz-shellforge deploy --dry-run

  # Backup existing files before deploying
  gz-shellforge deploy --backup

  # Deploy from custom build directory
  gz-shellforge deploy --build-dir ~/staging

  # Combined workflow
  gz-shellforge build && gz-shellforge deploy --backup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(flags)
		},
	}

	cmd.Flags().StringVarP(&flags.buildDir, "build-dir", "d", "./build", "Build directory containing files to deploy")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Preview deployment without making changes")
	cmd.Flags().BoolVar(&flags.backup, "backup", false, "Backup existing files before overwriting")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed output")

	return cmd
}

func runDeploy(flags *deployFlags) error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}

	// Expand build directory path
	buildDir := flags.buildDir
	if buildDir != "" {
		expanded, err := helpers.ExpandHomePath(buildDir)
		if err != nil {
			return clierrors.InvalidPath("build-dir", err)
		}
		buildDir = expanded
	}

	// Verbose output
	if flags.verbose {
		printDeployHeader(flags, buildDir)
	}

	// Create services
	services := factory.NewServices()
	deployer := services.NewDeployer()

	// Deploy options
	opts := app.DeployOptions{
		BuildDir:     buildDir,
		DryRun:       flags.dryRun,
		CreateBackup: flags.backup,
		Verbose:      flags.verbose,
		HomeDir:      homeDir,
	}

	// Execute deploy
	result, err := deployer.Deploy(opts)
	if err != nil {
		return clierrors.WrapError("deploy", err)
	}

	// Display results
	printDeployResult(flags, result)

	return nil
}

func printDeployHeader(flags *deployFlags, buildDir string) {
	fmt.Printf("Deploying shell configuration...\n")
	fmt.Printf("  Build directory: %s\n", buildDir)
	if flags.dryRun {
		fmt.Printf("  Dry run: yes (no files will be written)\n")
	}
	if flags.backup {
		fmt.Printf("  Backup: enabled\n")
	}
	fmt.Println()
}

func printDeployResult(flags *deployFlags, result *app.DeployResult) {
	if flags.dryRun {
		fmt.Printf("✓ Deployment preview (dry run)\n")
		fmt.Printf("  Files to deploy: %d\n\n", result.TotalFiles)

		for _, file := range result.DeployedFiles {
			fmt.Printf("  • %s → %s\n", file.SourcePath, file.DestPath)
		}
		fmt.Println()
		fmt.Printf("Run without --dry-run to deploy these files.\n")
		return
	}

	if result.ErrorCount > 0 {
		fmt.Printf("⚠ Deployment completed with errors\n")
	} else {
		fmt.Printf("✓ Deployment completed successfully\n")
	}

	fmt.Printf("  Deployed: %d/%d files\n", result.DeployedCount, result.TotalFiles)

	if result.ErrorCount > 0 {
		fmt.Printf("  Errors: %d\n", result.ErrorCount)
	}

	if flags.verbose || result.TotalFiles <= 5 {
		fmt.Println()
		for _, file := range result.DeployedFiles {
			status := "✓"
			if file.Error != nil {
				status = "✗"
			}
			fmt.Printf("  %s %s → %s\n", status, file.SourcePath, file.DestPath)
			if file.BackupPath != "" {
				fmt.Printf("    Backup: %s\n", file.BackupPath)
			}
			if file.Error != nil {
				fmt.Printf("    Error: %v\n", file.Error)
			}
		}
	}

	if len(result.BackupPaths) > 0 && !flags.verbose {
		fmt.Printf("\n  Backups created: %d\n", len(result.BackupPaths))
	}
}
