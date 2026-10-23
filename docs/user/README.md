# User Documentation

Welcome to Shellforge user documentation! Find everything you need to effectively use Shellforge for managing your shell configurations.

______________________________________________________________________

## 🚀 Getting Started

New to Shellforge? Start here:

1. **[Quick Start Guide](00-quick-start.md)** ⭐ **Start Here!**

   - Install in 1 command
   - Migrate your config in 5 minutes
   - Your first successful build

1. **[Installation Guide](10-installation.md)**

   - Detailed installation for all platforms
   - Troubleshooting installation issues
   - Verifying installation

______________________________________________________________________

## 📚 Core Documentation

### Essential Guides

- **[Basic Usage](20-basic-usage.md)**

  - Essential commands (validate, build, list)
  - Common flags and options
  - Your first workflow

- **[Complete Workflows](30-workflows.md)**

  - Step-by-step workflow guide
  - Migration → Build → Deploy
  - Best practices

- **[Command Reference](40-command-reference.md)** 📖

  - All commands with examples
  - Real-world scenarios
  - Tips and tricks

### Advanced Topics

- **[Advanced Usage](50-advanced-usage.md)**

  - CI/CD integration
  - Multi-OS workflows
  - Custom templates
  - Scripting with Shellforge

- **[Troubleshooting](60-troubleshooting.md)** 🔧

  - Common issues and solutions
  - Error message reference
  - Debug techniques

- **[FAQ](70-faq.md)** ❓

  - Frequently asked questions
  - "Why use Shellforge?"
  - Common misconceptions

______________________________________________________________________

## 📖 Documentation by Task

### I Want To...

#### Get Started

