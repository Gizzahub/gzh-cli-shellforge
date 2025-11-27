package app

import (
	"fmt"
	"testing"
	"time"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock SnapshotManager
type MockSnapshotManager struct {
	mock.Mock
}

func (m *MockSnapshotManager) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSnapshotManager) CreateSnapshot(sourcePath string) (*domain.Snapshot, error) {
	args := m.Called(sourcePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Snapshot), args.Error(1)
}

func (m *MockSnapshotManager) ListSnapshots(fileName string) (*domain.SnapshotList, error) {
	args := m.Called(fileName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SnapshotList), args.Error(1)
}

func (m *MockSnapshotManager) UpdateCurrent(sourcePath string) error {
	args := m.Called(sourcePath)
	return args.Error(0)
}

func (m *MockSnapshotManager) RestoreSnapshot(snapshot *domain.Snapshot, targetPath string) error {
	args := m.Called(snapshot, targetPath)
	return args.Error(0)
}

func (m *MockSnapshotManager) GetSnapshotByTimestamp(fileName, timestampStr string) (*domain.Snapshot, error) {
	args := m.Called(fileName, timestampStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Snapshot), args.Error(1)
}

func (m *MockSnapshotManager) CleanupSnapshots(fileName string, keepCount, keepDays int) ([]domain.Snapshot, error) {
	args := m.Called(fileName, keepCount, keepDays)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Snapshot), args.Error(1)
}

// Mock GitRepository
type MockGitRepository struct {
	mock.Mock
}

func (m *MockGitRepository) IsGitInstalled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockGitRepository) Init() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGitRepository) IsInitialized() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockGitRepository) ConfigUser(name, email string) error {
	args := m.Called(name, email)
	return args.Error(0)
}

func (m *MockGitRepository) AddAndCommit(message string, paths ...string) error {
	args := m.Called(message, paths)
	return args.Error(0)
}

func (m *MockGitRepository) HasChanges() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func setupTestService(t *testing.T) (*BackupService, *MockSnapshotManager, *MockGitRepository, *domain.BackupConfig) {
	snapshotMgr := new(MockSnapshotManager)
	gitRepo := new(MockGitRepository)
	config := domain.NewBackupConfig("/backup/shellforge")

	service := NewBackupService(snapshotMgr, gitRepo, config)

	return service, snapshotMgr, gitRepo, config
}

func TestNewBackupService(t *testing.T) {
	snapshotMgr := new(MockSnapshotManager)
	gitRepo := new(MockGitRepository)
	config := domain.NewBackupConfig("/test")

	service := NewBackupService(snapshotMgr, gitRepo, config)

	assert.NotNil(t, service)
	assert.Equal(t, snapshotMgr, service.snapshotMgr)
	assert.Equal(t, gitRepo, service.gitRepo)
	assert.Equal(t, config, service.config)
}

