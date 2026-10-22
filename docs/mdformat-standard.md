# Markdown formatter 계약

## 적용 범위

gzh-cli-shellforge는 ISSUE-352의 mdformat 소비자다. 선언, 설치,
`.make/mdformat-constraints.txt`, `MDFORMAT_RUN`, `fmt-md`는
gzh-cli-quality에 채택된 구현을 그대로 가져왔다. Go 도구 pin,
`make quality`(`fmt` + `lint` + `test`), 공개 API, 스키마는 바꾸지 않았다.
새 검사 게이트와 Markdown 본문 일괄 수정은 없다.

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
Git이 추적하거나 ignore하지 않은 `*.md`만 고르고, 일반 파일이 아니거나
symlink이면 제외한다. 경로는 NUL로 구분하고 `./`를 붙여, 이름이 `-`로
시작하는 파일이 옵션으로 읽히지 않게 한다. Bash `pipefail`과 NUL 입력을
받는 `xargs`를 쓰며 Git 조회와 formatter 실패는 Make로 전달된다.
완료 메시지는 그 다음에만 출력한다.

이전 `fmt-md`는 PATH의 mdformat을 `find`로 호출했다. 도구가 없으면 경고 후
성공했고, formatter가 실패해도 성공 메시지를 이어서 출력해 target이 성공할
수 있었다. `quality`, `check`, `all`에는 `fmt-md`를 넣지 않는다.

## 독립 리뷰

구현 요구 등급은 strong이다. 이 변경은 strong 등급의 독립 리뷰가 따로
필요하다. 이 문서는 리뷰 판정, 설치 결과, 회귀 비교, `make quality` 결과를
기록하지 않는다.
