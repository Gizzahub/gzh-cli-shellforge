package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// MockFileReader is a mock implementation of FileReader
type MockFileReader struct {
	mock.Mock
}

func (m *MockFileReader) ReadFile(path string) (string, error) {
	args := m.Called(path)
	return args.String(0), args.Error(1)
}

func (m *MockFileReader) FileExists(path string) bool {
	args := m.Called(path)
	return args.Bool(0)
}

func TestNewMigrationService(t *testing.T) {
	reader := &MockFileReader{}
	writer := &MockFileWriter{}

	service := NewMigrationService(reader, writer)

	assert.NotNil(t, service)
	assert.Equal(t, reader, service.reader)
	assert.Equal(t, writer, service.writer)
}

func TestMigrationService_Analyze(t *testing.T) {
	t.Run("analyzes RC file with explicit section headers", func(t *testing.T) {
		reader := &MockFileReader{}
		writer := &MockFileWriter{}
		service := NewMigrationService(reader, writer)

		rcContent := `# --- PATH Setup ---
export PATH=/usr/local/bin:$PATH

# === NVM Init ===
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"

# --- My Aliases ---
alias ll='ls -la'
alias gs='git status'
`

		reader.On("ReadFile", "/path/to/.zshrc").Return(rcContent, nil)

		result, err := service.Analyze("/path/to/.zshrc")

		require.NoError(t, err)
		assert.Equal(t, "/path/to/.zshrc", result.SourceFile)
		assert.Equal(t, 3, len(result.Sections))
		assert.Equal(t, 3, result.ModulesCreated)

		// Check first section
		assert.Equal(t, "PATH Setup", result.Sections[0].Name)
		assert.Equal(t, domain.CategoryInitD, result.Sections[0].Category)

		// Check second section
		assert.Equal(t, "NVM Init", result.Sections[1].Name)
		assert.Equal(t, domain.CategoryRcPreD, result.Sections[1].Category)

		// Check third section
		assert.Equal(t, "My Aliases", result.Sections[2].Name)
		assert.Equal(t, domain.CategoryRcPostD, result.Sections[2].Category)

		reader.AssertExpectations(t)
	})

	t.Run("handles RC file without section headers", func(t *testing.T) {
		reader := &MockFileReader{}
		writer := &MockFileWriter{}
		service := NewMigrationService(reader, writer)

		rcContent := `export PATH=/usr/local/bin:$PATH
alias ll='ls -la'
`

		reader.On("ReadFile", "/path/to/.bashrc").Return(rcContent, nil)

		result, err := service.Analyze("/path/to/.bashrc")

		require.NoError(t, err)
		assert.Equal(t, 1, len(result.Sections))
		assert.Equal(t, "main-config", result.Sections[0].Name)

		reader.AssertExpectations(t)
	})

	t.Run("returns error when reading file fails", func(t *testing.T) {
		reader := &MockFileReader{}
		writer := &MockFileWriter{}
		service := NewMigrationService(reader, writer)

		reader.On("ReadFile", "/nonexistent").Return("", assert.AnError)

		result, err := service.Analyze("/nonexistent")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to read RC file")

		reader.AssertExpectations(t)
	})
}

func TestMigrationService_Migrate(t *testing.T) {
	t.Run("migrates RC file successfully", func(t *testing.T) {
		reader := &MockFileReader{}
		writer := &MockFileWriter{}
		service := NewMigrationService(reader, writer)

		rcContent := `# --- PATH Setup ---
export PATH=/usr/local/bin:$PATH

# --- My Aliases ---
alias ll='ls -la'
`

		reader.On("ReadFile", "/path/to/.zshrc").Return(rcContent, nil)

		// Expect module files to be written
		writer.On("WriteFile", "modules/init.d/10-path-setup.sh", mock.MatchedBy(func(content string) bool {
			return strings.Contains(content, "#!/bin/bash") &&
				strings.Contains(content, "# Module: path-setup") &&
				strings.Contains(content, "export PATH=/usr/local/bin:$PATH")
		})).Return(nil)

		writer.On("WriteFile", "modules/rc_post.d/my-aliases.sh", mock.MatchedBy(func(content string) bool {
			return strings.Contains(content, "#!/bin/bash") &&
				strings.Contains(content, "# Module: my-aliases") &&
				strings.Contains(content, "alias ll='ls -la'")
		})).Return(nil)

		// Expect manifest to be written
		writer.On("WriteFile", "manifest.yaml", mock.MatchedBy(func(content string) bool {
			return strings.Contains(content, "modules:") &&
				strings.Contains(content, "name: path-setup") &&
				strings.Contains(content, "name: my-aliases")
		})).Return(nil)

		result, err := service.Migrate("/path/to/.zshrc", "modules", "manifest.yaml")

		require.NoError(t, err)
		assert.Equal(t, "/path/to/.zshrc", result.SourceFile)
		assert.Equal(t, 2, result.ModulesCreated)
		assert.Equal(t, "manifest.yaml", result.ManifestPath)
		assert.Equal(t, 2, len(result.ModuleFilesPaths))

		reader.AssertExpectations(t)
		writer.AssertExpectations(t)
	})

	t.Run("returns error when writing module file fails", func(t *testing.T) {
		reader := &MockFileReader{}
		writer := &MockFileWriter{}
		service := NewMigrationService(reader, writer)

		rcContent := `# --- PATH Setup ---
export PATH=/usr/local/bin:$PATH
`

		reader.On("ReadFile", "/path/to/.zshrc").Return(rcContent, nil)
		writer.On("WriteFile", mock.Anything, mock.Anything).Return(assert.AnError)

		result, err := service.Migrate("/path/to/.zshrc", "modules", "manifest.yaml")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to write module file")

		reader.AssertExpectations(t)
		writer.AssertExpectations(t)
	})
}

