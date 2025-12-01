# Product Requirements Document: Shellforge

**Version**: 1.0
**Status**: Draft
**Last Updated**: 2025-11-27
**Author**: CEE (@archmagece)

---

## Executive Summary

Shellforge is a build tool that transforms modular shell scripts into unified shell configurations (`.zshrc`, `.bashrc`, `.config/fish/config.fish`) with automatic dependency resolution and OS-specific filtering. This Go reimplementation aims to provide the same functionality as the Python version while following Go idioms and improving performance.

---

## Problem Statement

### Current Challenges

Managing shell configurations across multiple machines and operating systems is error-prone and time-consuming:

1. **Manual Concatenation**: Developers manually merge shell scripts, leading to ordering errors
2. **Dependency Hell**: No automatic resolution of module dependencies (e.g., nvm requires brew-path on macOS)
3. **OS Fragmentation**: OS-specific logic scattered across monolithic RC files
4. **No Validation**: Syntax errors discovered only after deployment
5. **Difficult Migration**: Moving from monolithic `.zshrc` to modular structure is manual work
6. **No Version Control**: Changes to shell configs lack proper backup/restore mechanisms

### Impact

- **Time waste**: Hours spent debugging shell startup errors
- **Machine-specific hacks**: Different configurations per machine, hard to sync
- **Fear of change**: Developers avoid updating shell configs due to risk
- **Lost productivity**: Shell misconfigurations block development work

---

## Solution Overview

Shellforge solves these problems by providing:

1. **Declarative Manifest**: Define modules and dependencies in YAML
2. **Automatic Resolution**: Topological sort ensures correct load order
3. **OS Filtering**: Include/exclude modules based on target OS
4. **Validation**: Check manifest integrity before deployment
5. **Migration Tools**: Convert existing monolithic RC files to modular structure
6. **Backup/Restore**: Git-backed versioning with timestamped snapshots
7. **Template System**: Generate common module patterns quickly

### How It Works

```
YAML Manifest → Dependency Graph → Topological Sort → OS Filter → Concatenate → Deploy
```

---

## Target Users

### Primary Audience

1. **Software Developers**: Manage dotfiles across work laptop, personal machines, servers
2. **DevOps Engineers**: Maintain consistent shell environments across teams
3. **Power Users**: Heavy shell users with complex configurations (aliases, functions, tool initialization)

### User Personas

**Persona 1: Multi-Machine Developer**
- Works on macOS laptop, Linux servers, Arch home desktop
- Uses Homebrew on Mac, pacman on Arch
- Needs different PATH configs per OS
- Wants to keep dotfiles in sync via Git

**Persona 2: Tool-Heavy DevOps Engineer**
- Uses 10+ development tools (nvm, rbenv, pyenv, gcloud, kubectl, docker, etc.)
- Each tool requires specific initialization order
- Needs to disable certain tools per project/machine
- Wants rollback capability when configs break

**Persona 3: Shell Customization Enthusiast**
- 500+ lines of custom functions and aliases
- Organizes by category (git, docker, system, productivity)
- Experiments frequently, needs safe testing workflow
- Shares config modules with team

---

## Success Criteria

### Must-Have (v1.0)

1. **Feature Parity**: All 12 Python commands implemented
   - build, validate, init, migrate, diff
   - deploy, restore, clean-snapshots
   - list-modules, info
   - template list, template generate

2. **Performance**:
   - Startup time <50ms (Python: ~200ms)
   - Build time <500ms for 50 modules
   - Memory usage <50MB
   - Binary size <10MB

3. **Go Quality**:
   - Idiomatic Go code (gofmt, golint clean)
   - Test coverage >80%
   - Zero C dependencies (static binary)
   - Cross-compilation support (Linux, macOS, BSD)

4. **User Experience**:
   - Same CLI interface as Python version
   - Same manifest.yaml format (backward compatible)
   - Same error messages and help text
   - Self-documenting CLI with contextual guidance

### Should-Have (v1.1+)

