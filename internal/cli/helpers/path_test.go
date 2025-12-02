package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandHomePath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name: "empty path",
			path: "",
			want: "",
		},
		{
			name: "absolute path",
			path: "/usr/local/bin",
			want: "/usr/local/bin",
		},
		{
			name: "relative path",
			path: "relative/path",
			want: "relative/path",
		},
		{
			name: "tilde only",
			path: "~",
			want: home,
		},
		{
			name: "tilde with path",
			path: "~/.zshrc",
			want: filepath.Join(home, ".zshrc"),
		},
		{
			name: "tilde with nested path",
			path: "~/config/shell/modules",
			want: filepath.Join(home, "config/shell/modules"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandHomePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveBackupDir(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name      string
		specified string
		want      string
		wantErr   bool
	}{
		{
			name:      "empty uses default",
			specified: "",
			want:      filepath.Join(home, ".backup", "shellforge"),
		},
		{
			name:      "absolute path",
			specified: "/var/backups/shell",
			want:      "/var/backups/shell",
		},
		{
			name:      "tilde path",
			specified: "~/my-backups",
			want:      filepath.Join(home, "my-backups"),
		},
		{
			name:      "relative path",
			specified: "backups",
			want:      "backups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveBackupDir(tt.specified)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDefaultBackupDir(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	got, err := DefaultBackupDir()
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(got, home))
	assert.Contains(t, got, ".backup")
	assert.Contains(t, got, "shellforge")
}
