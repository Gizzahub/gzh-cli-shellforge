package snapshot

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestManager(t *testing.T) (*Manager, afero.Fs, *domain.BackupConfig) {
	fs := afero.NewMemMapFs()
	config := domain.NewBackupConfig("/backup/shellforge")

	manager := NewManager(fs, config)

	return manager, fs, config
}

func TestNewManager(t *testing.T) {
	fs := afero.NewMemMapFs()
	config := domain.NewBackupConfig("/test/backup")

	manager := NewManager(fs, config)

	assert.NotNil(t, manager)
	assert.Equal(t, fs, manager.fs)
	assert.Equal(t, config, manager.config)
}

func TestManager_Initialize(t *testing.T) {
	manager, fs, config := setupTestManager(t)

	t.Run("creates backup directory structure", func(t *testing.T) {
		err := manager.Initialize()
		require.NoError(t, err)

		// Verify snapshots directory exists
		exists, err := afero.DirExists(fs, config.SnapshotsDir)
		require.NoError(t, err)
		assert.True(t, exists, "snapshots directory should exist")

		// Verify current directory exists
		exists, err = afero.DirExists(fs, config.CurrentDir)
		require.NoError(t, err)
		assert.True(t, exists, "current directory should exist")
	})

	t.Run("idempotent - multiple calls do not error", func(t *testing.T) {
		err := manager.Initialize()
		require.NoError(t, err)

		// Second call should not error
		err = manager.Initialize()
		require.NoError(t, err)
	})
}

func TestManager_CreateSnapshot(t *testing.T) {
	manager, fs, _ := setupTestManager(t)

	err := manager.Initialize()
	require.NoError(t, err)

	t.Run("creates snapshot of existing file", func(t *testing.T) {
		// Create source file
		sourcePath := "/home/user/.zshrc"
		content := "# zsh configuration\nexport PATH=/usr/local/bin:$PATH"
		err := afero.WriteFile(fs, sourcePath, []byte(content), 0644)
		require.NoError(t, err)

		// Create snapshot
		snapshot, err := manager.CreateSnapshot(sourcePath)
		require.NoError(t, err)
		require.NotNil(t, snapshot)

		// Verify snapshot properties
		assert.Equal(t, "zshrc", snapshot.FileName)
		assert.True(t, time.Since(snapshot.Timestamp) < time.Second)
		assert.Greater(t, snapshot.Size, int64(0))

		// Verify snapshot file exists
		exists, err := afero.Exists(fs, snapshot.FilePath)
		require.NoError(t, err)
		assert.True(t, exists)

		// Verify snapshot content matches source
		snapshotContent, err := afero.ReadFile(fs, snapshot.FilePath)
		require.NoError(t, err)
		assert.Equal(t, content, string(snapshotContent))
	})

	t.Run("handles file with leading dot", func(t *testing.T) {
		sourcePath := "/home/user/.bashrc"
		err := afero.WriteFile(fs, sourcePath, []byte("bash config"), 0644)
		require.NoError(t, err)

		snapshot, err := manager.CreateSnapshot(sourcePath)
		require.NoError(t, err)

		// Should strip leading dot from filename
		assert.Equal(t, "bashrc", snapshot.FileName)
		assert.Contains(t, snapshot.FilePath, "/snapshots/bashrc/")
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := manager.CreateSnapshot("/non/existent/file")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file does not exist")
	})

	t.Run("creates multiple snapshots with different timestamps", func(t *testing.T) {
		sourcePath := "/home/user/.profile"
		err := afero.WriteFile(fs, sourcePath, []byte("profile"), 0644)
		require.NoError(t, err)

		// Create first snapshot
		snapshot1, err := manager.CreateSnapshot(sourcePath)
		require.NoError(t, err)

		// Wait a bit to ensure different timestamp
		time.Sleep(1100 * time.Millisecond)

		// Create second snapshot
		snapshot2, err := manager.CreateSnapshot(sourcePath)
		require.NoError(t, err)

		// Timestamps should be different
		assert.NotEqual(t, snapshot1.Timestamp, snapshot2.Timestamp)
		assert.NotEqual(t, snapshot1.FilePath, snapshot2.FilePath)
	})
}

