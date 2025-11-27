# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Quick Reference

**Binary Name**: `gz-shellforge` (execution command)
**Code Path**: `cmd/shellforge/main.go` (source code location)
**Current Version**: 0.2.0-alpha

## Essential Commands

### Build & Test

```bash
# Build the binary (always use make, not go build directly)
make build
./build/gz-shellforge --version

# Run tests
make test

# Test with coverage
make test-coverage
open coverage.html

# Clean build artifacts
make clean

# Install to $GOPATH/bin
make install
```

### Development Workflow

```bash
# Run without building
go run cmd/shellforge/main.go build --os Mac --dry-run

# Run specific test
go test ./internal/domain -run TestModule_AppliesTo -v

# Run tests for specific package
go test ./internal/app -v

# Check for race conditions
go test -race ./...

# Format and lint
make lint
```

## Architecture Overview

This project follows **Hexagonal Architecture** (ports & adapters) with **Clean Architecture** principles. The codebase is strictly organized into 4 layers with clear dependency rules:

### Layer Structure

```
internal/
├── domain/      # Pure business logic (NO external dependencies)
├── app/         # Use cases (depends on domain interfaces only)
├── infra/       # Infrastructure adapters (implements domain interfaces)
└── cli/         # CLI commands (depends on app layer)
```

### Dependency Rules (Critical)

1. **Domain Layer** (`internal/domain/`)
   - Contains: Module, Manifest, Graph, Resolver entities
   - NO imports from other internal packages
   - NO imports of external libraries (except standard library)
   - Pure Go business logic only
   - When adding domain logic: NEVER import from app/infra/cli

2. **Application Layer** (`internal/app/`)
   - Contains: BuilderService, ValidatorService (use cases)
   - Depends ONLY on domain interfaces
   - Defines interfaces for infrastructure (ManifestParser, FileReader, FileWriter)
   - Orchestrates domain logic
   - When adding services: Define interfaces, don't import infra directly

3. **Infrastructure Layer** (`internal/infra/`)
   - Contains: YAML parser, filesystem, git wrappers
   - Implements interfaces defined by app layer
   - Uses external libraries (yaml.v3, afero, os/exec)
   - When adding adapters: Implement app layer interfaces

4. **CLI Layer** (`internal/cli/`)
   - Contains: Cobra commands (build, validate, list)
   - Depends on app layer services
   - Handles flag parsing and output formatting
   - When adding commands: Inject app services via constructors

### Key Architectural Patterns

**Dependency Injection Example:**
```go
// main.go wires everything together
fs := afero.NewOsFs()
parser := yamlparser.New()
reader := filesystem.NewReader(fs)
writer := filesystem.NewWriter(fs)
builder := app.NewBuilderService(parser, reader, writer)

buildCmd := cli.NewBuildCmd(builder)  // Inject service
```

**Interface-Based Design:**
```go
// app/builder.go defines what it needs
type ManifestParser interface {
    Parse(path string) (*domain.Manifest, error)
}

// infra/yamlparser/parser.go implements it
type Parser struct{}
func (p *Parser) Parse(path string) (*domain.Manifest, error) { ... }
```

## Core Domain Logic

### Module & Dependency Resolution

The heart of Shellforge is topological sorting of shell module dependencies:

1. **Graph Construction** (`internal/domain/graph.go`)
   - Modules are nodes, dependencies are directed edges
   - Each module has in-degree count (number of dependencies)

2. **Topological Sort** (`internal/domain/resolver.go`)
   - Uses Kahn's Algorithm (BFS-based)
   - Complexity: O(V + E) where V=modules, E=dependencies
   - OS filtering is integrated into the sort
   - Cycle detection reports clear error paths

3. **OS Filtering** (`internal/domain/module.go`)
   - Module.AppliesTo(targetOS) checks OS field
   - Empty OS field = applies to all platforms
   - Case-insensitive matching

### Critical Implementation Details

**Circular Dependency Detection:**
- After topological sort, if `result.length != nodes.length`, cycle exists
- Remaining nodes are part of the cycle
- Implementation: `resolver.go:detectCycle()`

**Module Load Order:**
- Modules with zero dependencies load first
- Each module loads only after all its dependencies
- Order is deterministic (not random)

## Testing Strategy

### Test Coverage Targets

- Domain layer: >80% (critical business logic)
- Application layer: >70%
- Infrastructure layer: >60%
- CLI layer: >50%

### Testing Patterns

**Domain Tests** (pure logic, no mocks):
```go
func TestModule_AppliesTo(t *testing.T) {
    tests := []struct {
        name     string
        module   Module
        targetOS string
        want     bool
    }{
        // Table-driven test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test pure logic
        })
    }
}
```

