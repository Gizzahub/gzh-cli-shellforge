package diffcomparator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// Benchmark file comparison with identical files
func BenchmarkComparator_Compare_Identical(b *testing.B) {
	sizes := []struct {
		name  string
		lines int
	}{
		{"Small_10", 10},
		{"Medium_100", 100},
		{"Large_1000", 1000},
		{"VeryLarge_10000", 10000},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			fs := afero.NewMemMapFs()
			comp := NewComparator(fs)

			// Generate file content
			content := generateFileContent(size.lines)
			afero.WriteFile(fs, "/original.sh", []byte(content), 0o644)
			afero.WriteFile(fs, "/generated.sh", []byte(content), 0o644)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := comp.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark file comparison with line additions
func BenchmarkComparator_Compare_Additions(b *testing.B) {
	scenarios := []struct {
		name          string
		originalLines int
		addedLines    int
	}{
		{"Small_10pct", 100, 10},
		{"Small_50pct", 100, 50},
		{"Medium_10pct", 1000, 100},
		{"Medium_50pct", 1000, 500},
		{"Large_10pct", 10000, 1000},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			fs := afero.NewMemMapFs()
			comp := NewComparator(fs)

			originalContent := generateFileContent(scenario.originalLines)
			generatedContent := originalContent + generateFileContent(scenario.addedLines)

			afero.WriteFile(fs, "/original.sh", []byte(originalContent), 0o644)
			afero.WriteFile(fs, "/generated.sh", []byte(generatedContent), 0o644)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := comp.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark file comparison with modifications
func BenchmarkComparator_Compare_Modifications(b *testing.B) {
	scenarios := []struct {
		name        string
		totalLines  int
		modifiedPct int
	}{
		{"Small_10pct", 100, 10},
		{"Small_50pct", 100, 50},
		{"Medium_10pct", 1000, 10},
		{"Medium_50pct", 1000, 50},
		{"Large_10pct", 10000, 10},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			fs := afero.NewMemMapFs()
			comp := NewComparator(fs)

			originalContent := generateFileContent(scenario.totalLines)
			generatedContent := modifyContent(originalContent, scenario.modifiedPct)

			afero.WriteFile(fs, "/original.sh", []byte(originalContent), 0o644)
			afero.WriteFile(fs, "/generated.sh", []byte(generatedContent), 0o644)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := comp.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark different diff formats
func BenchmarkComparator_Compare_Formats(b *testing.B) {
	formats := []struct {
		name   string
		format domain.DiffFormat
	}{
		{"Summary", domain.DiffFormatSummary},
		{"Unified", domain.DiffFormatUnified},
		{"Context", domain.DiffFormatContext},
		{"SideBySide", domain.DiffFormatSideBySide},
	}

	// Use a medium-sized file with 50% modifications
	originalContent := generateFileContent(1000)
	generatedContent := modifyContent(originalContent, 50)

	for _, format := range formats {
		b.Run(format.name, func(b *testing.B) {
			fs := afero.NewMemMapFs()
			comp := NewComparator(fs)

			afero.WriteFile(fs, "/original.sh", []byte(originalContent), 0o644)
			afero.WriteFile(fs, "/generated.sh", []byte(generatedContent), 0o644)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := comp.Compare("/original.sh", "/generated.sh", format.format)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark real-world scenario: shell configuration files
func BenchmarkComparator_Compare_RealWorld(b *testing.B) {
	scenarios := []struct {
		name         string
		originalSize int
		changeType   string
	}{
		{"SmallConfig_MinorChanges", 50, "minor"},         // .zshrc with few aliases
		{"MediumConfig_ModerateChanges", 200, "moderate"}, // typical .bashrc
		{"LargeConfig_MajorChanges", 500, "major"},        // complex shell setup
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			fs := afero.NewMemMapFs()
			comp := NewComparator(fs)

			originalContent := generateShellConfig(scenario.originalSize)
			generatedContent := applyRealWorldChanges(originalContent, scenario.changeType)

			afero.WriteFile(fs, "/original.sh", []byte(originalContent), 0o644)
			afero.WriteFile(fs, "/generated.sh", []byte(generatedContent), 0o644)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := comp.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark parseStatistics function specifically
func BenchmarkComparator_parseStatistics(b *testing.B) {
	sizes := []struct {
		name  string
		lines int
	}{
		{"Small_100", 100},
		{"Medium_1000", 1000},
		{"Large_10000", 10000},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			comp := NewComparator(afero.NewMemMapFs())
			result := domain.NewDiffResult("/original.sh", "/generated.sh", domain.DiffFormatSummary)

			original := generateLines(size.lines)
			generated := modifyLines(original, 30) // 30% modified

			diffText := ""

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				comp.parseStatistics(result, diffText, original, generated)
			}
		})
	}
}

// Benchmark splitLines function
func BenchmarkComparator_splitLines(b *testing.B) {
	sizes := []struct {
		name  string
		lines int
	}{
		{"Small_100", 100},
		{"Medium_1000", 1000},
		{"Large_10000", 10000},
		{"VeryLarge_100000", 100000},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			content := generateFileContent(size.lines)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = splitLines(content)
			}
		})
	}
}

// Helper functions for benchmark data generation

// generateFileContent generates file content with specified number of lines
func generateFileContent(lines int) string {
	var sb strings.Builder
	for i := 0; i < lines; i++ {
		sb.WriteString(fmt.Sprintf("# Line %d: This is a sample line for testing purposes\n", i+1))
	}
	return sb.String()
}

// generateLines generates a slice of lines
func generateLines(count int) []string {
	lines := make([]string, count)
	for i := 0; i < count; i++ {
		lines[i] = fmt.Sprintf("# Line %d: This is a sample line for testing purposes", i+1)
	}
	return lines
}

// modifyContent modifies a percentage of lines in the content
func modifyContent(content string, modifyPct int) string {
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	modifyCount := len(lines) * modifyPct / 100

	for i := 0; i < modifyCount && i < len(lines); i++ {
		// Modify every nth line
		idx := i * (100 / modifyPct)
		if idx < len(lines) {
			lines[idx] = lines[idx] + " # MODIFIED"
		}
	}

	return strings.Join(lines, "\n") + "\n"
}

// modifyLines modifies a percentage of lines in a slice
func modifyLines(lines []string, modifyPct int) []string {
	result := make([]string, len(lines))
	copy(result, lines)

	modifyCount := len(lines) * modifyPct / 100
	for i := 0; i < modifyCount && i < len(result); i++ {
		idx := i * (100 / modifyPct)
		if idx < len(result) {
			result[idx] = result[idx] + " # MODIFIED"
		}
	}

	return result
}

// generateShellConfig generates realistic shell configuration content
func generateShellConfig(lines int) string {
	var sb strings.Builder

	// Header
	sb.WriteString("#!/bin/bash\n")
	sb.WriteString("# Shell configuration file\n\n")

	sectionSize := lines / 5
	remaining := lines - 2 // Already wrote 2 lines

	// Section 1: Environment variables
	sb.WriteString("# === Environment Variables ===\n")
	for i := 0; i < sectionSize && remaining > 0; i++ {
		sb.WriteString(fmt.Sprintf("export VAR_%d=\"value_%d\"\n", i, i))
		remaining--
	}
	sb.WriteString("\n")

	// Section 2: PATH setup
	sb.WriteString("# === PATH Setup ===\n")
	for i := 0; i < sectionSize && remaining > 0; i++ {
		sb.WriteString(fmt.Sprintf("export PATH=\"/path/to/bin%d:$PATH\"\n", i))
		remaining--
	}
	sb.WriteString("\n")

	// Section 3: Aliases
	sb.WriteString("# === Aliases ===\n")
	for i := 0; i < sectionSize && remaining > 0; i++ {
		sb.WriteString(fmt.Sprintf("alias cmd%d='command %d'\n", i, i))
		remaining--
	}
	sb.WriteString("\n")

	// Section 4: Functions
	sb.WriteString("# === Functions ===\n")
	for i := 0; i < sectionSize && remaining > 0; i++ {
		sb.WriteString(fmt.Sprintf("function func%d() {\n  echo \"Function %d\"\n}\n", i, i))
		remaining -= 3
	}
	sb.WriteString("\n")

	// Section 5: Remaining lines
	for remaining > 0 {
		sb.WriteString("# Additional configuration line\n")
		remaining--
	}

	return sb.String()
}

// applyRealWorldChanges simulates real-world migration changes
func applyRealWorldChanges(content string, changeType string) string {
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	var result []string

	// Add module headers (realistic migration output)
	result = append(result, "# Generated by shellforge")
	result = append(result, "# Modules: 5")
	result = append(result, "")

	switch changeType {
	case "minor":
		// 10% changes: add module comments
		for i, line := range lines {
			if i%10 == 0 {
				result = append(result, "# --- module-"+fmt.Sprint(i/10)+" ---")
			}
			result = append(result, line)
		}

	case "moderate":
		// 30% changes: reorder sections, add headers
		sections := make(map[string][]string)
		currentSection := "preamble"
		sections[currentSection] = []string{}

		for _, line := range lines {
			if strings.HasPrefix(line, "# ===") {
				currentSection = line
				sections[currentSection] = []string{}
			} else {
				sections[currentSection] = append(sections[currentSection], line)
			}
		}

		// Reorder and add module headers
		order := []string{"preamble"}
		for k := range sections {
			if k != "preamble" {
				order = append(order, k)
			}
		}

		for _, section := range order {
			if section != "preamble" {
				result = append(result, "")
				result = append(result, "# --- "+section+" ---")
			}
			result = append(result, sections[section]...)
		}

	case "major":
		// 60% changes: complete restructuring with module metadata
		for i, line := range lines {
			if i%5 == 0 {
				result = append(result, "")
				result = append(result, fmt.Sprintf("# --- module-%d ---", i/5))
				result = append(result, "#!/bin/bash")
				result = append(result, fmt.Sprintf("# Module: module-%d", i/5))
				result = append(result, "# OS: Mac, Linux")
				result = append(result, "")
			}
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n") + "\n"
}
