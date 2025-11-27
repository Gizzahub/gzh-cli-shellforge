package app

import (
	"fmt"
	"path/filepath"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// TemplateRenderer defines the interface for template rendering
type TemplateRenderer interface {
	RenderModuleFile(template *domain.Template, data *domain.TemplateData) (string, error)
}

// FileWriter defines the interface for file writing
type FileWriter interface {
	WriteFile(path, content string) error
}

// TemplateService handles template generation operations
type TemplateService struct {
	renderer TemplateRenderer
	writer   FileWriter
}

// NewTemplateService creates a new template service
func NewTemplateService(renderer TemplateRenderer, writer FileWriter) *TemplateService {
	return &TemplateService{
		renderer: renderer,
		writer:   writer,
	}
}

// GenerateResult contains the result of template generation
type GenerateResult struct {
	ModuleName string
	FilePath   string
	Category   string
	Message    string
}

// Generate generates a module file from a template
func (s *TemplateService) Generate(
	template *domain.Template,
	data *domain.TemplateData,
	configDir string,
) (*GenerateResult, error) {
	// Render the module file
	content, err := s.renderer.RenderModuleFile(template, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// Determine output path
	fileName := fmt.Sprintf("%s.sh", data.ModuleName)
	categoryDir := filepath.Join(configDir, string(template.Category))
	filePath := filepath.Join(categoryDir, fileName)

	// Write the file
	if err := s.writer.WriteFile(filePath, content); err != nil {
		return nil, fmt.Errorf("failed to write module file: %w", err)
	}

	result := &GenerateResult{
		ModuleName: data.ModuleName,
		FilePath:   filePath,
		Category:   string(template.Category),
		Message:    fmt.Sprintf("Generated %s module at %s", data.ModuleName, filePath),
	}

	return result, nil
}
