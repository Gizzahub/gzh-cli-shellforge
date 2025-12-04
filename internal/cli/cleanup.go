package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	clierrors "github.com/gizzahub/gzh-cli-shellforge/internal/cli/errors"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/factory"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/helpers"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/output"
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
	filePath, err := helpers.ExpandHomePath(flags.file)
	if err != nil {
		return clierrors.InvalidPath("file", err)
	}

	// Determine backup directory
	backupDir, err := helpers.ResolveBackupDir(flags.backupDir)
	if err != nil {
		return err
	}

	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return clierrors.DirNotFound(backupDir)
	}

	// Validate retention policy
	if flags.keepCount < 1 {
		return clierrors.MinValue("keep-count", 1)
	}
	if flags.keepDays < 1 {
		return clierrors.MinValue("keep-days", 1)
	}

	output.NewConfigPrinter("Cleanup configuration").
		Add("File", filePath).
		Add("Backup dir", backupDir).
		Add("Keep count", flags.keepCount).
		Add("Keep days", flags.keepDays).
		Add("Git enabled", !flags.noGit).
		Add("Dry run", flags.dryRun).
		Print(flags.verbose)

	// Initialize services
	services := factory.NewBackupServices(factory.BackupOptions{
		BackupDir:  backupDir,
		GitEnabled: !flags.noGit,
		KeepCount:  flags.keepCount,
		KeepDays:   flags.keepDays,
	})
	config := services.Config
	backupService := services.BackupService

	// Extract file name from path for snapshot lookup
	fileName := filepath.Base(filePath)

	// Perform cleanup
	result, err := backupService.Cleanup(fileName, flags.dryRun)
	if err != nil {
		return clierrors.WrapError("cleanup", err)
	}

	// Display results
	if flags.dryRun {
		output.DryRunNotice()
	}

	if result.DeletedCount == 0 {
		output.SuccessResult("No snapshots to delete")
		fmt.Println()
		output.NewSummary().
			Add("Total snapshots", result.RemainingCount).
			Add("Policy", fmt.Sprintf("keep %d snapshots or %d days", flags.keepCount, flags.keepDays)).
			Print()
		return nil
	}

	if flags.dryRun {
		fmt.Printf("Would delete %d snapshot(s):\n\n", result.DeletedCount)
	} else {
		output.SuccessResult("Cleanup completed successfully")
		fmt.Println()
		fmt.Printf("Deleted %d snapshot(s):\n\n", result.DeletedCount)
	}

	// List snapshots to be deleted (or deleted)
	for i, snapshot := range result.DeletedSnapshots {
		fmt.Printf("  %d. %s (%s)\n", i+1, snapshot.FormatTimestamp(), snapshot.FormatSize())
	}

	summary := output.NewSummary()
	if flags.dryRun {
		summary.Add("Would delete", result.DeletedCount)
	} else {
		summary.Add("Deleted", result.DeletedCount)
	}
	summary.Add("Remaining", result.RemainingCount).
		Add("Policy", fmt.Sprintf("keep %d snapshots or %d days", flags.keepCount, flags.keepDays))

	if config.GitEnabled && !flags.dryRun && result.DeletedCount > 0 {
		summary.Add("Git", "Committed")
	}
	summary.Print()

	output.PrintDetails(flags.verbose, result.Message)

	if flags.dryRun && result.DeletedCount > 0 {
		output.PrintApplyHint(fmt.Sprintf("gz-shellforge cleanup --file %s --keep-count %d --keep-days %d",
			filePath, flags.keepCount, flags.keepDays))
	}

	return nil
}
