# Complete Workflow Guide

This guide demonstrates the complete end-to-end workflow for converting a monolithic shell configuration into a modular, maintainable structure using `gz-shellforge`.

## Workflow Overview

```
┌─────────────┐
│  Original   │
│   .zshrc    │
│ (monolithic)│
└──────┬──────┘
       │
       │ 1. migrate
       ▼
┌─────────────┐     ┌──────────────┐
│  manifest   │────▶│   modules/   │
│   .yaml     │     │  (modular)   │
└──────┬──────┘     └──────┬───────┘
       │                   │
       │ 2. build          │
       └──────────┬────────┘
                  ▼
         ┌────────────────┐
         │  Generated     │
         │  .zshrc.new    │
         └────────┬───────┘
                  │
                  │ 3. diff
                  ▼
         ┌────────────────┐
         │  Comparison    │
         │    Report      │
         └────────────────┘
```

## Prerequisites

- `gz-shellforge` installed (run `make install` from project root)
- An existing shell configuration file (.zshrc, .bashrc, etc.)

## Step-by-Step Workflow

### Step 1: Analyze Your Configuration (Optional)

Before migrating, you can analyze your RC file to see what sections will be detected:

```bash
gz-shellforge migrate ~/.zshrc --dry-run
```

**Output:**
```
Analyzing: /Users/you/.zshrc

Detected Sections:
  1. OS Detection              → init.d/00-os-detection.sh
  2. PATH Setup                → init.d/05-path-setup.sh
  3. Homebrew                  → init.d/10-homebrew.sh
  4. NVM Setup                 → rc_pre.d/nvm.sh
  5. Git Aliases               → rc_post.d/git-aliases.sh
  6. Helper Functions          → rc_post.d/helpers.sh

Would create 6 module files in 3 categories.
```

### Step 2: Migrate to Modular Structure

Convert your monolithic RC file into organized modules:

```bash
gz-shellforge migrate ~/.zshrc \
  --output-dir ~/shellforge/modules \
  --manifest ~/shellforge/manifest.yaml
```

