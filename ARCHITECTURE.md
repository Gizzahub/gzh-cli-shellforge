# System Architecture: Shellforge Go Implementation

**Version**: 1.0
**Status**: Draft
**Last Updated**: 2025-11-27

---

## Overview

This document describes the high-level architecture of the Shellforge Go implementation, following Go best practices and standard project layout.

---

## Architecture Principles

### 1. Standard Go Project Layout

Follow [golang-standards/project-layout](https://github.com/golang-standards/project-layout):
- `cmd/`: Main applications
- `internal/`: Private application code
- `pkg/`: Public libraries (if needed)
- `data/`: Embedded data files

### 2. Clean Architecture Layers

Separate concerns into distinct layers:
- **Domain**: Pure business logic, no external dependencies
- **Application**: Use cases orchestrating domain logic
- **Infrastructure**: Adapters for external systems (filesystem, git, YAML)
- **Interface**: CLI commands and user interaction

### 3. Dependency Inversion

- Domain layer has no dependencies on outer layers
- Application layer depends on domain interfaces
- Infrastructure implements domain interfaces
- CLI depends on application use cases

### 4. Interface-Based Design

- Define interfaces for all external dependencies
- Use dependency injection for testability
- Mock infrastructure layer in tests

### 5. Minimal External Dependencies

- Prefer standard library
- Use well-established libraries (Cobra, yaml.v3)
- Implement simple algorithms (topological sort) instead of heavy libraries

---

## Project Structure

```
gzh-cli-shellforge/
├── cmd/
│   └── shellforge/
│       └── main.go                 # Application entry point
│
├── internal/                       # Private application code
│   ├── domain/                     # Business logic (pure Go)
│   │   ├── module.go               # Module entity
│   │   ├── manifest.go             # Manifest entity
│   │   ├── graph.go                # Dependency graph
│   │   ├── resolver.go             # Dependency resolver
│   │   ├── shellmeta.go            # Shell metadata types
│   │   └── errors.go               # Domain errors
│   │
│   ├── app/                        # Use cases
│   │   ├── builder.go              # Build shell config
│   │   ├── validator.go            # Validate manifest
│   │   ├── deployer.go             # Deploy with backup
│   │   ├── migrator.go             # Migrate RC files
│   │   ├── initializer.go          # Auto-generate manifest
│   │   ├── differ.go               # Compare configs
│   │   └── templates.go            # Template generator
│   │
│   ├── infra/                      # Infrastructure adapters
│   │   ├── yamlparser/             # YAML parsing
│   │   │   └── parser.go
│   │   ├── filesystem/             # File operations
│   │   │   ├── reader.go
│   │   │   └── writer.go
│   │   ├── git/                    # Git operations
│   │   │   └── backup.go
│   │   └── shelldata/              # Shell metadata loader
│   │       └── loader.go
│   │
│   └── cli/                        # CLI interface (Cobra)
│       ├── root.go                 # Root command
│       ├── build.go                # build command
│       ├── validate.go             # validate command
│       ├── init.go                 # init command
│       ├── migrate.go              # migrate command
│       ├── diff.go                 # diff command
│       ├── deploy.go               # deploy command (if separate)
│       ├── restore.go              # restore command
│       ├── clean.go                # clean-snapshots command
│       ├── list.go                 # list-modules command
│       ├── info.go                 # info command
│       ├── template.go             # template commands
│       └── helpers.go              # Shared CLI utilities
│
├── pkg/                            # Public libraries (if needed)
│   └── (none initially)
│
├── data/                           # Embedded data files
│   └── shell_configs.yaml          # Shell metadata
│
├── examples/
│   ├── manifest.yaml               # Example manifest
│   └── modules/                    # Example modules
│       ├── init.d/
│       ├── rc_pre.d/
│       └── rc_post.d/
│
├── docs/
│   ├── PRD.md
│   ├── REQUIREMENTS.md
│   ├── ARCHITECTURE.md             # This document
│   └── TECH_STACK.md
│
├── Makefile                        # Build automation
├── go.mod                          # Go module definition
├── go.sum                          # Dependency checksums
├── README.md                       # User documentation
└── .gitignore
```

---

## Domain Layer

### Purpose
Pure business logic with no external dependencies. Defines core types and interfaces.

### Components

#### 1. Module (`internal/domain/module.go`)

**Entity**: Represents a shell module

```go
package domain

type Module struct {
    Name        string
    File        string
    Requires    []string
    OS          []string
    Description string
}

func (m *Module) AppliesTo(targetOS string) bool {
    if len(m.OS) == 0 {
        return true // No OS restriction
    }
    for _, os := range m.OS {
        if strings.EqualFold(os, targetOS) {
            return true
        }
    }
    return false
}

func (m *Module) Validate() error {
    if m.Name == "" {
        return NewValidationError("module missing 'name' field")
    }
    if m.File == "" {
        return NewValidationError("module '%s' missing 'file' field", m.Name)
    }
    return nil
}
```

---

#### 2. Manifest (`internal/domain/manifest.go`)

**Entity**: Collection of modules

```go
package domain

type Manifest struct {
    Modules []Module
}

func (m *Manifest) FindModule(name string) (*Module, bool) {
    for i := range m.Modules {
        if m.Modules[i].Name == name {
            return &m.Modules[i], true
        }
    }
    return nil, false
}

func (m *Manifest) Validate() []error {
    var errors []error

    // Check for duplicate names
    seen := make(map[string]bool)
    for _, mod := range m.Modules {
        if seen[mod.Name] {
            errors = append(errors, NewValidationError("duplicate module name: %s", mod.Name))
        }
        seen[mod.Name] = true

        // Validate each module
        if err := mod.Validate(); err != nil {
            errors = append(errors, err)
        }
    }

    return errors
}
```

---

#### 3. Graph (`internal/domain/graph.go`)

**Data Structure**: Dependency graph for topological sort

```go
package domain

type Graph struct {
    nodes map[string]*Node
    edges map[string][]string // node -> list of dependents
}

type Node struct {
    Module   *Module
    InDegree int
}

func NewGraph() *Graph {
    return &Graph{
        nodes: make(map[string]*Node),
        edges: make(map[string][]string),
    }
}

func (g *Graph) AddNode(module *Module) {
    g.nodes[module.Name] = &Node{
        Module:   module,
        InDegree: 0,
    }
}

func (g *Graph) AddEdge(from, to string) error {
    if _, exists := g.nodes[from]; !exists {
        return NewValidationError("dependency '%s' not found", from)
    }
    if _, exists := g.nodes[to]; !exists {
        return NewValidationError("module '%s' not found", to)
    }

    g.edges[from] = append(g.edges[from], to)
    g.nodes[to].InDegree++
    return nil
}

func (g *Graph) Size() int {
    return len(g.nodes)
}
```

---

#### 4. Resolver (`internal/domain/resolver.go`)

**Algorithm**: Topological sort with cycle detection

```go
package domain

type Resolver struct{}

func NewResolver() *Resolver {
    return &Resolver{}
}

// BuildGraph creates dependency graph from manifest
func (r *Resolver) BuildGraph(manifest *Manifest) (*Graph, error) {
    graph := NewGraph()

    // Add all modules as nodes
    for i := range manifest.Modules {
        graph.AddNode(&manifest.Modules[i])
    }

    // Add edges for dependencies
    for _, module := range manifest.Modules {
        for _, dep := range module.Requires {
            if err := graph.AddEdge(dep, module.Name); err != nil {
                return nil, err
            }
        }
    }

    return graph, nil
}

// TopologicalSort performs Kahn's algorithm
func (r *Resolver) TopologicalSort(graph *Graph, targetOS string) ([]Module, error) {
    // Create working copy of in-degrees
    inDegree := make(map[string]int)
    for name, node := range graph.nodes {
        if node.Module.AppliesTo(targetOS) {
            inDegree[name] = node.InDegree
        }
    }

    // Find all nodes with in-degree 0
    queue := []string{}
    for name, degree := range inDegree {
        if degree == 0 {
            queue = append(queue, name)
        }
    }

    // Process queue
    var result []Module
    for len(queue) > 0 {
        // Dequeue
        current := queue[0]
        queue = queue[1:]

        node := graph.nodes[current]
        result = append(result, *node.Module)

        // Process dependents
        for _, dependent := range graph.edges[current] {
            if _, ok := inDegree[dependent]; !ok {
                continue // Skip filtered modules
            }

            inDegree[dependent]--
            if inDegree[dependent] == 0 {
                queue = append(queue, dependent)
            }
        }
    }

    // Check for cycles
    if len(result) != len(inDegree) {
        return nil, r.detectCycle(graph, inDegree)
    }

    return result, nil
}

func (r *Resolver) detectCycle(graph *Graph, inDegree map[string]int) error {
    // Find nodes still in graph (part of cycle)
    var cycleNodes []string
    for name, degree := range inDegree {
        if degree > 0 {
            cycleNodes = append(cycleNodes, name)
        }
    }

    // Build cycle path (simplified - just show nodes in cycle)
    cyclePath := strings.Join(cycleNodes, " → ")
    return NewCircularDependencyError("circular dependency detected: %s", cyclePath)
}
```

---

#### 5. Shell Metadata (`internal/domain/shellmeta.go`)

**Types**: Shell configuration metadata

```go
package domain

type ConfigFile struct {
    Path         string
    Scope        string // system | user
    Priority     int
    Note         string
    Alternatives []string
    Optional     bool
}

type ShellConfig struct {
    OS          []string
    Shell       string // bash | zsh | fish
    SessionType string // login_interactive | interactive_non_login | always
    Description string
    Files       []ConfigFile
}

func (sc *ShellConfig) Matches(os, shell, session string) bool {
    osMatch := false
    for _, o := range sc.OS {
        if strings.EqualFold(o, os) {
            osMatch = true
            break
        }
    }

    shellMatch := strings.EqualFold(sc.Shell, shell)
    sessionMatch := strings.EqualFold(sc.SessionType, session)

    return osMatch && shellMatch && sessionMatch
}

type ShellMetadata struct {
    Configs            []ShellConfig
    RecommendedTargets map[string]interface{}
}

func (sm *ShellMetadata) FindConfig(os, shell, session string) *ShellConfig {
    for i := range sm.Configs {
        if sm.Configs[i].Matches(os, shell, session) {
            return &sm.Configs[i]
        }
    }
    return nil
}

func (sm *ShellMetadata) GetRecommendedTarget(os, shell, session string) string {
    // Navigate nested map structure
    // Implementation depends on shell_configs.yaml format
    return ""
}
```

---

#### 6. Domain Errors (`internal/domain/errors.go`)

**Custom Error Types**

```go
package domain

import "fmt"

type ValidationError struct {
    Message string
}

func (e *ValidationError) Error() string {
    return e.Message
}

func NewValidationError(format string, args ...interface{}) *ValidationError {
    return &ValidationError{Message: fmt.Sprintf(format, args...)}
}

type CircularDependencyError struct {
    Message string
}

func (e *CircularDependencyError) Error() string {
    return e.Message
}

func NewCircularDependencyError(format string, args ...interface{}) *CircularDependencyError {
    return &CircularDependencyError{Message: fmt.Sprintf(format, args...)}
}

type FileNotFoundError struct {
    Path string
}

func (e *FileNotFoundError) Error() string {
    return fmt.Sprintf("file not found: %s", e.Path)
}
```

---

## Application Layer

### Purpose
Orchestrate domain logic to implement use cases. Each use case is a distinct operation.

### Components

#### 1. Builder (`internal/app/builder.go`)

**Use Case**: Build shell configuration from modules

```go
package app

import (
    "github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

type BuilderService struct {
    manifestParser ManifestParser
    fileReader     FileReader
    fileWriter     FileWriter
    resolver       *domain.Resolver
}

type ManifestParser interface {
    Parse(path string) (*domain.Manifest, error)
}

type FileReader interface {
    ReadFile(path string) (string, error)
    FileExists(path string) bool
}

type FileWriter interface {
    WriteFile(path string, content string) error
}

func NewBuilderService(parser ManifestParser, reader FileReader, writer FileWriter) *BuilderService {
    return &BuilderService{
        manifestParser: parser,
        fileReader:     reader,
        fileWriter:     writer,
        resolver:       domain.NewResolver(),
    }
}

type BuildOptions struct {
    ConfigDir string
    Manifest  string
    Output    string
    OS        string
    DryRun    bool
    Verbose   bool
}

func (s *BuilderService) Build(opts BuildOptions) (string, error) {
    // 1. Parse manifest
    manifest, err := s.manifestParser.Parse(opts.Manifest)
    if err != nil {
        return "", fmt.Errorf("failed to parse manifest: %w", err)
    }

    // 2. Build dependency graph
    graph, err := s.resolver.BuildGraph(manifest)
    if err != nil {
        return "", fmt.Errorf("failed to build graph: %w", err)
    }

    // 3. Topological sort with OS filtering
    modules, err := s.resolver.TopologicalSort(graph, opts.OS)
    if err != nil {
        return "", fmt.Errorf("failed to resolve dependencies: %w", err)
    }

    // 4. Read and concatenate module files
    var lines []string
    lines = append(lines, "# Generated by shellforge")
    lines = append(lines, fmt.Sprintf("# OS: %s", opts.OS))
    lines = append(lines, fmt.Sprintf("# Modules: %d", len(modules)))
    lines = append(lines, "")

    for _, module := range modules {
        filePath := filepath.Join(opts.ConfigDir, module.File)

        if !s.fileReader.FileExists(filePath) {
            return "", domain.NewFileNotFoundError(filePath)
        }

        content, err := s.fileReader.ReadFile(filePath)
        if err != nil {
            return "", fmt.Errorf("failed to read %s: %w", filePath, err)
        }

        lines = append(lines, fmt.Sprintf("\n# --- %s ---", module.Name))
        lines = append(lines, fmt.Sprintf("# %s", module.Description))
        lines = append(lines, content)
        lines = append(lines, "")
    }

    output := strings.Join(lines, "\n")

    // 5. Write output (unless dry-run)
    if !opts.DryRun && opts.Output != "" {
        if err := s.fileWriter.WriteFile(opts.Output, output); err != nil {
            return "", fmt.Errorf("failed to write output: %w", err)
        }
    }

    return output, nil
}
```

---

#### 2. Validator (`internal/app/validator.go`)

**Use Case**: Validate manifest and modules

```go
package app

type ValidatorService struct {
    manifestParser ManifestParser
    fileReader     FileReader
    resolver       *domain.Resolver
}

func NewValidatorService(parser ManifestParser, reader FileReader) *ValidatorService {
    return &ValidatorService{
        manifestParser: parser,
        fileReader:     reader,
        resolver:       domain.NewResolver(),
    }
}

type ValidationResult struct {
    Valid     bool
    Errors    []error
    Stats     ValidationStats
}

type ValidationStats struct {
    TotalModules  int
    MacModules    int
    LinuxModules  int
    Dependencies  int
}

func (s *ValidatorService) Validate(configDir, manifestPath string) (*ValidationResult, error) {
    result := &ValidationResult{
        Valid:  true,
        Errors: []error{},
    }

    // 1. Parse manifest
    manifest, err := s.manifestParser.Parse(manifestPath)
    if err != nil {
        result.Valid = false
        result.Errors = append(result.Errors, err)
        return result, nil
    }

    // 2. Validate manifest structure
    if errs := manifest.Validate(); len(errs) > 0 {
        result.Valid = false
        result.Errors = append(result.Errors, errs...)
    }

    // 3. Check file existence
    for _, module := range manifest.Modules {
        filePath := filepath.Join(configDir, module.File)
        if !s.fileReader.FileExists(filePath) {
            result.Valid = false
            result.Errors = append(result.Errors, domain.NewFileNotFoundError(filePath))
        }
    }

    // 4. Check for circular dependencies
    graph, err := s.resolver.BuildGraph(manifest)
    if err != nil {
        result.Valid = false
        result.Errors = append(result.Errors, err)
        return result, nil
    }

    for _, os := range []string{"Mac", "Linux"} {
        _, err := s.resolver.TopologicalSort(graph, os)
        if err != nil {
            result.Valid = false
            result.Errors = append(result.Errors, err)
            break
        }
    }

    // 5. Calculate statistics
    result.Stats = s.calculateStats(manifest)

    return result, nil
}

func (s *ValidatorService) calculateStats(manifest *domain.Manifest) ValidationStats {
    stats := ValidationStats{
        TotalModules: len(manifest.Modules),
    }

    depSet := make(map[string]bool)
    for _, module := range manifest.Modules {
        if module.AppliesTo("Mac") {
            stats.MacModules++
        }
        if module.AppliesTo("Linux") {
            stats.LinuxModules++
        }
        for _, dep := range module.Requires {
            depSet[dep] = true
        }
    }
    stats.Dependencies = len(depSet)

    return stats
}
```

---

#### 3. Other Application Services

**Deployer** (`internal/app/deployer.go`):
- Orchestrates: build → create snapshot → git commit → deploy
- Uses: GitBackup, FileWriter, BuilderService

**Migrator** (`internal/app/migrator.go`):
- Parse monolithic RC file
- Detect sections
- Categorize modules
- Create modular structure

**Initializer** (`internal/app/initializer.go`):
- Scan directories
- Infer dependencies
- Generate manifest

**Differ** (`internal/app/differ.go`):
- Build config
- Read existing file
- Generate diff (use github.com/sergi/go-diff)
- Calculate statistics

**Templates** (`internal/app/templates.go`):
- Define template types
- Field substitution
- Generate module files

---

## Infrastructure Layer

### Purpose
Implement interfaces defined by application layer. Adapters for external systems.

### Components

#### 1. YAML Parser (`internal/infra/yamlparser/parser.go`)

```go
package yamlparser

import (
    "gopkg.in/yaml.v3"
    "github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

type Parser struct{}

func New() *Parser {
    return &Parser{}
}

func (p *Parser) Parse(path string) (*domain.Manifest, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read manifest: %w", err)
    }

    var manifest domain.Manifest
    if err := yaml.Unmarshal(data, &manifest); err != nil {
        return nil, fmt.Errorf("failed to parse YAML: %w", err)
    }

    return &manifest, nil
}
```

---

#### 2. Filesystem (`internal/infra/filesystem/`)

**Reader** (`reader.go`):
```go
package filesystem

import (
    "github.com/spf13/afero"
)

type Reader struct {
    fs afero.Fs
}

func NewReader(fs afero.Fs) *Reader {
    return &Reader{fs: fs}
}

func (r *Reader) ReadFile(path string) (string, error) {
    data, err := afero.ReadFile(r.fs, path)
    if err != nil {
        return "", err
    }
    return string(data), nil
}

func (r *Reader) FileExists(path string) bool {
    exists, err := afero.Exists(r.fs, path)
    return err == nil && exists
}
```

**Writer** (`writer.go`):
```go
package filesystem

type Writer struct {
    fs afero.Fs
}

func NewWriter(fs afero.Fs) *Writer {
    return &Writer{fs: fs}
}

func (w *Writer) WriteFile(path string, content string) error {
    // Create parent directories
    dir := filepath.Dir(path)
    if err := w.fs.MkdirAll(dir, 0755); err != nil {
        return err
    }

    return afero.WriteFile(w.fs, path, []byte(content), 0644)
}
```

---

#### 3. Git Backup (`internal/infra/git/backup.go`)

```go
package git

import (
    "os/exec"
)

type Backup struct {
    backupDir string
}

func NewBackup(backupDir string) *Backup {
    return &Backup{backupDir: backupDir}
}

func (b *Backup) Init() error {
    // Check if git is available
    if _, err := exec.LookPath("git"); err != nil {
        return fmt.Errorf("git is not installed: %w", err)
    }

    // Check if already initialized
    gitDir := filepath.Join(b.backupDir, ".git")
    if _, err := os.Stat(gitDir); err == nil {
        return nil // Already initialized
    }

    // Initialize git repo
    cmd := exec.Command("git", "init", b.backupDir)
    return cmd.Run()
}

func (b *Backup) CreateSnapshot(targetFile, snapshotDir string) (string, error) {
    timestamp := time.Now().Format("2006-01-02_15-04-05")
    snapshotPath := filepath.Join(snapshotDir, timestamp)

    // Copy file to snapshot location
    input, err := os.ReadFile(targetFile)
    if err != nil {
        return "", err
    }

    if err := os.MkdirAll(filepath.Dir(snapshotPath), 0755); err != nil {
        return "", err
    }

    if err := os.WriteFile(snapshotPath, input, 0644); err != nil {
        return "", err
    }

    return snapshotPath, nil
}

func (b *Backup) Commit(message string) error {
    // Add all changes
    cmd := exec.Command("git", "-C", b.backupDir, "add", ".")
    if err := cmd.Run(); err != nil {
        return err
    }

    // Commit
    cmd = exec.Command("git", "-C", b.backupDir, "commit", "-m", message)
    return cmd.Run()
}
```

---

#### 4. Shell Data Loader (`internal/infra/shelldata/loader.go`)

```go
package shelldata

import (
    _ "embed"
    "gopkg.in/yaml.v3"
    "github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

//go:embed ../../data/shell_configs.yaml
var shellConfigsData []byte

type Loader struct{}

func New() *Loader {
    return &Loader{}
}

func (l *Loader) Load() (*domain.ShellMetadata, error) {
    var metadata domain.ShellMetadata
    if err := yaml.Unmarshal(shellConfigsData, &metadata); err != nil {
        return nil, fmt.Errorf("failed to parse shell metadata: %w", err)
    }
    return &metadata, nil
}
```

---

## CLI Layer

### Purpose
Handle user interaction via Cobra commands. Parse flags, call application services, format output.

### Structure

#### Root Command (`internal/cli/root.go`)

```go
package cli

import (
    "github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "shellforge",
        Short: "Build tool for modular shell configurations",
        Long: `Shellforge - Forge your shell configuration from modular pieces.

🚀 Quick Start Guide:
  ...
`,
        Version: "1.0.0",
    }

    // Add subcommands
    cmd.AddCommand(NewBuildCmd())
    cmd.AddCommand(NewValidateCmd())
    cmd.AddCommand(NewInitCmd())
    cmd.AddCommand(NewMigrateCmd())
    cmd.AddCommand(NewDiffCmd())
    cmd.AddCommand(NewRestoreCmd())
    cmd.AddCommand(NewCleanCmd())
    cmd.AddCommand(NewListCmd())
    cmd.AddCommand(NewInfoCmd())
    cmd.AddCommand(NewTemplateCmd())

    return cmd
}
```

---

#### Build Command (`internal/cli/build.go`)

```go
package cli

import (
    "github.com/spf13/cobra"
    "github.com/gizzahub/gzh-cli-shellforge/internal/app"
)

func NewBuildCmd() *cobra.Command {
    var opts struct {
        configDir  string
        manifest   string
        output     string
        autoOutput bool
        os         string
        shell      string
        session    string
        deploy     bool
        dryRun     bool
        verbose    bool
    }

    cmd := &cobra.Command{
        Use:   "build",
        Short: "Generate shell configuration file",
        Long:  `Build shell configuration from modular pieces with dependency resolution.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // Initialize services
            fs := afero.NewOsFs()
            parser := yamlparser.New()
            reader := filesystem.NewReader(fs)
            writer := filesystem.NewWriter(fs)
            builder := app.NewBuilderService(parser, reader, writer)

            // Detect OS if not specified
            if opts.os == "" {
                opts.os = detectOS()
            }

            // Auto-detect output if requested
            if opts.autoOutput && opts.output == "" {
                // Use ShellMetadata to find recommended target
                // ...
            }

            // Build
            buildOpts := app.BuildOptions{
                ConfigDir: opts.configDir,
                Manifest:  opts.manifest,
                Output:    opts.output,
                OS:        opts.os,
                DryRun:    opts.dryRun,
                Verbose:   opts.verbose,
            }

            output, err := builder.Build(buildOpts)
            if err != nil {
                return formatError(err)
            }

            // Print success message
            fmt.Println("✓ Build completed successfully")

            if opts.dryRun {
                fmt.Println("\n--- Generated output ---")
                fmt.Println(output)
            }

            // Deploy if requested
            if opts.deploy && !opts.dryRun {
                // Call deployer service
                // ...
            }

            return nil
        },
    }

    // Flags
    cmd.Flags().StringVarP(&opts.configDir, "config-dir", "c", "", "Directory containing shell modules (required)")
    cmd.Flags().StringVarP(&opts.manifest, "manifest", "m", "", "Path to manifest.yaml (required)")
    cmd.Flags().StringVarP(&opts.output, "output", "o", "", "Output file path")
    cmd.Flags().BoolVar(&opts.autoOutput, "auto-output", false, "Auto-detect output path")
    cmd.Flags().StringVar(&opts.os, "os", "", "Target OS (macos, linux, etc.)")
    cmd.Flags().StringVar(&opts.shell, "shell", "", "Target shell (bash, zsh, fish)")
    cmd.Flags().StringVar(&opts.session, "session", "", "Session type (login, interactive, always)")
    cmd.Flags().BoolVar(&opts.deploy, "deploy", false, "Deploy with backup")
    cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview without writing")
    cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", false, "Verbose output")

    cmd.MarkFlagRequired("config-dir")
    cmd.MarkFlagRequired("manifest")

    return cmd
}

