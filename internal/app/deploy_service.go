package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DirectoryReader extends FileReader with directory listing.
type DirectoryReader interface {
	FileReader
	ListDir(path string) ([]string, error)
}

// BackupWriter extends FileWriter with backup support.
type BackupWriter interface {
	FileWriter
	Copy(src, dst string) error
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

	// List files in build directory
	files, err := s.reader.ListDir(opts.BuildDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list build directory: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in build directory: %s\n\nRun 'gz-shellforge build' first to generate configuration files", opts.BuildDir)
	}

	result := &DeployResult{
		TotalFiles:  len(files),
		BackupPaths: make(map[string]string),
		DeployedAt:  time.Now(),
	}

	// Process each file
	for _, file := range files {
		sourcePath := filepath.Join(opts.BuildDir, file)
		destPath := s.resolveDestPath(file, opts.HomeDir)

		deployed := DeployedFile{
			SourcePath: sourcePath,
			DestPath:   destPath,
		}

		// Skip if dry-run
		if opts.DryRun {
			deployed.Skipped = true
			result.SkippedCount++
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

// resolveDestPath converts a build file name to its actual destination path.
// For example: ".zshrc" -> "~/.zshrc"
func (s *DeployService) resolveDestPath(filename string, homeDir string) string {
	// RC files typically go to home directory
	if strings.HasPrefix(filename, ".") {
		return filepath.Join(homeDir, filename)
	}

	// Handle non-dotfiles (e.g., "zshrc" -> "~/.zshrc")
	knownRCFiles := map[string]string{
		"zshrc":        ".zshrc",
		"zprofile":     ".zprofile",
		"zshenv":       ".zshenv",
		"bashrc":       ".bashrc",
		"profile":      ".profile",
		"bash_profile": ".bash_profile",
	}

	if dotName, ok := knownRCFiles[filename]; ok {
		return filepath.Join(homeDir, dotName)
	}

	// Default: prefix with dot and place in home
	return filepath.Join(homeDir, "."+filename)
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
