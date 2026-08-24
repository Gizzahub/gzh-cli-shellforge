<!-- size-limit: 1000 -->

<!-- A changelog is an append-only record, not a guide read front to back, so the
     default 500-line document budget models the wrong genre. 1000 is the
     family-wide ceiling, matching gzh-cli-gitforge so there is one rule rather
     than one number per repository.

     It is a ceiling, not an exemption, and this repository's own history says it
     is not tight: the nine releases split out below were 780 lines in total and
     the largest single release was 128. Hitting 1000 therefore means a release
     is overdue, not that this file needs trimming — cut the release and move
     that line into docs/changelog/.

     Do not shrink the budget to a value a single batch can exhaust mid-write.
     Exceeding it is a hard error that blocks edits to this file, and the remedy
     is itself an edit to this file — that is how this file sat unchanged from
     2026-06-01 to 2026-08-25 at 587 prose lines against a 500-line limit, with
     the prepare command shipped and unrecorded. -->
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
## [Unreleased]

### Added

- **System-Wide Config Targets** (`etc-profile`, `etc-zshrc`, `etc-zshenv`): modules can now target `/etc/profile`, `/etc/zshrc`, and `/etc/zsh/zshenv`
  - `IsSystemTarget(name)` in `internal/domain` identifies system targets across all shell types
  - System targets resolve directly to absolute paths; `TargetResolver.GetRelativePath` returns the absolute path for these targets — callers must use `filepath.IsAbs` and must not join with HomeDir
  - `GetValidTargets` and `IsValidTarget` now include system targets for every shell type
  - Deploy service: absolute `DestPath` is used directly (not joined with HomeDir); `filepath.IsAbs` is the detection mechanism
  - Permission check before any system-file write; clear `requires elevated privileges — re-run with sudo` message on denial (no silent failure, no auto-escalation)
  - Backup **forced** for system files before overwrite, regardless of `--backup` flag; prevents a bad `/etc/profile` from breaking all-user logins
  - Dry-run mode carries a sudo hint in the `Error` field for system targets
  - `PermissionChecker` interface added to `DeployService` for test injection; `NewDeployServiceWithChecker` constructor



- **BSD OS Detection** (per-OS granularity): `DetectOS()` now returns `"FreeBSD"`, `"OpenBSD"`, `"NetBSD"` instead of a single `"BSD"` umbrella
  - `os-detection.sh` example module updated with per-OS BSD branches (`uname -s` → `FreeBSD`, `OpenBSD`, `NetBSD`)
  - `InferOSSupport()` in migration now detects `FreeBSD)`, `OpenBSD)`, `NetBSD)` case branches when inferring manifest `os:` values
  - Manifest modules can now target specific BSDs: `os: [FreeBSD]` or `os: [FreeBSD, OpenBSD]`
  - PRODUCT.md "지원 환경" updated to list FreeBSD/OpenBSD/NetBSD

- **Pluggable Validation Pipeline** (Phase 1): Introduced a `Validator` interface and `ValidationPipeline` in `internal/app`
  - `Finding` struct with `error`/`warn` severity levels; `validate` exit code reflects errors only
  - Built-in validators: `ManifestStructureValidator`, `CircularDependencyValidator`, `FileExistenceValidator`
  - `PrereqValidator` wraps `DoctorService` so prereq checking logic is shared between `validate` and `doctor`
  - `validate` command uses the pipeline; `--check-prereqs` flag enables opt-in prereq warnings
  - 100% coverage on all new validator logic

- **`prepare` command**: manifest modules can declare external packages (`packages:` with brew/cask/apt) and `gz-shellforge prepare` verifies or installs them before validate/build/deploy
  - `--check` reports missing packages and exits non-zero; `--dry-run` prints the plan — neither installs
  - Apply skips already-installed packages and reports an install failure as a non-zero exit without aborting the remaining installs
  - `PackageManager` detection lives in `internal/domain`; the orchestration is `internal/app/prepare_service.go`

