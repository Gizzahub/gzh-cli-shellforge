package diffcomparator

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/spf13/afero"
)

// Comparator implements file comparison with multiple output formats.
type Comparator struct {
	fs afero.Fs
}

// New creates a new Comparator.
func New(fs afero.Fs) *Comparator {
	return &Comparator{fs: fs}
}

// Compare compares two files and returns a DiffResult with the specified format.
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

	// Create result
	result := domain.NewDiffResult(originalPath, generatedPath, format)

	// Split into lines
	originalLines := splitLines(string(originalContent))
	generatedLines := splitLines(string(generatedContent))

	// Check if identical
	if string(originalContent) == string(generatedContent) {
		result.IsIdentical = true
		result.Statistics.TotalLines = len(originalLines)
		result.Statistics.LinesUnchanged = len(originalLines)
		result.Content = "Files are identical"
		return result, nil
	}

	// Calculate diff using simple line-by-line comparison
	stats, diffLines := c.calculateDiff(originalLines, generatedLines)
	result.Statistics = stats

	// Format output based on requested format
	switch format {
	case domain.DiffFormatSummary:
		result.Content = c.formatSummary(result)
	case domain.DiffFormatUnified:
		result.Content = c.formatUnified(originalPath, generatedPath, diffLines)
	case domain.DiffFormatContext:
		result.Content = c.formatContext(originalPath, generatedPath, diffLines)
	case domain.DiffFormatSideBySide:
		result.Content = c.formatSideBySide(diffLines)
	default:
		return nil, fmt.Errorf("unsupported diff format: %s", format)
	}

	return result, nil
}

// calculateDiff performs a simple line-by-line diff using longest common subsequence.
func (c *Comparator) calculateDiff(original, generated []string) (domain.DiffStatistics, []diffLine) {
	stats := domain.DiffStatistics{}
	var diffLines []diffLine

	// Simple LCS-based diff algorithm
	lcs := c.longestCommonSubsequence(original, generated)
	lcsSet := make(map[int]bool)
	for _, idx := range lcs {
		lcsSet[idx] = true
	}

	// Track which generated lines have been used
	genIdx := 0
	genUsed := make(map[int]bool)

	// Process original lines
	for i, line := range original {
		if lcsSet[i] {
			// Line is unchanged
			diffLines = append(diffLines, diffLine{
				lineType: lineUnchanged,
				content:  line,
				lineNum1: i + 1,
				lineNum2: genIdx + 1,
			})
			stats.LinesUnchanged++
			genUsed[genIdx] = true
			genIdx++
		} else {
			// Line was removed
			diffLines = append(diffLines, diffLine{
				lineType: lineRemoved,
				content:  line,
				lineNum1: i + 1,
				lineNum2: 0,
			})
			stats.LinesRemoved++
		}
	}

	// Process remaining generated lines (additions)
	for i, line := range generated {
		if !genUsed[i] {
			diffLines = append(diffLines, diffLine{
				lineType: lineAdded,
				content:  line,
				lineNum1: 0,
				lineNum2: i + 1,
			})
			stats.LinesAdded++
		}
	}

	stats.TotalLines = max(len(original), len(generated))
	return stats, diffLines
}

// longestCommonSubsequence finds the LCS between two slices of strings.
// Returns indices in the original slice that are part of the LCS.
func (c *Comparator) longestCommonSubsequence(a, b []string) []int {
	m, n := len(a), len(b)
	if m == 0 || n == 0 {
		return []int{}
	}

	// Build LCS table
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	// Backtrack to find LCS indices
	var result []int
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			result = append([]int{i - 1}, result...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return result
}

// formatSummary formats the result as a summary with statistics.
func (c *Comparator) formatSummary(result *domain.DiffResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Comparing:\n"))
	sb.WriteString(fmt.Sprintf("  Original:  %s\n", result.OriginalFile))
	sb.WriteString(fmt.Sprintf("  Generated: %s\n\n", result.GeneratedFile))
	sb.WriteString(fmt.Sprintf("Statistics:\n"))
	sb.WriteString(fmt.Sprintf("  Total lines:    %d\n", result.Statistics.TotalLines))
	sb.WriteString(fmt.Sprintf("  Lines added:    %d\n", result.Statistics.LinesAdded))
	sb.WriteString(fmt.Sprintf("  Lines removed:  %d\n", result.Statistics.LinesRemoved))
	sb.WriteString(fmt.Sprintf("  Lines modified: %d\n", result.Statistics.LinesModified))
	sb.WriteString(fmt.Sprintf("  Lines unchanged: %d\n\n", result.Statistics.LinesUnchanged))
	sb.WriteString(fmt.Sprintf("Summary: %s\n", result.Statistics.Summary()))
	return sb.String()
}

// formatUnified formats the diff in unified format (git diff style).
func (c *Comparator) formatUnified(originalPath, generatedPath string, lines []diffLine) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- %s\n", originalPath))
	sb.WriteString(fmt.Sprintf("+++ %s\n", generatedPath))

	for _, line := range lines {
		switch line.lineType {
		case lineAdded:
			sb.WriteString(fmt.Sprintf("+%s\n", line.content))
		case lineRemoved:
			sb.WriteString(fmt.Sprintf("-%s\n", line.content))
		case lineUnchanged:
			sb.WriteString(fmt.Sprintf(" %s\n", line.content))
		}
	}

	return sb.String()
}

// formatContext formats the diff in context format.
func (c *Comparator) formatContext(originalPath, generatedPath string, lines []diffLine) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*** %s\n", originalPath))
	sb.WriteString(fmt.Sprintf("--- %s\n", generatedPath))

	// Context format groups changes
	for _, line := range lines {
		switch line.lineType {
		case lineAdded:
			sb.WriteString(fmt.Sprintf("+ %s\n", line.content))
		case lineRemoved:
			sb.WriteString(fmt.Sprintf("- %s\n", line.content))
		case lineUnchanged:
			sb.WriteString(fmt.Sprintf("  %s\n", line.content))
		}
	}

	return sb.String()
}

// formatSideBySide formats the diff in side-by-side format.
func (c *Comparator) formatSideBySide(lines []diffLine) string {
	var sb strings.Builder
	const maxWidth = 40

	for _, line := range lines {
		switch line.lineType {
		case lineAdded:
			sb.WriteString(fmt.Sprintf("%-*s | + %s\n", maxWidth, "", line.content))
		case lineRemoved:
			content := line.content
			if len(content) > maxWidth {
				content = content[:maxWidth-3] + "..."
			}
			sb.WriteString(fmt.Sprintf("%-*s | - %s\n", maxWidth, content, ""))
		case lineUnchanged:
			content := line.content
			if len(content) > maxWidth {
				content = content[:maxWidth-3] + "..."
			}
			sb.WriteString(fmt.Sprintf("%-*s |   %s\n", maxWidth, content, content))
		}
	}

	return sb.String()
}

// lineType represents the type of diff line.
type lineType int

const (
	lineUnchanged lineType = iota
	lineAdded
	lineRemoved
	lineModified
)

// diffLine represents a single line in the diff output.
type diffLine struct {
	lineType lineType
	content  string
	lineNum1 int // Line number in original file
	lineNum2 int // Line number in generated file
}

// splitLines splits content into lines, preserving empty lines.
func splitLines(content string) []string {
	if content == "" {
		return []string{}
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
