# Performance Benchmarks

This document records CLI startup and memory observations alongside diff comparison benchmarks.

## 현재 G2 판정 — control 상대값과 안정화된 세션

2026-10-26 사용자가 채택한 프로토콜이다. 기존 절대 warm 10ms 기준의 관측·미달은 아래에
보존하며 새 estimator로 소급 PASS 처리하지 않는다. 생산 Go 코드 최적화나 다른 플랫폼의
성능 보증을 포함하지 않는다. `make perf-startup`과 기본 script 실행은 기존 개발 관측이다.

정식 판정은 `make perf-startup-qualify`가 호출하는 script의 `--qualify` 모드다.
operator는 익명 공개 release archive/checksums를 먼저 검증하고 추출 binary의 SHA256,
정확한 release/source를 입력한다. script는 hash, stable module version, vcs.revision,
vcs.modified=false, macOS/architecture, CGO=0, trimpath와 현재 Go compiler 일치를 요구한다.
control은 같은 compiler/target으로 `-trimpath -ldflags '-s -w'`, CGO=0 빌드하고 실제
buildinfo를 다시 대조한다. strip flag의 배포 provenance는 검증한 공개 artifact SHA와
소유 release 설정으로 묶는다. compiler 불일치는 판정 불가이며 이전 Go 강제 설치나
전역 PATH 변경으로 우회하지 않는다. quarantine 속성을 제거하지도 않는다.

필수 script 인자는 `--qualify`, `--binary`, `--expected-binary-sha256`,
`--expected-source-sha`, `--release-version`, `--own-heavy-state-file`,
`--own-heavy-owner`, `--output`이다. `--preflight-only`를 추가하면 metadata만 확인하고
control build·제품 실행·witness 검증은 하지 않는다. 입력 hash는 예시나 설치본 추정값이
아니라 실제 공개 다운로드 검증 결과를 사용한다.

parent가 관리하는 witness JSON은 다음 필드를 요구한다: `schema_version: 1`,
`owner`(입력 owner와 일치), 비어 있지 않은 `execution_id`, `valid_until_epoch`
(유효한 Unix seconds), `active_heavy_jobs`(소유 heavy job ID 배열).
측정 자체는 배열에서 제외한다. parent는 covered execution 동안 다른 소유 heavy job을
시작하기 전에 witness를 갱신해야 한다. 각 checkpoint에서 원문을 읽어 보존하고 누락,
잘못된 형식, 만료, owner/execution 변경 또는 active job이 있으면 시도를 무효로 중단한다.
이는 parent 소유 작업에 관한 증언이며 미등록 작업·다른 세션·VM의 전역 탐지가 아니다.

한 batch는 최대 5시도이며 처음 유효한 3세션을 사용한다. 시도 종료부터 다음 시도 시작까지
최소 60초다. 간격은 재현 가능한 분리 규칙일 뿐 cache 초기화나 부하 해소 보증이 아니다.
각 세션의 version/control warm 30쌍은 AB/BA 교대하고 다음 세션의 시작 순서를 반전한다.
세션 시작/종료, workload 경계, warm pair 경계의 load1/5/15와 logical CPU를 보존한다.
어느 checkpoint라도 load1/CPU >1.0이면 그 시도를 무효로 중단한다. 정확히 1.0은 허용한다.
명령 실패·필수 metric 누락도 무효이며 부분 stdout/stderr와 얻은 표본을 즉시 저장한다.
공개 binary의 기대 hash와 control build 직후 hash, 시작 시 harness hash를 모든 identity
checkpoint에서 다시 확인한다. provenance 변경·검증 불가 또는 process cleanup 실패는
해당 시도뿐 아니라 전체 batch를 INCONCLUSIVE로 중단한다. 원래 bytes로 복구한 뒤
다음 세션을 통과시켜 같은 batch를 살리지 않는다. setup·간격 대기·측정 중 interruption도
최종 failure metadata와 기존 원문을 보존한다. 실패 시도도 5시도에 포함한다. 좋은 결과를 선택하거나 budget 초과를
무효 사유로 삼지 않는다. 유효 세션이 3개보다 적으면 INCONCLUSIVE다.

