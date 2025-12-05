// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package logger provides simple structured logging for shellforge.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// Level represents logging severity level.
type Level int

const (
	// LevelDebug for detailed debugging information.
	LevelDebug Level = iota
	// LevelInfo for general informational messages.
	LevelInfo
	// LevelWarn for warning messages.
	LevelWarn
	// LevelError for error messages.
	LevelError
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

// Logger provides structured logging functionality.
type Logger struct {
	component string
	level     Level
	output    io.Writer
	logger    *log.Logger
}

// New creates a new Logger for the given component.
func New(component string) *Logger {
	return &Logger{
		component: component,
		level:     LevelInfo,
		output:    os.Stdout,
		logger:    log.New(os.Stdout, "", 0),
	}
}

// SetLevel sets the minimum logging level.
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// SetOutput sets the output writer for the logger.
func (l *Logger) SetOutput(w io.Writer) {
	l.output = w
	l.logger = log.New(w, "", 0)
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(LevelDebug, msg, args...)
}

// Info logs an informational message.
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(LevelInfo, msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(LevelWarn, msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(LevelError, msg, args...)
}

// Fatal logs an error message and exits the program.
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log(LevelError, msg, args...)
	os.Exit(1)
}

// log formats and writes a log message.
func (l *Logger) log(level Level, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := levelNames[level]

	var message string
	if len(args) > 0 {
		message = fmt.Sprintf(msg, args...)
	} else {
		message = msg
	}

	// Add key-value pairs if present
	var kvPairs []string
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key := fmt.Sprint(args[i])
			val := fmt.Sprint(args[i+1])
			kvPairs = append(kvPairs, fmt.Sprintf("%s=%s", key, val))
		}
	}

	var formatted string
	if len(kvPairs) > 0 {
		formatted = fmt.Sprintf("[%s] %-5s [%s] %s | %s",
			timestamp, levelName, l.component, message, strings.Join(kvPairs, " "))
	} else {
		formatted = fmt.Sprintf("[%s] %-5s [%s] %s",
			timestamp, levelName, l.component, message)
	}

	l.logger.Println(formatted)
}

// WithComponent creates a new logger with a different component name.
func (l *Logger) WithComponent(component string) *Logger {
	newLogger := &Logger{
		component: component,
		level:     l.level,
		output:    l.output,
		logger:    log.New(l.output, "", 0),
	}
	return newLogger
}

// Default logger instance.
var defaultLogger = New("shellforge")

// SetDefaultLevel sets the level for the default logger.
func SetDefaultLevel(level Level) {
	defaultLogger.SetLevel(level)
}

// SetDefaultOutput sets the output for the default logger.
func SetDefaultOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
}

// Debug logs a debug message using the default logger.
func Debug(msg string, args ...interface{}) {
	defaultLogger.Debug(msg, args...)
}

// Info logs an informational message using the default logger.
func Info(msg string, args ...interface{}) {
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning message using the default logger.
func Warn(msg string, args ...interface{}) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message using the default logger.
func Error(msg string, args ...interface{}) {
	defaultLogger.Error(msg, args...)
}

// Fatal logs an error message using the default logger and exits.
func Fatal(msg string, args ...interface{}) {
	defaultLogger.Fatal(msg, args...)
}
