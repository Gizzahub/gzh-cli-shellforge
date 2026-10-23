# Developer Documentation

Welcome to Shellforge developer documentation! Everything you need to contribute to the project.

______________________________________________________________________

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

**See:** [Contributing Guide](CONTRIBUTING.md)

______________________________________________________________________

## 📚 Documentation Index

### Architecture & Design

- **[Architecture Overview](00-architecture.md)** 📐

  - Hexagonal architecture
  - Layer dependencies
  - Design patterns

- **[Tech Stack](50-tech-stack.md)** ⚙️

  - Go libraries used
  - Rationale for choices
  - Alternatives considered

- **[Performance Benchmarks](60-benchmarks.md)** 📊

  - Benchmark results
  - Performance comparisons
  - Optimization notes

### Development Guides

- **[Contributing Guide](CONTRIBUTING.md)** 🤝

  - Code of conduct
  - Development workflow
  - PR process
  - Testing requirements

- **[API Documentation](API.md)** 📖

  - Public API (`pkg/cmd`)
  - Embedding Shellforge
  - Integration examples

______________________________________________________________________

## 🏗️ Project Structure

```
gzh-cli-shellforge/
├── cmd/shellforge/          # Main entry point
├── internal/                # Private application code
│   ├── domain/             # Business logic (NO external deps)
│   ├── app/                # Use cases (orchestration)
│   ├── infra/              # Infrastructure adapters
│   └── cli/                # CLI commands
├── pkg/cmd/                # Public API for embedding
├── examples/               # Example manifests
└── docs/                   # Documentation
    ├── user/              # User documentation
    └── dev/               # Developer documentation (you are here)
```

See [Architecture Document](00-architecture.md) for details.

______________________________________________________________________

## 🧪 Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
open coverage.html

# Run specific package
go test ./internal/domain -v

# Run specific test
go test ./internal/domain -run TestModule_AppliesTo -v

# Check for race conditions
go test -race ./...
```

**Current Coverage:** 70.1% (235 tests)

**Testing Strategy:**

- Domain layer: Pure logic tests (no mocks)
- App layer: Use `afero.MemMapFs` for filesystem mocking
- Infra layer: Integration tests with real implementations
- CLI layer: Command behavior tests

______________________________________________________________________

## 🛠️ Development Workflow

### Building

```bash
# Build binary
make build

# Install to $GOPATH/bin
make install

# Clean build artifacts
make clean
```

**IMPORTANT:** Always use `make build`, never `go build` directly. The Makefile configures the correct binary name (`gz-shellforge`) and build flags.

### Code Quality

```bash
# Format code
make fmt

# Lint
make lint

# Vet
go vet ./...
```

### Dependency Management

```bash
# Add dependency
go get github.com/example/package

# Update dependencies
go get -u ./...
go mod tidy

# Verify
go mod verify
```

______________________________________________________________________

## 📋 Common Development Tasks

### Adding a New CLI Command

1. Create `internal/cli/newcommand.go`
1. Follow pattern from `build.go` or `validate.go`
1. Add to root command in `internal/cli/root.go`
1. Create test file `internal/cli/newcommand_test.go`
1. Update user documentation

### Adding a New Use Case

1. Define interface in `internal/app/`
1. Implement service with injected dependencies
1. Add tests with mocked infrastructure
1. Wire up in `cmd/shellforge/main.go`

### Adding Infrastructure Adapter

1. Check if interface exists in `app/`
1. Create adapter in `infra/` subdirectory
1. Implement interface methods
1. Add tests (use in-memory or temp files)
1. Wire up in `main.go`

______________________________________________________________________

## 🎯 Architecture Rules (Critical)

### Dependency Flow

```
CLI Layer (internal/cli/)
    ↓ depends on
App Layer (internal/app/)
    ↓ depends on
Domain Layer (internal/domain/)
```

**Infrastructure Layer** (`internal/infra/`) implements interfaces defined by App Layer.

### Layer Constraints

1. **Domain Layer** - NO external dependencies

   - Pure Go business logic
   - NEVER import from app/infra/cli
   - Contains: Module, Manifest, Graph, Resolver

1. **App Layer** - Depends ONLY on domain interfaces

   - Defines interfaces for infrastructure
   - Orchestrates domain logic
   - Contains: BuilderService, ValidatorService

1. **Infrastructure Layer** - Implements app interfaces

   - Uses external libraries (yaml.v3, afero)
   - Contains: YAML parser, filesystem, git wrappers

1. **CLI Layer** - Depends on app services

   - Handles flag parsing and output
   - Injects services via constructors

______________________________________________________________________

## 📖 Documentation

- **User Documentation**: See [docs/user/](../user/)
- **Architecture Details**: See [00-architecture.md](00-architecture.md)
- **Tech Stack Rationale**: See [50-tech-stack.md](50-tech-stack.md)
- **Performance Data**: See [60-benchmarks.md](60-benchmarks.md)
- **API Reference**: See [API.md](API.md)

______________________________________________________________________

## 🤔 FAQ

### How do I run the CLI during development?

```bash
# Option 1: Build and run
make build
./build/gz-shellforge --help

# Option 2: Run directly with go run
go run cmd/shellforge/main.go --help
```

### What's the difference between cmd/shellforge and pkg/cmd?

- `cmd/shellforge/` - Main entry point for the standalone binary
- `pkg/cmd/` - Public API for embedding Shellforge in other Go applications

### Why use Hexagonal Architecture?

- **Testability**: Pure domain logic without external dependencies
- **Flexibility**: Easy to swap infrastructure implementations
- **Maintainability**: Clear boundaries between layers
- **Independence**: Business logic doesn't depend on frameworks

### What external dependencies are allowed?

- **Domain layer**: NONE (only standard library)
- **App layer**: NONE (only domain + standard library)
- **Infra layer**: yaml.v3, afero, os/exec
- **CLI layer**: cobra, pflag

### How do I update test coverage?

```bash
make test-coverage
open coverage.html
# Focus on domain and app layers (target >70%)
```

______________________________________________________________________

## 🗺️ Developer Documentation Roadmap

### ✅ Completed

1. Architecture Overview (00-architecture.md)
1. Tech Stack (50-tech-stack.md)
1. Performance Benchmarks (60-benchmarks.md)
1. Contributing Guide (CONTRIBUTING.md)
1. API Documentation (API.md)

### 📝 Future Enhancements

1. Development Setup Guide (detailed environment setup)
1. Testing Guide (comprehensive testing strategies)
1. Code Style Guide (Go-specific conventions)
1. Release Process Guide (versioning, changelog, deployment)

______________________________________________________________________

## 💬 Getting Help

- **Questions?** Open a [GitHub Discussion](https://github.com/gizzahub/gzh-cli-shellforge/discussions)
- **Bug?** File an [Issue](https://github.com/gizzahub/gzh-cli-shellforge/issues)
- **Contributing?** Read [CONTRIBUTING.md](CONTRIBUTING.md) first

______________________________________________________________________

**Last Updated:** 2025-12-01
**Maintained by:** Shellforge Contributors
