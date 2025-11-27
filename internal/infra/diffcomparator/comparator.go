package diffcomparator

import (
	"fmt"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/afero"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// Comparator compares two files and generates diff results
type Comparator struct {
	fs afero.Fs
}

// NewComparator creates a new file comparator
func NewComparator(fs afero.Fs) *Comparator {
	return &Comparator{fs: fs}
}

// Compare compares two files and returns the diff result
func (c *Comparator) Compare(originalPath, generatedPath string, format domain.DiffFormat) (*domain.DiffResult, error) {
	// Read both files
	originalContent, err := afero.ReadFile(c.fs, originalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read original file: %w", err)
	}

	generatedContent, err := afero.ReadFile(c.fs, generatedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read generated file: %w", err)
	}

	// Convert to strings
	originalStr := string(originalContent)
	generatedStr := string(generatedContent)

	// Split into lines
	originalLines := splitLines(originalStr)
	generatedLines := splitLines(generatedStr)

	// Create diff result
	result := domain.NewDiffResult(originalPath, generatedPath, format)

	// Check if files are identical
	if originalStr == generatedStr {
		result.IsIdentical = true
		result.Statistics.TotalLines = len(originalLines)
		result.Statistics.LinesUnchanged = len(originalLines)
		result.Content = "Files are identical\n"
		return result, nil
	}

	// Generate diff based on format
	switch format {
	case domain.DiffFormatSummary:
		return c.generateSummary(result, originalLines, generatedLines)
	case domain.DiffFormatUnified:
		return c.generateUnified(result, originalLines, generatedLines)
	case domain.DiffFormatContext:
		return c.generateContext(result, originalLines, generatedLines)
	case domain.DiffFormatSideBySide:
		return c.generateSideBySide(result, originalLines, generatedLines)
	default:
		return nil, fmt.Errorf("unsupported diff format: %s", format)
	}
}

// generateSummary generates summary format (statistics only)
func (c *Comparator) generateSummary(result *domain.DiffResult, original, generated []string) (*domain.DiffResult, error) {
	// Calculate statistics using unified diff
	diff := difflib.UnifiedDiff{
		A:        original,
		B:        generated,
		FromFile: result.OriginalFile,
		ToFile:   result.GeneratedFile,
		Context:  0,
	}

	diffText, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return nil, fmt.Errorf("failed to generate diff: %w", err)
	}

	// Parse diff to extract statistics
	c.parseStatistics(result, diffText, original, generated)

	// Generate summary content
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Original:  %s (%d lines)\n", result.OriginalFile, len(original)))
	sb.WriteString(fmt.Sprintf("Generated: %s (%d lines)\n\n", result.GeneratedFile, len(generated)))
	sb.WriteString("Statistics:\n")
	sb.WriteString(fmt.Sprintf("  Added:      +%d lines\n", result.Statistics.LinesAdded))
	sb.WriteString(fmt.Sprintf("  Removed:    -%d lines\n", result.Statistics.LinesRemoved))
	sb.WriteString(fmt.Sprintf("  Modified:   ~%d lines\n", result.Statistics.LinesModified))
	sb.WriteString(fmt.Sprintf("  Unchanged:   %d lines\n", result.Statistics.LinesUnchanged))
	sb.WriteString(fmt.Sprintf("  Total:       %d lines\n\n", result.Statistics.TotalLines))
	sb.WriteString(fmt.Sprintf("Change rate: %.1f%%\n", result.Statistics.ChangePercentage()))

	result.Content = sb.String()
	return result, nil
}

// generateUnified generates unified diff format (git diff style)
func (c *Comparator) generateUnified(result *domain.DiffResult, original, generated []string) (*domain.DiffResult, error) {
	diff := difflib.UnifiedDiff{
		A:        original,
		B:        generated,
		FromFile: result.OriginalFile,
		ToFile:   result.GeneratedFile,
		Context:  3,
	}

	diffText, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return nil, fmt.Errorf("failed to generate unified diff: %w", err)
	}

	c.parseStatistics(result, diffText, original, generated)
	result.Content = diffText
	return result, nil
}