각 세션의 `delta_s = median(version warm) - median(control warm)`를 구한 뒤
`median(delta_1, delta_2, delta_3) <= 3ms`로 기동을 판정한다. 이는 paired difference의
median이나 세 delta 모두 \<=3ms 조건과 다르다. 각 delta와 paired median을 별도로
공개한다. 각 유효 세션의 10모듈 build warm median은 \<50ms, 유효 세션 전체 version
memory 표본(각 5개)의 peak footprint 최댓값은 \<10,000,000 bytes여야 한다.
first-observed·새 inode 5표본·RSS·inittrace는 참고 관측이며 기동 판정으로 바꾸지 않는다.
schema2는 raw pair 순서/값, 무효 checkpoint/이유, partial failure, input hash와 집계를
모두 보존한다. write는 `.partial` 뒤 atomic replace를 사용한다.
subprocess는 각자의 새 process session/group에서 실행한다. timeout·interruption·종료 뒤
소유 group만 TERM/KILL로 제한 시간 안에 정리하고 direct child를 reap한다. group 확인은
해당 PGID의 상태만 조회하며 EPERM을 종료 증거로 간주하지 않는다. live/unknown 또는
조회 실패는 fail-closed다. 다른 process/group을 중지하지 않으며 global process dump도 남기지 않는다.
기동의 scored elapsed는 Popen 직전부터 communicate 완료(자식 종료·출력 회수)까지다.
이후 group/poll/reap 검증 비용은 별도 `cleanup_elapsed_ms_unscored`로 보존하며 기동에
더하지 않는다. cleanup이 검증되기 전 다음 호출로 진행하지 않는다. malformed ps 행이나
비정상 witness JSON의 NaN/Infinity는 unknown/invalid로 남기고 실패 metadata를 직렬화한다.

`make test-startup-protocol`은 가짜 clock/load/측정값으로 순서·경계·무효 처리·첫 유효
선택·estimator·부분 보존을 검사하고 `make quality`에도 포함된다. 실제 측정은 독립 소스
리뷰와 소유 CPU 조정 뒤 한 batch로 수행하며, 측정 후 관련 evidence/docs를 동결하고
최종 소유 quality/coverage/build gate를 통과해야 인도 검증이 완료된다.

## 채택 프로토콜의 단일 batch 관측 — 2026-10-26 KST

실제 실행은 2026-10-25 16:12:13 UTC에 시작한 기록이며 재측정하지 않았다.
Apple M5 Max, macOS 27.0.1, logical CPU 18개, Python 3.14.5, Go 1.27.3이다.
공개 v0.6.0 darwin/arm64 archive와 공개 checksums를 익명 다운로드로 대조했다.
archive SHA256 `3a15e20e043764c2c26672e5a5cfe22539f95934930b09e9ea2adc7769b8bf85`,
binary SHA256 `8b0954ef84349a368455541b97a1cafbe2aa41214f7be11fa0ca21cdb88b20b1`,
source `89baf366d806f694d4e6530bbf13a30f760b021c`이며 binary bytes는 과거 관측과 같다.
control도 현재 Go 1.27.3, darwin/arm64, CGO=0, trimpath, `-s -w`로 빌드했다.
strip flag provenance는 검증한 공개 bytes·해당 release source 설정으로 묶은 것이며
buildinfo가 strip flag 자체를 증명한다고 주장하지 않는다.

측정 시 repository HEAD는 `f07733851968da261d082865070939111359c545`였지만 새 harness는
그 HEAD의 코드가 아니라 독립 리뷰를 마친 작업 트리 입력이었다. tracked diff SHA256은
`a244e34920ab1ddbd9e4b499290b15c0f6a75336580d8ba417bd0a3772798e92`다.
이 측정 입력을 보존한 controller commit은 `88c7421197f242e423a1ef4e9695ce47fc18933e`다.
아래 SHA256은 해당 측정 당시의 harness를 가리키며 이후 launcher와 구분한다.

| 파일                          | 측정 입력 SHA256                                                   |
| ----------------------------- | ------------------------------------------------------------------ |
| `scripts/startup-profile.py`  | `6d5a7a2953939205870e0afdaadc02f62ad95f130ced0c0428a9547f9b7e411b` |
| `scripts/startup_protocol.py` | `39bb85806f5ba58dcf343518915081b6f86b4eb28fc6b7cfc8fe9f10afa35376` |
| `scripts/startup_process.py`  | `667276db30bc324895123fc65d3478a9cc372d131ef332399f1326f2d63adb81` |

