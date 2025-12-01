package shellmeta_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain/shellmeta"
	"github.com/spf13/afero"
)

func findProjectRoot(t *testing.T) string {
	t.Helper()

	// Start from current working directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Walk up to find data/shell-profiles
	for {
		shellProfilesPath := filepath.Join(dir, "data", "shell-profiles")
		if _, err := os.Stat(shellProfilesPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("Could not find project root with data/shell-profiles")
		}
		dir = parent
	}
}

func TestIntegration_LoadRealYAMLFiles(t *testing.T) {
	projectRoot := findProjectRoot(t)
	shellProfilesDir := filepath.Join(projectRoot, "data", "shell-profiles")

	loader := shellmeta.NewLoader(afero.NewOsFs())
	profiles, err := loader.Load(shellProfilesDir)
	if err != nil {
		t.Fatalf("Failed to load real YAML files: %v", err)
	}

	// Verify Core loaded
	if profiles.Core == nil {
		t.Error("Core should not be nil")
	}

	// Verify Contexts loaded
	if profiles.Contexts == nil {
		t.Error("Contexts should not be nil")
	}

	// Verify Dev loaded
	if profiles.Dev == nil {
		t.Error("Dev should not be nil")
	}

	// Verify Automation loaded
	if profiles.Automation == nil {
		t.Error("Automation should not be nil")
	}
}

func TestIntegration_QueryRealData(t *testing.T) {
	projectRoot := findProjectRoot(t)
	shellProfilesDir := filepath.Join(projectRoot, "data", "shell-profiles")

	loader := shellmeta.NewLoader(afero.NewOsFs())
	profiles, err := loader.Load(shellProfilesDir)
	if err != nil {
		t.Fatalf("Failed to load real YAML files: %v", err)
	}

	// Test macOS queries
	t.Run("Mac init files", func(t *testing.T) {
		files := profiles.GetInitFilesForOS("mac", "bash")
		if len(files) == 0 {
			t.Error("GetInitFilesForOS(mac, bash) should return files")
		}
	})

	t.Run("Mac default shell", func(t *testing.T) {
		shell := profiles.GetDefaultShell("mac")
		if shell != "zsh" {
			t.Errorf("GetDefaultShell(mac) = %v, want zsh", shell)
		}
	})

	// Test Linux queries
	t.Run("Ubuntu init files", func(t *testing.T) {
		files := profiles.GetInitFilesForOS("ubuntu", "bash")
		if len(files) == 0 {
			t.Error("GetInitFilesForOS(ubuntu, bash) should return files")
		}
	})

	// Test language version managers
	t.Run("rbenv", func(t *testing.T) {
		mgr := profiles.GetLanguageVersionManager("rbenv")
		if mgr == nil {
			t.Error("GetLanguageVersionManager(rbenv) should not be nil")
		}
		if mgr != nil && mgr.InitCommand == "" {
			t.Error("rbenv InitCommand should not be empty")
		}
	})

	// Test context checks
	t.Run("cron context", func(t *testing.T) {
		loaded := profiles.IsProfileLoadedInContext("cron")
		if loaded {
			t.Error("IsProfileLoadedInContext(cron) should be false")
		}
	})

	// Test shell modes
	t.Run("login shell mode", func(t *testing.T) {
		mode := profiles.GetShellMode("login")
		if mode == nil {
			t.Error("GetShellMode(login) should not be nil")
		}
	})

	// Test terminal multiplexers
	t.Run("tmux", func(t *testing.T) {
		tm := profiles.GetTerminalMultiplexer("tmux")
		if tm == nil {
			t.Error("GetTerminalMultiplexer(tmux) should not be nil")
		}
	})

	// Test desktop environments
	t.Run("gnome", func(t *testing.T) {
		de := profiles.GetDesktopEnvironment("gnome")
		if de == nil {
			t.Error("GetDesktopEnvironment(gnome) should not be nil")
		}
	})

	// Test list functions
	t.Run("list distributions", func(t *testing.T) {
		distros := profiles.ListSupportedDistributions()
		if len(distros) == 0 {
			t.Error("ListSupportedDistributions() should not be empty")
		}
	})

	t.Run("list language managers", func(t *testing.T) {
		managers := profiles.ListLanguageVersionManagers()
		if len(managers) == 0 {
			t.Error("ListLanguageVersionManagers() should not be empty")
		}
	})

	t.Run("list desktop environments", func(t *testing.T) {
		des := profiles.ListDesktopEnvironments()
		if len(des) == 0 {
			t.Error("ListDesktopEnvironments() should not be empty")
		}
	})
}
