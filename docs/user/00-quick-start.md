# Quick Start Guide

Get started with Shellforge in 5 minutes.

---

## What is Shellforge?

Shellforge transforms your monolithic shell configuration (`.zshrc`, `.bashrc`) into organized, modular files with automatic dependency resolution and OS-specific filtering.

**Why use it?**
- ✅ Automatic dependency resolution - modules load in correct order
- ✅ OS-specific filtering - Mac and Linux modules managed separately
- ✅ Modular structure - easy to maintain and share
- ✅ Version control friendly - track changes module by module

---

## Installation

### Option 1: Install via Go (Recommended)

```bash
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest
```

### Option 2: Build from Source

```bash
git clone https://github.com/gizzahub/gzh-cli-shellforge.git
cd gzh-cli-shellforge
make install
```

### Verify Installation

```bash
gz-shellforge --version
# Should output: shellforge version 0.2.0-alpha
```

---

## 5-Minute Tutorial

### Step 1: Backup Your Current Config (30 seconds)

```bash
# Backup your .zshrc or .bashrc
gz-shellforge backup --file ~/.zshrc --message "Before migration"
```

### Step 2: Migrate to Modular Structure (1 minute)

```bash
# Create a working directory
mkdir ~/shellforge
cd ~/shellforge

# Migrate your monolithic config to modules
gz-shellforge migrate ~/.zshrc
```

**What happens:**
- Detects sections in your `.zshrc`
- Creates organized modules in `modules/` directory
- Generates `manifest.yaml` with dependencies

**Generated structure:**
```
~/shellforge/
├── manifest.yaml          # Module definitions and dependencies
└── modules/
    ├── init.d/            # Early initialization (PATH, OS detection)
    ├── rc_pre.d/          # Tool setup (nvm, rbenv, conda)
    └── rc_post.d/         # Aliases and functions
```

### Step 3: Validate Configuration (30 seconds)

```bash
# Check for any issues
gz-shellforge validate --verbose
```

**Expected output:**
```
✓ Validation successful!
  Modules: 6
  Manifest: /Users/you/shellforge/manifest.yaml
```

### Step 4: Build Your Shell Config (1 minute)

```bash
# Build for your OS (Mac or Linux)
gz-shellforge build --os Mac --output ~/.zshrc.new
```

**What happens:**
- Filters modules by target OS
- Resolves dependencies (topological sort)
- Generates unified `.zshrc` in correct load order

### Step 5: Compare and Deploy (1 minute)

```bash
# Compare original vs generated
gz-shellforge diff ~/.zshrc ~/.zshrc.new --format summary

# If satisfied, deploy
mv ~/.zshrc.new ~/.zshrc
source ~/.zshrc
```

### Step 6: Test It Works (1 minute)

```bash
# Test in a new shell
zsh -c 'echo "Shell config loaded: $SHELL"'

# Verify your aliases and functions work
# (try some of your custom commands)
```

**🎉 Done!** Your shell config is now modular and maintainable.

---

## What You Just Learned

✅ **Backup**: Create versioned backups before changes
✅ **Migrate**: Convert monolithic config to modular structure
✅ **Validate**: Check configuration for errors
✅ **Build**: Generate unified config for specific OS
✅ **Compare**: Verify changes before deploying
✅ **Deploy**: Apply new configuration

---

## Common First Tasks

### View All Your Modules

```bash
gz-shellforge list --verbose
```

### Add a New Module

```bash
# Generate from template
gz-shellforge template generate alias my-aliases \
  -f aliases="alias k='kubectl',alias d='docker'"

# Or create manually
vim modules/rc_post.d/my-custom.sh

# Add to manifest.yaml
# Then rebuild
gz-shellforge build --os Mac --output ~/.zshrc
```

### Build for Different OS

```bash
# For macOS
gz-shellforge build --os Mac --output ~/.zshrc.mac

# For Linux
gz-shellforge build --os Linux --output ~/.zshrc.linux
```

### Preview Without Saving

```bash
gz-shellforge build --os Mac --dry-run
```

---

## Troubleshooting

### Command Not Found

```bash
# Ensure Go bin is in PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Or check installation
which gz-shellforge
```

### Circular Dependency Error

```
Error: circular dependency detected: A → B → C → A
```

**Solution:** Edit `manifest.yaml` and remove the circular `requires` dependency.

### Module File Not Found

```
Error: module file not found: modules/init.d/missing.sh
```

**Solution:** Ensure the file path in `manifest.yaml` matches the actual file location.

### Migration Didn't Detect Section

If a section wasn't auto-detected, manually create the module:

```bash
# Create module file
vim modules/rc_post.d/custom.sh

# Add to manifest.yaml
vim manifest.yaml
```

---

## Next Steps

### Learn More Commands

- **[Command Reference](40-command-reference.md)** - All commands with examples
- **[Workflows Guide](30-workflows.md)** - Complete workflow examples
- **[Troubleshooting](60-troubleshooting.md)** - Common issues and solutions

### Advanced Features

- **Templates**: Generate modules from predefined templates
- **Backup/Restore**: Version control your shell configs
- **Multi-OS**: Maintain configs for Mac and Linux
- **CI/CD**: Automate config deployment

See **[Advanced Usage Guide](50-advanced-usage.md)** for details.

---

## Getting Help

```bash
# General help
gz-shellforge --help

# Command-specific help
gz-shellforge migrate --help
gz-shellforge build --help

# List available commands
gz-shellforge help
```

**Need more help?**
- [Documentation](README.md)
- [GitHub Issues](https://github.com/gizzahub/gzh-cli-shellforge/issues)
- [Examples](../../examples/)

---

**Happy shell config management!** 🚀
