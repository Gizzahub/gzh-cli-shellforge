package template

import (
	"fmt"
	"strings"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// Renderer handles template rendering with field substitution
type Renderer struct{}

// NewRenderer creates a new template renderer
func NewRenderer() *Renderer {
	return &Renderer{}
}

// Render renders a template with the given data
func (r *Renderer) Render(template *domain.Template, data *domain.TemplateData) (string, error) {
	// Validate data first
	if err := data.Validate(template); err != nil {
		return "", err
	}

	content := template.Content

	// Replace module name
	content = strings.ReplaceAll(content, "{{MODULE_NAME}}", data.ModuleName)

	// Replace description
	if data.Description != "" {
		content = strings.ReplaceAll(content, "{{DESCRIPTION}}", data.Description)
	} else {
		content = strings.ReplaceAll(content, "{{DESCRIPTION}}", template.Description)
	}

	// Replace field values
	for key, value := range data.Fields {
		placeholder := fmt.Sprintf("{{%s}}", strings.ToUpper(key))
		content = strings.ReplaceAll(content, placeholder, value)
	}

	// Replace dependencies
	if len(data.Requires) > 0 {
		content = strings.ReplaceAll(content, "{{REQUIRES}}", strings.Join(data.Requires, ", "))
	} else {
		content = strings.ReplaceAll(content, "{{REQUIRES}}", "")
	}

	return content, nil
}

// RenderModuleFile renders a complete module file with header
func (r *Renderer) RenderModuleFile(template *domain.Template, data *domain.TemplateData) (string, error) {
	content, err := r.Render(template, data)
	if err != nil {
		return "", err
	}

	// Build header
	var header strings.Builder
	header.WriteString("#!/bin/bash\n")
	header.WriteString(fmt.Sprintf("# %s\n", data.ModuleName))

	description := data.Description
	if description == "" {
		description = template.Description
	}
	header.WriteString(fmt.Sprintf("# %s\n", description))

	if len(data.Requires) > 0 {
		header.WriteString(fmt.Sprintf("# Requires: %s\n", strings.Join(data.Requires, ", ")))
	}

	header.WriteString("\n")
	header.WriteString(content)

	return header.String(), nil
}