func formatError(err error) error {
    // Format domain errors with helpful messages
    switch e := err.(type) {
    case *domain.ValidationError:
        return fmt.Errorf("✗ Validation failed:\n  %s\n\n💡 How to fix:\n  Check manifest.yaml for errors", e.Message)
    case *domain.CircularDependencyError:
        return fmt.Errorf("✗ %s\n\n💡 How to fix:\n  Edit manifest.yaml and remove one dependency from the circular chain", e.Message)
    default:
        return fmt.Errorf("✗ Error: %w", err)
    }
}
```

---

## Data Flow

### Build Command Flow

```
User Input (CLI Flags)
    ↓
CLI Layer (build.go)
    ↓ Parse flags, validate
    ↓
Application Layer (BuilderService)
    ↓ Orchestrate use case
    ├─→ Infrastructure (YAMLParser)
    │      ↓ Parse manifest.yaml
    │      ↓
    ├─→ Domain (Resolver)
    │      ↓ Build graph
    │      ↓ Topological sort
    │      ↓ OS filtering
    │      ↓
    ├─→ Infrastructure (FileReader)
    │      ↓ Read module files
    │      ↓
    └─→ Infrastructure (FileWriter)
           ↓ Write output file
           ↓
User Output (Terminal)
```

---

## Testing Strategy

### Unit Tests

**Domain Layer** (100% coverage):
- Module.AppliesTo()
- Manifest.Validate()
- Graph construction
- Topological sort
- Cycle detection

**Application Layer** (>80% coverage):
- BuilderService.Build() with mock dependencies
- ValidatorService.Validate()
- Each use case independently

**Infrastructure Layer** (>70% coverage):
- YAML parser with sample files
- Filesystem with afero.MemMapFs
- Git backup with mock exec (or integration test)

**CLI Layer** (>60% coverage):
- Flag parsing
- Error formatting
- Command execution (integration-style)

### Integration Tests

- End-to-end: CLI → build → validate output
- Real filesystem with temp directories
- Real YAML files (use examples/)

### Table-Driven Tests

```go
func TestTopologicalSort(t *testing.T) {
    tests := []struct {
        name     string
        modules  []domain.Module
        os       string
        expected []string
        wantErr  bool
    }{
        {
            name: "linear dependencies",
            modules: []domain.Module{
                {Name: "a", Requires: []string{}},
                {Name: "b", Requires: []string{"a"}},
                {Name: "c", Requires: []string{"b"}},
            },
            os:       "Mac",
            expected: []string{"a", "b", "c"},
            wantErr:  false,
        },
        {
            name: "circular dependency",
            modules: []domain.Module{
                {Name: "a", Requires: []string{"b"}},
                {Name: "b", Requires: []string{"a"}},
            },
            os:       "Mac",
            expected: nil,
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

---

## Error Handling

### Error Wrapping

```go
if err := parser.Parse(path); err != nil {
    return nil, fmt.Errorf("failed to parse manifest: %w", err)
}
```

### Custom Error Types

```go
switch err := err.(type) {
case *domain.ValidationError:
    // Handle validation error
case *domain.CircularDependencyError:
    // Handle circular dependency
default:
    // Generic error
}
```

### Exit Codes

```go
func main() {
    if err := cli.NewRootCmd().Execute(); err != nil {
        switch err.(type) {
        case *domain.ValidationError, *domain.CircularDependencyError:
            os.Exit(1) // User error
        default:
            os.Exit(2) // System error
        }
    }
}
```

---

## Appendix

### Dependency Injection Example

```go
// main.go
func main() {
    // Infrastructure
    fs := afero.NewOsFs()
    parser := yamlparser.New()
    reader := filesystem.NewReader(fs)
    writer := filesystem.NewWriter(fs)

    // Application
    builder := app.NewBuilderService(parser, reader, writer)
    validator := app.NewValidatorService(parser, reader)

    // CLI
    rootCmd := cli.NewRootCmd()
    rootCmd.AddCommand(cli.NewBuildCmd(builder))
    rootCmd.AddCommand(cli.NewValidateCmd(validator))

    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

---

**Document Status**: Ready for review
**Next Steps**: Write TECH_STACK.md with library decisions
