package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/git"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/snapshot"
)

type cleanupFlags struct {
	file      string
	backupDir string
	keepCount int
	keepDays  int
	noGit     bool
	dryRun    bool
	verbose   bool
}

func newCleanupCmd() *cobra.Command {
	flags := &cleanupFlags{}

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Remove old backup snapshots according to retention policy",
		Long: `Cleanup removes old backup snapshots according to retention policy.

The cleanup operation will:
1. List all snapshots for the specified file
2. Determine which snapshots to delete based on retention rules
3. Delete old snapshots (keeping at least one snapshot for safety)
4. Commit the cleanup operation to git (if enabled)

Retention Policy:
- Keep snapshots by count (--keep-count): Keep N most recent snapshots
- Keep snapshots by age (--keep-days): Keep snapshots from last N days
- Both policies work together (union): keeps snapshots matching EITHER rule
- Safety: Always keeps at least one snapshot regardless of policies

Use --dry-run to preview what would be deleted without making any changes.`,
		Example: `  # Cleanup keeping last 10 snapshots
  gz-shellforge cleanup --file ~/.zshrc --keep-count 10

  # Cleanup keeping snapshots from last 30 days
  gz-shellforge cleanup --file ~/.zshrc --keep-days 30

  # Cleanup with both count and age policies
  gz-shellforge cleanup --file ~/.zshrc --keep-count 10 --keep-days 30

  # Preview cleanup without deleting
  gz-shellforge cleanup --file ~/.zshrc --keep-count 10 --dry-run

  # Cleanup from custom backup directory
  gz-shellforge cleanup --file ~/.zshrc --keep-count 10 --backup-dir ~/my-backups

  # Cleanup without git operations
  gz-shellforge cleanup --file ~/.zshrc --keep-count 10 --no-git`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCleanup(flags)
		},
	}

	// Required flags
	cmd.Flags().StringVarP(&flags.file, "file", "f", "", "File pattern to cleanup (required)")
	cmd.MarkFlagRequired("file")

	// Retention policy flags
	cmd.Flags().IntVar(&flags.keepCount, "keep-count", 10, "Number of snapshots to keep (default: 10)")
	cmd.Flags().IntVar(&flags.keepDays, "keep-days", 30, "Days of snapshots to keep (default: 30)")

	// Optional flags
	cmd.Flags().StringVar(&flags.backupDir, "backup-dir", "", "Backup directory (default: ~/.backup/shellforge)")
	cmd.Flags().BoolVar(&flags.noGit, "no-git", false, "Disable git versioning")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Preview deletions without executing")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed output")

	return cmd
}

func runCleanup(flags *cleanupFlags) error {
	// Expand home directory in file path
	filePath, err := expandHomePath(flags.file)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Determine backup directory
	backupDir := flags.backupDir
	if backupDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		backupDir = filepath.Join(home, ".backup", "shellforge")
	} else {
		backupDir, err = expandHomePath(backupDir)
		if err != nil {
			return fmt.Errorf("invalid backup directory: %w", err)
		}
	}

	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("backup directory does not exist: %s", backupDir)
	}

	// Validate retention policy
	if flags.keepCount < 1 {
		return fmt.Errorf("keep-count must be at least 1")
	}
	if flags.keepDays < 1 {
		return fmt.Errorf("keep-days must be at least 1")
	}

	if flags.verbose {
		fmt.Printf("Cleanup configuration:\n")
		fmt.Printf("  File:        %s\n", filePath)
		fmt.Printf("  Backup dir:  %s\n", backupDir)
		fmt.Printf("  Keep count:  %d\n", flags.keepCount)
		fmt.Printf("  Keep days:   %d\n", flags.keepDays)
		fmt.Printf("  Git enabled: %t\n", !flags.noGit)
		fmt.Printf("  Dry run:     %t\n", flags.dryRun)
		fmt.Println()
	}

	// Initialize services
	fs := afero.NewOsFs()
	config := domain.NewBackupConfig(backupDir)
	config.GitEnabled = !flags.noGit
	config.KeepCount = flags.keepCount
	config.KeepDays = flags.keepDays

	snapshotMgr := snapshot.NewManager(fs, config)
	gitRepo := newGitRepositoryAdapter(git.NewRepository(backupDir))
	backupService := app.NewBackupService(snapshotMgr, gitRepo, config)

	// Extract file name from path for snapshot lookup
	fileName := filepath.Base(filePath)

	// Perform cleanup
	result, err := backupService.Cleanup(fileName, flags.dryRun)
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	// Display results
	if flags.dryRun {
		fmt.Printf("🔍 Dry run - no changes made\n\n")
	}

	if result.DeletedCount == 0 {
		fmt.Printf("✓ No snapshots to delete\n\n")
		fmt.Printf("Summary:\n")
		fmt.Printf("  Total snapshots: %d\n", result.RemainingCount)
		fmt.Printf("  Policy: keep %d snapshots or %d days\n", flags.keepCount, flags.keepDays)
		return nil
	}

	if flags.dryRun {
		fmt.Printf("Would delete %d snapshot(s):\n\n", result.DeletedCount)
	} else {
		fmt.Printf("✓ Cleanup completed successfully\n\n")
		fmt.Printf("Deleted %d snapshot(s):\n\n", result.DeletedCount)
	}

	// List snapshots to be deleted (or deleted)
	for i, snapshot := range result.DeletedSnapshots {
		fmt.Printf("  %d. %s (%s)\n", i+1, snapshot.FormatTimestamp(), snapshot.FormatSize())
	}

	fmt.Printf("\nSummary:\n")
	if flags.dryRun {
		fmt.Printf("  Would delete: %d\n", result.DeletedCount)
	} else {
		fmt.Printf("  Deleted:      %d\n", result.DeletedCount)
	}
	fmt.Printf("  Remaining:    %d\n", result.RemainingCount)
	fmt.Printf("  Policy:       keep %d snapshots or %d days\n", flags.keepCount, flags.keepDays)

	if config.GitEnabled && !flags.dryRun {
		if result.DeletedCount > 0 {
			fmt.Printf("  Git:          Committed\n")
		}
	}

	if flags.verbose && result.Message != "" {
		fmt.Printf("\nDetails:\n")
		fmt.Printf("  %s\n", result.Message)
	}

	if flags.dryRun && result.DeletedCount > 0 {
		fmt.Printf("\nTo apply this cleanup:\n")
		fmt.Printf("  gz-shellforge cleanup --file %s --keep-count %d --keep-days %d\n",
			filePath, flags.keepCount, flags.keepDays)
	}

	return nil
}
