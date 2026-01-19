package app

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// ManifestParser defines the interface for parsing YAML manifests.
type ManifestParser interface {
	Parse(path string) (*domain.Manifest, error)
}

// FileReader defines the interface for reading files.
type FileReader interface {
	ReadFile(path string) (string, error)
	FileExists(path string) bool
}

// FileWriter defines the interface for writing files.
type FileWriter interface {
	WriteFile(path string, content string) error
}

// BackupCreator defines the interface for creating backups.
type BackupCreator interface {
	CreateBackup(path string) (string, error)
}

// BuilderService implements the build use case.
type BuilderService struct {
	manifestParser ManifestParser
	fileReader     FileReader
	fileWriter     FileWriter
	backupCreator  BackupCreator
	resolver       *domain.Resolver
}

// NewBuilderService creates a new builder service.
func NewBuilderService(parser ManifestParser, reader FileReader, writer FileWriter) *BuilderService {
	return &BuilderService{
		manifestParser: parser,
		fileReader:     reader,
		fileWriter:     writer,
		resolver:       domain.NewResolver(),
	}
}

// SetBackupCreator sets the backup creator for the builder service.
func (s *BuilderService) SetBackupCreator(bc BackupCreator) {
	s.backupCreator = bc
}

// BuildOptions contains options for building shell configuration.
type BuildOptions struct {
	ConfigDir string   // Directory containing module files
	Manifest  string   // Path to manifest.yaml
	OS        string   // Target OS (Mac, Linux, etc.)
	DryRun    bool     // If true, don't write output file
	Verbose   bool     // Show detailed output
	OutputDir string   // Output directory for builds (default: ./build)
	Shell     string   // Shell type override (zsh, bash, fish)
	Targets   []string // Specific targets to build (empty = all)
	HomeDir   string   // Home directory for path resolution
}

// TargetResult contains the result for a single target file.
type TargetResult struct {
	Target      string   // Target name (e.g., "zshrc")
	FilePath    string   // Full file path
	Content     string   // Generated content
	ModuleCount int      // Number of modules in this target
	ModuleNames []string // Module names in order
	BackupPath  string   // Path to backup file (if created)
}

// BuildResult contains the result of a build operation.
type BuildResult struct {
	Targets          []TargetResult
	TotalModuleCount int
	GeneratedAt      time.Time
	ShellType        string
	TargetOS         string
}

