# Frequently Asked Questions (FAQ)

Common questions about Shellforge answered.

---

## General Questions

### What is Shellforge?

Shellforge is a build tool that transforms your monolithic shell configuration files (`.zshrc`, `.bashrc`) into organized, modular structures with automatic dependency resolution and OS-specific filtering.

**Key features:**
- Automatic dependency resolution using topological sort
- OS-specific module filtering (Mac/Linux)
- Migration from existing configs
- Template generation
- Backup and restore system

### Why should I use Shellforge?

**Problems it solves:**
- ❌ Manual script concatenation causes ordering errors
- ❌ No dependency tracking leads to tools loading before their dependencies
- ❌ OS-specific logic scattered everywhere
- ❌ No validation until runtime errors

**Benefits:**
- ✅ Automatic dependency resolution
- ✅ OS-specific filtering
- ✅ Pre-deployment validation
- ✅ Modular, maintainable configuration
- ✅ Version control friendly

### Is Shellforge stable for production use?

**Current status:** Alpha (v0.2.0-alpha)

- **Core features (build, validate, list)** - Stable, ready for use
- **Migration and templates** - Stable
- **Backup/restore** - Stable
- **API** - May change in future releases

**Recommendation:**
- ✅ **Personal use** - Go ahead! Always backup first
- ✅ **Team use** - Yes, with proper testing
- ⚠️ **Enterprise** - Evaluate carefully, test thoroughly

### How is this different from other dotfile managers?

| Feature | Shellforge | Dotbot | GNU Stow | yadm |
|---------|------------|--------|----------|------|
| **Dependency resolution** | ✅ Automatic | ❌ Manual | ❌ None | ❌ None |
| **OS filtering** | ✅ Built-in | ⚠️ Manual | ❌ None | ⚠️ Manual |
| **Validation** | ✅ Pre-build | ❌ None | ❌ None | ❌ None |
| **Migration tools** | ✅ Automatic | ❌ Manual | ❌ Manual | ❌ Manual |
| **Template system** | ✅ Built-in | ❌ None | ❌ None | ❌ None |
| **Focus** | Shell configs | All dotfiles | All dotfiles | All dotfiles |

**Use Shellforge if:**
- You have complex shell configurations
- You need dependency management
- You maintain configs for multiple OSes
- You want automatic migration

**Use others if:**
- You manage all dotfiles (not just shell)
- You prefer symlink-based approach
- You need simpler solution

---

## Installation & Setup

### Do I need to install Go?

**For installation:** Yes (temporarily)

```bash
# Install Go
brew install go  # macOS
apt install golang-go  # Ubuntu

# Install Shellforge
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest

# After installation, Go is optional
# (unless you want to update Shellforge later)
```

**For usage:** No

Once installed, `gz-shellforge` is a standalone binary. You can uninstall Go if desired.

### Can I use Shellforge without Go?

**Coming soon:** Pre-built binaries

Currently, Go is required for installation. Future releases will provide:
- Homebrew formula (macOS)
- APT/RPM packages (Linux)
- Pre-built binaries (all platforms)

### Which shells are supported?

**Fully supported:**
- ✅ zsh 5.8+ (macOS default)
- ✅ bash 4.0+ (Linux default)

**Planned:**
- ⏳ fish 3.0+
- ⏳ PowerShell (experimental)

**Not supported:**
- ❌ csh/tcsh
- ❌ ksh

### Which operating systems are supported?

**Fully supported:**
- ✅ macOS 10.15+ (Catalina and later)
- ✅ Linux (Ubuntu 20.04+, Debian 11+, Arch, Fedora, CentOS)

**Experimental:**
- ⚠️ FreeBSD 13+ (basic features work)

**Not supported:**
- ❌ Windows (use WSL instead)

---

## Usage Questions

### How do I get started?

Follow the [Quick Start Guide](00-quick-start.md):

1. **Install**: `go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest`
2. **Backup**: `gz-shellforge backup --file ~/.zshrc`
3. **Migrate**: `gz-shellforge migrate ~/.zshrc`
4. **Build**: `gz-shellforge build --os Mac --output ~/.zshrc.new`
5. **Deploy**: `mv ~/.zshrc.new ~/.zshrc`

**Time to complete:** 5 minutes

### Can I use Shellforge with my existing dotfiles?

**Yes!** Shellforge integrates with existing setups:

