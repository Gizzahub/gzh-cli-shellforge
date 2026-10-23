# Markdown formatter 계약

## 적용 범위

gzh-cli-shellforge는 ISSUE-352의 mdformat 소비자다. 선언, 설치,
`.make/mdformat-constraints.txt`, `MDFORMAT_RUN`, `fmt-md`는
gzh-cli-quality에 채택된 구현을 그대로 가져왔다. Go 도구 pin, 공개 API, 스키마는 바꾸지 않았다.
기존 `make quality`의 Go fmt·lint·test 구성은 유지한다.
기존 Markdown 본문 일괄 수정은 도구·게이트 변경과 별도 커밋으로 진행한다.

## 고정 묶음

- mdformat 1.0.0
- mdformat-gfm 1.0.0
- mdformat-frontmatter 2.1.2
- constraints의 하위 의존성 다섯 개: markdown-it-py 4.2.0,
  mdit-py-plugins 0.6.1, mdurl 0.1.2, ruamel-yaml 0.19.1, wcwidth 0.9.2

gfm이 표를 처리하므로 tables 플러그인은 설치하지 않는다.
constraints 주석은 패밀리 기준을 gzh-cli-gitforge로 둔다.
최상위 pin이 바뀌면 constraints를 다시 해석한다.

## 설치

`make install-mdformat`는 `uv`로 이 저장소의 `bin/tools`에 묶음을 설치한다.
`bin/`은 Git에서 무시된다. `--version` 출력은 변수에 담은 뒤 검사해서,
출력이 나온 뒤에 실패한 명령도 거부한다. 허용 문자열은 `mdformat 1.0.0`과
플러그인 `mdformat-gfm 1.0.0`, `mdformat_frontmatter 2.1.2`이며 플러그인
순서는 어느 쪽이든 된다. 이어서 설치 환경의 Python으로 constraints 각 줄을
`name==version`과 대조한다. `uv`가 없거나 버전, 플러그인, 하위 의존성이
어긋나면 target은 실패한다. `make install-tools`도 이 설치에 의존한다.

## 정렬

`make fmt-md`는 설치와 검증이 끝난 저장소 전용 실행 파일만 사용한다.
전체 writer는 Git이 추적하거나 ignore하지 않은 일반 파일을 고른다.
변경 선택과 읽기 전용 검사 계약은 아래의 채택 내용을 따른다.

## 독립 리뷰

구현 요구 등급은 strong이다. 이 변경은 strong 등급의 독립 리뷰가 따로
필요하다. 이 문서는 리뷰 판정, 설치 결과, 회귀 비교, `make quality` 결과를
기록하지 않는다.

## Python 실행·변경 검사·CI 채택 — 2026-10-23

구현·독립 리뷰 요구 등급은 strong이며 리뷰는 구현자와 분리한다.
설치 전 재사용 검사와 설치 후 검사는 실제 venv의 Python 3.11 이상을 요구한다.
새 설치에서만 unset `UV_PYTHON`에 `>=3.11`을 적용하고 명시된 빈 값도 보존한다.
건강한 최신 venv는 더 낮은 명시 요청이 있어도 재사용한다. 도구 pin은 그대로다.

`format-md-check`는 `MDFORMAT_BASE`(기본 origin/master) merge-base 이후 커밋,
index, 작업트리 및 비추적 일반 Markdown을 수정 없이 검사한다.
Git·formatter 실패를 전파하고 공백·개행·선행 대시 인자를 보존한다.
symlink는 Windows index mode 120000 표현까지 제외하되 실제 일반 파일 타입 변경은
선택한다. staged 내용과 작업 파일이 다르면 index를 수정하지 않고 실패한다.
`fmt-md`/`format-md`는 전체 writer이며 `format-md-all-check`는 명시적 전체 검사다.

`quality`는 기존 Go fmt·lint·test에 변경 Markdown 검사와 installer·selector
회귀 검사를 더한다. `.github/workflows/mdformat.yml`은 Linux·Windows에서 실제
설치·launcher·선택기 smoke를 실행한다. 작업 브랜치는 매번 origin/master,
PR은 대상 SHA, trunk push는 이전 SHA, 수동 실행은 전체 검사를 기준으로 한다.
10분 CI job 제한은 초기 정책값이며 상세 메트릭 설계는 후속 범위다.
Windows 실제 성공 여부는 ISSUE-352 campaign 검증 기록으로 구분하며 POSIX 모의
fixture 결과로 대체하지 않는다. 공개 API·manifest schema는 추가하지 않는다.
