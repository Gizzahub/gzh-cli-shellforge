package domain

import "fmt"

// ValidationError represents a validation failure.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error.
func NewValidationError(format string, args ...interface{}) *ValidationError {
	return &ValidationError{Message: fmt.Sprintf(format, args...)}
}

// CircularDependencyError represents a circular dependency in the module graph.
type CircularDependencyError struct {
	Message string
}

func (e *CircularDependencyError) Error() string {
	return e.Message
}

// NewCircularDependencyError creates a new circular dependency error.
func NewCircularDependencyError(format string, args ...interface{}) *CircularDependencyError {
	return &CircularDependencyError{Message: fmt.Sprintf(format, args...)}
}

// FileNotFoundError represents a missing file.
type FileNotFoundError struct {
	Path string
}

func (e *FileNotFoundError) Error() string {
	return fmt.Sprintf("file not found: %s", e.Path)
}

// NewFileNotFoundError creates a new file not found error.
func NewFileNotFoundError(path string) *FileNotFoundError {
	return &FileNotFoundError{Path: path}
}
