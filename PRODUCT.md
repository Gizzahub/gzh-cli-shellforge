# Shellforge 제품 소개

> 다중 쉘 설정 파일을 모듈형 구조로 통합 관리하는 빌드/배포 도구

---

## 핵심 가치

**"모든 쉘 설정 파일을 하나의 설정으로 통합 관리합니다"**

### 관리 대상 파일

- `~/.zshrc`, `~/.bashrc` — 인터랙티브 쉘 설정
- `~/.zprofile`, `~/.bash_profile` — 로그인 쉘 설정
- `~/.profile` — 공통 프로파일
- `~/.config/fish/config.fish` — Fish 쉘 설정
- `~/.config/fish/conf.d/*.fish` — Fish 모듈형 설정 (자동 로드)
- `/etc/profile`, `/etc/zshrc` — 시스템 전역 설정 (예정)

### 해결하는 문제

| 기존 문제 | Shellforge 해결책 |
|----------|------------------|
| 여러 설정 파일 개별 관리 | **단일 manifest로 통합 관리** |
| `.zshrc`가 수백 줄로 비대해짐 | 모듈별 분리로 관리 용이 |
| 스크립트 순서 의존성 수동 관리 | 의존성 자동 해결 (위상 정렬) |
| Mac/Linux 설정 혼재 | OS 자동 감지 및 필터링 |
| 런타임까지 오류 발견 불가 | 빌드 전 검증으로 조기 발견 |
| 설정 변경 후 롤백 어려움 | Git 기반 스냅샷으로 즉시 복구 |

---

## 주요 기능

### 1. 통합 빌드 (Build)

```bash
# 모든 설정 파일을 빌드 디렉토리에 생성 (OS 자동 감지)
gz-shellforge build

# 빌드 출력 디렉토리 지정
gz-shellforge build --output-dir ./staging

# 특정 타겟만 빌드
gz-shellforge build --target zshrc --target zprofile

# 다른 쉘용으로 빌드
gz-shellforge build --shell fish
```

- **모든 대상 파일을 한번에 빌드** (zshrc, profile, zprofile 등)
- OS 미지정 시 **현재 OS 자동 감지**
- **위상 정렬 알고리즘**으로 모듈 로드 순서 자동 결정
- 순환 의존성 감지 및 경고

### 2. 배포 (Deploy)

```bash
# 빌드 결과물을 각 파일의 실제 경로에 배포
gz-shellforge deploy

# 미리보기 (실제 배포 없이 확인)
gz-shellforge deploy --dry-run

# 백업 후 배포
gz-shellforge deploy --backup
```

- 빌드된 파일을 **각각의 원래 경로에 자동 배포**
- 배포 전 자동 백업 옵션
- dry-run으로 안전하게 미리보기

### 3. OS 자동 감지 및 필터링

```yaml
modules:
  - name: brew-path
    file: init.d/05-brew-path.sh
    requires: [os-detection]
    os: [Mac]  # macOS에서만 로드
```

- **--os 옵션 생략 시 현재 OS 자동 감지**
- Mac, Linux 환경별 모듈 자동 필터링
- 하나의 저장소로 다중 환경 관리
- 불필요한 조건문 제거

### 3.1 Fish 쉘 지원

```bash
# Fish 쉘용 빌드
gz-shellforge build --shell fish

# Fish conf.d 모듈형 설정 (모듈별 개별 파일 생성)
gz-shellforge build --shell fish --target conf.d
```

**지원 타겟**:
- `config` → `~/.config/fish/config.fish` (단일 파일)
- `conf.d` → `~/.config/fish/conf.d/*.fish` (모듈별 파일)

**XDG_CONFIG_HOME 지원**:
- `$XDG_CONFIG_HOME` 환경변수 자동 인식
- 기본값: `~/.config`
- 사용자 지정 경로 지원 (예: `~/.myconfig/fish/`)

### 4. 마이그레이션 도구 (Migrate)

```bash
gz-shellforge migrate ~/.zshrc
```

- 기존 모놀리식 설정 자동 분석
- 섹션 감지 및 모듈 분리
- 의존성 추론 및 manifest.yaml 생성
- **5분 내 마이그레이션 완료**

### 5. 템플릿 시스템 (Template)

```bash
gz-shellforge template generate alias my-aliases
```

6종 내장 템플릿:
- PATH 설정
- 환경 변수
- 별칭(alias)
- 도구 초기화 (nvm, rbenv 등)
- 함수 정의
- 커스텀

### 6. 백업/복원 (Backup/Restore)

```bash
# 백업 생성
gz-shellforge backup --file ~/.zshrc --message "before update"

# 스냅샷 복원
gz-shellforge restore --file ~/.zshrc --snapshot 2025-01-19_14-30-45
```

- Git 기반 버전 관리
- 타임스탬프 스냅샷
- 즉각적인 롤백

### 7. 설정 비교 (Diff)

```bash
gz-shellforge diff ~/.zshrc ~/.zshrc.new --format unified
```

4가지 출력 포맷:
- `summary`: 변경 요약
- `unified`: Git 스타일 unified diff
- `context`: 컨텍스트 diff
- `side-by-side`: 양쪽 비교

### 8. 검증 (Validate)

```bash
gz-shellforge validate --verbose
```

