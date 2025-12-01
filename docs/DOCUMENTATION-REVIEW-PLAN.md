# Documentation Review and Improvement Plan

**Date**: 2025-12-01
**Status**: Analysis & Planning
**Purpose**: Separate human-focused and LLM-focused documentation

---

## 1. Current Documentation Analysis

### 1.1 Existing Documentation Inventory

| File | Target Audience | Content Type | Quality | Issues |
|------|----------------|--------------|---------|--------|
| **README.md** | Human Users | User Guide | ⭐⭐⭐⭐☆ | Mixed audience, too long (837 lines) |
| **CLAUDE.md** | LLM (Claude) | Development Guide | ⭐⭐⭐⭐⭐ | Good, but dev-focused |
| **ARCHITECTURE.md** | Developers | Architecture | ⭐⭐⭐⭐☆ | Technical, good structure |
| **PRD.md** | Product/Dev | Requirements | ⭐⭐⭐☆☆ | Not read yet |
| **REQUIREMENTS.md** | Developers | Functional Spec | ⭐⭐⭐☆☆ | Not read yet |
| **TECH_STACK.md** | Developers | Tech Decisions | ⭐⭐⭐⭐☆ | Good for devs |
| **examples/WORKFLOW.md** | Human Users | Tutorial | ⭐⭐⭐⭐⭐ | Excellent workflow guide |
| **examples/CLI-EXAMPLES.md** | Human Users | Quick Reference | ⭐⭐⭐⭐⭐ | Excellent command reference |
| **docs/BENCHMARKS.md** | Developers/Users | Performance | ⭐⭐⭐☆☆ | Niche audience |
| **CHANGELOG.md** | Users | Release Notes | ⭐⭐⭐☆☆ | Standard |
| **TODO.md** | Developers | Planning | ⭐⭐☆☆☆ | Development artifact |

### 1.2 Critical Findings

#### ✅ **Strengths**
1. **Excellent CLI documentation**: CLI-EXAMPLES.md and WORKFLOW.md are comprehensive
2. **Good architecture docs**: CLAUDE.md provides clear dev guidance
3. **Comprehensive README**: Covers all features and commands
4. **Real examples**: examples/ directory with working code

#### ❌ **Problems**
1. **No clear separation**: Human vs LLM documentation mixed
2. **README too long**: 837 lines - information overload
3. **Missing docs**:
   - Quick start guide (5-minute setup)
   - Troubleshooting guide
   - API reference (if used as library)
   - Contributing guide
   - Security documentation
4. **Poor discoverability**: Users don't know where to find what
5. **LLM context pollution**: Human docs consume LLM context unnecessarily
6. **No user personas**: Documentation not targeted to specific user types

---

## 2. User Personas & Documentation Needs

### 2.1 Human User Personas

#### **Persona 1: New User (First-time user)**
- **Goal**: Quick setup and first successful build
- **Needs**:
  - 5-minute quick start guide
  - Clear installation instructions
  - Simple example that works
  - What is this tool and why use it?
- **Current pain**: README is too long, overwhelming

#### **Persona 2: Regular User (Daily user)**
- **Goal**: Efficient workflow, command reference
- **Needs**:
  - Quick command reference
  - Common workflows
  - Troubleshooting guide
  - Tips and tricks
- **Current pain**: Must search through long README

#### **Persona 3: Advanced User (Power user)**
- **Goal**: Automation, integration, customization
- **Needs**:
  - Advanced features
  - CI/CD integration
  - Scripting examples
  - API reference (pkg/cmd)
- **Current pain**: No advanced usage documentation

#### **Persona 4: Contributor (Developer)**
- **Goal**: Understand codebase, contribute features
- **Needs**:
  - Architecture overview
  - Development setup
  - Testing guide
  - Contribution workflow
- **Current pain**: Mixed with user docs

### 2.2 LLM User Personas

#### **Persona 5: LLM Developer Assistant (Claude Code)**
- **Goal**: Understand codebase structure, make changes correctly
- **Needs**:
  - Architecture patterns (Hexagonal)
  - Dependency rules
  - Testing patterns
  - Build commands
  - Code conventions
- **Current state**: CLAUDE.md is excellent

#### **Persona 6: LLM User Assistant (General AI)**
- **Goal**: Help users with commands and workflows
- **Needs**:
  - Command syntax
  - Common use cases
  - Error messages and solutions
  - Feature availability
- **Current pain**: Must parse long README

---

## 3. Proposed Documentation Structure

