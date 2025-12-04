package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/factory"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/helpers"
)

type restoreFlags struct {
	file      string
	snapshot  string
	backupDir string
	noGit     bool
	dryRun    bool
	verbose   bool
}

func newRestoreCmd() *cobra.Command {
	flags := &restoreFlags{}

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore a shell configuration file from a backup snapshot",
		Long: `Restore restores a previously backed up shell configuration file from a snapshot.

The restore operation will:
1. Create a safety backup of the current file (if git is enabled)
2. Restore the file from the specified snapshot
3. Commit the restore operation to git (if enabled)

Use --dry-run to preview the restore operation without making any changes.`,
		Example: `  # Restore from a specific snapshot
  gz-shellforge restore --file ~/.zshrc --snapshot 2025-11-27_14-30-45

  # Preview restore without applying changes
  gz-shellforge restore --file ~/.zshrc --snapshot 2025-11-27_14-30-45 --dry-run

  # Restore from custom backup directory
  gz-shellforge restore --file ~/.zshrc --snapshot 2025-11-27_14-30-45 --backup-dir ~/my-backups

  # Restore without git operations
  gz-shellforge restore --file ~/.zshrc --snapshot 2025-11-27_14-30-45 --no-git`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRestore(flags)
		},
	}

	// Required flags
	cmd.Flags().StringVarP(&flags.file, "file", "f", "", "File to restore to (required)")
	cmd.Flags().StringVarP(&flags.snapshot, "snapshot", "s", "", "Snapshot timestamp to restore (required)")
	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("snapshot")

	// Optional flags
	cmd.Flags().StringVar(&flags.backupDir, "backup-dir", "", "Backup directory (default: ~/.backup/shellforge)")
	cmd.Flags().BoolVar(&flags.noGit, "no-git", false, "Disable git versioning")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Preview restore without executing")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed output")

	return cmd
}

func runRestore(flags *restoreFlags) error {
	// Expand home directory in file path
	filePath, err := helpers.ExpandHomePath(flags.file)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Determine backup directory
	backupDir, err := helpers.ResolveBackupDir(flags.backupDir)
	if err != nil {
		return err
	}

	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("backup directory does not exist: %s", backupDir)
	}

	if flags.verbose {
		fmt.Printf("Restore configuration:\n")
		fmt.Printf("  Target file: %s\n", filePath)
		fmt.Printf("  Snapshot:    %s\n", flags.snapshot)
		fmt.Printf("  Backup dir:  %s\n", backupDir)
		fmt.Printf("  Git enabled: %t\n", !flags.noGit)
		fmt.Printf("  Dry run:     %t\n", flags.dryRun)
		fmt.Println()
	}

	// Initialize services
	services := factory.NewBackupServices(factory.BackupOptions{
		BackupDir:  backupDir,
		GitEnabled: !flags.noGit,
	})
	config := services.Config
	backupService := services.BackupService

	// Extract file name from path for snapshot lookup
	fileName := filepath.Base(filePath)

	// Perform restore
	result, err := backupService.Restore(fileName, flags.snapshot, filePath, flags.dryRun)
	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	// Display results
	if flags.dryRun {
		fmt.Printf("🔍 Dry run - no changes made\n\n")
	} else {
		fmt.Printf("✓ Restore completed successfully\n\n")
	}

	fmt.Printf("Snapshot:\n")
	fmt.Printf("  Timestamp: %s\n", result.Snapshot.FormatTimestamp())
	fmt.Printf("  Size:      %s\n", result.Snapshot.FormatSize())
	fmt.Printf("  Target:    %s\n", result.TargetPath)

	if config.GitEnabled && !flags.dryRun {
		if result.GitCommitted {
			fmt.Printf("  Git:       Committed\n")
		} else {
			fmt.Printf("  Git:       Not committed (see details below)\n")
		}
	}

	if flags.verbose && result.Message != "" {
		fmt.Printf("\nDetails:\n")
		fmt.Printf("  %s\n", result.Message)
	}

	if flags.dryRun {
		fmt.Printf("\nTo apply this restore:\n")
		fmt.Printf("  gz-shellforge restore --file %s --snapshot %s\n", filePath, flags.snapshot)
	}

	return nil
}
