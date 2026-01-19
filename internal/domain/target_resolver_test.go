package domain

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTargetResolver_Resolve(t *testing.T) {
	tests := []struct {
		name      string
		shellType string
		homeDir   string
		target    string
		want      string
		wantErr   bool
	}{
		// zsh targets
		{
			name:      "zsh zshrc",
			shellType: "zsh",
			homeDir:   "/home/user",
			target:    "zshrc",
			want:      "/home/user/.zshrc",
			wantErr:   false,
		},
		{
			name:      "zsh zprofile",
			shellType: "zsh",
			homeDir:   "/home/user",
			target:    "zprofile",
			want:      "/home/user/.zprofile",
			wantErr:   false,
		},
		{
			name:      "zsh zshenv",
			shellType: "zsh",
			homeDir:   "/home/user",
			target:    "zshenv",
			want:      "/home/user/.zshenv",
			wantErr:   false,
		},
		// bash targets
		{
			name:      "bash bashrc",
			shellType: "bash",
			homeDir:   "/home/user",
			target:    "bashrc",
			want:      "/home/user/.bashrc",
			wantErr:   false,
		},
		{
			name:      "bash bash_profile",
			shellType: "bash",
			homeDir:   "/home/user",
			target:    "bash_profile",
			want:      "/home/user/.bash_profile",
			wantErr:   false,
		},
		// fish targets
		{
			name:      "fish config",
			shellType: "fish",
			homeDir:   "/home/user",
			target:    "config",
			want:      "/home/user/.config/fish/config.fish",
			wantErr:   false,
		},
		// Error cases
		{
			name:      "invalid target for shell",
			shellType: "zsh",
			homeDir:   "/home/user",
			target:    "bashrc",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "invalid shell type",
			shellType: "unknown",
			homeDir:   "/home/user",
			target:    "zshrc",
			want:      "",
			wantErr:   true,
		},
		// Case insensitivity
		{
			name:      "case insensitive target",
			shellType: "zsh",
			homeDir:   "/home/user",
			target:    "ZSHRC",
			want:      "/home/user/.zshrc",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewTargetResolver(tt.shellType, tt.homeDir)
			got, err := resolver.Resolve(tt.target)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTargetResolver_GetValidTargets(t *testing.T) {
	tests := []struct {
		name      string
		shellType string
		want      []string
	}{
		{
			name:      "zsh targets",
			shellType: "zsh",
			want:      []string{"zshrc", "zprofile", "zshenv", "zlogin", "zlogout", "profile"},
		},
		{
			name:      "bash targets",
			shellType: "bash",
			want:      []string{"bashrc", "bash_profile", "profile", "bash_login", "bash_logout"},
		},
		{
			name:      "fish targets",
			shellType: "fish",
			want:      []string{"config", "conf.d"},
		},
		{
			name:      "unknown shell",
			shellType: "unknown",
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewTargetResolver(tt.shellType, "/home/user")
			got := resolver.GetValidTargets()

			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				// Check all expected targets are present (order may vary)
				assert.Len(t, got, len(tt.want))
				for _, target := range tt.want {
					assert.Contains(t, got, target)
				}
			}
		})
	}
}

func TestTargetResolver_IsValidTarget(t *testing.T) {
	tests := []struct {
		name      string
		shellType string
		target    string
		want      bool
	}{
		{
			name:      "valid zsh target",
			shellType: "zsh",
			target:    "zshrc",
			want:      true,
		},
		{
			name:      "invalid target for zsh",
			shellType: "zsh",
			target:    "bashrc",
			want:      false,
		},
		{
			name:      "case insensitive",
			shellType: "zsh",
			target:    "ZSHRC",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewTargetResolver(tt.shellType, "/home/user")
			got := resolver.IsValidTarget(tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTargetResolver_ValidateTargets(t *testing.T) {
	tests := []struct {
		name      string
		shellType string
		modules   []Module
		wantErr   bool
	}{
		{
			name:      "all valid targets",
			shellType: "zsh",
			modules: []Module{
				{Name: "a", Target: "zshrc"},
				{Name: "b", Target: "zprofile"},
			},
			wantErr: false,
		},
		{
			name:      "default target (empty) is valid",
			shellType: "zsh",
			modules: []Module{
				{Name: "a", Target: ""},
			},
			wantErr: false,
		},
		{
			name:      "invalid target",
			shellType: "zsh",
			modules: []Module{
				{Name: "a", Target: "bashrc"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewTargetResolver(tt.shellType, "/home/user")
			err := resolver.ValidateTargets(tt.modules)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTargetResolver_GetDefaultTarget(t *testing.T) {
	tests := []struct {
		name      string
		shellType string
		want      string
	}{
		{
			name:      "zsh default",
			shellType: "zsh",
			want:      "zshrc",
		},
		{
			name:      "bash default",
			shellType: "bash",
			want:      "bashrc",
		},
		{
			name:      "fish default",
			shellType: "fish",
			want:      "config",
		},
		{
			name:      "unknown shell",
			shellType: "unknown",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewTargetResolver(tt.shellType, "/home/user")
			got := resolver.GetDefaultTarget()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetTargetDescription(t *testing.T) {
	tests := []struct {
		target string
		want   string
	}{
		{
			target: "zshrc",
			want:   "Interactive shell configuration",
		},
		{
			target: "zprofile",
			want:   "Login shell configuration",
		},
		{
			target: "conf.d",
			want:   "Fish modular configs",
		},
		{
			target: "unknown",
			want:   "Target file: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			got := GetTargetDescription(tt.target)
			assert.Contains(t, got, tt.want)
		})
	}
}

func TestTargetResolver_FishConfD(t *testing.T) {
	resolver := NewTargetResolver("fish", "/home/user")

	// Test conf.d target resolution
	got, err := resolver.Resolve("conf.d")
	require.NoError(t, err)
	expected := filepath.Join("/home/user", ".config", "fish", "conf.d")
	assert.Equal(t, expected, got)

	// Test conf.d relative path
	relPath, err := resolver.GetRelativePath("conf.d")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(".config", "fish", "conf.d"), relPath)

	// Test conf.d is valid target
	assert.True(t, resolver.IsValidTarget("conf.d"))
}

func TestTargetResolver_IsDirectoryTarget(t *testing.T) {
	resolver := NewTargetResolver("fish", "/home/user")

	tests := []struct {
		target string
		isDir  bool
	}{
		{"conf.d", true},
		{"config", false},
		{"zshrc", false},
		{"bashrc", false},
		{"CONF.D", true}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			assert.Equal(t, tt.isDir, resolver.IsDirectoryTarget(tt.target))
		})
	}
}

func TestTargetResolver_XDGConfigHome(t *testing.T) {
	tests := []struct {
		name           string
		xdgConfigHome  string
		homeDir        string
		wantConfigPath string
		wantConfDPath  string
	}{
		{
			name:           "default (no XDG_CONFIG_HOME)",
			xdgConfigHome:  "",
			homeDir:        "/home/user",
			wantConfigPath: "/home/user/.config/fish/config.fish",
			wantConfDPath:  "/home/user/.config/fish/conf.d",
		},
		{
			name:           "relative XDG_CONFIG_HOME",
			xdgConfigHome:  ".myconfig",
			homeDir:        "/home/user",
			wantConfigPath: "/home/user/.myconfig/fish/config.fish",
			wantConfDPath:  "/home/user/.myconfig/fish/conf.d",
		},
		{
			name:           "absolute XDG_CONFIG_HOME inside homeDir",
			xdgConfigHome:  "/home/user/.customconfig",
			homeDir:        "/home/user",
			wantConfigPath: "/home/user/.customconfig/fish/config.fish",
			wantConfDPath:  "/home/user/.customconfig/fish/conf.d",
		},
		{
			name:           "absolute XDG_CONFIG_HOME outside homeDir (falls back to default)",
			xdgConfigHome:  "/opt/config",
			homeDir:        "/home/user",
			wantConfigPath: "/home/user/.config/fish/config.fish",
			wantConfDPath:  "/home/user/.config/fish/conf.d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original env
			origXDG := os.Getenv("XDG_CONFIG_HOME")
			defer func() {
				if origXDG == "" {
					os.Unsetenv("XDG_CONFIG_HOME")
				} else {
					os.Setenv("XDG_CONFIG_HOME", origXDG)
				}
			}()

			// Set test env
			if tt.xdgConfigHome == "" {
				os.Unsetenv("XDG_CONFIG_HOME")
			} else {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)
			}

			resolver := NewTargetResolver("fish", tt.homeDir)

			// Test config path
			configPath, err := resolver.Resolve("config")
			require.NoError(t, err)
			assert.Equal(t, tt.wantConfigPath, configPath)

			// Test conf.d path
			confDPath, err := resolver.Resolve("conf.d")
			require.NoError(t, err)
			assert.Equal(t, tt.wantConfDPath, confDPath)
		})
	}
}