// Build generates shell configuration from modules.
func (s *BuilderService) Build(opts BuildOptions) (*BuildResult, error) {
	// 1. Parse manifest
	manifest, err := s.manifestParser.Parse(opts.Manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// 2. Determine shell type
	shellType := s.determineShellType(opts, manifest)

	// 3. Build dependency graph and resolve
	graph, err := s.resolver.BuildGraph(manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}

	modules, err := s.resolver.TopologicalSort(graph, opts.OS)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	now := time.Now()

	// 4. Build multi-target output
	return s.buildMultiTarget(opts, manifest, modules, shellType, now)
}

// determineShellType returns the shell type to use.
func (s *BuilderService) determineShellType(opts BuildOptions, manifest *domain.Manifest) string {
	// Priority: CLI option > manifest > default
	if opts.Shell != "" {
		return strings.ToLower(opts.Shell)
	}
	if manifest.Shell.Type != "" {
		return strings.ToLower(manifest.Shell.Type)
	}
	return "zsh"
}

// buildMultiTarget generates multiple RC files based on module targets.
func (s *BuilderService) buildMultiTarget(opts BuildOptions, manifest *domain.Manifest, modules []domain.Module, shellType string, now time.Time) (*BuildResult, error) {
	// Determine output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = manifest.GetOutputDirectory()
	}
	if outputDir == "" {
		outputDir = "./build"
	}

	// Expand home directory
	homeDir := opts.HomeDir
	if homeDir == "" {
		homeDir = "~"
	}
	if strings.HasPrefix(outputDir, "~") {
		outputDir = strings.Replace(outputDir, "~", homeDir, 1)
	}

	// Create target resolver
	resolver := domain.NewTargetResolver(shellType, outputDir)

	// Validate targets
	if err := resolver.ValidateTargets(modules); err != nil {
		return nil, err
	}

	// Group modules by target
	targetGroups := s.groupModulesByTarget(modules)

	// Filter targets if specific ones requested
	if len(opts.Targets) > 0 {
		targetGroups = s.filterTargets(targetGroups, opts.Targets)
	}

	// Sort modules within each target by priority
	for target := range targetGroups {
		s.sortByPriority(targetGroups[target])
	}

	// Generate content for each target
	var results []TargetResult
	totalModuleCount := 0

	// Process targets in deterministic order
	targetNames := make([]string, 0, len(targetGroups))
	for target := range targetGroups {
		targetNames = append(targetNames, target)
	}
	sort.Strings(targetNames)

	// Collect metadata for deploy
	var metaFiles []domain.BuildFileInfo

	for _, target := range targetNames {
		mods := targetGroups[target]
		if len(mods) == 0 {
			continue
		}

		// Check if this is a directory target (e.g., conf.d)
		if resolver.IsDirectoryTarget(target) {
			// Handle directory target: one file per module
			dirResults, dirMetaFiles, err := s.buildDirectoryTarget(opts, mods, resolver, target, shellType, now)
			if err != nil {
				return nil, err
			}
			results = append(results, dirResults...)
			metaFiles = append(metaFiles, dirMetaFiles...)
			totalModuleCount += len(mods)
			continue
		}

		filePath, err := resolver.Resolve(target)
		if err != nil {
			return nil, err
		}

		// Get relative path for deploy metadata
		destPath, err := resolver.GetRelativePath(target)
		if err != nil {
			return nil, err
		}

		content, moduleNames := s.generateContent(mods, opts, shellType, target, now)
		totalModuleCount += len(mods)

		result := TargetResult{
			Target:      target,
			FilePath:    filePath,
			Content:     content,
			ModuleCount: len(mods),
			ModuleNames: moduleNames,
		}

		// Write file (unless dry-run)
		if !opts.DryRun {
			if err := s.fileWriter.WriteFile(filePath, content); err != nil {
				return nil, fmt.Errorf("failed to write %s: %w", filePath, err)
			}
		}

		results = append(results, result)

		// Add to metadata
		metaFiles = append(metaFiles, domain.BuildFileInfo{
			Source:   filepath.Base(filePath),
			Target:   target,
			DestPath: destPath,
		})
	}

	// Write metadata file (unless dry-run)
	if !opts.DryRun {
		metadata := &domain.BuildMetadata{
			Shell:       shellType,
			OS:          opts.OS,
			GeneratedAt: now,
			Files:       metaFiles,
		}
		metaJSON, err := metadata.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize metadata: %w", err)
		}
		metaPath := filepath.Join(outputDir, domain.MetadataFileName)
		if err := s.fileWriter.WriteFile(metaPath, string(metaJSON)); err != nil {
			return nil, fmt.Errorf("failed to write metadata: %w", err)
		}
	}

	return &BuildResult{
		Targets:          results,
		TotalModuleCount: totalModuleCount,
		GeneratedAt:      now,
		ShellType:        shellType,
		TargetOS:         opts.OS,
	}, nil
}

// groupModulesByTarget groups modules by their target RC file.
func (s *BuilderService) groupModulesByTarget(modules []domain.Module) map[string][]domain.Module {
	groups := make(map[string][]domain.Module)
	for _, mod := range modules {
		target := mod.GetTarget()
		groups[target] = append(groups[target], mod)
	}
	return groups
}

// filterTargets filters target groups to only include specified targets.
func (s *BuilderService) filterTargets(groups map[string][]domain.Module, targets []string) map[string][]domain.Module {
	filtered := make(map[string][]domain.Module)
	for _, target := range targets {
		target = strings.ToLower(target)
		if mods, ok := groups[target]; ok {
			filtered[target] = mods
		}
	}
	return filtered
}

// sortByPriority sorts modules by priority (lower = earlier).
func (s *BuilderService) sortByPriority(modules []domain.Module) {
	sort.SliceStable(modules, func(i, j int) bool {
		return modules[i].GetPriority() < modules[j].GetPriority()
	})
}

// generateContent generates the shell configuration content for a list of modules.
func (s *BuilderService) generateContent(modules []domain.Module, opts BuildOptions, shellType, target string, now time.Time) (string, []string) {
	var lines []string

	// Header
	lines = append(lines, "# Generated by shellforge")
	lines = append(lines, fmt.Sprintf("# Shell: %s", shellType))
	if target != "" {
		lines = append(lines, fmt.Sprintf("# Target: %s", target))
	}
	lines = append(lines, fmt.Sprintf("# OS: %s", opts.OS))
	lines = append(lines, fmt.Sprintf("# Modules: %d", len(modules)))
	lines = append(lines, fmt.Sprintf("# Generated at: %s", now.Format(time.RFC3339)))
	lines = append(lines, "")

	moduleNames := make([]string, 0, len(modules))

	for _, module := range modules {
		moduleNames = append(moduleNames, module.Name)

		// Construct full path
		filePath := filepath.Join(opts.ConfigDir, module.File)

		// Check if file exists
		if !s.fileReader.FileExists(filePath) {
			// Skip missing files in content generation (error will be caught elsewhere)
			lines = append(lines, fmt.Sprintf("\n# --- %s --- (FILE NOT FOUND: %s)", module.Name, filePath))
			continue
		}

		// Read module content
		content, err := s.fileReader.ReadFile(filePath)
		if err != nil {
			lines = append(lines, fmt.Sprintf("\n# --- %s --- (READ ERROR: %v)", module.Name, err))
			continue
		}

		// Add module header
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("# --- %s ---", module.Name))
		if module.Description != "" {
			lines = append(lines, fmt.Sprintf("# %s", module.Description))
		}
		if module.Priority != 0 {
			lines = append(lines, fmt.Sprintf("# Priority: %d", module.Priority))
		}

		// Add module content (trim trailing whitespace)
		lines = append(lines, strings.TrimRight(content, " \t\n"))
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n"), moduleNames
}

