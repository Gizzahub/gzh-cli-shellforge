package cli

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/helpers"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/git"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/snapshot"
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

	if flags.verbose {
		fmt.Printf("Backup configuration:\n")
		fmt.Printf("  Source file: %s\n", filePath)
		fmt.Printf("  Backup dir:  %s\n", backupDir)
		fmt.Printf("  Git enabled: %t\n", !flags.noGit)
		fmt.Println()
	}

	// Initialize services
	fs := afero.NewOsFs()
	config := domain.NewBackupConfig(backupDir)
	config.GitEnabled = !flags.noGit

	snapshotMgr := snapshot.NewManager(fs, config)
	gitRepo := newGitRepositoryAdapter(git.NewRepository(backupDir))
	backupService := app.NewBackupService(snapshotMgr, gitRepo, config)

	// Perform backup
	result, err := backupService.Backup(filePath, flags.message)
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Display results
	fmt.Printf("✓ Backup created successfully\n")
	fmt.Printf("\n")
	fmt.Printf("Snapshot:\n")
	fmt.Printf("  Timestamp: %s\n", result.Snapshot.FormatTimestamp())
	fmt.Printf("  Size:      %s\n", result.Snapshot.FormatSize())
	fmt.Printf("  Location:  %s\n", result.Snapshot.FilePath)

	if config.GitEnabled {
		if result.GitCommitted {
			fmt.Printf("  Git:       Committed\n")
		} else {
			fmt.Printf("  Git:       Not committed (see details below)\n")
		}
	}

	if flags.verbose {
		fmt.Printf("\nDetails:\n")
		fmt.Printf("  %s\n", result.Message)
	}

	fmt.Printf("\nTo restore this backup:\n")
	fmt.Printf("  gz-shellforge restore --file %s --snapshot %s\n", filePath, result.Snapshot.FormatTimestamp())

	return nil
}

// gitRepositoryAdapter adapts git.Repository to app.GitRepository interface
type gitRepositoryAdapter struct {
	repo *git.Repository
}

func newGitRepositoryAdapter(repo *git.Repository) *gitRepositoryAdapter {
	return &gitRepositoryAdapter{repo: repo}
}

func (a *gitRepositoryAdapter) IsGitInstalled() bool {
	return git.IsGitInstalled()
}

func (a *gitRepositoryAdapter) Init() error {
	return a.repo.Init()
}

func (a *gitRepositoryAdapter) IsInitialized() bool {
	return a.repo.IsInitialized()
}

func (a *gitRepositoryAdapter) ConfigUser(name, email string) error {
	return a.repo.ConfigUser(name, email)
}

func (a *gitRepositoryAdapter) AddAndCommit(message string, paths ...string) error {
	return a.repo.AddAndCommit(message, paths...)
}

func (a *gitRepositoryAdapter) HasChanges() (bool, error) {
	return a.repo.HasChanges()
}
