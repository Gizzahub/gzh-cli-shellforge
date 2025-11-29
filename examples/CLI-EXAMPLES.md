# CLI Usage Examples

Quick reference guide for common `gz-shellforge` command usage patterns.

## Table of Contents

- [Migrate Command](#migrate-command)
- [Build Command](#build-command)
- [Validate Command](#validate-command)
- [List Command](#list-command)
- [Diff Command](#diff-command)
- [Template Command](#template-command)
- [Backup Command](#backup-command)
- [Restore Command](#restore-command)
- [Cleanup Command](#cleanup-command)
- [Real-World Scenarios](#real-world-scenarios)

---

## Migrate Command

Convert monolithic RC files to modular structure.

### Basic Migration

```bash
# Migrate your .zshrc to modular structure
gz-shellforge migrate ~/.zshrc

# Migrate with custom output directory
gz-shellforge migrate ~/.zshrc --output-dir ~/my-modules

# Migrate with custom manifest path
gz-shellforge migrate ~/.bashrc --manifest ~/my-config/manifest.yaml
```

### Dry Run (Analysis Only)

```bash
# Analyze what would be created without creating files
gz-shellforge migrate ~/.zshrc --dry-run

# Analyze with verbose output
gz-shellforge migrate ~/.zshrc --dry-run --verbose
```

### Complete Examples

```bash
# Example 1: Migrate .zshrc to specific directory
gz-shellforge migrate ~/.zshrc \
  --output-dir ~/shellforge/modules \
  --manifest ~/shellforge/manifest.yaml

# Example 2: Analyze .bashrc before migrating
gz-shellforge migrate ~/.bashrc --dry-run -v

# Example 3: Migrate with default paths
cd ~/my-shell-config
gz-shellforge migrate ~/.zshrc
# Creates: ./modules/ and ./manifest.yaml
```

---

## Build Command

Generate shell configuration from modules.

### Basic Build

```bash
# Build for Mac
gz-shellforge build --os Mac --output ~/.zshrc

# Build for Linux
gz-shellforge build --os Linux --output ~/.bashrc

# Build with custom paths
gz-shellforge build \
  --manifest ~/config/manifest.yaml \
  --config-dir ~/config/modules \
  --os Mac \
  --output ~/.zshrc
```

### Dry Run (Preview Only)

```bash
# Preview what would be generated for Mac
gz-shellforge build --os Mac --dry-run

# Preview with verbose output
gz-shellforge build --os Mac --dry-run --verbose

# Preview specific manifest
gz-shellforge build \
  --manifest custom-manifest.yaml \
  --config-dir custom-modules \
  --os Linux \
  --dry-run
```

### Short Flags

```bash
# Using short flags
gz-shellforge build -m manifest.yaml -c modules -o ~/.zshrc --os Mac

# Dry run with verbose
gz-shellforge build --os Mac --dry-run -v
```

### Complete Examples

```bash
# Example 1: Build and save for Mac
gz-shellforge build \
  --manifest manifest.yaml \
  --config-dir modules \
  --os Mac \
  --output ~/.zshrc.new

# Example 2: Preview Linux build
gz-shellforge build --os Linux --dry-run -v

# Example 3: Build from custom location
gz-shellforge build \
  -m ~/dotfiles/shell-manifest.yaml \
  -c ~/dotfiles/shell-modules \
  -o ~/.zshrc \
  --os Mac

# Example 4: Test before deploying
gz-shellforge build --os Mac --dry-run > preview.txt
less preview.txt
```

---

## Validate Command

Check manifest and modules for errors.

### Basic Validation

```bash
# Validate with default paths
gz-shellforge validate

# Validate with custom paths
gz-shellforge validate \
  --manifest custom-manifest.yaml \
  --config-dir custom-modules
```

### Verbose Validation

```bash
# Show detailed validation steps
gz-shellforge validate --verbose

# Validate specific manifest with details
gz-shellforge validate \
  --manifest ~/config/manifest.yaml \
  --config-dir ~/config/modules \
  --verbose
```

### Complete Examples

```bash
# Example 1: Quick validation check
cd ~/shellforge
gz-shellforge validate

# Example 2: Detailed validation report
gz-shellforge validate -v

# Example 3: Validate before build
gz-shellforge validate && \
gz-shellforge build --os Mac --output ~/.zshrc

# Example 4: Validate custom configuration
gz-shellforge validate \
  -m ~/dotfiles/manifest.yaml \
  -c ~/dotfiles/modules \
  -v
```

---

## List Command

Display modules and their dependencies.

### Basic Listing

```bash
# List all modules
gz-shellforge list

# List with custom paths
gz-shellforge list \
  --manifest custom-manifest.yaml \
  --config-dir custom-modules
```

### OS Filtering

```bash
# List only Mac-compatible modules
gz-shellforge list --filter Mac

# List only Linux-compatible modules
gz-shellforge list --filter Linux

# Case-insensitive filtering
gz-shellforge list --filter mac
```

### Verbose Listing

```bash
# Show detailed information
gz-shellforge list --verbose

# Show Mac modules with details
gz-shellforge list --filter Mac --verbose
```

### Complete Examples

```bash
# Example 1: Quick module overview
gz-shellforge list

# Example 2: Check Mac-specific modules
gz-shellforge list --filter Mac

# Example 3: Detailed module information
gz-shellforge list -v

# Example 4: Find modules for Linux with details
gz-shellforge list --filter Linux --verbose

# Example 5: List from custom manifest
gz-shellforge list \
  -m ~/config/manifest.yaml \
  -c ~/config/modules \
  -v
```

---

## Diff Command

Compare original and generated configurations.

### Summary Format

```bash
# Show statistics only
gz-shellforge diff ~/.zshrc ~/.zshrc.new

# Equivalent explicit format
gz-shellforge diff ~/.zshrc ~/.zshrc.new --format summary
```

### Unified Format (Git Style)

```bash
# Show git-style diff
gz-shellforge diff ~/.zshrc ~/.zshrc.new --format unified

# With verbose output
gz-shellforge diff ~/.zshrc ~/.zshrc.new --format unified -v
```

### Context Format

```bash
# Show context diff
gz-shellforge diff ~/.zshrc ~/.zshrc.new --format context
```

### Side-by-Side Format

```bash
# Show side-by-side comparison
gz-shellforge diff ~/.zshrc ~/.zshrc.new --format side-by-side

# Pipe to less for scrolling
gz-shellforge diff ~/.zshrc ~/.zshrc.new --format side-by-side | less
```

### Complete Examples

```bash
# Example 1: Quick comparison
gz-shellforge diff ~/.zshrc ~/.zshrc.new

# Example 2: Detailed git-style diff
gz-shellforge diff ~/.zshrc ~/.zshrc.new -f unified > changes.diff

# Example 3: Visual comparison
gz-shellforge diff ~/.zshrc ~/.zshrc.new -f side-by-side | less

# Example 4: All formats for review
for fmt in summary unified context side-by-side; do
  echo "=== Format: $fmt ==="
  gz-shellforge diff ~/.zshrc ~/.zshrc.new -f $fmt
  echo ""
done

# Example 5: Compare before deploying
gz-shellforge diff ~/.zshrc ~/.zshrc.new -f summary
if [ $? -eq 0 ]; then
  echo "Files are identical, no changes needed"
else
  echo "Files differ, review changes above"
fi
```

---

## Template Command

Generate modules from templates.

### List Available Templates

```bash
# Show all available templates
gz-shellforge template list

# Show templates with verbose descriptions
gz-shellforge template list -v
```

### Generate from Templates

```bash
# Generate PATH module
gz-shellforge template generate path \
  --config-dir modules \
  -f name=custom-path \
  -f description="Custom PATH setup" \
  -f path_entry="/opt/custom/bin"

# Generate environment variable module
gz-shellforge template generate env \
  --config-dir modules \
  -f name=api-keys \
  -f description="API key configuration" \
  -f var_name=API_KEY \
  -f var_value="your-key-here"

# Generate alias module
gz-shellforge template generate alias \
  --config-dir modules \
  -f name=docker-aliases \
  -f description="Docker shortcuts" \
  -f aliases="alias dps='docker ps',alias dex='docker exec -it'"
```

### With Dependencies

```bash
# Generate with required dependencies
gz-shellforge template generate tool-init \
  --config-dir modules \
  -f name=pyenv \
  -f description="Python version manager" \
  -f tool_name=pyenv \
  -f init_command='eval "$(pyenv init -)"' \
  -r os-detection
```

### Complete Examples

```bash
# Example 1: Create custom PATH module
gz-shellforge template generate path \
  -c ~/shellforge/modules \
  -f name=golang-path \
  -f description="Go language PATH setup" \
  -f path_entry="$HOME/go/bin" \
  -r os-detection

# Example 2: Create API keys module
gz-shellforge template generate env \
  -c modules \
  -f name=github-token \
  -f description="GitHub API token" \
  -f var_name=GITHUB_TOKEN \
  -f var_value="\${GITHUB_TOKEN:-default}"

# Example 3: Create tool initialization
gz-shellforge template generate tool-init \
  -c modules \
  -f name=nvm \
  -f description="Node Version Manager" \
  -f tool_name=nvm \
  -f init_command='[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"' \
  -r os-detection

# Example 4: OS-specific module
gz-shellforge template generate os-specific \
  -c modules \
  -f name=mac-homebrew \
  -f description="Homebrew for macOS" \
  -f os_name=Mac \
  -f content='export PATH="/opt/homebrew/bin:$PATH"' \
  -r os-detection
```

---

## Backup Command

Create timestamped backups of shell configurations.

### Basic Backup

```bash
# Backup .zshrc
gz-shellforge backup --file ~/.zshrc

# Backup with custom message
gz-shellforge backup \
  --file ~/.zshrc \
  --message "Before major refactoring"

# Backup to custom directory
gz-shellforge backup \
  --file ~/.zshrc \
  --backup-dir ~/my-backups
```

### Without Git

```bash
# Backup without git versioning
gz-shellforge backup --file ~/.zshrc --no-git
```

### Verbose Output

```bash
# Show detailed backup process
gz-shellforge backup --file ~/.zshrc --verbose

# With custom message and verbose
gz-shellforge backup \
  --file ~/.zshrc \
  --message "Pre-migration backup" \
  -v
```

### Complete Examples

```bash
# Example 1: Quick backup before changes
gz-shellforge backup -f ~/.zshrc

# Example 2: Backup with descriptive message
gz-shellforge backup \
  -f ~/.zshrc \
  -m "Before switching to modular configuration"

# Example 3: Backup multiple files
for file in ~/.zshrc ~/.bashrc ~/.bash_profile; do
  if [ -f "$file" ]; then
    gz-shellforge backup -f "$file" -m "Backup before migration"
  fi
done

# Example 4: Backup to external drive
gz-shellforge backup \
  -f ~/.zshrc \
  --backup-dir /Volumes/Backup/shellforge \
  -v
```

---

## Restore Command

Restore shell configuration from backup snapshots.

### Basic Restore

```bash
# Restore from specific snapshot
gz-shellforge restore \
  --file ~/.zshrc \
  --snapshot 2025-11-28_14-30-45

# Restore with short flags
gz-shellforge restore \
  -f ~/.zshrc \
  -s 2025-11-28_14-30-45
```

### Dry Run (Preview)

```bash
# Preview restore without executing
gz-shellforge restore \
  --file ~/.zshrc \
  --snapshot 2025-11-28_14-30-45 \
  --dry-run

# Preview with verbose output
gz-shellforge restore \
  -f ~/.zshrc \
  -s 2025-11-28_14-30-45 \
  --dry-run -v
```

### Custom Backup Directory

```bash
# Restore from custom backup location
gz-shellforge restore \
  -f ~/.zshrc \
  -s 2025-11-28_14-30-45 \
  --backup-dir ~/my-backups
```

### Complete Examples

```bash
# Example 1: List available snapshots first
ls -lh ~/.backup/shellforge/

# Example 2: Restore specific version
gz-shellforge restore \
  -f ~/.zshrc \
  -s 2025-11-28_14-30-45

# Example 3: Preview before restoring
gz-shellforge restore \
  -f ~/.zshrc \
  -s 2025-11-28_14-30-45 \
  --dry-run -v

# Example 4: Restore and verify
gz-shellforge restore -f ~/.zshrc -s 2025-11-28_14-30-45
source ~/.zshrc
echo "Restore successful, shell reloaded"
```

---

## Cleanup Command

Remove old backup snapshots.

### By Count

```bash
# Keep last 10 snapshots
gz-shellforge cleanup --file ~/.zshrc --keep-count 10

# Keep last 5 snapshots
gz-shellforge cleanup -f ~/.zshrc --keep-count 5
```

### By Age

```bash
# Keep snapshots from last 30 days
gz-shellforge cleanup --file ~/.zshrc --keep-days 30

# Keep snapshots from last 7 days
gz-shellforge cleanup -f ~/.zshrc --keep-days 7
```

### Combined Policy

```bash
# Keep last 10 OR from last 30 days (union)
gz-shellforge cleanup \
  --file ~/.zshrc \
  --keep-count 10 \
  --keep-days 30
```

### Dry Run (Preview)

```bash
# Preview what would be deleted
gz-shellforge cleanup \
  --file ~/.zshrc \
  --keep-count 10 \
  --dry-run

# Preview with verbose output
gz-shellforge cleanup \
  -f ~/.zshrc \
  --keep-count 5 \
  --keep-days 14 \
  --dry-run -v
```

### Complete Examples

```bash
# Example 1: Regular cleanup (keep last 10)
gz-shellforge cleanup -f ~/.zshrc --keep-count 10

# Example 2: Aggressive cleanup (keep last 3)
gz-shellforge cleanup -f ~/.zshrc --keep-count 3

# Example 3: Time-based cleanup
gz-shellforge cleanup -f ~/.zshrc --keep-days 30

# Example 4: Preview before cleanup
gz-shellforge cleanup \
  -f ~/.zshrc \
  --keep-count 10 \
  --keep-days 30 \
  --dry-run -v

# Example 5: Cleanup all RC files
for file in ~/.zshrc ~/.bashrc ~/.bash_profile; do
  if [ -f "$file" ]; then
    gz-shellforge cleanup -f "$file" --keep-count 10
  fi
done
```

---

## Real-World Scenarios

### Scenario 1: Initial Setup

```bash
# 1. Backup your current .zshrc
gz-shellforge backup -f ~/.zshrc -m "Before modular migration"

# 2. Migrate to modular structure
gz-shellforge migrate ~/.zshrc \
  --output-dir ~/shellforge/modules \
  --manifest ~/shellforge/manifest.yaml

# 3. Validate the migration
cd ~/shellforge
gz-shellforge validate -v

# 4. Build for your OS
gz-shellforge build --os Mac --output ~/.zshrc.new

# 5. Compare changes
gz-shellforge diff ~/.zshrc ~/.zshrc.new -f summary

# 6. Test the new config
zsh --init-file ~/.zshrc.new

# 7. Deploy if satisfied
mv ~/.zshrc.new ~/.zshrc
source ~/.zshrc
```

### Scenario 2: Adding a New Module

```bash
# 1. Create new module from template
gz-shellforge template generate alias \
  -c ~/shellforge/modules \
  -f name=kubernetes-aliases \
  -f description="Kubernetes shortcuts" \
  -f aliases="alias k='kubectl',alias kgp='kubectl get pods'"

# 2. Edit manifest to add the new module
vim ~/shellforge/manifest.yaml

# 3. Validate configuration
cd ~/shellforge
gz-shellforge validate

# 4. Build and preview
gz-shellforge build --os Mac --dry-run | grep -A 5 "kubernetes"

# 5. Build to new file
gz-shellforge build --os Mac --output ~/.zshrc.new

# 6. Compare and deploy
gz-shellforge diff ~/.zshrc ~/.zshrc.new -f summary
mv ~/.zshrc.new ~/.zshrc && source ~/.zshrc
```

### Scenario 3: Multi-OS Workflow

```bash
# Setup once
cd ~/dotfiles
gz-shellforge migrate ~/.zshrc

# Build for Mac
gz-shellforge build --os Mac --output ~/.zshrc.mac

# Build for Linux
gz-shellforge build --os Linux --output ~/.zshrc.linux

# Deploy based on OS (add to shell startup)
case "$(uname -s)" in
  Darwin)
    [ -f ~/.zshrc.mac ] && ln -sf ~/.zshrc.mac ~/.zshrc
    ;;
  Linux)
    [ -f ~/.zshrc.linux ] && ln -sf ~/.zshrc.linux ~/.zshrc
    ;;
esac
```

### Scenario 4: Rollback After Issues

```bash
# Something went wrong, rollback!

# 1. List available snapshots
ls -lt ~/.backup/shellforge/ | head -10

# 2. Restore previous version
gz-shellforge restore \
  -f ~/.zshrc \
  -s 2025-11-28_12-00-00

# 3. Verify restoration
source ~/.zshrc
echo "Rolled back successfully"
```

### Scenario 5: Regular Maintenance

```bash
#!/bin/bash
# maintenance.sh - Regular shellforge maintenance

cd ~/shellforge

echo "=== Validating configuration ==="
gz-shellforge validate -v

echo ""
echo "=== Rebuilding configurations ==="
gz-shellforge build --os Mac --output ~/.zshrc.new

echo ""
echo "=== Comparing changes ==="
if gz-shellforge diff ~/.zshrc ~/.zshrc.new -f summary; then
  echo "No changes detected"
else
  echo "Changes detected, review above"
  read -p "Deploy changes? (y/n) " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    mv ~/.zshrc.new ~/.zshrc
    source ~/.zshrc
    echo "Deployed successfully"
  fi
fi

echo ""
echo "=== Cleaning up old backups ==="
gz-shellforge cleanup -f ~/.zshrc --keep-count 10 --keep-days 30 -v

echo ""
echo "=== Maintenance complete ==="
```

### Scenario 6: Team Sharing

```bash
# Share your modular configuration with team

# 1. Setup git repository
cd ~/shellforge
git init
git add manifest.yaml modules/
git commit -m "Initial modular shell configuration"
git remote add origin git@github.com:team/shell-config.git
git push -u origin main

# Team members clone and use:
git clone git@github.com:team/shell-config.git ~/shellforge
cd ~/shellforge

# Each member builds for their OS
gz-shellforge build --os Mac --output ~/.zshrc  # or Linux
source ~/.zshrc

# Update and share changes
vim modules/rc_post.d/team-aliases.sh
git add modules/rc_post.d/team-aliases.sh
git commit -m "Add team aliases"
git push

# Team members pull updates
git pull
gz-shellforge build --os Mac --output ~/.zshrc
source ~/.zshrc
```

---

## Tips and Tricks

### Quick Commands with Aliases

Add to your shell configuration:

```bash
# Shellforge aliases
alias sf='gz-shellforge'
alias sfb='gz-shellforge build --os Mac --output ~/.zshrc'
alias sfv='gz-shellforge validate -v'
alias sfl='gz-shellforge list'
alias sfbak='gz-shellforge backup -f ~/.zshrc'
alias sfdiff='gz-shellforge diff ~/.zshrc ~/.zshrc.new'

# Quick rebuild and reload
alias sfreload='sfb && source ~/.zshrc'

# Safe deploy with backup
alias sfdeploy='sfbak -m "Pre-deploy backup" && sfb && source ~/.zshrc'
```

### Chaining Commands

```bash
# Validate, build, diff in one line
gz-shellforge validate && \
gz-shellforge build --os Mac --output ~/.zshrc.new && \
gz-shellforge diff ~/.zshrc ~/.zshrc.new

# Backup, rebuild, and deploy
gz-shellforge backup -f ~/.zshrc -m "Auto backup" && \
gz-shellforge build --os Mac --output ~/.zshrc && \
source ~/.zshrc
```

### Finding Help

```bash
# General help
gz-shellforge --help

# Command-specific help
gz-shellforge migrate --help
gz-shellforge build --help

# Version information
gz-shellforge --version
```

---

For more detailed workflow guides, see [WORKFLOW.md](WORKFLOW.md).
