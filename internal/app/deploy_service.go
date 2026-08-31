// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// DirectoryReader extends FileReader with directory listing.
type DirectoryReader interface {
	FileReader
	ListDir(path string) ([]string, error)
}

// BackupWriter extends FileWriter with backup and directory support.
type BackupWriter interface {
	FileWriter
	Copy(src, dst string) error
	MkdirAll(path string) error
}

// PermissionChecker verifies that a path is writable before attempting a write.
type PermissionChecker interface {
	CheckWritable(path string) error
}

// osPermissionChecker probes writability by creating and immediately removing a
// temporary file in the target directory.
type osPermissionChecker struct{}

func (c *osPermissionChecker) CheckWritable(path string) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".shellforge-check-*")
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("requires elevated privileges to write to %s — re-run with sudo", path)
		}
		return fmt.Errorf("cannot write to %s: %w", path, err)
	}
	f.Close()
	os.Remove(f.Name())
	return nil
}

// DeployService implements the deploy use case.
type DeployService struct {
	reader  DirectoryReader
	writer  BackupWriter
	checker PermissionChecker
}

// NewDeployService creates a new deploy service with the default OS permission checker.
func NewDeployService(reader DirectoryReader, writer BackupWriter) *DeployService {
	return &DeployService{
		reader:  reader,
		writer:  writer,
		checker: &osPermissionChecker{},
	}
}

// NewDeployServiceWithChecker creates a deploy service with an injectable permission checker.
// Use this in tests to supply a mock PermissionChecker.
func NewDeployServiceWithChecker(reader DirectoryReader, writer BackupWriter, checker PermissionChecker) *DeployService {
	return &DeployService{
		reader:  reader,
		writer:  writer,
		checker: checker,
	}
}

// DeployOptions contains options for deploying built configuration.
type DeployOptions struct {
	BuildDir     string // Directory containing built files (default: ./build)
	DryRun       bool   // Preview without deploying
	CreateBackup bool   // Backup existing files before overwriting
	Verbose      bool   // Show detailed output
	HomeDir      string // Home directory for path resolution
}

// DeployedFile represents a single deployed file.
type DeployedFile struct {
	SourcePath string // Path in build directory
	DestPath   string // Deployed destination path
	BackupPath string // Path to backup (if created)
	Deployed   bool   // Whether deployment succeeded
	Skipped    bool   // Whether file was skipped
	Error      error  // Error if any
}

// DeployResult contains the result of a deploy operation.
type DeployResult struct {
	DeployedFiles []DeployedFile
	TotalFiles    int
	DeployedCount int
	SkippedCount  int
	ErrorCount    int
	BackupPaths   map[string]string // source -> backup path
	DeployedAt    time.Time
}

// Deploy copies built configuration files to their actual paths.
func (s *DeployService) Deploy(opts DeployOptions) (*DeployResult, error) {
	var err error
	opts, err = defaultDeployOptions(opts)
	if err != nil {
		return nil, err
	}

	metadata, err := s.loadBuildMetadata(opts.BuildDir)
	if err != nil {
		return nil, err
	}

	result := &DeployResult{
		TotalFiles:  len(metadata.Files),
		BackupPaths: make(map[string]string),
		DeployedAt:  time.Now(),
	}

	for _, fileInfo := range metadata.Files {
		s.deployFile(fileInfo, opts, result)
	}

	return result, nil
}

func defaultDeployOptions(opts DeployOptions) (DeployOptions, error) {
	if opts.BuildDir == "" {
		opts.BuildDir = defaultBuildDir
	}
	if opts.HomeDir != "" {
		return opts, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return DeployOptions{}, fmt.Errorf("failed to get home directory: %w", err)
	}
	opts.HomeDir = home
	return opts, nil
}

