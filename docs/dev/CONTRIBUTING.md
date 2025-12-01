# Contributing to Shellforge

Thank you for your interest in contributing to Shellforge! This guide will help you get started.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Code Style](#code-style)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Project Structure](#project-structure)

---

## Code of Conduct

We are committed to providing a welcoming and inspiring community for all. Please be respectful and constructive in all interactions.

**Expected Behavior:**
- Be respectful of differing viewpoints and experiences
- Accept constructive criticism gracefully
- Focus on what is best for the community
- Show empathy towards other community members

**Unacceptable Behavior:**
- Harassment, trolling, or discriminatory language
- Personal attacks or insults
- Public or private harassment
- Publishing others' private information

---

## Getting Started

### Prerequisites

- **Go 1.21+** - [Download](https://golang.org/dl/)
- **Git** - For version control
- **Make** - For build automation (optional but recommended)
- **A GitHub account** - For pull requests

### Finding Work

Good places to start:

1. **Good First Issues**: Look for issues labeled `good-first-issue`
2. **Help Wanted**: Issues labeled `help-wanted` need contributors
3. **Documentation**: Improvements to docs are always welcome
4. **Bug Reports**: Try to reproduce and fix reported bugs

---

## Development Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/gzh-cli-shellforge.git
cd gzh-cli-shellforge

# Add upstream remote
git remote add upstream https://github.com/gizzahub/gzh-cli-shellforge.git
```

### 2. Install Dependencies

```bash
# Download Go modules
go mod download

# Verify dependencies
go mod verify
```

### 3. Build the Project

```bash
# Build using Make
make build

# Verify build
./build/gz-shellforge --version
```

### 4. Run Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
open coverage.html

# Run specific test
go test ./internal/domain -run TestModule_AppliesTo -v
```

---

## Development Workflow

### Creating a Feature Branch

```bash
# Update your fork
git checkout master
git pull upstream master

# Create a feature branch
git checkout -b feature/your-feature-name
```

**Branch Naming Convention:**
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test improvements

### Making Changes

1. **Write Code**
   - Follow the [code style](#code-style)
   - Add tests for new functionality
   - Update documentation as needed

2. **Test Your Changes**
   ```bash
   # Run tests
   make test

   # Check formatting
   go fmt ./...

   # Run linter
   go vet ./...
   ```

3. **Commit Your Changes**
   ```bash
   # Stage changes
   git add .

   # Commit with descriptive message
   git commit -m "feat(domain): add OS filtering to module resolver"
   ```

**Commit Message Format:**
```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Scope**: The area of code affected (e.g., `domain`, `cli`, `app`, `infra`)

**Examples:**
```
feat(cli): add list command for modules
fix(resolver): handle circular dependencies correctly
docs(readme): update installation instructions
test(builder): add test for OS filtering
```

### Keeping Your Fork Updated

```bash
# Fetch upstream changes
git fetch upstream

# Rebase your branch
git rebase upstream/master

# Force push to your fork (if already pushed)
git push --force-with-lease origin feature/your-feature-name
```

---

## Code Style

### Go Code Standards

#### Formatting

```bash
# Format all Go files
go fmt ./...

# Check with gofmt
gofmt -l .
```

#### Naming Conventions

```go
// Package comments
// Package domain contains the core business logic for Shellforge.
package domain

// Exported types use PascalCase
type BuilderService struct {
    manifestParser ManifestParser
}

// Unexported types use camelCase
type internalHelper struct {}

// Interface names are nouns
type ManifestParser interface {
    Parse(path string) (*Manifest, error)
}

// Function names are verbs
func NewBuilderService(...) *BuilderService {}
func (m *Module) AppliesTo(targetOS string) bool {}
```

#### Error Handling

```go
// Wrap errors with context
if err := parser.Parse(path); err != nil {
    return nil, fmt.Errorf("failed to parse manifest: %w", err)
}

// Use custom error types for domain errors
type ValidationError struct {
    Message string
}

func (e *ValidationError) Error() string {
    return e.Message
}
```

#### Comments

```go
// Exported functions require doc comments
// AppliesTo checks if this module should be loaded for the given OS.
// Empty OS field means the module applies to all operating systems.
func (m *Module) AppliesTo(targetOS string) bool {
    // Implementation
}

// Complex logic needs explanatory comments
// Use Kahn's algorithm for topological sort
// Complexity: O(V + E) where V = nodes, E = edges
```

### Project-Specific Patterns

#### Dependency Injection

```go
// main.go - wire up dependencies
fs := afero.NewOsFs()
parser := yamlparser.New()
reader := filesystem.NewReader(fs)
builder := app.NewBuilderService(parser, reader, writer)

// CLI layer - inject services
buildCmd := cli.NewBuildCmd(builder)
```

#### Table-Driven Tests

```go
func TestModule_AppliesTo(t *testing.T) {
    tests := []struct {
        name     string
        module   Module
        targetOS string
        want     bool
    }{
        {
            name:     "matches single OS",
            module:   Module{OS: []string{"Mac"}},
            targetOS: "Mac",
            want:     true,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.module.AppliesTo(tt.targetOS)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

---

## Testing

### Test Organization

```
internal/
├── domain/
│   ├── module.go
│   └── module_test.go       # Unit tests
├── app/
│   ├── builder.go
│   └── builder_test.go      # Service tests with mocks
└── infra/
    ├── yamlparser/
    │   ├── parser.go
    │   └── parser_test.go   # Integration-style tests
```

### Test Requirements

**For Pull Requests:**
- ✅ All existing tests must pass
- ✅ New features must include tests
- ✅ Bug fixes must include regression tests
- ✅ Code coverage should not decrease

**Coverage Targets:**
- Domain layer: >80%
- Application layer: >70%
- Infrastructure layer: >60%
- CLI layer: >50%

### Running Tests

```bash
# Run all tests
make test

# Run specific package
go test ./internal/domain -v

# Run specific test
go test ./internal/domain -run TestModule_AppliesTo -v

# With coverage
make test-coverage

# With race detection
go test -race ./...

# Verbose output
go test -v ./...
```

### Writing Good Tests

```go
// Good test - clear name, focused, uses table-driven pattern
func TestModule_AppliesTo_MatchesSingleOS(t *testing.T) {
    module := Module{OS: []string{"Mac"}}
    result := module.AppliesTo("Mac")
    assert.True(t, result)
}

// Better test - comprehensive coverage with table-driven
func TestModule_AppliesTo(t *testing.T) {
    tests := []struct {
        name     string
        module   Module
        targetOS string
        want     bool
    }{
        {
            name:     "matches single OS",
            module:   Module{OS: []string{"Mac"}},
            targetOS: "Mac",
            want:     true,
        },
        {
            name:     "case insensitive match",
            module:   Module{OS: []string{"Mac"}},
            targetOS: "mac",
            want:     true,
        },
        {
            name:     "no match",
            module:   Module{OS: []string{"Mac"}},
            targetOS: "Linux",
            want:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.module.AppliesTo(tt.targetOS)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

---

## Pull Request Process

### Before Submitting

**Checklist:**
- [ ] Code follows project style
- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] No lint errors (`go vet ./...`)
- [ ] Documentation is updated
- [ ] Commit messages follow convention
- [ ] Branch is rebased on latest master

### Submitting a PR

1. **Push to Your Fork**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create Pull Request**
   - Go to GitHub and create a PR from your fork
   - Use a clear, descriptive title
   - Fill out the PR template
   - Link related issues (e.g., "Closes #123")

3. **PR Title Format**
   ```
   <type>(<scope>): <description>

   Examples:
   feat(cli): add list command for modules
   fix(resolver): handle circular dependencies
   docs(readme): update installation guide
   ```

4. **PR Description Template**
   ```markdown
   ## Description
   Brief description of changes

   ## Motivation
   Why is this change needed?

   ## Changes
   - List of changes made
   - Another change

   ## Testing
   - How did you test this?
   - What test cases did you add?

   ## Related Issues
   Closes #123

   ## Screenshots (if applicable)
   ```

### Review Process

1. **Automated Checks**: CI/CD runs tests and linters
2. **Code Review**: Maintainers review your code
3. **Feedback**: Address any requested changes
4. **Approval**: Once approved, a maintainer will merge

**Responding to Feedback:**
- Be open to suggestions
- Ask questions if unclear
- Make requested changes
- Push updates to the same branch

---

## Project Structure

### Architecture

Shellforge follows **Hexagonal Architecture** (Ports & Adapters) with **Clean Architecture** principles.

```
internal/
├── domain/      # Pure business logic (NO external dependencies)
├── app/         # Use cases (depends on domain interfaces only)
├── infra/       # Infrastructure adapters (implements domain interfaces)
└── cli/         # CLI commands (depends on app layer)
```

### Dependency Rules

**Critical Rules:**

1. **Domain Layer** (`internal/domain/`)
   - NO imports from other internal packages
   - NO imports of external libraries (except stdlib)
   - Pure Go business logic only

2. **Application Layer** (`internal/app/`)
   - Depends ONLY on domain interfaces
   - Defines interfaces for infrastructure
   - Orchestrates domain logic

3. **Infrastructure Layer** (`internal/infra/`)
   - Implements interfaces defined by app layer
   - Uses external libraries (yaml.v3, afero)

4. **CLI Layer** (`internal/cli/`)
   - Depends on app layer services
   - Handles flag parsing and output

**Example Violation (DON'T DO THIS):**
```go
// domain/module.go
package domain

import "github.com/gizzahub/gzh-cli-shellforge/internal/app" // ❌ WRONG!
```

**Correct Approach:**
```go
// app/builder.go
package app

import "github.com/gizzahub/gzh-cli-shellforge/internal/domain" // ✅ CORRECT
```

### Adding New Features

When adding a new feature:

1. **Domain Layer**: Add entities and business logic
2. **Application Layer**: Add use case service
3. **Infrastructure Layer**: Add adapters if needed
4. **CLI Layer**: Add command

**Example: Adding a new command**

```bash
# 1. Create CLI command
internal/cli/newcommand.go

# 2. Add service (if needed)
internal/app/newservice.go

# 3. Add tests
internal/cli/newcommand_test.go
internal/app/newservice_test.go

# 4. Update root command
internal/cli/root.go
```

---

## Documentation

### When to Update Docs

**Always update docs when:**
- Adding new features
- Changing CLI interface
- Modifying configuration format
- Fixing bugs that affect user behavior

**Documentation Files:**
- `README.md` - Overview and quick start
- `QUICK_START.md` - 5-minute tutorial
- `FAQ.md` - Common questions
- `docs/user/` - User documentation
- `docs/dev/` - Developer documentation
- `CLAUDE.md` - AI development context

### Documentation Style

- Use clear, simple language
- Include code examples
- Provide real-world scenarios
- Keep sections focused and concise

---

## Getting Help

### Questions?

- **GitHub Discussions**: Ask questions
- **GitHub Issues**: Report bugs or request features
- **Code Comments**: Read inline documentation
- **Architecture Docs**: See `docs/dev/00-architecture.md`

### Resources

- **[Architecture Guide](00-architecture.md)** - System design
- **[Tech Stack](50-tech-stack.md)** - Library choices
- **[CLAUDE.md](../../CLAUDE.md)** - Development guidelines

---

## License

By contributing to Shellforge, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to Shellforge!** 🚀
