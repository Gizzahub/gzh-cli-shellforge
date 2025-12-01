# Shellforge

> Build tool for modular shell configurations with automatic dependency resolution

A Go implementation of Shellforge - transform modular shell scripts into unified `.zshrc`/`.bashrc` with dependency resolution and OS-specific filtering.

[![Tests](https://img.shields.io/badge/tests-50%2F50_passing-brightgreen)]()
[![Coverage](https://img.shields.io/badge/coverage-71%25--100%25-brightgreen)]()
[![Go Version](https://img.shields.io/badge/go-1.21%2B-blue)]()
[![License](https://img.shields.io/badge/license-MIT-blue)]()

**[Quick Start](QUICK_START.md)** | **[Documentation](docs/user/)** | **[FAQ](FAQ.md)** | **[Examples](examples/)**

---

## What It Does

- **Reads shell modules** from your config directory
- **Resolves dependencies** automatically via topological sort
- **Filters by OS** (macOS/Linux) - include/exclude modules per platform
- **Validates configuration** before building (circular dependencies, missing files)
- **Generates unified** `.zshrc`, `.bashrc`, or custom shell config

## Why Shellforge?

### Before
- ❌ Manual concatenation of shell scripts → ordering errors
- ❌ No dependency tracking → tools load before dependencies
- ❌ OS-specific logic scattered → hard to maintain
- ❌ No validation → discover errors after deployment

### After
- ✅ Automatic dependency resolution
- ✅ OS-specific filtering
- ✅ Pre-deployment validation
- ✅ Modular, maintainable configuration
- ✅ Version control friendly

---

## Quick Start

### Install

```bash
# From source (requires Go 1.21+)
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest

# Or build locally
git clone https://github.com/gizzahub/gzh-cli-shellforge.git
cd gzh-cli-shellforge
make install
```

### 5-Minute Tutorial

```bash
# 1. Backup your current config
gz-shellforge backup --file ~/.zshrc

# 2. Convert to modular structure
gz-shellforge migrate ~/.zshrc

# 3. Validate
gz-shellforge validate

# 4. Build for your OS
gz-shellforge build --os Mac --output ~/.zshrc.new

# 5. Compare and deploy
gz-shellforge diff ~/.zshrc ~/.zshrc.new
mv ~/.zshrc.new ~/.zshrc && source ~/.zshrc
```

**[→ Complete Quick Start Guide](QUICK_START.md)**

---

## Core Features

### ✅ Implemented

- **Automatic Dependency Resolution**: Topological sort ensures correct module load order
- **OS Filtering**: Include/exclude modules based on target OS (Mac, Linux)
- **Validation**: Detect circular dependencies and missing files before building
- **Migration Tools**: Convert monolithic RC files to modular structure
- **Template Generation**: Create modules from 6 built-in templates
- **Backup/Restore**: Git-backed versioning with timestamped snapshots
- **Diff Comparison**: 4 output formats (summary, unified, context, side-by-side)

### ⏳ Planned

- **Plugin System**: Extensible module types and custom validators

---

## Commands

### Essential Commands

```bash
# Build shell configuration
gz-shellforge build --manifest manifest.yaml --os Mac --output ~/.zshrc

# Validate manifest file
gz-shellforge validate --manifest manifest.yaml

# List all modules (with optional OS filter)
gz-shellforge list --filter Mac

# Convert monolithic RC to modular structure
gz-shellforge migrate ~/.zshrc --output-dir modules

# Compare configurations
gz-shellforge diff ~/.zshrc ~/.zshrc.new
```

### Backup & Restore

```bash
# Create backup snapshot
gz-shellforge backup --file ~/.zshrc --message "Before changes"

# Restore from snapshot
gz-shellforge restore --file ~/.zshrc --snapshot 2025-11-27_14-30-45

# Cleanup old snapshots
gz-shellforge cleanup --file ~/.zshrc --keep-count 10 --keep-days 30
```

### Template System

```bash
# List available templates
gz-shellforge template list

# Generate module from template
gz-shellforge template generate path my-bin -f path_dir=/usr/local/bin
gz-shellforge template generate alias my-aliases -f aliases='alias ll="ls -la"'
```

**[→ Complete Command Reference](docs/user/20-commands.md)**

---

## Example Manifest

```yaml
modules:
  - name: os-detection
    file: init.d/00-os-detection.sh
    requires: []
    os: [Mac, Linux]
    description: Detect operating system

  - name: brew-path
    file: init.d/05-brew-path.sh
    requires: [os-detection]
    os: [Mac]
    description: Homebrew PATH initialization

  - name: nvm
    file: rc_pre.d/nvm.sh
    requires: [brew-path]
    os: [Mac, Linux]
    description: Node Version Manager
```

---

## Documentation

### For Users

- **[Quick Start](QUICK_START.md)** - Get started in 5 minutes
- **[FAQ](FAQ.md)** - Frequently asked questions
- **[Command Reference](docs/user/20-commands.md)** - All commands and options
- **[Workflow Guide](docs/user/30-workflows.md)** - Step-by-step workflows
- **[Examples](docs/user/40-examples.md)** - Real-world usage examples

### For Developers

- **[Architecture](ARCHITECTURE.md)** - System design (Hexagonal + Clean Architecture)
- **[Tech Stack](TECH_STACK.md)** - Technology choices and rationale
- **[Contributing](docs/dev/CONTRIBUTING.md)** - How to contribute
- **[API Reference](docs/dev/API.md)** - Public API documentation

---

## Examples

### Try the Demo

```bash
cd examples/
./workflow-demo.sh  # Automated demonstration of complete workflow
```

This shows the entire migrate → build → diff workflow for both Mac and Linux.

### Example Configuration

The `examples/` directory contains:
- `manifest.yaml` - Example manifest with 10 modules
- `modules/` - Shell modules organized by category (init.d/, rc_pre.d/, rc_post.d/)
- `sample.zshrc` - Sample RC file for testing migration
- `WORKFLOW.md` - Complete workflow guide

---

## Performance

Shellforge (Go) is significantly faster than the Python version:

| Metric | Python | Go | Improvement |
|--------|--------|----|----|
| Startup time | ~200ms | <10ms | **20x faster** |
| Build (10 modules) | ~300ms | <50ms | **6x faster** |
| Memory usage | ~80MB | <10MB | **8x lighter** |
| Binary size | ~40MB (venv) | ~8MB | **5x smaller** |

---

## Platform Support

- ✅ **macOS**: 10.15+ (Catalina and later)
- ✅ **Linux**: Ubuntu 20.04+, Debian 11+, Arch, Manjaro
- ✅ **Shells**: Zsh 5.8+, Bash 4.0+, Fish 3.0+
- ⏳ **BSD**: FreeBSD 13+ (planned)
- ❌ **Windows**: Not supported (use WSL)

---

## Testing

```bash
# Run all tests
make test

# With coverage report
make test-coverage
open coverage.html

# Run benchmarks
make bench
```

**Test Coverage:**
- Domain layer: 88.7%
- Infrastructure layer: 77.6-100%
- Application layer: 86.5%
- **Overall: 70.1%** (235 tests passing)

---

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Write tests for new features
4. Ensure tests pass (`make test`)
5. Format code (`go fmt ./...`)
6. Open a Pull Request

See **[CONTRIBUTING.md](docs/dev/CONTRIBUTING.md)** for detailed guidelines.

---

## Status

- **Version**: 0.2.0-alpha
- **Development**: Active
- **Stability**: Alpha (core features stable, API may change)
- **Production Ready**: Core build/validate features ready for use

**Recent Releases:**
- v0.5.0: Diff comparison with 4 output formats
- v0.4.0: Migration tools with auto-categorization
- v0.3.0: Template generation system
- v0.2.1: Restore and cleanup commands
- v0.2.0: List and backup commands
- v0.1.0: Initial release with build/validate

---

## License

MIT License - see LICENSE file for details

---

## Support

- **Issues**: [GitHub Issues](https://github.com/gizzahub/gzh-cli-shellforge/issues)
- **Discussions**: [GitHub Discussions](https://github.com/gizzahub/gzh-cli-shellforge/discussions)
- **Documentation**: [docs/](docs/)

---

**Made with ❤️ for better shell config management**