func TestManager_ListSnapshots(t *testing.T) {
	manager, fs, config := setupTestManager(t)

	err := manager.Initialize()
	require.NoError(t, err)

	t.Run("returns empty list for file with no snapshots", func(t *testing.T) {
		list, err := manager.ListSnapshots("zshrc")
		require.NoError(t, err)
		assert.NotNil(t, list)
		assert.Equal(t, "zshrc", list.FileName)
		assert.Empty(t, list.Snapshots)
	})

	t.Run("lists all snapshots for a file", func(t *testing.T) {
		// Create test snapshots directly
		snapshotDir := filepath.Join(config.SnapshotsDir, "zshrc")
		err := fs.MkdirAll(snapshotDir, 0755)
		require.NoError(t, err)

		timestamps := []string{
			"2025-11-25_10-00-00",
			"2025-11-26_14-30-00",
			"2025-11-27_09-15-00",
		}

		for _, ts := range timestamps {
			path := filepath.Join(snapshotDir, ts)
			err := afero.WriteFile(fs, path, []byte("content"), 0644)
			require.NoError(t, err)
		}

		// List snapshots
		list, err := manager.ListSnapshots("zshrc")
		require.NoError(t, err)
		assert.Len(t, list.Snapshots, 3)

		// Should be sorted by newest first
		assert.Equal(t, "2025-11-27_09-15-00", list.Snapshots[0].FormatTimestamp())
		assert.Equal(t, "2025-11-26_14-30-00", list.Snapshots[1].FormatTimestamp())
		assert.Equal(t, "2025-11-25_10-00-00", list.Snapshots[2].FormatTimestamp())
	})

	t.Run("handles filename with leading dot", func(t *testing.T) {
		list, err := manager.ListSnapshots(".zshrc")
		require.NoError(t, err)
		assert.Equal(t, "zshrc", list.FileName)
	})

	t.Run("ignores files with invalid timestamp format", func(t *testing.T) {
		snapshotDir := filepath.Join(config.SnapshotsDir, "testfile")
		err := fs.MkdirAll(snapshotDir, 0755)
		require.NoError(t, err)

		// Create valid snapshot
		validPath := filepath.Join(snapshotDir, "2025-11-27_10-00-00")
		err = afero.WriteFile(fs, validPath, []byte("valid"), 0644)
		require.NoError(t, err)

		// Create invalid snapshot (wrong format)
		invalidPath := filepath.Join(snapshotDir, "invalid-timestamp.txt")
		err = afero.WriteFile(fs, invalidPath, []byte("invalid"), 0644)
		require.NoError(t, err)

		list, err := manager.ListSnapshots("testfile")
		require.NoError(t, err)
		assert.Len(t, list.Snapshots, 1) // Only valid snapshot
	})
}