func (s *DeployService) loadBuildMetadata(buildDir string) (*domain.BuildMetadata, error) {
	if !s.reader.FileExists(buildDir) {
		return nil, fmt.Errorf("build directory not found: %s\n\nRun 'gz-shellforge build' first to generate configuration files", buildDir)
	}

	metaPath := filepath.Join(buildDir, domain.MetadataFileName)
	if !s.reader.FileExists(metaPath) {
		return nil, fmt.Errorf("metadata file not found: %s\n\nRun 'gz-shellforge build' to regenerate", metaPath)
	}

	metaContent, err := s.reader.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	metadata, err := domain.ParseBuildMetadata([]byte(metaContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}
	if len(metadata.Files) == 0 {
		return nil, fmt.Errorf("no files found in build metadata\n\nRun 'gz-shellforge build' first to generate configuration files")
	}
	return metadata, nil
}

func (s *DeployService) deployFile(fileInfo domain.BuildFileInfo, opts DeployOptions, result *DeployResult) {
	sourcePath := filepath.Join(opts.BuildDir, fileInfo.Source)
	destPath := deployDestination(opts.HomeDir, fileInfo.DestPath)
	isSystem := filepath.IsAbs(fileInfo.DestPath)
	deployed := DeployedFile{SourcePath: sourcePath, DestPath: destPath}

	if !s.reader.FileExists(sourcePath) {
		deployed.Error = fmt.Errorf("source file not found: %s", sourcePath)
		s.recordFailure(result, deployed)
		return
	}
	if opts.DryRun {
		if isSystem {
			deployed.Error = fmt.Errorf("dry-run: %s requires elevated privileges — re-run with sudo", destPath)
		}
		deployed.Skipped = true
		result.SkippedCount++
		result.DeployedFiles = append(result.DeployedFiles, deployed)
		return
	}
	if isSystem {
		if err := s.checker.CheckWritable(destPath); err != nil {
			deployed.Error = err
			s.recordFailure(result, deployed)
			return
		}
	}
	if err := s.ensureDir(filepath.Dir(destPath)); err != nil {
		deployed.Error = fmt.Errorf("failed to create directory %s: %w", filepath.Dir(destPath), err)
		s.recordFailure(result, deployed)
		return
	}
	if err := s.backupDestination(sourcePath, destPath, opts.CreateBackup || isSystem, result, &deployed); err != nil {
		deployed.Error = fmt.Errorf("backup failed: %w", err)
		s.recordFailure(result, deployed)
		return
	}
	if err := s.writer.Copy(sourcePath, destPath); err != nil {
		deployed.Error = fmt.Errorf("copy failed: %w", err)
		s.recordFailure(result, deployed)
		return
	}
	deployed.Deployed = true
	result.DeployedCount++
	result.DeployedFiles = append(result.DeployedFiles, deployed)
}

func deployDestination(homeDir, destPath string) string {
	if filepath.IsAbs(destPath) {
		return destPath
	}
	return filepath.Join(homeDir, destPath)
}

func (s *DeployService) backupDestination(sourcePath, destPath string, needed bool, result *DeployResult, deployed *DeployedFile) error {
	if !needed || !s.reader.FileExists(destPath) {
		return nil
	}
	backupPath, err := s.createBackup(destPath)
	if err != nil {
		return err
	}
	deployed.BackupPath = backupPath
	result.BackupPaths[sourcePath] = backupPath
	return nil
}

func (s *DeployService) recordFailure(result *DeployResult, deployed DeployedFile) {
	result.ErrorCount++
	result.DeployedFiles = append(result.DeployedFiles, deployed)
}

// ensureDir creates a directory if it doesn't exist.
func (s *DeployService) ensureDir(dir string) error {
	return s.writer.MkdirAll(dir)
}

// createBackup creates a timestamped backup of a file.
func (s *DeployService) createBackup(path string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.backup.%s", path, timestamp)

	if err := s.writer.Copy(path, backupPath); err != nil {
		return "", err
	}

	return backupPath, nil
}
