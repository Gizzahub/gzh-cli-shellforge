package cli

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/diffcomparator"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/filesystem"
)

type diffFlags struct {
	format  string
	verbose bool
}

func newDiffCmd() *cobra.Command {
	flags := &diffFlags{}

	cmd := &cobra.Command{
		Use:   "diff <original-file> <generated-file>",
		Short: "Compare original and generated configuration files",
		Long: `Compare two shell configuration files and show differences.

The diff command analyzes differences between an original RC file and a
generated configuration, providing statistics and formatted output.

Supported formats:
  - summary:     Statistics only (lines added/removed/modified)
  - unified:     Unified diff format (git diff style)
  - context:     Context diff format (with surrounding lines)
  - side-by-side: Side-by-side comparison view

Use this command to review changes before deploying generated configurations.`,
		Example: `  # Show summary statistics
  gz-shellforge diff ~/.zshrc ~/.zshrc.new

  # Show unified diff (git diff style)
  gz-shellforge diff ~/.zshrc ~/.zshrc.new --format unified

  # Show side-by-side comparison
  gz-shellforge diff ~/.zshrc ~/.zshrc.new --format side-by-side

  # Compare with verbose output
  gz-shellforge diff ~/.zshrc ~/.zshrc.new -v`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			originalPath := args[0]
			generatedPath := args[1]
			return runDiff(originalPath, generatedPath, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.format, "format", "f", "summary", "Output format (summary, unified, context, side-by-side)")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed output")

	return cmd
}

func runDiff(originalPath, generatedPath string, flags *diffFlags) error {
	// Expand home directory
	var err error
	originalPath, err = expandHomePath(originalPath)
	if err != nil {
		return fmt.Errorf("invalid original path: %w", err)
	}
	generatedPath, err = expandHomePath(generatedPath)
	if err != nil {
		return fmt.Errorf("invalid generated path: %w", err)
	}

	// Validate format
	format := domain.DiffFormat(flags.format)
	if err := domain.ValidateFormat(flags.format); err != nil {
		return fmt.Errorf("invalid format: %w", err)
	}

	if flags.verbose {
		fmt.Printf("Comparing files:\n")
		fmt.Printf("  Original:  %s\n", originalPath)
		fmt.Printf("  Generated: %s\n", generatedPath)
		fmt.Printf("  Format:    %s\n\n", flags.format)
	}

	// Initialize services
	fs := afero.NewOsFs()
	reader := filesystem.NewReader(fs)
	comparator := diffcomparator.NewComparator(fs)
	diffService := app.NewDiffService(comparator, reader)

	// Perform comparison
	result, err := diffService.Compare(originalPath, generatedPath, format)
	if err != nil {
		return fmt.Errorf("comparison failed: %w", err)
	}

	// Display results
	if flags.verbose && result.DiffResult.IsIdentical {
		fmt.Println("✓ Files are identical")
		fmt.Printf("  Total lines: %d\n", result.DiffResult.Statistics.TotalLines)
		return nil
	}

	if result.DiffResult.IsIdentical {
		fmt.Println("Files are identical")
		return nil
	}

	// Show diff content
	fmt.Print(result.DiffResult.Content)

	// Show summary in verbose mode for non-summary formats
	if flags.verbose && format != domain.DiffFormatSummary {
		fmt.Println("\n" + formatDiffSummary(result.DiffResult))
	}

	return nil
}

func formatDiffSummary(result *domain.DiffResult) string {
	stats := result.Statistics
	return fmt.Sprintf("Summary: %s", stats.Summary())
}
