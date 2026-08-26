package domain

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PrereqLookup checks whether external prerequisites exist on the current system.
// The interface is small so tests can inject a mock without touching the filesystem.
type PrereqLookup interface {
	HasBinary(name string) bool
	PathExists(expanded string) bool
}

// OsPrereqLookup is the production implementation backed by exec.LookPath / os.Stat.
type OsPrereqLookup struct{}

// HasBinary reports whether name resolves to an executable on PATH.
func (OsPrereqLookup) HasBinary(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// PathExists reports whether expanded identifies an existing filesystem path.
func (OsPrereqLookup) PathExists(expanded string) bool {
	_, err := os.Stat(expanded)
	return err == nil
}

// ExpandPath expands ~ and $VAR references in a path string.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	return os.ExpandEnv(path)
}
