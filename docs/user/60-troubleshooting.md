# Troubleshooting Guide

Common issues and solutions for Shellforge users.

---

## Table of Contents

- [Installation Issues](#installation-issues)
- [Build Errors](#build-errors)
- [Validation Errors](#validation-errors)
- [Migration Issues](#migration-issues)
- [Runtime Errors](#runtime-errors)
- [Performance Issues](#performance-issues)
- [Platform-Specific Issues](#platform-specific-issues)
- [Getting More Help](#getting-more-help)

---

## Installation Issues

### Command Not Found: gz-shellforge

**Problem:**
```bash
$ gz-shellforge --version
zsh: command not found: gz-shellforge
```

**Solutions:**

1. **Check if Go bin is in PATH:**
   ```bash
   echo $PATH | grep "$(go env GOPATH)/bin"
   ```

   If not found, add to your shell config:
   ```bash
   # For zsh (~/.zshrc)
   export PATH="$PATH:$(go env GOPATH)/bin"

   # For bash (~/.bashrc)
   export PATH="$PATH:$(go env GOPATH)/bin"

   # Reload shell
   source ~/.zshrc  # or source ~/.bashrc
   ```

2. **Verify installation:**
   ```bash
   ls -la $(go env GOPATH)/bin/gz-shellforge
   ```

3. **Reinstall if missing:**
   ```bash
   go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest
   ```

### Go Version Too Old

**Problem:**
```
go: module requires Go 1.21 or later
```

**Solution:**
Update Go to 1.21 or later:

```bash
# macOS (Homebrew)
brew upgrade go

# Linux (download from golang.org)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Verify
go version
```

### Build Fails During Installation

**Problem:**
```
# github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge
build error: ...
```

**Solutions:**

1. **Clean Go cache:**
   ```bash
   go clean -modcache
   go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest
   ```

2. **Check network connectivity:**
   ```bash
   ping pkg.go.dev
   ```

3. **Use proxy if behind firewall:**
   ```bash
   export GOPROXY=https://proxy.golang.org,direct
   go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest
   ```

---

## Build Errors

### Circular Dependency Detected

**Problem:**
```
Error: circular dependency detected
Cycle: module-a → module-b → module-c → module-a
```

**Solution:**

1. **Identify the cycle** in `manifest.yaml`:
   ```yaml
   # WRONG - creates cycle
   - name: module-a
     requires: [module-c]

   - name: module-b
     requires: [module-a]

   - name: module-c
     requires: [module-b]
   ```

2. **Break the cycle** by removing unnecessary dependency:
   ```yaml
   # CORRECT
   - name: module-a
     requires: []

   - name: module-b
     requires: [module-a]

   - name: module-c
     requires: [module-b]
   ```

3. **Validate after fix:**
   ```bash
   gz-shellforge validate -v
   ```

### Module File Not Found

**Problem:**
```
Error: module file not found: modules/init.d/missing.sh
Module: my-module
```

**Solutions:**

1. **Check file path in manifest:**
   ```yaml
   # Ensure file path is relative to config-dir
   - name: my-module
     file: init.d/my-module.sh  # NOT: /absolute/path/to/file.sh
   ```

2. **Verify file exists:**
   ```bash
   ls -la modules/init.d/my-module.sh
   ```

3. **Check file permissions:**
   ```bash
   chmod 644 modules/init.d/my-module.sh
   ```

4. **Use correct config-dir flag:**
   ```bash
   gz-shellforge build --config-dir modules --os Mac
   ```

### OS Not Specified

**Problem:**
```
Error: --os flag is required
```

**Solution:**
Always specify OS when building:

```bash
# Correct
gz-shellforge build --os Mac --output ~/.zshrc
gz-shellforge build --os Linux --output ~/.bashrc

# Wrong
gz-shellforge build --output ~/.zshrc  # Missing --os
```

### Output File Permission Denied

**Problem:**
```
Error: failed to write output file: permission denied
File: /etc/zshrc
```

**Solutions:**

1. **Use user home directory:**
   ```bash
   # Use ~ not /
   gz-shellforge build --os Mac --output ~/.zshrc
   ```

2. **Check file permissions:**
   ```bash
   ls -la ~/.zshrc
   ```

3. **Use sudo only if really needed:**
   ```bash
   sudo gz-shellforge build --os Mac --output /etc/zshrc
   ```

---

## Validation Errors

### Invalid YAML Syntax

**Problem:**
```
Error: failed to parse manifest
Line 10: mapping values are not allowed here
```

**Solutions:**

1. **Check YAML syntax:**
   ```yaml
   # WRONG - missing quotes
   - name: my-module
     description: This is: invalid

   # CORRECT - quoted
   - name: my-module
     description: "This is: valid"
   ```

2. **Validate YAML online:**
   Use [YAML Lint](http://www.yamllint.com/) to check syntax

3. **Check indentation:**
   ```yaml
   # WRONG - inconsistent indentation
   modules:
   - name: module-a
     requires: [module-b]

   # CORRECT - consistent 2-space indentation
   modules:
     - name: module-a
       requires: [module-b]
   ```

### Missing Required Fields

**Problem:**
```
Error: module missing required field: name
Module index: 3
```

**Solution:**
Ensure all required fields are present:

```yaml
# Required fields: name, file
modules:
  - name: my-module       # Required
    file: init.d/file.sh  # Required
    requires: []          # Optional (default: [])
    os: [Mac]             # Optional (default: all)
    description: "..."    # Optional
```

### Duplicate Module Names

**Problem:**
```
Error: duplicate module name: os-detection
```

**Solution:**
Each module name must be unique:

```yaml
# WRONG - duplicate names
modules:
  - name: os-detection
    file: init.d/os-detection.sh

  - name: os-detection  # Duplicate!
    file: init.d/other.sh

# CORRECT - unique names
modules:
  - name: os-detection
    file: init.d/os-detection.sh

  - name: os-detection-extended
    file: init.d/other.sh
```

---

## Migration Issues

### No Sections Detected

**Problem:**
```
Warning: No sections detected in RC file
File: ~/.zshrc
```

**Solutions:**

1. **Check section header formats** Shellforge supports:
   ```bash
   # Supported formats:
   # --- Section Name ---
   # === Section Name ===
   ## Section Name
   # SECTION NAME
   ```

2. **Add headers manually** if missing:
   ```bash
   # Edit your .zshrc before migration
   # Add section headers like:
   # --- PATH Setup ---
   export PATH="/usr/local/bin:$PATH"

   # --- NVM ---
   [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
   ```

3. **Use dry-run to preview:**
   ```bash
   gz-shellforge migrate ~/.zshrc --dry-run -v
   ```

### Wrong Categorization

**Problem:**
Migration put modules in wrong categories (init.d vs rc_pre.d vs rc_post.d).

**Solutions:**

1. **Review generated manifest:**
   ```bash
   cat manifest.yaml
   ```

2. **Manually move files:**
   ```bash
   mv modules/init.d/nvm.sh modules/rc_pre.d/nvm.sh
   ```

3. **Update manifest:**
   ```yaml
   - name: nvm
     file: rc_pre.d/nvm.sh  # Update path
   ```

4. **Re-categorize guidelines:**
   - `init.d/` - PATH, OS detection, early init
   - `rc_pre.d/` - Tool initialization (nvm, rbenv, pyenv)
   - `rc_post.d/` - Aliases, functions, customization

### Dependencies Not Inferred

**Problem:**
Dependencies not automatically detected during migration.

**Solution:**
Manually add dependencies in `manifest.yaml`:

```yaml
# If module B uses variables from module A
- name: module-a
  file: init.d/module-a.sh
  requires: []

- name: module-b
  file: rc_pre.d/module-b.sh
  requires: [module-a]  # Add manually
```

**Common dependency patterns:**
- If module uses `$MACHINE` → requires `os-detection`
- If module uses `brew` → requires `brew-path` (Mac only)
- If module sources a file → requires module that creates that file

---

## Runtime Errors

### Shell Functions Not Available

**Problem:**
Functions defined in modules don't work after sourcing generated config.

**Solutions:**

1. **Check module load order:**
   ```bash
   gz-shellforge list -v
   # Ensure function definitions load before usage
   ```

2. **Verify dependencies:**
   ```yaml
   - name: helper-functions
     file: rc_post.d/helpers.sh
     requires: []

   - name: my-functions
     file: rc_post.d/my-functions.sh
     requires: [helper-functions]  # Add if needed
   ```

3. **Test generated config:**
   ```bash
   # Source in new shell
   zsh -c 'source ~/.zshrc.new; type my_function'
   ```

### Environment Variables Not Set

**Problem:**
Environment variables from modules not available.

**Solutions:**

1. **Check if module loads for your OS:**
   ```yaml
   - name: mac-env
     file: rc_pre.d/mac-env.sh
     os: [Mac]  # Won't load on Linux!
   ```

2. **Verify build OS matches:**
   ```bash
   # If on Mac, build for Mac
   gz-shellforge build --os Mac --output ~/.zshrc
   ```

3. **Check load order:**
   ```bash
   # ENV vars should load early
   gz-shellforge list -v
   ```

### Aliases Not Working

**Problem:**
Aliases defined in modules don't work.

**Solutions:**

1. **Ensure aliases load last:**
   ```yaml
   - name: my-aliases
     file: rc_post.d/aliases.sh  # Should be in rc_post.d/
     requires: []
   ```

2. **Check shell type:**
   ```bash
   # Some aliases work only in zsh or bash
   echo $SHELL
   ```

3. **Verify syntax:**
   ```bash
   # Correct alias syntax
   alias ll='ls -la'
   alias k='kubectl'
   ```

---

## Performance Issues

### Slow Shell Startup

**Problem:**
Shell takes too long to start after using Shellforge.

**Solutions:**

1. **Profile shell startup:**
   ```bash
   # Zsh profiling
   zsh -i -c exit

   # Or add to .zshrc temporarily:
   zmodload zsh/zprof
   # ... rest of config ...
   zprof
   ```

2. **Identify slow modules:**
   Look for modules doing:
   - Network calls
   - Heavy computation
   - Large file reads

3. **Optimize slow modules:**
   ```bash
   # Use conditional loading
   if [ -d "$NVM_DIR" ]; then
     [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
   fi

   # Lazy load tools
   function nvm() {
     unfunction nvm
     [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
     nvm "$@"
   }
   ```

4. **Defer non-critical modules:**
   Move to separate file and source asynchronously.

### Large Output File

**Problem:**
Generated shell config is too large.

**Solutions:**

1. **Remove unused modules:**
   ```bash
   # List all modules
   gz-shellforge list -v

   # Remove unused from manifest.yaml
   ```

2. **Split large modules:**
   Break large modules into smaller, focused ones.

3. **Use external files:**
   Instead of inline content, source external files:
   ```bash
   # In module:
   [ -f ~/.aliases ] && source ~/.aliases
   ```

---

## Platform-Specific Issues

### macOS Specific

#### Homebrew Not Found After Build

**Problem:**
```
brew: command not found
```

**Solution:**

1. **Check brew-path module exists:**
   ```bash
   gz-shellforge list --filter Mac -v | grep brew
   ```

2. **Verify PATH in module:**
   ```bash
   cat modules/init.d/05-brew-path.sh
   # Should contain:
   export PATH="/opt/homebrew/bin:$PATH"  # M1/M2 Mac
   # or
   export PATH="/usr/local/bin:$PATH"     # Intel Mac
   ```

3. **Ensure module loads:**
   ```yaml
   - name: brew-path
     file: init.d/05-brew-path.sh
     requires: [os-detection]
     os: [Mac]
   ```

#### macOS Catalina+ zsh Warning

**Problem:**
```
The default interactive shell is now zsh.
To update your account to use zsh, please run `chsh -s /bin/zsh`.
```

**Solution:**
This is normal. Shellforge works with both zsh and bash. Ignore or switch to zsh:
```bash
chsh -s /bin/zsh
```

### Linux Specific

#### System-wide Config Issues

**Problem:**
Changes don't apply system-wide.

**Solution:**

1. **User config:**
   ```bash
   gz-shellforge build --os Linux --output ~/.bashrc
   ```

2. **System-wide config:**
   ```bash
   sudo gz-shellforge build --os Linux --output /etc/bashrc
   # Or
   sudo gz-shellforge build --os Linux --output /etc/bash.bashrc
   ```

#### Package Manager Detection

**Problem:**
Module uses `apt` but system has `yum`.

**Solution:**
Add OS detection logic in module:

```bash
# In module file
if command -v apt >/dev/null 2>&1; then
  # Debian/Ubuntu
  alias update='sudo apt update && sudo apt upgrade'
elif command -v yum >/dev/null 2>&1; then
  # CentOS/RHEL
  alias update='sudo yum update'
elif command -v pacman >/dev/null 2>&1; then
  # Arch
  alias update='sudo pacman -Syu'
fi
```

---

## Getting More Help

### Enable Verbose Mode

Get detailed output for debugging:

```bash
# Verbose validation
gz-shellforge validate --verbose

# Verbose build
gz-shellforge build --os Mac --verbose

# Verbose migration
gz-shellforge migrate ~/.zshrc --verbose
```

### Check Logs

```bash
# Build with verbose and save output
gz-shellforge build --os Mac -v 2>&1 | tee build.log

# Review log
less build.log
```

### Dry Run Testing

Test without making changes:

```bash
# Migration dry run
gz-shellforge migrate ~/.zshrc --dry-run

# Build dry run
gz-shellforge build --os Mac --dry-run
```

### Report Issues

If problem persists:

1. **Check existing issues:**
   [GitHub Issues](https://github.com/gizzahub/gzh-cli-shellforge/issues)

2. **Gather information:**
   ```bash
   gz-shellforge --version
   go version
   uname -a
   echo $SHELL
   ```

3. **Create minimal reproduction:**
   - Simplify manifest to minimal case
   - Include exact commands and error messages

4. **Submit issue:**
   Include:
   - Shellforge version
   - Operating system
   - Complete error message
   - Steps to reproduce
   - Manifest.yaml (if applicable)

---

## Common Error Messages Reference

| Error Message | Likely Cause | Solution |
|---------------|--------------|----------|
| `command not found: gz-shellforge` | Not in PATH | Add Go bin to PATH |
| `circular dependency detected` | Dependency cycle | Remove circular requires |
| `module file not found` | Wrong file path | Check path in manifest |
| `--os flag is required` | Missing OS flag | Add --os Mac or --os Linux |
| `permission denied` | Wrong output location | Use ~/.zshrc not /etc/zshrc |
| `invalid YAML syntax` | YAML parsing error | Check indentation and quotes |
| `duplicate module name` | Name collision | Use unique module names |
| `failed to parse manifest` | Corrupt YAML | Validate YAML syntax |

---

## Best Practices to Avoid Issues

### 1. Always Validate Before Building
```bash
gz-shellforge validate && gz-shellforge build --os Mac --output ~/.zshrc
```

### 2. Use Dry Run First
```bash
gz-shellforge build --os Mac --dry-run | less
```

### 3. Backup Before Changing
```bash
gz-shellforge backup --file ~/.zshrc --message "Before rebuild"
```

### 4. Test in New Shell
```bash
zsh -c 'source ~/.zshrc.new; echo "Test passed"'
```

### 5. Keep Modules Simple
- One concern per module
- Clear dependencies
- Descriptive names

### 6. Use Version Control
```bash
cd ~/shellforge
git init
git add .
git commit -m "Initial shellforge setup"
```

---

## Still Having Issues?

- **Documentation**: [User Guide](README.md)
- **Examples**: [examples/](../../examples/)
- **FAQ**: [FAQ](70-faq.md)
- **GitHub Issues**: [Report a bug](https://github.com/gizzahub/gzh-cli-shellforge/issues/new)
- **Discussions**: [Ask a question](https://github.com/gizzahub/gzh-cli-shellforge/discussions)

---

**Last Updated**: 2025-12-01