func TestManager_DeleteSnapshot(t *testing.T) {
	manager, fs, config := setupTestManager(t)

	err := manager.Initialize()
	require.NoError(t, err)

	t.Run("deletes existing snapshot", func(t *testing.T) {
		// Create test snapshot
		snapshotDir := filepath.Join(config.SnapshotsDir, "zshrc")
		err := fs.MkdirAll(snapshotDir, 0755)
		require.NoError(t, err)

		snapshotPath := filepath.Join(snapshotDir, "2025-11-27_10-00-00")
		err = afero.WriteFile(fs, snapshotPath, []byte("content"), 0644)
		require.NoError(t, err)

		// Create snapshot object
		timestamp, _ := time.Parse("2006-01-02_15-04-05", "2025-11-27_10-00-00")
		snapshot := &domain.Snapshot{
			Timestamp: timestamp,
			FilePath:  snapshotPath,
			FileName:  "zshrc",
			Size:      7,
		}

		// Delete snapshot
		err = manager.DeleteSnapshot(snapshot)
		require.NoError(t, err)

		// Verify file is deleted
		exists, err := afero.Exists(fs, snapshotPath)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("does not error for non-existent snapshot", func(t *testing.T) {
		snapshot := &domain.Snapshot{
			FilePath: "/non/existent/snapshot",
		}

		err := manager.DeleteSnapshot(snapshot)
		require.NoError(t, err) // Should not error
	})
}

func TestManager_DeleteSnapshots(t *testing.T) {
	manager, fs, config := setupTestManager(t)

	err := manager.Initialize()
	require.NoError(t, err)

	t.Run("deletes multiple snapshots", func(t *testing.T) {
		// Create test snapshots
		snapshotDir := filepath.Join(config.SnapshotsDir, "zshrc")
		err := fs.MkdirAll(snapshotDir, 0755)
		require.NoError(t, err)

		var snapshots []domain.Snapshot
		for i := 1; i <= 3; i++ {
			ts := time.Date(2025, 11, i, 10, 0, 0, 0, time.UTC)
			path := filepath.Join(snapshotDir, ts.Format("2006-01-02_15-04-05"))
			err = afero.WriteFile(fs, path, []byte("content"), 0644)
			require.NoError(t, err)

			snapshots = append(snapshots, domain.Snapshot{
				Timestamp: ts,
				FilePath:  path,
				FileName:  "zshrc",
				Size:      7,
			})
		}

		// Delete all snapshots
		err = manager.DeleteSnapshots(snapshots)
		require.NoError(t, err)

		// Verify all are deleted
		for _, snapshot := range snapshots {
			exists, err := afero.Exists(fs, snapshot.FilePath)
			require.NoError(t, err)
			assert.False(t, exists)
		}
	})
}

func TestManager_RestoreSnapshot(t *testing.T) {
	manager, fs, config := setupTestManager(t)

	err := manager.Initialize()
	require.NoError(t, err)

	t.Run("restores snapshot to target path", func(t *testing.T) {
		// Create test snapshot
		snapshotDir := filepath.Join(config.SnapshotsDir, "zshrc")
		err := fs.MkdirAll(snapshotDir, 0755)
		require.NoError(t, err)

		snapshotPath := filepath.Join(snapshotDir, "2025-11-27_10-00-00")
		content := "restored content"
		err = afero.WriteFile(fs, snapshotPath, []byte(content), 0644)
		require.NoError(t, err)

		timestamp, _ := time.Parse("2006-01-02_15-04-05", "2025-11-27_10-00-00")
		snapshot := &domain.Snapshot{
			Timestamp: timestamp,
			FilePath:  snapshotPath,
			FileName:  "zshrc",
			Size:      int64(len(content)),
		}

		// Restore to target
		targetPath := "/home/user/.zshrc"
		err = manager.RestoreSnapshot(snapshot, targetPath)
		require.NoError(t, err)

		// Verify target file exists with correct content
		restored, err := afero.ReadFile(fs, targetPath)
		require.NoError(t, err)
		assert.Equal(t, content, string(restored))
	})

	t.Run("creates target directory if needed", func(t *testing.T) {
		snapshotDir := filepath.Join(config.SnapshotsDir, "bashrc")
		err := fs.MkdirAll(snapshotDir, 0755)
		require.NoError(t, err)

		snapshotPath := filepath.Join(snapshotDir, "2025-11-27_10-00-00")
		err = afero.WriteFile(fs, snapshotPath, []byte("content"), 0644)
		require.NoError(t, err)

		timestamp, _ := time.Parse("2006-01-02_15-04-05", "2025-11-27_10-00-00")
		snapshot := &domain.Snapshot{
			Timestamp: timestamp,
			FilePath:  snapshotPath,
			FileName:  "bashrc",
			Size:      7,
		}

		// Restore to path with non-existent directory
		targetPath := "/new/directory/.bashrc"
		err = manager.RestoreSnapshot(snapshot, targetPath)
		require.NoError(t, err)

		// Verify target exists
		exists, err := afero.Exists(fs, targetPath)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("returns error for non-existent snapshot", func(t *testing.T) {
		snapshot := &domain.Snapshot{
			FilePath: "/non/existent/snapshot",
		}

		err := manager.RestoreSnapshot(snapshot, "/target/path")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "snapshot does not exist")
	})
}

func TestManager_UpdateCurrent(t *testing.T) {
	manager, fs, config := setupTestManager(t)

	err := manager.Initialize()
	require.NoError(t, err)

	t.Run("updates current copy of file", func(t *testing.T) {
		sourcePath := "/home/user/.zshrc"
		content := "current zsh config"
		err := afero.WriteFile(fs, sourcePath, []byte(content), 0644)
		require.NoError(t, err)

		err = manager.UpdateCurrent(sourcePath)
		require.NoError(t, err)

		// Verify current file exists
		currentPath := filepath.Join(config.CurrentDir, "zshrc")
		exists, err := afero.Exists(fs, currentPath)
		require.NoError(t, err)
		assert.True(t, exists)

		// Verify content
		currentContent, err := afero.ReadFile(fs, currentPath)
		require.NoError(t, err)
		assert.Equal(t, content, string(currentContent))
	})

	t.Run("strips leading dot from filename", func(t *testing.T) {
		sourcePath := "/home/user/.profile"
		err := afero.WriteFile(fs, sourcePath, []byte("profile"), 0644)
		require.NoError(t, err)

		err = manager.UpdateCurrent(sourcePath)
		require.NoError(t, err)

		// Should create file without leading dot
		currentPath := filepath.Join(config.CurrentDir, "profile")
		exists, err := afero.Exists(fs, currentPath)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestManager_CleanupSnapshots(t *testing.T) {
	manager, fs, config := setupTestManager(t)

	err := manager.Initialize()
	require.NoError(t, err)

	t.Run("deletes snapshots according to retention policy", func(t *testing.T) {
		// Create test snapshots
		snapshotDir := filepath.Join(config.SnapshotsDir, "zshrc")
		err := fs.MkdirAll(snapshotDir, 0755)
		require.NoError(t, err)

		// Create 5 snapshots
		for i := 1; i <= 5; i++ {
			ts := time.Now().AddDate(0, 0, -i)
			path := filepath.Join(snapshotDir, ts.Format("2006-01-02_15-04-05"))
			err = afero.WriteFile(fs, path, []byte("content"), 0644)
			require.NoError(t, err)
		}

		// Keep only 3 most recent
		deleted, err := manager.CleanupSnapshots("zshrc", 3, 0)
		require.NoError(t, err)
		assert.Len(t, deleted, 2) // Should delete 2 oldest

		// Verify remaining snapshots
		list, err := manager.ListSnapshots("zshrc")
		require.NoError(t, err)
		assert.Len(t, list.Snapshots, 3)
	})

	t.Run("keeps at least one snapshot when policy would delete all", func(t *testing.T) {
		// Create one snapshot
		snapshotDir := filepath.Join(config.SnapshotsDir, "testfile")
		err := fs.MkdirAll(snapshotDir, 0755)
		require.NoError(t, err)

		ts := time.Now().AddDate(0, 0, -100) // 100 days old
		path := filepath.Join(snapshotDir, ts.Format("2006-01-02_15-04-05"))
		err = afero.WriteFile(fs, path, []byte("content"), 0644)
		require.NoError(t, err)

		// Try to keep only snapshots from last 7 days (this one is 100 days old)
		deleted, err := manager.CleanupSnapshots("testfile", 0, 7)
		require.NoError(t, err)
		assert.Len(t, deleted, 0) // Should not delete the only snapshot

		// Verify snapshot still exists
		list, err := manager.ListSnapshots("testfile")
		require.NoError(t, err)
		assert.Len(t, list.Snapshots, 1)
	})
}

func TestManager_GetSnapshotByTimestamp(t *testing.T) {
	manager, fs, config := setupTestManager(t)

	err := manager.Initialize()
	require.NoError(t, err)

	t.Run("finds snapshot by timestamp string", func(t *testing.T) {
		// Create test snapshot
		snapshotDir := filepath.Join(config.SnapshotsDir, "zshrc")
		err := fs.MkdirAll(snapshotDir, 0755)
		require.NoError(t, err)

		timestampStr := "2025-11-27_14-30-45"
		path := filepath.Join(snapshotDir, timestampStr)
		err = afero.WriteFile(fs, path, []byte("content"), 0644)
		require.NoError(t, err)

		// Find snapshot
		snapshot, err := manager.GetSnapshotByTimestamp("zshrc", timestampStr)
		require.NoError(t, err)
		require.NotNil(t, snapshot)

		assert.Equal(t, timestampStr, snapshot.FormatTimestamp())
		assert.Equal(t, "zshrc", snapshot.FileName)
	})

	t.Run("returns error for non-existent timestamp", func(t *testing.T) {
		_, err := manager.GetSnapshotByTimestamp("zshrc", "2025-01-01_00-00-00")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "snapshot not found")
	})
}
