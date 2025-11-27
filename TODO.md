# TODO: Backup/Restore Implementation

## Status: In Progress (Paused)

**Started**: 2025-11-27
**Current Version**: v0.2.0-alpha
**Target Version**: v0.2.0

---

## ✅ Completed Layers

### Domain Layer (Complete)
- [x] `internal/domain/snapshot.go` - Snapshot entity and list operations
- [x] `internal/domain/snapshot_test.go` - 9 comprehensive tests
- [x] Retention policy logic (keep by count/days)
- [x] Always keep at least one snapshot safety rule
- [x] Human-readable size formatting

**Commit**: `44111fb` - feat(domain): add snapshot domain model for backup/restore

### Infrastructure Layer (Partial)
- [x] `internal/infra/git/repository.go` - Git wrapper for version control
- [x] Git installation check
- [x] Repository initialization
- [x] Add/commit operations
- [x] Status checking

**Commit**: `2021ee1` - feat(infra): add Git repository wrapper for backup operations

**Missing**:
- [ ] `internal/infra/git/repository_test.go` - Unit tests for git wrapper
- [ ] Snapshot file operations (copy, list, delete)

---

## ⏳ Remaining Work

### 1. Infrastructure Layer (Filesystem Operations)

**File**: `internal/infra/snapshot/manager.go` (to create)

**Responsibilities**:
- Create snapshot directory structure
- Copy files to snapshot location with timestamp
- List existing snapshots from filesystem
- Delete snapshots based on retention policy
- Verify file integrity

**Functions Needed**:
```go
type SnapshotManager struct {
    fs afero.Fs
    config *domain.BackupConfig
}

func (sm *SnapshotManager) CreateSnapshot(sourcePath string) (*domain.Snapshot, error)
func (sm *SnapshotManager) ListSnapshots(fileName string) (*domain.SnapshotList, error)
func (sm *SnapshotManager) DeleteSnapshot(snapshot *domain.Snapshot) error
func (sm *SnapshotManager) RestoreSnapshot(snapshot *domain.Snapshot, targetPath string) error
func (sm *SnapshotManager) CleanupSnapshots(fileName string, keepCount, keepDays int) ([]domain.Snapshot, error)
```

**Estimated**: 2-3 hours

---

### 2. Application Layer (Business Logic)

**File**: `internal/app/backup_service.go` (to create)

**Responsibilities**:
- Orchestrate backup operations
- Coordinate snapshot manager and git repository
- Handle backup workflow:
  1. Create snapshot
  2. Update current/ directory
  3. Git add/commit
- Handle restore workflow:
  1. Verify snapshot exists
  2. Backup current file
  3. Restore from snapshot
  4. Git commit
- Manage cleanup operations

**Interface**:
```go
type BackupService struct {
    snapshotMgr  *snapshot.SnapshotManager
    gitRepo      *git.Repository
    config       *domain.BackupConfig
}

func (bs *BackupService) Backup(sourcePath string, message string) (*domain.Snapshot, error)
func (bs *BackupService) Restore(snapshot *domain.Snapshot, targetPath string, dryRun bool) error
func (bs *BackupService) ListSnapshots(fileName string) (*domain.SnapshotList, error)
func (bs *BackupService) Cleanup(fileName string, keepCount, keepDays int, dryRun bool) ([]domain.Snapshot, error)
```

**Estimated**: 2-3 hours

---

### 3. CLI Layer (User Interface)

**Files to Create**:
- `internal/cli/backup.go` - Backup command
- `internal/cli/restore.go` - Restore command
- `internal/cli/list_snapshots.go` - List snapshots command (optional)

**backup command**:
```bash
shellforge backup --file ~/.zshrc [--message "description"]

Flags:
  -f, --file string      File to backup (required)
  -m, --message string   Backup description
  --backup-dir string    Backup directory (default: ~/.backup/shellforge)
  --no-git              Skip git versioning
```

**restore command**:
```bash
shellforge restore --file ~/.zshrc --snapshot 2025-11-27_14-30-45

Flags:
  -f, --file string        File to restore to (required)
  -s, --snapshot string    Snapshot timestamp (required)
  --backup-dir string      Backup directory (default: ~/.backup/shellforge)
  --dry-run               Preview restore without executing
```

**cleanup command** (optional):
```bash
shellforge cleanup --file ~/.zshrc --keep-count 10 --keep-days 30

Flags:
  -f, --file string      File pattern to cleanup
  --keep-count int       Number of snapshots to keep
  --keep-days int        Days of snapshots to keep
  --dry-run             Preview deletions
```

**Estimated**: 3-4 hours

---

### 4. Testing

**Unit Tests Needed**:
- [ ] `internal/infra/git/repository_test.go` - Git wrapper tests
- [ ] `internal/infra/snapshot/manager_test.go` - Snapshot manager tests
- [ ] `internal/app/backup_service_test.go` - Backup service tests
- [ ] `internal/cli/backup_test.go` - Backup CLI tests
- [ ] `internal/cli/restore_test.go` - Restore CLI tests

**Integration Tests Needed**:
- [ ] End-to-end backup workflow
- [ ] End-to-end restore workflow
- [ ] Cleanup with retention policy
- [ ] Git integration (commit history)

**Estimated**: 2-3 hours

---

### 5. Documentation

**Files to Update**:
- [ ] README.md - Add backup/restore commands section
- [ ] CHANGELOG.md - Update v0.2.0 with backup features
- [ ] Update examples/ with backup usage

**Estimated**: 30 minutes

---

## Total Remaining Effort

**Estimated Time**: 10-15 hours total
- Infrastructure: 2-3 hours
- Application: 2-3 hours
- CLI: 3-4 hours
- Testing: 2-3 hours
- Documentation: 30 minutes
- Integration/debugging: 2 hours buffer

---

## Implementation Notes

### Directory Structure
```
~/.backup/shellforge/
├── .git/                    # Git repository
├── current/                 # Current deployed files
│   └── zshrc
└── snapshots/              # Timestamped snapshots
    └── zshrc/
        ├── 2025-11-27_14-30-45
        ├── 2025-11-26_10-15-30
        └── 2025-11-25_16-45-00
```

### Design Decisions

1. **Afero Filesystem Abstraction**: Use afero.Fs throughout for testability
2. **Git Optional**: System should work without git (filesystem-only backup)
3. **Safety First**: Always keep at least one snapshot
4. **Timestamps**: Use YYYY-MM-DD_HH-MM-SS format for consistency
5. **Dry-Run Support**: All destructive operations support --dry-run

### Requirements Reference

- **FR-007**: Git Operations (partial - git wrapper done)
- **FR-008**: Snapshot Management (domain model done)

---

## Next Session Checklist

1. [ ] Review this TODO
2. [ ] Implement `internal/infra/snapshot/manager.go`
3. [ ] Add tests for snapshot manager
4. [ ] Implement `internal/app/backup_service.go`
5. [ ] Add tests for backup service
6. [ ] Implement CLI commands
7. [ ] Integration testing
8. [ ] Update documentation
9. [ ] Commit and tag v0.2.0

---

## Alternative: Minimal MVP

If time-constrained, implement minimal backup-only feature:

1. Skip restore command (future release)
2. Skip cleanup command (future release)
3. Implement basic backup only:
   - Copy file to timestamped snapshot
   - Optional git commit
   - No listing or management

This would provide core value with ~5-6 hours work instead of 10-15.

---

**Last Updated**: 2025-11-27
**Status**: Domain + Infrastructure (git) complete, Application/CLI layers pending