1. **BSD Support**: FreeBSD, OpenBSD OS filtering
2. **Shell Completion**: Bash, Zsh, Fish completion scripts
3. **Watch Mode**: Auto-rebuild on manifest changes
4. **Config Validation**: Lint shell scripts before build
5. **Metrics**: Report module load times, file sizes

### Nice-to-Have (v2.0+)

1. **Remote Deployment**: Deploy to remote servers via SSH
2. **Hooks**: Pre-build, post-build, pre-deploy hooks
3. **Plugin System**: Extend with custom template generators
4. **Web UI**: Visual editor for manifest.yaml

### Non-Goals (Out of Scope)

1. ❌ Windows PowerShell support (different shell paradigm)
2. ❌ Shell script transpilation (bash → zsh conversion)
3. ❌ Runtime shell config hot-reloading
4. ❌ Cloud-hosted dotfile storage service

---

## Core Features

### 1. Manifest-Based Configuration

**Description**: Define shell modules and dependencies in YAML

**User Story**: As a developer, I want to declare my shell modules and their relationships so that I don't have to manually track dependencies.

**Acceptance Criteria**:
- Support YAML manifest with modules array
- Each module has: name, file, requires, os, description
- Validate manifest syntax and schema
- Provide clear error messages for invalid manifests

**Example**:
```yaml
modules:
  - name: brew-path
    file: init.d/brew.sh
    requires: []
    os: [Mac]
    description: Homebrew PATH initialization

  - name: nvm
    file: rc_pre.d/nvm.sh
    requires: [brew-path]
    os: [Mac, Linux]
    description: Node Version Manager
```

---

### 2. Automatic Dependency Resolution

**Description**: Resolve module load order via topological sort

**User Story**: As a user, I want modules to load in the correct order automatically so that I don't have to manually sequence them.

**Acceptance Criteria**:
- Build directed dependency graph from manifest
- Perform topological sort to determine load order
- Detect circular dependencies and report them clearly
- Handle modules with no dependencies (load first)

**Algorithm**: Kahn's algorithm or DFS-based topological sort

---

### 3. OS Filtering

**Description**: Include/exclude modules based on target operating system

**User Story**: As a multi-platform user, I want to specify which modules apply to which OS so that macOS-only tools don't load on Linux.

**Acceptance Criteria**:
- Support OS values: Mac, Linux (v1.0); BSD (v1.1+)
- Filter modules during build based on --os flag
- Default to current OS if not specified
- Modules with no OS field apply to all platforms

---

### 4. Shell Metadata System

**Description**: Built-in knowledge of shell configuration files across OS/shell/session types

**User Story**: As a user, I want the tool to know where to deploy my config so that I don't have to remember zsh vs bash file differences.

**Acceptance Criteria**:
- Support shells: zsh, bash, fish
- Support session types: login, interactive, always, gui
- Support OS: macOS, Ubuntu, Debian, Arch, Manjaro
- Provide `info` command to show config metadata
- Provide `--auto-output` flag to use recommended target file

**Example**:
```bash
shellforge build -c config -m manifest.yaml --auto-output
# Detects: macOS + zsh + interactive → outputs to ~/.zshrc
```

---

### 5. Build Command

**Description**: Generate shell configuration file from modules

**User Story**: As a user, I want to build my shell config from modules so that I can test it before deploying.

**Acceptance Criteria**:
- Read manifest, resolve dependencies, concatenate modules
- Support --output flag for custom output path
- Support --auto-output for automatic path detection
- Support --os, --shell, --session for cross-platform builds
- Support --dry-run to preview output without writing file
- Support --verbose to show detailed processing steps
- Generate header comment with metadata (generator, OS, module count)

---

### 6. Validation

**Description**: Check manifest and module files for errors

**User Story**: As a user, I want to validate my config before deploying so that I catch errors early.

**Acceptance Criteria**:
- Check YAML syntax
- Verify all module files exist
- Check all dependencies reference existing modules
- Detect circular dependencies
- Report missing required fields (name, file)
- Provide actionable error messages

