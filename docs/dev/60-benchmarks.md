# Performance Benchmarks

This document records CLI startup and memory observations alongside diff comparison benchmarks.

## CLI 시작·메모리 관측 — 2026-10-23

PRODUCT.md G2 / Quality Gates의 startup \<10ms, 10모듈 build \<50ms,
메모리 \<10MB 목표는 그대로 유지한다. PRODUCT는 cold/warm, 표본 수,
percentile, 메모리 metric과 MB 단위를 정하지 않았다. 아래는 그 빈틈을 드러내는
재현 가능한 **관측 방법**이며 승인된 수용 기준을 새로 정의하지 않는다.
특히 warm median이나 footprint만 골라 전체 성능 PASS라고 해석하지 않는다.

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

공개 binary는 이 표본에서 warm median도 10ms를 넘었고 maximum RSS도 10MB를
넘었다. stripped 크기나 작은 footprint를 이유로 목표를 충족했다고 해석하지 않는다.
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
startup의 첫 실행과 maximum RSS는 목표보다 높으며 전체 성능 PASS는 아니다.
수용 workload/metric 정의와 더 세분화한 launch tracing이 남는다.
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
