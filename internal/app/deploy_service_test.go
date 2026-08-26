package app

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// MockDirectoryReader implements DirectoryReader for testing.
type MockDirectoryReader struct {
	files       map[string]string
	directories map[string][]string
}

func NewMockDirectoryReader() *MockDirectoryReader {
	return &MockDirectoryReader{
		files:       make(map[string]string),
		directories: make(map[string][]string),
	}
}

func (m *MockDirectoryReader) ReadFile(path string) (string, error) {
	if content, ok := m.files[path]; ok {
		return content, nil
	}
	return "", nil
}

func (m *MockDirectoryReader) FileExists(path string) bool {
	_, ok := m.files[path]
	if ok {
		return true
	}
	_, ok = m.directories[path]
	return ok
}

func (m *MockDirectoryReader) ListDir(path string) ([]string, error) {
	if files, ok := m.directories[path]; ok {
		return files, nil
	}
	return []string{}, nil
}

func (m *MockDirectoryReader) AddFile(path, content string) {
	m.files[path] = content
}

func (m *MockDirectoryReader) AddDirectory(path string, files []string) {
	m.directories[path] = files
	for _, f := range files {
		m.files[filepath.Join(path, f)] = "content of " + f
	}
}

// MockBackupWriter implements BackupWriter for testing.
type MockBackupWriter struct {
	files map[string]string
}

func NewMockBackupWriter() *MockBackupWriter {
	return &MockBackupWriter{
		files: make(map[string]string),
	}
}

func (m *MockBackupWriter) WriteFile(path, content string) error {
	m.files[path] = content
	return nil
}

func (m *MockBackupWriter) Copy(src, dst string) error {
	m.files[dst] = "copied from " + src
	return nil
}

func (m *MockBackupWriter) MkdirAll(path string) error {
	// Mock implementation - just track that the directory was created
	return nil
}

func (m *MockBackupWriter) GetFile(path string) (string, bool) {
	content, ok := m.files[path]
	return content, ok
}

// MockPermissionChecker implements PermissionChecker for testing.
type MockPermissionChecker struct {
	err error // nil = writable, non-nil = denied
}

func (m *MockPermissionChecker) CheckWritable(_ string) error {
	return m.err
}

// createTestMetadata creates a JSON metadata file content for testing.
func createTestMetadata(files []domain.BuildFileInfo) string {
	meta := &domain.BuildMetadata{
		Shell:       "zsh",
		OS:          "Mac",
		GeneratedAt: time.Now(),
		Files:       files,
	}
	data, _ := meta.ToJSON()
	return string(data)
}

func TestDeployService_Deploy_Success(t *testing.T) {
	reader := NewMockDirectoryReader()
	writer := NewMockBackupWriter()
	service := NewDeployService(reader, writer)

	// Setup: build directory with files and metadata
	reader.AddDirectory("./build", []string{".zshrc", ".zprofile"})
	reader.AddFile("build/.zshrc", "zshrc content")
	reader.AddFile("build/.zprofile", "zprofile content")
	reader.AddFile("build/"+domain.MetadataFileName, createTestMetadata([]domain.BuildFileInfo{
		{Source: ".zshrc", Target: "zshrc", DestPath: ".zshrc"},
		{Source: ".zprofile", Target: "zprofile", DestPath: ".zprofile"},
	}))

	opts := DeployOptions{
		BuildDir: "./build",
		HomeDir:  "/home/test",
	}

	result, err := service.Deploy(opts)
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	if result.TotalFiles != 2 {
		t.Errorf("TotalFiles = %d, want 2", result.TotalFiles)
	}

	if result.DeployedCount != 2 {
		t.Errorf("DeployedCount = %d, want 2", result.DeployedCount)
	}

	// Verify files were written
	if _, ok := writer.GetFile("/home/test/.zshrc"); !ok {
		t.Error("Expected .zshrc to be deployed")
	}
	if _, ok := writer.GetFile("/home/test/.zprofile"); !ok {
		t.Error("Expected .zprofile to be deployed")
	}
}

func TestDeployService_Deploy_FishNestedPath(t *testing.T) {
	reader := NewMockDirectoryReader()
	writer := NewMockBackupWriter()
	service := NewDeployService(reader, writer)

	// Setup: fish shell with nested config path
	reader.AddDirectory("./build", []string{"config.fish"})
	reader.AddFile("build/config.fish", "fish config content")
	reader.AddFile("build/"+domain.MetadataFileName, createTestMetadata([]domain.BuildFileInfo{
		{Source: "config.fish", Target: "config", DestPath: ".config/fish/config.fish"},
	}))

	opts := DeployOptions{
		BuildDir: "./build",
		HomeDir:  "/home/test",
	}

	result, err := service.Deploy(opts)
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	if result.DeployedCount != 1 {
		t.Errorf("DeployedCount = %d, want 1", result.DeployedCount)
	}

	// Verify file was written to nested path
	if _, ok := writer.GetFile("/home/test/.config/fish/config.fish"); !ok {
		t.Error("Expected config.fish to be deployed to ~/.config/fish/config.fish")
	}
}