---

### 7. Migration

**Description**: Convert monolithic RC file to modular structure

**User Story**: As a user migrating from a traditional .zshrc, I want to automatically split it into modules so that I don't have to do it manually.

**Acceptance Criteria**:
- Parse existing RC file and detect section boundaries
- Categorize sections into init.d/, rc_pre.d/, rc_post.d/
- Create module files with extracted content
- Auto-generate manifest.yaml with detected dependencies
- Create backup of original file (default, --no-backup to skip)
- Support --dry-run to preview migration

---

### 8. Auto-Init

**Description**: Generate manifest from existing modular config structure

**User Story**: As a user with existing init.d/rc_pre.d/rc_post.d directories, I want to auto-generate the manifest so that I don't have to write it by hand.

**Acceptance Criteria**:
- Scan init.d/, rc_pre.d/, rc_post.d/ directories
- Infer dependencies from module content (e.g., `$MACHINE` → requires os-detection)
- Detect OS support from `case $MACHINE` statements
- Extract descriptions from comments
- Generate manifest.yaml with helpful category comments

---

### 9. Deployment & Backup

**Description**: Deploy with automatic backup and version control

**User Story**: As a user, I want to safely deploy my config with automatic backups so that I can rollback if something breaks.

**Acceptance Criteria**:
- Create timestamped snapshot before deployment
- Commit snapshot to git repo (dual protection)
- Support `--deploy` flag in build command
- Support `restore` command to rollback
- Support `clean-snapshots` command to manage retention

**Backup Structure**:
```
~/.backup/shellforge/
├── .git/                          # Git version control
├── current/                       # Current deployed files
└── snapshots/
    ├── zshrc/
    │   ├── 2025-11-27_10-30-15
    │   └── 2025-11-27_11-45-22
    └── bashrc/
        └── 2025-11-26_09-15-00
```

---

### 10. Diff Comparison

**Description**: Compare generated config with existing RC file

**User Story**: As a user, I want to see what will change before deploying so that I can review modifications.

**Acceptance Criteria**:
- Support --auto-detect-existing to find current RC file
- Support --existing flag for explicit file path
- Support output formats: summary, unified, context
- Show statistics: added lines, removed lines, modified lines
- Suggest next steps (deploy, review specific changes)

---

### 11. Template Generation

**Description**: Generate common module patterns from templates

**User Story**: As a user, I want to quickly create standard modules (PATH, env vars, aliases) so that I don't have to write boilerplate.

**Acceptance Criteria**:
- Support templates: path, env, alias, conditional-source, tool-init, os-specific
- Support `-f` flag for field values
- Support `-i` for interactive mode
- Support `-r` flag to specify dependencies
- Auto-categorize into init.d/, rc_pre.d/, or rc_post.d/
- Provide `template list` command to show available templates

---

### 12. Self-Documenting CLI

**Description**: Contextual help and next-step suggestions

**User Story**: As a user, I want the CLI to guide me through workflows so that I don't have to constantly check documentation.

**Acceptance Criteria**:
- Show Quick Start Guide in --help
- Provide next-step suggestions after commands
- Show smart statistics (module counts, dependencies)
- Support --verbose for detailed output
- Provide error-specific troubleshooting steps

---

## Go-Specific Goals

### Idiomatic Go Design

1. **Standard Library First**: Use stdlib where possible (avoid over-dependency)
2. **Interfaces for Testability**: Define interfaces for all external dependencies
3. **Error Handling**: Use `errors` package, wrap errors with context
4. **Table-Driven Tests**: Use Go's idiomatic test structure
5. **Project Layout**: Follow `cmd/`, `internal/`, `pkg/` standard layout
6. **Documentation**: godoc for all exported functions/types

### Performance Targets

| Metric | Target | Rationale |
|--------|--------|-----------|
| Startup time | <50ms | Faster than Python (~200ms) |
| Build time (50 modules) | <500ms | Near-instant feedback |
| Memory usage | <50MB | Lightweight, suitable for containers |
| Binary size | <10MB | Easy distribution, fast download |

