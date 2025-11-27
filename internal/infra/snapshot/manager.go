package snapshot

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/spf13/afero"
)

// Manager handles snapshot file operations
type Manager struct {
	fs     afero.Fs
	config *domain.BackupConfig
}

// NewManager creates a new snapshot manager
func NewManager(fs afero.Fs, config *domain.BackupConfig) *Manager {
	return &Manager{
		fs:     fs,
		config: config,
	}
}

// Initialize creates the backup directory structure
func (m *Manager) Initialize() error {
	// Create snapshots directory
	if err := m.fs.MkdirAll(m.config.SnapshotsDir, 0755); err != nil {
		return domain.NewSnapshotError("initialize", m.config.SnapshotsDir, err)
	}

	// Create current directory
	if err := m.fs.MkdirAll(m.config.CurrentDir, 0755); err != nil {
		return domain.NewSnapshotError("initialize", m.config.CurrentDir, err)
	}

	return nil
}

// CreateSnapshot creates a timestamped snapshot of a file
func (m *Manager) CreateSnapshot(sourcePath string) (*domain.Snapshot, error) {
	// Verify source file exists
	exists, err := afero.Exists(m.fs, sourcePath)
	if err != nil {
		return nil, domain.NewSnapshotError("check source", sourcePath, err)
	}
	if !exists {
		return nil, domain.NewSnapshotError("check source", sourcePath, fmt.Errorf("file does not exist"))
	}

	// Get file info
	info, err := m.fs.Stat(sourcePath)
	if err != nil {
		return nil, domain.NewSnapshotError("stat source", sourcePath, err)
	}

	// Extract filename (e.g., .zshrc → zshrc)
	fileName := filepath.Base(sourcePath)
	fileName = strings.TrimPrefix(fileName, ".")

	// Create snapshot directory for this file
	snapshotDir := filepath.Join(m.config.SnapshotsDir, fileName)
	if err := m.fs.MkdirAll(snapshotDir, 0755); err != nil {
		return nil, domain.NewSnapshotError("create directory", snapshotDir, err)
	}

	// Generate timestamp
	timestamp := time.Now()
	timestampStr := timestamp.Format("2006-01-02_15-04-05")

	// Create snapshot path
	snapshotPath := filepath.Join(snapshotDir, timestampStr)

	// Copy file to snapshot location
	if err := m.copyFile(sourcePath, snapshotPath); err != nil {
		return nil, domain.NewSnapshotError("copy file", snapshotPath, err)
	}

	// Create snapshot object
	snapshot := &domain.Snapshot{
		Timestamp: timestamp,
		FilePath:  snapshotPath,
		FileName:  fileName,
		Size:      info.Size(),
	}

	return snapshot, nil
}

// ListSnapshots returns all snapshots for a given file
func (m *Manager) ListSnapshots(fileName string) (*domain.SnapshotList, error) {
	// Remove leading dot if present
	fileName = strings.TrimPrefix(fileName, ".")

	snapshotDir := filepath.Join(m.config.SnapshotsDir, fileName)

	// Check if directory exists
	exists, err := afero.DirExists(m.fs, snapshotDir)
	if err != nil {
		return nil, domain.NewSnapshotError("check directory", snapshotDir, err)
	}
	if !exists {
		// No snapshots yet - return empty list
		return &domain.SnapshotList{
			Snapshots: []domain.Snapshot{},
			FileName:  fileName,
		}, nil
	}

	// Read directory entries
	entries, err := afero.ReadDir(m.fs, snapshotDir)
	if err != nil {
		return nil, domain.NewSnapshotError("read directory", snapshotDir, err)
	}

	var snapshots []domain.Snapshot
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}

		// Parse timestamp from filename
		timestamp, err := time.Parse("2006-01-02_15-04-05", entry.Name())
		if err != nil {
			// Skip files that don't match timestamp format
			continue
		}

		snapshot := domain.Snapshot{
			Timestamp: timestamp,
			FilePath:  filepath.Join(snapshotDir, entry.Name()),
			FileName:  fileName,
			Size:      entry.Size(),
		}
		snapshots = append(snapshots, snapshot)
	}

	list := &domain.SnapshotList{
		Snapshots: snapshots,
		FileName:  fileName,
	}

	// Sort by newest first
	list.SortByNewest()

	return list, nil
}