func TestBackupService_Backup(t *testing.T) {
	service, snapshotMgr, gitRepo, _ := setupTestService(t)

	testSnapshot := &domain.Snapshot{
		Timestamp: time.Now(),
		FilePath:  "/backup/snapshots/zshrc/2025-11-27_10-00-00",
		FileName:  "zshrc",
		Size:      1024,
	}

	t.Run("creates backup without git", func(t *testing.T) {
		service.config.GitEnabled = false

		snapshotMgr.On("Initialize").Return(nil).Once()
		snapshotMgr.On("CreateSnapshot", "/home/user/.zshrc").Return(testSnapshot, nil).Once()
		snapshotMgr.On("UpdateCurrent", "/home/user/.zshrc").Return(nil).Once()

		result, err := service.Backup("/home/user/.zshrc", "Test backup")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, testSnapshot, result.Snapshot)
		assert.False(t, result.GitCommitted)
		assert.Contains(t, result.Message, "Backup created")

		snapshotMgr.AssertExpectations(t)
	})

	t.Run("creates backup with git commit", func(t *testing.T) {
		service.config.GitEnabled = true

		snapshotMgr.On("Initialize").Return(nil).Once()
		snapshotMgr.On("CreateSnapshot", "/home/user/.zshrc").Return(testSnapshot, nil).Once()
		snapshotMgr.On("UpdateCurrent", "/home/user/.zshrc").Return(nil).Once()

		gitRepo.On("IsGitInstalled").Return(true).Once()
		gitRepo.On("IsInitialized").Return(true).Once()
		gitRepo.On("AddAndCommit", "Test backup", mock.Anything).Return(nil).Once()

		result, err := service.Backup("/home/user/.zshrc", "Test backup")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.GitCommitted)
		assert.Contains(t, result.Message, "committed to git")

		snapshotMgr.AssertExpectations(t)
		gitRepo.AssertExpectations(t)
	})

	t.Run("uses default message when empty", func(t *testing.T) {
		service.config.GitEnabled = true

		snapshotMgr.On("Initialize").Return(nil).Once()
		snapshotMgr.On("CreateSnapshot", "/home/user/.zshrc").Return(testSnapshot, nil).Once()
		snapshotMgr.On("UpdateCurrent", "/home/user/.zshrc").Return(nil).Once()

		gitRepo.On("IsGitInstalled").Return(true).Once()
		gitRepo.On("IsInitialized").Return(true).Once()
		gitRepo.On("AddAndCommit", mock.MatchedBy(func(msg string) bool {
			return msg == fmt.Sprintf("Backup zshrc at %s", testSnapshot.FormatTimestamp())
		}), mock.Anything).Return(nil).Once()

		result, err := service.Backup("/home/user/.zshrc", "")
		require.NoError(t, err)
		require.NotNil(t, result)

		snapshotMgr.AssertExpectations(t)
		gitRepo.AssertExpectations(t)
	})

	t.Run("handles git init on first backup", func(t *testing.T) {
		service.config.GitEnabled = true

		snapshotMgr.On("Initialize").Return(nil).Once()
		snapshotMgr.On("CreateSnapshot", "/home/user/.zshrc").Return(testSnapshot, nil).Once()
		snapshotMgr.On("UpdateCurrent", "/home/user/.zshrc").Return(nil).Once()

		gitRepo.On("IsGitInstalled").Return(true).Once()
		gitRepo.On("IsInitialized").Return(false).Once()
		gitRepo.On("Init").Return(nil).Once()
		gitRepo.On("ConfigUser", "Shellforge Backup", "backup@shellforge.local").Return(nil).Once()
		gitRepo.On("AddAndCommit", "First backup", mock.Anything).Return(nil).Once()

		result, err := service.Backup("/home/user/.zshrc", "First backup")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.GitCommitted)

		snapshotMgr.AssertExpectations(t)
		gitRepo.AssertExpectations(t)
	})

	t.Run("continues on git commit failure", func(t *testing.T) {
		service.config.GitEnabled = true

		snapshotMgr.On("Initialize").Return(nil).Once()
		snapshotMgr.On("CreateSnapshot", "/home/user/.zshrc").Return(testSnapshot, nil).Once()
		snapshotMgr.On("UpdateCurrent", "/home/user/.zshrc").Return(nil).Once()

		gitRepo.On("IsGitInstalled").Return(true).Once()
		gitRepo.On("IsInitialized").Return(true).Once()
		gitRepo.On("AddAndCommit", "Test backup", mock.Anything).Return(fmt.Errorf("commit failed")).Once()

		result, err := service.Backup("/home/user/.zshrc", "Test backup")
		require.NoError(t, err) // Should not error
		require.NotNil(t, result)

		assert.False(t, result.GitCommitted)
		assert.Contains(t, result.Message, "git commit failed")

		snapshotMgr.AssertExpectations(t)
		gitRepo.AssertExpectations(t)
	})

	t.Run("returns error when snapshot creation fails", func(t *testing.T) {
		service.config.GitEnabled = false

		snapshotMgr.On("Initialize").Return(nil).Once()
		snapshotMgr.On("CreateSnapshot", "/home/user/.zshrc").Return(nil, fmt.Errorf("snapshot failed")).Once()

		result, err := service.Backup("/home/user/.zshrc", "Test backup")
		require.Error(t, err)
		require.Nil(t, result)

		assert.Contains(t, err.Error(), "failed to create snapshot")

		snapshotMgr.AssertExpectations(t)
	})

	t.Run("returns error when update current fails", func(t *testing.T) {
		service.config.GitEnabled = false

		snapshotMgr.On("Initialize").Return(nil).Once()
		snapshotMgr.On("CreateSnapshot", "/home/user/.zshrc").Return(testSnapshot, nil).Once()
		snapshotMgr.On("UpdateCurrent", "/home/user/.zshrc").Return(fmt.Errorf("update failed")).Once()

		result, err := service.Backup("/home/user/.zshrc", "Test backup")
		require.Error(t, err)
		require.Nil(t, result)

		assert.Contains(t, err.Error(), "failed to update current copy")

		snapshotMgr.AssertExpectations(t)
	})
}