측정 후 macOS Intel preflight의 host architecture `x86_64`를 Go의 `amd64`로 정규화하는
수리를 추가했다. 기존 `aarch64`/`arm64` 경로와 timing·estimator·collection은 유지한다.
수정된 launcher는 재측정하지 않았으며 아래 ARM64 관측을 새 launcher의 측정으로
재표기하지 않는다. 원문 SHA256 `4ded8b3f…adba3`와 위 측정 harness 해시는 그대로다.

한 batch에서 처음 세 시도가 모두 유효했다. 각 세션 warm 30쌍, memory 5개·fresh-copy
5개를 그대로 수집했고 무효 시도나 추가 시도는 없었다. 결과를 골라 재실행하지 않았다.

| 세션 | version median ms | control median ms | delta ms  | paired median 참고 ms | build median ms | version 최대 footprint bytes |
| ---- | ----------------- | ----------------- | --------- | --------------------- | --------------- | ---------------------------- |
| 1    | 5.445500          | 3.481146          | 1.964354  | 2.009771              | 5.5790415       | 5,390,984                    |
| 2    | 5.414063          | 3.361542          | 2.052521  | 2.0581035             | 6.3170625       | 5,210,784                    |
| 3    | 5.266125          | 3.2271875         | 2.0389375 | 2.039479              | 5.8863545       | 5,374,600                    |

판정 estimator는 세 delta의 median **2.0389375ms \<=3ms**다. build 세 median은 모두
\<50ms, 전체 version peak footprint 최댓값은 **5,390,984 bytes \<10,000,000**다.
따라서 이 정확한 macOS 공개 binary·harness·조건의 measurement outcome은 PASS다.
이는 모든 환경의 안정성이나 생산 코드 최적화 효과를 증명하지 않는다.

129개 load/witness checkpoint의 최대 load1/CPU는 0.5322265625였다.
session-end checkpoint에서 다음 session-start checkpoint까지의 관측 간격은
60.130709초·60.120758초다. attempt 완료 시각을 별도로 기록한 값은 아니다.
harness는 attempt 완료 후 다음 시작까지 최소 60초 대기를 별도로 강제한다.
witness의 소유 heavy job은 비어 있었으며 유효 owner/execution을
유지했다. 다른 세션·VM의 전역 작업 부재를 증명하는 자료는 아니다. 462개 실행의 소유
group cleanup 기록은 모두 running_members_after=false다. 가장 큰 unscored cleanup은
0.086542ms이며 scored elapsed에 더하지 않았다. coordinator 실행 PID와 소유 control
임시 fixture도 실행 후 부재를 확인했다. 제거된 control의 현재 hash를 다시 잰 것이 아니라
원문의 buildinfo/hash와 실행 중 identity 검증 결과를 독립 검토했다.

참고값도 버리지 않았다. 세션 첫 version은 292.032166/9.599833/5.675291ms이고,
새 inode version의 15표본은 84.557125–228.547ms다. version maximum RSS는
11,370,496–11,780,096 bytes로 10MB를 넘지만 판정 metric인 peak footprint와 구분한다.
과거 10.588ms 미달 및 다른 30×3·25×3 계열은 이 estimator로 소급 판정하거나 이번
counterbalanced 계열에 섞지 않는다. OS loader·cache·보안 검사의 인과 효과도 분리하지 않았다.

[원문 JSON](startup-qualification-2026-10-26.json)은 parent가 수집한 276,100 bytes를
수정 없이 보존한다. SHA256은
`4ded8b3f15628f16715d3daf10f3ec8c5a1881f5aa6281532aaaedc1c77adba3`다.
warm/fresh/first/memory 전체 표본과 실행 순서·source/control identity·witness·cleanup을
포함한다. full process 목록이나 환경 dump는 없다. 별도 strong reviewer가 원문으로
표본 수·산술·부하·provenance를 재계산해 PASS로 판정했다. 이 문서의 측정 PASS는
최종 소유 quality/coverage/build gate와 source 통합 완료를 선기록하는 것이 아니다.

## CLI 시작·메모리 관측 — 2026-10-23

당시 PRODUCT.md G2 / Quality Gates의 startup \<10ms, 10모듈 build \<50ms,
메모리 \<10MB 목표에 대한 역사적 관측이다. 이 관측 당시에는 metric이 정해지지 않았고,
이후 당시 G2가 warm median과
peak footprint를 판정값으로 정했다. 아래는 재현 가능한 **관측 방법**과 그
기준에 따른 판정이다.

### 재현 방법과 경계