**What happens:**
1. Analyzes your .zshrc file
2. Detects sections using header patterns (===, ---, # CAPS)
3. Categorizes each section:
   - `init.d/` - Early initialization (PATH, OS detection)
   - `rc_pre.d/` - Tool initialization (nvm, rbenv, pyenv)
   - `rc_post.d/` - Aliases, functions, customizations
4. Generates module files with proper headers
5. Creates manifest.yaml with dependencies and metadata

**Generated Structure:**
```
~/shellforge/
├── manifest.yaml
└── modules/
    ├── init.d/
    │   ├── 00-os-detection.sh
    │   ├── 05-path-setup.sh
    │   └── 10-homebrew.sh
    ├── rc_pre.d/
    │   └── nvm.sh
    └── rc_post.d/
        ├── git-aliases.sh
        └── helpers.sh
```

### Step 3: Review the Manifest

Examine the generated manifest.yaml:

```bash
cat ~/shellforge/manifest.yaml
```

**Example manifest.yaml:**
```yaml
modules:
  - name: os-detection
    file: init.d/00-os-detection.sh
    description: Detect operating system and set MACHINE variable
    os: [Mac, Linux]

  - name: path-setup
    file: init.d/05-path-setup.sh
    description: PATH Setup
    requires:
      - os-detection
    os: [Mac, Linux]

  - name: homebrew
    file: init.d/10-homebrew.sh
    description: Homebrew
    requires:
      - os-detection
    os: [Mac]

  - name: nvm
    file: rc_pre.d/nvm.sh
    description: NVM Setup
    requires:
      - os-detection
    os: [Mac, Linux]
```

### Step 4: Validate Configuration

Check that modules are valid and have no circular dependencies:

```bash
gz-shellforge validate \
  --manifest ~/shellforge/manifest.yaml \
  --config-dir ~/shellforge/modules
```

**Output:**
```
✓ Validation successful!
  Modules: 6
  Manifest: /Users/you/shellforge/manifest.yaml
```

### Step 5: List Modules

View all modules and their dependencies:

```bash
gz-shellforge list \
  --manifest ~/shellforge/manifest.yaml \
  --config-dir ~/shellforge/modules
```

**Output:**
```
Modules (6)
Manifest: /Users/you/shellforge/manifest.yaml

1. os-detection [Mac, Linux]
   Detect operating system and set MACHINE variable

2. path-setup [Mac, Linux]
   PATH Setup
   → os-detection

3. homebrew [Mac]
   Homebrew
   → os-detection
```

### Step 6: Build for Your OS

Generate a new .zshrc file optimized for your operating system:

**For macOS:**
```bash
gz-shellforge build \
  --manifest ~/shellforge/manifest.yaml \
  --config-dir ~/shellforge/modules \
  --os Mac \
  --output ~/.zshrc.new
```

**For Linux:**
```bash
gz-shellforge build \
  --manifest ~/shellforge/manifest.yaml \
  --config-dir ~/shellforge/modules \
  --os Linux \
  --output ~/.zshrc.new
```

**What happens:**
1. Reads manifest.yaml
2. Filters modules by target OS
3. Resolves dependencies (topological sort)
4. Loads modules in correct order
5. Generates single .zshrc file with headers

**Generated .zshrc.new:**
```bash
# Generated by shellforge
# OS: Mac
# Modules: 5
# Generated at: 2025-11-28T20:00:00+09:00

# --- os-detection ---
# Detect operating system and set MACHINE variable
#!/bin/bash
case "$(uname -s)" in
  Darwin)
    export MACHINE="Mac"
    ;;
  ...
esac

# --- path-setup ---
# PATH Setup
#!/bin/bash
export PATH="/usr/local/bin:$PATH"

# --- homebrew ---
# Homebrew
#!/bin/bash
if [[ "$MACHINE" == "Mac" ]]; then
    export PATH="/opt/homebrew/bin:$PATH"
fi

...
```

### Step 7: Preview Build (Dry Run)

Preview what would be generated without creating a file:

```bash
gz-shellforge build \
  --manifest ~/shellforge/manifest.yaml \
  --config-dir ~/shellforge/modules \
  --os Mac \
  --dry-run
```

### Step 8: Compare Original vs Generated

Compare your original .zshrc with the generated version:

**Summary format (statistics only):**
```bash
gz-shellforge diff ~/.zshrc ~/.zshrc.new --format summary
```

**Output:**
```
File Comparison
Original:  /Users/you/.zshrc
Generated: /Users/you/.zshrc.new

Statistics:
  Total lines:      45
  Added lines:      12 (26.7%)
  Removed lines:    0 (0.0%)
  Modified lines:   0 (0.0%)
  Unchanged lines:  33 (73.3%)

Total changes: 12 lines (26.7%)

Status: Files are different
```

**Unified diff format (git diff style):**
```bash
gz-shellforge diff ~/.zshrc ~/.zshrc.new --format unified
```

**Side-by-side format (visual comparison):**
```bash
gz-shellforge diff ~/.zshrc ~/.zshrc.new --format side-by-side
```

### Step 9: Test the New Configuration

Before replacing your original .zshrc:

```bash
# Source the new configuration in a subshell
zsh -c 'source ~/.zshrc.new; echo "Test: $MACHINE"'

# Or test in a new terminal window
zsh --init-file ~/.zshrc.new
```

### Step 10: Deploy the New Configuration

Once you've verified everything works:

**Option A: Backup and replace**
```bash
# Create backup
gz-shellforge backup --file ~/.zshrc --message "Before modular migration"

# Replace with new version
mv ~/.zshrc.new ~/.zshrc

# Source the new configuration
source ~/.zshrc
```

**Option B: Symlink approach**
```bash
# Keep modules managed separately
ln -sf ~/shellforge/output/.zshrc ~/.zshrc
```

## Advanced Workflows

### Multi-OS Setup

Maintain configurations for multiple operating systems:

```bash
# Build for Mac
gz-shellforge build --os Mac --output ~/.zshrc.mac

# Build for Linux
gz-shellforge build --os Linux --output ~/.zshrc.linux

# Deploy based on OS
case "$(uname -s)" in
  Darwin) cp ~/.zshrc.mac ~/.zshrc ;;
  Linux)  cp ~/.zshrc.linux ~/.zshrc ;;
esac
```

### Version Control Integration

```bash
cd ~/shellforge

# Initialize git repository
git init
git add manifest.yaml modules/
git commit -m "Initial modular shell configuration"

# Add remote and push
git remote add origin git@github.com:username/my-shell-config.git
git push -u origin main
```

### Continuous Updates

```bash
# Edit a module
vim ~/shellforge/modules/rc_post.d/git-aliases.sh

# Rebuild
gz-shellforge build --os Mac --output ~/.zshrc.new

# Compare changes
gz-shellforge diff ~/.zshrc ~/.zshrc.new

# Deploy if satisfied
mv ~/.zshrc.new ~/.zshrc && source ~/.zshrc
```

### Template-Based Module Creation

Create new modules from templates:

```bash
# Create a new alias module
gz-shellforge template generate alias \
  --config-dir ~/shellforge/modules \
  -f name=docker-aliases \
  -f description="Docker shortcuts" \
  -f aliases="alias dps='docker ps',alias dex='docker exec -it'"

# Update manifest
vim ~/shellforge/manifest.yaml  # Add the new module

# Rebuild
gz-shellforge build --os Mac --output ~/.zshrc.new
```

## Automated Demo

Run the complete workflow demo script:

```bash
cd examples/
./workflow-demo.sh
```

This script automatically:
1. Creates a sample .zshrc
2. Migrates it to modules
3. Builds configurations for Mac and Linux
4. Compares original with generated
5. Lists and validates modules

## Best Practices

1. **Start Small**: Migrate a copy of your .zshrc first, not the original
2. **Review Categorization**: Check that modules are in the right directories
3. **Test Thoroughly**: Source the new .zshrc in a test environment first
4. **Version Control**: Keep your modules in git for easy rollback
5. **Incremental Migration**: You don't have to migrate everything at once
6. **Document Dependencies**: Add clear descriptions in manifest.yaml
7. **Use Dry-Run**: Always preview builds before deploying

## Troubleshooting

### Circular Dependencies Detected

```bash
Error: circular dependency detected: A → B → C → A
```

**Solution:** Review your `requires` fields in manifest.yaml and break the cycle.

### Module File Not Found

```bash
Error: module file not found: modules/init.d/missing.sh
```

**Solution:** Ensure the file path in manifest.yaml matches the actual file location.

### Section Not Detected During Migration

If a section wasn't detected, manually create a module:

```bash
# Create the module file
mkdir -p ~/shellforge/modules/rc_post.d
vim ~/shellforge/modules/rc_post.d/custom.sh

# Add to manifest.yaml
vim ~/shellforge/manifest.yaml
```

## Getting Help

```bash
# General help
gz-shellforge --help

# Command-specific help
gz-shellforge migrate --help
gz-shellforge build --help
gz-shellforge diff --help

# Examples
gz-shellforge build --help | grep -A 10 Examples
```

## Next Steps

- Explore the [Template System](../README.md#template---generate-module-from-template) for creating new modules
- Set up [Backup/Restore](../README.md#backup---create-backup) for configuration versioning
- Learn about [Advanced Features](../README.md#features) in the main README

---

**Happy shell config management!** 🚀
