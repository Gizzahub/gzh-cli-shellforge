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
	ConfigDir string // Directory containing module files
	Manifest  string // Path to manifest.yaml
	OS        string // Target OS (Mac, Linux, etc.)
	DryRun    bool   // If true, don't write output file
	Verbose   bool   // Show detailed output

	// Multi-target options (v2)
	OutputDir    string   // Output directory for multi-target builds
	Shell        string   // Shell type override (zsh, bash, fish)
	Targets      []string // Specific targets to build (empty = all)
	CreateBackup bool     // Create backup of existing files
	HomeDir      string   // Home directory for path resolution

	// Legacy single-output mode
	Output string // Single output file path (legacy mode)
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
	// Multi-target results
	Targets []TargetResult

	// Summary
	TotalModuleCount int
	GeneratedAt      time.Time
	ShellType        string
	TargetOS         string

	// Legacy compatibility
	Output      string   // Combined output (for legacy mode or dry-run)
	ModuleCount int      // Total module count (legacy alias)
	ModuleNames []string // All module names (legacy alias)
}

// Build generates shell configuration from modules.
// Supports both legacy single-output mode and multi-target mode.
func (s *BuilderService) Build(opts BuildOptions) (*BuildResult, error) {
	// 1. Parse manifest
	manifest, err := s.manifestParser.Parse(opts.Manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// 2. Determine shell type and build mode
	shellType := s.determineShellType(opts, manifest)
	isLegacyMode := s.isLegacyMode(opts, manifest)

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

	// 4. Route to appropriate build mode
	if isLegacyMode {
		return s.buildLegacy(opts, modules, shellType, now)
	}
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

// isLegacyMode determines if we should use legacy single-output mode.
func (s *BuilderService) isLegacyMode(opts BuildOptions, manifest *domain.Manifest) bool {
	// Explicit single output = legacy mode
	if opts.Output != "" {
		return true
	}
	// No output dir and legacy manifest = legacy mode
	if opts.OutputDir == "" && manifest.IsLegacy() {
		return true
	}
	return false
}

// buildLegacy generates a single output file (legacy mode).
func (s *BuilderService) buildLegacy(opts BuildOptions, modules []domain.Module, shellType string, now time.Time) (*BuildResult, error) {
	content, moduleNames := s.generateContent(modules, opts, shellType, "", now)

	// Write output (unless dry-run)
	if !opts.DryRun && opts.Output != "" {
		if err := s.fileWriter.WriteFile(opts.Output, content); err != nil {
			return nil, fmt.Errorf("failed to write output: %w", err)
		}
	}

	return &BuildResult{
		Targets: []TargetResult{{
			Target:      s.getDefaultTarget(shellType),
			FilePath:    opts.Output,
			Content:     content,
			ModuleCount: len(modules),
			ModuleNames: moduleNames,
		}},
		TotalModuleCount: len(modules),
		GeneratedAt:      now,
		ShellType:        shellType,
		TargetOS:         opts.OS,
		// Legacy fields
		Output:      content,
		ModuleCount: len(modules),
		ModuleNames: moduleNames,
	}, nil
}

// buildMultiTarget generates multiple RC files based on module targets.
func (s *BuilderService) buildMultiTarget(opts BuildOptions, manifest *domain.Manifest, modules []domain.Module, shellType string, now time.Time) (*BuildResult, error) {
	// Determine output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = manifest.GetOutputDirectory()
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
	var allModuleNames []string
	var combinedContent []string
	totalModuleCount := 0

	// Process targets in deterministic order
	targetNames := make([]string, 0, len(targetGroups))
	for target := range targetGroups {
		targetNames = append(targetNames, target)
	}
	sort.Strings(targetNames)

	for _, target := range targetNames {
		mods := targetGroups[target]
		if len(mods) == 0 {
			continue
		}

		filePath, err := resolver.Resolve(target)
		if err != nil {
			return nil, err
		}

		content, moduleNames := s.generateContent(mods, opts, shellType, target, now)
		allModuleNames = append(allModuleNames, moduleNames...)
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
			// Create backup if requested
			if (opts.CreateBackup || manifest.Output.Backup) && s.backupCreator != nil {
				if s.fileReader.FileExists(filePath) {
					backupPath, err := s.backupCreator.CreateBackup(filePath)
					if err != nil {
						return nil, fmt.Errorf("failed to backup %s: %w", filePath, err)
					}
					result.BackupPath = backupPath
				}
			}

			if err := s.fileWriter.WriteFile(filePath, content); err != nil {
				return nil, fmt.Errorf("failed to write %s: %w", filePath, err)
			}
		}

		results = append(results, result)
		combinedContent = append(combinedContent, fmt.Sprintf("# === %s ===\n%s", target, content))
	}

	return &BuildResult{
		Targets:          results,
		TotalModuleCount: totalModuleCount,
		GeneratedAt:      now,
		ShellType:        shellType,
		TargetOS:         opts.OS,
		// Legacy fields for compatibility
		Output:      strings.Join(combinedContent, "\n\n"),
		ModuleCount: totalModuleCount,
		ModuleNames: allModuleNames,
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
