# Frequently Asked Questions (FAQ)

## General Questions

### What is Shellforge?

Shellforge is a build tool that transforms modular shell scripts into unified shell configurations (like `.zshrc` or `.bashrc`). It automatically resolves dependencies between modules, filters by operating system, and validates your configuration before deployment.

### Why should I use Shellforge instead of manually managing my shell config?

**Problems Shellforge solves:**
- **Dependency management**: Automatically loads modules in the correct order
- **OS compatibility**: Easy multi-OS configuration (Mac/Linux) with filtering
- **Modularity**: Organize your shell config into logical, reusable pieces
- **Validation**: Catch errors before deployment (circular dependencies, missing files)
- **Version control**: Git-backed backups and snapshots
- **Team sharing**: Share modular configs across teams with ease

### Is Shellforge compatible with my current shell configuration?

Yes! Shellforge supports:
- **Zsh** (5.8+)
- **Bash** (4.0+)
- **Fish** (3.0+)

The `migrate` command can convert your existing monolithic `.zshrc` or `.bashrc` into modular structure automatically.

---

## Installation & Setup

### How do I install Shellforge?

**Option 1: From source (requires Go 1.21+)**
```bash
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest
```

**Option 2: Build locally**
```bash
git clone https://github.com/gizzahub/gzh-cli-shellforge.git
cd gzh-cli-shellforge
make install
```

See [QUICK_START.md](QUICK_START.md) for detailed instructions.

### Why is the command `gz-shellforge` instead of just `shellforge`?

The `gz-` prefix is part of the Gizzahub CLI tool family naming convention. It helps avoid naming conflicts with other tools while maintaining consistency across the gzh-cli ecosystem.

### Do I need Git installed?

Git is **optional** but recommended:
- ✅ **Required for**: Backup/restore features with version control
- ❌ **Not required for**: Build, validate, migrate, template commands

You can disable git features with the `--no-git` flag.

---

## Usage Questions

### How do I convert my existing .zshrc to modular structure?

Use the `migrate` command:

```bash
# Preview what would be created
gz-shellforge migrate ~/.zshrc --dry-run

# Perform the migration
gz-shellforge migrate ~/.zshrc --output-dir ~/dotfiles/modules

# Validate the result
cd ~/dotfiles
gz-shellforge validate

# Build and compare
gz-shellforge build --os Mac --output ~/.zshrc.new
gz-shellforge diff ~/.zshrc ~/.zshrc.new
```

See [docs/user/30-workflows.md](docs/user/30-workflows.md) for complete workflow.

### What does "circular dependency detected" mean?

A circular dependency occurs when modules depend on each other in a loop:

```
Module A requires Module B
Module B requires Module C
Module C requires Module A  ← Circular!
```

**How to fix:**
1. Run `gz-shellforge validate --verbose` to see the cycle
2. Edit `manifest.yaml` and remove one dependency from the loop
3. Re-run `gz-shellforge validate` to confirm the fix

### How do I add a new module?

**Option 1: From template**
```bash
gz-shellforge template list  # See available templates
gz-shellforge template generate alias my-aliases \
  -f aliases='alias ll="ls -la"'
```

**Option 2: Manual creation**
```bash
# Create the file
vim modules/rc_post.d/my-module.sh

# Add to manifest.yaml
modules:
  - name: my-module
    file: rc_post.d/my-module.sh
    requires: [os-detection]  # Dependencies
    os: [Mac, Linux]          # Supported OSes
    description: My custom module
```

Then rebuild:
```bash
gz-shellforge build --os Mac --output ~/.zshrc
```

### How do I handle OS-specific configurations?

Use the `os` field in your manifest:

```yaml
modules:
  - name: homebrew-path
    file: init.d/homebrew.sh
    os: [Mac]              # Only load on macOS
    description: Homebrew PATH setup

  - name: apt-setup
    file: init.d/apt.sh
    os: [Linux]            # Only load on Linux
    description: APT configuration
```

Leave `os` empty (or omit it) for modules that work on all platforms.

---

## Troubleshooting

### Why doesn't my module load?

**Check these common issues:**

1. **File path is wrong**: Ensure the `file` path in `manifest.yaml` matches the actual location
   ```bash
   gz-shellforge validate  # Will report missing files
   ```

2. **OS filtering**: Module might be filtered out for your OS
   ```bash
   gz-shellforge list --filter Mac  # See which modules apply
   ```

3. **Dependency not met**: A required module might be missing
   ```bash
   gz-shellforge validate --verbose  # Shows dependency issues
   ```