저장소 루트의 `make perf-startup`은 `make build`로 로컬 바이너리를 만들고
`scripts/startup-profile.py`를 실행한다. Python 3 표준 라이브러리, Go,
macOS `/usr/bin/time -l`, `xattr`, `sysctl`, `sw_vers`가 필요하다.
선택한 binary의 quarantine 속성이 있으면 거부하며 제거하지 않는다.
현재 구현은 macOS 전용 관측기이며 다른 OS의 메모리 숫자로 바꾸어 출력하지 않는다.
`--binary`로 이미 빌드된 파일을 선택할 수 있다. Make와 공개 release flags를
혼동하지 않도록 `--label`과 binary/control buildinfo를 함께 기록한다.

- 바이너리 SHA256·Go buildinfo·source HEAD·OS/build·CPU·시각을 기록한다.
  source HEAD와 실제 binary buildinfo를 모두 보존해 dirty build를 숨기지 않는다.
- `--version`, `--help`, 10모듈 build, 빈 Go `main` 대조군을 각각 실행한다.
  build는 task 임시 manifest와 파일을 사용한다. 사용자 dotfile은 읽거나 배포하지 않는다.
- 각 workload의 첫 관측 1회와 같은 경로의 후속 30회를 구분한다.
  후속 표본의 raw 값·min·median·nearest-rank p95·max를 모두 기록한다.
  외부 Python 프로세스의 `perf_counter_ns`로 subprocess 생성부터 종료까지 잰다.
  프로세스 생성과 stdout/stderr 수집 비용이 포함되므로 Go 함수 실행 시간과 다르다.
- 로컬 executable을 새 inode로 복사해 첫 실행을 각각 5회 더 관측한다.
  shellforge와 빈 Go 대조군의 실행 순서를 번갈아 배치한다. 이는 **fresh-copy**
  표본이며 filesystem/dyld/security cache가 차가움을 입증하지 않는다.
  reboot, cache purge, 전역 보안 설정 변경이나 quarantine 우회를 하지 않는다.
- 메모리는 각 workload를 별도 프로세스로 5회 실행한 `/usr/bin/time -l`의
  `maximum resident set size`와 `peak memory footprint`를 각각 bytes로 기록한다.
  둘은 서로 대체할 수 없으며 elapsed 표본과도 별도 실행이다.
- `GODEBUG=inittrace=1`은 마지막 별도 프로세스에서만 사용한다. tracing 자체의
  overhead가 있으므로 그 시간과 uninstrumented elapsed를 직접 빼지 않는다.
- 전체 raw JSON은 `tmp/startup-profile.json`에 쓴다. 측정 fixture와 대조군은
  이 실행이 만든 고유 tmp 하위에서 생성·정리한다. 결과 JSON은 검토를 위해 남긴다.
  직접 호출할 때 `--samples`, `--fresh-copies`, `--output`으로 관측 수와 출력만
  바꿀 수 있다. 임의의 count를 제품의 정식 수용 기준으로 간주하지 않는다.

### 공개 v0.6.0 바이너리 관측 — 주 제품 표본

Make 개발 빌드는 debug 정보를 포함하지만 배포 설정 `.goreleaser.yaml`은
`CGO_ENABLED=0`, `-trimpath`, `-s -w`를 사용한다. 따라서 아래 공개 artifact를
주 제품 표본으로 따로 측정했고 Make 관측과 섞어 개선율을 계산하지 않았다.

- URL: 공개 v0.6.0의 `gz-shellforge_0.6.0_darwin_arm64.tar.gz`.
- archive SHA256: `3a15e20e043764c2c26672e5a5cfe22539f95934930b09e9ea2adc7769b8bf85`.
- 추출 binary SHA256: `8b0954ef84349a368455541b97a1cafbe2aa41214f7be11fa0ca21cdb88b20b1`.
- binary 5,710,418 bytes, Go 1.27.3, `vcs.revision=89baf366`, `vcs.modified=false`.
  task tmp에만 다운로드·추출했으며 사용자 설치본과 공개 asset은 바꾸지 않았다.
- 2026-10-23 14:49:56 UTC, 같은 Apple M5 Max / macOS 27.0.1.
  workload별 첫 관측 1회·warm 30회·fresh-copy 5회·별도 메모리 5회다.