### 3.1 Human-Focused Documentation

```
docs/
├── user/                           # User documentation
│   ├── README.md                   # Landing page for users
│   ├── 00-quick-start.md          # 5-minute setup guide ⚠️ NEW
│   ├── 10-installation.md         # Detailed installation ⚠️ NEW
│   ├── 20-basic-usage.md          # Basic commands ⚠️ NEW
│   ├── 30-workflows.md            # Common workflows (moved from examples/)
│   ├── 40-command-reference.md    # All commands (moved from examples/)
│   ├── 50-advanced-usage.md       # Advanced features ⚠️ NEW
│   ├── 60-troubleshooting.md      # Common issues ⚠️ NEW
│   ├── 70-faq.md                  # Frequently Asked Questions ⚠️ NEW
│   └── 80-changelog.md            # Release notes (symlink)
│
├── developer/                      # Developer documentation
│   ├── README.md                   # Landing page for developers
│   ├── 00-architecture.md         # System architecture (moved)
│   ├── 10-development-setup.md    # Dev environment setup ⚠️ NEW
│   ├── 20-testing-guide.md        # Testing strategy ⚠️ NEW
│   ├── 30-contributing.md         # Contribution guide ⚠️ NEW
│   ├── 40-code-style.md           # Code conventions ⚠️ NEW
│   ├── 50-tech-stack.md           # Technology decisions (moved)
│   └── 60-benchmarks.md           # Performance data (moved)
│
├── reference/                      # Reference documentation
│   ├── api.md                      # pkg/cmd API reference ⚠️ NEW
│   ├── manifest-schema.md          # YAML manifest spec ⚠️ NEW
│   ├── template-reference.md       # Template system ⚠️ NEW
│   └── error-codes.md              # Error reference ⚠️ NEW
│
└── design/                         # Design documents
    ├── prd.md                      # Product requirements (moved)
    ├── requirements.md             # Functional spec (moved)
    └── decisions/                  # Architecture Decision Records ⚠️ NEW
        ├── 001-hexagonal-architecture.md
        ├── 002-dependency-resolution.md
        └── README.md
```

### 3.2 LLM-Focused Documentation

```
.claude/
├── CONTEXT.md                      # Main LLM context (new)
├── DEVELOPMENT.md                  # Current CLAUDE.md renamed
├── commands/                       # Custom slash commands
├── ctx/                            # Context files
└── schemas/                        # Document schemas
```

**CONTEXT.md** (New LLM landing page):
```markdown
# Shellforge Context for LLM

## What is Shellforge?
Modular shell configuration builder with dependency resolution.

## Quick Reference
- Binary: gz-shellforge
- Code: cmd/shellforge/main.go
- Architecture: Hexagonal (4 layers)
- Commands: build, validate, list, migrate, template, backup, restore, cleanup, diff

## For Development Tasks
See: DEVELOPMENT.md (detailed developer guide)

## For User Assistance
See: docs/user/command-reference.md (command syntax)

## Critical Rules
1. Always use `make build`, never `go build`
2. Domain layer NEVER imports app/infra/cli
3. Add tests for all new features
4. Follow dependency rules strictly
```

### 3.3 Root Documentation

```
/
├── README.md                       # Short landing page (150 lines max)
├── LICENSE
├── CHANGELOG.md
├── CONTRIBUTING.md                 # Link to docs/dev/contributing.md
└── docs/                           # All detailed docs moved here
```

**New README.md** (Simplified):
```markdown
# Shellforge

Build tool for modular shell configurations with automatic dependency resolution.

## Quick Start

[5-minute setup guide](docs/user/00-quick-start.md)

## Documentation

- **Users**: [User Guide](docs/user/)
- **Developers**: [Developer Guide](docs/dev/)
- **API Reference**: [API Docs](docs/reference/)

## Features

- ✅ Dependency resolution
- ✅ OS-specific filtering
- ✅ Migration from monolithic configs
- ✅ Template generation
- ✅ Backup/restore system

[See all features →](docs/user/README.md)

## Installation

```bash
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest
```

[Detailed installation →](docs/user/10-installation.md)

## Support

- [Documentation](docs/)
- [Issues](https://github.com/gizzahub/gzh-cli-shellforge/issues)
- [Contributing](CONTRIBUTING.md)

## License

MIT
```

---

## 4. Documentation Strategy

### 4.1 Separation Principles

