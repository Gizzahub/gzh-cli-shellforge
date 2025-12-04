package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/factory"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

type validateFlags struct {
	configDir string
	manifest  string
	verbose   bool
}

func newValidateCmd() *cobra.Command {
	flags := &validateFlags{}

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate manifest file and module files",
		Long: `Validate checks the manifest file for syntax errors, validates module
definitions, checks for circular dependencies, and verifies that all
referenced module files exist.

This command performs validation without building the configuration,
making it useful for quickly checking manifest correctness during
development.`,
		Example: `  # Validate default manifest
  shellforge validate

  # Validate custom manifest
  shellforge validate --manifest custom.yaml --config-dir modules

  # Verbose validation with detailed output
  shellforge validate --verbose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(flags)
		},
	}

	// Define flags
	cmd.Flags().StringVarP(&flags.configDir, "config-dir", "c", "modules", "Directory containing module files")
	cmd.Flags().StringVarP(&flags.manifest, "manifest", "m", "manifest.yaml", "Path to manifest file")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed validation output")

	return cmd
}

func runValidate(flags *validateFlags) error {
	if flags.verbose {
		fmt.Printf("Validating manifest: %s\n", flags.manifest)
		fmt.Printf("Module directory: %s\n", flags.configDir)
		fmt.Println()
	}

	// Create services
	services := factory.NewServices()
	parser := services.Parser
	reader := services.Reader

	// 1. Parse manifest
	if flags.verbose {
		fmt.Println("1. Parsing manifest file...")
	}

	manifest, err := parser.Parse(flags.manifest)
	if err != nil {
		return fmt.Errorf("✗ Manifest parsing failed: %w", err)
	}

	if flags.verbose {
		fmt.Printf("   ✓ Manifest parsed successfully (%d modules)\n", len(manifest.Modules))
	}

	// 2. Validate manifest structure
	if flags.verbose {
		fmt.Println("\n2. Validating manifest structure...")
	}

	validationErrors := manifest.Validate()
	if len(validationErrors) > 0 {
		fmt.Println("✗ Validation errors found:")
		for i, err := range validationErrors {
			fmt.Printf("   %d. %s\n", i+1, err.Error())
		}
		return fmt.Errorf("manifest validation failed with %d error(s)", len(validationErrors))
	}

	if flags.verbose {
		fmt.Println("   ✓ Manifest structure is valid")
	}

	// 3. Check for circular dependencies (try both Mac and Linux)
	if flags.verbose {
		fmt.Println("\n3. Checking for circular dependencies...")
	}

	// We'll use a simple resolver to check both common OSes
	resolver := &dependencyValidator{}
	if err := resolver.checkCircularDependencies(manifest); err != nil {
		return fmt.Errorf("✗ Circular dependency detected: %w", err)
	}

	if flags.verbose {
		fmt.Println("   ✓ No circular dependencies found")
	}

	// 4. Verify module files exist
	if flags.verbose {
		fmt.Println("\n4. Verifying module files...")
	}

	missingFiles := []string{}
	for _, module := range manifest.Modules {
		filePath := filepath.Join(flags.configDir, module.File)
		if !reader.FileExists(filePath) {
			missingFiles = append(missingFiles, fmt.Sprintf("%s (referenced by module '%s')", filePath, module.Name))
		}
	}

	if len(missingFiles) > 0 {
		fmt.Println("✗ Missing module files:")
		for i, file := range missingFiles {
			fmt.Printf("   %d. %s\n", i+1, file)
		}
		return fmt.Errorf("validation failed: %d module file(s) not found", len(missingFiles))
	}

	if flags.verbose {
		fmt.Printf("   ✓ All %d module files exist\n", len(manifest.Modules))
	}

	// 5. Summary
	if flags.verbose {
		fmt.Println("\n" + strings.Repeat("=", 50))
	}

	fmt.Printf("✓ Validation successful!\n")
	fmt.Printf("  Modules: %d\n", len(manifest.Modules))
	fmt.Printf("  Manifest: %s\n", flags.manifest)

	return nil
}

// dependencyValidator is a simple helper to check for circular dependencies
type dependencyValidator struct{}

func (v *dependencyValidator) checkCircularDependencies(manifest *domain.Manifest) error {
	// Build a simple dependency map
	depMap := make(map[string][]string)
	moduleSet := make(map[string]bool)

	for _, module := range manifest.Modules {
		depMap[module.Name] = module.Requires
		moduleSet[module.Name] = true
	}

	// Check each module for circular dependencies using DFS
	for moduleName := range moduleSet {
		visited := make(map[string]bool)
		if v.hasCycle(moduleName, depMap, visited, make(map[string]bool)) {
			return fmt.Errorf("circular dependency involving module '%s'", moduleName)
		}
	}

	return nil
}

func (v *dependencyValidator) hasCycle(node string, graph map[string][]string, visited, recStack map[string]bool) bool {
	visited[node] = true
	recStack[node] = true

	for _, neighbor := range graph[node] {
		if !visited[neighbor] {
			if v.hasCycle(neighbor, graph, visited, recStack) {
				return true
			}
		} else if recStack[neighbor] {
			return true
		}
	}

	recStack[node] = false
	return false
}
