# Shellforge

> Build tool for modular shell configurations with automatic dependency resolution

Transform your monolithic `.zshrc`/`.bashrc` into organized, maintainable modules with automatic dependency resolution and OS-specific filtering.

[![Tests](https://img.shields.io/badge/tests-285_passing-brightgreen)]()
[![Coverage](https://img.shields.io/badge/coverage-77.5%25-brightgreen)]()
[![Go Version](https://img.shields.io/badge/go-1.21%2B-blue)]()
[![License](https://img.shields.io/badge/license-MIT-blue)]()

**[Quick Start](docs/user/00-quick-start.md)** | **[Documentation](docs/user/)** | **[Examples](examples/)**

---

## What It Does

- ✅ **Automatic dependency resolution** - Modules load in correct order
- ✅ **OS-specific filtering** - Different configs for Mac/Linux
- ✅ **Validation** - Catch errors before deployment
- ✅ **Migration** - Convert existing configs automatically
- ✅ **Templates** - Generate common modules quickly
- ✅ **Backup/Restore** - Version control with git

---

## Quick Start

### Install

```bash
# Install via Go
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest

# Or build from source
make install
```

### Migrate Your Config (5 minutes)

```bash
# 1. Backup your current config
gz-shellforge backup --file ~/.zshrc

# 2. Migrate to modular structure
mkdir ~/shellforge && cd ~/shellforge
gz-shellforge migrate ~/.zshrc

# 3. Build for your OS
gz-shellforge build --os Mac --output ~/.zshrc.new

# 4. Test and deploy
gz-shellforge diff ~/.zshrc ~/.zshrc.new
mv ~/.zshrc.new ~/.zshrc
```

**[Complete tutorial →](docs/user/00-quick-start.md)**

---

## Core Features

### Dependency Resolution
Automatically sorts modules using topological sort (Kahn's algorithm). No more manual ordering or "works by accident" configs.

### OS Filtering
Write once, deploy everywhere. Modules tagged with `os: [Mac]` only load on macOS, `os: [Linux]` only on Linux.

### Migration Tools
Convert your existing monolithic `.zshrc` to organized modules automatically. Detects sections, infers dependencies, and categorizes content.

### Template System
Generate common modules from 6 built-in templates: PATH, environment variables, aliases, tool initialization, and more.

### Backup & Restore
Git-backed versioning with timestamped snapshots. Rollback to any previous configuration instantly.

### Diff Comparison
Compare original and generated configs with 4 output formats: summary, unified, context, side-by-side.

---

## Example: Module Structure

```yaml
# manifest.yaml
modules:
  - name: os-detection
    file: init.d/00-os-detection.sh
    requires: []
    os: [Mac, Linux]

  - name: brew-path
    file: init.d/05-brew-path.sh
    requires: [os-detection]
    os: [Mac]

  - name: nvm
    file: rc_pre.d/nvm.sh
    requires: [brew-path]
    os: [Mac, Linux]
```

```
modules/
├── init.d/             # Early initialization
├── rc_pre.d/           # Tool setup (nvm, rbenv)
└── rc_post.d/          # Aliases, functions
```

**[See complete examples →](examples/)**

---

## Commands

```bash
# Validate configuration
gz-shellforge validate

# Build for specific OS
gz-shellforge build --os Mac --output ~/.zshrc

# List modules with filtering
gz-shellforge list --filter Mac

# Migrate existing config
gz-shellforge migrate ~/.zshrc

# Generate from template
gz-shellforge template generate alias my-aliases

# Backup current config
gz-shellforge backup --file ~/.zshrc

# Restore from snapshot
gz-shellforge restore --file ~/.zshrc --snapshot 2025-11-28_14-30-45

# Compare configs
gz-shellforge diff ~/.zshrc ~/.zshrc.new
```

**[Full command reference →](docs/user/40-command-reference.md)**

---

## Documentation

### For Users
- **[Quick Start Guide](docs/user/00-quick-start.md)** - Get started in 5 minutes
- **[Complete Workflows](docs/user/30-workflows.md)** - Step-by-step guide
- **[Command Reference](docs/user/40-command-reference.md)** - All commands with examples
- **[User Documentation](docs/user/)** - Complete user guide

### For Developers
- **[Architecture](docs/dev/00-architecture.md)** - System design (Hexagonal)
- **[Tech Stack](docs/dev/50-tech-stack.md)** - Technology decisions
- **[Contributing Guide](docs/dev/CONTRIBUTING.md)** - How to contribute
- **[Developer Documentation](docs/dev/)** - Complete developer guide

---

## Performance

Go implementation is significantly faster than Python version:

| Metric | Python | Go | Improvement |
|--------|--------|----|----|
| Startup | ~200ms | <10ms | **20x faster** |
| Build (10 modules) | ~300ms | <50ms | **6x faster** |
| Memory | ~80MB | <10MB | **8x lighter** |
| Binary size | ~40MB | ~8MB | **5x smaller** |

---

## Why Shellforge?

### Problem
- Manual shell script concatenation causes ordering errors
- No dependency tracking leads to tools loading before dependencies
- OS-specific logic scattered everywhere, hard to maintain
- No validation until runtime errors occur

### Solution
- Automatic dependency resolution with topological sort
- OS-specific module filtering (Mac/Linux)
- Pre-deployment validation catches errors early
- Modular structure makes maintenance easy
- Version control friendly for team collaboration

---

## Status

**Version**: 0.5.1
**Test Coverage**: 70.1% (235 tests passing)
**Status**: Active development

### Implemented Features
- ✅ Build, validate, list commands
- ✅ Migration from monolithic configs
- ✅ Template generation (6 types)
- ✅ Backup/restore with git versioning
- ✅ Diff comparison (4 formats)
- ✅ OS filtering (Mac/Linux)

### Planned Features
- ⏳ Plugin system for custom validators
- ⏳ BSD support (FreeBSD 13+)
- ⏳ Fish shell support

---

## Installation

### Prerequisites
- Go 1.21 or later
- Git (for backup/restore features)

### From Source

```bash
git clone https://github.com/gizzahub/gzh-cli-shellforge.git
cd gzh-cli-shellforge
make install
```

### Verify Installation

```bash
gz-shellforge --version
# Output: shellforge version 0.5.1
```

---

## Contributing

Contributions welcome! Please:

1. Read the [Contributing Guide](CONTRIBUTING.md)
2. Fork the repository
3. Create a feature branch
4. Write tests for new features
5. Ensure all tests pass: `make test`
6. Submit a pull request

See [Development Guide](.claude/DEVELOPMENT.md) for architecture details.

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

## Support

- **Documentation**: [docs/user/](docs/user/)
- **Issues**: [GitHub Issues](https://github.com/gizzahub/gzh-cli-shellforge/issues)
- **Examples**: [examples/](examples/)
- **Discussions**: [GitHub Discussions](https://github.com/gizzahub/gzh-cli-shellforge/discussions)

---

**Built with Go, designed for shell power users** 🚀
