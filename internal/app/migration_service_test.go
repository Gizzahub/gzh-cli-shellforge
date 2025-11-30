package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRCParser is a mock implementation of RCParser for testing
type mockRCParser struct {
	result *domain.MigrationResult
	err    error
}

func (m *mockRCParser) ParseFile(path string) (*domain.MigrationResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

// mockFileReader is a simple mock for FileReader
type mockFileReader struct {
	fs afero.Fs
}

func (m *mockFileReader) ReadFile(path string) (string, error) {
	data, err := afero.ReadFile(m.fs, path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (m *mockFileReader) FileExists(path string) bool {
	exists, _ := afero.Exists(m.fs, path)
	return exists
}

// mockFileWriter is a simple mock for FileWriter
type mockFileWriter struct {
	fs afero.Fs
}

func (m *mockFileWriter) WriteFile(path string, content string) error {
	return afero.WriteFile(m.fs, path, []byte(content), 0644)
}

func TestNewMigrationService(t *testing.T) {
	fs := afero.NewMemMapFs()
	parser := &mockRCParser{}
	reader := &mockFileReader{fs: fs}
	writer := &mockFileWriter{fs: fs}

	service := NewMigrationService(parser, reader, writer)

	require.NotNil(t, service)
	assert.NotNil(t, service.parser)
	assert.NotNil(t, service.reader)
	assert.NotNil(t, service.writer)
}

func TestMigrationService_Analyze(t *testing.T) {
	tests := []struct {
		name          string
		rcFilePath    string
		parserResult  *domain.MigrationResult
		parserError   error
		expectError   bool
		errorContains string
		validate      func(t *testing.T, result *MigrateResult)
	}{
		{
			name:       "successful analysis",
			rcFilePath: "/home/user/.zshrc",
			parserResult: &domain.MigrationResult{
				Sections: []domain.Section{
					{
						Name:        "PATH Setup",
						Content:     "export PATH=/usr/local/bin:$PATH",
						Category:    "init.d",
						LineStart:   1,
						LineEnd:     3,
						Description: "Path configuration",
					},
					{
						Name:        "Aliases",
						Content:     "alias ll='ls -la'",
						Category:    "rc_post.d",
						LineStart:   5,
						LineEnd:     7,
						Description: "Command aliases",
					},
				},
				Modules: []domain.Module{
					{
						Name:        "path-setup",
						File:        "init.d/10-path-setup.sh",
						Description: "Path configuration",
					},
					{
						Name:        "aliases",
						File:        "rc_post.d/50-aliases.sh",
						Description: "Command aliases",
					},
				},
				Manifest: &domain.Manifest{
					Modules: []domain.Module{
						{Name: "path-setup", File: "init.d/10-path-setup.sh"},
						{Name: "aliases", File: "rc_post.d/50-aliases.sh"},
					},
				},
				Warnings: []string{},
			},
			expectError: false,
			validate: func(t *testing.T, result *MigrateResult) {
				assert.Equal(t, "/home/user/.zshrc", result.SourceFile)
				assert.Len(t, result.Sections, 2)
				assert.Equal(t, 2, result.ModulesCreated)
				assert.Len(t, result.Warnings, 0)
			},
		},
		{
			name:       "analysis with warnings",
			rcFilePath: "/home/user/.bashrc",
			parserResult: &domain.MigrationResult{
				Sections: []domain.Section{
					{Name: "Config", Content: "export VAR=value", Category: "init.d"},
				},
				Modules: []domain.Module{
					{Name: "config", File: "init.d/10-config.sh"},
				},
				Manifest: &domain.Manifest{
					Modules: []domain.Module{{Name: "config", File: "init.d/10-config.sh"}},
				},
				Warnings: []string{"some unsegmented content found"},
			},
			expectError: false,
			validate: func(t *testing.T, result *MigrateResult) {
				assert.Len(t, result.Warnings, 1)
				assert.Contains(t, result.Warnings[0], "unsegmented")
			},
		},
		{
			name:          "parser error",
			rcFilePath:    "/invalid/path",
			parserError:   fmt.Errorf("file not found"),
			expectError:   true,
			errorContains: "failed to parse RC file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			parser := &mockRCParser{result: tt.parserResult, err: tt.parserError}
			reader := &mockFileReader{fs: fs}
			writer := &mockFileWriter{fs: fs}

			service := NewMigrationService(parser, reader, writer)

			result, err := service.Analyze(tt.rcFilePath)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestMigrationService_Migrate(t *testing.T) {
	tests := []struct {
		name          string
		rcFilePath    string
		outputDir     string
		manifestPath  string
		parserResult  *domain.MigrationResult
		parserError   error
		expectError   bool
		errorContains string
		validate      func(t *testing.T, fs afero.Fs, result *MigrateResult)
	}{
		{
			name:         "successful migration",
			rcFilePath:   "/home/user/.zshrc",
			outputDir:    "/output/modules",
			manifestPath: "/output/manifest.yaml",
			parserResult: &domain.MigrationResult{
				Sections: []domain.Section{
					{
						Name:        "PATH Setup",
						Content:     "export PATH=/usr/local/bin:$PATH\n",
						Category:    "init.d",
						Description: "Path configuration",
					},
				},
				Modules: []domain.Module{
					{
						Name:        "path-setup",
						File:        "init.d/10-path-setup.sh",
						Description: "Path configuration",
						OS:          []string{"Mac", "Linux"},
						Requires:    []string{},
					},
				},
				Manifest: &domain.Manifest{
					Modules: []domain.Module{
						{
							Name:        "path-setup",
							File:        "init.d/10-path-setup.sh",
							Description: "Path configuration",
							OS:          []string{"Mac", "Linux"},
							Requires:    []string{},
						},
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, fs afero.Fs, result *MigrateResult) {
				// Verify result
				assert.Equal(t, "/home/user/.zshrc", result.SourceFile)
				assert.Equal(t, 1, result.ModulesCreated)
				assert.Equal(t, "/output/manifest.yaml", result.ManifestPath)
				assert.Len(t, result.ModuleFilesPaths, 1)

				// Verify module file was created
				moduleExists, _ := afero.Exists(fs, "/output/modules/init.d/10-path-setup.sh")
				assert.True(t, moduleExists, "module file should be created")

				// Verify module content
				moduleContent, _ := afero.ReadFile(fs, "/output/modules/init.d/10-path-setup.sh")
				contentStr := string(moduleContent)
				assert.Contains(t, contentStr, "#!/bin/bash")
				assert.Contains(t, contentStr, "# Module: path-setup")
				assert.Contains(t, contentStr, "# Description: Path configuration")
				assert.Contains(t, contentStr, "# OS: Mac, Linux")
				assert.Contains(t, contentStr, "export PATH=/usr/local/bin:$PATH")

				// Verify manifest file was created
				manifestExists, _ := afero.Exists(fs, "/output/manifest.yaml")
				assert.True(t, manifestExists, "manifest file should be created")

				// Verify manifest content
				manifestContent, _ := afero.ReadFile(fs, "/output/manifest.yaml")
				manifestStr := string(manifestContent)
				assert.Contains(t, manifestStr, "modules:")
				assert.Contains(t, manifestStr, "name: path-setup")
				assert.Contains(t, manifestStr, "file: init.d/10-path-setup.sh")
				assert.Contains(t, manifestStr, "- Mac")
				assert.Contains(t, manifestStr, "- Linux")
			},
		},
		{
			name:         "migration with no sections - warning added",
			rcFilePath:   "/home/user/.zshrc",
			outputDir:    "/output/modules",
			manifestPath: "/output/manifest.yaml",
			parserResult: &domain.MigrationResult{
				Sections: []domain.Section{},
				Modules:  []domain.Module{},
				Manifest: &domain.Manifest{Modules: []domain.Module{}},
				Warnings: []string{},
			},
			expectError: false,
			validate: func(t *testing.T, fs afero.Fs, result *MigrateResult) {
				assert.Len(t, result.Warnings, 1)
				assert.Contains(t, result.Warnings[0], "no sections detected")
			},
		},
		{
			name:         "multiple modules with dependencies",
			rcFilePath:   "/home/user/.zshrc",
			outputDir:    "/output/modules",
			manifestPath: "/output/manifest.yaml",
			parserResult: &domain.MigrationResult{
				Sections: []domain.Section{
					{Name: "PATH", Content: "export PATH=/bin:$PATH\n", Category: "init.d"},
					{Name: "NVM", Content: "export NVM_DIR=\"$HOME/.nvm\"\n", Category: "rc_pre.d"},
					{Name: "Aliases", Content: "alias gs='git status'\n", Category: "rc_post.d"},
				},
				Modules: []domain.Module{
					{Name: "path", File: "init.d/10-path.sh"},
					{Name: "nvm", File: "rc_pre.d/20-nvm.sh", Requires: []string{"path"}},
					{Name: "aliases", File: "rc_post.d/50-aliases.sh"},
				},
				Manifest: &domain.Manifest{
					Modules: []domain.Module{
						{Name: "path", File: "init.d/10-path.sh"},
						{Name: "nvm", File: "rc_pre.d/20-nvm.sh", Requires: []string{"path"}},
						{Name: "aliases", File: "rc_post.d/50-aliases.sh"},
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, fs afero.Fs, result *MigrateResult) {
				assert.Equal(t, 3, result.ModulesCreated)
				assert.Len(t, result.ModuleFilesPaths, 3)

				// Verify all module files created
				for _, path := range result.ModuleFilesPaths {
					exists, _ := afero.Exists(fs, path)
					assert.True(t, exists, "module file %s should exist", path)
				}

				// Verify manifest contains dependencies
				manifestContent, _ := afero.ReadFile(fs, "/output/manifest.yaml")
				manifestStr := string(manifestContent)
				assert.Contains(t, manifestStr, "- path")
			},
		},
		{
			name:          "parser error",
			rcFilePath:    "/invalid/path",
			outputDir:     "/output/modules",
			manifestPath:  "/output/manifest.yaml",
			parserError:   fmt.Errorf("file not found"),
			expectError:   true,
			errorContains: "failed to parse RC file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			parser := &mockRCParser{result: tt.parserResult, err: tt.parserError}
			reader := &mockFileReader{fs: fs}
			writer := &mockFileWriter{fs: fs}

			service := NewMigrationService(parser, reader, writer)

			result, err := service.Migrate(tt.rcFilePath, tt.outputDir, tt.manifestPath)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.validate != nil {
				tt.validate(t, fs, result)
			}
		})
	}
}

func TestMigrationService_generateModuleContent(t *testing.T) {
	tests := []struct {
		name     string
		module   domain.Module
		content  string
		validate func(t *testing.T, output string)
	}{
		{
			name: "module with all metadata",
			module: domain.Module{
				Name:        "git-config",
				Description: "Git configuration",
				Requires:    []string{"path-setup", "env-vars"},
				OS:          []string{"Mac", "Linux"},
			},
			content: "export GIT_EDITOR=vim\nalias gs='git status'\n",
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "#!/bin/bash")
				assert.Contains(t, output, "# Module: git-config")
				assert.Contains(t, output, "# Description: Git configuration")
				assert.Contains(t, output, "# Requires: path-setup, env-vars")
				assert.Contains(t, output, "# OS: Mac, Linux")
				assert.Contains(t, output, "export GIT_EDITOR=vim")
				assert.Contains(t, output, "alias gs='git status'")
				assert.True(t, strings.HasSuffix(output, "\n"), "should end with newline")
			},
		},
		{
			name: "module with minimal metadata",
			module: domain.Module{
				Name: "simple-module",
			},
			content: "echo 'hello'\n",
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "#!/bin/bash")
				assert.Contains(t, output, "# Module: simple-module")
				assert.NotContains(t, output, "# Description:")
				assert.NotContains(t, output, "# Requires:")
				assert.NotContains(t, output, "# OS:")
				assert.Contains(t, output, "echo 'hello'")
			},
		},
		{
			name: "content without trailing newline",
			module: domain.Module{
				Name: "test-module",
			},
			content: "export VAR=value",
			validate: func(t *testing.T, output string) {
				assert.True(t, strings.HasSuffix(output, "\n"), "should add trailing newline")
				assert.Contains(t, output, "export VAR=value")
			},
		},
		{
			name: "empty requires and OS arrays",
			module: domain.Module{
				Name:     "empty-arrays",
				Requires: []string{},
				OS:       []string{},
			},
			content: "echo 'test'\n",
			validate: func(t *testing.T, output string) {
				assert.NotContains(t, output, "# Requires:")
				assert.NotContains(t, output, "# OS:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			parser := &mockRCParser{}
			reader := &mockFileReader{fs: fs}
			writer := &mockFileWriter{fs: fs}

			service := NewMigrationService(parser, reader, writer)

			output := service.generateModuleContent(tt.module, tt.content)

			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestMigrationService_generateManifestYAML(t *testing.T) {
	tests := []struct {
		name     string
		manifest *domain.Manifest
		validate func(t *testing.T, output string)
	}{
		{
			name: "manifest with multiple modules",
			manifest: &domain.Manifest{
				Modules: []domain.Module{
					{
						Name:        "path-setup",
						File:        "init.d/10-path-setup.sh",
						Description: "PATH configuration",
						OS:          []string{"Mac", "Linux"},
						Requires:    []string{},
					},
					{
						Name:        "nvm-init",
						File:        "rc_pre.d/20-nvm.sh",
						Description: "NVM initialization",
						Requires:    []string{"path-setup"},
						OS:          []string{"Mac"},
					},
					{
						Name:     "aliases",
						File:     "rc_post.d/50-aliases.sh",
						Requires: []string{},
					},
				},
			},
			validate: func(t *testing.T, output string) {
				// Check header
				assert.Contains(t, output, "# Generated by gz-shellforge migrate")
				assert.Contains(t, output, "modules:")

				// Check first module
				assert.Contains(t, output, "- name: path-setup")
				assert.Contains(t, output, "file: init.d/10-path-setup.sh")
				assert.Contains(t, output, "requires: []")
				assert.Contains(t, output, "description: PATH configuration")
				assert.Contains(t, output, "- Mac")
				assert.Contains(t, output, "- Linux")

				// Check second module with dependency
				assert.Contains(t, output, "- name: nvm-init")
				assert.Contains(t, output, "file: rc_pre.d/20-nvm.sh")
				assert.Contains(t, output, "- path-setup")

				// Check third module
				assert.Contains(t, output, "- name: aliases")
				assert.Contains(t, output, "file: rc_post.d/50-aliases.sh")
			},
		},
		{
			name: "module without description",
			manifest: &domain.Manifest{
				Modules: []domain.Module{
					{
						Name: "simple",
						File: "init.d/10-simple.sh",
					},
				},
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "- name: simple")
				assert.Contains(t, output, "file: init.d/10-simple.sh")
				// Description line should not appear
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					if strings.Contains(line, "simple") {
						assert.NotContains(t, line, "description:")
					}
				}
			},
		},
		{
			name: "empty manifest",
			manifest: &domain.Manifest{
				Modules: []domain.Module{},
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "# Generated by gz-shellforge migrate")
				assert.Contains(t, output, "modules:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			parser := &mockRCParser{}
			reader := &mockFileReader{fs: fs}
			writer := &mockFileWriter{fs: fs}

			service := NewMigrationService(parser, reader, writer)

			output := service.generateManifestYAML(tt.manifest)

			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}