### Changed

- **`validate --verbose` output format**: Replaced the numbered step-by-step progress report (1. Parsing… 2. Validating structure…) with a consolidated findings list. Findings now carry severity icons (✗ error, ⚠ warning) and the module name where applicable.
- **Minimum Go version raised to 1.26** (`go.mod`, CI, and `.golangci.yml` `run.go` now agree). Consumers building the module from source need a 1.26 toolchain.

### Fixed

- **`make lint` no longer reports PASS when golangci-lint is absent.** The target fell back to `go vet` or skipped silently, so a missing linter was indistinguishable from a clean run.
- **`go.sum` was missing the `/go.mod` hashes** for `pflag v1.0.10` and `golang.org/x/text v0.37.0`, which left the module unbuildable from a clean module cache.

### Refactored

- **CLI Architecture Improvements**: Complete refactoring for code reusability and maintainability
  - **Phase 1 - ServiceFactory**: Centralized dependency injection for all CLI commands
    - Created `internal/cli/factory/` package with `Services` and `BackupServices` structs
    - Eliminated duplicate service initialization across 10+ CLI files
    - Extracted `GitRepositoryAdapter` to shared factory package
  - **Phase 2 - Output Package**: Standardized CLI output formatting
    - Created `internal/cli/output/` package with `ConfigPrinter` (fluent API) and result helpers
    - Replaced ~90 lines of duplicate `fmt.Printf` patterns with reusable functions
    - Consistent verbose output formatting across backup, restore, cleanup commands
  - **Phase 2 - Error Handling**: Unified error message formatting
    - Created `internal/cli/errors/` package with standardized error helpers
    - Applied consistent error wrapping across all 10 CLI command files
    - Error helpers: `WrapError`, `InvalidPath`, `FileNotFound`, `DirNotFound`, `MinValue`
  - **Phase 3 - Performance Optimization**: Improved string processing efficiency
    - Optimized `splitLines()` function in migrate.go using `strings.Builder` (O(n²) → O(n))
    - Refactored git command execution with helper methods in `repository.go`
    - Reduced boilerplate code in git operations (Add, ConfigUser, GetStatus methods)
  - **Phase 4 - Error Consistency Completion**: Finalized error handling migration
    - Updated validate.go, list.go, template.go to use `clierrors` package
    - Fixed test assertion in list_test.go to match new error format
  - **Impact**: ~150 lines removed, 3 new packages created, improved code maintainability

---

## Past releases

Released versions are archived one file per release line. This file carries only
unreleased changes.

| Line                           | Releases                                                                    |
| ------------------------------ | --------------------------------------------------------------------------- |
| [0.6.x](docs/changelog/0.6.md) | 0.6.0 (2025-12-02)                                                          |
| [0.5.x](docs/changelog/0.5.md) | 0.5.1 (2025-11-30), 0.5.0 (2025-11-28)                                      |
| [0.4.x](docs/changelog/0.4.md) | 0.4.0 (2025-11-27)                                                          |
| [0.3.x](docs/changelog/0.3.md) | 0.3.0 (2025-11-27)                                                          |
| [0.2.x](docs/changelog/0.2.md) | 0.2.1 (2025-11-27), 0.2.0-beta (2025-11-27), 0.2.0-alpha (2025-11-27)       |
| [0.1.x](docs/changelog/0.1.md) | 0.1.0 (2025-11-27)                                                          |

---

## Development

### Project Status
- **Stability**: Alpha (core features stable, API may change)
- **Production Ready**: Yes, for build and validate use cases
- **Test Coverage**: 71-100% across modules

### Contributors
- Initial Go implementation by Claude (Anthropic)
- Based on Python version: gzh-cli-shellforge-py

---

[Unreleased]: https://github.com/gizzahub/gzh-cli-shellforge/compare/v0.6.0...HEAD