- harness의 `--binary`로 해당 파일을 선택했다. 빈 Go 대조군도 같은 Go 1.27.3,
  CGO=0, trimpath, `-s -w`로 빌드했다. flags는 공개 배포 설정에서 가져온 것이며
  관측기는 `--control-trimpath`, `--control-ldflags`, `--control-cgo` 입력과
  실제 control buildinfo를 모두 JSON에 보존한다. `--label`은 표본 설명이다.

| workload                 | 첫 관측 ms | warm median ms | warm p95 ms | warm max ms |
| ------------------------ | ---------: | -------------: | ----------: | ----------: |
| 공개 binary `--version`  |    359.014 |         10.588 |      13.039 |      13.670 |
| 공개 binary `--help`     |     11.255 |         11.214 |      15.733 |      16.757 |
| 공개 binary 10모듈 build |     16.428 |         13.508 |      18.941 |      23.393 |
| 배포 flags 빈 Go main    |    242.874 |          7.723 |      11.170 |      12.723 |

공개 binary의 fresh-copy 원시 표본(ms)은
113.560, 250.998, 117.317, 164.061, 112.931이며,
대조군은 76.066, 73.559, 80.535, 104.728, 73.713이다.

| workload                 | maximum RSS 범위 bytes | peak footprint 범위 bytes |
| ------------------------ | ---------------------: | ------------------------: |
| 공개 binary `--version`  |  11,468,800–11,649,024 |       5,063,304–5,243,528 |
| 공개 binary `--help`     |  11,337,728–11,747,328 |       4,932,232–5,341,832 |
| 공개 binary 10모듈 build |  12,140,544–12,500,992 |       5,325,424–5,702,280 |
| 배포 flags 빈 Go main    |    4,341,760–4,489,216 |       3,113,488–3,293,760 |

당시 G2의 절대 warm median 기준으로 판정하면 공개 binary는 `version`
warm median 10.588ms로 기동 목표(10ms 미만)를 충족하지 못했다. 메모리는 `version`
peak footprint 최댓값 5,243,528 bytes로 10MB 목표를 충족한다. maximum RSS는
10MB를 넘었지만 참고값이며 판정에 쓰지 않는다. 10모듈 build warm median
13.508ms는 50ms 목표를 충족한다.
다른 시각에 측정한 Make 표본과의 차이만으로 strip flags의 인과 효과를 주장하지 않는다.
전체 raw 값은 `tmp/startup-profile-public-v0.6.0.json`에 남겼다.

### Make 개발 바이너리 관측 — 보조 표본

2026-10-23 14:41:33 UTC, Apple M5 Max, macOS 27.0.1 (26A434),
Go 1.27.3 darwin/arm64, Python 3.14.5. 생산 소스는 `89baf366`이며
Make 바이너리 크기는 8,463,778 bytes다. 이 후속에서는 생산 Go 코드를 바꾸지 않았다. 이 보조 표본의 빈 main은
trimpath만 추가한 참고 대조군이었고, 위 공개 표본에서만 배포 flags를 맞췄다.
측정 binary의 `vcs.modified=true`는 README·측정기 작업 diff를 포함한다. 공개 릴리스
자산을 바꾼 것이 아니며, 해당 릴리스의 정확한 bytes를 측정했다고 주장하지 않는다.
이전 25회 측정의 288.850ms 첫 호출과 다른 표본이므로 전후 개선율로 비교하지 않는다.

| workload          | 첫 관측 ms | warm median ms | warm p95 ms | warm max ms |
| ----------------- | ---------: | -------------: | ----------: | ----------: |
| `--version`       |    284.105 |          7.652 |       9.455 |       9.597 |
| `--help`          |      7.863 |          6.872 |       7.967 |       8.613 |
| 10모듈 build      |      8.034 |          7.619 |       8.324 |       8.396 |
| 빈 Go main 대조군 |    233.366 |          4.276 |       5.039 |       5.417 |

fresh-copy 첫 실행의 원시 표본(ms):

- shellforge: 111.929, 109.877, 102.379, 102.787, 100.496.
- 빈 Go main: 64.131, 64.198, 86.183, 60.779, 64.446.

| workload     | maximum RSS 범위 bytes | peak footprint 범위 bytes |
| ------------ | ---------------------: | ------------------------: |
| `--version`  |  11,501,568–11,812,864 |       5,128,840–5,456,568 |
| `--help`     |  11,567,104–11,616,256 |       5,128,840–5,194,400 |
| 10모듈 build |  12,206,080–12,419,072 |       5,407,368–5,620,360 |
| 빈 Go main   |    4,243,456–4,571,136 |       3,015,184–3,359,272 |

