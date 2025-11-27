package filesystem

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriter_WriteFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content string
		wantErr bool
	}{
		{
			name:    "write simple file",
			path:    "test.txt",
			content: "hello world",
			wantErr: false,
		},
		{
			name:    "write with newlines",
			path:    "multiline.txt",
			content: "line1\nline2\nline3",
			wantErr: false,
		},
		{
			name:    "write nested file (creates directories)",
			path:    "a/b/c/nested.txt",
			content: "nested content",
			wantErr: false,
		},
		{
			name:    "overwrite existing file",
			path:    "overwrite.txt",
			content: "new content",
			wantErr: false,
		},
		{
			name:    "empty content",
			path:    "empty.txt",
			content: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup in-memory filesystem
			fs := afero.NewMemMapFs()
			writer := NewWriter(fs)

			// For overwrite test, create existing file
			if tt.name == "overwrite existing file" {
				afero.WriteFile(fs, tt.path, []byte("old content"), 0644)
			}

			// Write file
			err := writer.WriteFile(tt.path, tt.content)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify file was written
				data, err := afero.ReadFile(fs, tt.path)
				require.NoError(t, err)
				assert.Equal(t, tt.content, string(data))

				// Verify parent directories were created
				exists, err := afero.DirExists(fs, ".")
				require.NoError(t, err)
				assert.True(t, exists)
			}
		})
	}
}

func TestWriter_WriteFile_DirectoryCreation(t *testing.T) {
	fs := afero.NewMemMapFs()
	writer := NewWriter(fs)

	// Write file with deep nesting
	err := writer.WriteFile("a/b/c/d/e/file.txt", "deep content")
	require.NoError(t, err)

	// Verify all parent directories were created
	for _, dir := range []string{"a", "a/b", "a/b/c", "a/b/c/d", "a/b/c/d/e"} {
		exists, err := afero.DirExists(fs, dir)
		require.NoError(t, err)
		assert.True(t, exists, "directory %s should exist", dir)
	}

	// Verify file content
	data, err := afero.ReadFile(fs, "a/b/c/d/e/file.txt")
	require.NoError(t, err)
	assert.Equal(t, "deep content", string(data))
}

func TestWriter_WriteFile_Permissions(t *testing.T) {
	fs := afero.NewMemMapFs()
	writer := NewWriter(fs)

	// Write file
	err := writer.WriteFile("perm-test.txt", "content")
	require.NoError(t, err)

	// Check file permissions (0644)
	info, err := fs.Stat("perm-test.txt")
	require.NoError(t, err)
	assert.Equal(t, "-rw-r--r--", info.Mode().String())
}
