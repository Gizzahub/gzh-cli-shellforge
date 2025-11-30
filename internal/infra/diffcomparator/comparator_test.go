package diffcomparator

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

func TestNewComparator(t *testing.T) {
	fs := afero.NewMemMapFs()
	comp := NewComparator(fs)

	assert.NotNil(t, comp)
	assert.Equal(t, fs, comp.fs)
}

func TestComparator_Compare_IdenticalFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	comp := NewComparator(fs)

	// Create identical files
	content := "line1\nline2\nline3\n"
	afero.WriteFile(fs, "/original.sh", []byte(content), 0644)
	afero.WriteFile(fs, "/generated.sh", []byte(content), 0644)

	result, err := comp.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.NoError(t, err)
	assert.True(t, result.IsIdentical)
	assert.Equal(t, 3, result.Statistics.TotalLines)
	assert.Equal(t, 3, result.Statistics.LinesUnchanged)
	assert.Equal(t, 0, result.Statistics.LinesAdded)
	assert.Equal(t, 0, result.Statistics.LinesRemoved)
	assert.Equal(t, 0, result.Statistics.LinesModified)
	assert.Contains(t, result.Content, "Files are identical")
}

func TestComparator_Compare_AddedLines(t *testing.T) {
	fs := afero.NewMemMapFs()
	comp := NewComparator(fs)

	originalContent := "line1\nline2\n"
	afero.WriteFile(fs, "/original.sh", []byte(originalContent), 0644)

	generatedContent := "line1\nline2\nline3\nline4\n"
	afero.WriteFile(fs, "/generated.sh", []byte(generatedContent), 0644)

	result, err := comp.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.NoError(t, err)
	assert.False(t, result.IsIdentical)
	assert.Equal(t, 4, result.Statistics.TotalLines)
	assert.Equal(t, 2, result.Statistics.LinesAdded)
	assert.Equal(t, 0, result.Statistics.LinesRemoved)
}

func TestComparator_Compare_UnifiedFormat(t *testing.T) {
	fs := afero.NewMemMapFs()
	comp := NewComparator(fs)

	originalContent := "line1\nline2\nline3\n"
	afero.WriteFile(fs, "/original.sh", []byte(originalContent), 0644)

	generatedContent := "line1\nline2_modified\nline3\nline4\n"
	afero.WriteFile(fs, "/generated.sh", []byte(generatedContent), 0644)

	result, err := comp.Compare("/original.sh", "/generated.sh", domain.DiffFormatUnified)

	require.NoError(t, err)
	assert.False(t, result.IsIdentical)
	assert.Contains(t, result.Content, "---")
	assert.Contains(t, result.Content, "+++")
	assert.Contains(t, result.Content, "-line2")
	assert.Contains(t, result.Content, "+line2_modified")
	assert.Contains(t, result.Content, "+line4")
}

func TestComparator_Compare_ContextFormat(t *testing.T) {
	fs := afero.NewMemMapFs()
	comp := NewComparator(fs)

	originalContent := "line1\nline2\nline3\n"
	afero.WriteFile(fs, "/original.sh", []byte(originalContent), 0644)

	generatedContent := "line1\nline2_modified\nline3\n"
	afero.WriteFile(fs, "/generated.sh", []byte(generatedContent), 0644)

	result, err := comp.Compare("/original.sh", "/generated.sh", domain.DiffFormatContext)

	require.NoError(t, err)
	assert.False(t, result.IsIdentical)
	assert.Contains(t, result.Content, "***")
	assert.Contains(t, result.Content, "---")
	// Context diff uses ! for changed lines
	assert.Contains(t, result.Content, "! line2")
}

func TestComparator_Compare_SideBySideFormat(t *testing.T) {
	fs := afero.NewMemMapFs()
	comp := NewComparator(fs)

	originalContent := "line1\nline2\nline3\n"
	afero.WriteFile(fs, "/original.sh", []byte(originalContent), 0644)

	generatedContent := "line1\nline2_modified\nline3\nline4\n"
	afero.WriteFile(fs, "/generated.sh", []byte(generatedContent), 0644)

	result, err := comp.Compare("/original.sh", "/generated.sh", domain.DiffFormatSideBySide)

	require.NoError(t, err)
	assert.False(t, result.IsIdentical)
	assert.Contains(t, result.Content, "original.sh")
	assert.Contains(t, result.Content, "generated.sh")
	assert.Contains(t, result.Content, "line1")
	assert.Contains(t, result.Content, "line2")
	assert.Contains(t, result.Content, "line3")
}

func TestComparator_Compare_RemovedLines(t *testing.T) {
	fs := afero.NewMemMapFs()
	comp := NewComparator(fs)

	originalContent := "line1\nline2\nline3\nline4\n"
	afero.WriteFile(fs, "/original.sh", []byte(originalContent), 0644)

	generatedContent := "line1\nline3\n"
	afero.WriteFile(fs, "/generated.sh", []byte(generatedContent), 0644)

	result, err := comp.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.NoError(t, err)
	assert.False(t, result.IsIdentical)
	// Based on LCS algorithm, line2 and line4 are removed, line1 and line3 are unchanged
	assert.True(t, result.Statistics.LinesRemoved >= 1, "Should have at least 1 line removed")
	assert.True(t, result.Statistics.LinesUnchanged >= 1, "Should have at least 1 line unchanged")
}

func TestComparator_Compare_ModifiedLines(t *testing.T) {
	fs := afero.NewMemMapFs()
	comp := NewComparator(fs)

	originalContent := "line1\nline2\nline3\n"
	afero.WriteFile(fs, "/original.sh", []byte(originalContent), 0644)

	generatedContent := "line1_modified\nline2\nline3_modified\n"
	afero.WriteFile(fs, "/generated.sh", []byte(generatedContent), 0644)

	result, err := comp.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.NoError(t, err)
	assert.False(t, result.IsIdentical)
	assert.Equal(t, 2, result.Statistics.LinesModified)
	assert.Equal(t, 1, result.Statistics.LinesUnchanged)
}

func TestComparator_Compare_FileNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	comp := NewComparator(fs)

	_, err := comp.Compare("/nonexistent.sh", "/also-nonexistent.sh", domain.DiffFormatSummary)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read original file")
}

func TestComparator_Compare_EmptyFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	comp := NewComparator(fs)

	afero.WriteFile(fs, "/empty1.sh", []byte(""), 0644)
	afero.WriteFile(fs, "/empty2.sh", []byte(""), 0644)

	result, err := comp.Compare("/empty1.sh", "/empty2.sh", domain.DiffFormatSummary)

	require.NoError(t, err)
	assert.True(t, result.IsIdentical)
	assert.Equal(t, 0, result.Statistics.TotalLines)
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "needs truncation",
			input:    "hello world",
			maxLen:   8,
			expected: "hello...",
		},
		{
			name:     "very short max",
			input:    "hello world",
			maxLen:   3,
			expected: "...", // maxLen-3 = 0, so just "..."
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
			assert.LessOrEqual(t, len(result), tt.maxLen)
		})
	}
}
