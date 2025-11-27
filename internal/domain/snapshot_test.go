package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSnapshot_FormatTimestamp(t *testing.T) {
	snapshot := Snapshot{
		Timestamp: time.Date(2025, 11, 27, 14, 30, 45, 0, time.UTC),
	}

	formatted := snapshot.FormatTimestamp()
	assert.Equal(t, "2025-11-27_14-30-45", formatted)
}

func TestSnapshot_FormatSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{"bytes", 512, "512 B"},
		{"kilobytes", 1536, "1.5 KB"},
		{"megabytes", 1048576, "1.0 MB"},
		{"large file", 5242880, "5.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot := Snapshot{Size: tt.size}
			assert.Equal(t, tt.expected, snapshot.FormatSize())
		})
	}
}

func TestSnapshotList_SortByNewest(t *testing.T) {
	now := time.Now()
	list := &SnapshotList{
		Snapshots: []Snapshot{
			{Timestamp: now.Add(-2 * time.Hour), FileName: "old"},
			{Timestamp: now, FileName: "newest"},
			{Timestamp: now.Add(-1 * time.Hour), FileName: "middle"},
		},
	}

	list.SortByNewest()

	assert.Equal(t, "newest", list.Snapshots[0].FileName)
	assert.Equal(t, "middle", list.Snapshots[1].FileName)
	assert.Equal(t, "old", list.Snapshots[2].FileName)
}

func TestSnapshotList_FilterByAge(t *testing.T) {
	now := time.Now()
	list := &SnapshotList{
		Snapshots: []Snapshot{
			{Timestamp: now.AddDate(0, 0, -5), FileName: "5days"},
			{Timestamp: now.AddDate(0, 0, -15), FileName: "15days"},
			{Timestamp: now.AddDate(0, 0, -35), FileName: "35days"},
		},
	}

	filtered := list.FilterByAge(30)

	assert.Len(t, filtered, 2)
	assert.Equal(t, "5days", filtered[0].FileName)
	assert.Equal(t, "15days", filtered[1].FileName)
}

func TestSnapshotList_KeepNewest(t *testing.T) {
	now := time.Now()
	list := &SnapshotList{
		Snapshots: []Snapshot{
			{Timestamp: now.Add(-3 * time.Hour), FileName: "3"},
			{Timestamp: now, FileName: "1"},
			{Timestamp: now.Add(-1 * time.Hour), FileName: "2"},
		},
	}

	kept := list.KeepNewest(2)

	assert.Len(t, kept, 2)
	assert.Equal(t, "1", kept[0].FileName)
	assert.Equal(t, "2", kept[1].FileName)
}

func TestSnapshotList_KeepNewest_EnsuresMinimumOne(t *testing.T) {
	now := time.Now()
	list := &SnapshotList{
		Snapshots: []Snapshot{
			{Timestamp: now, FileName: "only"},
		},
	}

	kept := list.KeepNewest(0)

	assert.Len(t, kept, 1, "should keep at least one snapshot even with count=0")
}

func TestSnapshotList_GetToDelete(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		snapshots     []Snapshot
		keepCount     int
		keepDays      int
		expectedCount int
		expectedNames []string
	}{
		{
			name: "keep by count",
			snapshots: []Snapshot{
				{Timestamp: now, FileName: "1", FilePath: "/path/1"},
				{Timestamp: now.Add(-1 * time.Hour), FileName: "2", FilePath: "/path/2"},
				{Timestamp: now.Add(-2 * time.Hour), FileName: "3", FilePath: "/path/3"},
			},
			keepCount:     2,
			keepDays:      0,
			expectedCount: 1,
			expectedNames: []string{"3"},
		},
		{
			name: "keep by days",
			snapshots: []Snapshot{
				{Timestamp: now.AddDate(0, 0, -5), FileName: "recent", FilePath: "/path/recent"},
				{Timestamp: now.AddDate(0, 0, -35), FileName: "old", FilePath: "/path/old"},
			},
			keepCount:     0,
			keepDays:      30,
			expectedCount: 1,
			expectedNames: []string{"old"},
		},
		{
			name: "keep by both (union)",
			snapshots: []Snapshot{
				{Timestamp: now, FileName: "1", FilePath: "/path/1"},
				{Timestamp: now.AddDate(0, 0, -20), FileName: "2", FilePath: "/path/2"},
				{Timestamp: now.AddDate(0, 0, -40), FileName: "3", FilePath: "/path/3"},
			},
			keepCount:     1,
			keepDays:      30,
			expectedCount: 1,
			expectedNames: []string{"3"},
		},
		{
			name: "no policy keeps all",
			snapshots: []Snapshot{
				{Timestamp: now, FileName: "1", FilePath: "/path/1"},
				{Timestamp: now.Add(-1 * time.Hour), FileName: "2", FilePath: "/path/2"},
			},
			keepCount:     0,
			keepDays:      0,
			expectedCount: 0,
		},
		{
			name: "always keep at least one",
			snapshots: []Snapshot{
				{Timestamp: now.AddDate(0, 0, -100), FileName: "veryold", FilePath: "/path/veryold"},
			},
			keepCount:     0,
			keepDays:      30,
			expectedCount: 0, // Should keep the one snapshot
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := &SnapshotList{Snapshots: tt.snapshots}
			toDelete := list.GetToDelete(tt.keepCount, tt.keepDays)

			assert.Len(t, toDelete, tt.expectedCount)

			if len(tt.expectedNames) > 0 {
				deleteNames := make([]string, len(toDelete))
				for i, s := range toDelete {
					deleteNames[i] = s.FileName
				}
				assert.ElementsMatch(t, tt.expectedNames, deleteNames)
			}
		})
	}
}

func TestNewBackupConfig(t *testing.T) {
	config := NewBackupConfig("/home/user/.backup/shellforge")

	assert.Equal(t, "/home/user/.backup/shellforge", config.BackupDir)
	assert.Equal(t, "/home/user/.backup/shellforge/snapshots", config.SnapshotsDir)
	assert.Equal(t, "/home/user/.backup/shellforge/current", config.CurrentDir)
	assert.True(t, config.GitEnabled)
	assert.Equal(t, 10, config.KeepCount)
	assert.Equal(t, 30, config.KeepDays)
}

func TestSnapshotError(t *testing.T) {
	err := NewSnapshotError("create", "/path/to/file", assert.AnError)

	assert.Contains(t, err.Error(), "snapshot create failed")
	assert.Contains(t, err.Error(), "/path/to/file")
	assert.ErrorIs(t, err, assert.AnError)
}
