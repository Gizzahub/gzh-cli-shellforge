package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsGitInstalled(t *testing.T) {
	// This test assumes git is installed on the test system
	// If git is not installed, the test will fail
	installed := IsGitInstalled()
	assert.True(t, installed, "git should be installed for tests to run")
}

func TestRepository_Init(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git is not installed")
	}

	t.Run("creates new repository", func(t *testing.T) {
		// Create temp directory
		tmpDir := t.TempDir()

		repo := NewRepository(tmpDir)
		err := repo.Init()
		require.NoError(t, err)

		// Verify .git directory was created
		gitDir := filepath.Join(tmpDir, ".git")
		stat, err := os.Stat(gitDir)
		require.NoError(t, err)
		assert.True(t, stat.IsDir(), ".git should be a directory")
	})

	t.Run("idempotent - multiple inits do not error", func(t *testing.T) {
		tmpDir := t.TempDir()

		repo := NewRepository(tmpDir)

		// First init
		err := repo.Init()
		require.NoError(t, err)

		// Second init should not error
		err = repo.Init()
		require.NoError(t, err)
	})
}

func TestRepository_IsInitialized(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git is not installed")
	}

	t.Run("returns false for non-git directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := NewRepository(tmpDir)

		initialized := repo.IsInitialized()
		assert.False(t, initialized)
	})

	t.Run("returns true for git directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := NewRepository(tmpDir)

		err := repo.Init()
		require.NoError(t, err)

		initialized := repo.IsInitialized()
		assert.True(t, initialized)
	})
}

func TestRepository_ConfigUser(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git is not installed")
	}

	tmpDir := t.TempDir()
	repo := NewRepository(tmpDir)

	err := repo.Init()
	require.NoError(t, err)

	t.Run("sets user name and email", func(t *testing.T) {
		err := repo.ConfigUser("Test User", "test@example.com")
		require.NoError(t, err)
	})
}

func TestRepository_Add(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git is not installed")
	}

	tmpDir := t.TempDir()
	repo := NewRepository(tmpDir)

	err := repo.Init()
	require.NoError(t, err)

	err = repo.ConfigUser("Test User", "test@example.com")
	require.NoError(t, err)

	t.Run("adds single file", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		// Add the file
		err = repo.Add("test.txt")
		require.NoError(t, err)

		// Verify file is staged
		status, err := repo.GetStatus()
		require.NoError(t, err)
		assert.Contains(t, status, "test.txt")
	})

	t.Run("adds all files when no paths specified", func(t *testing.T) {
		// Create another test file
		testFile2 := filepath.Join(tmpDir, "test2.txt")
		err := os.WriteFile(testFile2, []byte("test content 2"), 0644)
		require.NoError(t, err)

		// Add all files
		err = repo.Add()
		require.NoError(t, err)

		// Verify files are staged
		status, err := repo.GetStatus()
		require.NoError(t, err)
		assert.Contains(t, status, "test2.txt")
	})

	t.Run("adds multiple files", func(t *testing.T) {
		// Create test files
		testFile3 := filepath.Join(tmpDir, "test3.txt")
		testFile4 := filepath.Join(tmpDir, "test4.txt")
		err := os.WriteFile(testFile3, []byte("test content 3"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(testFile4, []byte("test content 4"), 0644)
		require.NoError(t, err)

		// Add multiple files
		err = repo.Add("test3.txt", "test4.txt")
		require.NoError(t, err)

		// Verify files are staged
		status, err := repo.GetStatus()
		require.NoError(t, err)
		assert.Contains(t, status, "test3.txt")
		assert.Contains(t, status, "test4.txt")
	})
}

func TestRepository_Commit(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git is not installed")
	}

	tmpDir := t.TempDir()
	repo := NewRepository(tmpDir)

	err := repo.Init()
	require.NoError(t, err)

	err = repo.ConfigUser("Test User", "test@example.com")
	require.NoError(t, err)

	t.Run("creates commit with staged changes", func(t *testing.T) {
		// Create and stage a file
		testFile := filepath.Join(tmpDir, "commit_test.txt")
		err := os.WriteFile(testFile, []byte("commit test"), 0644)
		require.NoError(t, err)

		err = repo.Add("commit_test.txt")
		require.NoError(t, err)

		// Commit the changes
		err = repo.Commit("Test commit message")
		require.NoError(t, err)

		// Verify working directory is clean after commit
		hasChanges, err := repo.HasChanges()
		require.NoError(t, err)
		assert.False(t, hasChanges, "should have no changes after commit")
	})

	t.Run("handles nothing to commit gracefully", func(t *testing.T) {
		// Try to commit with no staged changes
		err := repo.Commit("Empty commit")
		require.NoError(t, err, "should not error when nothing to commit")
	})
}

