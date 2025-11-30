package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/diffcomparator"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/filesystem"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/rcparser"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/yamlparser"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigrateBuildDiffWorkflow tests the complete end-to-end workflow:
// 1. Migrate RC file to modular structure
// 2. Build output from modules
// 3. Compare original with generated
func TestMigrateBuildDiffWorkflow(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()

	// Create example RC file content
	rcContent := `# Sample .zshrc
# --- OS Detection ---
case "$(uname -s)" in
  Darwin)
    export MACHINE="Mac"
    ;;
  Linux)
    export MACHINE="Linux"
    ;;
esac

# === PATH Setup ===
export PATH="/usr/local/bin:$PATH"

# --- Git Aliases ---
alias gs='git status'
alias ga='git add'
alias gc='git commit'

# === Helper Functions ===
function mkcd() {
  mkdir -p "$1" && cd "$1"
}
`

	// Write RC file
	rcPath := "/test/.zshrc"
	err := afero.WriteFile(fs, rcPath, []byte(rcContent), 0644)
	require.NoError(t, err)

	// Step 1: Migrate RC file
	t.Run("Step1_Migrate", func(t *testing.T) {
		reader := filesystem.NewReader(fs)
		writer := filesystem.NewWriter(fs)
		parser := rcparser.New(fs)
		migrationService := app.NewMigrationService(parser, reader, writer)

		outputDir := "/test/modules"
		manifestPath := "/test/manifest.yaml"

		result, err := migrationService.Migrate(rcPath, outputDir, manifestPath)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Greater(t, result.ModulesCreated, 0, "should create at least one module")

		// Verify manifest was created
		exists := reader.FileExists(manifestPath)
		assert.True(t, exists, "manifest.yaml should exist")

		// Verify at least one module file was created
		assert.Greater(t, len(result.ModuleFilesPaths), 0, "should have created module files")
	})

	// Step 2: Build output from modules
	t.Run("Step2_Build", func(t *testing.T) {
		reader := filesystem.NewReader(fs)
		writer := filesystem.NewWriter(fs)
		yamlParser := yamlparser.New(fs)
		builderService := app.NewBuilderService(yamlParser, reader, writer)

		manifestPath := "/test/manifest.yaml"
		configDir := "/test/modules"
		outputPath := "/test/.zshrc.new"
		targetOS := "Mac"

		result, err := builderService.Build(app.BuildOptions{
			ConfigDir: configDir,
			Manifest:  manifestPath,
			Output:    outputPath,
			OS:        targetOS,
			DryRun:    false,
			Verbose:   false,
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Greater(t, len(result.Output), 0, "generated content should not be empty")

		// Verify output file was created
		exists := reader.FileExists(outputPath)
		assert.True(t, exists, ".zshrc.new should exist")

		// Verify content contains expected sections
		assert.Contains(t, result.Output, "PATH", "should contain PATH setup")
	})

	// Step 3: Compare original with generated
	t.Run("Step3_Diff", func(t *testing.T) {
		reader := filesystem.NewReader(fs)
		comparator := diffcomparator.NewComparator(fs)
		diffService := app.NewDiffService(comparator, reader)

		originalPath := "/test/.zshrc"
		generatedPath := "/test/.zshrc.new"

		result, err := diffService.Compare(originalPath, generatedPath, domain.DiffFormatSummary)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.DiffResult)

		// Files should not be identical (module headers are added)
		assert.False(t, result.DiffResult.IsIdentical, "files should have differences due to module headers")

		// Should have some statistics
		stats := result.DiffResult.Statistics
		assert.Greater(t, stats.TotalLines, 0, "should have total lines")

		// Content should be preserved (lines added for headers, but original content present)
		assert.Greater(t, stats.LinesUnchanged+stats.LinesAdded, 0, "should have unchanged or added lines")
	})

	// Step 4: Full workflow validation
	t.Run("Step4_ValidateWorkflow", func(t *testing.T) {
		reader := filesystem.NewReader(fs)

		// Verify all expected files exist
		expectedFiles := []string{
			"/test/.zshrc",        // Original
			"/test/manifest.yaml", // Generated manifest
			"/test/.zshrc.new",    // Generated output
		}

		for _, file := range expectedFiles {
			exists := reader.FileExists(file)
			assert.True(t, exists, "file should exist: %s", file)
		}

		// Verify modules directory has files
		entries, err := afero.ReadDir(fs, "/test/modules")
		require.NoError(t, err)
		assert.Greater(t, len(entries), 0, "modules directory should contain files")
	})
}

// TestDiffFormats tests all diff output formats in the workflow
func TestDiffFormats(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create two test files with known differences
	original := `line1
line2
line3`

	generated := `line1
line2_modified
line3
line4`

	afero.WriteFile(fs, "/original.sh", []byte(original), 0644)
	afero.WriteFile(fs, "/generated.sh", []byte(generated), 0644)

	reader := filesystem.NewReader(fs)
	comparator := diffcomparator.NewComparator(fs)
	diffService := app.NewDiffService(comparator, reader)

	formats := []domain.DiffFormat{
		domain.DiffFormatSummary,
		domain.DiffFormatUnified,
		domain.DiffFormatContext,
		domain.DiffFormatSideBySide,
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			result, err := diffService.Compare("/original.sh", "/generated.sh", format)
			require.NoError(t, err)
			assert.NotNil(t, result.DiffResult)
			assert.False(t, result.DiffResult.IsIdentical)
			assert.NotEmpty(t, result.DiffResult.Content, "content should not be empty for format: %s", format)

			// Verify format-specific content
			switch format {
			case domain.DiffFormatSummary:
				assert.Contains(t, result.DiffResult.Content, "Statistics")
			case domain.DiffFormatUnified:
				assert.Contains(t, result.DiffResult.Content, "---")
				assert.Contains(t, result.DiffResult.Content, "+++")
			case domain.DiffFormatContext:
				assert.Contains(t, result.DiffResult.Content, "***")
			case domain.DiffFormatSideBySide:
				assert.Contains(t, result.DiffResult.Content, "|")
			}
		})
	}
}

// TestRealWorldExample tests with the actual example file
func TestRealWorldExample(t *testing.T) {
	// This test uses the real filesystem to test with examples/sample.zshrc
	if testing.Short() {
		t.Skip("skipping real-world example test in short mode")
	}

	// Get project root
	cwd, err := os.Getwd()
	require.NoError(t, err)
	projectRoot := filepath.Join(cwd, "..", "..")

	exampleFile := filepath.Join(projectRoot, "examples", "sample.zshrc")

	// Check if example file exists
	if _, err := os.Stat(exampleFile); os.IsNotExist(err) {
		t.Skip("examples/sample.zshrc not found, skipping real-world test")
	}

	// Use real filesystem for this test
	fs := afero.NewOsFs()
	reader := filesystem.NewReader(fs)

	// Verify we can read the example
	exists := reader.FileExists(exampleFile)
	assert.True(t, exists, "example file should exist")

	content, err := reader.ReadFile(exampleFile)
	require.NoError(t, err)
	assert.NotEmpty(t, content, "example file should have content")
	assert.Contains(t, content, "OS Detection", "example should have OS Detection section")
}
