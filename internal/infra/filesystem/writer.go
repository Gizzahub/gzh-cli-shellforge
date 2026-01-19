package filesystem

import (
	"path/filepath"

	"github.com/spf13/afero"
)

// Writer implements file writing operations.
type Writer struct {
	fs afero.Fs
}

// NewWriter creates a new filesystem writer.
func NewWriter(fs afero.Fs) *Writer {
	return &Writer{fs: fs}
}

// WriteFile writes content to a file, creating parent directories if needed.
func (w *Writer) WriteFile(path string, content string) error {
	// Create parent directories
	dir := filepath.Dir(path)
	if err := w.fs.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Write file
	return afero.WriteFile(w.fs, path, []byte(content), 0o644)
}

// Copy copies a file from src to dst, creating parent directories if needed.
func (w *Writer) Copy(src, dst string) error {
	// Read source file
	data, err := afero.ReadFile(w.fs, src)
	if err != nil {
		return err
	}

	// Create parent directories for destination
	dir := filepath.Dir(dst)
	if err := w.fs.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Write to destination
	return afero.WriteFile(w.fs, dst, data, 0o644)
}
