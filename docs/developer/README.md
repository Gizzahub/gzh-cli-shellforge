# Developer Documentation

Welcome to Shellforge developer documentation! Everything you need to contribute to the project.

---

## 🚀 Quick Start for Contributors

### 1. Setup Development Environment

```bash
# Clone repository
git clone https://github.com/gizzahub/gzh-cli-shellforge.git
cd gzh-cli-shellforge

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Verify
./build/gz-shellforge --version
```

**Time to setup:** 5 minutes

### 2. Make Your First Contribution

```bash
# Create feature branch
git checkout -b feature/my-feature

# Make changes
vim internal/domain/module.go

# Run tests
make test

# Commit (with proper format)
git commit -m "feat(domain): add new feature"

# Push and create PR
git push origin feature/my-feature
```

**See:** [Contributing Guide](30-contributing.md)

---

## 📚 Core Documentation

### Architecture & Design

- **[Architecture Overview](00-architecture.md)** 📐
  - Hexagonal architecture (4 layers)
  - Dependency rules
  - Design patterns
  - Project structure

- **[Tech Stack](50-tech-stack.md)** ⚙️
  - Technology decisions
  - Library choices
  - Rationale and trade-offs

- **[Performance Benchmarks](60-benchmarks.md)** 📊
  - Benchmark results
  - Performance comparisons
  - Optimization guidelines

### Development Guides

- **[Development Setup](10-development-setup.md)** 🛠️
  - Prerequisites
  - IDE configuration
  - Development workflow
  - Debugging tips

- **[Testing Guide](20-testing-guide.md)** 🧪
  - Testing strategy
  - Writing unit tests
  - Integration tests
  - Coverage targets

- **[Contributing Guide](30-contributing.md)** 🤝
  - How to contribute
  - Code review process
  - PR guidelines
  - Issue triage

- **[Code Style Guide](40-code-style.md)** 📝
  - Go conventions
  - Project-specific patterns
  - Naming conventions
  - Error handling

---

## 🏗️ Architecture Quick Reference

### 4-Layer Hexagonal Architecture

```
cmd/shellforge/          # Entry point
    ↓
internal/cli/            # CLI commands (Cobra)
    ↓
internal/app/            # Use cases (business workflows)
    ↓
internal/domain/         # Pure business logic
    ↑
internal/infra/          # Infrastructure adapters
```

### Critical Dependency Rules

1. **Domain**: NEVER imports app/infra/cli
2. **App**: Only imports domain interfaces
3. **Infra**: Implements app interfaces
4. **CLI**: Depends on app layer

**Violation = Breaking change!**

See [Architecture Document](00-architecture.md) for details.

---

## 🧪 Testing Quick Reference

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./internal/domain -v

# With coverage
make test-coverage
open coverage.html

# Watch mode
go test ./... -v -count=1

# Benchmarks
make bench
```

### Test Coverage Targets

| Layer | Target | Current |
|-------|--------|---------|
| Domain | >80% | 88.7% ✅ |
| Application | >70% | 86.5% ✅ |
| Infrastructure | >60% | 77.6-100% ✅ |
| CLI | >50% | 45.8% ⚠️ |

### Writing Tests

```go
// Domain tests (pure logic, no mocks)
func TestModule_AppliesTo(t *testing.T) {
    tests := []struct {
        name     string
        module   Module
        targetOS string
        want     bool
    }{
        // Table-driven cases
    }
}

// App tests (with mocks)
func TestBuilderService_Build(t *testing.T) {
    fs := afero.NewMemMapFs()
    // Use in-memory filesystem
}
```

See [Testing Guide](20-testing-guide.md) for details.

---

## 💻 Development Workflow

### Standard Development Cycle

```bash
1. Create feature branch
   git checkout -b feature/my-feature

2. Write code + tests
   vim internal/domain/module.go
   vim internal/domain/module_test.go

3. Run tests locally
   make test

4. Commit with proper format
   git commit -m "feat(domain): add feature"

5. Push and create PR
   git push origin feature/my-feature

6. Address review feedback
   git commit --amend  # or new commit

7. Merge after approval
   (via GitHub PR)
```

### Build Commands

```bash
# Build binary
make build

# Build for all platforms
make build-all

# Install to $GOPATH/bin
make install

# Clean artifacts
make clean

# Format code
make lint

# Full validation
make validate
```

**CRITICAL:** Always use `make build`, never `go build` directly.

---

## 📦 Project Structure

### Source Code

```
cmd/
└── shellforge/
    └── main.go              # Entry point

internal/
├── domain/                  # Pure business logic
│   ├── module.go           # Module entity
│   ├── manifest.go         # Manifest entity
│   ├── resolver.go         # Dependency resolver
│   └── graph.go            # Dependency graph
├── app/                    # Use cases
│   ├── builder.go          # Build service
│   ├── validator.go        # Validation service
│   └── migrator.go         # Migration service
├── infra/                  # Infrastructure
│   ├── yamlparser/         # YAML parsing
│   ├── filesystem/         # File operations
│   └── git/                # Git operations
└── cli/                    # CLI commands
    ├── root.go             # Root command
    ├── build.go            # Build command
    └── validate.go         # Validate command

pkg/
└── cmd/                    # Public API (for library usage)
```

### Documentation

```
docs/
├── user/                   # User documentation
├── developer/              # Developer documentation (you are here)
├── reference/              # API reference
└── design/                 # Design documents