// generateContext generates context diff format
func (c *Comparator) generateContext(result *domain.DiffResult, original, generated []string) (*domain.DiffResult, error) {
	diff := difflib.ContextDiff{
		A:        original,
		B:        generated,
		FromFile: result.OriginalFile,
		ToFile:   result.GeneratedFile,
		Context:  3,
	}

	diffText, err := difflib.GetContextDiffString(diff)
	if err != nil {
		return nil, fmt.Errorf("failed to generate context diff: %w", err)
	}

	// Parse unified diff for statistics
	unifiedDiff := difflib.UnifiedDiff{
		A:        original,
		B:        generated,
		FromFile: result.OriginalFile,
		ToFile:   result.GeneratedFile,
		Context:  0,
	}
	unifiedText, _ := difflib.GetUnifiedDiffString(unifiedDiff)
	c.parseStatistics(result, unifiedText, original, generated)

	result.Content = diffText
	return result, nil
}

// generateSideBySide generates side-by-side comparison format
func (c *Comparator) generateSideBySide(result *domain.DiffResult, original, generated []string) (*domain.DiffResult, error) {
	// Generate unified diff for statistics
	unifiedDiff := difflib.UnifiedDiff{
		A:        original,
		B:        generated,
		FromFile: result.OriginalFile,
		ToFile:   result.GeneratedFile,
		Context:  0,
	}
	unifiedText, _ := difflib.GetUnifiedDiffString(unifiedDiff)
	c.parseStatistics(result, unifiedText, original, generated)

	// Generate side-by-side view
	var sb strings.Builder
	maxLen := len(original)
	if len(generated) > maxLen {
		maxLen = len(generated)
	}

	// Header
	sb.WriteString(fmt.Sprintf("%-40s | %-40s\n", result.OriginalFile, result.GeneratedFile))
	sb.WriteString(strings.Repeat("-", 83) + "\n")

	// Line-by-line comparison
	for i := 0; i < maxLen; i++ {
		var origLine, genLine string

		if i < len(original) {
			origLine = truncate(original[i], 40)
		} else {
			origLine = ""
		}

		if i < len(generated) {
			genLine = truncate(generated[i], 40)
		} else {
			genLine = ""
		}

		// Determine line status
		marker := " "
		if i < len(original) && i < len(generated) {
			if original[i] != generated[i] {
				marker = "~"
			}
		} else if i >= len(original) {
			marker = "+"
		} else {
			marker = "-"
		}

		sb.WriteString(fmt.Sprintf("%-40s %s %-40s\n", origLine, marker, genLine))
	}

	result.Content = sb.String()
	return result, nil
}

// parseStatistics calculates statistics by directly comparing line arrays
func (c *Comparator) parseStatistics(result *domain.DiffResult, diffText string, original, generated []string) {
	// Direct comparison for accurate statistics
	origLen := len(original)
	genLen := len(generated)

	// Calculate unchanged, modified, added, and removed lines
	unchanged := 0
	modified := 0

	// Compare lines that exist in both files
	minLen := origLen
	if genLen < minLen {
		minLen = genLen
	}

	for i := 0; i < minLen; i++ {
		if original[i] == generated[i] {
			unchanged++
		} else {
			modified++
		}
	}

	// Calculate pure additions and removals
	added := 0
	removed := 0

	if genLen > origLen {
		added = genLen - origLen
	} else if origLen > genLen {
		removed = origLen - genLen
	}

	// Total lines is the max of both
	totalLines := origLen
	if genLen > totalLines {
		totalLines = genLen
	}

	result.Statistics = domain.DiffStatistics{
		LinesAdded:     added,
		LinesRemoved:   removed,
		LinesModified:  modified,
		LinesUnchanged: unchanged,
		TotalLines:     totalLines,
	}
}

// splitLines splits content into lines, preserving empty lines
func splitLines(content string) []string {
	if content == "" {
		return []string{}
	}

	lines := strings.Split(content, "\n")

	// Remove trailing empty line if present (from final newline)
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
