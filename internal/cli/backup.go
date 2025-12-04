package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/factory"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/helpers"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/output"
)

type backupFlags struct {
	file      string
	message   string
	backupDir string
	noGit     bool
	verbose   bool
}

func newBackupCmd() *cobra.Command {
	flags := &backupFlags{}

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Create a backup of a shell configuration file",
		Long: `Backup creates a timestamped snapshot of your shell configuration file.

The backup is stored in a structured directory with optional git versioning
for history tracking. This allows you to safely experiment with configuration
changes and restore previous versions if needed.`,
		Example: `  # Backup your zsh configuration
  gz-shellforge backup --file ~/.zshrc

  # Backup with custom message
  gz-shellforge backup --file ~/.zshrc --message "Before major refactor"

  # Backup without git versioning
  gz-shellforge backup --file ~/.bashrc --no-git

  # Backup to custom directory
  gz-shellforge backup --file ~/.zshrc --backup-dir ~/my-backups`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBackup(flags)
		},
	}

	// Required flags
	cmd.Flags().StringVarP(&flags.file, "file", "f", "", "File to backup (required)")
	cmd.MarkFlagRequired("file")

	// Optional flags
	cmd.Flags().StringVarP(&flags.message, "message", "m", "", "Backup description message")
	cmd.Flags().StringVar(&flags.backupDir, "backup-dir", "", "Backup directory (default: ~/.backup/shellforge)")
	cmd.Flags().BoolVar(&flags.noGit, "no-git", false, "Disable git versioning")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed output")

	return cmd
}

func runBackup(flags *backupFlags) error {
	// Expand home directory in file path
	filePath, err := helpers.ExpandHomePath(flags.file)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Determine backup directory
	backupDir, err := helpers.ResolveBackupDir(flags.backupDir)
	if err != nil {
		return err
	}

	output.NewConfigPrinter("Backup configuration").
		Add("Source file", filePath).
		Add("Backup dir", backupDir).
		Add("Git enabled", !flags.noGit).
		Print(flags.verbose)

	// Initialize services
	services := factory.NewBackupServices(factory.BackupOptions{
		BackupDir:  backupDir,
		GitEnabled: !flags.noGit,
	})
	config := services.Config
	backupService := services.BackupService

	// Perform backup
	result, err := backupService.Backup(filePath, flags.message)
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Display results
	output.SuccessResult("Backup created successfully")
	fmt.Println()

	(&output.SnapshotInfo{
		Timestamp:    result.Snapshot.FormatTimestamp(),
		Size:         result.Snapshot.FormatSize(),
		Location:     result.Snapshot.FilePath,
		ShowGit:      config.GitEnabled,
		GitCommitted: result.GitCommitted,
	}).Print()

	output.PrintDetails(flags.verbose, result.Message)
	output.PrintRestoreHint(filePath, result.Snapshot.FormatTimestamp())

	return nil
}