생산 binary를 다시 빌드하지 않은 확인 실행(14:47:01 UTC, 같은 표본 수)에서는
`--version` 첫 관측 5.689ms, warm median 4.539ms / p95 5.046ms /
max 5.116ms였고, 10모듈 build는 warm median 5.033ms / max 6.381ms였다.
반면 매 실행 새로 만든 빈 Go main의 첫 관측은 255.333ms였다.
이 실행의 fresh-copy 원시 표본(ms)은 shellforge
308.338, 120.950, 157.037, 127.030, 105.800이고, 빈 Go main은
72.510, 74.592, 67.688, 67.123, 114.774였다. 같은 application code에서도
관측 분포가 달라진다. 이를 코드 최적화나 안정적인 cold-start 통과로 보고하지 않는다.
확인 실행 JSON은 `tmp/startup-profile-confirmation.json`에 따로 남겼다.

### 초기화 원인 조사와 최적화 판단

`BenchmarkNewRootCmd`는 프로세스 launch와 command registration을 분리한다.
3회 결과는 79,800 / 101,710 / 80,475 ns/op, 약 52,802 B/op,
305 allocs/op다. allocation profile의 주요 항목은 pflag의 flag 등록과
FlagSet이 사용하는 buffered writer다. 반복 benchmark의 누적 allocation을
한 번의 실행 RSS로 보고하지 않는다.

별도 inittrace의 initializer clock 합은 0.319ms, 약 99,304 bytes였다.
그중 domain의 exported regexp 초기화는 0.050ms, YAML은 0.036ms였다.
공개 바이너리의 별도 inittrace clock 합도 0.446ms였으며 allocation은 99,304 bytes였다.
이 단일 traced 표본이나 command constructor의 0.08–0.10ms만으로
100–284ms 첫 실행 tail을 설명할 수 없다. 빈 Go main에서도 큰 첫 실행 tail을
관측했으므로 문제가 CLI command 등록에만 있다고 단정할 근거도 없다.
이 대조군은 OS loader·cache·보안 검사를 각각 격리한 실험이 아니므로 어느 것을
확정 원인으로 지목하지 않는다. executable 크기와 import graph도 두 바이너리 사이에 다르다.

따라서 lazy registration이나 exported regexp의 변경을 추가하지 않았다.
공개 API·CLI 도움말·배포 안전 계약을 바꾸는 추측성 최적화는 이 근거로 정당화되지 않는다.
startup warm median 10.588ms가 목표보다 높으므로 전체 성능 PASS는 아니다.
첫 실행 시간과 maximum RSS는 참고값이다. 더 세분화한 launch tracing이 남는다.
아래의 기존 diff benchmark는 별도 workload의 과거 자료이며 이번 CLI 관측의 대용이 아니다.

## Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./internal/infra/diffcomparator/

# Run specific benchmark
go test -bench=BenchmarkComparator_Compare_Identical -benchmem ./internal/infra/diffcomparator/

# Run with CPU profiling
go test -bench=. -benchmem -cpuprofile=cpu.prof ./internal/infra/diffcomparator/

# Run with memory profiling
go test -bench=. -benchmem -memprofile=mem.prof ./internal/infra/diffcomparator/

