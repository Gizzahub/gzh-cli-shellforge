# Contributing to Shellforge

Thank you for your interest in contributing to Shellforge! This document provides guidelines and instructions for contributing.

---

## 🚀 Quick Start

### 1. Fork and Clone

```bash
# Fork repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/gzh-cli-shellforge.git
cd gzh-cli-shellforge

# Add upstream remote
git remote add upstream https://github.com/gizzahub/gzh-cli-shellforge.git
```

### 2. Setup Development Environment

```bash
# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Verify
./build/gz-shellforge --version
```

### 3. Create Feature Branch

```bash
git checkout -b feature/my-awesome-feature
```

### 4. Make Changes and Test

```bash
# Make your changes
vim internal/domain/module.go

# Run tests
make test

# Run linter
make lint
```

### 5. Commit and Push

```bash
# Commit with proper format
git commit -m "feat(domain): add awesome feature"

# Push to your fork
git push origin feature/my-awesome-feature
```

### 6. Create Pull Request

1. Go to your fork on GitHub
2. Click "New Pull Request"
3. Fill out the PR template
4. Submit for review

---

## 📋 Contribution Guidelines

### What We're Looking For

**High Priority:**
- 🐛 Bug fixes with test cases
- 📝 Documentation improvements
- 🧪 Test coverage improvements
- ⚡ Performance optimizations

**Welcome:**
- ✨ New features (discuss first!)
- 🎨 Code quality improvements
- 🌐 Platform-specific enhancements
- 🔧 Developer tooling

**Please Discuss First:**
- 🏗️ Major architectural changes
- 💥 Breaking API changes
- 🔄 Large refactoring

---

## 💻 Code Contributions

### Before You Start