#### Human Documentation
- **Location**: `docs/user/`, `docs/dev/`
- **Style**: Tutorial, narrative, example-driven
- **Format**: Markdown with rich formatting
- **Length**: As long as needed for clarity
- **Audience**: People learning or using the tool
- **Language**: Korean for policies, English for technical docs

#### LLM Documentation
- **Location**: `.claude/`, `CLAUDE.md` → `.claude/DEVELOPMENT.md`
- **Style**: Concise, imperative, rule-based
- **Format**: Structured with clear sections
- **Length**: Token-optimized (<10KB preferred)
- **Audience**: AI coding assistants
- **Language**: English only

### 4.2 Content Distribution

| Content Type | Human Docs | LLM Docs | Both |
|--------------|------------|----------|------|
| Quick Start | ✅ | ❌ | |
| Command Reference | ✅ | ✅ | Shared |
| Architecture | ✅ | ✅ | Different level of detail |
| Code Examples | ✅ | ❌ | |
| Build Commands | ❌ | ✅ | |
| Dependency Rules | ❌ | ✅ | |
| API Reference | ✅ | ✅ | Same source |
| Troubleshooting | ✅ | ✅ | Different format |
| Design Decisions | ✅ | ❌ | |

### 4.3 Documentation Quality Metrics

#### For Human Documentation
- ✅ Can a new user complete setup in 5 minutes?
- ✅ Can a regular user find any command in <30 seconds?
- ✅ Are all error messages documented with solutions?
- ✅ Are there examples for every feature?
- ✅ Is the documentation searchable?

#### For LLM Documentation
- ✅ Can LLM find architectural constraints in <100 tokens?
- ✅ Are all critical rules clearly stated?
- ✅ Is context size optimized (<10KB per file)?
- ✅ Are file paths and commands accurate?
- ✅ Is there no ambiguity in instructions?

---

## 5. Implementation Plan

### Phase 1: Reorganization (Week 1)
**Priority**: HIGH
**Effort**: 2-3 hours

- [ ] Create new directory structure
  ```bash
  mkdir -p docs/{user,developer,reference,design/decisions}
  ```
- [ ] Move existing files:
  - [ ] examples/WORKFLOW.md → docs/user/30-workflows.md
  - [ ] examples/CLI-EXAMPLES.md → docs/user/40-command-reference.md
  - [x] ARCHITECTURE.md → docs/dev/00-architecture.md (COMPLETED)
  - [x] TECH_STACK.md → docs/dev/50-tech-stack.md (COMPLETED)
  - [ ] docs/BENCHMARKS.md → docs/dev/60-benchmarks.md
  - [ ] PRD.md → docs/design/prd.md
  - [ ] REQUIREMENTS.md → docs/design/requirements.md
- [ ] Rename CLAUDE.md → .claude/DEVELOPMENT.md
- [ ] Create symlinks for backward compatibility:
  ```bash
  ln -s docs/user/30-workflows.md examples/WORKFLOW.md
  ln -s docs/user/40-command-reference.md examples/CLI-EXAMPLES.md
  ```
- [ ] Update all internal links in moved files

### Phase 2: Create Missing Human Documentation (Week 1-2)
**Priority**: HIGH
**Effort**: 4-6 hours

#### Critical (Must have)
- [ ] **docs/user/00-quick-start.md** (5-minute guide)
  - Install in 1 command
  - Migrate sample file
  - Build first config
  - Done!
- [ ] **docs/user/10-installation.md** (all platforms)
  - macOS (Homebrew, go install, from source)
  - Linux (apt, yum, go install, from source)
  - Verify installation
- [ ] **docs/user/60-troubleshooting.md** (common issues)
  - Installation issues
  - Build errors
  - Validation errors
  - OS detection problems
  - Permission issues
- [ ] **docs/user/README.md** (user docs landing page)
  - Documentation map
  - Quick navigation
  - What to read first

#### Important (Should have)
- [ ] **docs/user/20-basic-usage.md** (essential commands)
  - validate, build, list basics
  - Common flags
  - First workflow
- [ ] **docs/user/50-advanced-usage.md** (power user features)
  - CI/CD integration
  - Multi-OS workflows
  - Custom templates
  - Scripting with gz-shellforge
- [ ] **docs/user/70-faq.md** (frequently asked questions)
  - "Why use Shellforge?"
  - "How is this different from ...?"
  - Common misconceptions

### Phase 3: Create Missing Developer Documentation (Week 2)
**Priority**: MEDIUM
**Effort**: 3-4 hours

