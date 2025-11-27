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
