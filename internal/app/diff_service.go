package app

import (
	"fmt"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// DiffComparator defines the interface for comparing files.
type DiffComparator interface {
	Compare(originalPath, generatedPath string, format domain.DiffFormat) (*domain.DiffResult, error)
}

// DiffService orchestrates file comparison operations.
type DiffService struct {
	comparator DiffComparator
	reader     FileReader
}

// NewDiffService creates a new DiffService.
func NewDiffService(comparator DiffComparator, reader FileReader) *DiffService {
	return &DiffService{
		comparator: comparator,
		reader:     reader,
	}
}

// CompareResult contains the result of a diff comparison operation.
type CompareResult struct {
	DiffResult *domain.DiffResult
	Error      error
}

// Compare compares two files and returns the diff result.
func (s *DiffService) Compare(originalPath, generatedPath string, format domain.DiffFormat) (*CompareResult, error) {
	// Validate format
	if err := domain.ValidateFormat(string(format)); err != nil {
		return nil, fmt.Errorf("invalid format: %w", err)
	}

	// Check if files exist
	if err := s.validateFileExists(originalPath); err != nil {
		return nil, fmt.Errorf("original file validation failed: %w", err)
	}

	if err := s.validateFileExists(generatedPath); err != nil {
		return nil, fmt.Errorf("generated file validation failed: %w", err)
	}

	// Perform comparison
	diffResult, err := s.comparator.Compare(originalPath, generatedPath, format)
	if err != nil {
		return &CompareResult{Error: err}, err
	}

	return &CompareResult{DiffResult: diffResult}, nil
}

// validateFileExists checks if a file exists and is readable.
func (s *DiffService) validateFileExists(path string) error {
	exists := s.reader.FileExists(path)
	if !exists {
		return fmt.Errorf("file does not exist: %s", path)
	}

	return nil
}
