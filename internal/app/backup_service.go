package app

import (
	"fmt"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// SnapshotManager defines the interface for snapshot operations
type SnapshotManager interface {
	Initialize() error
	CreateSnapshot(sourcePath string) (*domain.Snapshot, error)
	ListSnapshots(fileName string) (*domain.SnapshotList, error)
	UpdateCurrent(sourcePath string) error
	RestoreSnapshot(snapshot *domain.Snapshot, targetPath string) error
	GetSnapshotByTimestamp(fileName, timestampStr string) (*domain.Snapshot, error)
	CleanupSnapshots(fileName string, keepCount, keepDays int) ([]domain.Snapshot, error)
}

// GitRepository defines the interface for git operations
type GitRepository interface {
	IsGitInstalled() bool
	Init() error
	IsInitialized() bool
	ConfigUser(name, email string) error
	AddAndCommit(message string, paths ...string) error
	HasChanges() (bool, error)
}

// BackupService orchestrates backup operations
type BackupService struct {
	snapshotMgr SnapshotManager
	gitRepo     GitRepository
	config      *domain.BackupConfig
}

// NewBackupService creates a new backup service
func NewBackupService(snapshotMgr SnapshotManager, gitRepo GitRepository, config *domain.BackupConfig) *BackupService {
	return &BackupService{
		snapshotMgr: snapshotMgr,
		gitRepo:     gitRepo,
		config:      config,
	}
}

// BackupResult contains information about a backup operation
type BackupResult struct {
	Snapshot     *domain.Snapshot
	GitCommitted bool
	Message      string
}

// Backup creates a backup of the source file
func (s *BackupService) Backup(sourcePath, message string) (*BackupResult, error) {
	// Initialize backup directories if needed
	if err := s.snapshotMgr.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize backup directories: %w", err)
	}

	// Create snapshot
	snapshot, err := s.snapshotMgr.CreateSnapshot(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Update current copy
	if err := s.snapshotMgr.UpdateCurrent(sourcePath); err != nil {
		return nil, fmt.Errorf("failed to update current copy: %w", err)
	}

	result := &BackupResult{
		Snapshot:     snapshot,
		GitCommitted: false,
		Message:      fmt.Sprintf("Backup created: %s", snapshot.FormatTimestamp()),
	}

	// Git operations (optional, enabled by config)
	if s.config.GitEnabled {
		if err := s.initializeGit(); err != nil {
			// Git initialization failure is not fatal
			result.Message += fmt.Sprintf(" (git init failed: %v)", err)
			return result, nil
		}

		// Commit the snapshot
		commitMsg := message
		if commitMsg == "" {
			commitMsg = fmt.Sprintf("Backup %s at %s", snapshot.FileName, snapshot.FormatTimestamp())
		}

		if err := s.gitRepo.AddAndCommit(commitMsg); err != nil {
			// Git commit failure is not fatal
			result.Message += fmt.Sprintf(" (git commit failed: %v)", err)
			return result, nil
		}

		result.GitCommitted = true
		result.Message += " (committed to git)"
	}

	return result, nil
}

// RestoreResult contains information about a restore operation
type RestoreResult struct {
	Snapshot     *domain.Snapshot
	TargetPath   string
	GitCommitted bool
	Message      string
}

// Restore restores a snapshot to the target path
func (s *BackupService) Restore(fileName, timestampStr, targetPath string, dryRun bool) (*RestoreResult, error) {
	// Find the snapshot
	snapshot, err := s.snapshotMgr.GetSnapshotByTimestamp(fileName, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("failed to find snapshot: %w", err)
	}

	result := &RestoreResult{
		Snapshot:     snapshot,
		TargetPath:   targetPath,
		GitCommitted: false,
	}

	if dryRun {
		result.Message = fmt.Sprintf("Would restore %s (%s) to %s",
			snapshot.FormatTimestamp(),
			snapshot.FormatSize(),
			targetPath)
		return result, nil
	}

	// Create backup of current file before restoring (if git is enabled)
	if s.config.GitEnabled {
		if _, err := s.Backup(targetPath, fmt.Sprintf("Pre-restore backup of %s", fileName)); err != nil {
			// Backup failure is not fatal, but warn
			result.Message += fmt.Sprintf("Warning: pre-restore backup failed: %v\n", err)
		}
	}

	// Restore the snapshot
	if err := s.snapshotMgr.RestoreSnapshot(snapshot, targetPath); err != nil {
		return nil, fmt.Errorf("failed to restore snapshot: %w", err)
	}

	result.Message = fmt.Sprintf("Restored %s (%s) to %s",
		snapshot.FormatTimestamp(),
		snapshot.FormatSize(),
		targetPath)

	// Commit the restore operation (if git is enabled)
	if s.config.GitEnabled {
		if err := s.initializeGit(); err == nil {
			commitMsg := fmt.Sprintf("Restore %s from %s", fileName, snapshot.FormatTimestamp())
			if err := s.gitRepo.AddAndCommit(commitMsg); err == nil {
				result.GitCommitted = true
				result.Message += " (committed to git)"
			}
		}
	}

	return result, nil
}

// CleanupResult contains information about a cleanup operation
type CleanupResult struct {
	DeletedSnapshots []domain.Snapshot
	DeletedCount     int
	RemainingCount   int
	Message          string
}

// Cleanup removes old snapshots according to retention policy
func (s *BackupService) Cleanup(fileName string, dryRun bool) (*CleanupResult, error) {
	// Get all snapshots
	list, err := s.snapshotMgr.ListSnapshots(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	// Determine which snapshots to delete
	toDelete := list.GetToDelete(s.config.KeepCount, s.config.KeepDays)

	result := &CleanupResult{
		DeletedSnapshots: toDelete,
		DeletedCount:     len(toDelete),
		RemainingCount:   len(list.Snapshots) - len(toDelete),
	}

	if dryRun {
		result.Message = fmt.Sprintf("Would delete %d snapshot(s), keeping %d",
			result.DeletedCount,
			result.RemainingCount)
		return result, nil
	}

	// Delete snapshots
	deleted, err := s.snapshotMgr.CleanupSnapshots(fileName, s.config.KeepCount, s.config.KeepDays)
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup snapshots: %w", err)
	}

	result.DeletedSnapshots = deleted
	result.DeletedCount = len(deleted)
	result.Message = fmt.Sprintf("Deleted %d snapshot(s), kept %d",
		result.DeletedCount,
		result.RemainingCount)

	// Commit the cleanup (if git is enabled and there are changes)
	if s.config.GitEnabled && len(deleted) > 0 {
		if err := s.initializeGit(); err == nil {
			commitMsg := fmt.Sprintf("Cleanup %s: deleted %d old snapshots", fileName, len(deleted))
			if err := s.gitRepo.AddAndCommit(commitMsg); err == nil {
				result.Message += " (committed to git)"
			}
		}
	}

	return result, nil
}

// ListSnapshots returns all snapshots for a file
func (s *BackupService) ListSnapshots(fileName string) (*domain.SnapshotList, error) {
	return s.snapshotMgr.ListSnapshots(fileName)
}

// initializeGit initializes git repository if needed
func (s *BackupService) initializeGit() error {
	if !s.gitRepo.IsGitInstalled() {
		return fmt.Errorf("git is not installed")
	}

	if s.gitRepo.IsInitialized() {
		return nil // Already initialized
	}

	// Initialize git repo
	if err := s.gitRepo.Init(); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	// Configure user (use default shellforge identity)
	if err := s.gitRepo.ConfigUser("Shellforge Backup", "backup@shellforge.local"); err != nil {
		return fmt.Errorf("git config failed: %w", err)
	}

	return nil
}