1. **Check existing issues**: [GitHub Issues](https://github.com/gizzahub/gzh-cli-shellforge/issues)
2. **Discuss large changes**: Open an issue first for discussion
3. **Read architecture docs**: [docs/dev/00-architecture.md](docs/dev/00-architecture.md)
4. **Follow Go best practices**: See [Effective Go](https://golang.org/doc/effective_go.html)

### Development Workflow

```bash
# 1. Sync with upstream
git fetch upstream
git rebase upstream/master

# 2. Create branch
git checkout -b feature/my-feature

# 3. Make changes + write tests
# ...

# 4. Run full validation
make validate

# 5. Commit
git commit -m "feat(scope): description"

# 6. Push
git push origin feature/my-feature

# 7. Create PR
```

### Code Quality Requirements

**Before submitting PR:**
- ✅ All tests pass: `make test`
- ✅ Code is formatted: `go fmt ./...`
- ✅ No linter errors: `make lint`
- ✅ New code has tests
- ✅ Documentation updated (if needed)

**Test coverage targets:**
- Domain layer: >80%
- Application layer: >70%
- Infrastructure layer: >60%
- CLI layer: >50%

---

## 📝 Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/) with **mandatory scope**.

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Type

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, missing semicolons, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvement
- `test`: Adding or updating tests
- `chore`: Maintenance tasks (dependencies, build, etc.)

### Scope (Required)

Specifies which part of the codebase is affected:

- `domain`: Domain layer changes
- `app`: Application layer changes
- `infra`: Infrastructure layer changes
- `cli`: CLI layer changes
- `build`: Build system changes
- `docs`: Documentation changes
- `deps`: Dependency updates

### Examples

```bash
# Good commits
feat(domain): add OS filtering support for BSD
fix(cli): handle missing manifest file gracefully
docs(user): add troubleshooting guide for installation
test(domain): increase resolver test coverage to 95%
refactor(infra): simplify YAML parser error handling

# Bad commits (missing scope)
feat: add new feature  # ❌ Missing scope
fix bug                # ❌ Missing scope and description
update docs            # ❌ Missing scope and type
```

---

## 🧪 Testing Guidelines

### Writing Tests

**Location:**
- Place tests next to the code they test
- Use `_test.go` suffix
- Example: `module.go` → `module_test.go`

**Naming:**
```go
// Format: Test{Type}_{Method}_{Scenario}
func TestModule_AppliesTo_MacOS(t *testing.T) { ... }
func TestResolver_Resolve_CircularDependency(t *testing.T) { ... }
```

**Table-Driven Tests:**
```go
func TestModule_AppliesTo(t *testing.T) {
    tests := []struct {
        name     string
        module   Module
        targetOS string
        want     bool
    }{
        {
            name:     "empty OS applies to all",
            module:   Module{OS: []string{}},
            targetOS: "Mac",
            want:     true,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.module.AppliesTo(tt.targetOS)
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./internal/domain -v

# With coverage
make test-coverage
open coverage.html

# Run specific test
go test ./internal/domain -run TestModule_AppliesTo -v

# Benchmarks
make bench
```

---

## 📖 Documentation Contributions

### Types of Documentation

**User Documentation** (`docs/user/`):
- Installation guides
- Usage tutorials
- Troubleshooting
- FAQ

**Developer Documentation** (`docs/dev/`):
- Architecture
- Development setup
- Testing guide
- Contributing guide

**API Reference** (`docs/reference/`):
- Public API documentation
- Manifest schema
- Template reference

### Documentation Style

**User docs:**
- Friendly, encouraging tone
- Clear examples
- Step-by-step instructions
- Screenshots where helpful

**Developer docs:**
- Technical, precise
- Code examples
- Architecture diagrams
- Best practices

### Updating Documentation

```bash
# Edit documentation
vim docs/user/00-quick-start.md

# Preview (if markdown tool installed)
grip docs/user/00-quick-start.md

# Commit
git commit -m "docs(user): improve quick start guide"
```

---

## 🐛 Bug Reports

### Before Reporting

1. **Search existing issues**: [GitHub Issues](https://github.com/gizzahub/gzh-cli-shellforge/issues)
2. **Verify it's a bug**: Not expected behavior
3. **Gather information**: Version, OS, error messages

### Bug Report Template

```markdown
**Description:**
Clear description of the bug.

**To Reproduce:**
1. Step 1
2. Step 2
3. See error

**Expected Behavior:**
What should happen.

**Actual Behavior:**
What actually happens.

**Environment:**
- Shellforge version: `gz-shellforge --version`
- OS: macOS 14 / Ubuntu 22.04 / etc.
- Shell: zsh 5.9 / bash 5.0 / etc.
- Go version (if building): `go version`

**Additional Context:**
- Error messages
- Relevant configuration (manifest.yaml)
- Screenshots (if applicable)
```

---

## ✨ Feature Requests

### Before Requesting

1. **Check existing issues**: Might already be planned
2. **Consider scope**: Should it be in Shellforge?
3. **Think about implementation**: How would it work?

### Feature Request Template

```markdown
**Problem:**
What problem does this solve?

**Proposed Solution:**
How should it work?

**Alternatives Considered:**
Other ways to solve this?

**Use Case:**
Real-world scenario where this helps.

**Additional Context:**
Examples, mockups, etc.
```

---

## 🔍 Code Review Process

### What Reviewers Look For

**Code Quality:**
- ✅ Follows project architecture
- ✅ Tests included and passing
- ✅ No obvious bugs
- ✅ Error handling appropriate
- ✅ Documentation updated

**Style:**
- ✅ Consistent with codebase
- ✅ Clear variable names
- ✅ Appropriate comments
- ✅ No unnecessary complexity

**Design:**
- ✅ Follows SOLID principles
- ✅ Dependency rules respected
- ✅ No tight coupling
- ✅ Testable design

### Responding to Feedback

**Good responses:**
- Ask questions if unclear
- Explain your reasoning
- Make requested changes
- Update PR with fixes

**Not so good:**
- Ignore feedback
- Argue without reason
- Make unrelated changes
- Take feedback personally

### PR Tips

**Good PR:**
- ✅ Single focused change
- ✅ Clear description
- ✅ Tests included
- ✅ Small (<400 lines)
- ✅ Ready for review

**Needs work:**
- ❌ Multiple unrelated changes
- ❌ No description
- ❌ No tests
- ❌ Too large (>1000 lines)
- ❌ Work in progress

---

## 🏗️ Architecture Guidelines

### Dependency Rules (Critical)

```
internal/
├── domain/      # NO imports from app/infra/cli
├── app/         # Only imports domain interfaces
├── infra/       # Implements app interfaces
└── cli/         # Depends on app layer
```

**Rules:**
1. Domain layer: Pure Go, no external dependencies
2. App layer: Defines interfaces, orchestrates domain
3. Infra layer: Implements interfaces, uses external libraries
4. CLI layer: Wires everything together

See [Architecture Document](docs/dev/00-architecture.md) for details.

### Adding New Features

**Step-by-step:**
1. Define domain entity/logic (`internal/domain/`)
2. Define use case interface (`internal/app/`)
3. Implement use case
4. Create infrastructure adapter (`internal/infra/`)
5. Add CLI command (`internal/cli/`)
6. Write tests for each layer
7. Update documentation

---

## 🎯 First-Time Contributors

### Good First Issues

Look for issues labeled `good first issue`:
- Documentation improvements
- Test coverage additions
- Small bug fixes
- Code cleanup

### Getting Help

- **GitHub Discussions**: [Ask questions](https://github.com/gizzahub/gzh-cli-shellforge/discussions)
- **Issue Comments**: Ask on the issue you're working on
- **PR Comments**: Ask questions in your PR

### Learning Resources

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Project Documentation](docs/dev/)

---

## 📜 Code of Conduct

### Our Standards

**Be respectful:**
- Use welcoming and inclusive language
- Respect differing viewpoints
- Accept constructive criticism gracefully
- Focus on what's best for the community

**Be collaborative:**
- Help others learn
- Share knowledge
- Give credit where due
- Be patient with beginners

**Be professional:**
- Keep discussions technical
- Avoid personal attacks
- Don't harass or troll
- Follow project guidelines

### Reporting Issues

If you experience unacceptable behavior:
1. Contact project maintainers
2. Provide details of the incident
3. We will investigate and respond

---

## ⚖️ License

By contributing, you agree that your contributions will be licensed under the MIT License.

All contributions must be:
- Your own original work
- Not infringing on third-party rights
- Licensed under MIT License

---

## 🙏 Recognition

Contributors are recognized in:
- [README.md](README.md) contributors section
- Git commit history
- [CHANGELOG.md](CHANGELOG.md) for each release

**Thank you for contributing to Shellforge!** 🎉

---

## 📚 Additional Resources

- **[Developer Documentation](docs/dev/)** - Complete developer guide
- **[Architecture](docs/dev/00-architecture.md)** - System design
- **[Testing Guide](docs/dev/20-testing-guide.md)** - Testing best practices
- **[Code Style](docs/dev/40-code-style.md)** - Style guidelines

---

**Questions?** Open a [GitHub Discussion](https://github.com/gizzahub/gzh-cli-shellforge/discussions)

**Found a bug?** Open a [GitHub Issue](https://github.com/gizzahub/gzh-cli-shellforge/issues)

**Want to contribute?** Create a [Pull Request](https://github.com/gizzahub/gzh-cli-shellforge/pulls)
