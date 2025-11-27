package template

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

func TestGetBuiltinTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()

	// Check all 6 template types exist
	assert.Len(t, templates, 6)

	assert.Contains(t, templates, domain.TemplateTypePath)
	assert.Contains(t, templates, domain.TemplateTypeEnv)
	assert.Contains(t, templates, domain.TemplateTypeAlias)
	assert.Contains(t, templates, domain.TemplateTypeConditionalSource)
	assert.Contains(t, templates, domain.TemplateTypeToolInit)
	assert.Contains(t, templates, domain.TemplateTypeOSSpecific)

	// Verify each template is valid
	for templateType, template := range templates {
		assert.Equal(t, templateType, template.Type, "template type mismatch for %s", templateType)
		assert.NotEmpty(t, template.Name)
		assert.NotEmpty(t, template.Description)
		assert.NotEmpty(t, template.Category)
		assert.NotEmpty(t, template.Content)
	}
}

func TestGetBuiltinTemplate(t *testing.T) {
	t.Run("returns existing template", func(t *testing.T) {
		template, ok := GetBuiltinTemplate(domain.TemplateTypePath)
		assert.True(t, ok)
		assert.NotNil(t, template)
		assert.Equal(t, domain.TemplateTypePath, template.Type)
	})

	t.Run("returns false for non-existent template", func(t *testing.T) {
		_, ok := GetBuiltinTemplate(domain.TemplateType("non-existent"))
		assert.False(t, ok)
	})
}

func TestPathTemplate(t *testing.T) {
	template := getPathTemplate()

	assert.Equal(t, domain.TemplateTypePath, template.Type)
	assert.Equal(t, domain.CategoryInitD, template.Category)
	assert.Len(t, template.Fields, 1)

	// Check required field
	field := template.Fields[0]
	assert.Equal(t, "path_dir", field.Name)
	assert.True(t, field.Required)

	// Check content
	assert.Contains(t, template.Content, "export PATH")
	assert.Contains(t, template.Content, "{{PATH_DIR}}")
}

func TestEnvTemplate(t *testing.T) {
	template := getEnvTemplate()

	assert.Equal(t, domain.TemplateTypeEnv, template.Type)
	assert.Equal(t, domain.CategoryRcPreD, template.Category)
	assert.Len(t, template.Fields, 2)

	// Check both required fields
	assert.Equal(t, "var_name", template.Fields[0].Name)
	assert.True(t, template.Fields[0].Required)

	assert.Equal(t, "var_value", template.Fields[1].Name)
	assert.True(t, template.Fields[1].Required)

	// Check content
	assert.Contains(t, template.Content, "export {{VAR_NAME}}")
}

func TestAliasTemplate(t *testing.T) {
	template := getAliasTemplate()

	assert.Equal(t, domain.TemplateTypeAlias, template.Type)
	assert.Equal(t, domain.CategoryRcPostD, template.Category)
	assert.Len(t, template.Fields, 1)

	// Check required field
	field := template.Fields[0]
	assert.Equal(t, "aliases", field.Name)
	assert.True(t, field.Required)
}

func TestConditionalSourceTemplate(t *testing.T) {
	template := getConditionalSourceTemplate()

	assert.Equal(t, domain.TemplateTypeConditionalSource, template.Type)
	assert.Equal(t, domain.CategoryRcPreD, template.Category)
	assert.Len(t, template.Fields, 1)

	// Check content
	assert.Contains(t, template.Content, "if [ -f")
	assert.Contains(t, template.Content, "source")
}

func TestToolInitTemplate(t *testing.T) {
	template := getToolInitTemplate()

	assert.Equal(t, domain.TemplateTypeToolInit, template.Type)
	assert.Equal(t, domain.CategoryRcPreD, template.Category)
	assert.Len(t, template.Fields, 2)

	// Check content
	assert.Contains(t, template.Content, "command -v")
	assert.Contains(t, template.Content, "{{INIT_COMMAND}}")
}

func TestOSSpecificTemplate(t *testing.T) {
	template := getOSSpecificTemplate()

	assert.Equal(t, domain.TemplateTypeOSSpecific, template.Type)
	assert.Equal(t, domain.CategoryRcPreD, template.Category)
	assert.Len(t, template.Fields, 2)

	// Check fields are optional
	assert.False(t, template.Fields[0].Required)
	assert.False(t, template.Fields[1].Required)

	// Check content
	assert.Contains(t, template.Content, "case \"$(uname -s)\" in")
	assert.Contains(t, template.Content, "Darwin")
	assert.Contains(t, template.Content, "Linux")
}
