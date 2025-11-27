package app

import (
	"errors"
	"testing"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDiffComparator is a mock implementation of DiffComparator.
type MockDiffComparator struct {
	mock.Mock
}

func (m *MockDiffComparator) Compare(originalPath, generatedPath string, format domain.DiffFormat) (*domain.DiffResult, error) {
	args := m.Called(originalPath, generatedPath, format)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DiffResult), args.Error(1)
}

// MockFileReader is a mock implementation of FileReader.
type MockFileReader struct {
	mock.Mock
}

func (m *MockFileReader) ReadFile(path string) (string, error) {
	args := m.Called(path)
	return args.String(0), args.Error(1)
}

func (m *MockFileReader) FileExists(path string) bool {
	args := m.Called(path)
	return args.Bool(0)
}

func TestNewDiffService(t *testing.T) {
	comparator := new(MockDiffComparator)
	reader := new(MockFileReader)

	service := NewDiffService(comparator, reader)

	assert.NotNil(t, service)
	assert.NotNil(t, service.comparator)
	assert.NotNil(t, service.reader)
}

func TestDiffService_Compare_Success(t *testing.T) {
	comparator := new(MockDiffComparator)
	reader := new(MockFileReader)

	// Setup mocks
	reader.On("FileExists", "/original.sh").Return(true)
	reader.On("FileExists", "/generated.sh").Return(true)

	expectedResult := &domain.DiffResult{
		OriginalFile:  "/original.sh",
		GeneratedFile: "/generated.sh",
		Format:        domain.DiffFormatSummary,
		Statistics: domain.DiffStatistics{
			LinesAdded:   5,
			LinesRemoved: 3,
			TotalLines:   100,
		},
	}
	comparator.On("Compare", "/original.sh", "/generated.sh", domain.DiffFormatSummary).
		Return(expectedResult, nil)

	service := NewDiffService(comparator, reader)
	result, err := service.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.DiffResult)
	assert.Equal(t, "/original.sh", result.DiffResult.OriginalFile)
	assert.Equal(t, "/generated.sh", result.DiffResult.GeneratedFile)
	assert.Equal(t, 5, result.DiffResult.Statistics.LinesAdded)
	assert.Equal(t, 3, result.DiffResult.Statistics.LinesRemoved)

	comparator.AssertExpectations(t)
	reader.AssertExpectations(t)
}

func TestDiffService_Compare_InvalidFormat(t *testing.T) {
	comparator := new(MockDiffComparator)
	reader := new(MockFileReader)

	service := NewDiffService(comparator, reader)
	_, err := service.Compare("/original.sh", "/generated.sh", domain.DiffFormat("invalid"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}

func TestDiffService_Compare_OriginalFileNotExists(t *testing.T) {
	comparator := new(MockDiffComparator)
	reader := new(MockFileReader)

	// Setup mock - original file doesn't exist
	reader.On("FileExists", "/original.sh").Return(false)

	service := NewDiffService(comparator, reader)
	_, err := service.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "original file validation failed")
	assert.Contains(t, err.Error(), "file does not exist")

	reader.AssertExpectations(t)
}

func TestDiffService_Compare_GeneratedFileNotExists(t *testing.T) {
	comparator := new(MockDiffComparator)
	reader := new(MockFileReader)

	// Setup mocks
	reader.On("FileExists", "/original.sh").Return(true)
	reader.On("FileExists", "/generated.sh").Return(false)

	service := NewDiffService(comparator, reader)
	_, err := service.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "generated file validation failed")
	assert.Contains(t, err.Error(), "file does not exist")

	reader.AssertExpectations(t)
}

func TestDiffService_Compare_FileNotExist(t *testing.T) {
	comparator := new(MockDiffComparator)
	reader := new(MockFileReader)

	// Setup mock - file doesn't exist
	reader.On("FileExists", "/original.sh").Return(false)

	service := NewDiffService(comparator, reader)
	_, err := service.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "original file validation failed")
	assert.Contains(t, err.Error(), "file does not exist")

	reader.AssertExpectations(t)
}

func TestDiffService_Compare_ComparatorError(t *testing.T) {
	comparator := new(MockDiffComparator)
	reader := new(MockFileReader)

	// Setup mocks
	reader.On("FileExists", "/original.sh").Return(true)
	reader.On("FileExists", "/generated.sh").Return(true)
	comparator.On("Compare", "/original.sh", "/generated.sh", domain.DiffFormatSummary).
		Return(nil, errors.New("comparison failed"))

	service := NewDiffService(comparator, reader)
	result, err := service.Compare("/original.sh", "/generated.sh", domain.DiffFormatSummary)

	require.Error(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Error)
	assert.Contains(t, err.Error(), "comparison failed")

	comparator.AssertExpectations(t)
	reader.AssertExpectations(t)
}

func TestDiffService_Compare_AllFormats(t *testing.T) {
	formats := []domain.DiffFormat{
		domain.DiffFormatSummary,
		domain.DiffFormatUnified,
		domain.DiffFormatContext,
		domain.DiffFormatSideBySide,
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			comparator := new(MockDiffComparator)
			reader := new(MockFileReader)

			// Setup reader mocks
			reader.On("FileExists", "/original.sh").Return(true)
			reader.On("FileExists", "/generated.sh").Return(true)

			expectedResult := &domain.DiffResult{
				OriginalFile:  "/original.sh",
				GeneratedFile: "/generated.sh",
				Format:        format,
			}
			comparator.On("Compare", "/original.sh", "/generated.sh", format).
				Return(expectedResult, nil)

			service := NewDiffService(comparator, reader)
			result, err := service.Compare("/original.sh", "/generated.sh", format)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, format, result.DiffResult.Format)

			comparator.AssertExpectations(t)
			reader.AssertExpectations(t)
		})
	}
}

func TestDiffService_ValidateFileExists(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		exists      bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "file exists",
			path:        "/existing.sh",
			exists:      true,
			expectError: false,
		},
		{
			name:        "file does not exist",
			path:        "/missing.sh",
			exists:      false,
			expectError: true,
			errorMsg:    "file does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comparator := new(MockDiffComparator)
			reader := new(MockFileReader)

			reader.On("FileExists", tt.path).Return(tt.exists)

			service := NewDiffService(comparator, reader)
			err := service.validateFileExists(tt.path)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}

			reader.AssertExpectations(t)
		})
	}
}