func TestDeployService_Deploy_DryRun(t *testing.T) {
	reader := NewMockDirectoryReader()
	writer := NewMockBackupWriter()
	service := NewDeployService(reader, writer)

	reader.AddDirectory("./build", []string{".zshrc"})
	reader.AddFile("build/.zshrc", "zshrc content")
	reader.AddFile("build/"+domain.MetadataFileName, createTestMetadata([]domain.BuildFileInfo{
		{Source: ".zshrc", Target: "zshrc", DestPath: ".zshrc"},
	}))

	opts := DeployOptions{
		BuildDir: "./build",
		HomeDir:  "/home/test",
		DryRun:   true,
	}

	result, err := service.Deploy(opts)
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	if result.DeployedCount != 0 {
		t.Errorf("DeployedCount = %d, want 0 in dry-run", result.DeployedCount)
	}

	if result.SkippedCount != 1 {
		t.Errorf("SkippedCount = %d, want 1 in dry-run", result.SkippedCount)
	}

	// Verify no files were written
	if _, ok := writer.GetFile("/home/test/.zshrc"); ok {
		t.Error("File should not be deployed in dry-run mode")
	}
}

func TestDeployService_Deploy_WithBackup(t *testing.T) {
	reader := NewMockDirectoryReader()
	writer := NewMockBackupWriter()
	service := NewDeployService(reader, writer)

	// Setup: build directory and existing destination file
	reader.AddDirectory("./build", []string{".zshrc"})
	reader.AddFile("build/.zshrc", "new zshrc content")
	reader.AddFile("build/"+domain.MetadataFileName, createTestMetadata([]domain.BuildFileInfo{
		{Source: ".zshrc", Target: "zshrc", DestPath: ".zshrc"},
	}))
	reader.AddFile("/home/test/.zshrc", "existing content")

	opts := DeployOptions{
		BuildDir:     "./build",
		HomeDir:      "/home/test",
		CreateBackup: true,
	}

	result, err := service.Deploy(opts)
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	if len(result.BackupPaths) != 1 {
		t.Errorf("BackupPaths count = %d, want 1", len(result.BackupPaths))
	}

	// Verify backup was created
	if result.DeployedFiles[0].BackupPath == "" {
		t.Error("Expected backup path to be set")
	}
}

func TestDeployService_Deploy_BuildDirNotFound(t *testing.T) {
	reader := NewMockDirectoryReader()
	writer := NewMockBackupWriter()
	service := NewDeployService(reader, writer)

	opts := DeployOptions{
		BuildDir: "./nonexistent",
		HomeDir:  "/home/test",
	}

	_, err := service.Deploy(opts)
	if err == nil {
		t.Error("Deploy() expected error for missing build directory")
	}
}

func TestDeployService_Deploy_MetadataNotFound(t *testing.T) {
	reader := NewMockDirectoryReader()
	writer := NewMockBackupWriter()
	service := NewDeployService(reader, writer)

	// Build directory exists but no metadata
	reader.AddDirectory("./build", []string{".zshrc"})

	opts := DeployOptions{
		BuildDir: "./build",
		HomeDir:  "/home/test",
	}

	_, err := service.Deploy(opts)
	if err == nil {
		t.Error("Deploy() expected error for missing metadata file")
	}
}

func TestDeployService_Deploy_EmptyMetadata(t *testing.T) {
	reader := NewMockDirectoryReader()
	writer := NewMockBackupWriter()
	service := NewDeployService(reader, writer)

	reader.AddDirectory("./build", []string{})
	reader.AddFile("build/"+domain.MetadataFileName, createTestMetadata([]domain.BuildFileInfo{}))

	opts := DeployOptions{
		BuildDir: "./build",
		HomeDir:  "/home/test",
	}

	_, err := service.Deploy(opts)
	if err == nil {
		t.Error("Deploy() expected error for empty metadata")
	}
}

// --- System target tests ---

func TestDeployService_Deploy_SystemTarget_UsesAbsolutePath(t *testing.T) {
	reader := NewMockDirectoryReader()
	writer := NewMockBackupWriter()
	checker := &MockPermissionChecker{err: nil} // writable
	service := NewDeployServiceWithChecker(reader, writer, checker)

	reader.AddDirectory("./build", []string{"etc-profile"})
	reader.AddFile("build/etc-profile", "# system profile content")
	reader.AddFile("build/"+domain.MetadataFileName, createTestMetadata([]domain.BuildFileInfo{
		{Source: "etc-profile", Target: "etc-profile", DestPath: "/etc/profile"},
	}))

	opts := DeployOptions{
		BuildDir: "./build",
		HomeDir:  "/home/test",
	}

	result, err := service.Deploy(opts)
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	if result.DeployedCount != 1 {
		t.Errorf("DeployedCount = %d, want 1", result.DeployedCount)
	}

	// Must be written to the absolute path, NOT /home/test/etc/profile
	if _, ok := writer.GetFile("/etc/profile"); !ok {
		t.Error("system target must deploy to /etc/profile, not a home-relative path")
	}
	if _, ok := writer.GetFile("/home/test/etc/profile"); ok {
		t.Error("system target must NOT be joined with HomeDir")
	}
}