# Generate benchmark comparison (after making changes)
go test -bench=. -benchmem ./internal/infra/diffcomparator/ > new.txt
benchstat old.txt new.txt
```

## Benchmark Categories

### 1. Identical File Comparison

Tests performance with files that have no differences.

**Key Metrics** (Apple M1 Ultra):

- **Small (10 lines)**: ~1.2 µs/op, 4 KB allocated
- **Medium (100 lines)**: ~6.9 µs/op, 28 KB allocated
- **Large (1,000 lines)**: ~57 µs/op, 262 KB allocated
- **Very Large (10,000 lines)**: ~540 µs/op, 2.6 MB allocated

**Complexity**: O(n) where n = number of lines
**Memory**: Linear growth with file size

### 2. File Additions

Tests performance when lines are added to the original file.

**Key Metrics**:

- **Small 10% additions (100→110 lines)**: ~30 µs/op, 67 KB allocated
- **Small 50% additions (100→150 lines)**: ~41 µs/op, 81 KB allocated
- **Medium 10% additions (1,000→1,100 lines)**: ~290 µs/op, 705 KB allocated
- **Medium 50% additions (1,000→1,500 lines)**: ~353 µs/op, 864 KB allocated
- **Large 10% additions (10,000→11,000 lines)**: ~2.9 ms/op, 6.5 MB allocated

**Complexity**: O(n × m) where n, m = line counts
**Memory**: Proportional to diff output size

### 3. Line Modifications

Tests performance when existing lines are modified.

**Key Metrics**:

- **Small 10% modified (100 lines)**: ~74 µs/op, 155 KB allocated
- **Small 50% modified (100 lines)**: ~243 µs/op, 414 KB allocated
- **Medium 10% modified (1,000 lines)**: ~5.3 ms/op, 9.6 MB allocated
- **Medium 50% modified (1,000 lines)**: ~20 ms/op, 31 MB allocated
- **Large 10% modified (10,000 lines)**: ~516 ms/op, 895 MB allocated

**Complexity**: O(n × m × d) where d = percentage of differences
**Memory**: Higher allocation with more modifications (diff hunks)

**⚠️ Performance Note**: Modifications are more expensive than additions due to the LCS (Longest Common Subsequence) algorithm used by go-difflib.

### 4. Different Output Formats

Compares performance of different diff output formats on a 1,000-line file with 50% modifications.

**Key Metrics**:

- **Summary**: ~18.8 ms/op, 31 MB allocated
- **Unified**: ~18.2 ms/op, 31 MB allocated
- **Context**: ~39 ms/op, 62 MB allocated (2x slower)
- **SideBySide**: ~18.6 ms/op, 32 MB allocated

**Recommendation**:

- Use `summary` format for statistics-only comparisons
- Use `unified` for fastest diff output
- Avoid `context` format for large files (2x overhead)

### 5. Real-World Scenarios

Simulates typical shell configuration file migrations.

**Key Metrics**:

- **Small config (50 lines, minor changes)**: ~41 µs/op, 70 KB allocated
- **Medium config (200 lines, moderate changes)**: ~87 µs/op, 155 KB allocated
- **Large config (500 lines, major restructuring)**: ~2.5 ms/op, 4.3 MB allocated

**Real-World Performance**:

- Typical .zshrc (50-200 lines): **< 100 µs** (sub-millisecond)
- Complex .bashrc (500+ lines): **< 3 ms** (imperceptible to user)

### 6. Function-Level Benchmarks

#### parseStatistics

Direct line comparison without diff generation.

**Key Metrics**:

- **Small (100 lines)**: ~200 ns/op, 0 allocations
- **Medium (1,000 lines)**: ~2 µs/op, 0 allocations
- **Large (10,000 lines)**: ~20 µs/op, 0 allocations

**Complexity**: O(n), zero allocations
**Performance**: Extremely fast, no garbage collection overhead

#### splitLines

String splitting into line arrays.

**Key Metrics**:

- **Small (100 lines)**: ~1.5 µs/op, 1.8 KB allocated
- **Medium (1,000 lines)**: ~14.6 µs/op, 16 KB allocated
- **Large (10,000 lines)**: ~135 µs/op, 164 KB allocated
- **Very Large (100,000 lines)**: ~1.3 ms/op, 1.6 MB allocated

**Complexity**: O(n), single allocation
**Memory**: Proportional to file size

## Performance Optimization Tips

### 1. File Size Considerations

- **< 1,000 lines**: Excellent performance (< 100 µs)
- **1,000-10,000 lines**: Good performance (< 10 ms)
- **> 10,000 lines**: Consider streaming or chunking

### 2. Diff Format Selection

**For Quick Comparison**:

```go
// Fastest - statistics only
result, _ := comparator.Compare(orig, gen, domain.DiffFormatSummary)
```

**For Human Review**:

```go
// Fast - git-style diff
result, _ := comparator.Compare(orig, gen, domain.DiffFormatUnified)
```

**Avoid for Large Files**:

```go
// Slowest - 2x overhead
result, _ := comparator.Compare(orig, gen, domain.DiffFormatContext)
```

### 3. Memory Management

**Current Implementation**:

- Loads entire files into memory
- Suitable for typical config files (< 1 MB)

**For Very Large Files** (future optimization):

- Consider streaming line-by-line comparison
- Implement chunked diff generation
- Add file size limit checks

### 4. Algorithmic Complexity

**Current go-difflib Implementation**:

- LCS algorithm: O(n × m) time, O(n × m) space
- Works well for typical config files
- May struggle with very large files or high modification rates

**Optimization Opportunities**:

- For simple additions: Use faster append-only comparison
- For identical files: Early exit after byte comparison
- For structured configs: Use domain-aware diffing

## Benchmark Interpretation Guide

### Understanding the Metrics

```
BenchmarkComparator_Compare_Identical/Small_10-20    971155    1191 ns/op    4000 B/op    11 allocs/op
                                              │         │         │             │            │
                                              │         │         │             │            └─ Allocations per operation
                                              │         │         │             └─ Bytes allocated per operation
                                              │         │         └─ Nanoseconds per operation
                                              │         └─ Number of iterations
                                              └─ CPU cores used (20 = 10 cores × 2)