**Option 1: Migrate your existing config**
```bash
gz-shellforge migrate ~/.zshrc
```

**Option 2: Start fresh and copy sections**
```bash
# Create new modular structure
mkdir ~/shellforge
cd ~/shellforge

# Copy sections from old config manually
vim modules/rc_post.d/my-aliases.sh
# Paste aliases from old .zshrc

# Create manifest
vim manifest.yaml

# Build
gz-shellforge build --os Mac --output ~/.zshrc
```

**Option 3: Hybrid approach**
```bash
# Use Shellforge for core config
gz-shellforge build --os Mac --output ~/.zshrc.shellforge

# Source from main config
echo 'source ~/.zshrc.shellforge' >> ~/.zshrc
# Keep other stuff in .zshrc
```

### Do I have to migrate my entire config at once?

**No!** Start small:

```bash
# Start with just a few modules
gz-shellforge migrate ~/.zshrc --dry-run

# Review detected sections
# Manually select which to migrate

# Create minimal manifest
cat > manifest.yaml <<EOF
modules:
  - name: aliases
    file: rc_post.d/aliases.sh
    requires: []
EOF

# Build just these modules
gz-shellforge build --os Mac --dry-run
```

**Incremental approach:**
1. Migrate core PATH and env vars
2. Add tool initialization (nvm, rbenv)
3. Add aliases and functions
4. Add custom configurations

### How do I add new modules?

**Method 1: Use templates**
```bash
gz-shellforge template generate alias my-aliases \
  -f aliases="alias k='kubectl'"
```

**Method 2: Create manually**
```bash
# Create file
vim modules/rc_post.d/my-module.sh

# Add to manifest
vim manifest.yaml

# Rebuild
gz-shellforge build --os Mac --output ~/.zshrc
```

### Can I test changes without affecting my current shell?

**Yes!** Multiple safe testing methods:

**Method 1: Dry run**
```bash
gz-shellforge build --os Mac --dry-run | less
```

**Method 2: Output to different file**
```bash
gz-shellforge build --os Mac --output ~/.zshrc.test
```

**Method 3: Test in new shell**
```bash
zsh -c 'source ~/.zshrc.test; echo "Test passed"'
```

**Method 4: Test in new terminal**
```bash
gz-shellforge build --os Mac --output ~/.zshrc.test
# Open new terminal, run: source ~/.zshrc.test
```

---

## Features Questions

### What is dependency resolution?

**Automatic ordering** of modules based on dependencies.

**Example problem:**
```yaml
# Without dependency resolution, you might load:
1. nvm (needs brew-path)
2. brew-path (defines PATH)
# Result: nvm fails because brew not in PATH!
```

**With Shellforge:**
```yaml
- name: brew-path
  requires: []

- name: nvm
  requires: [brew-path]  # Shellforge loads brew-path first
```

