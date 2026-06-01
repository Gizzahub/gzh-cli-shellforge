package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	clierrors "github.com/gizzahub/gzh-cli-shellforge/internal/cli/errors"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/factory"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/helpers"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

type validateFlags struct {
	configDir    string
	manifest     string
	verbose      bool
	checkPrereqs bool
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
  shellforge validate --verbose

  # Also check external tool prerequisites (requires_bin / requires_path)
  shellforge validate --check-prereqs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(flags)
		},
	}

	cmd.Flags().StringVarP(&flags.configDir, "config-dir", "c", "modules", "Directory containing module files")
	cmd.Flags().StringVarP(&flags.manifest, "manifest", "m", "manifest.yaml", "Path to manifest file")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed validation output")
	cmd.Flags().BoolVar(&flags.checkPrereqs, "check-prereqs", false, "Also check requires_bin / requires_path (warn only)")

	return cmd
}

func runValidate(flags *validateFlags) error {
	if flags.verbose {
		fmt.Printf("Validating manifest: %s\n", flags.manifest)
		fmt.Printf("Module directory: %s\n", flags.configDir)
		fmt.Println()
	}

	services := factory.NewServices()

	manifest, err := services.Parser.Parse(flags.manifest)
	if err != nil {
		return clierrors.WrapError("manifest parsing", err)
	}

	if flags.verbose {
		fmt.Printf("✓ Manifest parsed (%d modules)\n\n", len(manifest.Modules))
	}

	// Build the validation pipeline.
	validators := []app.Validator{
		app.ManifestStructureValidator{},
		app.CircularDependencyValidator{},
		app.NewFileExistenceValidator(services.Reader),
	}
	if flags.checkPrereqs {
		targetOS := helpers.DetectOS()
		validators = append(validators, app.NewPrereqValidator(targetOS, domain.OsPrereqLookup{}))
	}

	pipeline := app.NewValidationPipeline(validators...)
	findings := pipeline.Run(manifest, flags.configDir)

	if len(findings) == 0 {
		fmt.Printf("✓ Validation successful!\n")
		fmt.Printf("  Modules: %d\n", len(manifest.Modules))
		fmt.Printf("  Manifest: %s\n", flags.manifest)
		return nil
	}

	printFindings(findings)

	if app.HasErrors(findings) {
		errCount := countBySeverity(findings, app.SeverityError)
		return fmt.Errorf("validation failed with %d error(s)", errCount)
	}

	// Warnings only — still success.
	fmt.Printf("✓ Validation successful!\n")
	fmt.Printf("  Modules: %d\n", len(manifest.Modules))
	fmt.Printf("  Manifest: %s\n", flags.manifest)
	return nil
}

func printFindings(findings []app.Finding) {
	errCount := countBySeverity(findings, app.SeverityError)
	warnCount := countBySeverity(findings, app.SeverityWarn)

	if errCount > 0 {
		fmt.Printf("✗ Validation errors found (%d error(s), %d warning(s)):\n", errCount, warnCount)
	} else {
		fmt.Printf("⚠ Validation warnings (%d):\n", warnCount)
	}

	for i, f := range findings {
		icon := "✗"
		if !f.IsError() {
			icon = "⚠"
		}
		if f.Module != "" {
			fmt.Printf("   %d. %s [%s] %s\n", i+1, icon, f.Module, f.Message)
		} else {
			fmt.Printf("   %d. %s %s\n", i+1, icon, f.Message)
		}
	}
	fmt.Println()
}

func countBySeverity(findings []app.Finding, severity string) int {
	n := 0
	for _, f := range findings {
		if f.Severity == severity {
			n++
		}
	}
	return n
}
