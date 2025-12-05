// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package errors provides shellforge-specific error types and utilities.
package errors

import (
	"errors"
	"fmt"
)

// Common sentinel errors for shellforge operations.
var (
	// ErrNotFound indicates a requested resource was not found.
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists indicates a resource already exists.
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidInput indicates invalid user input.
	ErrInvalidInput = errors.New("invalid input")

	// ErrPermissionDenied indicates insufficient permissions.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrInvalidPath indicates an invalid file or directory path.
	ErrInvalidPath = errors.New("invalid path")

	// ErrCommandFailed indicates a shell command execution failure.
	ErrCommandFailed = errors.New("command execution failed")

	// ErrBackupFailed indicates a backup operation failure.
	ErrBackupFailed = errors.New("backup operation failed")

	// ErrSyncFailed indicates a synchronization failure.
	ErrSyncFailed = errors.New("synchronization failed")

	// ErrTemplateFailed indicates a template processing failure.
	ErrTemplateFailed = errors.New("template processing failed")

	// ErrDotfileInvalid indicates an invalid dotfile.
	ErrDotfileInvalid = errors.New("invalid dotfile")

	// ErrShellConfigInvalid indicates an invalid shell configuration.
	ErrShellConfigInvalid = errors.New("invalid shell configuration")
)

// Error represents a shellforge error with additional context.
type Error struct {
	Op      string // operation being performed
	Path    string // file path (if applicable)
	Command string // shell command (if applicable)
	Err     error  // underlying error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Path != "" && e.Command != "" {
		return fmt.Sprintf("%s: path=%s, command=%s: %v", e.Op, e.Path, e.Command, e.Err)
	}
	if e.Path != "" {
		return fmt.Sprintf("%s: path=%s: %v", e.Op, e.Path, e.Err)
	}
	if e.Command != "" {
		return fmt.Sprintf("%s: command=%s: %v", e.Op, e.Command, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Err
}

// Wrap wraps an error with operation context.
func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{
		Op:  op,
		Err: err,
	}
}

// WrapWithPath wraps an error with operation and path context.
func WrapWithPath(op, path string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

// WrapWithCommand wraps an error with operation and command context.
func WrapWithCommand(op, command string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{
		Op:      op,
		Command: command,
		Err:     err,
	}
}

// WrapWithContext wraps an error with full context.
func WrapWithContext(op, path, command string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{
		Op:      op,
		Path:    path,
		Command: command,
		Err:     err,
	}
}

// Is checks if the error matches the target error.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// New creates a new error with the given message.
func New(text string) error {
	return errors.New(text)
}