func TestDeployService_Deploy_SystemTarget_PermissionDenied(t *testing.T) {
	reader := NewMockDirectoryReader()
	writer := NewMockBackupWriter()
	permErr := fmt.Errorf("requires elevated privileges to write to /etc/profile — re-run with sudo")
	checker := &MockPermissionChecker{err: permErr}
	service := NewDeployServiceWithChecker(reader, writer, checker)

	reader.AddDirectory("./build", []string{"etc-profile"})
	reader.AddFile("build/etc-profile", "# content")
	reader.AddFile("build/"+domain.MetadataFileName, createTestMetadata([]domain.BuildFileInfo{
		{Source: "etc-profile", Target: "etc-profile", DestPath: "/etc/profile"},
	}))

	opts := DeployOptions{
		BuildDir: "./build",
		HomeDir:  "/home/test",
	}

	result, err := service.Deploy(opts)
	if err != nil {
		t.Fatalf("Deploy() unexpected fatal error = %v", err)
	}

	if result.ErrorCount != 1 {
		t.Errorf("ErrorCount = %d, want 1", result.ErrorCount)
	}
	if result.DeployedFiles[0].Error == nil {
		t.Error("expected permission error on system target")
	}
	// Verify no file was written
	if _, ok := writer.GetFile("/etc/profile"); ok {
		t.Error("must not write to system file when permission check fails")
	}
}

func TestDeployService_Deploy_SystemTarget_BackupForcedWithoutFlag(t *testing.T) {
	reader := NewMockDirectoryReader()
	writer := NewMockBackupWriter()
	checker := &MockPermissionChecker{err: nil}
	service := NewDeployServiceWithChecker(reader, writer, checker)

	reader.AddDirectory("./build", []string{"etc-profile"})
	reader.AddFile("build/etc-profile", "# new content")
	reader.AddFile("build/"+domain.MetadataFileName, createTestMetadata([]domain.BuildFileInfo{
		{Source: "etc-profile", Target: "etc-profile", DestPath: "/etc/profile"},
	}))
	// Existing system file
	reader.AddFile("/etc/profile", "# original system content")

	opts := DeployOptions{
		BuildDir:     "./build",
		HomeDir:      "/home/test",
		CreateBackup: false, // not requested by user, but must still happen for system files
	}

	result, err := service.Deploy(opts)
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	if result.DeployedCount != 1 {
		t.Errorf("DeployedCount = %d, want 1", result.DeployedCount)
	}

	// Backup must be created even though CreateBackup was false
	if result.DeployedFiles[0].BackupPath == "" {
		t.Error("backup must be forced for system targets even when CreateBackup=false")
	}
}

func TestDeployService_Deploy_SystemTarget_DryRun(t *testing.T) {
	reader := NewMockDirectoryReader()
	writer := NewMockBackupWriter()
	checker := &MockPermissionChecker{err: nil}
	service := NewDeployServiceWithChecker(reader, writer, checker)

	reader.AddDirectory("./build", []string{"etc-profile"})
	reader.AddFile("build/etc-profile", "# content")
	reader.AddFile("build/"+domain.MetadataFileName, createTestMetadata([]domain.BuildFileInfo{
		{Source: "etc-profile", Target: "etc-profile", DestPath: "/etc/profile"},
	}))

	opts := DeployOptions{
		BuildDir: "./build",
		HomeDir:  "/home/test",
		DryRun:   true,
	}

	result, err := service.Deploy(opts)
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	if result.SkippedCount != 1 {
		t.Errorf("SkippedCount = %d, want 1 in dry-run", result.SkippedCount)
	}

	// Dry-run for system target must carry a sudo hint in the error field
	df := result.DeployedFiles[0]
	if df.Error == nil {
		t.Error("dry-run system target must carry sudo hint in Error field")
	}
	// No actual file written
	if _, ok := writer.GetFile("/etc/profile"); ok {
		t.Error("dry-run must not write files")
	}
}

func TestDeployDestination_PreservesSystemPathsAndResolvesUserPaths(t *testing.T) {
	tests := []struct {
		name     string
		homeDir  string
		destPath string
		want     string
	}{
		{
			name:     "system path is not joined with home directory",
			homeDir:  "/home/test",
			destPath: "/etc/profile",
			want:     "/etc/profile",
		},
		{
			name:     "user path is joined with home directory",
			homeDir:  "/home/test",
			destPath: ".config/fish/config.fish",
			want:     "/home/test/.config/fish/config.fish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deployDestination(tt.homeDir, tt.destPath); got != tt.want {
				t.Errorf("deployDestination() = %q, want %q", got, tt.want)
			}
		})
	}
}