```

### Performance Thresholds

**Excellent** (< 1 ms):

- Imperceptible to users
- Suitable for interactive CLI operations
- No optimization needed

**Good** (1-10 ms):

- Acceptable for CLI operations
- May benefit from progress indicators
- Consider caching for repeated operations

**Acceptable** (10-100 ms):

- Noticeable but tolerable
- Should show progress for user feedback
- Consider optimization if frequently used

**Slow** (> 100 ms):

- User will notice delay
- Requires progress indicators
- Should optimize or warn about large files

### Memory Allocation Analysis

**Low Allocation** (< 1 MB):

- Minimal GC pressure
- No memory concerns

**Moderate Allocation** (1-10 MB):

- Acceptable for CLI tools
- May trigger minor GC collections

**High Allocation** (> 10 MB):

- Consider memory optimization
- May cause GC pauses
- Monitor for large file operations

## Regression Testing

To detect performance regressions:

```bash
# Before changes
go test -bench=. -benchmem ./internal/infra/diffcomparator/ > before.txt

# After changes
go test -bench=. -benchmem ./internal/infra/diffcomparator/ > after.txt

# Compare
benchstat before.txt after.txt
```

**Acceptable Regression**:

- < 10% slower: Minor variation
- 10-20% slower: Acceptable if feature adds value
- > 20% slower: Investigate and optimize

**Performance Improvement**:

- > 10% faster: Good improvement
- > 50% faster: Significant optimization
- > 2x faster: Major algorithmic improvement

## Future Optimization Opportunities

### Short-Term (< 1 week)

1. **Early Exit for Identical Files**

   - Compare file hashes before line-by-line diff
   - Expected improvement: 10-50% for identical files

1. **Lazy Statistics Calculation**

   - Only calculate statistics when requested
   - Expected improvement: 20-30% for non-summary formats

1. **Buffer Pool for String Building**

   - Reuse buffers for repeated operations
   - Expected improvement: 5-10% reduction in allocations

### Medium-Term (1-4 weeks)

1. **Streaming Line Comparison**

   - Process files line-by-line without loading fully into memory
   - Target: Support files > 100 MB

1. **Parallel Processing**

   - Split file into chunks for parallel comparison
   - Expected improvement: 2-4x for very large files

1. **Smart Diff Algorithm Selection**

   - Use simpler algorithms for append-only changes
   - Expected improvement: 50-100% for specific patterns

### Long-Term (1-3 months)

1. **Custom LCS Implementation**

   - Optimize for shell config file patterns
   - Target: 2x improvement over go-difflib

1. **Incremental Diff Caching**

   - Cache previous diff results
   - Target: 10x improvement for repeated comparisons

1. **Binary Format Support**

   - Efficient storage of diff metadata
   - Target: 50% reduction in memory usage

## Related Documentation

- [ARCHITECTURE.md](../ARCHITECTURE.md) - System design
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Development guidelines
- [comparator_test.go](../internal/infra/diffcomparator/comparator_test.go) - Unit tests
- [comparator_bench_test.go](../internal/infra/diffcomparator/comparator_bench_test.go) - Benchmark code

## Benchmark Maintenance

**Update Frequency**: Run benchmarks for each release

**Regression Check**: Compare against previous version

**Documentation**: Update this file when:

- Adding new benchmark scenarios
- Making algorithmic changes
- Observing significant performance changes
- Updating hardware or Go version

______________________________________________________________________

**Last Updated**: 2025-11-30
**Go Version**: 1.21+
**Test Hardware**: Apple M1 Ultra (20 cores)
**Benchmark Version**: v0.5.0
