# Performance Benchmarks

This document describes the performance characteristics of the diff comparison system and provides benchmark results for reference.

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

2. **Lazy Statistics Calculation**
   - Only calculate statistics when requested
   - Expected improvement: 20-30% for non-summary formats

3. **Buffer Pool for String Building**
   - Reuse buffers for repeated operations
   - Expected improvement: 5-10% reduction in allocations

### Medium-Term (1-4 weeks)

1. **Streaming Line Comparison**
   - Process files line-by-line without loading fully into memory
   - Target: Support files > 100 MB

2. **Parallel Processing**
   - Split file into chunks for parallel comparison
   - Expected improvement: 2-4x for very large files

3. **Smart Diff Algorithm Selection**
   - Use simpler algorithms for append-only changes
   - Expected improvement: 50-100% for specific patterns

### Long-Term (1-3 months)

1. **Custom LCS Implementation**
   - Optimize for shell config file patterns
   - Target: 2x improvement over go-difflib

2. **Incremental Diff Caching**
   - Cache previous diff results
   - Target: 10x improvement for repeated comparisons

3. **Binary Format Support**
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

---

**Last Updated**: 2025-11-30
**Go Version**: 1.21+
**Test Hardware**: Apple M1 Ultra (20 cores)
**Benchmark Version**: v0.5.0