### Quality Targets

| Metric | Target | Rationale |
|--------|--------|-----------|
| Test coverage | >80% | High confidence in refactoring |
| Zero C deps | Static binary | Easy cross-compilation, no libc issues |
| Cross-compile | Linux, macOS, BSD | Support all Unix-like platforms |
| Code quality | gofmt, golint clean | Follow Go community standards |

---

## Out of Scope (v1.0)

These features are explicitly **not** included in the initial Go version:

1. **Windows Support**: Different shell paradigm (PowerShell, cmd.exe)
2. **Remote Deployment**: Requires SSH/SCP integration (add in v1.1+)
3. **Config Linting**: Shell script validation (use shellcheck separately)
4. **Watch Mode**: Auto-rebuild on file changes (add in v1.1+)
5. **GUI Interface**: Web-based manifest editor (v2.0+)
6. **Plugin System**: Custom template generators (v2.0+)

---

## Technical Constraints

### Must Use

- Go 1.21+ (for embed, generics if needed)
- Cobra CLI framework (industry standard)
- gopkg.in/yaml.v3 (YAML parsing)
- Standard library wherever possible

### Must Avoid

- Heavy dependencies (no NetworkX equivalent, use custom graph)
- CGO (require static binary)
- Non-standard project layouts
- Vendor directory (use go.mod only)

---

## Success Metrics

### Development Metrics

- Time to implement: <40 hours (1 week full-time)
- Lines of Go code: ~3000 (match Python)
- Test files: >20 (match Python)
- Dependencies: <10 external packages

### User Adoption Metrics

- Migration from Python version: >80% of Python users
- New user onboarding: <10 minutes from install to first build
- User satisfaction: Maintain same UX as Python version
- Bug reports: <5 critical bugs in v1.0

### Performance Metrics

- Build time improvement: >50% faster than Python
- Binary size: <10MB (vs Python's ~40MB virtualenv)
- Memory usage: <50MB (vs Python's ~80MB)
- Startup time: <50ms (vs Python's ~200ms)

---

## Dependencies

### From Python Version

- Manifest format: 100% backward compatible
- CLI interface: Identical command names and flags
- Error messages: Same or better clarity
- Help text: Same content, Go-styled formatting

### On External Tools

- Git: Required for backup/deployment features
- Operating system: Unix-like (Linux, macOS, BSD)
- Shell: Any POSIX shell (bash, zsh, fish, etc.)

---

## Assumptions

1. Users have Go 1.21+ installed (for development) OR use pre-built binaries
2. Users have Git installed (for backup/deployment features)
3. Users are on Unix-like systems (Linux, macOS, BSD)
4. Users' shell configs are text files (not binary)
5. Users' module files are <1MB each (reasonable for shell scripts)

---

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Go library lacks feature | High | Low | Use stdlib or implement custom (e.g., topological sort) |
| Performance regression vs Python | Medium | Low | Benchmark early, Go is typically faster |
| Breaking manifest compatibility | High | Medium | Strict YAML schema validation, comprehensive tests |
| User confusion during migration | Medium | Medium | Provide migration guide, maintain Python docs |
| Git not available on system | Medium | Low | Check at startup, fail with clear message |

---

## Appendix

### Related Projects

- **Python Version**: [gzh-cli-shellforge-py](https://github.com/Gizzahub/gzh-cli-shellforge)
- **Chezmoi**: Dotfile manager (complementary tool)
- **GNU Stow**: Symlink farm manager (different approach)
- **dotbot**: Dotfile bootstrap (different scope)

### References

- Python README: `/home/archmagece/myopen/gizzahub/gzh-cli-shellforge-py/README.md`
- Python codebase: `/home/archmagece/myopen/gizzahub/gzh-cli-shellforge-py/src/shellforge/`
- Shell startup files: `docs/deployment-guide.md` (Python version)

---

**Document Status**: Ready for review
**Next Steps**: Write REQUIREMENTS.md with detailed functional specifications
