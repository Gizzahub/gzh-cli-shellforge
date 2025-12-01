# Shellforge Public API

> Using Shellforge as a Go library

This document describes the public API for embedding Shellforge in your own Go applications.

---

## Overview

Shellforge can be used in two ways:
1. **As a CLI tool** - The `gz-shellforge` binary
2. **As a Go library** - Import `pkg/cmd` in your Go code

The public API allows you to:
- Embed Shellforge commands in your own CLI
- Integrate Shellforge functionality into your Go applications
- Build custom tools on top of Shellforge

---

## Installation

### As a Library

```bash
go get github.com/gizzahub/gzh-cli-shellforge
```

### Minimum Go Version

- **Go 1.21+** required

---

## Public Package: `pkg/cmd`

### Import

```go
import "github.com/gizzahub/gzh-cli-shellforge/pkg/cmd"
```

### Functions

#### `NewRootCmd() *cobra.Command`

Returns the root Shellforge command for embedding in other CLIs.

**Signature:**
```go
func NewRootCmd() *cobra.Command
```

**Returns:**
- `*cobra.Command` - The root command with all subcommands attached

**Example:**
```go
package main

import (
    "os"

    "github.com/gizzahub/gzh-cli-shellforge/pkg/cmd"
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "mycli",
        Short: "My custom CLI tool",
    }

    // Add shellforge as a subcommand
    shellforgeCmd := cmd.NewRootCmd()
    shellforgeCmd.Use = "shellforge"  // Customize if needed
    rootCmd.AddCommand(shellforgeCmd)

    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

**Usage:**
```bash
# Your custom CLI with embedded shellforge
mycli shellforge build --os Mac --output ~/.zshrc
mycli shellforge validate
mycli shellforge list
```

---

## Use Cases

### 1. Embed in Existing CLI

Integrate Shellforge into your existing command-line tool:

```go
package main

import (
    "github.com/gizzahub/gzh-cli-shellforge/pkg/cmd"
    "github.com/spf13/cobra"
)

