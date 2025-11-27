package app

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// MockTemplateRenderer is a mock implementation of TemplateRenderer
type MockTemplateRenderer struct {
	mock.Mock
}

func (m *MockTemplateRenderer) RenderModuleFile(template *domain.Template, data *domain.TemplateData) (string, error) {
	args := m.Called(template, data)
	return args.String(0), args.Error(1)
}

// MockFileWriter is a mock implementation of FileWriter
type MockFileWriter struct {
	mock.Mock
}

func (m *MockFileWriter) WriteFile(path, content string) error {
	args := m.Called(path, content)
	return args.Error(0)
}

func TestNewTemplateService(t *testing.T) {
	renderer := &MockTemplateRenderer{}
	writer := &MockFileWriter{}

	service := NewTemplateService(renderer, writer)

	assert.NotNil(t, service)
	assert.Equal(t, renderer, service.renderer)
	assert.Equal(t, writer, service.writer)
}

func TestTemplateService_Generate(t *testing.T) {
	t.Run("generates module file successfully", func(t *testing.T) {
		renderer := &MockTemplateRenderer{}
		writer := &MockFileWriter{}
		service := NewTemplateService(renderer, writer)

		template := &domain.Template{
			Type:     domain.TemplateTypePath,
			Category: domain.CategoryInitD,
		}
		data := &domain.TemplateData{
			ModuleName: "my-bin",
			Fields: map[string]string{
				"path_dir": "/usr/local/bin",
			},
		}

		renderedContent := "#!/bin/bash\n# my-bin\nexport PATH=\"/usr/local/bin:$PATH\""
		renderer.On("RenderModuleFile", template, data).Return(renderedContent, nil)
		writer.On("WriteFile", "modules/init.d/my-bin.sh", renderedContent).Return(nil)

		result, err := service.Generate(template, data, "modules")

		require.NoError(t, err)
		assert.Equal(t, "my-bin", result.ModuleName)
		assert.Equal(t, "modules/init.d/my-bin.sh", result.FilePath)
		assert.Equal(t, "init.d", result.Category)
		assert.Contains(t, result.Message, "Generated my-bin module")

		renderer.AssertExpectations(t)
		writer.AssertExpectations(t)
	})

	t.Run("returns error when rendering fails", func(t *testing.T) {
		renderer := &MockTemplateRenderer{}
		writer := &MockFileWriter{}
		service := NewTemplateService(renderer, writer)

		template := &domain.Template{
			Type:     domain.TemplateTypePath,
			Category: domain.CategoryInitD,
		}
		data := &domain.TemplateData{
			ModuleName: "my-bin",
		}

		renderer.On("RenderModuleFile", template, data).Return("", fmt.Errorf("render error"))

		result, err := service.Generate(template, data, "modules")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to render template")

		renderer.AssertExpectations(t)
		writer.AssertNotCalled(t, "WriteFile")
	})

	t.Run("returns error when writing fails", func(t *testing.T) {
		renderer := &MockTemplateRenderer{}
		writer := &MockFileWriter{}
		service := NewTemplateService(renderer, writer)

		template := &domain.Template{
			Type:     domain.TemplateTypePath,
			Category: domain.CategoryInitD,
		}
		data := &domain.TemplateData{
			ModuleName: "my-bin",
			Fields: map[string]string{
				"path_dir": "/usr/local/bin",
			},
		}

		renderedContent := "module content"
		renderer.On("RenderModuleFile", template, data).Return(renderedContent, nil)
		writer.On("WriteFile", "modules/init.d/my-bin.sh", renderedContent).Return(fmt.Errorf("write error"))

		result, err := service.Generate(template, data, "modules")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to write module file")

		renderer.AssertExpectations(t)
		writer.AssertExpectations(t)
	})

	t.Run("generates file in correct category directory", func(t *testing.T) {
		testCases := []struct {
			category     domain.TemplateCategory
			expectedPath string
		}{
			{domain.CategoryInitD, "modules/init.d/test.sh"},
			{domain.CategoryRcPreD, "modules/rc_pre.d/test.sh"},
			{domain.CategoryRcPostD, "modules/rc_post.d/test.sh"},
		}

		for _, tc := range testCases {
			t.Run(string(tc.category), func(t *testing.T) {
				renderer := &MockTemplateRenderer{}
				writer := &MockFileWriter{}
				service := NewTemplateService(renderer, writer)

				template := &domain.Template{
					Category: tc.category,
				}
				data := &domain.TemplateData{
					ModuleName: "test",
				}

				renderer.On("RenderModuleFile", template, data).Return("content", nil)
				writer.On("WriteFile", tc.expectedPath, "content").Return(nil)

				result, err := service.Generate(template, data, "modules")

				require.NoError(t, err)
				assert.Equal(t, tc.expectedPath, result.FilePath)
				assert.Equal(t, string(tc.category), result.Category)

				renderer.AssertExpectations(t)
				writer.AssertExpectations(t)
			})
		}
	})
}
