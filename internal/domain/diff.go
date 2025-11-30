package domain

import "fmt"

// DiffFormat represents the output format for diff results.
type DiffFormat string

const (
	// DiffFormatSummary shows only statistics (lines added/removed/modified)
	DiffFormatSummary DiffFormat = "summary"

	// DiffFormatUnified shows unified diff format (git diff style)
	DiffFormatUnified DiffFormat = "unified"

	// DiffFormatContext shows context diff format (with context lines)
	DiffFormatContext DiffFormat = "context"

	// DiffFormatSideBySide shows side-by-side comparison
	DiffFormatSideBySide DiffFormat = "side-by-side"
)

// DiffStatistics contains statistics about differences between files.
type DiffStatistics struct {
	LinesAdded     int
	LinesRemoved   int
	LinesModified  int
	LinesUnchanged int
	TotalLines     int
}

// DiffResult represents the result of comparing two files.
type DiffResult struct {
	OriginalFile  string
	GeneratedFile string
	Statistics    DiffStatistics
	Format        DiffFormat
	Content       string // Formatted diff content based on Format
	IsIdentical   bool
}

// NewDiffResult creates a new DiffResult.
func NewDiffResult(original, generated string, format DiffFormat) *DiffResult {
	return &DiffResult{
		OriginalFile:  original,
		GeneratedFile: generated,
		Format:        format,
		Statistics:    DiffStatistics{},
	}
}

// CalculateTotalChanges returns the total number of changed lines.
func (s *DiffStatistics) CalculateTotalChanges() int {
	return s.LinesAdded + s.LinesRemoved + s.LinesModified
}

// ChangePercentage calculates the percentage of changed lines.
func (s *DiffStatistics) ChangePercentage() float64 {
	if s.TotalLines == 0 {
		return 0.0
	}
	changes := float64(s.CalculateTotalChanges())
	return (changes / float64(s.TotalLines)) * 100.0
}

// Summary returns a human-readable summary of statistics.
func (s *DiffStatistics) Summary() string {
	return fmt.Sprintf(
		"+%d -%d ~%d (%.1f%% changed)",
		s.LinesAdded,
		s.LinesRemoved,
		s.LinesModified,
		s.ChangePercentage(),
	)
}

// ValidateFormat checks if the diff format is valid.
func ValidateFormat(format string) error {
	switch DiffFormat(format) {
	case DiffFormatSummary, DiffFormatUnified, DiffFormatContext, DiffFormatSideBySide:
		return nil
	default:
		return NewValidationError("invalid diff format '%s', must be one of: summary, unified, context, side-by-side", format)
	}
}
