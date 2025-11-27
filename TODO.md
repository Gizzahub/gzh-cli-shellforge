# TODO: Backup/Restore Implementation

## Status: Backup Complete ✅ | Restore/Cleanup Planned

**Started**: 2025-11-27
**Current Version**: v0.2.0-alpha
**Target Version**: v0.2.1 (restore/cleanup)

---

## ✅ Completed: v0.2.0-alpha Backup Feature

### Domain Layer (Complete)
- [x] `internal/domain/snapshot.go` - Snapshot entity and list operations
- [x] `internal/domain/snapshot_test.go` - 9 comprehensive tests
- [x] Retention policy logic (keep by count/days)
- [x] Always keep at least one snapshot safety rule
- [x] Human-readable size formatting

**Commit**: `44111fb` - feat(domain): add snapshot domain model for backup/restore

### Infrastructure Layer (Complete)
- [x] `internal/infra/git/repository.go` - Git wrapper for version control
- [x] `internal/infra/git/repository_test.go` - Comprehensive git tests (81.2% coverage)
- [x] `internal/infra/snapshot/manager.go` - Snapshot file operations
- [x] `internal/infra/snapshot/manager_test.go` - 26 subtests (77.6% coverage)
- [x] Git installation check, init, add/commit, status operations
- [x] Snapshot create, list, delete, restore, cleanup operations

**Commits**:
- `2021ee1` - feat(infra): add Git repository wrapper for backup operations
- `9b56de9` - test(infra): add comprehensive git repository tests
- `52688f9` - feat(infra): implement snapshot manager for backup/restore

### Application Layer (Complete)
- [x] `internal/app/backup_service.go` - Business logic orchestration
- [x] `internal/app/backup_service_test.go` - Mock-based tests (78.0% coverage)
- [x] Backup workflow (snapshot + git commit)
- [x] Restore workflow (with dry-run support)
- [x] Cleanup workflow (retention policies)
- [x] Git integration (optional, non-fatal)

**Commit**: `5ff61e7` - feat(app): implement backup service with orchestration logic

### CLI Layer (Complete)
- [x] `internal/cli/backup.go` - Backup command implementation
- [x] Git repository adapter pattern
- [x] Home path expansion (~) support
- [x] Verbose mode with detailed output
- [x] Integration with root command

**Commit**: `1bd441d` - feat(cli): add backup command for shell configuration files

### Documentation (Complete)
- [x] README.md - Backup command documentation
- [x] Updated features section (backup moved to implemented)
- [x] Updated status section (v0.2.0-alpha achievements)

**Commit**: `99b4896` - docs(readme): document backup command implementation

---

## ⏳ Planned: v0.2.1 Restore/Cleanup Features

### 1. CLI Layer - Restore Command

**File**: `internal/cli/restore.go` (to create)

**restore command**:
```bash
gz-shellforge restore --file ~/.zshrc --snapshot 2025-11-27_14-30-45

Flags:
  -f, --file string        File to restore to (required)
  -s, --snapshot string    Snapshot timestamp (required)
  --backup-dir string      Backup directory (default: ~/.backup/shellforge)
  --dry-run               Preview restore without executing
  -v, --verbose           Show detailed output
```

**Implementation Notes**:
- Use existing `BackupService.Restore()` method
- Add snapshot listing/selection if timestamp not provided
- Create safety backup before restore
- Git commit after successful restore

**Estimated**: 2 hours

---

### 2. CLI Layer - Cleanup Command

**File**: `internal/cli/cleanup.go` (to create)

**cleanup command**:
```bash
gz-shellforge cleanup --file ~/.zshrc --keep-count 10 --keep-days 30

Flags:
  -f, --file string      File pattern to cleanup (required)
  --keep-count int       Number of snapshots to keep (default: 10)
  --keep-days int        Days of snapshots to keep (default: 30)
  --backup-dir string    Backup directory (default: ~/.backup/shellforge)
  --dry-run             Preview deletions without executing
  -v, --verbose         Show detailed output
```

**Implementation Notes**:
- Use existing `BackupService.Cleanup()` method
- Show what will be deleted before confirmation
- Git commit cleanup actions
- Safety: always keep at least one snapshot

**Estimated**: 2 hours

---

### 3. CLI Layer - List Snapshots Command (Optional)

**Enhancement to existing list command or new subcommand**:
```bash
gz-shellforge snapshots --file ~/.zshrc

# Or as subcommand
gz-shellforge backup list --file ~/.zshrc
```

**Estimated**: 1 hour

---

### 4. Testing

**Unit Tests Needed**:
- [ ] `internal/cli/restore_test.go` - Restore CLI tests
- [ ] `internal/cli/cleanup_test.go` - Cleanup CLI tests

**Integration Tests**:
- [ ] End-to-end restore workflow
- [ ] End-to-end cleanup workflow

**Estimated**: 2 hours

---

### 5. Documentation

**Files to Update**:
- [ ] README.md - Add restore/cleanup commands
- [ ] Update examples/ with restore/cleanup usage

**Estimated**: 30 minutes

---

## Total v0.2.1 Effort

**Estimated Time**: 7-8 hours
- Restore CLI: 2 hours
- Cleanup CLI: 2 hours
- List snapshots: 1 hour
- Testing: 2 hours
- Documentation: 30 minutes
- Integration/debugging: 30 minutes buffer

---

## Design Decisions (Implemented)

1. ✅ **Afero Filesystem Abstraction**: Used throughout for testability
2. ✅ **Git Optional**: Works without git (filesystem-only backup)
3. ✅ **Safety First**: Always keep at least one snapshot
4. ✅ **Timestamps**: YYYY-MM-DD_HH-MM-SS format
5. ✅ **Dry-Run Support**: Implemented in BackupService
6. ✅ **Adapter Pattern**: CLI layer uses adapters for infrastructure dependencies
7. ✅ **Non-Fatal Git**: Git failures don't break backup operations

---

## Architecture Summary

**4-Layer Clean Architecture**:
```
CLI Layer (internal/cli/)
  ↓ uses
Application Layer (internal/app/)
  ↓ uses
Infrastructure Layer (internal/infra/)
  ↓ depends on
Domain Layer (internal/domain/)
```

**Test Coverage**:
- Domain: 76.9%
- Infrastructure: 77.6-91.7%
- Application: 78.0-89.2%
- CLI: 71.3%
- **Overall**: 50/50 tests passing

---

**Last Updated**: 2025-11-27
**Status**: v0.2.0-alpha backup feature complete (100%). Restore/cleanup planned for v0.2.1.