.claude/
├── CONTEXT.md              # LLM entry point
└── DEVELOPMENT.md          # Detailed dev guide for AI
```

---

## 🎯 Common Development Tasks

### Adding New CLI Command

1. Create `internal/cli/newcommand.go`
2. Follow pattern from `build.go` or `validate.go`
3. Add to root in `internal/cli/root.go`
4. Create tests `internal/cli/newcommand_test.go`
5. Update README.md

**Example structure:**
```go
package cli

func NewMyCmd(service *app.MyService) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "mycommand",
        Short: "Description",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
        },
    }
    // Add flags
    return cmd
}
```

### Adding New Use Case

1. Define interface in `internal/app/`
2. Implement service with injected dependencies
3. Add tests with mocked infrastructure
4. Wire up in CLI command
5. Update documentation

### Adding Infrastructure Adapter

1. Check if interface exists in `app/`
2. Create adapter in `infra/` subdirectory
3. Implement interface methods
4. Add tests (use in-memory or temp files)
5. Wire up in `main.go`

### Adding New Feature

1. **Plan**: Write design document in `docs/design/decisions/`
2. **Domain**: Implement business logic in `internal/domain/`
3. **Tests**: Write comprehensive tests
4. **App**: Create use case in `internal/app/`
5. **Infra**: Implement adapters in `internal/infra/`
6. **CLI**: Add command in `internal/cli/`
7. **Docs**: Update user documentation
8. **Validate**: Run full test suite
9. **PR**: Submit pull request

---

## 🐛 Debugging

### Debug with Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug binary
dlv debug cmd/shellforge/main.go -- build --os Mac --dry-run

# Debug tests
dlv test ./internal/domain -- -test.run TestModule_AppliesTo
```

### Debug with Print Statements

```go
// Temporary debugging
import "fmt"

func MyFunc() {
    fmt.Printf("DEBUG: value=%v\n", value)
    // Remove before commit!
}
```

### Verbose Logging

```bash
# Enable verbose output
gz-shellforge build --os Mac --verbose

# Or in code
if verbose {
    log.Printf("Detailed info: %v", data)
}
```

---

## 📊 Code Quality Standards

### Go Best Practices

- ✅ Use `go fmt` before committing
- ✅ Run `go vet` to catch common mistakes
- ✅ Follow [Effective Go](https://golang.org/doc/effective_go.html)
- ✅ Use table-driven tests
- ✅ Write descriptive commit messages

### Project-Specific Conventions

**Naming:**
- Interface names: Nouns (e.g., `ManifestParser`, `FileReader`)
- Function names: Verbs (e.g., `Build()`, `Validate()`)
- Test names: `Test{Type}_{Method}_{Scenario}`

**Error Handling:**
```go
// Always wrap errors with context
return fmt.Errorf("failed to build: %w", err)

// Use custom domain errors
return &ValidationError{Message: "invalid module"}
```

**Dependencies:**
```go
// Constructor pattern with dependency injection
func NewBuilderService(
    parser ManifestParser,
    reader FileReader,
    writer FileWriter,
) *BuilderService {
    return &BuilderService{
        parser: parser,
        reader: reader,
        writer: writer,
    }
}
```

---

## 🔧 Tools & IDE Setup

### Recommended IDEs

**Visual Studio Code:**
- Install Go extension
- Enable format on save
- Configure gopls

**GoLand:**
- Best IDE for Go
- Built-in test runner
- Excellent debugger

### Essential Tools

```bash
# Linters
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Test coverage
go install github.com/axw/gocov/gocov@latest

# Debugger
go install github.com/go-delve/delve/cmd/dlv@latest
```

---

## 📖 Additional Resources

### Internal Resources

- **[Design Documents](../design/)** - PRD, requirements, decisions
- **[API Reference](../reference/api.md)** - Public API documentation
- **[User Documentation](../user/)** - User-facing docs

### External Resources

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

---

## 🤝 Getting Help

### For Development Questions

- **GitHub Discussions**: [Ask a question](https://github.com/gizzahub/gzh-cli-shellforge/discussions)
- **Architecture Questions**: See [00-architecture.md](00-architecture.md)
- **Testing Questions**: See [20-testing-guide.md](20-testing-guide.md)

### For Contributing

- **Contributing Guide**: [30-contributing.md](30-contributing.md)
- **Code Style**: [40-code-style.md](40-code-style.md)
- **GitHub Issues**: [Open issues](https://github.com/gizzahub/gzh-cli-shellforge/issues)

---

## 📝 Documentation Status

| Document | Status | Priority |
|----------|--------|----------|
| README.md | ✅ Complete | - |
| 00-architecture.md | ✅ Complete | - |
| 10-development-setup.md | ⚠️ TODO | High |
| 20-testing-guide.md | ⚠️ TODO | High |
| 30-contributing.md | ⚠️ TODO | High |
| 40-code-style.md | ⚠️ TODO | Medium |
| 50-tech-stack.md | ✅ Complete | - |
| 60-benchmarks.md | ✅ Complete | - |

---

## 🎯 Quick Links

### For New Contributors
1. [Development Setup](10-development-setup.md) (TODO)
2. [Contributing Guide](30-contributing.md) (TODO)
3. [Code Style Guide](40-code-style.md) (TODO)

### For Core Contributors
1. [Architecture Overview](00-architecture.md)
2. [Testing Guide](20-testing-guide.md) (TODO)
3. [Tech Stack](50-tech-stack.md)

### For Maintainers
1. [Design Documents](../design/)
2. [Performance Benchmarks](60-benchmarks.md)
3. [API Reference](../reference/) (TODO)

---

**Last Updated**: 2025-12-01
**Maintainers**: Shellforge Contributors
**License**: MIT
