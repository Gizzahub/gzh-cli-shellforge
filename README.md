# gzh-cli-shellforge

> Build tool for modular shell configurations with automatic dependency resolution

A Go implementation of Shellforge - transform modular shell scripts into unified `.zshrc`/`.bashrc` with dependency resolution and OS-specific filtering.

---

## What It Does

- **Reads shell modules** from your config directory
- **Resolves dependencies** automatically via topological sort
- **Filters by OS** (macOS/Linux) - include/exclude modules per platform
- **Generates optimized** `.zshrc`, `.bashrc`, or `.config/fish/config.fish`
- **Validates configuration** integrity before deployment
- **Backs up automatically** with git version control and timestamped snapshots

---

## Quick Start

### Installation

```bash
# From source (requires Go 1.21+)
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest

# Or clone and build locally
git clone https://github.com/gizzahub/gzh-cli-shellforge.git
cd gzh-cli-shellforge
make build
```

### Basic Usage

```bash
# 1. Auto-generate manifest from existing config structure
shellforge init -c config/shellrc -o manifest.yaml

# 2. Validate configuration
shellforge validate -c config/shellrc -m manifest.yaml

# 3. Build shell config
shellforge build -c config/shellrc -m manifest.yaml --auto-output

# 4. Deploy with automatic backup
shellforge build -c config/shellrc -m manifest.yaml --auto-output --deploy
```

### Example Manifest

```yaml
# manifest.yaml
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

## Features

### Core Features

- **Automatic Dependency Resolution**: Topological sort ensures correct module load order
- **OS Filtering**: Include/exclude modules based on target OS (Mac, Linux, BSD)
- **Shell Support**: Build configs for bash, zsh, fish
- **Validation**: Detect circular dependencies and missing files before deployment
- **Migration Tools**: Convert monolithic `.zshrc` to modular structure
- **Backup & Restore**: Git-backed versioning with timestamped snapshots
- **Template Generation**: Create common modules (PATH, env vars, aliases) from templates

### Advanced Features

- **Shell Metadata System**: Built-in knowledge of shell config files (where to deploy per OS/shell/session)
- **Diff Comparison**: Preview changes before deployment (summary/unified/context formats)
- **Auto-Init**: Generate manifest from existing `init.d/`/`rc_pre.d/`/`rc_post.d/` structure
- **Self-Documenting CLI**: Contextual help and next-step suggestions
- **Verbose Mode**: Detailed output for debugging dependency resolution

---

## Commands

### Core Operations

```bash
# Build shell configuration
shellforge build -c config -m manifest.yaml --auto-output

# Validate manifest and modules
shellforge validate -c config -m manifest.yaml

# Generate manifest from existing config
shellforge init -c config -o manifest.yaml
```

### Migration & Comparison

```bash
# Convert monolithic .zshrc to modular structure
shellforge migrate -s ~/.zshrc -t config/shellrc

# Compare generated vs existing config
shellforge diff -c config -m manifest.yaml --auto-detect-existing
```

### Backup & Restore

```bash
# Deploy with automatic backup
shellforge build -c config -m manifest.yaml --deploy

# List available snapshots
shellforge restore -t ~/.zshrc --list

# Restore previous version
shellforge restore -t ~/.zshrc

# Clean old snapshots
shellforge clean-snapshots -t ~/.zshrc --keep-count 10
```

### Templates

```bash
# List available templates
shellforge template list

# Generate PATH module
shellforge template generate path my-bin \
  -f path_dir=/usr/local/mybin

# Generate environment variable
shellforge template generate env EDITOR \
  -f var_name=EDITOR -f var_value=vim

# Generate with dependencies
shellforge template generate tool-init nvm \
  -r brew-path -r os-detection
```

### Information

```bash
# List modules in load order
shellforge list-modules -c config -m manifest.yaml

# Show shell config metadata
shellforge info --os macos --shell zsh --session interactive
```

---

## Project Structure

### Recommended Layout

```
your-dotfiles/
├── shellrc/
│   ├── init.d/              # Initialization (PATH setup, env detection)
│   │   ├── 00-os-detection.sh
│   │   └── 05-brew-path.sh
│   ├── rc_pre.d/            # Pre-configuration (tool setup)
│   │   ├── conda.sh
│   │   └── nvm.sh
│   └── rc_post.d/           # Post-configuration (aliases, functions)
│       └── aliases.sh
└── manifest.yaml
```

### Module Categories

- **init.d/**: Early initialization (PATH, OS detection, environment setup)
- **rc_pre.d/**: Pre-configuration (tool initialization - nvm, rbenv, conda, etc.)
- **rc_post.d/**: Post-configuration (aliases, functions, customizations)

---

## Why Shellforge?

### Problems It Solves

**Before Shellforge**:
- Manual concatenation of shell scripts → ordering errors
- No dependency tracking → tools load before dependencies
- OS-specific logic scattered → hard to maintain
- No validation → discover errors after deployment
- Fear of change → avoid updating shell config

**With Shellforge**:
- ✅ Automatic dependency resolution
- ✅ OS-specific filtering
- ✅ Pre-deployment validation
- ✅ Safe rollback with backups
- ✅ Modular, maintainable configuration

### Benefits

- **Multi-Machine**: Maintain different configs per OS from single source
- **Version Control**: Git-friendly modular structure
- **Safe Deployment**: Automatic backups before changes
- **Easy Migration**: Convert existing `.zshrc` to modular structure
- **Team Sharing**: Share individual modules without full config

---

## Documentation

- **[PRD.md](docs/PRD.md)**: Product requirements and feature specifications
- **[REQUIREMENTS.md](docs/REQUIREMENTS.md)**: Detailed functional requirements
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)**: System architecture and design
- **[TECH_STACK.md](docs/TECH_STACK.md)**: Technology choices and rationale
- **[examples/](examples/)**: Example manifests and modules

---

## Development

### Prerequisites

- Go 1.21 or later
- Git (for backup features)
- Make (optional, for build automation)

### Build

```bash
# Install dependencies
go mod download

