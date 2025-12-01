# Shellforge Context for LLM

> **Purpose**: Quick reference for AI coding assistants working with Shellforge codebase

---

## Project Overview

**Name**: Shellforge (Go implementation)
**Binary**: `gz-shellforge`
**Code Entry**: `cmd/shellforge/main.go`
**Version**: 0.2.0-alpha
**Purpose**: Modular shell configuration builder with automatic dependency resolution

---

## Core Concepts

### What It Does
- Converts monolithic `.zshrc`/`.bashrc` to modular structure
- Resolves module dependencies (topological sort)
- Filters modules by OS (Mac, Linux)
- Generates unified shell configs
- Provides backup/restore system

### Key Features
- ✅ Dependency resolution (Kahn's algorithm)
- ✅ OS-specific filtering
- ✅ Migration from monolithic configs
- ✅ Template generation (6 built-in types)
- ✅ Backup/restore with git versioning
- ✅ Diff comparison (4 formats)

---

## Quick Command Reference

### Build Commands (CRITICAL: Use make, not go build)
```bash
make build          # Build binary to build/gz-shellforge
make test           # Run all tests
make test-coverage  # Generate coverage report
make install        # Install to $GOPATH/bin
make clean          # Remove build artifacts
```

### CLI Commands
```bash
gz-shellforge build --manifest manifest.yaml --os Mac --output ~/.zshrc
gz-shellforge validate --manifest manifest.yaml
gz-shellforge list --filter Mac
gz-shellforge migrate ~/.zshrc
gz-shellforge template generate <type> <name>
gz-shellforge backup --file ~/.zshrc
gz-shellforge restore --file ~/.zshrc --snapshot <timestamp>
gz-shellforge diff <file1> <file2> --format summary
```

---

## Architecture (Hexagonal)

### 4-Layer Structure
```
internal/
├── domain/      # Pure business logic (NO external deps)
├── app/         # Use cases (depends on domain only)
├── infra/       # Infrastructure adapters
└── cli/         # CLI commands (Cobra)
```

### CRITICAL Dependency Rules
1. **Domain**: NEVER imports app/infra/cli
2. **App**: Only imports domain interfaces
3. **Infra**: Implements app interfaces
4. **CLI**: Depends on app layer

**Violation = Breaking change!**

---

## Key Domain Logic

### Module Resolution
- **Algorithm**: Topological sort (Kahn's BFS-based)
- **Location**: `internal/domain/resolver.go`
- **Complexity**: O(V + E) where V=modules, E=dependencies
- **Cycle Detection**: Built-in, reports full cycle path

### OS Filtering
- **Implementation**: `Module.AppliesTo(targetOS string) bool`
- **Empty OS field**: Applies to all platforms
- **Case-insensitive**: "Mac", "mac", "MAC" all valid

### File Organization
```
manifest.yaml         # Module definitions
modules/
├── init.d/          # Early init (PATH, OS detection)
├── rc_pre.d/        # Tool setup (nvm, rbenv)
└── rc_post.d/       # Aliases, functions
```

---

## Critical Rules for Development

### Build System
- ✅ **ALWAYS use `make build`**
- ❌ **NEVER use `go build` directly**
- Binary name: `gz-shellforge`
- Code path: `cmd/shellforge/`

### Testing
- Domain layer: >80% coverage required
- Use `afero.MemMapFs` for filesystem mocking
- Table-driven tests preferred
- Integration tests in `examples/`

### Error Handling
- Wrap with context: `fmt.Errorf("context: %w", err)`
- Custom domain errors: `ValidationError`, `CircularDependencyError`
- CLI formats errors for users

### Naming Conventions
- Binary: `gz-shellforge`
- Module: `github.com/gizzahub/gzh-cli-shellforge`
- Command in help: `shellforge` (not gz-shellforge)

---

## File Paths Reference

### Source Code
```
cmd/shellforge/main.go              # Entry point
internal/domain/resolver.go         # Topological sort
internal/domain/module.go           # Module entity
internal/app/builder.go             # Build use case
internal/cli/build.go               # Build command
internal/infra/yamlparser/parser.go # YAML parsing
```

### Documentation
```
README.md                           # Main documentation (302 lines - simplified)
CLAUDE.md                           # Development guide
docs/user/                          # User documentation
docs/dev/                           # Developer documentation
.claude/CONTEXT.md                  # This file (LLM entry point)
```

### Examples
```
examples/manifest.yaml              # Example manifest
examples/modules/                   # Example shell modules
examples/sample.zshrc               # Sample RC file
examples/workflow-demo.sh           # Automated demo
```

---

## Common Tasks

### Adding New CLI Command
1. Create `internal/cli/newcommand.go`
2. Follow pattern from `build.go`
3. Add to root in `internal/cli/root.go`
4. Create tests in `internal/cli/newcommand_test.go`
5. Update README.md command reference

### Adding New Use Case
1. Define interface in `internal/app/`
2. Implement with injected dependencies
3. Add tests with mocked infrastructure
4. Wire up in CLI command

### Adding Infrastructure Adapter
1. Check if interface exists in `app/`
2. Create in `infra/` subdirectory
3. Implement interface methods
4. Add tests (in-memory or temp files)
5. Wire up in `main.go`

---

## Testing Patterns

### Domain Tests (Pure Logic)
```go
func TestModule_AppliesTo(t *testing.T) {
    tests := []struct {
        name     string
        module   Module
        targetOS string
        want     bool
    }{
        // Table-driven cases
    }
    // No mocks needed
}
```

### App Tests (With Mocks)
```go
fs := afero.NewMemMapFs()
afero.WriteFile(fs, "test.yaml", []byte("..."), 0644)
reader := filesystem.NewReader(fs)
// Test with mocked filesystem
```

---

## Performance Benchmarks

| Metric | Python | Go | Improvement |
|--------|--------|----|----|
| Startup | ~200ms | <10ms | 20x faster |
| Build (10 modules) | ~300ms | <50ms | 6x faster |
| Memory | ~80MB | <10MB | 8x lighter |
| Binary size | ~40MB | ~8MB | 5x smaller |

---

## For Detailed Information

### Development Tasks
→ See `DEVELOPMENT.md` (current CLAUDE.md)
- Complete architecture details
- Dependency injection patterns
- Code style guidelines
- Testing strategy

### User Assistance
→ See `docs/user/40-command-reference.md`
- All command syntax
- Real-world examples
- Common workflows

### API Integration
→ See `docs/reference/api.md` (TODO)
- Public API reference
- Integration guide

---

## Quick Debugging

### Build Fails
```bash
make clean
make build
```

### Tests Fail
```bash
go test ./... -v
go test ./internal/domain -v  # Specific package
```

### Check Coverage
```bash
make test-coverage
open coverage.html
```

---

## Current Status

**Version**: 0.2.0-alpha
**Test Coverage**: 70.1% overall
- Domain: 88.7%
- Infrastructure: 77.6-100%
- Application: 86.5%
- CLI: 45.8%

**Features**:
- ✅ Build, validate, list, migrate, template, backup, restore, cleanup, diff
- ⏳ Plugin system (planned)

---

## Important Notes

1. **NEVER commit without tests** for new features
2. **NEVER push to remote** without user permission
3. **ALWAYS validate manifest** before building
4. **ALWAYS use make build**, never go build
5. **ALWAYS check dependency rules** in architecture

---

**Last Updated**: 2025-12-01
**For**: Claude Code and AI coding assistants
**Maintained by**: Shellforge contributors
