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

| 제공하는 것 (Is)                            | 되지 않을 것 (Is Not)              |
| ------------------------------------------- | ---------------------------------- |
| 다중 쉘 설정의 모듈형 빌드/배포             | 쉘 자체·쉘 런타임 대체             |
| 위상 정렬 의존성 해결·OS 필터링             | 모든 dotfile 대상 범용 동기화 도구 |
| build ↔ deploy 분리·dry-run·git 스냅샷 롤백 | 원격 설정 저장소 호스팅            |
| 쉘 설정에 특화된 마이그레이션·검증          | GUI·웹·IDE 플러그인                |

______________________________________________________________________

## Goals (Measurable Targets)

G1. **Multi-shell / multi-file build**

- Target: zsh·bash·fish 및 zshrc/zprofile/zshenv/config.fish/conf.d 등을
  단일 manifest에서 빌드

G2. **Build performance**

- Target: 기동은 matched empty-Go control 대비 중앙값 차이 \<= 3ms,
  10모듈 빌드 < 50ms, 메모리 < 10MB, 바이너리 ~8MB
- Metric: macOS 에서 릴리스 바이너리에
  `make perf-startup-qualify STARTUP_ARGS='<검증한 바이너리·provenance·witness·output 인자>'`
  로 잰다. 정확한 인자는 [측정 프로토콜](docs/dev/60-benchmarks.md)의 현재 판정 절을 따른다.
  스크립트는 `com.apple.quarantine` 속성이 붙은 바이너리를
  거부하므로 릴리스 산출물은 `gh release download` 나 `curl` 로 받는다
  (`make perf-startup`과 기본 script 실행은 개발용 관측이며 판정하지 않는다).
  각 세션은 version/control warm 30쌍을 AB/BA 교차하며 세션 시작 순서를 반전한다.
  `delta_s = median(version 30개) - median(control 30개)`를 계산하고 처음 유효한
  3세션의 `median(delta_1, delta_2, delta_3) <= 3ms`로 판정한다. paired-difference
  median은 별도 참고값이며 각 세션 모두 \<= 3ms라는 규칙과도 구분한다.
  시도 종료부터 다음 시작까지 최소 60초, 최대 5시도다. 유효 3세션 부족은 판정 불가다.
  세션 시작/종료·워크로드 경계·warm pair 경계에서 load1/5/15와 논리 CPU 수를 기록하며
  `load1/CPU > 1.0` 또는 소유 heavy 작업 겹침은 해당 시도를 무효로 남긴다.
  소유 heavy 상태는 유효기간·owner·execution ID가 있는 parent 관리 witness로 확인한다.
  누락·만료·불일치도 무효이며 전역 프로세스 탐지를 보증하지 않는다.
  모든 원문·무효 사유·부분 실패를 보존하고 좋은 결과를 골라 재시도하지 않는다.
  공개 release/control/harness provenance 변경·검증 불가나 소유 process 정리 실패는
  전체 batch를 판정 불가로 중단한다. interruption·setup 실패도 부분 결과와 함께 남긴다.
  10모듈 빌드는 각 유효 세션의 warm median < 50ms, 메모리는 유효 세션 전체 version
  memory 표본(각 5개)의 peak_footprint_bytes 최댓값 < 10,000,000 bytes로 판정한다.
  first-observed·새 inode 첫 실행(각 5개)·maximum RSS는 참고값이다.

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

- 기동 control 대비 \<= 3ms; 10모듈 빌드 < 50ms — G2 Metric 정의로 판정한다

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
