package filesystem

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReader_ReadFile(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(afero.Fs)
		path     string
		expected string
		wantErr  bool
	}{
		{
			name: "read existing file",
			setup: func(fs afero.Fs) {
				afero.WriteFile(fs, "test.txt", []byte("hello world"), 0644)
			},
			path:     "test.txt",
			expected: "hello world",
			wantErr:  false,
		},
		{
			name: "read file with newlines",
			setup: func(fs afero.Fs) {
				afero.WriteFile(fs, "multiline.txt", []byte("line1\nline2\nline3"), 0644)
			},
			path:     "multiline.txt",
			expected: "line1\nline2\nline3",
			wantErr:  false,
		},
		{
			name:    "file not found",
			setup:   func(fs afero.Fs) {},
			path:    "nonexistent.txt",
			wantErr: true,
		},
		{
			name: "read nested file",
			setup: func(fs afero.Fs) {
				fs.MkdirAll("a/b/c", 0755)
				afero.WriteFile(fs, "a/b/c/nested.txt", []byte("nested content"), 0644)
			},
			path:     "a/b/c/nested.txt",
			expected: "nested content",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup in-memory filesystem
			fs := afero.NewMemMapFs()
			tt.setup(fs)

			// Create reader
			reader := NewReader(fs)

			// Read file
			content, err := reader.ReadFile(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, content)
			}
		})
	}
}

func TestReader_FileExists(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(afero.Fs)
		path     string
		expected bool
	}{
		{
			name: "file exists",
			setup: func(fs afero.Fs) {
				afero.WriteFile(fs, "exists.txt", []byte("content"), 0644)
			},
			path:     "exists.txt",
			expected: true,
		},
		{
			name:     "file does not exist",
			setup:    func(fs afero.Fs) {},
			path:     "missing.txt",
			expected: false,
		},
		{
			name: "directory exists",
			setup: func(fs afero.Fs) {
				fs.MkdirAll("testdir", 0755)
			},
			path:     "testdir",
			expected: true,
		},
		{
			name: "nested file exists",
			setup: func(fs afero.Fs) {
				fs.MkdirAll("a/b", 0755)
				afero.WriteFile(fs, "a/b/file.txt", []byte("test"), 0644)
			},
			path:     "a/b/file.txt",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup in-memory filesystem
			fs := afero.NewMemMapFs()
			tt.setup(fs)

			// Create reader
			reader := NewReader(fs)

			// Check existence
			exists := reader.FileExists(tt.path)
			assert.Equal(t, tt.expected, exists)
		})
	}
}

// TestReader_RealFile tests reading actual module files
func TestReader_RealFile(t *testing.T) {
	// Test with real OS filesystem
	fs := afero.NewOsFs()
	reader := NewReader(fs)

	// Try to read an example module (skip if not exists)
	examplePath := "../../../examples/modules/init.d/00-os-detection.sh"
	if !reader.FileExists(examplePath) {
		t.Skip("Example module not found, skipping")
	}

	content, err := reader.ReadFile(examplePath)
	require.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.Contains(t, content, "#!/bin/bash")
}
