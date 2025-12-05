# CLAUDE.md

This file provides LLM-optimized guidance for Claude Code when working with this repository.

---

## Project Context

**Binary**: `gz-shellforge`
**Module**: `github.com/gizzahub/gzh-cli-shellforge`
**Go Version**: 1.25+
**Architecture**: Standard CLI (Cobra-based)

### Core Principles

- **Interface-driven design**: Use Go interfaces for abstraction
- **Direct constructors**: No DI containers, simple factory pattern
- **Shell safety**: Sanitize shell commands, prevent injection attacks
- **Dotfile management**: Safe handling of dotfiles and configuration files
- **Modular packages**: Separation of shell operations, parsing, and validation

---

## Shared Library (gzh-cli-core)

**IMPORTANT**: Use `gzh-cli-core` for common utilities. DO NOT create local duplicates.

| Package | Import | Purpose |
|---------|--------|---------|
| logger | `gzh-cli-core/logger` | Structured logging |
| testutil | `gzh-cli-core/testutil` | Test helpers (TempDir, Assert*, Capture) |
| errors | `gzh-cli-core/errors` | Error types and wrapping |
| config | `gzh-cli-core/config` | Config loading utilities |
| cli | `gzh-cli-core/cli` | CLI flags and output |
| version | `gzh-cli-core/version` | Version info |

```go
import (
    "github.com/gizzahub/gzh-cli-core/logger"
    "github.com/gizzahub/gzh-cli-core/errors"
    "github.com/gizzahub/gzh-cli-core/testutil"
)
```

---

## Module-Specific Guides (AGENTS.md)

**Read these before modifying code:**

| Guide | Location | Purpose |
|-------|----------|---------|
| Common Rules | `cmd/AGENTS_COMMON.md` | Project-wide conventions |
| CLI Module | `cmd/shellforge/AGENTS.md` | CLI-specific rules |

---

## Internal Packages

| Package | Purpose | Key Functions |
|---------|---------|---------------|
| `internal/app` | Application services | Builder, Service implementations |
| `internal/cli` | CLI commands | Command handlers |
| `internal/cli/errors` | CLI error handling | Error display |
| `internal/cli/output` | CLI output | Result formatting |
| `internal/domain` | Domain models | Module, Manifest, Graph |
| `internal/infra` | Infrastructure | File system, External services |

## Public Packages (pkg/)

| Package | Purpose |
|---------|---------|
| `pkg/dotfiles` | Dotfile management |
| `pkg/shell` | Shell configuration |
| `pkg/template` | Template engine |
| `pkg/backup` | Backup operations |
| `pkg/sync` | Synchronization |

---

## Development Workflow

### Before Code Modification

1. **Read AGENTS.md** for the module you're modifying
2. Check existing patterns in `internal/` and `pkg/`
3. Review CONTRIBUTING.md for guidelines

### Code Modification Process

```bash
# 1. Write code + tests
# 2. Quality checks (CRITICAL)
make quality    # runs fmt + lint + test

# Quick development cycle
make dev-fast   # format + unit tests only

# Pre-PR verification
make pr-check
```

---

## Essential Commands Reference

### Development Workflow

```bash
# One-time setup
make deps
make install-tools

# Before every commit (CRITICAL)
make quality

# Build & install
make build
make install

# Quick development
make dev-fast   # format + unit tests
make dev        # format + lint + test
```

### Testing

```bash
make test           # All tests
make test-unit      # Unit tests only
make test-coverage  # With coverage report
make bench          # Benchmarks
```

### Code Quality

```bash
make fmt            # Format code
make lint           # Run linters
make fmt-diff       # Format changed files only
make lint-diff      # Lint changed files only
```

---

## Project Structure

```
.
├── cmd/
│   └── shellforge/
│       ├── AGENTS.md           # Module-specific guide
│       └── main.go             # Entry point
├── internal/                    # Private packages
│   ├── app/                    # Application services
│   ├── cli/                    # CLI commands
│   │   ├── errors/             # CLI error handling
│   │   ├── output/             # CLI output formatting
│   │   └── factory/            # Command factory
│   ├── domain/                 # Domain models
│   ├── infra/                  # Infrastructure
│   └── integration/            # Integration helpers
├── data/                        # Data files
├── docs/                        # Documentation
├── examples/                    # Usage examples
├── .golangci.yml               # Linter config
├── CLAUDE.md                   # This file
├── go.mod                      # Go module
├── Makefile                    # Build automation
└── README.md                   # Project documentation
```

---

## Important Rules

### Critical Requirements

- **Read AGENTS.md** before modifying any module
- Always run `make quality` before commit
- Test coverage: 80%+ for core logic
- **Sanitize shell inputs** - prevent command injection
- **Safe file operations** - validate paths, check permissions

### Code Style

- **Binary name**: `gz-shellforge`
- **Interface-driven**: Use interfaces for testability
- **Error handling**: Use structured errors with context
- **Shell safety**: Always validate and sanitize user inputs
- **Dotfile handling**: Preserve original files, create backups

### Commit Format

```
{type}({scope}): {description}

{body}

Model: claude-{model}
Co-Authored-By: Claude <noreply@anthropic.com>
```

**Types**: feat, fix, docs, refactor, test, chore
**Scope**: REQUIRED (e.g., cmd, internal, pkg/dotfiles, pkg/shell)

---

## FAQ

**Q: Where to add new commands?**
A: `cmd/shellforge/` - create new command file

**Q: Where to add shell execution logic?**
A: `internal/shell/` - safe command execution

**Q: Where to add output parsing?**
A: `internal/parser/` - shell output parsing

**Q: Where to add public APIs?**
A: `pkg/{feature}/` directory

**Q: How to handle dotfiles safely?**
A: Use `pkg/dotfiles` - always create backups, validate paths

**Q: What files should AI not modify?**
A: See `.claudeignore`

---

**Last Updated**: 2025-12-05