func TestMigrationService_extractSections(t *testing.T) {
	service := NewMigrationService(nil, nil)

	t.Run("extracts sections with dash headers", func(t *testing.T) {
		content := `# --- Section 1 ---
content1

# --- Section 2 ---
content2
`
		sections := service.extractSections(content)

		assert.Equal(t, 2, len(sections))
		assert.Equal(t, "Section 1", sections[0].Name)
		assert.Contains(t, sections[0].Content, "content1")
		assert.Equal(t, "Section 2", sections[1].Name)
		assert.Contains(t, sections[1].Content, "content2")
	})

	t.Run("extracts sections with equals headers", func(t *testing.T) {
		content := `# === Section A ===
contentA

# === Section B ===
contentB
`
		sections := service.extractSections(content)

		assert.Equal(t, 2, len(sections))
		assert.Equal(t, "Section A", sections[0].Name)
		assert.Equal(t, "Section B", sections[1].Name)
	})

	t.Run("treats entire file as one section when no headers", func(t *testing.T) {
		content := `export PATH=/usr/local/bin:$PATH
alias ll='ls -la'
`
		sections := service.extractSections(content)

		assert.Equal(t, 1, len(sections))
		assert.Equal(t, "main-config", sections[0].Name)
		assert.Contains(t, sections[0].Content, "export PATH")
		assert.Contains(t, sections[0].Content, "alias ll")
	})

	t.Run("skips empty sections", func(t *testing.T) {
		content := `# --- Section 1 ---
content1

# --- Empty Section ---

# --- Section 2 ---
content2
`
		sections := service.extractSections(content)

		assert.Equal(t, 2, len(sections))
		assert.Equal(t, "Section 1", sections[0].Name)
		assert.Equal(t, "Section 2", sections[1].Name)
	})
}

func TestMigrationService_generateModuleContent(t *testing.T) {
	service := NewMigrationService(nil, nil)

	t.Run("generates module content with full metadata", func(t *testing.T) {
		module := domain.Module{
			Name:        "path-setup",
			File:        "init.d/10-path-setup.sh",
			Requires:    []string{"os-detection"},
			OS:          []string{"Mac", "Linux"},
			Description: "Initialize PATH",
		}

		content := "export PATH=/usr/local/bin:$PATH"

		result := service.generateModuleContent(module, content)

		assert.Contains(t, result, "#!/bin/bash")
		assert.Contains(t, result, "# Module: path-setup")
		assert.Contains(t, result, "# Description: Initialize PATH")
		assert.Contains(t, result, "# Requires: os-detection")
		assert.Contains(t, result, "# OS: Mac, Linux")
		assert.Contains(t, result, "export PATH=/usr/local/bin:$PATH")
		assert.True(t, strings.HasSuffix(result, "\n"))
	})

	t.Run("generates minimal module content", func(t *testing.T) {
		module := domain.Module{
			Name: "simple",
			File: "rc_post.d/simple.sh",
		}

		content := "alias ll='ls -la'"

		result := service.generateModuleContent(module, content)

		assert.Contains(t, result, "#!/bin/bash")
		assert.Contains(t, result, "# Module: simple")
		assert.NotContains(t, result, "# Description:")
		assert.NotContains(t, result, "# Requires:")
		assert.NotContains(t, result, "# OS:")
		assert.Contains(t, result, "alias ll='ls -la'")
	})
}

func TestMigrationService_generateManifestYAML(t *testing.T) {
	service := NewMigrationService(nil, nil)

	t.Run("generates manifest with multiple modules", func(t *testing.T) {
		manifest := &domain.Manifest{
			Modules: []domain.Module{
				{
					Name:        "os-detection",
					File:        "init.d/00-os-detection.sh",
					Requires:    []string{},
					OS:          []string{"Mac", "Linux"},
					Description: "Detect OS",
				},
				{
					Name:     "nvm",
					File:     "rc_pre.d/nvm.sh",
					Requires: []string{"os-detection"},
					OS:       []string{"Mac", "Linux"},
				},
			},
		}

		result := service.generateManifestYAML(manifest)

		assert.Contains(t, result, "modules:")
		assert.Contains(t, result, "name: os-detection")
		assert.Contains(t, result, "file: init.d/00-os-detection.sh")
		assert.Contains(t, result, "requires: []")
		assert.Contains(t, result, "description: Detect OS")
		assert.Contains(t, result, "name: nvm")
		assert.Contains(t, result, "file: rc_pre.d/nvm.sh")
		assert.Contains(t, result, "- os-detection")
	})
}