// getDefaultTarget returns the default target for a shell type.
func (s *BuilderService) getDefaultTarget(shellType string) string {
	switch shellType {
	case "zsh":
		return "zshrc"
	case "bash":
		return "bashrc"
	case "fish":
		return "config"
	default:
		return "shellrc"
	}
}

// buildDirectoryTarget handles directory targets like conf.d where each module
// gets its own file instead of being merged into a single file.
func (s *BuilderService) buildDirectoryTarget(opts BuildOptions, mods []domain.Module, resolver *domain.TargetResolver, target, shellType string, now time.Time) ([]TargetResult, []domain.BuildFileInfo, error) {
	// Get the directory path
	dirPath, err := resolver.Resolve(target)
	if err != nil {
		return nil, nil, err
	}

	// Get relative path for deploy metadata (e.g., ".config/fish/conf.d")
	relDirPath, err := resolver.GetRelativePath(target)
	if err != nil {
		return nil, nil, err
	}

	var results []TargetResult
	var metaFiles []domain.BuildFileInfo

	for _, mod := range mods {
		// Generate filename from module name: {module-name}.fish
		fileName := sanitizeModuleName(mod.Name) + ".fish"
		filePath := filepath.Join(dirPath, fileName)

		// Generate content for single module
		content := s.generateSingleModuleContent(mod, opts, shellType, target, now)

		result := TargetResult{
			Target:      target,
			FilePath:    filePath,
			Content:     content,
			ModuleCount: 1,
			ModuleNames: []string{mod.Name},
		}

		// Write file (unless dry-run)
		if !opts.DryRun {
			if err := s.fileWriter.WriteFile(filePath, content); err != nil {
				return nil, nil, fmt.Errorf("failed to write %s: %w", filePath, err)
			}
		}

		results = append(results, result)

		// Add to metadata with full destination path
		metaFiles = append(metaFiles, domain.BuildFileInfo{
			Source:   filepath.Join("conf.d", fileName),
			Target:   target,
			DestPath: filepath.Join(relDirPath, fileName),
		})
	}

	return results, metaFiles, nil
}

// generateSingleModuleContent generates shell configuration content for a single module.
// Used for directory targets where each module gets its own file.
func (s *BuilderService) generateSingleModuleContent(mod domain.Module, opts BuildOptions, shellType, target string, now time.Time) string {
	var lines []string

	// Header
	lines = append(lines, "# Generated by shellforge")
	lines = append(lines, fmt.Sprintf("# Shell: %s", shellType))
	lines = append(lines, fmt.Sprintf("# Module: %s", mod.Name))
	if target != "" {
		lines = append(lines, fmt.Sprintf("# Target: %s", target))
	}
	lines = append(lines, fmt.Sprintf("# OS: %s", opts.OS))
	lines = append(lines, fmt.Sprintf("# Generated at: %s", now.Format(time.RFC3339)))
	lines = append(lines, "")

	// Construct full path
	filePath := filepath.Join(opts.ConfigDir, mod.File)

	// Check if file exists
	if !s.fileReader.FileExists(filePath) {
		lines = append(lines, fmt.Sprintf("# --- %s --- (FILE NOT FOUND: %s)", mod.Name, filePath))
		return strings.Join(lines, "\n")
	}

	// Read module content
	content, err := s.fileReader.ReadFile(filePath)
	if err != nil {
		lines = append(lines, fmt.Sprintf("# --- %s --- (READ ERROR: %v)", mod.Name, err))
		return strings.Join(lines, "\n")
	}

	// Add module description as comment
	if mod.Description != "" {
		lines = append(lines, fmt.Sprintf("# %s", mod.Description))
	}
	if mod.Priority != 0 {
		lines = append(lines, fmt.Sprintf("# Priority: %d", mod.Priority))
	}
	lines = append(lines, "")

	// Add module content (trim trailing whitespace)
	lines = append(lines, strings.TrimRight(content, " \t\n"))
	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

// sanitizeModuleName converts a module name to a safe filename.
// Replaces spaces and special characters with underscores.
func sanitizeModuleName(name string) string {
	// Replace common unsafe characters
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ToLower(name)
	return name
}
