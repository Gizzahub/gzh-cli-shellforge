# Product Goals (No-PRD)

**Project**: gzh-cli-shellforge (`gz-shellforge` binary)
**Doc Type**: Goals + Constraints + Quality Gates
**Status**: Active
**Last Updated**: 2026-07-16

______________________________________________________________________

## Product Intent

gzh-cli-shellforge **builds and deploys modular shell configurations** across
zsh / bash / fish from a single manifest. It:

- resolves module load order by topological sort and filters modules by OS,
- separates build from deploy (dry-run + git-backed backup/rollback),
- and wraps shell-config management rather than replacing the shell (SOUL 신념 1).

This is a feature-library project — a single PRODUCT.md is sufficient. It
replaces the earlier marketing-style document; its product positioning now lives
in [README.md](README.md).

| 제공하는 것 (Is)                              | 되지 않을 것 (Is Not)                       |
| --------------------------------------------- | ------------------------------------------- |
| 다중 쉘 설정의 모듈형 빌드/배포               | 쉘 자체·쉘 런타임 대체                      |
| 위상 정렬 의존성 해결·OS 필터링               | 모든 dotfile 대상 범용 동기화 도구          |
| build ↔ deploy 분리·dry-run·git 스냅샷 롤백  | 원격 설정 저장소 호스팅                     |
| 쉘 설정에 특화된 마이그레이션·검증           | GUI·웹·IDE 플러그인                         |

______________________________________________________________________

## Goals (Measurable Targets)

G1. **Multi-shell / multi-file build**

- Target: zsh·bash·fish 및 zshrc/zprofile/zshenv/config.fish/conf.d 등을
  단일 manifest에서 빌드

G2. **Build performance**

- Target: 기동 < 10ms, 10모듈 빌드 < 50ms, 메모리 < 10MB, 바이너리 ~8MB

G3. **Safe deploy**

- Target: `deploy`는 `--backup`/`--dry-run` 지원; 원본 dotfile은 배포 전 백업 가능

G4. **Migration speed**

- Target: 모놀리식 `.zshrc` → 모듈 manifest 자동 생성 5분 내

G5. **Test reliability**

- Target: 커버리지 >= 80% (현재 78%, 291 테스트 → 목표 상향)

______________________________________________________________________

## Non-Goals (Explicitly Out of Scope)

- No 쉘 자체·쉘 런타임 대체 (기존 쉘을 감쌀 뿐)
- No 범용 dotfile 매니저 — shellforge는 쉘 설정에 특화된다
- No 원격 설정 저장소 호스팅
- No GUI·웹·IDE 플러그인
- No 시스템 전역 설정(`/etc/*`) 자동 변경 — 감지만 하고 명시 승인 전 미변경

______________________________________________________________________

## Guardrails and Technical Constraints

**Architecture**

- Hexagonal: `domain`/`app`/`infra`/`cli` 관심사 분리
- 모듈 로드 순서는 위상 정렬(Kahn)로 결정; 순환 의존성은 감지·경고

**Dependency Boundaries**

- `gzh-cli-core`만 의존 가능; 다른 feature 라이브러리 의존 금지 (GUIDELINES §2)

**Compatibility**

- Go 1.25+ (`go.mod` go 1.25.7; devbox 툴체인 1.26)
- macOS 10.15+/Linux/WSL2/BSD; 백업·복원은 Git 필요

**Safety**

- 파괴적 배포는 `--backup`/`--dry-run`을 요구; 원본 dotfile 보존이 기본
- 쉘 입력을 sanitize한다 (shell injection 방지)

______________________________________________________________________

## Quality Gates (Release Readiness)

**Build and Lint**

- `make quality` (fmt + lint + test) pass with no warnings

**Testing**

- `make test-coverage` pass; 커버리지 >= 80%

**Performance**

- 기동 < 10ms; 10모듈 빌드 < 50ms

**Docs**

- 명령 레퍼런스가 실제 명령·플래그와 일치

______________________________________________________________________

## Decision Rules

- 새 기능은 SOUL.md 4-게이트(틈 · 라이브러리 · 대량/전환 · 날카로움)를 통과해야 한다
- 쉘 도구 재구현은 게이트 1(재발명 금지)에서 거절된다 — 감싸되 대체하지 않는다
- 최소 하나의 goal에 매핑되거나 명시적으로 승인되어야 한다
- Guardrails 위반은 문서화된 예외를 요구한다
- Quality Gates 미충족 시 릴리스는 차단된다

______________________________________________________________________

**End of Document**