- [ ] **docs/dev/README.md** (dev docs landing page)
- [ ] **docs/dev/10-development-setup.md**
  - Prerequisites
  - Clone and build
  - Running tests
  - IDE setup (VSCode, GoLand)
- [ ] **docs/dev/20-testing-guide.md**
  - Testing strategy
  - Writing unit tests
  - Writing integration tests
  - Coverage targets
  - Running benchmarks
- [ ] **docs/dev/30-contributing.md**
  - How to contribute
  - Code review process
  - PR guidelines
  - Issue triage
- [ ] **docs/dev/40-code-style.md**
  - Go conventions
  - Project-specific patterns
  - Naming conventions
  - Error handling

### Phase 4: Create Reference Documentation (Week 2-3)
**Priority**: MEDIUM
**Effort**: 4-5 hours

- [ ] **docs/reference/api.md** (pkg/cmd public API)
  - Exported types
  - Public functions
  - Usage examples
  - Integration guide
- [ ] **docs/reference/manifest-schema.md**
  - Complete YAML schema
  - Field descriptions
  - Validation rules
  - Examples
- [ ] **docs/reference/template-reference.md**
  - All template types
  - Required fields
  - Optional fields
  - Custom templates
- [ ] **docs/reference/error-codes.md**
  - All error types
  - Error codes
  - Causes
  - Solutions

### Phase 5: Create LLM Documentation (Week 3)
**Priority**: MEDIUM
**Effort**: 2-3 hours

- [ ] **/.claude/CONTEXT.md** (LLM landing page)
  - Project overview (50 lines)
  - Quick reference
  - Where to find detailed info
  - Critical rules
- [ ] Update **/.claude/DEVELOPMENT.md** (current CLAUDE.md)
  - Remove user-facing content
  - Add more build/test commands
  - Add more code patterns
  - Optimize for token usage
- [ ] Create **/.claude/COMMANDS.md** (command reference for LLM)
  - All CLI commands with syntax
  - Common flag combinations
  - Error message reference

### Phase 6: Simplify Root README (Week 3)
**Priority**: HIGH
**Effort**: 1-2 hours

- [ ] Rewrite README.md (target: 150-200 lines)
  - What is Shellforge? (3 sentences)
  - Quick Start (5 lines)
  - Links to detailed docs
  - Feature overview (bullet points only)
  - Installation (1 command)
  - Support links
- [ ] Create CONTRIBUTING.md
  - Link to docs/dev/contributing.md
  - Code of conduct
  - Quick contribution checklist

### Phase 7: Documentation Automation (Week 4)
**Priority**: LOW
**Effort**: 2-3 hours

- [ ] Add documentation validation to CI
  - Check for broken links
  - Verify code examples
  - Check for outdated versions
- [ ] Create documentation generator for API reference
  - Auto-generate from godoc
  - Keep docs/reference/api.md in sync
- [ ] Add documentation coverage metrics
  - Track which features are documented
  - Identify missing docs

---

## 6. Documentation Maintenance Guidelines

### 6.1 When to Update Documentation

| Change Type | Documentation to Update |
|-------------|------------------------|
| New feature | User guide, Command reference, API reference, LLM context |
| Bug fix | Troubleshooting guide (if user-facing) |
| Breaking change | Migration guide, CHANGELOG, Quick start |
| New command | Command reference, LLM commands, CLI examples |
| Architecture change | Architecture doc, LLM development doc |
| Performance improvement | Benchmarks |
| New error message | Error codes reference, Troubleshooting |

### 6.2 Documentation Review Checklist

Before merging any PR that adds/changes documentation:

- [ ] Is the target audience clear? (User vs Developer vs LLM)
- [ ] Is the file in the correct directory?
- [ ] Are all code examples tested and working?
- [ ] Are all links valid?
- [ ] Is the language appropriate? (Korean for policies, English for technical)
- [ ] Is it following the file size limits? (<10KB for LLM docs)
- [ ] Are there screenshots where helpful? (user docs)
- [ ] Is it indexed in the parent README?

### 6.3 Documentation Style Guide

#### User Documentation
- **Tone**: Friendly, helpful, encouraging
- **Structure**: Problem → Solution → Example
- **Length**: As long as needed, use sections
- **Code examples**: Always include, with comments
- **Images**: Use when beneficial (screenshots, diagrams)

#### Developer Documentation
- **Tone**: Technical, precise, authoritative
- **Structure**: Concept → Implementation → Rationale
- **Length**: Detailed but focused
- **Code examples**: Show best practices
- **Diagrams**: Use for architecture and flows

