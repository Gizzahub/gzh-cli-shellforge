package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/factory"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/template"
)

type templateFlags struct {
	configDir string
	fields    []string
	requires  []string
	verbose   bool
}

func newTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Generate module from template",
		Long: `Template generates a shell module from predefined templates.

Available template types:
  path                  Add directory to PATH (init.d/)
  env                   Set environment variable (rc_pre.d/)
  alias                 Define shell aliases (rc_post.d/)
  conditional-source    Source file if it exists (rc_pre.d/)
  tool-init             Initialize development tool (rc_pre.d/)
  os-specific           OS-specific configuration (rc_pre.d/)`,
	}

	cmd.AddCommand(newTemplateGenerateCmd())
	cmd.AddCommand(newTemplateListCmd())

	return cmd
}

func newTemplateGenerateCmd() *cobra.Command {
	flags := &templateFlags{}

	cmd := &cobra.Command{
		Use:   "generate <template-type> <module-name>",
		Short: "Generate a module from a template",
		Long: `Generate creates a shell module file from a predefined template.

Template types:
  path                  Add directory to PATH
  env                   Set environment variable
  alias                 Define shell aliases
  conditional-source    Source file if it exists
  tool-init             Initialize development tool
  os-specific           OS-specific configuration

Fields are specified with -f key=value flags.`,
		Example: `  # Generate path module
  gz-shellforge template generate path my-bin -f path_dir=/usr/local/bin

  # Generate env module
  gz-shellforge template generate env EDITOR -f var_name=EDITOR -f var_value=vim

  # Generate with dependencies
  gz-shellforge template generate tool-init nvm -f tool_name=nvm -f init_command='eval "$(nvm init)"' -r brew-path`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateType := args[0]
			moduleName := args[1]
			return runTemplateGenerate(templateType, moduleName, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.configDir, "config-dir", "c", "modules", "Module directory")
	cmd.Flags().StringSliceVarP(&flags.fields, "field", "f", []string{}, "Template field (key=value)")
	cmd.Flags().StringSliceVarP(&flags.requires, "requires", "r", []string{}, "Module dependencies")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed output")

	return cmd
}

func newTemplateListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available templates",
		Long:  `List displays all available module templates with descriptions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateList()
		},
	}

	return cmd
}

func runTemplateGenerate(templateType, moduleName string, flags *templateFlags) error {
	// Validate template type
	if !domain.IsValidTemplateType(templateType) {
		return fmt.Errorf("invalid template type: %s\nRun 'gz-shellforge template list' to see available templates", templateType)
	}

	// Get template
	tmpl, ok := template.GetBuiltinTemplate(domain.TemplateType(templateType))
	if !ok {
		return fmt.Errorf("template not found: %s", templateType)
	}

	// Parse fields
	fieldMap, err := parseFields(flags.fields)
	if err != nil {
		return err
	}

	// Create template data
	data := &domain.TemplateData{
		ModuleName: moduleName,
		Fields:     fieldMap,
		Requires:   flags.requires,
	}

	if flags.verbose {
		fmt.Printf("Generating module:\n")
		fmt.Printf("  Type:     %s\n", templateType)
		fmt.Printf("  Name:     %s\n", moduleName)
		fmt.Printf("  Category: %s\n", tmpl.Category)
		if len(flags.requires) > 0 {
			fmt.Printf("  Requires: %s\n", strings.Join(flags.requires, ", "))
		}
		fmt.Println()
	}

	// Initialize services
	services := factory.NewServices()
	renderer := template.NewRenderer()
	service := app.NewTemplateService(renderer, services.Writer)

	// Generate module
	result, err := service.Generate(tmpl, data, flags.configDir)
	if err != nil {
		return fmt.Errorf("failed to generate module: %w", err)
	}

	// Display results
	fmt.Printf("✓ %s\n", result.Message)
	if flags.verbose {
		fmt.Printf("\nModule file: %s\n", result.FilePath)
		fmt.Printf("Category:    %s\n", result.Category)
	}

	return nil
}

func runTemplateList() error {
	fmt.Println("Available templates:")
	fmt.Println()

	templates := template.GetBuiltinTemplates()

	// Display in order
	types := []domain.TemplateType{
		domain.TemplateTypePath,
		domain.TemplateTypeEnv,
		domain.TemplateTypeAlias,
		domain.TemplateTypeConditionalSource,
		domain.TemplateTypeToolInit,
		domain.TemplateTypeOSSpecific,
	}

	for _, templateType := range types {
		tmpl := templates[templateType]
		fmt.Printf("  %-22s %s (%s/)\n", tmpl.Name, tmpl.Description, tmpl.Category)

		// Show required fields
		var requiredFields []string
		for _, field := range tmpl.Fields {
			if field.Required {
				requiredFields = append(requiredFields, field.Name)
			}
		}
		if len(requiredFields) > 0 {
			fmt.Printf("  %22s Required: %s\n", "", strings.Join(requiredFields, ", "))
		}
		fmt.Println()
	}

	fmt.Println("Usage:")
	fmt.Println("  gz-shellforge template generate <template-name> <module-name> [flags]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gz-shellforge template generate path my-bin -f path_dir=/usr/local/bin")
	fmt.Println("  gz-shellforge template generate env EDITOR -f var_name=EDITOR -f var_value=vim")
	fmt.Println("  gz-shellforge template generate tool-init nvm -f tool_name=nvm -f init_command='eval \"$(nvm init)\"'")

	return nil
}

func parseFields(fields []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid field format: %s (expected key=value)", field)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("empty field name in: %s", field)
		}

		result[key] = value
	}

	return result, nil
}
