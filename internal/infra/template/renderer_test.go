package template

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

func TestRenderer_Render(t *testing.T) {
	renderer := NewRenderer()

	t.Run("renders path template", func(t *testing.T) {
		template := getPathTemplate()
		data := &domain.TemplateData{
			ModuleName: "my-bin",
			Fields: map[string]string{
				"path_dir": "/usr/local/bin",
			},
		}

		content, err := renderer.Render(template, data)
		require.NoError(t, err)

		assert.Contains(t, content, "/usr/local/bin")
		assert.Contains(t, content, "export PATH")
	})

	t.Run("renders env template", func(t *testing.T) {
		template := getEnvTemplate()
		data := &domain.TemplateData{
			ModuleName: "editor",
			Fields: map[string]string{
				"var_name":  "EDITOR",
				"var_value": "vim",
			},
		}

		content, err := renderer.Render(template, data)
		require.NoError(t, err)

		assert.Contains(t, content, "export EDITOR=\"vim\"")
	})

	t.Run("renders with custom description", func(t *testing.T) {
		template := getPathTemplate()
		data := &domain.TemplateData{
			ModuleName:  "my-bin",
			Description: "My custom bin directory",
			Fields: map[string]string{
				"path_dir": "/usr/local/bin",
			},
		}

		content, err := renderer.Render(template, data)
		require.NoError(t, err)
		assert.NotEmpty(t, content)
	})

	t.Run("uses template description when data description is empty", func(t *testing.T) {
		template := getPathTemplate()
		data := &domain.TemplateData{
			ModuleName: "my-bin",
			Fields: map[string]string{
				"path_dir": "/usr/local/bin",
			},
		}

		content, err := renderer.Render(template, data)
		require.NoError(t, err)
		assert.NotEmpty(t, content)
	})

	t.Run("returns error for missing required field", func(t *testing.T) {
		template := getPathTemplate()
		data := &domain.TemplateData{
			ModuleName: "my-bin",
			Fields:     map[string]string{},
		}

		_, err := renderer.Render(template, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required field 'path_dir' is missing")
	})
}

func TestRenderer_RenderModuleFile(t *testing.T) {
	renderer := NewRenderer()

	t.Run("renders complete module file with header", func(t *testing.T) {
		template := getPathTemplate()
		data := &domain.TemplateData{
			ModuleName:  "my-bin",
			Description: "My custom bin directory",
			Fields: map[string]string{
				"path_dir": "/usr/local/bin",
			},
		}

		content, err := renderer.RenderModuleFile(template, data)
		require.NoError(t, err)

		// Check header
		assert.True(t, strings.HasPrefix(content, "#!/bin/bash\n"))
		assert.Contains(t, content, "# my-bin")
		assert.Contains(t, content, "# My custom bin directory")

		// Check content
		assert.Contains(t, content, "/usr/local/bin")
		assert.Contains(t, content, "export PATH")
	})

	t.Run("includes dependencies in header", func(t *testing.T) {
		template := getPathTemplate()
		data := &domain.TemplateData{
			ModuleName: "my-bin",
			Fields: map[string]string{
				"path_dir": "/usr/local/bin",
			},
			Requires: []string{"os-detection", "brew-path"},
		}

		content, err := renderer.RenderModuleFile(template, data)
		require.NoError(t, err)

		assert.Contains(t, content, "# Requires: os-detection, brew-path")
	})

	t.Run("no requires line when no dependencies", func(t *testing.T) {
		template := getPathTemplate()
		data := &domain.TemplateData{
			ModuleName: "my-bin",
			Fields: map[string]string{
				"path_dir": "/usr/local/bin",
			},
		}

		content, err := renderer.RenderModuleFile(template, data)
		require.NoError(t, err)

		assert.NotContains(t, content, "# Requires:")
	})
}
