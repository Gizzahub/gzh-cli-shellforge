package domain

import (
	"fmt"
	"strings"
)

// TemplateType represents the type of template
type TemplateType string

const (
	TemplateTypePath              TemplateType = "path"
	TemplateTypeEnv               TemplateType = "env"
	TemplateTypeAlias             TemplateType = "alias"
	TemplateTypeConditionalSource TemplateType = "conditional-source"
	TemplateTypeToolInit          TemplateType = "tool-init"
	TemplateTypeOSSpecific        TemplateType = "os-specific"
)

// TemplateCategory represents the directory category for the module
type TemplateCategory string

const (
	CategoryInitD    TemplateCategory = "init.d"
	CategoryRcPreD   TemplateCategory = "rc_pre.d"
	CategoryRcPostD  TemplateCategory = "rc_post.d"
)

// Template represents a module template definition
type Template struct {
	Type        TemplateType
	Name        string
	Description string
	Category    TemplateCategory
	Fields      []TemplateField
	Content     string // Template content with placeholders
}

// TemplateField represents a field that needs to be filled in the template
type TemplateField struct {
	Name        string
	Description string
	Required    bool
	Default     string
}

// TemplateData holds the data to render a template
type TemplateData struct {
	ModuleName  string
	Description string
	Fields      map[string]string
	Requires    []string // Dependencies
}

// Validate validates the template data
func (td *TemplateData) Validate(template *Template) error {
	if td.ModuleName == "" {
		return fmt.Errorf("module name is required")
	}

	// Check required fields
	for _, field := range template.Fields {
		if field.Required {
			value, ok := td.Fields[field.Name]
			if !ok || strings.TrimSpace(value) == "" {
				return fmt.Errorf("required field '%s' is missing", field.Name)
			}
		}
	}

	return nil
}

// GetAllTemplateTypes returns all available template types
func GetAllTemplateTypes() []TemplateType {
	return []TemplateType{
		TemplateTypePath,
		TemplateTypeEnv,
		TemplateTypeAlias,
		TemplateTypeConditionalSource,
		TemplateTypeToolInit,
		TemplateTypeOSSpecific,
	}
}

// IsValidTemplateType checks if a template type is valid
func IsValidTemplateType(t string) bool {
	templateType := TemplateType(t)
	for _, valid := range GetAllTemplateTypes() {
		if templateType == valid {
			return true
		}
	}
	return false
}

// GetTemplateCategory returns the category for a template type
func GetTemplateCategory(t TemplateType) TemplateCategory {
	switch t {
	case TemplateTypePath:
		return CategoryInitD
	case TemplateTypeEnv:
		return CategoryRcPreD
	case TemplateTypeAlias:
		return CategoryRcPostD
	case TemplateTypeConditionalSource:
		return CategoryRcPreD
	case TemplateTypeToolInit:
		return CategoryRcPreD
	case TemplateTypeOSSpecific:
		return CategoryRcPreD
	default:
		return CategoryRcPreD
	}
}
