# Shell Profiles Metadata

Comprehensive metadata about shell initialization files across different operating systems, shells, and execution contexts.

## Overview

This directory contains YAML files that define:
- Which initialization files are loaded by different shells
- How execution context affects profile loading
- Common problems and solutions for development environments
- Workarounds for automated and isolated execution contexts

## Files

### core.yaml
**OS and Shell Basics** (~165 lines, ~7KB)

Contains fundamental information about shell initialization:
- OS-specific profiles (Linux distributions, macOS)
- Shell types (bash, zsh, fish, sh)
- Load priority and order
- Shell detection commands
- OS detection patterns
- Default shells per OS

**Use this when**: You need to know which files a shell loads on a specific OS.

**Example questions**:
- What files does bash load on Ubuntu?
- What's the load order for zsh?
- How do I detect which OS I'm running on?

---

### contexts.yaml
**Execution Contexts** (~280 lines, ~10KB)

Defines how shell profiles are loaded in different execution contexts:
- SSH profiles
- X Window System initialization
- Desktop Environments (GNOME, KDE, XFCE, etc.)
- Display Managers (GDM, LightDM, SDDM)
- Shell execution modes (login/non-login, interactive/non-interactive)
- Session types (TTY, GUI, SSH)
- systemd user sessions
- XDG environment variables

**Use this when**: You need to understand how execution context affects profile loading.

**Example questions**:
- Why doesn't my profile load in SSH non-interactive mode?
- What's the difference between login and non-login shells?
- Which files are loaded when I start a GUI session?
- How do Display Managers set environment variables?

---

### dev.yaml
**Development Environments** (~145 lines, ~5KB)

Addresses common development environment issues:
- GUI IDE integrated terminals (VSCode, IntelliJ, PyCharm, etc.)
- Desktop launcher problems
- Language version managers:
  - Ruby: rbenv, rvm
  - Node.js: nvm
  - Python: pyenv, conda
  - Java: jenv, sdkman
  - Go: gvm
  - Rust: rustup

**Use this when**: Your development tools don't see your shell environment.

**Example problems**:
- VSCode terminal can't find rbenv
- nvm not available in IntelliJ
- pyenv missing when opening IDE from application menu
- conda not activated in GUI editor

**Solutions included**: VSCode settings, .desktop file modifications, environment.d configuration

---

### automation.yaml
**Automation and Isolated Environments** (~270 lines, ~10KB)

Covers scenarios where shell profiles are typically NOT loaded:
- **Scheduled execution**: cron, systemd timers, macOS launchd
- **User switching**: su, sudo variants
- **Containers**: Docker, chroot, Flatpak, Snap
- **Virtualization**: WSL, Android Termux
- **Remote execution**: SSH non-interactive, git hooks, CI/CD (GitHub Actions, GitLab CI, Jenkins)
- **Terminal multiplexers**: tmux, screen, zellij
- **System environment**: PAM, /etc/environment, profile.d, environment.d

**Use this when**: Scripts or automated tasks don't have your shell environment.

**Example problems**:
- Cron job can't find commands in custom PATH
- Docker exec doesn't see rbenv
- SSH command execution fails with "command not found"
- GitHub Actions can't find tools installed via version managers
- tmux new pane doesn't load nvm

**Solutions included**: Workarounds for each context, environment configuration strategies

---

## Quick Reference

### By Problem Type

| Problem | File | Section |
|---------|------|---------|
| Which files does bash load on Linux? | core.yaml | `os_profiles` |
| VSCode can't find rbenv | dev.yaml | `gui_app_contexts` |
| Cron job needs custom PATH | automation.yaml | `scheduled_execution.cron` |
| SSH command fails | automation.yaml | `remote_execution.ssh_non_interactive` |
| Login vs non-login shell | contexts.yaml | `shell_modes` |
| Desktop Environment autostart | contexts.yaml | `desktop_environments` |
| Docker exec environment | automation.yaml | `container_contexts.docker` |
| tmux new pane profile | automation.yaml | `terminal_multiplexers.tmux` |

### By Tool/Technology

| Tool | File | Section |
|------|------|---------|
| rbenv, rvm, nvm, pyenv | dev.yaml | `language_version_managers` |
| VSCode, IntelliJ, PyCharm | dev.yaml | `gui_app_contexts` |
| cron, systemd timer | automation.yaml | `scheduled_execution` |
| Docker, WSL, Flatpak | automation.yaml | `container_contexts` |
| tmux, screen, zellij | automation.yaml | `terminal_multiplexers` |
| GNOME, KDE, XFCE | contexts.yaml | `desktop_environments` |
| GDM, LightDM, SDDM | contexts.yaml | `display_managers` |

### By Use Case

**I'm a developer**:
1. Start with `dev.yaml` for IDE and language version manager issues
2. Refer to `core.yaml` for OS-specific shell behavior
3. Check `contexts.yaml` for session-specific problems

**I'm writing automation**:
1. Start with `automation.yaml` for cron/systemd/CI-CD
2. Refer to `core.yaml` for correct file paths
3. Check `contexts.yaml` for understanding execution modes

**I'm debugging shell issues**:
1. Start with `contexts.yaml` to understand execution context
2. Refer to `core.yaml` for correct load order
3. Check relevant sections in `dev.yaml` or `automation.yaml` for specific scenarios

---

## File Size Summary

| File | Lines | Size | Focus |
|------|-------|------|-------|
| core.yaml | ~165 | ~7KB | OS/Shell fundamentals |
| contexts.yaml | ~280 | ~10KB | Execution contexts |
| dev.yaml | ~145 | ~5KB | Development tools |
| automation.yaml | ~270 | ~10KB | Automation/isolation |
| **Total** | **~860** | **~32KB** | **All contexts** |

All files are optimized for LLM processing (< 10KB each).

---

## Usage in Code

### Go Example

```go
package shellmeta

import (
    "io/ioutil"
    "path/filepath"
    "gopkg.in/yaml.v3"
)

// LoadAllProfiles loads all profile metadata
func LoadAllProfiles(dir string) (*ShellProfiles, error) {
    profiles := &ShellProfiles{}

    // Load each file
    files := []string{"core.yaml", "contexts.yaml", "dev.yaml", "automation.yaml"}
    for _, file := range files {
        data, err := ioutil.ReadFile(filepath.Join(dir, file))
        if err != nil {
            return nil, err
        }

        // Merge into profiles struct
        var partial map[string]interface{}
        if err := yaml.Unmarshal(data, &partial); err != nil {
            return nil, err
        }

        // ... merge logic ...
    }

    return profiles, nil
}
```

---

## Contributing

When adding new content:
1. **Choose the right file**: Follow the categorization above
2. **Keep files small**: Each file should stay under 300 lines / 10KB
3. **Add cross-references**: Reference related sections in other files
4. **Update this README**: Add entries to the quick reference tables
5. **Test YAML validity**: Ensure correct YAML syntax

---

## Version History

- **v1.0** (2025-12-02): Initial split from monolithic shell-profiles.yaml
  - Split into 4 files for better LLM processing efficiency
  - Added comprehensive coverage of edge cases
  - Documented 860+ lines of shell profile metadata

---

## Related Documentation

- Project root: `../README.md`
- Examples: `../../examples/`
- Build system integration: `../../internal/domain/shellmeta/`
