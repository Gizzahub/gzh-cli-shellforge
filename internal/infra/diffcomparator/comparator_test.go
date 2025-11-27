package diffcomparator

import (
	"testing"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	fs := afero.NewMemMapFs()
	comparator := New(fs)

	assert.NotNil(t, comparator)
	assert.NotNil(t, comparator.fs)
}

func TestComparator_Compare_IdenticalFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	content := "line1\nline2\nline3\n"

	afero.WriteFile(fs, "/original.sh", []byte(content), 0644)
	afero.WriteFile(fs, "/generated.sh", []byte(content), 0644)

	comparator := New(fs)
	result, err := comparator.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.NoError(t, err)
	assert.True(t, result.IsIdentical)
	assert.Equal(t, 3, result.Statistics.TotalLines)
	assert.Equal(t, 3, result.Statistics.LinesUnchanged)
	assert.Equal(t, 0, result.Statistics.LinesAdded)
	assert.Equal(t, 0, result.Statistics.LinesRemoved)
	assert.Equal(t, "Files are identical", result.Content)
}

func TestComparator_Compare_DifferentFiles(t *testing.T) {
	fs := afero.NewMemMapFs()

	original := "line1\nline2\nline3\n"
	generated := "line1\nline2_modified\nline3\nline4\n"

	afero.WriteFile(fs, "/original.sh", []byte(original), 0644)
	afero.WriteFile(fs, "/generated.sh", []byte(generated), 0644)

	comparator := New(fs)
	result, err := comparator.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.NoError(t, err)
	assert.False(t, result.IsIdentical)
	assert.Greater(t, result.Statistics.TotalLines, 0)
	assert.Greater(t, result.Statistics.CalculateTotalChanges(), 0)
}

func TestComparator_Compare_MissingOriginalFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/generated.sh", []byte("content"), 0644)

	comparator := New(fs)
	_, err := comparator.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read original file")
}

func TestComparator_Compare_MissingGeneratedFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/original.sh", []byte("content"), 0644)

	comparator := New(fs)
	_, err := comparator.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read generated file")
}

func TestComparator_Compare_SummaryFormat(t *testing.T) {
	fs := afero.NewMemMapFs()

	original := "line1\nline2\nline3\n"
	generated := "line1\nline3\nline4\n"

	afero.WriteFile(fs, "/original.sh", []byte(original), 0644)
	afero.WriteFile(fs, "/generated.sh", []byte(generated), 0644)

	comparator := New(fs)
	result, err := comparator.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.NoError(t, err)
	assert.Contains(t, result.Content, "Comparing:")
	assert.Contains(t, result.Content, "Statistics:")
	assert.Contains(t, result.Content, "Total lines:")
	assert.Contains(t, result.Content, "Lines added:")
	assert.Contains(t, result.Content, "Lines removed:")
	assert.Contains(t, result.Content, "Summary:")
}

func TestComparator_Compare_UnifiedFormat(t *testing.T) {
	fs := afero.NewMemMapFs()

	original := "line1\nline2\nline3\n"
	generated := "line1\nline2_modified\nline3\nline4\n"

	afero.WriteFile(fs, "/original.sh", []byte(original), 0644)
	afero.WriteFile(fs, "/generated.sh", []byte(generated), 0644)

	comparator := New(fs)
	result, err := comparator.Compare("/original.sh", "/generated.sh", domain.DiffFormatUnified)

	require.NoError(t, err)
	assert.Contains(t, result.Content, "--- /original.sh")
	assert.Contains(t, result.Content, "+++ /generated.sh")
	// Unified format uses +/- prefixes
	assert.Contains(t, result.Content, "+")
	assert.Contains(t, result.Content, "-")
}

func TestComparator_Compare_ContextFormat(t *testing.T) {
	fs := afero.NewMemMapFs()

	original := "line1\nline2\nline3\n"
	generated := "line1\nline2_modified\nline3\nline4\n"

	afero.WriteFile(fs, "/original.sh", []byte(original), 0644)
	afero.WriteFile(fs, "/generated.sh", []byte(generated), 0644)

	comparator := New(fs)
	result, err := comparator.Compare("/original.sh", "/generated.sh", domain.DiffFormatContext)

	require.NoError(t, err)
	assert.Contains(t, result.Content, "*** /original.sh")
	assert.Contains(t, result.Content, "--- /generated.sh")
	// Context format uses +/- prefixes
	assert.Contains(t, result.Content, "+")
	assert.Contains(t, result.Content, "-")
}