func main() {
    // Your existing root command
    rootCmd := &cobra.Command{
        Use:   "devtools",
        Short: "Developer tools suite",
    }

    // Add your own commands
    rootCmd.AddCommand(deployCmd)
    rootCmd.AddCommand(testCmd)

    // Add shellforge
    shellforgeCmd := cmd.NewRootCmd()
    rootCmd.AddCommand(shellforgeCmd)

    rootCmd.Execute()
}
```

**Result:**
```bash
devtools deploy      # Your command
devtools test        # Your command
devtools shellforge  # Embedded shellforge
```

### 2. Custom Command Name

Change the command name to fit your CLI naming:

```go
shellforgeCmd := cmd.NewRootCmd()
shellforgeCmd.Use = "shell"  // Rename to 'shell'
rootCmd.AddCommand(shellforgeCmd)
```

**Usage:**
```bash
mycli shell build --os Mac
mycli shell validate
```

### 3. Programmatic Access (Advanced)

For direct programmatic use, you can access internal services:

> **Note**: Internal packages (`internal/*`) are not guaranteed stable. Use at your own risk.

```go
import (
    "github.com/gizzahub/gzh-cli-shellforge/internal/app"
    "github.com/gizzahub/gzh-cli-shellforge/internal/infra/yamlparser"
    "github.com/gizzahub/gzh-cli-shellforge/internal/infra/filesystem"
    "github.com/spf13/afero"
)

func buildShellConfig(manifestPath, configDir, targetOS string) (string, error) {
    // Setup infrastructure
    fs := afero.NewOsFs()
    parser := yamlparser.New()
    reader := filesystem.NewReader(fs)
    writer := filesystem.NewWriter(fs)

    // Create service
    builder := app.NewBuilderService(parser, reader, writer)

    // Build
    opts := app.BuildOptions{
        ConfigDir: configDir,
        Manifest:  manifestPath,
        OS:        targetOS,
        DryRun:    false,
        Output:    "output.sh",
    }

    return builder.Build(opts)
}
```

**⚠️ Warning**: Internal API may change without notice. Stick to `pkg/cmd` for stability.

---

## Architecture for Library Users

### Recommended Approach

**Use the public `pkg/cmd` API:**
```
Your CLI → pkg/cmd.NewRootCmd() → Shellforge Commands
```

**Benefits:**
- ✅ Stable API
- ✅ All commands work
- ✅ Proper flag parsing
- ✅ Help text and completion
- ✅ No maintenance burden

### Advanced Approach (Not Recommended)

**Direct use of internal services:**
```
Your Code → internal/app → internal/domain
```

**Drawbacks:**
- ⚠️ Unstable API
- ⚠️ May break between versions
- ⚠️ Requires understanding of internal architecture
- ⚠️ More maintenance work

---

## Examples

### Complete Example: Custom DevTools CLI

```go
package main

import (
    "fmt"
    "os"

    "github.com/gizzahub/gzh-cli-shellforge/pkg/cmd"
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "devtools",
        Short: "Development tools for my project",
        Long:  `A suite of development tools including shell config management.`,
    }

    // Add custom deploy command
    deployCmd := &cobra.Command{
        Use:   "deploy",
        Short: "Deploy the application",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Deploying...")
        },
    }

    // Add custom test command
    testCmd := &cobra.Command{
        Use:   "test",
        Short: "Run tests",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Running tests...")
        },
    }

    // Add shellforge
    shellforgeCmd := cmd.NewRootCmd()
    shellforgeCmd.Use = "shell"
    shellforgeCmd.Short = "Manage shell configurations"

    // Add all commands to root
    rootCmd.AddCommand(deployCmd)
    rootCmd.AddCommand(testCmd)
    rootCmd.AddCommand(shellforgeCmd)

    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

**Usage:**
```bash
# Your custom commands
devtools deploy
devtools test

# Shellforge embedded
devtools shell build --os Mac --output ~/.zshrc
devtools shell validate
devtools shell list --filter Mac
devtools shell migrate ~/.zshrc
```

### Example: Version Information

Add version information to the embedded command:

```go
package main

import (
    "github.com/gizzahub/gzh-cli-shellforge/pkg/cmd"
    "github.com/spf13/cobra"
)

const version = "1.0.0"

func main() {
    rootCmd := &cobra.Command{
        Use:     "mycli",
        Version: version,
    }

    shellforgeCmd := cmd.NewRootCmd()
    shellforgeCmd.Version = version
    rootCmd.AddCommand(shellforgeCmd)

    rootCmd.Execute()
}
```

### Example: Custom Help Template

Customize the help output:

```go
shellforgeCmd := cmd.NewRootCmd()

// Custom usage template
customUsage := `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}

{{if .HasAvailableSubCommands}}Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
`

shellforgeCmd.SetUsageTemplate(customUsage)
rootCmd.AddCommand(shellforgeCmd)
```

---

## API Stability

### Stable API (`pkg/cmd`)

- ✅ **Guaranteed stable** across minor versions
- ✅ Breaking changes only in major versions
- ✅ Follows semantic versioning

### Internal API (`internal/*`)

- ⚠️ **No stability guarantee**
- ⚠️ May change without notice
- ⚠️ Use at your own risk

**Recommendation:** Always use `pkg/cmd` unless you have specific needs.

---

## Versioning

Shellforge follows [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backwards-compatible)
- **PATCH**: Bug fixes (backwards-compatible)

**Import versioning:**
```go
// Current (v0.x.x - unstable)
import "github.com/gizzahub/gzh-cli-shellforge/pkg/cmd"

// Future v1 (stable)
import "github.com/gizzahub/gzh-cli-shellforge/v1/pkg/cmd"
```

---

## Dependencies

When using Shellforge as a library, you'll pull in:

### Direct Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/afero` - Filesystem abstraction
- `gopkg.in/yaml.v3` - YAML parsing

### Total Size
- ~3-4 MB added to your binary
- No C dependencies (pure Go)
- Cross-compiles easily

---

## Best Practices

### 1. Use the Public API

```go
// ✅ Good - stable API
import "github.com/gizzahub/gzh-cli-shellforge/pkg/cmd"

// ❌ Avoid - unstable API
import "github.com/gizzahub/gzh-cli-shellforge/internal/app"
```

### 2. Version Pinning

Use go.mod to pin specific versions:

```go
module github.com/yourorg/yourproject

go 1.21

require (
    github.com/gizzahub/gzh-cli-shellforge v0.2.0
)
```

### 3. Error Handling

Always handle errors from Shellforge commands:

```go
shellforgeCmd := cmd.NewRootCmd()
if err := shellforgeCmd.Execute(); err != nil {
    // Handle error appropriately
    log.Fatalf("Shellforge command failed: %v", err)
}
```

### 4. Testing

Mock the Shellforge command in tests:

```go
func TestMyCLI(t *testing.T) {
    // Create root command
    rootCmd := &cobra.Command{Use: "test"}

    // Add shellforge (use real command in tests)
    shellforgeCmd := cmd.NewRootCmd()
    rootCmd.AddCommand(shellforgeCmd)

    // Test command execution
    rootCmd.SetArgs([]string{"shellforge", "validate", "--help"})
    err := rootCmd.Execute()
    assert.NoError(t, err)
}
```

---

## Troubleshooting

### Import Errors

**Problem:**
```
cannot find package "github.com/gizzahub/gzh-cli-shellforge/pkg/cmd"
```

**Solution:**
```bash
go get github.com/gizzahub/gzh-cli-shellforge
go mod tidy
```

### Version Conflicts

**Problem:**
```
conflicting versions of github.com/spf13/cobra
```

**Solution:**
```bash
# Update to compatible versions
go get github.com/spf13/cobra@latest
go mod tidy
```

### Binary Size Concerns

**Problem:** Binary size increased significantly

**Solution:**
```bash
# Build with size optimization
go build -ldflags="-s -w" -o mycli

# Result: ~3-4 MB added (reasonable for full CLI)
```

---

## Migration Guide

### From Direct Binary Execution

**Before:**
```bash
# Running shellforge binary
gz-shellforge build --os Mac
```

**After (Embedded):**
```go
// Programmatic execution
cmd := cmd.NewRootCmd()
cmd.SetArgs([]string{"build", "--os", "Mac"})
cmd.Execute()
```

### From Custom Implementation

If you built custom shell config tooling:

1. Replace with Shellforge embedded command
2. Use `pkg/cmd.NewRootCmd()`
3. Customize command name if needed
4. Keep your other CLI features

**Benefits:**
- ✅ Less code to maintain
- ✅ Battle-tested dependency resolution
- ✅ Free updates and bug fixes
- ✅ Community support

---

## Future API Additions

Planned for future versions:

### v0.3.0
- `pkg/builder` - Direct builder service API
- `pkg/validator` - Direct validator service API

### v1.0.0
- Stable public API with guarantees
- More granular service access
- Plugin system

**Feedback Welcome!** Let us know what API you need: [GitHub Issues](https://github.com/gizzahub/gzh-cli-shellforge/issues)

---

## Support

### Questions?

- **GitHub Discussions**: [Ask questions](https://github.com/gizzahub/gzh-cli-shellforge/discussions)
- **GitHub Issues**: [Report API bugs](https://github.com/gizzahub/gzh-cli-shellforge/issues)
- **Examples**: See `examples/` directory

### Resources

- **[Architecture](00-architecture.md)** - Understanding the internals
- **[Contributing](CONTRIBUTING.md)** - Contributing to the API
- **[Tech Stack](50-tech-stack.md)** - Libraries used

---

## Example Projects

Projects using Shellforge as a library:

> Coming soon! Be the first - submit a PR to add your project here.

---

**Last Updated**: 2025-12-01
**API Version**: v0.2.0 (unstable)
**Stable API**: Planned for v1.0.0