func TestRepository_AddAndCommit(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git is not installed")
	}

	tmpDir := t.TempDir()
	repo := NewRepository(tmpDir)

	err := repo.Init()
	require.NoError(t, err)

	err = repo.ConfigUser("Test User", "test@example.com")
	require.NoError(t, err)

	t.Run("stages and commits in one operation", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tmpDir, "addcommit_test.txt")
		err := os.WriteFile(testFile, []byte("add and commit test"), 0644)
		require.NoError(t, err)

		// Add and commit
		err = repo.AddAndCommit("Add and commit test", "addcommit_test.txt")
		require.NoError(t, err)

		// Verify working directory is clean
		hasChanges, err := repo.HasChanges()
		require.NoError(t, err)
		assert.False(t, hasChanges)
	})

	t.Run("stages and commits all files when no paths specified", func(t *testing.T) {
		// Create test files
		testFile1 := filepath.Join(tmpDir, "file1.txt")
		testFile2 := filepath.Join(tmpDir, "file2.txt")
		err := os.WriteFile(testFile1, []byte("file 1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(testFile2, []byte("file 2"), 0644)
		require.NoError(t, err)

		// Add and commit all
		err = repo.AddAndCommit("Add all files")
		require.NoError(t, err)

		// Verify working directory is clean
		hasChanges, err := repo.HasChanges()
		require.NoError(t, err)
		assert.False(t, hasChanges)
	})
}

func TestRepository_GetStatus(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git is not installed")
	}

	tmpDir := t.TempDir()
	repo := NewRepository(tmpDir)

	err := repo.Init()
	require.NoError(t, err)

	err = repo.ConfigUser("Test User", "test@example.com")
	require.NoError(t, err)

	t.Run("returns empty string for clean working directory", func(t *testing.T) {
		status, err := repo.GetStatus()
		require.NoError(t, err)
		assert.Empty(t, status)
	})

	t.Run("returns status for untracked files", func(t *testing.T) {
		// Create an untracked file
		testFile := filepath.Join(tmpDir, "untracked.txt")
		err := os.WriteFile(testFile, []byte("untracked"), 0644)
		require.NoError(t, err)

		status, err := repo.GetStatus()
		require.NoError(t, err)
		assert.Contains(t, status, "untracked.txt")
		assert.Contains(t, status, "??") // Git short status for untracked
	})

	t.Run("returns status for staged files", func(t *testing.T) {
		// Stage the untracked file
		err := repo.Add("untracked.txt")
		require.NoError(t, err)

		status, err := repo.GetStatus()
		require.NoError(t, err)
		assert.Contains(t, status, "untracked.txt")
		assert.Contains(t, status, "A") // Git short status for added
	})
}

func TestRepository_HasChanges(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("git is not installed")
	}

	tmpDir := t.TempDir()
	repo := NewRepository(tmpDir)

	err := repo.Init()
	require.NoError(t, err)

	err = repo.ConfigUser("Test User", "test@example.com")
	require.NoError(t, err)

	t.Run("returns false for clean working directory", func(t *testing.T) {
		hasChanges, err := repo.HasChanges()
		require.NoError(t, err)
		assert.False(t, hasChanges)
	})

	t.Run("returns true for untracked files", func(t *testing.T) {
		// Create an untracked file
		testFile := filepath.Join(tmpDir, "changes_test.txt")
		err := os.WriteFile(testFile, []byte("changes"), 0644)
		require.NoError(t, err)

		hasChanges, err := repo.HasChanges()
		require.NoError(t, err)
		assert.True(t, hasChanges)
	})

	t.Run("returns false after committing all changes", func(t *testing.T) {
		// Commit the changes
		err := repo.AddAndCommit("Commit all changes")
		require.NoError(t, err)

		hasChanges, err := repo.HasChanges()
		require.NoError(t, err)
		assert.False(t, hasChanges)
	})

	t.Run("returns true for modified tracked files", func(t *testing.T) {
		// Modify the tracked file
		testFile := filepath.Join(tmpDir, "changes_test.txt")
		err := os.WriteFile(testFile, []byte("modified content"), 0644)
		require.NoError(t, err)

		hasChanges, err := repo.HasChanges()
		require.NoError(t, err)
		assert.True(t, hasChanges)
	})
}

func TestNewRepository(t *testing.T) {
	t.Run("creates repository with correct directory", func(t *testing.T) {
		dir := "/test/path"
		repo := NewRepository(dir)

		assert.NotNil(t, repo)
		assert.Equal(t, dir, repo.dir)
	})
}