### How do I roll back if something breaks?

Use the restore command:

```bash
# List available snapshots
ls -lt ~/.backup/shellforge/

# Restore from specific snapshot
gz-shellforge restore --file ~/.zshrc --snapshot 2025-11-27_14-30-45

# Reload your shell
source ~/.zshrc
```

### The generated config is different from my original - is this normal?

Yes, some differences are expected:

**Normal differences:**
- ✅ Module headers added (`# --- module-name ---`)
- ✅ Generated metadata at top (timestamp, OS, module count)
- ✅ Module order changed (dependency-based sorting)
- ✅ OS-specific sections filtered out

**Unexpected differences:**
- ❌ Missing content (check with `--verbose` during migration)
- ❌ Syntax errors (validate your original file first)

Use `gz-shellforge diff` to review changes:
```bash
gz-shellforge diff ~/.zshrc ~/.zshrc.new --format side-by-side
```

---

## Advanced Usage

### Can I use Shellforge in CI/CD?

Yes! Example GitHub Actions workflow:

```yaml
name: Validate Shell Config
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Install Shellforge
        run: go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest
      - name: Validate
        run: gz-shellforge validate --verbose
      - name: Build for Mac
        run: gz-shellforge build --os Mac --dry-run
      - name: Build for Linux
        run: gz-shellforge build --os Linux --dry-run
```

### How do I share my modular config with my team?

1. **Initialize git repository**:
   ```bash
   cd ~/dotfiles
   git init
   git add manifest.yaml modules/
   git commit -m "Initial modular shell config"
   git remote add origin git@github.com:team/shell-config.git
   git push -u origin main
   ```

2. **Team members clone and use**:
   ```bash
   git clone git@github.com:team/shell-config.git ~/dotfiles
   cd ~/dotfiles
   gz-shellforge build --os Mac --output ~/.zshrc
   source ~/.zshrc
   ```

3. **Update workflow**:
   ```bash
   # Make changes
   vim modules/rc_post.d/team-aliases.sh

   # Commit and push
   git add modules/rc_post.d/team-aliases.sh
   git commit -m "Add team aliases"
   git push

   # Team members pull and rebuild
   git pull
   gz-shellforge build --os Mac --output ~/.zshrc
   source ~/.zshrc
   ```

### Can I have different configs for work and personal machines?

Yes, use separate manifest files:

```bash
# Work machine
gz-shellforge build -m manifest.work.yaml -o ~/.zshrc

# Personal machine
gz-shellforge build -m manifest.personal.yaml -o ~/.zshrc
```

Or use OS filtering and environment variables in your modules.

---

## Performance & Compatibility

### How fast is Shellforge compared to the Python version?

Shellforge (Go) is significantly faster:

| Metric | Python | Go | Improvement |
|--------|--------|----|----|
| Startup time | ~200ms | <10ms | **20x faster** |
| Build (10 modules) | ~300ms | <50ms | **6x faster** |
| Memory usage | ~80MB | <10MB | **8x lighter** |

### Does Shellforge work on Windows?

Not directly. Use **WSL (Windows Subsystem for Linux)** to run Shellforge on Windows.

### What platforms are supported?

- ✅ **macOS**: 10.15+ (Catalina and later)
- ✅ **Linux**: Ubuntu 20.04+, Debian 11+, Arch, Manjaro
- ⏳ **BSD**: FreeBSD 13+ (planned)
- ❌ **Windows**: Use WSL

---

## Getting Help

### Where can I find more documentation?

- **[Quick Start Guide](QUICK_START.md)** - Get started in 5 minutes
- **[README](README.md)** - Complete feature overview
- **[User Documentation](docs/user/)** - Detailed guides and examples
- **[Developer Documentation](docs/dev/)** - Architecture and contributing

### Where do I report bugs or request features?

- **GitHub Issues**: https://github.com/gizzahub/gzh-cli-shellforge/issues
- Include:
  - Shellforge version (`gz-shellforge --version`)
  - Operating system and shell version
  - Steps to reproduce
  - Expected vs actual behavior

### How can I contribute?

See [docs/dev/CONTRIBUTING.md](docs/dev/CONTRIBUTING.md) for:
- Development setup
- Code style guidelines
- Testing requirements
- Pull request process

---

## Still have questions?

- 📖 Check the [complete documentation](docs/user/)
- 💬 Open a [GitHub Discussion](https://github.com/gizzahub/gzh-cli-shellforge/discussions)
- 🐛 Report bugs via [GitHub Issues](https://github.com/gizzahub/gzh-cli-shellforge/issues)