#### LLM Documentation
- **Tone**: Imperative, rule-based
- **Structure**: Rule → Example → Exception
- **Length**: Token-optimized (<10KB)
- **Code examples**: Only essential patterns
- **Diagrams**: No (not readable by LLM)

---

## 7. Success Metrics

### 7.1 User Success Metrics

- [ ] **New user success**: 80% complete quick-start in <10 minutes
- [ ] **Command discoverability**: Users find any command in <1 minute
- [ ] **Issue resolution**: 80% of issues have documented solutions
- [ ] **Documentation feedback**: 4+ star rating on documentation clarity

### 7.2 Developer Success Metrics

- [ ] **Onboarding time**: New contributors make first PR in <2 hours
- [ ] **Architecture understanding**: Developers correctly follow layer rules
- [ ] **Test coverage**: All new features have tests (enforced by docs)

### 7.3 LLM Success Metrics

- [ ] **Context efficiency**: LLM docs <30KB total
- [ ] **Accuracy**: LLM generates correct commands 95% of time
- [ ] **Rule compliance**: LLM follows architecture rules 100%

---

## 8. Migration Strategy

### 8.1 Backward Compatibility

During transition period (1 month):
- Keep old files with deprecation notice
- Create symlinks to new locations
- Update all internal links
- Add redirects in README

Example deprecation notice:
```markdown
⚠️ **DEPRECATED**: This file has moved to `docs/user/30-workflows.md`

This file will be removed in v0.6.0. Please update your bookmarks.
```

### 8.2 Communication Plan

1. **GitHub Release Note** (v0.5.1):
   ```markdown
   ## Documentation Reorganization

   We've reorganized documentation for better discoverability:
   - User documentation: docs/user/
   - Developer documentation: docs/dev/
   - API reference: docs/reference/

   Old links redirected for 1 month. Please update bookmarks.
   ```

2. **README Banner** (1 month):
   ```markdown
   > 📚 **Documentation has been reorganized!** See the new structure in [docs/](docs/)
   ```

3. **PR Template Update**:
   Add documentation checklist to PR template.

---

## 9. Appendix: Documentation Templates

### Template 1: User Guide Page

```markdown
# [Feature Name]

> Quick one-line description

## What is it?

[2-3 sentences explaining the feature]

## When to use it

[Use cases, scenarios]

## Quick Example

```bash
# Simple example that works
gz-shellforge command --flag value
```

## Step-by-Step Guide

### Step 1: [Action]
[Instructions]

### Step 2: [Action]
[Instructions]

## Advanced Usage

[Optional: advanced scenarios]

## Common Issues

| Problem | Solution |
|---------|----------|
| [Issue] | [Fix] |

## Related Commands

- [Command 1](link)
- [Command 2](link)

## Next Steps

- [What to do next]
```

### Template 2: Developer Guide Page

```markdown
# [Technical Topic]

**Status**: [Draft/Stable/Deprecated]
**Last Updated**: YYYY-MM-DD

## Overview

[Technical description]

## Architecture

[Architecture diagram or description]

## Implementation

### Code Structure
[Package organization]

### Key Components
[Main types/interfaces]

### Design Decisions
[Why this approach]

## Usage

```go
// Code example
```

## Testing

```go
// Test example
```

## References

- [Related doc]
- [External resource]
```

### Template 3: LLM Context Page

```markdown
# [Component] Context

## Quick Facts
- Purpose: [one sentence]
- Location: path/to/file
- Dependencies: [list]

## Critical Rules
1. [Rule with consequence]
2. [Rule with consequence]

## Common Operations
[Command or pattern]

## Error Patterns
[Common mistakes to avoid]
```

---

## 10. Review Checklist

Before considering documentation complete:

### Content Completeness
- [ ] All features documented
- [ ] All commands documented
- [ ] All errors documented
- [ ] All workflows documented

### Quality
- [ ] All code examples tested
- [ ] All links verified
- [ ] Spelling/grammar checked
- [ ] Screenshots up to date

### Organization
- [ ] Clear information architecture
- [ ] Logical file naming
- [ ] Proper cross-linking
- [ ] Search-friendly structure

### Accessibility
- [ ] New users can start in 5 minutes
- [ ] Regular users can find commands quickly
- [ ] Developers can understand architecture
- [ ] LLM can find rules efficiently

---

**End of Documentation Review and Improvement Plan**
