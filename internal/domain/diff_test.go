package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDiffResult(t *testing.T) {
	result := NewDiffResult("original.sh", "generated.sh", DiffFormatSummary)

	assert.Equal(t, "original.sh", result.OriginalFile)
	assert.Equal(t, "generated.sh", result.GeneratedFile)
	assert.Equal(t, DiffFormatSummary, result.Format)
	assert.Equal(t, 0, result.Statistics.LinesAdded)
	assert.Equal(t, 0, result.Statistics.LinesRemoved)
	assert.Equal(t, 0, result.Statistics.LinesModified)
	assert.False(t, result.IsIdentical)
}

func TestDiffStatistics_CalculateTotalChanges(t *testing.T) {
	tests := []struct {
		name     string
		stats    DiffStatistics
		expected int
	}{
		{
			name: "no changes",
			stats: DiffStatistics{
				LinesAdded:    0,
				LinesRemoved:  0,
				LinesModified: 0,
			},
			expected: 0,
		},
		{
			name: "only additions",
			stats: DiffStatistics{
				LinesAdded:    5,
				LinesRemoved:  0,
				LinesModified: 0,
			},
			expected: 5,
		},
		{
			name: "only removals",
			stats: DiffStatistics{
				LinesAdded:    0,
				LinesRemoved:  3,
				LinesModified: 0,
			},
			expected: 3,
		},
		{
			name: "only modifications",
			stats: DiffStatistics{
				LinesAdded:    0,
				LinesRemoved:  0,
				LinesModified: 2,
			},
			expected: 2,
		},
		{
			name: "mixed changes",
			stats: DiffStatistics{
				LinesAdded:    5,
				LinesRemoved:  3,
				LinesModified: 2,
			},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := tt.stats.CalculateTotalChanges()
			assert.Equal(t, tt.expected, total)
		})
	}
}

func TestDiffStatistics_ChangePercentage(t *testing.T) {
	tests := []struct {
		name     string
		stats    DiffStatistics
		expected float64
	}{
		{
			name: "zero total lines",
			stats: DiffStatistics{
				LinesAdded:    5,
				LinesRemoved:  3,
				LinesModified: 2,
				TotalLines:    0,
			},
			expected: 0.0,
		},
		{
			name: "no changes",
			stats: DiffStatistics{
				LinesAdded:    0,
				LinesRemoved:  0,
				LinesModified: 0,
				TotalLines:    100,
			},
			expected: 0.0,
		},
		{
			name: "50% changed",
			stats: DiffStatistics{
				LinesAdded:    25,
				LinesRemoved:  15,
				LinesModified: 10,
				TotalLines:    100,
			},
			expected: 50.0,
		},
		{
			name: "100% changed",
			stats: DiffStatistics{
				LinesAdded:    50,
				LinesRemoved:  30,
				LinesModified: 20,
				TotalLines:    100,
			},
			expected: 100.0,
		},
		{
			name: "33.3% changed",
			stats: DiffStatistics{
				LinesAdded:    10,
				LinesRemoved:  0,
				LinesModified: 0,
				TotalLines:    30,
			},
			expected: 33.333333333333336,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			percentage := tt.stats.ChangePercentage()
			assert.InDelta(t, tt.expected, percentage, 0.0001)
		})
	}
}

func TestDiffStatistics_Summary(t *testing.T) {
	tests := []struct {
		name     string
		stats    DiffStatistics
		expected string
	}{
		{
			name: "no changes",
			stats: DiffStatistics{
				LinesAdded:    0,
				LinesRemoved:  0,
				LinesModified: 0,
				TotalLines:    100,
			},
			expected: "+0 -0 ~0 (0.0% changed)",
		},
		{
			name: "mixed changes",
			stats: DiffStatistics{
				LinesAdded:    25,
				LinesRemoved:  15,
				LinesModified: 10,
				TotalLines:    100,
			},
			expected: "+25 -15 ~10 (50.0% changed)",
		},
		{
			name: "only additions",
			stats: DiffStatistics{
				LinesAdded:    10,
				LinesRemoved:  0,
				LinesModified: 0,
				TotalLines:    30,
			},
			expected: "+10 -0 ~0 (33.3% changed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.stats.Summary()
			assert.Equal(t, tt.expected, summary)
		})
	}
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		wantError bool
	}{
		{
			name:      "valid summary format",
			format:    "summary",
			wantError: false,
		},
		{
			name:      "valid unified format",
			format:    "unified",
			wantError: false,
		},
		{
			name:      "valid context format",
			format:    "context",
			wantError: false,
		},
		{
			name:      "valid side-by-side format",
			format:    "side-by-side",
			wantError: false,
		},
		{
			name:      "invalid format",
			format:    "invalid",
			wantError: true,
		},
		{
			name:      "empty format",
			format:    "",
			wantError: true,
		},
		{
			name:      "case sensitive",
			format:    "Summary",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormat(tt.format)
			if tt.wantError {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDiffFormat_Constants(t *testing.T) {
	// Ensure constants have expected values
	assert.Equal(t, DiffFormat("summary"), DiffFormatSummary)
	assert.Equal(t, DiffFormat("unified"), DiffFormatUnified)
	assert.Equal(t, DiffFormat("context"), DiffFormatContext)
	assert.Equal(t, DiffFormat("side-by-side"), DiffFormatSideBySide)
}