**App Layer Tests** (with mocks):
```go
// Use afero.MemMapFs for filesystem mocking
fs := afero.NewMemMapFs()
afero.WriteFile(fs, "test.yaml", []byte("..."), 0644)
reader := filesystem.NewReader(fs)
// Test service with mocked filesystem
```

**Integration Tests:**
- Use `examples/` directory for real manifest testing
- Create temp directories for file operations
- Clean up after tests

## File Organization Rules

### DO NOT Create/Modify These Without Discussion:

1. **Domain Entities**: Adding new entities to `internal/domain/` changes core model
2. **Interface Definitions**: New interfaces in `app/` affect all layers below
3. **Manifest Format**: Changes break backward compatibility

### Safe to Add/Modify:

1. **CLI Commands**: New commands in `internal/cli/` (follow existing patterns)
2. **Infrastructure Adapters**: New implementations in `internal/infra/`
3. **Test Files**: Always add tests alongside new code
4. **Documentation**: README, docs/* files

## Common Development Tasks

### Adding a New CLI Command

1. Create `internal/cli/newcommand.go`
2. Follow pattern from `build.go` or `validate.go`
3. Add to root command in `internal/cli/root.go`
4. Create test file `internal/cli/newcommand_test.go`
5. Update README.md with command documentation

### Adding a New Use Case

1. Define interface in `internal/app/`
2. Implement service with injected dependencies
3. Add tests with mocked infrastructure
4. Wire up in CLI command

### Adding Infrastructure Adapter

1. Check if interface exists in `app/`
2. Create adapter in `infra/` subdirectory
3. Implement interface methods
4. Add tests (use in-memory or temp files)
5. Wire up in main.go

## Important Constraints

### Naming Convention

- **Binary name**: `gz-shellforge` (for gzh-cli tool family consistency)
- **Code path**: `cmd/shellforge/` (not `cmd/gz-shellforge/`)
- **Module name**: `github.com/gizzahub/gzh-cli-shellforge`
- **Command name in help**: `shellforge` (not gz-shellforge)

### Build System

- **ALWAYS use `make build`**, never `go build` directly
- Makefile defines `BINARY_NAME=gz-shellforge` and `MAIN_PATH=cmd/shellforge/main.go`
- Build output: `build/gz-shellforge`

### Dependencies

- Prefer standard library over external packages
- Current external deps: cobra, yaml.v3, afero, testify
- NO CGO dependencies (must produce static binary)
- Custom implementations for simple algorithms (e.g., topological sort)

### Error Handling

- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Use custom error types in domain layer (ValidationError, CircularDependencyError)
- CLI layer formats errors for user-friendly output
- Never panic in production code (only in tests with t.Fatal)

## Code Style

### Follow Go Standards

- Run `gofmt` before committing (enforced by `make lint`)
- Use `go vet` to catch common mistakes
- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use table-driven tests for multiple scenarios

### Project-Specific Conventions

**Package Comments:**
```go
// Package domain contains the core business logic for Shellforge.
// It defines entities (Module, Manifest) and algorithms (topological sort).
package domain
```

**Interface Naming:**
- Use noun phrases: `ManifestParser`, `FileReader`, `FileWriter`
- NOT verb phrases: `ParseManifest`, `ReadFile`

**Constructor Pattern:**
```go
func NewBuilderService(parser ManifestParser, reader FileReader, writer FileWriter) *BuilderService {
    return &BuilderService{
        manifestParser: parser,
        fileReader:     reader,
        fileWriter:     writer,
        resolver:       domain.NewResolver(),
    }
}
```

## Backward Compatibility

### Must Maintain

1. **Manifest YAML format**: 100% compatible with Python version
2. **CLI interface**: Same command names, flags, and behavior
3. **Error messages**: Same or better clarity
4. **Output format**: Compatible with existing workflows

### Can Improve

1. Performance (faster is better)
2. Error messages (clearer is better)
3. Internal implementation (as long as interface stays same)

## Key Files Reference

- `cmd/shellforge/main.go` - Entry point, wiring
- `internal/domain/resolver.go` - Topological sort algorithm
- `internal/app/builder.go` - Build use case
- `internal/cli/root.go` - CLI root command
- `Makefile` - Build automation
- `examples/manifest.yaml` - Example configuration
- `ARCHITECTURE.md` - Detailed architecture documentation
- `TECH_STACK.md` - Library choices and rationale

## Troubleshooting

### Build fails with "command not found"
- Ensure Go 1.21+ is installed: `go version`
- Check GOPATH: `echo $GOPATH`

### Tests fail with filesystem errors
- Tests should use `afero.MemMapFs`, not real filesystem
- Check test setup creates required files in memory

### Binary name doesn't match
- Check Makefile: `BINARY_NAME=gz-shellforge`
- After changes, run `make clean && make build`

### Import cycle detected
- Domain should NEVER import app/infra/cli
- App should NEVER import infra/cli
- Infra should NEVER import cli
- Check import statements follow dependency rules