- manifest.yaml 구문 검사
- 모듈 파일 존재 확인
- 의존성 그래프 유효성 검사
- 순환 참조 감지

---

## 제품 특징

### 성능

| 지표 | Python 버전 | Go 버전 | 개선 |
|------|------------|---------|------|
| 시작 시간 | ~200ms | <10ms | **20배** |
| 빌드 (10모듈) | ~300ms | <50ms | **6배** |
| 메모리 | ~80MB | <10MB | **8배** |
| 바이너리 크기 | ~40MB | ~8MB | **5배** |

### 모듈 구조

```
modules/
├── init.d/       # 초기화 (PATH, OS 감지)
├── rc_pre.d/     # 도구 설정 (nvm, rbenv)
└── rc_post.d/    # 별칭, 함수
```

### manifest.yaml 형식

```yaml
version: "2"

shell:
  type: zsh  # zsh, bash, fish

output:
  directory: ~
  backup: true

modules:
  # zshenv - 모든 쉘에서 로드 (환경변수)
  - name: os-detection
    file: init.d/00-os-detection.sh
    target: zshenv       # 대상 RC 파일 (zshrc, zprofile, zshenv 등)
    priority: 10         # 우선순위 (0-100, 낮을수록 먼저)
    os: [Mac, Linux]
    description: Detect operating system

  # zprofile - 로그인 쉘에서만 로드 (PATH 설정)
  - name: brew-path
    file: init.d/05-brew-path.sh
    target: zprofile
    priority: 10
    requires: [os-detection]
    os: [Mac]
    description: Initialize Homebrew PATH

  # zshrc - 인터랙티브 쉘 설정 (별칭, 함수)
  - name: aliases
    file: rc_post.d/aliases.sh
    target: zshrc
    priority: 80
    os: [Mac, Linux]
    description: Common aliases and functions
```

---

## 사용 사례

### 1. 개인 개발자

- 여러 머신 간 쉘 설정 동기화
- Mac과 Linux 환경 통합 관리
- 설정 변경 이력 추적

### 2. 팀/조직

- 표준 개발 환경 설정 공유
- 신규 입사자 온보딩 간소화
- 버전 관리를 통한 협업

### 3. DevOps

- CI/CD 환경 쉘 설정 자동화
- 컨테이너 이미지 쉘 설정 관리
- 인프라 코드로서의 설정

---

## 지원 환경

### 운영체제

- macOS 10.15+
- Linux (Ubuntu, Debian, Fedora, Arch)
- WSL2

### 쉘

- Zsh
- Bash
- Fish (XDG_CONFIG_HOME 지원, conf.d 모듈형 설정)

### 요구사항

- Go 1.21+ (빌드 시)
- Git (백업/복원 기능)

---

## 제품 현황

**버전**: 0.5.1
**테스트**: 291개 통과 (커버리지 78%)
**상태**: 활발한 개발 중

### 구현 완료

- ✅ build, deploy, validate, list 명령어
- ✅ 다중 설정 파일 통합 관리 (zshrc, profile, zprofile 등)
- ✅ OS 자동 감지
- ✅ 모놀리식 설정 마이그레이션
- ✅ 6종 템플릿 생성
- ✅ Git 기반 백업/복원
- ✅ 4가지 형식 diff 비교
- ✅ Fish 쉘 지원 (config, conf.d)
- ✅ XDG_CONFIG_HOME 환경변수 지원

### 개발 예정

- ⏳ 커스텀 검증기 플러그인 시스템
- ⏳ BSD 지원 (FreeBSD 13+)
- ⏳ 시스템 전역 설정 (/etc/profile, /etc/zshrc)

---

## 빠른 시작

```bash
# 1. 설치
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest

# 2. 기존 설정 백업
gz-shellforge backup --file ~/.zshrc

# 3. 모듈화 마이그레이션
mkdir ~/shellforge && cd ~/shellforge
gz-shellforge migrate ~/.zshrc

# 4. 모든 설정 파일 빌드 (OS 자동 감지)
gz-shellforge build

# 5. 빌드 결과 확인 후 배포
gz-shellforge deploy --dry-run  # 미리보기
gz-shellforge deploy --backup   # 백업 후 배포
```

---

## 경쟁 제품 대비 장점

| 기능 | Shellforge | dotbot | chezmoi | stow |
|------|-----------|--------|---------|------|
| **다중 설정 파일 통합 관리** | ✅ | ❌ | △ | ❌ |
| **빌드/배포 분리** | ✅ | ❌ | ❌ | ❌ |
| 의존성 자동 해결 | ✅ | ❌ | ❌ | ❌ |
| OS 자동 감지 | ✅ | ❌ | ✅ | ❌ |
| 쉘 설정 특화 | ✅ | ❌ | ❌ | ❌ |
| 마이그레이션 도구 | ✅ | ❌ | △ | ❌ |
| 템플릿 시스템 | ✅ | △ | ✅ | ❌ |
| 단일 바이너리 | ✅ | ❌ | ✅ | ✅ |

---

## 참고 자료

- [빠른 시작 가이드](docs/user/00-quick-start.md)
- [명령어 레퍼런스](docs/user/40-command-reference.md)
- [워크플로우 가이드](docs/user/30-workflows.md)
- [예제](examples/)

---

**"복잡한 쉘 설정을 단순하게"** — Shellforge
