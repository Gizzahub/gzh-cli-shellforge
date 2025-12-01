# Technology Stack: Shellforge Go Implementation

**Version**: 1.0
**Status**: Draft
**Last Updated**: 2025-11-27

---

## Overview

This document details all library choices for the Shellforge Go implementation, with rationale and alternatives considered.

---

## Core Principles

1. **Prefer Standard Library**: Use Go stdlib wherever possible
2. **Well-Established Libraries**: Choose mature, widely-used packages
3. **Minimal Dependencies**: Keep dependency count low (<10 external packages)
4. **Zero C Dependencies**: Pure Go for easy cross-compilation
5. **Active Maintenance**: Libraries must be actively maintained

---

## Library Decisions

### 1. CLI Framework

#### Choice: Cobra

**Package**: `github.com/spf13/cobra`

**Version**: Latest stable (v1.8+)

**Rationale**:
- **Industry Standard**: Used by kubectl, hugo, gh (GitHub CLI), docker CLI
- **Excellent Subcommand Support**: Built for complex CLIs with nested commands
- **Auto-Generated Help**: Rich help text with command groups and examples
- **Completion Scripts**: Built-in bash/zsh/fish completion generation
- **Persistent Flags**: Global flags that work across all subcommands
- **Pre/Post Run Hooks**: Execute code before/after commands
- **Large Community**: Extensive documentation and examples

**Key Features Used**:
```go
// Command groups
cmd.AddCommand(buildCmd, validateCmd, initCmd)

// Persistent flags (global)
rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

// Required flags
cmd.MarkFlagRequired("config-dir")

// Custom usage template
cmd.SetUsageTemplate(customTemplate)
```

**Alternatives Considered**:

| Library | Pros | Cons | Decision |
|---------|------|------|----------|
| `flag` (stdlib) | No dependencies | Too basic for complex CLI | ❌ Rejected |
| `urfave/cli` | Lighter than Cobra | Less popular, different API style | ❌ Rejected |
| `spf13/pflag` | POSIX-style flags | No subcommand support | ❌ Rejected |

---

### 2. YAML Parsing

#### Choice: gopkg.in/yaml.v3

**Package**: `gopkg.in/yaml.v3`

**Version**: v3.0+

**Rationale**:
- **Most Popular**: De facto standard Go YAML library
- **Full YAML 1.2 Support**: Handles all YAML features (anchors, aliases, tags)
- **Good API**: Clean Marshal/Unmarshal interface
- **Struct Tags**: Nice integration with Go structs
- **Well-Tested**: Used by Kubernetes, Docker, and many major projects
- **Active Development**: Regular updates and bug fixes

**Usage Example**:
```go
import "gopkg.in/yaml.v3"

type Manifest struct {
    Modules []Module `yaml:"modules"`
}

func parseManifest(path string) (*Manifest, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var manifest Manifest
    if err := yaml.Unmarshal(data, &manifest); err != nil {
        return nil, err
    }

    return &manifest, nil
}
```

**Alternatives Considered**:

| Library | Pros | Cons | Decision |
|---------|------|------|----------|
| `goccy/go-yaml` | Faster, colored output | Newer, less proven | ❌ Rejected |
| `go-yaml/yaml` (v2) | Older version | v3 is better | ❌ Rejected |

**Why Not JSON?**: YAML is more human-friendly for config files (comments, multiline strings, no quotes needed)

---

### 3. Graph Algorithm (Topological Sort)

#### Choice: Custom Implementation

**Package**: None (custom code in `internal/domain/resolver.go`)

**Rationale**:
- **Simple Algorithm**: Kahn's algorithm is ~50 lines of Go
- **No Heavy Dependency**: Avoid pulling in entire graph library for one algorithm
- **Full Control**: Easier to add custom error messages and cycle detection
- **Better Performance**: No overhead from generic graph library
- **Clearer Code**: Algorithm is visible and reviewable