func TestBackupService_Restore(t *testing.T) {
	service, snapshotMgr, gitRepo, _ := setupTestService(t)
	_ = gitRepo // May be used in future git-enabled restore tests

	testSnapshot := &domain.Snapshot{
		Timestamp: time.Date(2025, 11, 27, 10, 0, 0, 0, time.UTC),
		FilePath:  "/backup/snapshots/zshrc/2025-11-27_10-00-00",
		FileName:  "zshrc",
		Size:      1024,
	}

	t.Run("restores snapshot without git", func(t *testing.T) {
		service.config.GitEnabled = false

		snapshotMgr.On("GetSnapshotByTimestamp", "zshrc", "2025-11-27_10-00-00").Return(testSnapshot, nil).Once()
		snapshotMgr.On("RestoreSnapshot", testSnapshot, "/home/user/.zshrc").Return(nil).Once()

		result, err := service.Restore("zshrc", "2025-11-27_10-00-00", "/home/user/.zshrc", false)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, testSnapshot, result.Snapshot)
		assert.Equal(t, "/home/user/.zshrc", result.TargetPath)
		assert.False(t, result.GitCommitted)
		assert.Contains(t, result.Message, "Restored")

		snapshotMgr.AssertExpectations(t)
	})

	t.Run("dry run mode does not restore", func(t *testing.T) {
		snapshotMgr.On("GetSnapshotByTimestamp", "zshrc", "2025-11-27_10-00-00").Return(testSnapshot, nil).Once()

		result, err := service.Restore("zshrc", "2025-11-27_10-00-00", "/home/user/.zshrc", true)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Contains(t, result.Message, "Would restore")
		assert.False(t, result.GitCommitted)

		snapshotMgr.AssertExpectations(t)
		// RestoreSnapshot should NOT be called in dry run
		snapshotMgr.AssertNotCalled(t, "RestoreSnapshot")
	})

	t.Run("returns error when snapshot not found", func(t *testing.T) {
		snapshotMgr.On("GetSnapshotByTimestamp", "zshrc", "2025-01-01_00-00-00").Return(nil, fmt.Errorf("not found")).Once()

		result, err := service.Restore("zshrc", "2025-01-01_00-00-00", "/home/user/.zshrc", false)
		require.Error(t, err)
		require.Nil(t, result)

		assert.Contains(t, err.Error(), "failed to find snapshot")

		snapshotMgr.AssertExpectations(t)
	})
}

func TestBackupService_Cleanup(t *testing.T) {
	service, snapshotMgr, gitRepo, _ := setupTestService(t)
	_ = gitRepo // May be used in future git-enabled cleanup tests

	testList := &domain.SnapshotList{
		Snapshots: []domain.Snapshot{
			{Timestamp: time.Now(), FilePath: "/backup/snapshots/zshrc/1", FileName: "zshrc", Size: 100},
			{Timestamp: time.Now().AddDate(0, 0, -1), FilePath: "/backup/snapshots/zshrc/2", FileName: "zshrc", Size: 100},
			{Timestamp: time.Now().AddDate(0, 0, -2), FilePath: "/backup/snapshots/zshrc/3", FileName: "zshrc", Size: 100},
		},
		FileName: "zshrc",
	}

	toDelete := []domain.Snapshot{
		testList.Snapshots[2], // Oldest one
	}

	t.Run("cleans up old snapshots", func(t *testing.T) {
		service.config.GitEnabled = false

		snapshotMgr.On("ListSnapshots", "zshrc").Return(testList, nil).Once()
		snapshotMgr.On("CleanupSnapshots", "zshrc", 10, 30).Return(toDelete, nil).Once()

		result, err := service.Cleanup("zshrc", false)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, 1, result.DeletedCount)
		assert.Contains(t, result.Message, "Deleted 1 snapshot")

		snapshotMgr.AssertExpectations(t)
	})

	t.Run("dry run mode does not delete", func(t *testing.T) {
		snapshotMgr.On("ListSnapshots", "zshrc").Return(testList, nil).Once()

		result, err := service.Cleanup("zshrc", true)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Contains(t, result.Message, "Would delete")

		snapshotMgr.AssertExpectations(t)
		// CleanupSnapshots should NOT be called in dry run
		snapshotMgr.AssertNotCalled(t, "CleanupSnapshots")
	})
}

func TestBackupService_ListSnapshots(t *testing.T) {
	service, snapshotMgr, _, _ := setupTestService(t)

	testList := &domain.SnapshotList{
		Snapshots: []domain.Snapshot{
			{Timestamp: time.Now(), FilePath: "/backup/snapshots/zshrc/1", FileName: "zshrc", Size: 100},
		},
		FileName: "zshrc",
	}

	t.Run("lists snapshots", func(t *testing.T) {
		snapshotMgr.On("ListSnapshots", "zshrc").Return(testList, nil).Once()

		result, err := service.ListSnapshots("zshrc")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, testList, result)

		snapshotMgr.AssertExpectations(t)
	})

	t.Run("returns error when listing fails", func(t *testing.T) {
		snapshotMgr.On("ListSnapshots", "zshrc").Return(nil, fmt.Errorf("list failed")).Once()

		result, err := service.ListSnapshots("zshrc")
		require.Error(t, err)
		require.Nil(t, result)

		snapshotMgr.AssertExpectations(t)
	})
}
