# CLAUDE.md

LLM-optimized guidance for working with this repository.

---

## Project Context

**Binary**: `gz-shellforge`
**Module**: `github.com/gizzahub/gzh-cli-shellforge`
**Go Version**: 1.25+
**Architecture**: Standard CLI (Cobra-based)

### Core Principles

- **Interface-driven design** with Go interfaces
- **Shell safety**: Sanitize commands, prevent injection
- **Dotfile management**: Safe handling, always backup
- **Modular packages**: Separation of concerns

---

## Shared Library (gzh-cli-core)

**CRITICAL**: Use `gzh-cli-core` for common utilities. DO NOT duplicate locally.

| Package | Purpose |
|---------|---------|
| logger | Structured logging |
| testutil | Test helpers (TempDir, Assert*, Capture) |
| errors | Error types and wrapping |
| config | Config loading |
| cli | CLI flags and output |
| version | Version info |

```go
import (
    "github.com/gizzahub/gzh-cli-core/logger"
    "github.com/gizzahub/gzh-cli-core/errors"
    "github.com/gizzahub/gzh-cli-core/testutil"
)
```

---

## Package Architecture

### Internal Packages

| Package | Purpose |
|---------|---------|
| `internal/app` | Application services (Builder, DeployService) |
| `internal/cli` | Command handlers (build, deploy, validate, etc.) |
| `internal/cli/errors` | CLI error handling |
| `internal/cli/output` | Result formatting |
| `internal/domain` | Module, Manifest, Graph |
| `internal/infra` | File system, External services |

### Public Packages (pkg/)

| Package | Purpose |
|---------|---------|
| `pkg/dotfiles` | Dotfile management |
| `pkg/shell` | Shell configuration |
| `pkg/template` | Template engine |
| `pkg/backup` | Backup operations |
| `pkg/sync` | Synchronization |

---

## Development Workflow

### Quick Reference

```bash
# Before every commit (CRITICAL)
make quality    # fmt + lint + test

# Development cycle
make dev-fast   # format + unit tests only
make dev        # format + lint + test

# Build & test
make build
make test
make test-coverage
```

### Critical Requirements

- Always run `make quality` before commit
- Test coverage: 80%+ for core logic
- **Sanitize shell inputs** - prevent command injection
- **Safe file operations** - validate paths, check permissions

---

## CLI Commands

### Build & Deploy Workflow

```bash
# Build configuration (OS auto-detected, outputs to ./build/)
gz-shellforge build

# Preview build output
gz-shellforge build --dry-run

# Deploy to home directory with backup
gz-shellforge deploy --backup

# Full workflow
gz-shellforge build && gz-shellforge deploy --dry-run
gz-shellforge build && gz-shellforge deploy --backup
```

### Key Commands

| Command | Purpose |
|---------|---------|
| `build` | Generate shell config from modules (OS auto-detected, default: ./build/) |
| `deploy` | Copy built files to actual paths (~/.zshrc, etc.) |
| `validate` | Check manifest and module files |
| `list` | List available modules |
| `backup` | Backup current shell configurations |
| `restore` | Restore from backup |

### Build Options

- `--os` - Target OS (auto-detected if omitted)
- `--output-dir` - Output directory (default: `./build/`)
- `--shell` - Shell type (zsh, bash, fish)
- `--target` - Specific targets to build (zshrc, zprofile, etc.)
- `--dry-run` - Preview without writing files

### Deploy Options

- `--build-dir` - Build directory (default: `./build/`)
- `--backup` - Backup existing files before overwriting
- `--dry-run` - Preview without deploying

---

## Project Structure

```
.
├── cmd/shellforge/          # Entry point
├── internal/                # Private packages
│   ├── app/                 # Services
│   ├── cli/                 # Commands
│   ├── domain/              # Models
│   └── infra/               # Infrastructure
├── pkg/                     # Public APIs
├── data/                    # Data files
├── examples/                # Usage examples
└── docs/                    # Documentation
```

---

## Code Style

- **Binary name**: `gz-shellforge`
- **Interface-driven**: Use interfaces for testability
- **Error handling**: Use structured errors with context
- **Shell safety**: Always validate and sanitize user inputs
- **Dotfile handling**: Preserve originals, create backups

### Commit Format

```
{type}({scope}): {description}

Model: claude-{model}
Co-Authored-By: Claude <noreply@anthropic.com>
```

**Types**: feat, fix, docs, refactor, test, chore
**Scope**: REQUIRED (cmd, internal, pkg/dotfiles, pkg/shell)

---

## Quick FAQ

**Where to add new commands?** → `cmd/shellforge/`
**Where to add shell logic?** → `internal/shell/`
**Where to add public APIs?** → `pkg/{feature}/`
**How to handle dotfiles?** → Use `pkg/dotfiles` with backups
**AI-prohibited files?** → See `.claudeignore`

---

**Last Updated**: 2026-01-19 (Breaking change: removed legacy --single-output mode)