- [Install Shellforge](10-installation.md#installation)
- [Migrate my first config](00-quick-start.md#step-2-migrate-to-modular-structure)
- [Understand what Shellforge does](00-quick-start.md#what-is-shellforge)

#### Daily Usage

- [Build my shell config](40-command-reference.md#build-command)
- [Validate my manifest](40-command-reference.md#validate-command)
- [List my modules](40-command-reference.md#list-command)
- [Add a new module](00-quick-start.md#add-a-new-module)

#### Advanced Features

- [Use templates](40-command-reference.md#template-command)
- [Backup my config](40-command-reference.md#backup-command)
- [Restore from backup](40-command-reference.md#restore-command)
- [Compare configs](40-command-reference.md#diff-command)
- [Manage multiple OSes](50-advanced-usage.md#multi-os-setup)

#### Troubleshooting

- [Fix circular dependencies](60-troubleshooting.md#circular-dependencies)
- [Debug build errors](60-troubleshooting.md#build-errors)
- [Resolve validation issues](60-troubleshooting.md#validation-errors)

______________________________________________________________________

## 🎯 Documentation by User Type

### New Users

Start with these in order:

1. [Quick Start Guide](00-quick-start.md) - 5 minutes
1. [Basic Usage](20-basic-usage.md) - 10 minutes
1. [Workflows](30-workflows.md) - 15 minutes

**Time to productivity: ~30 minutes**

### Regular Users

Keep these bookmarked:

- [Command Reference](40-command-reference.md) - Quick lookup
- [Troubleshooting](60-troubleshooting.md) - Common issues
- [FAQ](70-faq.md) - Quick answers

### Power Users

Explore advanced topics:

- [Advanced Usage](50-advanced-usage.md) - CI/CD, automation
- [API Reference](../reference/api.md) - Library integration
- [Template Reference](../reference/template-reference.md) - Custom templates

______________________________________________________________________

## 📋 Complete Documentation Index

### Getting Started (00-19)

- [00 - Quick Start Guide](00-quick-start.md) ⭐
- [10 - Installation Guide](10-installation.md)

### Core Usage (20-49)

- [20 - Basic Usage](20-basic-usage.md)
- [30 - Complete Workflows](30-workflows.md)
- [40 - Command Reference](40-command-reference.md) 📖

### Advanced Topics (50-79)

- [50 - Advanced Usage](50-advanced-usage.md)
- [60 - Troubleshooting](60-troubleshooting.md) 🔧
- [70 - FAQ](70-faq.md) ❓

### Reference (80-89)

- [80 - Changelog](80-changelog.md)

______________________________________________________________________

## 🔍 Quick Command Lookup

### Most Common Commands

```bash
# Validate your configuration
gz-shellforge validate

# Build for your OS
gz-shellforge build --os Mac --output ~/.zshrc

# List all modules
gz-shellforge list

# Migrate existing config
gz-shellforge migrate ~/.zshrc

# Compare original vs generated
gz-shellforge diff ~/.zshrc ~/.zshrc.new
```

[See all commands →](40-command-reference.md)

______________________________________________________________________

## 📊 Feature Overview

### Core Features

- ✅ **Dependency Resolution** - Automatic topological sort
- ✅ **OS Filtering** - Mac/Linux module selection
- ✅ **Validation** - Pre-build error detection
- ✅ **Dry Run** - Preview before deploying

### Migration & Build

- ✅ **Auto Migration** - Convert monolithic configs
- ✅ **Section Detection** - Smart section parsing
- ✅ **Dependency Inference** - Auto-detect requirements
- ✅ **Template Generation** - Pre-defined module templates

### Backup & Restore

- ✅ **Git-Backed Versioning** - Full version history
- ✅ **Snapshot Management** - Timestamped backups
- ✅ **Retention Policies** - Auto-cleanup old backups
- ✅ **Restore with Safety** - Automatic safety backups

### Comparison & Validation

- ✅ **Diff Formats** - Summary, unified, context, side-by-side
- ✅ **LCS Algorithm** - Accurate change detection
- ✅ **Statistics** - Line-level change metrics

______________________________________________________________________

## 🎓 Learning Path

### Beginner Path (Day 1)

```
Quick Start (5 min)
    ↓
Install & Verify (5 min)
    ↓
First Migration (10 min)
    ↓
First Build (5 min)
    ↓
✅ You can now use Shellforge!
```

### Intermediate Path (Week 1)

```
Learn All Commands (1 hour)
    ↓
Complete Workflows (1 hour)
    ↓
Practice Daily Usage (1 week)
    ↓
✅ Comfortable with daily tasks
```

### Advanced Path (Month 1)

```
Advanced Features (2 hours)
    ↓
CI/CD Integration (2 hours)
    ↓
Custom Templates (1 hour)
    ↓
✅ Master of Shellforge
```

______________________________________________________________________

## 🆘 Getting Help

### Self-Service

1. Check [Troubleshooting Guide](60-troubleshooting.md)
1. Search [FAQ](70-faq.md)
1. Read [Command Reference](40-command-reference.md)

### Command Line Help

```bash
# General help
gz-shellforge --help

# Command-specific help
gz-shellforge build --help
gz-shellforge migrate --help
```

### Community & Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/gizzahub/gzh-cli-shellforge/issues)
- **Discussions**: [Ask questions](https://github.com/gizzahub/gzh-cli-shellforge/discussions)
- **Examples**: [Working examples](../../examples/)

______________________________________________________________________

## 📝 Documentation Notes

### Documentation Status

| Document          | Status      | Last Updated |
| ----------------- | ----------- | ------------ |
| Quick Start       | ✅ Complete | 2025-12-01   |
| Installation      | ⚠️ TODO     | -            |
| Basic Usage       | ⚠️ TODO     | -            |
| Workflows         | ✅ Complete | 2025-11-28   |
| Command Reference | ✅ Complete | 2025-11-28   |
| Advanced Usage    | ⚠️ TODO     | -            |
| Troubleshooting   | ⚠️ TODO     | -            |
| FAQ               | ⚠️ TODO     | -            |

### Contributing to Documentation

Found an error? Want to improve documentation?

1. [Read Contributing Guide](../developer/30-contributing.md)
1. Submit a pull request
1. Or open an issue

______________________________________________________________________

## 🗺️ Related Documentation

- **[Developer Documentation](../developer/)** - For contributors
- **[API Reference](../reference/)** - For library users
- **[Design Documents](../design/)** - Architecture and decisions

______________________________________________________________________

**Last Updated**: 2025-12-01
**Documentation Version**: 0.2.0
