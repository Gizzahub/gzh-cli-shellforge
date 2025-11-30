package domain

import (
	"fmt"
	"path/filepath"
	"time"
)

// Snapshot represents a timestamped backup of a configuration file
type Snapshot struct {
	Timestamp time.Time
	FilePath  string // Full path to snapshot file
	FileName  string // Original filename (e.g., "zshrc")
	Size      int64
}

// FormatTimestamp returns the timestamp in the standard format (YYYY-MM-DD_HH-MM-SS)
func (s *Snapshot) FormatTimestamp() string {
	return s.Timestamp.Format("2006-01-02_15-04-05")
}

// FormatSize returns a human-readable file size
func (s *Snapshot) FormatSize() string {
	const unit = 1024
	if s.Size < unit {
		return fmt.Sprintf("%d B", s.Size)
	}
	div, exp := int64(unit), 0
	for n := s.Size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(s.Size)/float64(div), "KMGTPE"[exp])
}

// SnapshotList represents a collection of snapshots for a file
type SnapshotList struct {
	Snapshots []Snapshot
	FileName  string
}

// SortByNewest sorts snapshots by timestamp (newest first)
func (sl *SnapshotList) SortByNewest() {
	// Simple bubble sort - adequate for small lists
	for i := 0; i < len(sl.Snapshots)-1; i++ {
		for j := i + 1; j < len(sl.Snapshots); j++ {
			if sl.Snapshots[j].Timestamp.After(sl.Snapshots[i].Timestamp) {
				sl.Snapshots[i], sl.Snapshots[j] = sl.Snapshots[j], sl.Snapshots[i]
			}
		}
	}
}

// FilterByAge returns snapshots within the specified number of days
func (sl *SnapshotList) FilterByAge(days int) []Snapshot {
	cutoff := time.Now().AddDate(0, 0, -days)
	var filtered []Snapshot
	for _, snapshot := range sl.Snapshots {
		if snapshot.Timestamp.After(cutoff) {
			filtered = append(filtered, snapshot)
		}
	}
	return filtered
}

// KeepNewest returns the N most recent snapshots
func (sl *SnapshotList) KeepNewest(count int) []Snapshot {
	sl.SortByNewest()
	if count >= len(sl.Snapshots) {
		return sl.Snapshots
	}
	if count < 1 {
		count = 1 // Never delete all snapshots
	}
	return sl.Snapshots[:count]
}

// GetToDelete returns snapshots that should be deleted based on retention policy
func (sl *SnapshotList) GetToDelete(keepCount int, keepDays int) []Snapshot {
	sl.SortByNewest()

	// Build map of snapshots to keep
	keepMap := make(map[string]bool)

	// Keep by count
	if keepCount > 0 {
		toKeep := sl.KeepNewest(keepCount)
		for _, s := range toKeep {
			keepMap[s.FilePath] = true
		}
	}

	// Keep by age
	if keepDays > 0 {
		toKeep := sl.FilterByAge(keepDays)
		for _, s := range toKeep {
			keepMap[s.FilePath] = true
		}
	}

	// If no policy specified, keep all
	if keepCount == 0 && keepDays == 0 {
		return []Snapshot{}
	}

	// Always keep at least one snapshot
	if len(keepMap) == 0 && len(sl.Snapshots) > 0 {
		keepMap[sl.Snapshots[0].FilePath] = true
	}

	// Collect snapshots to delete
	var toDelete []Snapshot
	for _, snapshot := range sl.Snapshots {
		if !keepMap[snapshot.FilePath] {
			toDelete = append(toDelete, snapshot)
		}
	}

	return toDelete
}

// BackupConfig represents the configuration for backup operations
type BackupConfig struct {
	BackupDir    string // ~/.backup/shellforge
	SnapshotsDir string // ~/.backup/shellforge/snapshots
	CurrentDir   string // ~/.backup/shellforge/current
	GitEnabled   bool   // Whether to use git for versioning
	KeepCount    int    // Number of snapshots to keep (0 = unlimited)
	KeepDays     int    // Days to keep snapshots (0 = unlimited)
}

// NewBackupConfig creates a default backup configuration
func NewBackupConfig(backupDir string) *BackupConfig {
	return &BackupConfig{
		BackupDir:    backupDir,
		SnapshotsDir: filepath.Join(backupDir, "snapshots"),
		CurrentDir:   filepath.Join(backupDir, "current"),
		GitEnabled:   true,
		KeepCount:    10, // Keep last 10 snapshots by default
		KeepDays:     30, // Keep 30 days by default
	}
}

// SnapshotError represents errors during snapshot operations
type SnapshotError struct {
	Operation string
	Path      string
	Err       error
}

func (e *SnapshotError) Error() string {
	return fmt.Sprintf("snapshot %s failed for %s: %v", e.Operation, e.Path, e.Err)
}

func (e *SnapshotError) Unwrap() error {
	return e.Err
}

// NewSnapshotError creates a new snapshot error
func NewSnapshotError(operation, path string, err error) *SnapshotError {
	return &SnapshotError{
		Operation: operation,
		Path:      path,
		Err:       err,
	}
}