func TestComparator_Compare_SideBySideFormat(t *testing.T) {
	fs := afero.NewMemMapFs()

	original := "line1\nline2\nline3\n"
	generated := "line1\nline2_modified\nline3\nline4\n"

	afero.WriteFile(fs, "/original.sh", []byte(original), 0644)
	afero.WriteFile(fs, "/generated.sh", []byte(generated), 0644)

	comparator := New(fs)
	result, err := comparator.Compare("/original.sh", "/generated.sh", domain.DiffFormatSideBySide)

	require.NoError(t, err)
	// Side-by-side format uses |
	assert.Contains(t, result.Content, "|")
	assert.Contains(t, result.Content, "+")
	assert.Contains(t, result.Content, "-")
}

func TestComparator_LongestCommonSubsequence(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected []int
	}{
		{
			name:     "empty slices",
			a:        []string{},
			b:        []string{},
			expected: []int{},
		},
		{
			name:     "identical slices",
			a:        []string{"line1", "line2", "line3"},
			b:        []string{"line1", "line2", "line3"},
			expected: []int{0, 1, 2},
		},
		{
			name:     "no common elements",
			a:        []string{"line1", "line2"},
			b:        []string{"line3", "line4"},
			expected: nil,
		},
		{
			name:     "some common elements",
			a:        []string{"line1", "line2", "line3"},
			b:        []string{"line1", "line3"},
			expected: []int{0, 2},
		},
		{
			name:     "interleaved common elements",
			a:        []string{"a", "b", "c", "d"},
			b:        []string{"b", "d"},
			expected: []int{1, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			comparator := New(fs)
			result := comparator.longestCommonSubsequence(tt.a, tt.b)
			if tt.expected == nil {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "empty string",
			content:  "",
			expected: []string{},
		},
		{
			name:     "single line",
			content:  "line1",
			expected: []string{"line1"},
		},
		{
			name:     "multiple lines",
			content:  "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "lines with trailing newline",
			content:  "line1\nline2\n",
			expected: []string{"line1", "line2"},
		},
		{
			name:     "empty lines",
			content:  "line1\n\nline3",
			expected: []string{"line1", "", "line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a greater than b",
			a:        10,
			b:        5,
			expected: 10,
		},
		{
			name:     "b greater than a",
			a:        5,
			b:        10,
			expected: 10,
		},
		{
			name:     "equal values",
			a:        5,
			b:        5,
			expected: 5,
		},
		{
			name:     "negative values",
			a:        -5,
			b:        -10,
			expected: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := max(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComparator_CalculateDiff(t *testing.T) {
	fs := afero.NewMemMapFs()
	comparator := New(fs)

	tests := []struct {
		name             string
		original         []string
		generated        []string
		expectedAdded    int
		expectedRemoved  int
		expectedUnchanged int
	}{
		{
			name:              "identical content",
			original:          []string{"line1", "line2", "line3"},
			generated:         []string{"line1", "line2", "line3"},
			expectedAdded:     0,
			expectedRemoved:   0,
			expectedUnchanged: 3,
		},
		{
			name:              "only additions",
			original:          []string{"line1", "line2"},
			generated:         []string{"line1", "line2", "line3"},
			expectedAdded:     1,
			expectedRemoved:   0,
			expectedUnchanged: 2,
		},
		{
			name:              "only removals",
			original:          []string{"line1", "line2", "line3"},
			generated:         []string{"line1", "line3"},
			expectedAdded:     0,
			expectedRemoved:   1,
			expectedUnchanged: 2,
		},
		{
			name:              "mixed changes",
			original:          []string{"line1", "line2", "line3"},
			generated:         []string{"line1", "line3", "line4"},
			expectedAdded:     1,
			expectedRemoved:   1,
			expectedUnchanged: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, _ := comparator.calculateDiff(tt.original, tt.generated)
			assert.Equal(t, tt.expectedAdded, stats.LinesAdded, "lines added")
			assert.Equal(t, tt.expectedRemoved, stats.LinesRemoved, "lines removed")
			assert.Equal(t, tt.expectedUnchanged, stats.LinesUnchanged, "lines unchanged")
		})
	}
}
