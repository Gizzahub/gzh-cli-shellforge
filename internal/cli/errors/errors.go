package errors

import (
	"fmt"
)

// WrapError wraps an error with operation context
func WrapError(operation string, err error) error {
	return fmt.Errorf("%s failed: %w", operation, err)
}

// InvalidPath returns a standardized invalid path error
func InvalidPath(pathType string, err error) error {
	return fmt.Errorf("invalid %s path: %w", pathType, err)
}

// FileNotFound returns a standardized file not found error
func FileNotFound(path string) error {
	return fmt.Errorf("file not found: %s", path)
}

// DirNotFound returns a standardized directory not found error
func DirNotFound(path string) error {
	return fmt.Errorf("directory does not exist: %s", path)
}

// ValidationError returns a standardized validation error
func ValidationError(message string) error {
	return fmt.Errorf("validation error: %s", message)
}

// RequiredFlag returns a standardized required flag error with help
func RequiredFlag(flagName string, examples ...string) error {
	msg := fmt.Sprintf("--%s flag is required", flagName)
	if len(examples) > 0 {
		msg += "\n\nExamples:"
		for _, ex := range examples {
			msg += "\n  " + ex
		}
	}
	return fmt.Errorf("%s", msg)
}

// MutuallyExclusive returns an error for mutually exclusive flags
func MutuallyExclusive(flag1, flag2 string) error {
	return fmt.Errorf("--%s and --%s cannot be used together", flag1, flag2)
}

// MinValue returns an error for values below minimum
func MinValue(flagName string, minValue int) error {
	return fmt.Errorf("%s must be at least %d", flagName, minValue)
}