**Algorithm:** Topological sort (Kahn's algorithm)
**Result:** Modules always load in correct order

### What is OS filtering?

**Load different modules** based on operating system.

**Example:**
```yaml
- name: brew-path
  file: init.d/brew-path.sh
  os: [Mac]  # Only loads on macOS

- name: apt-functions
  file: rc_post.d/apt.sh
  os: [Linux]  # Only loads on Linux

- name: universal-aliases
  file: rc_post.d/aliases.sh
  os: []  # Loads on all OSes (empty = all)
```

**Benefits:**
- Maintain single manifest for multiple OSes
- Deploy platform-specific configs easily
- Avoid conditional logic in shell scripts

### How does migration work?

**Automatic conversion** from monolithic to modular.

**Process:**
1. **Detect sections** using header patterns:
   ```bash
   # --- Section Name ---
   # === Section Name ===
   ## Section Name
   # SECTION NAME
   ```

2. **Categorize** by content:
   - PATH, env vars → `init.d/`
   - Tool init (nvm, rbenv) → `rc_pre.d/`
   - Aliases, functions → `rc_post.d/`

3. **Infer dependencies:**
   - Uses `$MACHINE` → requires `os-detection`
   - Uses `brew` → requires `brew-path`

4. **Generate manifest** with metadata

**Result:** Organized modules with correct dependencies

### What templates are available?

**6 built-in templates:**

1. **path** - Add directory to PATH
2. **env** - Set environment variable
3. **alias** - Define shell aliases
4. **conditional-source** - Source file if exists
5. **tool-init** - Initialize development tool
6. **os-specific** - OS-specific configuration

**Usage:**
```bash
gz-shellforge template list
gz-shellforge template generate <type> <name>
```

**Can I create custom templates?**
Not yet - planned for future release.

### How does backup/restore work?

**Git-backed versioning** with timestamps.

**Features:**
- Timestamped snapshots
- Full git history
- Retention policies
- Safety backups on restore

**Usage:**
```bash
# Backup
gz-shellforge backup --file ~/.zshrc

# List snapshots
ls -la ~/.backup/shellforge/

# Restore
gz-shellforge restore --file ~/.zshrc --snapshot 2025-12-01_14-30-45

# Cleanup old snapshots
gz-shellforge cleanup --file ~/.zshrc --keep-count 10
```

---

## Troubleshooting Questions

### My shell takes longer to start after using Shellforge

**Likely causes:**
1. Too many modules loading
2. Slow operations in modules (network calls, heavy computation)
3. Unoptimized tool initialization

**Solutions:**
```bash
# Profile shell startup
zsh -i -c exit

# Use lazy loading for slow tools
# Example: lazy load nvm
function nvm() {
  unfunction nvm
  [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
  nvm "$@"
}

# Defer non-critical modules
# Load asynchronously or on demand
```

See [Troubleshooting Guide](60-troubleshooting.md#performance-issues) for details.

### I get "circular dependency" error

**Cause:** Module A requires B, B requires C, C requires A (cycle!)

**Solution:**
```bash
# Find the cycle (error message shows it)
Error: circular dependency detected
Cycle: module-a → module-b → module-c → module-a

# Break the cycle in manifest.yaml
# Remove one of the requires to break chain
```

See [Troubleshooting Guide](60-troubleshooting.md#circular-dependency-detected).

### My custom functions don't work after migration

**Possible causes:**
1. Function defined after usage
2. Missing dependency
3. Incorrect categorization

**Solutions:**
```bash
# Check load order
gz-shellforge list -v

# Ensure definition before usage
# Add dependency if needed

# Verify module categorization
# Functions should be in rc_post.d/
```

### Can I see what Shellforge will generate before deploying?

**Yes!** Use `--dry-run`:

```bash
gz-shellforge build --os Mac --dry-run | less
```

Or compare with diff:
```bash
gz-shellforge build --os Mac --output ~/.zshrc.new
gz-shellforge diff ~/.zshrc ~/.zshrc.new
```

---

## Advanced Questions

### Can I use Shellforge in CI/CD?

**Yes!** Common use cases:

**Automated testing:**
```bash
#!/bin/bash
# ci-test.sh

# Validate manifest
gz-shellforge validate || exit 1

# Build for all OSes
gz-shellforge build --os Mac --output /tmp/test-mac.sh
gz-shellforge build --os Linux --output /tmp/test-linux.sh

# Test syntax
zsh -n /tmp/test-mac.sh || exit 1
bash -n /tmp/test-linux.sh || exit 1

echo "CI passed!"
```

**Automated deployment:**
```yaml
# .github/workflows/deploy-shell-config.yml
name: Deploy Shell Config
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Install Shellforge
        run: go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest
      - name: Build configs
        run: |
          gz-shellforge build --os Mac --output dist/zshrc-mac
          gz-shellforge build --os Linux --output dist/bashrc-linux
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: shell-configs
          path: dist/
```

### Can I use Shellforge as a Go library?

**Yes!** Use the public API:

```go
import "github.com/gizzahub/gzh-cli-shellforge/pkg/cmd"

// Build shell config programmatically
result, err := cmd.Build(cmd.BuildOptions{
    ManifestPath: "manifest.yaml",
    ConfigDir:    "modules",
    TargetOS:     "Mac",
})
```

See [API Reference](../reference/api.md) for full documentation.

### How can I contribute to Shellforge?

**We welcome contributions!**

1. Read [Contributing Guide](../developer/30-contributing.md)
2. Check [open issues](https://github.com/gizzahub/gzh-cli-shellforge/issues)
3. Fork repository
4. Create feature branch
5. Write tests
6. Submit pull request

**Areas needing help:**
- Documentation improvements
- More templates
- Platform-specific testing
- Performance optimizations

### Can I use Shellforge with Docker?

**Yes!** Example Dockerfile:

```dockerfile
FROM golang:1.21-alpine AS builder
RUN go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest

FROM alpine:latest
COPY --from=builder /go/bin/gz-shellforge /usr/local/bin/
WORKDIR /workspace
ENTRYPOINT ["gz-shellforge"]
```

**Usage:**
```bash
docker run -v $(pwd):/workspace shellforge:latest \
  build --os Linux --output /workspace/output.sh
```

### Does Shellforge support plugins?

**Not yet - planned for future release.**

**Planned features:**
- Custom validators
- Custom templates
- Custom module types
- Hook system

**Current workaround:**
Use template generation and custom modules.

---

## Performance Questions

### Is Shellforge fast?

**Yes!** Significantly faster than Python version:

| Metric | Python | Go | Improvement |
|--------|--------|----|----|
| Startup | ~200ms | <10ms | 20x faster |
| Build (10 modules) | ~300ms | <50ms | 6x faster |
| Memory | ~80MB | <10MB | 8x lighter |

**Benchmarks:** See [Performance Benchmarks](../developer/60-benchmarks.md)

### Will Shellforge make my shell slower?

**No!** Shellforge generates static shell scripts.

**Impact on shell startup:**
- Zero - Shellforge runs at build time, not shell startup
- Generated config is same as handwritten script
- No performance overhead

**Only build time matters:**
- Build happens manually when you change modules
- Not every shell startup

---

## Security Questions

### Is Shellforge safe to use?

**Yes**, with standard precautions:

**Security features:**
- ✅ No network calls during build
- ✅ No arbitrary code execution
- ✅ Git-backed versioning for rollback
- ✅ Validation before deployment
- ✅ Open source (audit anytime)

**Best practices:**
- Always backup before changes
- Review generated config before deploying
- Use version control for modules
- Don't store secrets in modules

### Can Shellforge leak my secrets?

**No**, unless you put secrets in modules.

**Safe practice:**
```bash
# DON'T put secrets directly
export API_KEY="secret-key-here"  # ❌ Bad

# DO reference external file
export API_KEY=$(cat ~/.secrets/api-key)  # ✅ Good

# DO use environment
export API_KEY="${API_KEY}"  # ✅ Good

# DO use password manager
export API_KEY=$(security find-generic-password -s "API_KEY" -w)  # ✅ Best
```

### Does Shellforge phone home?

**No.** Shellforge:
- ❌ No telemetry
- ❌ No analytics
- ❌ No network calls
- ❌ No data collection

**Completely offline** - all operations are local.

---

## Compatibility Questions

### Can I use Shellforge with oh-my-zsh?

**Yes!** Two approaches:

**Approach 1: Migrate oh-my-zsh config**
```bash
gz-shellforge migrate ~/.zshrc
# Includes oh-my-zsh if present
```

**Approach 2: Keep oh-my-zsh, use Shellforge for custom**
```bash
# In manifest.yaml
- name: oh-my-zsh
  file: init.d/oh-my-zsh.sh
  requires: []

- name: my-custom
  file: rc_post.d/custom.sh
  requires: [oh-my-zsh]
```

### Does Shellforge work with Homebrew?

**Yes!** Create brew-path module:

```bash
gz-shellforge template generate path brew-path \
  -f path_dir="/opt/homebrew/bin"
```

Or let migration detect it automatically.

### Can I use Shellforge with my team's shared dotfiles?

**Yes!** Perfect for team sharing:

```bash
# Team repository structure
dotfiles/
├── manifest.yaml
├── modules/
│   ├── init.d/
│   ├── rc_pre.d/
│   └── rc_post.d/
└── README.md

# Each team member:
git clone git@github.com:team/dotfiles.git
cd dotfiles
gz-shellforge build --os Mac --output ~/.zshrc  # or Linux
```

**Benefits for teams:**
- Version control friendly
- Easy to review changes
- Platform-specific configs
- Shared and personal modules

---

## Still Have Questions?

- **Documentation**: [User Guide](README.md)
- **Troubleshooting**: [Troubleshooting Guide](60-troubleshooting.md)
- **Examples**: [examples/](../../examples/)
- **GitHub Discussions**: [Ask a question](https://github.com/gizzahub/gzh-cli-shellforge/discussions)
- **GitHub Issues**: [Report a bug](https://github.com/gizzahub/gzh-cli-shellforge/issues)

---

**Last Updated**: 2025-12-01
