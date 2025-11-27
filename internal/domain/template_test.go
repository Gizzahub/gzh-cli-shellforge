package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateData_Validate(t *testing.T) {
	template := &Template{
		Type: TemplateTypePath,
		Fields: []TemplateField{
			{Name: "path_dir", Required: true},
			{Name: "description", Required: false},
		},
	}

	t.Run("valid data", func(t *testing.T) {
		data := &TemplateData{
			ModuleName: "my-bin",
			Fields: map[string]string{
				"path_dir":    "/usr/local/bin",
				"description": "My custom bin directory",
			},
		}

		err := data.Validate(template)
		assert.NoError(t, err)
	})

	t.Run("missing module name", func(t *testing.T) {
		data := &TemplateData{
			ModuleName: "",
			Fields: map[string]string{
				"path_dir": "/usr/local/bin",
			},
		}

		err := data.Validate(template)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "module name is required")
	})

	t.Run("missing required field", func(t *testing.T) {
		data := &TemplateData{
			ModuleName: "my-bin",
			Fields:     map[string]string{},
		}

		err := data.Validate(template)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required field 'path_dir' is missing")
	})

	t.Run("empty required field", func(t *testing.T) {
		data := &TemplateData{
			ModuleName: "my-bin",
			Fields: map[string]string{
				"path_dir": "   ",
			},
		}

		err := data.Validate(template)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required field 'path_dir' is missing")
	})

	t.Run("missing optional field is ok", func(t *testing.T) {
		data := &TemplateData{
			ModuleName: "my-bin",
			Fields: map[string]string{
				"path_dir": "/usr/local/bin",
			},
		}

		err := data.Validate(template)
		assert.NoError(t, err)
	})
}

func TestGetAllTemplateTypes(t *testing.T) {
	types := GetAllTemplateTypes()

	assert.Len(t, types, 6)
	assert.Contains(t, types, TemplateTypePath)
	assert.Contains(t, types, TemplateTypeEnv)
	assert.Contains(t, types, TemplateTypeAlias)
	assert.Contains(t, types, TemplateTypeConditionalSource)
	assert.Contains(t, types, TemplateTypeToolInit)
	assert.Contains(t, types, TemplateTypeOSSpecific)
}

func TestIsValidTemplateType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid: path", "path", true},
		{"valid: env", "env", true},
		{"valid: alias", "alias", true},
		{"valid: conditional-source", "conditional-source", true},
		{"valid: tool-init", "tool-init", true},
		{"valid: os-specific", "os-specific", true},
		{"invalid: unknown", "unknown", false},
		{"invalid: empty", "", false},
		{"invalid: PATH (wrong case)", "PATH", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidTemplateType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTemplateCategory(t *testing.T) {
	tests := []struct {
		templateType TemplateType
		expected     TemplateCategory
	}{
		{TemplateTypePath, CategoryInitD},
		{TemplateTypeEnv, CategoryRcPreD},
		{TemplateTypeAlias, CategoryRcPostD},
		{TemplateTypeConditionalSource, CategoryRcPreD},
		{TemplateTypeToolInit, CategoryRcPreD},
		{TemplateTypeOSSpecific, CategoryRcPreD},
	}

	for _, tt := range tests {
		t.Run(string(tt.templateType), func(t *testing.T) {
			category := GetTemplateCategory(tt.templateType)
			assert.Equal(t, tt.expected, category)
		})
	}
}
