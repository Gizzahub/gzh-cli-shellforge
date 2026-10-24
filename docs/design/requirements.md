# Functional Requirements: Shellforge Go Implementation

**Version**: 1.0
**Status**: Draft
**Last Updated**: 2025-11-27

______________________________________________________________________

## Overview

This document specifies the detailed functional and technical requirements for the Go implementation of Shellforge. Each requirement includes acceptance criteria for validation.

______________________________________________________________________

## Functional Requirements

### FR-001: YAML Manifest Parsing

**Description**: Parse YAML manifest file containing shell module definitions

**Priority**: CRITICAL

**Acceptance Criteria**:

- Read YAML file from filesystem path
- Parse into Go struct: `Manifest{Modules []Module}`
- Support Module fields: name (string), file (string), requires ([]string), os ([]string), description (string)
- Handle YAML syntax errors gracefully with clear error messages
- Support optional fields (requires, os, description can be empty)
- Validate YAML structure matches expected schema

**Data Structure**:

```go
type Module struct {
    Name        string   `yaml:"name"`
    File        string   `yaml:"file"`
    Requires    []string `yaml:"requires,omitempty"`
    OS          []string `yaml:"os,omitempty"`
    Description string   `yaml:"description,omitempty"`
}

type Manifest struct {
    Modules []Module `yaml:"modules"`
}
```

**Error Cases**:

- File not found → "Manifest file not found: {path}"
- Invalid YAML syntax → "Failed to parse manifest: {error details}"
- Missing modules key → "Manifest must contain 'modules' array"

______________________________________________________________________

### FR-002: Dependency Graph Construction

**Description**: Build directed graph representing module dependencies

**Priority**: CRITICAL

**Acceptance Criteria**:

- Create graph with modules as nodes
- Add edges from dependencies to dependents (dependency → module)
- Handle modules with zero dependencies
- Handle modules with multiple dependencies
- Store full module metadata in graph nodes

**Algorithm**: Adjacency list representation

**Data Structure**:

```go
type Graph struct {
    nodes map[string]*Node
    edges map[string][]string // node -> list of dependents
}

type Node struct {
    module   Module
    incoming int // for topological sort
}
```

**Error Cases**:

- Dependency references non-existent module → validation error (FR-014)

______________________________________________________________________

### FR-003: Topological Sort with Cycle Detection

**Description**: Sort modules in dependency order, detect circular dependencies

**Priority**: CRITICAL

**Acceptance Criteria**:

- Implement topological sort (Kahn's algorithm or DFS)
- Return modules in correct load order
- Detect circular dependencies before building
- Report full circular dependency chain (e.g., "A → B → C → A")
- Handle multiple independent dependency chains
- Handle modules with no dependencies (load first)

**Algorithm**: Kahn's algorithm (BFS-based topological sort)

**Pseudocode**:

```
1. Calculate in-degree for each node
2. Add all nodes with in-degree 0 to queue
3. While queue not empty:
   a. Dequeue node, add to result
   b. For each dependent of node:
      - Decrement its in-degree
      - If in-degree becomes 0, enqueue it
4. If result.length != nodes.length:
   - Circular dependency detected
   - Find cycle and report it
```

**Error Cases**:

- Circular dependency → "Circular dependency detected: {cycle path}"
- Example: "Circular dependency: nvm → brew-path → nvm"

______________________________________________________________________

### FR-004: OS Filtering

**Description**: Filter modules based on target operating system

**Priority**: HIGH

**Acceptance Criteria**:

- Support OS values: "Mac", "Linux" (case-insensitive)
- Modules with empty `os` field apply to all platforms
- Filter modules during topological sort
- Preserve dependency order after filtering
- Auto-detect current OS if not specified

**OS Detection Logic**:

```go
func DetectOS() string {
    switch runtime.GOOS {
    case "darwin":
        return "Mac"
    case "linux":
        return "Linux"
    default:
        return runtime.GOOS
    }
}
```

**Examples**:

- Module has `os: [Mac]` → include on macOS, exclude on Linux
- Module has `os: [Mac, Linux]` → include on both
- Module has `os: []` → include on all platforms

______________________________________________________________________

### FR-005: Shell Configuration Metadata

**Description**: Built-in metadata about shell config files for different OS/shell/session combinations

**Priority**: HIGH

**Acceptance Criteria**:

- Support operating systems: macos, ubuntu, debian, arch, manjaro
- Support shells: bash, zsh, fish
- Support session types: login, interactive, always, gui
- Load metadata from embedded `data/shell_configs.yaml`
- Provide query interface to find config by OS/shell/session
- Return recommended target file path

**Data Structure**:

```go
type ConfigFile struct {
    Path         string   `yaml:"path"`
    Scope        string   `yaml:"scope"` // system | user
    Priority     int      `yaml:"priority"`
    Note         string   `yaml:"note,omitempty"`
    Alternatives []string `yaml:"alternatives,omitempty"`
    Optional     bool     `yaml:"optional"`
}

type ShellConfig struct {
    OS          []string     `yaml:"os"`
    Shell       string       `yaml:"shell"`     // bash | zsh | fish
    SessionType string       `yaml:"session_type"` // login_interactive | interactive_non_login | always
    Description string       `yaml:"description"`
    Files       []ConfigFile `yaml:"files"`
}

type ShellMetadata struct {
    Configs            []ShellConfig          `yaml:"configs"`
    RecommendedTargets map[string]interface{} `yaml:"recommended_targets"`
}
```

**Embedded Data**: Use `//go:embed data/shell_configs.yaml` to bundle metadata into binary

______________________________________________________________________

### FR-006: File I/O Operations

**Description**: Read module files and write concatenated output

**Priority**: CRITICAL

**Acceptance Criteria**:

- Read shell script files from filesystem
- Handle missing files with clear error messages
- Write concatenated output to specified path
- Create parent directories if needed
- Support dry-run mode (return content without writing)
- Preserve file permissions (0644 for output files)

**Generated Output Format**:

```bash
# Generated by shellforge
# OS: Mac
# Modules: 10
# Generated at: 2025-11-27 12:00:00

# --- module-name ---
# Module description
<module content>

# --- next-module ---
# Next module description
<next module content>
```

**Error Cases**:

- Module file not found → "Module file not found: {path}"
- Output directory doesn't exist → create it
- No write permission → "Permission denied: {path}"

______________________________________________________________________

### FR-007: Git Operations

**Description**: Git repository management for backup/versioning

**Priority**: MEDIUM

**Acceptance Criteria**:

- Initialize git repo in backup directory (~/.backup/shellforge/)
- Create commits with descriptive messages
- Support git not being installed (fail gracefully)
- Detect existing git repo (don't re-initialize)
- Configure git identity if needed

**Commands to Execute**:

```bash
git init ~/.backup/shellforge
git -C ~/.backup/shellforge add .
git -C ~/.backup/shellforge commit -m "Shellforge deployment: {timestamp}"
```

**Error Cases**:

- Git not installed → "Git is required for backup features. Please install git."
- Git command fails → "Git operation failed: {error details}"

______________________________________________________________________

### FR-008: Snapshot Management

**Description**: Create, list, restore, and cleanup timestamped snapshots

**Priority**: MEDIUM

**Acceptance Criteria**:

**Create Snapshot**:

- Copy target file to `~/.backup/shellforge/snapshots/{filename}/{timestamp}`
- Timestamp format: YYYY-MM-DD_HH-MM-SS (e.g., 2025-11-27_12-30-15)
- Return snapshot path for user confirmation

**List Snapshots**:

- Find all snapshots for target file
- Sort by timestamp (newest first)
- Display with formatted timestamp and file size

**Restore Snapshot**:

- Copy snapshot back to target location
- Create backup of current file before restoring
- Support --dry-run to preview restore

**Cleanup Snapshots**:

- Keep N most recent snapshots (--keep-count)
- Keep snapshots from last N days (--keep-days)
- Support --dry-run to preview deletions
- Never delete all snapshots (keep at least 1)

**Snapshot Directory Structure**:

```
~/.backup/shellforge/
├── .git/
├── current/
│   ├── zshrc
│   └── bashrc
└── snapshots/
    ├── zshrc/
    │   ├── 2025-11-27_12-30-15
    │   ├── 2025-11-27_11-15-00
    │   └── 2025-11-26_09-00-00
    └── bashrc/
        └── 2025-11-25_14-45-30
```

______________________________________________________________________

### FR-009: Diff Comparison

**Description**: Compare generated configuration with existing RC file

**Priority**: MEDIUM

**Acceptance Criteria**:

- Support three output formats: summary, unified, context
- Auto-detect existing RC file based on OS/shell/session
- Calculate statistics: total lines, added, removed, modified, unchanged
- Use `github.com/sergi/go-diff` for diff generation
- Provide colored output for terminal display

**Output Format (Summary)**:

```
📊 Diff Statistics:
  Total lines: 250
  Added:      15 lines
  Removed:    5 lines
  Modified:   10 lines
  Unchanged:  220 lines

💡 Next steps:
  Deploy:  shellforge build -c modules -m manifest.yaml --deploy
  Review:  shellforge diff -c modules -m manifest.yaml -e ~/.zshrc --format unified
```

**Output Format (Unified)**:

```diff
--- /home/user/.zshrc  2025-11-26 10:30:00
+++ generated          2025-11-27 12:00:00
@@ -10,7 +10,7 @@
 export PATH="/usr/local/bin:$PATH"
-export EDITOR=nano
+export EDITOR=vim

 # Aliases
+alias gs='git status'
```

______________________________________________________________________

### FR-010: Template Generation

**Description**: Generate shell modules from predefined templates

**Priority**: MEDIUM

**Acceptance Criteria**:

- Support 6 template types: path, env, alias, conditional-source, tool-init, os-specific
- Support field substitution via `-f key=value` flag
- Support interactive mode (`-i`) to prompt for all fields
- Support `-r` flag to specify module dependencies
- Auto-categorize templates into init.d/, rc_pre.d/, or rc_post.d/
- Create module file with proper header comment

**Template Types**:

1. **path**: Add directory to PATH

   - Fields: path_dir, description
   - Category: init.d/
   - Example: `shellforge template generate path my-bin -f path_dir=/usr/local/mybin`

1. **env**: Set environment variable

   - Fields: var_name, var_value, description
   - Category: rc_pre.d/
   - Example: `shellforge template generate env EDITOR -f var_name=EDITOR -f var_value=vim`

1. **alias**: Define shell aliases

   - Fields: description, aliases (multiline)
   - Category: rc_post.d/
   - Example: `shellforge template generate alias git-shortcuts -f "aliases=alias gs='git status'"`

1. **conditional-source**: Source file if it exists

   - Fields: source_path, description
   - Category: rc_pre.d/
   - Example: `shellforge template generate conditional-source local-config -f source_path=~/.zshrc.local`

1. **tool-init**: Initialize development tool

   - Fields: tool_name, tool_command, init_command, description
   - Category: rc_pre.d/
   - Example: `shellforge template generate tool-init nvm -f tool_name=nvm -f init_command='eval "$(nvm init -)"'`

1. **os-specific**: OS-specific configuration

   - Fields: description, mac_content, linux_content
   - Category: rc_pre.d/
   - Example: `shellforge template generate os-specific package-manager`

**Template List Output**:

```
Available templates:

  path                  Add directory to PATH (init.d/)
  env                   Set environment variable (rc_pre.d/)
  alias                 Define shell aliases (rc_post.d/)
  conditional-source    Source file if it exists (rc_pre.d/)
  tool-init             Initialize development tool (rc_pre.d/)
  os-specific           OS-specific configuration (rc_pre.d/)

Usage:
  shellforge template generate <template-name> <module-name> [flags]
  shellforge template generate path my-bin -f path_dir=/usr/local/bin
  shellforge template generate env EDITOR -i  # interactive mode
```

______________________________________________________________________

### FR-011: Migration from Monolithic RC

**Description**: Convert traditional .zshrc/.bashrc to modular structure

**Priority**: MEDIUM

**Acceptance Criteria**:

- Parse RC file line by line
- Detect section boundaries (comments with ---, ===, ALL CAPS headers)
- Extract section names from comments
- Categorize sections: init.d/ (PATH, early setup), rc_pre.d/ (tool init), rc_post.d/ (aliases, functions)
- Create module files with extracted content
- Auto-generate manifest.yaml with inferred dependencies
- Create backup of original file (--no-backup to skip)
- Support --dry-run to preview migration

**Section Detection Patterns**:

```bash
# Matches section headers:
# --- Section Name ---
# === Section Name ===
# SECTION NAME (all caps line)
# ## Section Name
```

**Auto-Categorization Rules**:

- Lines before first section → init.d/00-preamble.sh
- Sections with PATH manipulation → init.d/
- Sections with tool initialization (nvm, rbenv, etc.) → rc_pre.d/
- Sections with aliases, functions → rc_post.d/
- Unknown sections → rc_pre.d/ (safe default)

**Dependency Inference**:

- Contains `$MACHINE` → requires os-detection
- Contains `brew` → requires brew-path
- Contains tool-specific patterns → infer tool dependencies

______________________________________________________________________

### FR-012: Auto-Init (Manifest Generation)

**Description**: Generate manifest.yaml from existing modular structure

**Priority**: MEDIUM

**Acceptance Criteria**:

- Scan init.d/, rc_pre.d/, rc_post.d/ directories
- Create module entry for each .sh file
- Infer dependencies from file content
- Detect OS support from `case $MACHINE` statements
- Extract descriptions from file header comments
- Generate manifest with category comments
- Support --dry-run to preview generated manifest

**Dependency Inference Logic**:

```
If file contains "$MACHINE":
  → add dependency: os-detection

If file contains "brew" command:
  → add dependency: brew-path

If file contains specific tool patterns:
  nvm → brew-path (on Mac)
  conda → brew-path (on Mac)
  asdf-<tool> → asdf-core
```

**OS Detection Logic**:

```bash
# If file contains this pattern:
case $MACHINE in
  Mac)
    # Mac-specific code
    ;;
  Linux)
    # Linux-specific code
    ;;
esac

# Then infer: os: [Mac, Linux]
```

**Generated Manifest Example**:

```yaml
# Shellforge Manifest
# Auto-generated from existing shell configuration
# Generated at: 2025-11-27 12:00:00

modules:
  # System Initialization (init.d/)
  - name: os-detection
    file: init.d/00-os-detection.sh
    requires: []
    os: [Mac, Linux]
    description: Detect operating system and set MACHINE variable

  # Pre-RC Configuration (rc_pre.d/)
  - name: conda
    file: rc_pre.d/conda.sh
    requires: [os-detection, brew-path]
    os: [Mac, Linux]
    description: Conda environment initialization

  # Post-RC Configuration (rc_post.d/)
  - name: aliases
    file: rc_post.d/aliases.sh
    requires: []
    os: [Mac, Linux]
    description: Common aliases and functions
```

______________________________________________________________________

### FR-013: Verbose Mode

**Description**: Detailed output for debugging and transparency

**Priority**: LOW

**Acceptance Criteria**:

- Support `-v` or `--verbose` flag on all commands
- Show dependency resolution steps
- Show module load order with reasons
- Show file operations (read, write, copy)
- Show git operations (init, add, commit)
- Use structured logging (not just debug prints)

**Verbose Output Example**:

```
Building shell configuration...

Dependency Resolution:
  ✓ Loaded 10 modules from manifest
  ✓ Built dependency graph (10 nodes, 15 edges)
  ✓ Topological sort completed

Module Load Order:
  1. os-detection (init.d/00-os-detection.sh) [no dependencies]
  2. brew-path (init.d/05-brew-path.sh) [depends on: os-detection]
  3. conda (rc_pre.d/conda.sh) [depends on: os-detection, brew-path]
  ...

File Operations:
  → Reading: init.d/00-os-detection.sh (120 bytes)
  → Reading: init.d/05-brew-path.sh (85 bytes)
  → Writing: /tmp/output.zsh (2,450 bytes)

✓ Build completed successfully
```

______________________________________________________________________

### FR-014: Validation & Error Handling

**Description**: Comprehensive validation with user-friendly error messages

**Priority**: CRITICAL

**Acceptance Criteria**:

**Manifest Validation**:

- ✓ YAML syntax is valid
- ✓ "modules" key exists and is array
- ✓ Each module has "name" field (non-empty string)
- ✓ Each module has "file" field (non-empty string)
- ✓ Module names are unique
- ✓ All module files exist on filesystem
- ✓ All "requires" entries reference existing modules
- ✓ No circular dependencies
- ✓ OS values are valid ("Mac" or "Linux")

**Error Message Format**:

```
✗ Validation failed:
  • Module 'nvm' missing 'file' field
  • Module file not found: init.d/missing.sh
  • Circular dependency detected: asdf-core → nvm → asdf-core

💡 How to fix:
  1. Add 'file' field to module 'nvm' in manifest.yaml
  2. Create missing file or update path in manifest
  3. Edit manifest.yaml and remove one dependency from the circular chain

Example:
  modules:
    - name: nvm
      file: rc_pre.d/nvm.sh  # <- Add this line
      requires: [brew-path]
```

**Exit Codes**:

- 0: Success
- 1: User error (invalid manifest, missing files, etc.)
- 2: System error (git not found, filesystem error, etc.)

______________________________________________________________________

## CLI Command Requirements

### CMD-001: build

**Usage**: `shellforge build [flags]`

**Description**: Generate shell configuration file from modules

**Flags**:

```
-c, --config-dir PATH      Directory containing shell modules (required)
-m, --manifest PATH        Path to manifest.yaml (required)
-o, --output PATH          Output file path (.zshrc, .bashrc, etc.)
    --auto-output          Auto-detect output path based on OS/shell/session
    --os STRING            Target OS (macos, ubuntu, debian, arch, manjaro)
    --shell STRING         Target shell (bash, zsh, fish)
    --session STRING       Session type (login, interactive, always)
    --deploy               Deploy with backup after building
    --dry-run              Preview output without writing file
-v, --verbose              Show detailed processing steps
```

**Examples**:

```bash
# Basic build with manual output
shellforge build -c config/shellrc -m manifest.yaml -o ~/.zshrc

# Auto-detect output path
shellforge build -c config/shellrc -m manifest.yaml --auto-output

# Build for different OS/shell
shellforge build -c config -m manifest.yaml --os linux --shell bash --auto-output

# Preview without writing
shellforge build -c config -m manifest.yaml --auto-output --dry-run

# Build and deploy
shellforge build -c config -m manifest.yaml --auto-output --deploy
```

______________________________________________________________________

### CMD-002: validate

**Usage**: `shellforge validate [flags]`

**Description**: Check manifest and module files for errors

**Flags**:

```
-c, --config-dir PATH      Directory containing shell modules (required)
-m, --manifest PATH        Path to manifest.yaml (required)
-v, --verbose              Show detailed validation steps
```

**Examples**:

```bash
# Validate configuration
shellforge validate -c config/shellrc -m manifest.yaml

# Verbose validation
shellforge validate -c config/shellrc -m manifest.yaml -v
```

**Output Example**:

```
✓ Validation passed

📊 Validation Statistics:
  • Total modules: 27
  • Mac modules: 27
  • Linux modules: 22
  • Total dependencies: 18

💡 Next steps:
  Preview:  shellforge build -c config -m manifest.yaml --dry-run
  Compare:  shellforge diff -c config -m manifest.yaml --auto-detect-existing
```

______________________________________________________________________

### CMD-003: init

**Usage**: `shellforge init [flags]`

**Description**: Auto-generate manifest from existing config structure

**Flags**:

```
-c, --config-dir PATH      Directory containing init.d/rc_pre.d/rc_post.d (required)
-o, --output PATH          Output manifest file path (default: manifest.yaml)
    --dry-run              Preview manifest without writing file
```

**Examples**:

```bash
# Generate manifest
shellforge init -c config/shellrc -o manifest.yaml

# Preview without writing
shellforge init -c config/shellrc --dry-run
```

______________________________________________________________________

### CMD-004: migrate

**Usage**: `shellforge migrate [flags]`

**Description**: Convert monolithic RC file to modular structure

**Flags**:

```
-s, --source PATH          Source RC file (.zshrc, .bashrc, etc.) (required)
-t, --target DIR           Target directory for modular structure (required)
    --dry-run              Preview migration without writing files
    --no-backup            Skip creating backup of source file
    --no-manifest          Skip auto-generating manifest.yaml
```

**Examples**:

```bash
# Migrate .zshrc
shellforge migrate -s ~/.zshrc -t config/shellrc

# Preview migration
shellforge migrate -s ~/.zshrc -t config/shellrc --dry-run

# Migrate without backup (dangerous!)
shellforge migrate -s ~/.zshrc -t config/shellrc --no-backup
```

______________________________________________________________________

### CMD-005: diff

**Usage**: `shellforge diff [flags]`

**Description**: Compare generated config with existing RC file

**Flags**:

```
-c, --config-dir PATH      Directory containing shell modules (required)
-m, --manifest PATH        Path to manifest.yaml (required)
-e, --existing PATH        Existing RC file to compare against
    --auto-detect-existing Auto-detect existing RC file
    --format STRING        Output format (summary, unified, context) (default: summary)
-v, --verbose              Show detailed diff analysis
```

**Examples**:

```bash
# Compare with auto-detected RC file
shellforge diff -c config -m manifest.yaml --auto-detect-existing

# Compare with specific file
shellforge diff -c config -m manifest.yaml -e ~/.zshrc

# Show unified diff
shellforge diff -c config -m manifest.yaml --auto-detect-existing --format unified
```

______________________________________________________________________

### CMD-006: restore

**Usage**: `shellforge restore [flags]`

**Description**: Restore previous configuration from backup

**Flags**:

```
-t, --target PATH          Target file to restore (required, e.g., ~/.zshrc)
    --list                 List available snapshots
-s, --snapshot PATH        Specific snapshot to restore
    --dry-run              Preview restore without writing file
```

**Examples**:

```bash
# List available snapshots
shellforge restore -t ~/.zshrc --list

# Interactive restore (shows recent snapshots)
shellforge restore -t ~/.zshrc

# Restore specific snapshot
shellforge restore -t ~/.zshrc -s ~/.backup/shellforge/snapshots/zshrc/2025-11-27_10-30-15
```

______________________________________________________________________

### CMD-007: clean-snapshots

**Usage**: `shellforge clean-snapshots [flags]`

**Description**: Manage snapshot retention

**Flags**:

```
-t, --target PATH          Target file (e.g., ~/.zshrc)
    --keep-count INT       Keep N most recent snapshots
    --keep-days INT        Keep snapshots from last N days
    --dry-run              Preview deletions without removing files
```

**Examples**:

```bash
# Keep only last 10 snapshots
shellforge clean-snapshots -t ~/.zshrc --keep-count 10

# Keep last 30 days
shellforge clean-snapshots -t ~/.zshrc --keep-days 30

# Preview cleanup
shellforge clean-snapshots -t ~/.zshrc --keep-count 10 --dry-run
```

______________________________________________________________________

### CMD-008: list-modules

**Usage**: `shellforge list-modules [flags]`

**Description**: Show modules in dependency load order

**Flags**:

```
-c, --config-dir PATH      Directory containing shell modules (required)
-m, --manifest PATH        Path to manifest.yaml (required)
    --os STRING            Filter by OS (macos, linux, etc.)
```

**Examples**:

```bash
# List all modules
shellforge list-modules -c config -m manifest.yaml

# List Mac-only modules
shellforge list-modules -c config -m manifest.yaml --os mac
```

**Output Example**:

```
Module Load Order:

  1. os-detection           init.d/00-os-detection.sh      [Mac, Linux]
  2. brew-path              init.d/05-brew-path.sh         [Mac]
  3. conda                  rc_pre.d/conda.sh              [Mac, Linux]
  4. nvm                    rc_pre.d/nvm.sh                [Mac, Linux]
  5. aliases                rc_post.d/aliases.sh           [Mac, Linux]

Total: 5 modules
```

______________________________________________________________________

### CMD-009: info

**Usage**: `shellforge info [flags]`

**Description**: Show shell configuration metadata

**Flags**:

```
    --os STRING            Operating system (macos, ubuntu, debian, arch, manjaro)
    --shell STRING         Shell type (bash, zsh, fish)
    --session STRING       Session type (login, interactive, always)
```

**Examples**:

```bash
# macOS zsh interactive
shellforge info --os macos --shell zsh --session interactive

# Ubuntu bash login
shellforge info --os ubuntu --shell bash --session login
```

**Output Example**:

```
Shell Configuration: macOS / zsh / interactive

Files loaded (in order):
  1. /etc/zshenv               [system, always]
  2. ~/.zshenv                 [user, always]
  3. ~/.zshrc                  [user, interactive] ← recommended build target

Recommended target for building:
  ~/.zshrc

Session type: Interactive non-login shell
  - Launched by: most terminal emulators (iTerm2, Terminal.app, etc.)
  - Used for: daily interactive work
```

______________________________________________________________________

### CMD-010: template list

**Usage**: `shellforge template list`

**Description**: Show available module templates

**Output**: List of templates with descriptions and categories

______________________________________________________________________

### CMD-011: template generate

**Usage**: `shellforge template generate <template-name> <module-name> [flags]`

**Description**: Create module from template

**Flags**:

```
-f, --field key=value      Set template field value (repeatable)
-i, --interactive          Prompt for all fields interactively
-c, --config-dir PATH      Config directory (default: current directory)
-r, --require MODULE       Add module dependency (repeatable)
```

**Examples**:

```bash
# Generate PATH module
shellforge template generate path my-bin -f path_dir=/usr/local/mybin

# Generate environment variable
shellforge template generate env EDITOR -f var_name=EDITOR -f var_value=vim

# Interactive mode
shellforge template generate tool-init nvm -i

# With dependencies
shellforge template generate tool-init nvm -r brew-path -r os-detection
```

______________________________________________________________________

## Performance Requirements

### PR-001: Startup Time

**Target**: \<10ms startup (PRODUCT.md G2)

**Rationale**: Near-instant CLI responsiveness (Python version: ~200ms)

**Measurement**: Warm median of the `version` workload on the release binary (PRODUCT.md G2 Metric)

**Validation**: `python3 scripts/startup-profile.py --binary <release binary>` on macOS

______________________________________________________________________

### PR-002: Build Time

**Target**: \<500ms to build 50 modules

**Rationale**: Fast feedback loop for development

**Measurement**: Time from `build` command start to file written

**Validation**: Build 50 modules, measure with `time shellforge build ...`

______________________________________________________________________

### PR-003: Memory Usage

**Target**: \<10MB peak memory footprint (PRODUCT.md G2)

**Rationale**: Lightweight, suitable for resource-constrained environments

**Measurement**: Largest `peak_footprint_bytes` of the `version` workload (maximum RSS is reference only)

**Validation**: The same `scripts/startup-profile.py` run as PR-001

______________________________________________________________________

### PR-004: Binary Size

**Target**: \<10MB compiled binary

**Rationale**: Fast download, easy distribution

**Measurement**: Size of `shellforge` binary after `go build`

**Validation**: `ls -lh shellforge` should show \<10MB

______________________________________________________________________

## Quality Requirements

### QR-001: Test Coverage

**Target**: >80% code coverage

**Measurement**: `go test -cover ./...`

**Critical Paths** (must have 100% coverage):

- Dependency graph construction
- Topological sort
- Manifest parsing
- Validation logic

______________________________________________________________________

### QR-002: Zero C Dependencies

**Target**: Pure Go binary with no CGO

**Rationale**: Easy cross-compilation, no libc version issues

**Validation**: `go build` with `CGO_ENABLED=0`

______________________________________________________________________

### QR-003: Cross-Compilation

**Target**: Build for Linux, macOS, BSD without modification

**Platforms**:

- linux/amd64, linux/arm64
- darwin/amd64, darwin/arm64
- freebsd/amd64

**Validation**: `GOOS=linux GOARCH=amd64 go build` succeeds

______________________________________________________________________

### QR-004: Error Message Quality

**Target**: All errors include:

1. What went wrong
1. Why it happened
1. How to fix it
1. Example of correct usage (where applicable)

**Example**:

```
✗ Module file not found: init.d/missing.sh

Cause: The file 'init.d/missing.sh' referenced in manifest.yaml does not exist.

How to fix:
  1. Check the file path in manifest.yaml
  2. Verify the file exists: ls init.d/missing.sh
  3. Create the file or update the path

Example manifest entry:
  - name: my-module
    file: init.d/existing-file.sh  # ← Use correct path
```

______________________________________________________________________

## Compatibility Requirements

### COMPAT-001: Manifest Format

**Requirement**: 100% backward compatible with Python version

**Validation**: Python-generated manifest.yaml works without modification

______________________________________________________________________

### COMPAT-002: CLI Interface

**Requirement**: Same command names, flag names, and flag behavior

**Exceptions**:

- Flag order may differ in help text
- Output formatting may use Go-style formatting
- Error message wording may improve but semantics must match

______________________________________________________________________

### COMPAT-003: Generated Output

**Requirement**: Generated shell config is functionally equivalent

**Validation**: Diffing Python output vs Go output shows only metadata differences (timestamps, generator version)

______________________________________________________________________

## Non-Functional Requirements

### NFR-001: Code Quality

**Requirements**:

- Pass `gofmt` (standard formatting)
- Pass `golint` (no lint warnings)
- Pass `go vet` (static analysis)
- Follow Go standard project layout

______________________________________________________________________

### NFR-002: Documentation

**Requirements**:

- godoc for all exported functions/types
- README with quick start guide
- Usage examples for all commands
- Architecture documentation

______________________________________________________________________

### NFR-003: Dependency Management

**Requirements**:

- Use `go.mod` for dependency tracking
- Pin dependency versions
- Minimal external dependencies (\<10 packages)
- Prefer standard library

______________________________________________________________________

## Appendix: Test Scenarios

### Scenario 1: Simple Linear Dependencies

```yaml
modules:
  - name: a
    file: a.sh
    requires: []
  - name: b
    file: b.sh
    requires: [a]
  - name: c
    file: c.sh
    requires: [b]
```

**Expected Load Order**: a → b → c

______________________________________________________________________

### Scenario 2: Complex DAG

```yaml
modules:
  - name: base
    file: base.sh
    requires: []
  - name: tool1
    file: tool1.sh
    requires: [base]
  - name: tool2
    file: tool2.sh
    requires: [base]
  - name: config
    file: config.sh
    requires: [tool1, tool2]
```

**Expected Load Order**: base → (tool1, tool2 in any order) → config

______________________________________________________________________

### Scenario 3: Circular Dependency

```yaml
modules:
  - name: a
    file: a.sh
    requires: [b]
  - name: b
    file: b.sh
    requires: [c]
  - name: c
    file: c.sh
    requires: [a]
```

**Expected Behavior**: Validation error with cycle path: a → b → c → a

______________________________________________________________________

### Scenario 4: OS Filtering

```yaml
modules:
  - name: base
    file: base.sh
    os: [Mac, Linux]
  - name: brew
    file: brew.sh
    os: [Mac]
    requires: [base]
  - name: pacman
    file: pacman.sh
    os: [Linux]
    requires: [base]
```

**On macOS**: base → brew
**On Linux**: base → pacman

______________________________________________________________________

**Document Status**: Ready for review
**Next Steps**: Write ARCHITECTURE.md with Go-specific design