**Algorithm**: Kahn's Algorithm (BFS-based topological sort)

**Pseudocode**:
```
1. Calculate in-degree for each node
2. Add all nodes with in-degree 0 to queue
3. While queue not empty:
   a. Dequeue node, add to result
   b. For each dependent:
      - Decrement in-degree
      - If in-degree becomes 0, enqueue
4. If result.length != nodes.length:
   - Circular dependency detected
```

**Complexity**: O(V + E) where V = nodes, E = edges

**Alternatives Considered**:

| Library | Pros | Cons | Decision |
|---------|------|------|----------|
| `gonum/graph` | Full graph library | Overkill (100+ KB), complex API | ❌ Rejected |
| `yourbasic/graph` | Simple API | Unmaintained | ❌ Rejected |
| Custom DFS | Also works | BFS (Kahn's) is clearer | ✅ Chose Kahn's |

---

### 4. File System Abstraction

#### Choice: Afero

**Package**: `github.com/spf13/afero`

**Version**: Latest stable (v1.11+)

**Rationale**:
- **Essential for Testing**: In-memory filesystem for fast, isolated tests
- **Standard Interface**: `fs.FS` compatible
- **Widely Used**: Same author as Cobra, used by Hugo and many others
- **Multiple Backends**: OsFs (real), MemMapFs (in-memory), ReadOnlyFs, etc.
- **Easy Mocking**: No need to mock individual file operations

**Usage Example**:
```go
import "github.com/spf13/afero"

// Production: real filesystem
fs := afero.NewOsFs()

// Testing: in-memory filesystem
fs := afero.NewMemMapFs()

// Both use same interface
type FileReader interface {
    ReadFile(path string) ([]byte, error)
    FileExists(path string) bool
}
```

**Test Example**:
```go
func TestBuild(t *testing.T) {
    // Setup in-memory filesystem
    fs := afero.NewMemMapFs()
    afero.WriteFile(fs, "manifest.yaml", []byte("modules: []"), 0644)
    afero.WriteFile(fs, "init.d/test.sh", []byte("echo test"), 0644)

    // Test without touching real filesystem
    reader := filesystem.NewReader(fs)
    // ...
}
```

**Alternatives Considered**:

| Library | Pros | Cons | Decision |
|---------|------|------|----------|
| `os` package directly | No dependency | Hard to test, requires temp files | ❌ Rejected |
| `io/fs` (stdlib) | Standard interface | No in-memory implementation | ❌ Rejected |
| `go-billy` | VFS interface | Less popular, complex | ❌ Rejected |

---

### 5. Git Operations

#### Choice: os/exec with git commands

**Package**: `os/exec` (stdlib)

**Rationale**:
- **Simple Commands**: Only need `git init`, `git add`, `git commit`
- **Reliable**: Directly use user's git installation
- **No Large Dependency**: Pure-Go git libraries are huge (go-git: ~5MB)
- **User Familiarity**: Uses same git version as user's shell
- **Easy Debugging**: Can see exact git commands in verbose mode

**Implementation**:
```go
import "os/exec"

func gitInit(dir string) error {
    cmd := exec.Command("git", "init", dir)
    return cmd.Run()
}

func gitCommit(dir, message string) error {
    // Add all changes
    cmd := exec.Command("git", "-C", dir, "add", ".")
    if err := cmd.Run(); err != nil {
        return err
    }

    // Commit
    cmd = exec.Command("git", "-C", dir, "commit", "-m", message)
    return cmd.Run()
}
```

**Dependency Check**:
```go
func checkGitAvailable() error {
    if _, err := exec.LookPath("git"); err != nil {
        return fmt.Errorf("git is required for backup features. Please install git.")
    }
    return nil
}
```

**Alternatives Considered**:

| Library | Pros | Cons | Decision |
|---------|------|------|----------|
| `go-git/go-git` | Pure Go, no git needed | Large dependency (~5MB), complex API | ❌ Rejected |
| `libgit2/git2go` | Native performance | C dependency (CGO), hard to cross-compile | ❌ Rejected |
| `os/exec` | Simple, reliable | Requires git installed | ✅ Chosen |

---

### 6. Diff Algorithm

#### Choice: sergi/go-diff

**Package**: `github.com/sergi/go-diff`

**Version**: Latest stable (v1.3+)

**Rationale**:
- **Unified Diff Format**: Implements standard unified diff (like `diff -u`)
- **Context Diff Format**: Also supports context diff
- **Well-Tested**: Used by many Go projects
- **Lightweight**: Small, focused library
- **Easy API**: Simple to use

**Usage Example**:
```go
import (
    "github.com/sergi/go-diff/diffmatchpatch"
)

func generateDiff(oldText, newText string) string {
    dmp := diffmatchpatch.New()
    diffs := dmp.DiffMain(oldText, newText, false)

    // Unified diff format
    patches := dmp.PatchMake(oldText, diffs)
    return dmp.PatchToText(patches)
}
```

**Alternatives Considered**:

| Library | Pros | Cons | Decision |
|---------|------|------|----------|
| Custom implementation | No dependency | Reinventing wheel, complex algorithm | ❌ Rejected |
| `pmezard/go-difflib` | Standard-ish | Less maintained | ❌ Rejected |
| `sergi/go-diff` | Widely used, good API | None | ✅ Chosen |

---

### 7. Testing Framework

#### Choice: stdlib testing + testify/assert

**Package**:
- `testing` (stdlib)
- `github.com/stretchr/testify` (assertions only)

**Rationale**:

**stdlib `testing`**:
- Built-in, no dependency
- Table-driven tests with `t.Run()`
- Benchmark support
- Standard Go way

**testify/assert**:
- Better assertion readability
- Clear failure messages
- Widely used in Go community
- Lightweight (just assertions, not full framework)

**Usage Example**:
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestModule_AppliesTo(t *testing.T) {
    tests := []struct {
        name      string
        module    Module
        targetOS  string
        expected  bool
    }{
        {
            name:     "matches single OS",
            module:   Module{OS: []string{"Mac"}},
            targetOS: "Mac",
            expected: true,
        },
        {
            name:     "no match",
            module:   Module{OS: []string{"Mac"}},
            targetOS: "Linux",
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.module.AppliesTo(tt.targetOS)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

**Alternatives Considered**:

| Library | Pros | Cons | Decision |
|---------|------|------|----------|
| `ginkgo/gomega` | BDD-style | Overkill, non-standard | ❌ Rejected |
| Pure stdlib | No dependencies | Verbose assertions | ⚠️ Acceptable |
| `testify` | Better DX | Small dependency | ✅ Chosen |

---

### 8. Embedded Data Files

#### Choice: embed package (Go 1.16+)

**Package**: `embed` (stdlib, Go 1.16+)

**Rationale**:
- **Built-in**: No external dependency
- **Single Binary**: Bundle `data/shell_configs.yaml` into binary
- **Simple API**: `//go:embed` directive
- **Compile-Time**: No runtime file I/O for bundled data

**Usage Example**:
```go
import (
    _ "embed"
)

//go:embed data/shell_configs.yaml
var shellConfigsData []byte

func loadShellMetadata() (*ShellMetadata, error) {
    var metadata ShellMetadata
    err := yaml.Unmarshal(shellConfigsData, &metadata)
    return &metadata, err
}
```

**Alternatives Considered**:

| Approach | Pros | Cons | Decision |
|----------|------|------|----------|
| Separate data file | Easy to edit | Requires runtime file | ❌ Rejected |
| `go-bindata` | Works on Go <1.16 | External tool, deprecated | ❌ Rejected |
| `embed` package | Built-in, clean | Requires Go 1.16+ | ✅ Chosen |

---

### 9. Logging (Future)

#### Choice: log/slog (Go 1.21+)

**Package**: `log/slog` (stdlib, Go 1.21+)

**Rationale**:
- **Structured Logging**: Key-value pairs for machine-parseable logs
- **Built-in**: No dependency
- **Levels**: Debug, Info, Warn, Error
- **Handlers**: Text, JSON formats

**Not Needed for v1.0**: Use `fmt.Println` and `fmt.Fprintf(os.Stderr, ...)` for now

**Future Usage**:
```go
import "log/slog"

slog.Info("building config",
    "modules", len(modules),
    "os", targetOS,
)

slog.Debug("reading module file",
    "path", filePath,
    "size", fileSize,
)
```

**Alternatives**:

| Library | Pros | Cons | Decision |
|---------|------|------|----------|
| `logrus` | Popular | Large dependency | ❌ Not needed |
| `zap` | Fast | Complex API | ❌ Not needed |
| `log/slog` | Stdlib, simple | Requires Go 1.21+ | ✅ Future use |

---

### 10. Error Handling

#### Choice: errors package (Go 1.13+)

**Package**: `errors` (stdlib)

**Rationale**:
- **Error Wrapping**: `fmt.Errorf("context: %w", err)`
- **Error Unwrapping**: `errors.Unwrap(err)`
- **Error Checking**: `errors.Is(err, target)`
- **Error Assertion**: `errors.As(err, &target)`

**Usage Example**:
```go
import "errors"

// Wrap errors with context
if err := parser.Parse(path); err != nil {
    return nil, fmt.Errorf("failed to parse manifest: %w", err)
}

// Check specific error types
if errors.Is(err, os.ErrNotExist) {
    return fmt.Errorf("manifest file not found")
}

// Custom error types
var validationErr *ValidationError
if errors.As(err, &validationErr) {
    fmt.Println("Validation failed:", validationErr.Message)
}
```

**Alternatives**:

| Library | Pros | Cons | Decision |
|---------|------|------|----------|
| `pkg/errors` | Stack traces | Deprecated, use stdlib | ❌ Rejected |
| `errors` (stdlib) | Built-in, standard | No stack traces | ✅ Chosen |

---

## Complete Dependency List

### External Dependencies (6 total)

1. `github.com/spf13/cobra` - CLI framework
2. `gopkg.in/yaml.v3` - YAML parsing
3. `github.com/spf13/afero` - Filesystem abstraction
4. `github.com/sergi/go-diff` - Diff generation
5. `github.com/stretchr/testify` - Test assertions (dev only)
6. (Future) `golang.org/x/term` - Terminal size detection (for formatting)

### Standard Library Only

- `os/exec` - Git operations
- `embed` - Embed data files
- `errors` - Error handling
- `fmt` - Formatting
- `os` - OS operations
- `path/filepath` - Path manipulation
- `strings` - String operations
- `time` - Timestamps
- `testing` - Test framework

---

## Go Version Requirement

### Minimum: Go 1.21

**Required Features**:
- `embed` package (Go 1.16+)
- `errors.Is`, `errors.As` (Go 1.13+)
- `go.mod` (Go 1.11+)
- (Optional) `log/slog` (Go 1.21+)

**Rationale**:
- Go 1.21 is stable and widely available
- Provides all needed stdlib features
- Good balance of modern features vs compatibility

---

## Build Configuration

### go.mod

```go
module github.com/gizzahub/gzh-cli-shellforge

go 1.21

require (
    github.com/sergi/go-diff v1.3.1
    github.com/spf13/afero v1.11.0
    github.com/spf13/cobra v1.8.0
    gopkg.in/yaml.v3 v3.0.1
)

require (
    github.com/inconshreveable/mousetrap v1.1.0 // indirect
    github.com/spf13/pflag v1.0.5 // indirect
    golang.org/x/text v0.14.0 // indirect
)

// Test dependencies
require (
    github.com/stretchr/testify v1.8.4
)
```

### Build Flags

```bash
# Production build
go build -ldflags="-s -w" -o shellforge cmd/shellforge/main.go

# Flags explained:
# -s: Strip symbol table (reduce binary size)
# -w: Strip DWARF debug info (reduce binary size)

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o shellforge-linux

# Static binary (no external dependencies)
CGO_ENABLED=0 go build -o shellforge
```

---

## Dependency Update Strategy

### Update Frequency

- **Minor updates**: Every 3 months
- **Security patches**: Immediately
- **Major versions**: Evaluate carefully

### Update Process

```bash
# Check for updates
go list -u -m all

# Update all dependencies
go get -u ./...

# Update specific package
go get -u github.com/spf13/cobra@latest

# Tidy up
go mod tidy

# Verify
go mod verify
```

### Security Scanning

```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan for vulnerabilities
govulncheck ./...
```

---

## Performance Considerations

### Dependency Impact on Binary Size

| Dependency | Estimated Size Impact |
|------------|----------------------|
| Cobra | ~1.5 MB |
| yaml.v3 | ~500 KB |
| afero | ~300 KB |
| go-diff | ~200 KB |
| Custom code | ~1 MB |
| **Total** | **~3.5 MB** |

**Target**: <10 MB (within target)

### Dependency Impact on Build Time

| Dependency | Build Time Impact |
|------------|-------------------|
| Cobra | ~2 seconds |
| yaml.v3 | ~1 second |
| Others | <1 second |
| **Total** | **~3-4 seconds** (cold build) |

**Incremental builds**: <1 second

---

## License Compatibility

All dependencies use permissive licenses compatible with MIT:

| Dependency | License |
|------------|---------|
| Cobra | Apache 2.0 |
| yaml.v3 | Apache 2.0 & MIT |
| afero | Apache 2.0 |
| go-diff | MIT |
| testify | MIT |

**Project License**: MIT (compatible with all dependencies)

---

## Alternatives NOT Chosen

### Why Not NetworkX Equivalent?

**gonum/graph**:
- ❌ 100+ KB dependency for single algorithm
- ❌ Complex API, steep learning curve
- ❌ Overkill for simple topological sort

**Custom implementation**:
- ✅ ~50 lines of clear Go code
- ✅ Full control over error messages
- ✅ No dependency

### Why Not go-git?

**go-git**:
- ❌ 5+ MB dependency
- ❌ Complex API
- ❌ Not needed for simple `init/add/commit`

**os/exec with git**:
- ✅ Simple, 10 lines of code
- ✅ Uses user's git installation
- ✅ Easy to debug
- ⚠️ Requires git installed (acceptable trade-off)

### Why Not BDD Testing?

**ginkgo/gomega**:
- ❌ Non-standard in Go community
- ❌ Adds complexity
- ❌ Overkill for this project

**stdlib testing + testify**:
- ✅ Standard Go approach
- ✅ Table-driven tests
- ✅ Clear, readable

---

## Future Considerations

### Potential Additions (v1.1+)

1. **Terminal UI**: `github.com/charmbracelet/bubbletea` (for interactive mode)
2. **Progress Bars**: `github.com/cheggaaa/pb` (for long operations)
3. **Config Linting**: `mvdan.cc/sh/v3` (shell script parsing)
4. **HTTP Client**: `net/http` (for remote manifest fetching)

### NOT Planning to Add

- ❌ ORM (no database)
- ❌ Web framework (no web UI)
- ❌ gRPC (no RPC)
- ❌ Protobuf (no binary protocol)

---

## Summary

**Total External Dependencies**: 6 (4 production, 2 dev)

**Philosophy**: Minimal, well-chosen dependencies. Prefer stdlib and custom implementations for simple tasks.

**Quality Bar**: All dependencies must be:
- ✅ Actively maintained
- ✅ Widely used in Go community
- ✅ Permissive license (MIT, Apache 2.0, BSD)
- ✅ No C dependencies (pure Go)

---

**Document Status**: Ready for review
**Next Steps**: Write README.md for user-facing documentation
