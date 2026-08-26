package domain_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

func TestExpandPath_Tilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := domain.ExpandPath("~/foo/bar")
	want := filepath.Join(home, "foo", "bar")
	if got != want {
		t.Errorf("ExpandPath: got %q, want %q", got, want)
	}
}

func TestExpandPath_EnvVar(t *testing.T) {
	t.Setenv("MYDIR", "/tmp/testdir")
	got := domain.ExpandPath("$MYDIR/sub")
	if got != "/tmp/testdir/sub" {
		t.Errorf("ExpandPath: got %q, want /tmp/testdir/sub", got)
	}
}

func TestExpandPath_NoExpansion(t *testing.T) {
	got := domain.ExpandPath("/absolute/path")
	if got != "/absolute/path" {
		t.Errorf("ExpandPath: got %q, want /absolute/path", got)
	}
}
