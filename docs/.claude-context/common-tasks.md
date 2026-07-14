# Common Tasks - gzh-cli-shellforge

## Architecture: Clean Architecture Layers

```text
cmd/shellforge/          # CLI entry point
internal/app/            # Application services (use cases)
  ├── builder.go           # Manifest → shell config builder
  ├── deploy_service.go    # Deploy configs to system
  ├── backup_service.go    # Backup before changes
  ├── diff_service.go      # Show config diff
  ├── doctor_service.go    # Diagnose system health
  ├── validator.go         # Manifest validation
  ├── prereq_validator.go  # Prerequisite checks
  └── migration_service.go # Config migration
internal/infra/          # Infrastructure adapters
  ├── yamlparser/           # YAML manifest parser
  ├── rcparser/             # .bashrc/.zshrc parser
  ├── filesystem/           # File read/write
  ├── template/             # Template rendering
  ├── snapshot/             # State snapshots
  ├── diffcomparator/       # Diff engine
  └── git/                  # Git operations
```

## Manifest Format

Shellforge reads YAML manifests describing shell configurations:

```yaml
version: "1.0"
shell: zsh
modules:
  - name: aliases
    entries:
      - key: ll
        value: "ls -la"
  - name: path
    entries:
      - key: GOPATH
        value: "$HOME/go"
```

## CLI Commands

| Command | Purpose |
|---------|---------|
| `shellforge build` | Build shell config from manifest |
| `shellforge deploy` | Deploy config to system |
| `shellforge diff` | Show changes vs current config |
| `shellforge backup` | Backup current shell config |
| `shellforge doctor` | Diagnose system prerequisites |
| `shellforge validate` | Validate manifest file |

## Adding a New Infrastructure Adapter

1. Create package under `internal/infra/<name>/`
2. Define interface in `internal/app/` that the adapter satisfies
3. Implement the adapter
4. Wire in `cmd/shellforge/main.go` or service constructors

## Validation Pipeline

Manifest validation flows through multiple validators:

```go
validator := app.NewValidator()
validator.AddRule(manifest_validators.RequiredFields{})
validator.AddRule(manifest_validators.ShellSupport{})
validator.AddRule(manifest_validators.PathExists{})
```
