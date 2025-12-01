package shellmeta

import (
	"testing"

	"github.com/spf13/afero"
)

func loadTestProfiles(t *testing.T) *ShellProfiles {
	t.Helper()

	fs := afero.NewMemMapFs()
	setupTestFiles(t, fs)

	loader := NewLoader(fs)
	profiles, err := loader.Load("/data/shell-profiles")
	if err != nil {
		t.Fatalf("Failed to load test profiles: %v", err)
	}
	return profiles
}

func TestShellProfiles_GetInitFilesForOS(t *testing.T) {
	profiles := loadTestProfiles(t)

	tests := []struct {
		name     string
		os       string
		shell    string
		wantLen  int
		wantNil  bool
	}{
		{"Mac bash", "mac", "bash", 2, false},
		{"Mac zsh", "Mac", "ZSH", 2, false},
		{"Ubuntu bash", "ubuntu", "bash", 2, false},
		{"Darwin alias", "darwin", "bash", 2, false},
		{"macOS alias", "macos", "bash", 2, false},
		{"Nonexistent OS", "windows", "bash", 0, true},
		{"Nonexistent shell", "mac", "powershell", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profiles.GetInitFilesForOS(tt.os, tt.shell)
			if tt.wantNil && got != nil {
				t.Errorf("GetInitFilesForOS() = %v, want nil", got)
			}
			if !tt.wantNil && len(got) != tt.wantLen {
				t.Errorf("GetInitFilesForOS() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestShellProfiles_GetLoginShellFiles(t *testing.T) {
	profiles := loadTestProfiles(t)

	tests := []struct {
		name    string
		os      string
		wantLen int
		wantNil bool
	}{
		{"Mac", "mac", 2, false},
		{"Ubuntu", "ubuntu", 2, false},
		{"Darwin alias", "darwin", 2, false},
		{"Nonexistent", "windows", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profiles.GetLoginShellFiles(tt.os)
			if tt.wantNil && got != nil {
				t.Errorf("GetLoginShellFiles() = %v, want nil", got)
			}
			if !tt.wantNil && len(got) != tt.wantLen {
				t.Errorf("GetLoginShellFiles() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestShellProfiles_GetDefaultShell(t *testing.T) {
	profiles := loadTestProfiles(t)

	tests := []struct {
		name string
		os   string
		want string
	}{
		{"Mac", "mac", "zsh"},
		{"macOS alias", "macos", "zsh"},
		{"Darwin alias", "darwin", "zsh"},
		{"Ubuntu", "ubuntu", "bash"},
		{"Nonexistent", "windows", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profiles.GetDefaultShell(tt.os)
			if got != tt.want {
				t.Errorf("GetDefaultShell() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShellProfiles_GetLanguageVersionManager(t *testing.T) {
	profiles := loadTestProfiles(t)

	tests := []struct {
		name    string
		mgr     string
		wantNil bool
	}{
		{"rbenv", "rbenv", false},
		{"nvm", "nvm", false},
		{"pyenv", "pyenv", false},
		{"conda", "conda", false},
		{"RBENV uppercase", "RBENV", false},
		{"nonexistent", "nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profiles.GetLanguageVersionManager(tt.mgr)
			if tt.wantNil && got != nil {
				t.Errorf("GetLanguageVersionManager() = %v, want nil", got)
			}
			if !tt.wantNil && got == nil {
				t.Error("GetLanguageVersionManager() = nil, want non-nil")
			}
		})
	}
}

func TestShellProfiles_GetDesktopEnvironment(t *testing.T) {
	profiles := loadTestProfiles(t)

	tests := []struct {
		name    string
		de      string
		wantNil bool
	}{
		{"gnome", "gnome", false},
		{"kde", "kde", false},
		{"GNOME uppercase", "GNOME", false},
		{"nonexistent", "windows", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profiles.GetDesktopEnvironment(tt.de)
			if tt.wantNil && got != nil {
				t.Errorf("GetDesktopEnvironment() = %v, want nil", got)
			}
			if !tt.wantNil && got == nil {
				t.Error("GetDesktopEnvironment() = nil, want non-nil")
			}
		})
	}
}

func TestShellProfiles_GetShellMode(t *testing.T) {
	profiles := loadTestProfiles(t)

	tests := []struct {
		name    string
		mode    string
		wantNil bool
	}{
		{"login_shell", "login_shell", false},
		{"login alias", "login", false},
		{"non_login_shell", "non_login_shell", false},
		{"non-login alias", "non-login", false},
		{"nonlogin alias", "nonlogin", false},
		{"interactive_shell", "interactive_shell", false},
		{"interactive alias", "interactive", false},
		{"nonexistent", "nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profiles.GetShellMode(tt.mode)
			if tt.wantNil && got != nil {
				t.Errorf("GetShellMode() = %v, want nil", got)
			}
			if !tt.wantNil && got == nil {
				t.Error("GetShellMode() = nil, want non-nil")
			}
		})
	}
}

func TestShellProfiles_GetTerminalMultiplexer(t *testing.T) {
	profiles := loadTestProfiles(t)

	tests := []struct {
		name    string
		tm      string
		wantNil bool
	}{
		{"tmux", "tmux", false},
		{"screen", "screen", false},
		{"zellij", "zellij", false},
		{"TMUX uppercase", "TMUX", false},
		{"nonexistent", "byobu", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profiles.GetTerminalMultiplexer(tt.tm)
			if tt.wantNil && got != nil {
				t.Errorf("GetTerminalMultiplexer() = %v, want nil", got)
			}
			if !tt.wantNil && got == nil {
				t.Error("GetTerminalMultiplexer() = nil, want non-nil")
			}
		})
	}
}

func TestShellProfiles_GetUserSwitch(t *testing.T) {
	profiles := loadTestProfiles(t)

	tests := []struct {
		name    string
		sw      string
		wantNil bool
	}{
		{"su", "su", false},
		{"su_login", "su_login", false},
		{"sudo", "sudo", false},
		{"SU uppercase", "SU", false},
		{"nonexistent", "doas", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profiles.GetUserSwitch(tt.sw)
			if tt.wantNil && got != nil {
				t.Errorf("GetUserSwitch() = %v, want nil", got)
			}
			if !tt.wantNil && got == nil {
				t.Error("GetUserSwitch() = nil, want non-nil")
			}
		})
	}
}

func TestShellProfiles_ListSupportedDistributions(t *testing.T) {
	profiles := loadTestProfiles(t)

	distros := profiles.ListSupportedDistributions()
	if len(distros) == 0 {
		t.Error("ListSupportedDistributions() should return at least one distro")
	}

	// Check ubuntu exists
	found := false
	for _, d := range distros {
		if d == "ubuntu" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ListSupportedDistributions() should include ubuntu")
	}
}

func TestShellProfiles_ListLanguageVersionManagers(t *testing.T) {
	profiles := loadTestProfiles(t)

	managers := profiles.ListLanguageVersionManagers()
	if len(managers) == 0 {
		t.Error("ListLanguageVersionManagers() should return at least one manager")
	}

	// Check rbenv exists
	found := false
	for _, m := range managers {
		if m == "rbenv" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ListLanguageVersionManagers() should include rbenv")
	}
}

func TestShellProfiles_ListDesktopEnvironments(t *testing.T) {
	profiles := loadTestProfiles(t)

	des := profiles.ListDesktopEnvironments()
	if len(des) == 0 {
		t.Error("ListDesktopEnvironments() should return at least one DE")
	}

	// Check gnome exists
	found := false
	for _, de := range des {
		if de == "gnome" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ListDesktopEnvironments() should include gnome")
	}
}

func TestShellProfiles_GetInteractiveShellFiles(t *testing.T) {
	profiles := loadTestProfiles(t)

	tests := []struct {
		name    string
		os      string
		wantLen int
		wantNil bool
	}{
		{"Mac", "mac", 2, false},
		{"Ubuntu", "ubuntu", 2, false},
		{"Darwin alias", "darwin", 2, false},
		{"Nonexistent", "windows", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profiles.GetInteractiveShellFiles(tt.os)
			if tt.wantNil && got != nil {
				t.Errorf("GetInteractiveShellFiles() = %v, want nil", got)
			}
			if !tt.wantNil && len(got) != tt.wantLen {
				t.Errorf("GetInteractiveShellFiles() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestShellProfiles_GetDisplayManager(t *testing.T) {
	profiles := loadTestProfiles(t)

	tests := []struct {
		name    string
		dm      string
		wantNil bool
	}{
		{"gdm", "gdm", false},
		{"GDM uppercase", "GDM", false},
		{"nonexistent", "lightdm", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profiles.GetDisplayManager(tt.dm)
			if tt.wantNil && got != nil {
				t.Errorf("GetDisplayManager() = %v, want nil", got)
			}
			if !tt.wantNil && got == nil {
				t.Error("GetDisplayManager() = nil, want non-nil")
			}
		})
	}
}

func TestShellProfiles_IsProfileLoadedInContext(t *testing.T) {
	profiles := loadTestProfiles(t)

	tests := []struct {
		name    string
		context string
		want    bool
	}{
		{"cron", "cron", false},
		{"at", "at", false},
		{"docker_exec", "docker_exec", false},
		{"docker-exec", "docker-exec", false},
		{"flatpak", "flatpak", false},
		{"git_hooks", "git_hooks", false},
		{"git-hooks", "git-hooks", false},
		{"github_actions", "github_actions", false},
		{"github-actions", "github-actions", false},
		{"gitlab_ci", "gitlab_ci", false},
		{"gitlab-ci", "gitlab-ci", false},
		{"jenkins", "jenkins", false},
		{"ssh_forced_command", "ssh_forced_command", false},
		{"ssh-forced-command", "ssh-forced-command", false},
		{"unknown context", "unknown", true}, // default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profiles.IsProfileLoadedInContext(tt.context)
			if got != tt.want {
				t.Errorf("IsProfileLoadedInContext(%q) = %v, want %v", tt.context, got, tt.want)
			}
		})
	}
}

func TestShellProfiles_NilProfiles(t *testing.T) {
	profiles := &ShellProfiles{}

	// All methods should handle nil gracefully
	if profiles.GetInitFilesForOS("mac", "bash") != nil {
		t.Error("GetInitFilesForOS() should return nil for nil Core")
	}
	if profiles.GetLoginShellFiles("mac") != nil {
		t.Error("GetLoginShellFiles() should return nil for nil Core")
	}
	if profiles.GetDefaultShell("mac") != "" {
		t.Error("GetDefaultShell() should return empty for nil Core")
	}
	if profiles.GetLanguageVersionManager("rbenv") != nil {
		t.Error("GetLanguageVersionManager() should return nil for nil Dev")
	}
	if profiles.GetDesktopEnvironment("gnome") != nil {
		t.Error("GetDesktopEnvironment() should return nil for nil Contexts")
	}
	if profiles.GetShellMode("login") != nil {
		t.Error("GetShellMode() should return nil for nil Contexts")
	}
	if profiles.GetTerminalMultiplexer("tmux") != nil {
		t.Error("GetTerminalMultiplexer() should return nil for nil Automation")
	}
	if profiles.GetUserSwitch("su") != nil {
		t.Error("GetUserSwitch() should return nil for nil Automation")
	}
	if profiles.ListSupportedDistributions() != nil {
		t.Error("ListSupportedDistributions() should return nil for nil Core")
	}
}
