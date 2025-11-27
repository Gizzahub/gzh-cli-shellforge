package template

import "github.com/gizzahub/gzh-cli-shellforge/internal/domain"

// GetBuiltinTemplates returns all built-in templates
func GetBuiltinTemplates() map[domain.TemplateType]*domain.Template {
	return map[domain.TemplateType]*domain.Template{
		domain.TemplateTypePath:              getPathTemplate(),
		domain.TemplateTypeEnv:               getEnvTemplate(),
		domain.TemplateTypeAlias:             getAliasTemplate(),
		domain.TemplateTypeConditionalSource: getConditionalSourceTemplate(),
		domain.TemplateTypeToolInit:          getToolInitTemplate(),
		domain.TemplateTypeOSSpecific:        getOSSpecificTemplate(),
	}
}

// GetBuiltinTemplate returns a specific built-in template
func GetBuiltinTemplate(templateType domain.TemplateType) (*domain.Template, bool) {
	templates := GetBuiltinTemplates()
	template, ok := templates[templateType]
	return template, ok
}

func getPathTemplate() *domain.Template {
	return &domain.Template{
		Type:        domain.TemplateTypePath,
		Name:        "path",
		Description: "Add directory to PATH",
		Category:    domain.CategoryInitD,
		Fields: []domain.TemplateField{
			{
				Name:        "path_dir",
				Description: "Directory path to add to PATH",
				Required:    true,
			},
		},
		Content: `# Add {{PATH_DIR}} to PATH
if [ -d "{{PATH_DIR}}" ]; then
    export PATH="{{PATH_DIR}}:$PATH"
fi`,
	}
}

func getEnvTemplate() *domain.Template {
	return &domain.Template{
		Type:        domain.TemplateTypeEnv,
		Name:        "env",
		Description: "Set environment variable",
		Category:    domain.CategoryRcPreD,
		Fields: []domain.TemplateField{
			{
				Name:        "var_name",
				Description: "Environment variable name",
				Required:    true,
			},
			{
				Name:        "var_value",
				Description: "Environment variable value",
				Required:    true,
			},
		},
		Content: `# Set {{VAR_NAME}} environment variable
export {{VAR_NAME}}="{{VAR_VALUE}}"`,
	}
}

func getAliasTemplate() *domain.Template {
	return &domain.Template{
		Type:        domain.TemplateTypeAlias,
		Name:        "alias",
		Description: "Define shell aliases",
		Category:    domain.CategoryRcPostD,
		Fields: []domain.TemplateField{
			{
				Name:        "aliases",
				Description: "Alias definitions (one per line)",
				Required:    true,
			},
		},
		Content: `# Shell aliases
{{ALIASES}}`,
	}
}

func getConditionalSourceTemplate() *domain.Template {
	return &domain.Template{
		Type:        domain.TemplateTypeConditionalSource,
		Name:        "conditional-source",
		Description: "Source file if it exists",
		Category:    domain.CategoryRcPreD,
		Fields: []domain.TemplateField{
			{
				Name:        "source_path",
				Description: "Path to file to source",
				Required:    true,
			},
		},
		Content: `# Source {{SOURCE_PATH}} if it exists
if [ -f "{{SOURCE_PATH}}" ]; then
    source "{{SOURCE_PATH}}"
fi`,
	}
}

func getToolInitTemplate() *domain.Template {
	return &domain.Template{
		Type:        domain.TemplateTypeToolInit,
		Name:        "tool-init",
		Description: "Initialize development tool",
		Category:    domain.CategoryRcPreD,
		Fields: []domain.TemplateField{
			{
				Name:        "tool_name",
				Description: "Name of the tool",
				Required:    true,
			},
			{
				Name:        "init_command",
				Description: "Initialization command",
				Required:    true,
			},
		},
		Content: `# Initialize {{TOOL_NAME}}
if command -v {{TOOL_NAME}} &> /dev/null; then
    {{INIT_COMMAND}}
fi`,
	}
}

func getOSSpecificTemplate() *domain.Template {
	return &domain.Template{
		Type:        domain.TemplateTypeOSSpecific,
		Name:        "os-specific",
		Description: "OS-specific configuration",
		Category:    domain.CategoryRcPreD,
		Fields: []domain.TemplateField{
			{
				Name:        "mac_content",
				Description: "Content for macOS",
				Required:    false,
			},
			{
				Name:        "linux_content",
				Description: "Content for Linux",
				Required:    false,
			},
		},
		Content: `# OS-specific configuration
case "$(uname -s)" in
    Darwin)
        # macOS
        {{MAC_CONTENT}}
        ;;
    Linux)
        # Linux
        {{LINUX_CONTENT}}
        ;;
esac`,
	}
}
