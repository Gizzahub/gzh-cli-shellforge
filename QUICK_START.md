# Quick Start Guide

> Get started with Shellforge in 5 minutes

This guide will help you convert your monolithic shell configuration into a modular, maintainable structure.

## Prerequisites

- Go 1.21 or later (for building from source)
- Git (for backup features)
- An existing shell configuration file (.zshrc, .bashrc, etc.)

## Installation

### Option 1: Install from source

```bash
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest
```

### Option 2: Build locally

```bash
git clone https://github.com/gizzahub/gzh-cli-shellforge.git
cd gzh-cli-shellforge
make install
```

### Verify installation

```bash
gz-shellforge --version
```

## 5-Minute Tutorial

### Step 1: Backup your current configuration

```bash
gz-shellforge backup --file ~/.zshrc --message "Before migration"
```

### Step 2: Migrate to modular structure

```bash
cd ~/dotfiles  # or your preferred location
gz-shellforge migrate ~/.zshrc
```

This creates:
- `manifest.yaml` - Module configuration
- `modules/` - Individual shell modules organized by category

### Step 3: Validate the migration

```bash
gz-shellforge validate
```

### Step 4: Build for your OS

```bash
gz-shellforge build --os Mac --output ~/.zshrc.new
```

Replace `Mac` with `Linux` if you're on Linux.

### Step 5: Compare and deploy

```bash
# Compare changes
gz-shellforge diff ~/.zshrc ~/.zshrc.new

# Test the new configuration
source ~/.zshrc.new

# If everything works, deploy it
mv ~/.zshrc.new ~/.zshrc
source ~/.zshrc
```

## What's Next?

### Learn more commands

```bash
# List all modules
gz-shellforge list

# List Mac-specific modules
gz-shellforge list --filter Mac

# Add a new module from template
gz-shellforge template list
gz-shellforge template generate alias my-aliases -f aliases='alias ll="ls -la"'
```

### Explore documentation

- **[Complete Command Reference](docs/user/20-commands.md)** - All commands and options
- **[Workflow Guide](docs/user/30-workflows.md)** - Step-by-step workflows
- **[Command Reference](docs/user/40-command-reference.md)** - All commands with examples
- **[FAQ](FAQ.md)** - Frequently asked questions

### Try the examples

```bash
cd examples/
gz-shellforge validate --verbose
gz-shellforge build --os Mac --dry-run
./workflow-demo.sh  # Automated demonstration
```

## Common Tasks

### Update a module

```bash
# Edit the module file
vim modules/rc_post.d/my-aliases.sh

# Rebuild and test
gz-shellforge build --os Mac --output ~/.zshrc.new
gz-shellforge diff ~/.zshrc ~/.zshrc.new
mv ~/.zshrc.new ~/.zshrc && source ~/.zshrc
```

### Roll back to previous version

```bash
# List available snapshots
ls -lt ~/.backup/shellforge/

# Restore specific snapshot
gz-shellforge restore --file ~/.zshrc --snapshot 2025-11-27_14-30-45
source ~/.zshrc
```

### Share with your team

```bash
cd ~/dotfiles
git init
git add manifest.yaml modules/
git commit -m "Initial modular shell configuration"
git remote add origin git@github.com:yourteam/shell-config.git
git push -u origin main
```

Team members can then:

```bash
git clone git@github.com:yourteam/shell-config.git ~/dotfiles
cd ~/dotfiles
gz-shellforge build --os Mac --output ~/.zshrc  # or Linux
source ~/.zshrc
```

## Troubleshooting

### Command not found

Make sure `$GOPATH/bin` is in your PATH:

```bash
echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### Circular dependency error

Edit `manifest.yaml` and remove one dependency from the cycle. Use `gz-shellforge validate --verbose` to see details.

### Module file not found

Ensure the `file` path in `manifest.yaml` matches the actual file location relative to the config directory.

## Get Help

```bash
# General help
gz-shellforge --help

# Command-specific help
gz-shellforge build --help
gz-shellforge migrate --help
```

For more detailed information, see the [complete documentation](README.md).

---

**You're all set!** Your shell configuration is now modular, version-controlled, and maintainable. 🚀
