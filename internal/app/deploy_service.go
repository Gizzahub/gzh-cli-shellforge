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

// DeployService implements the deploy use case.
type DeployService struct {
	reader DirectoryReader
	writer BackupWriter
}

// NewDeployService creates a new deploy service.
func NewDeployService(reader DirectoryReader, writer BackupWriter) *DeployService {
	return &DeployService{
		reader: reader,
		writer: writer,
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
	// Default values
	if opts.BuildDir == "" {
		opts.BuildDir = "./build"
	}
	if opts.HomeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		opts.HomeDir = home
	}

	// Check build directory exists
	if !s.reader.FileExists(opts.BuildDir) {
		return nil, fmt.Errorf("build directory not found: %s\n\nRun 'gz-shellforge build' first to generate configuration files", opts.BuildDir)
	}

	// Read metadata file
	metaPath := filepath.Join(opts.BuildDir, domain.MetadataFileName)
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

	result := &DeployResult{
		TotalFiles:  len(metadata.Files),
		BackupPaths: make(map[string]string),
		DeployedAt:  time.Now(),
	}

	// Process each file from metadata
	for _, fileInfo := range metadata.Files {
		sourcePath := filepath.Join(opts.BuildDir, fileInfo.Source)
		destPath := filepath.Join(opts.HomeDir, fileInfo.DestPath)

		deployed := DeployedFile{
			SourcePath: sourcePath,
			DestPath:   destPath,
		}

		// Check source file exists
		if !s.reader.FileExists(sourcePath) {
			deployed.Error = fmt.Errorf("source file not found: %s", sourcePath)
			result.ErrorCount++
			result.DeployedFiles = append(result.DeployedFiles, deployed)
			continue
		}

		// Skip if dry-run
		if opts.DryRun {
			deployed.Skipped = true
			result.SkippedCount++
			result.DeployedFiles = append(result.DeployedFiles, deployed)
			continue
		}

		// Ensure destination directory exists (for nested paths like .config/fish/)
		destDir := filepath.Dir(destPath)
		if err := s.ensureDir(destDir); err != nil {
			deployed.Error = fmt.Errorf("failed to create directory %s: %w", destDir, err)
			result.ErrorCount++
			result.DeployedFiles = append(result.DeployedFiles, deployed)
			continue
		}

		// Create backup if requested and file exists
		if opts.CreateBackup && s.reader.FileExists(destPath) {
			backupPath, err := s.createBackup(destPath)
			if err != nil {
				deployed.Error = fmt.Errorf("backup failed: %w", err)
				result.ErrorCount++
				result.DeployedFiles = append(result.DeployedFiles, deployed)
				continue
			}
			deployed.BackupPath = backupPath
			result.BackupPaths[sourcePath] = backupPath
		}

		// Copy file to destination
		if err := s.writer.Copy(sourcePath, destPath); err != nil {
			deployed.Error = fmt.Errorf("copy failed: %w", err)
			result.ErrorCount++
			result.DeployedFiles = append(result.DeployedFiles, deployed)
			continue
		}

		deployed.Deployed = true
		result.DeployedCount++
		result.DeployedFiles = append(result.DeployedFiles, deployed)
	}

	return result, nil
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
