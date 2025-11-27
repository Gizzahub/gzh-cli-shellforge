package filesystem

import (
	"github.com/spf13/afero"
)

// Reader implements file reading operations.
type Reader struct {
	fs afero.Fs
}

// NewReader creates a new filesystem reader.
func NewReader(fs afero.Fs) *Reader {
	return &Reader{fs: fs}
}

// ReadFile reads the entire contents of a file.
func (r *Reader) ReadFile(path string) (string, error) {
	data, err := afero.ReadFile(r.fs, path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FileExists checks if a file exists.
func (r *Reader) FileExists(path string) bool {
	exists, err := afero.Exists(r.fs, path)
	return err == nil && exists
}