# Build binary
go build -o shellforge cmd/shellforge/main.go

# Or use Makefile
make build
```

### Test

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Verbose output
go test -v ./...

# Or use Makefile
make test
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint
golangci-lint run

# Vet
go vet ./...
```

### Project Structure

```
gzh-cli-shellforge/
├── cmd/shellforge/          # Main application entry point
├── internal/
│   ├── domain/              # Business logic (pure Go)
│   ├── app/                 # Use cases
│   ├── infra/               # Infrastructure (YAML, filesystem, git)
│   └── cli/                 # CLI commands (Cobra)
├── data/                    # Embedded shell metadata
├── examples/                # Example configs
├── docs/                    # Documentation
└── Makefile
```

---

## Performance

Shellforge (Go) is significantly faster than the Python version:

| Metric | Python | Go | Improvement |
|--------|--------|----|----|
| Startup time | ~200ms | <50ms | **4x faster** |
| Build (50 modules) | ~800ms | <500ms | **1.6x faster** |
| Memory usage | ~80MB | <50MB | **1.6x lighter** |
| Binary size | ~40MB (venv) | <10MB | **4x smaller** |

---

## Compatibility

### Python Version Compatibility

- ✅ **Manifest format**: 100% compatible with Python version
- ✅ **CLI interface**: Identical commands and flags
- ✅ **Generated output**: Functionally equivalent shell configs
- ✅ **Migration path**: Drop-in replacement for Python version

### Supported Platforms

- **macOS**: 10.15+ (Catalina and later)
- **Linux**: Ubuntu 20.04+, Debian 11+, Arch, Manjaro
- **BSD**: FreeBSD 13+ (planned)

### Supported Shells

- **Zsh**: 5.8+ (default on macOS)
- **Bash**: 4.0+ (ubiquitous)
- **Fish**: 3.0+ (modern shell)

---

## Integration

### With Chezmoi

Shellforge works seamlessly with [Chezmoi](https://www.chezmoi.io/) for multi-machine dotfile management:

```bash
# Build config
shellforge build -c config -m manifest.yaml --auto-output

# Add to Chezmoi
chezmoi add ~/.zshrc

# Apply to other machines
chezmoi apply
```

### With Git

```bash
# Initialize dotfiles repo
git init ~/dotfiles
cd ~/dotfiles

# Generate and track manifest
shellforge init -c shellrc -o manifest.yaml
git add manifest.yaml shellrc/

# Build and test
shellforge build -c shellrc -m manifest.yaml --auto-output --dry-run
```

---

## FAQ

**Q: Do I need to uninstall the Python version?**
A: No, they can coexist. The Go version uses the same manifest format.

**Q: Can I migrate from the Python version?**
A: Yes, just replace the binary. Manifest files are 100% compatible.

**Q: Does it work on Windows?**
A: Not yet. Windows PowerShell has a different paradigm. Use WSL for now.

**Q: What if git is not installed?**
A: Backup features require git. Other features work without it.

**Q: Can I use it without a manifest?**
A: No, the manifest is required to define modules and dependencies.

---

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for new features
4. Ensure tests pass (`make test`)
5. Format code (`go fmt ./...`)
6. Commit with clear message
7. Push to branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

---

## Related Projects

- **[gzh-cli-shellforge-py](https://github.com/Gizzahub/gzh-cli-shellforge)**: Original Python version
- **[Chezmoi](https://www.chezmoi.io/)**: Dotfile manager (complementary)
- **[GNU Stow](https://www.gnu.org/software/stow/)**: Symlink farm manager
- **[dotbot](https://github.com/anishathalye/dotbot)**: Dotfile bootstrap tool

---

## License

MIT License - see [LICENSE](LICENSE) file for details

---

## Part of gzh-cli Series

Other tools in the gzh-cli series:
- `gzh-cli-shellforge` - This project
- More coming soon...

---

## Support

- **Issues**: [GitHub Issues](https://github.com/gizzahub/gzh-cli-shellforge/issues)
- **Discussions**: [GitHub Discussions](https://github.com/gizzahub/gzh-cli-shellforge/discussions)
- **Documentation**: See [docs/](docs/) directory

---

**Status**: Under development (Go reimplementation)
**Stability**: Alpha (not production-ready yet)
**Compatibility**: Manifest-compatible with Python version