// DeleteSnapshot deletes a single snapshot file
func (m *Manager) DeleteSnapshot(snapshot *domain.Snapshot) error {
	exists, err := afero.Exists(m.fs, snapshot.FilePath)
	if err != nil {
		return domain.NewSnapshotError("check file", snapshot.FilePath, err)
	}
	if !exists {
		// Already deleted, not an error
		return nil
	}

	if err := m.fs.Remove(snapshot.FilePath); err != nil {
		return domain.NewSnapshotError("delete", snapshot.FilePath, err)
	}

	return nil
}

// DeleteSnapshots deletes multiple snapshots
func (m *Manager) DeleteSnapshots(snapshots []domain.Snapshot) error {
	for _, snapshot := range snapshots {
		if err := m.DeleteSnapshot(&snapshot); err != nil {
			return err
		}
	}
	return nil
}

// RestoreSnapshot restores a snapshot to the target path
func (m *Manager) RestoreSnapshot(snapshot *domain.Snapshot, targetPath string) error {
	// Verify snapshot exists
	exists, err := afero.Exists(m.fs, snapshot.FilePath)
	if err != nil {
		return domain.NewSnapshotError("check snapshot", snapshot.FilePath, err)
	}
	if !exists {
		return domain.NewSnapshotError("check snapshot", snapshot.FilePath, fmt.Errorf("snapshot does not exist"))
	}

	// Create target directory if needed
	targetDir := filepath.Dir(targetPath)
	if err := m.fs.MkdirAll(targetDir, 0755); err != nil {
		return domain.NewSnapshotError("create target directory", targetDir, err)
	}

	// Copy snapshot to target
	if err := m.copyFile(snapshot.FilePath, targetPath); err != nil {
		return domain.NewSnapshotError("restore", targetPath, err)
	}

	return nil
}

// UpdateCurrent updates the "current" copy of a file
func (m *Manager) UpdateCurrent(sourcePath string) error {
	fileName := filepath.Base(sourcePath)
	fileName = strings.TrimPrefix(fileName, ".")

	currentPath := filepath.Join(m.config.CurrentDir, fileName)

	// Copy file to current location
	if err := m.copyFile(sourcePath, currentPath); err != nil {
		return domain.NewSnapshotError("update current", currentPath, err)
	}

	return nil
}

// CleanupSnapshots deletes snapshots according to retention policy
func (m *Manager) CleanupSnapshots(fileName string, keepCount, keepDays int) ([]domain.Snapshot, error) {
	// Get all snapshots
	list, err := m.ListSnapshots(fileName)
	if err != nil {
		return nil, err
	}

	// Determine which snapshots to delete
	toDelete := list.GetToDelete(keepCount, keepDays)

	// Delete the snapshots
	if err := m.DeleteSnapshots(toDelete); err != nil {
		return nil, err
	}

	return toDelete, nil
}

// copyFile copies a file from source to destination
func (m *Manager) copyFile(sourcePath, destPath string) error {
	// Open source file
	source, err := m.fs.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer source.Close()

	// Create destination file
	dest, err := m.fs.Create(destPath)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer dest.Close()

	// Copy contents
	if _, err := io.Copy(dest, source); err != nil {
		return fmt.Errorf("copy contents: %w", err)
	}

	// Get source file permissions
	sourceInfo, err := m.fs.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	// Set destination file permissions to match source
	if err := m.fs.Chmod(destPath, sourceInfo.Mode()); err != nil {
		return fmt.Errorf("chmod destination: %w", err)
	}

	return nil
}

// GetSnapshotByTimestamp finds a snapshot by timestamp string
func (m *Manager) GetSnapshotByTimestamp(fileName, timestampStr string) (*domain.Snapshot, error) {
	list, err := m.ListSnapshots(fileName)
	if err != nil {
		return nil, err
	}

	for _, snapshot := range list.Snapshots {
		if snapshot.FormatTimestamp() == timestampStr {
			return &snapshot, nil
		}
	}

	return nil, domain.NewSnapshotError("find snapshot", timestampStr, fmt.Errorf("snapshot not found"))
}
