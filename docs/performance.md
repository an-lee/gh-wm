# Performance Measurement Guide

This guide covers performance measurement strategies for gh-wm development.

## Quick Start

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Compare performance against main branch  
./scripts/bench-compare.sh

# Profile CPU usage during task resolution
./scripts/perf-profile.sh cpu ./gh-wm resolve --repo-root . --event-name issues --payload event.json
```

## Benchmark Coverage

gh-wm has benchmark tests covering critical code paths:

| Component | Benchmark | What It Measures |
|-----------|-----------|------------------|
| Config | `BenchmarkConfigLoad` | Full config + tasks loading |
| Config | `BenchmarkParseGlobal` | Global config parsing only |
| Engine | `BenchmarkResolveMatchingTasks` | Task resolution logic |
| Trigger | `BenchmarkMatchOnOR_Issues` | Issue event matching |
| Trigger | `BenchmarkMatchOnOR_SlashCommand` | Slash command matching |

### Performance Baselines (as of 2026-04-16)

- **Config loading**: ~22μs, 20KB allocations
- **Task resolution**: ~22μs, 20KB allocations  
- **Trigger matching**: 15-45ns, zero allocations
- **Binary size**: 9.2MB (optimized), ~29% reduction from debug build

## CI Performance Monitoring

The `.github/workflows/performance.yml` workflow runs on all PRs and tracks:

- Benchmark results with 3-run averages
- Binary size comparisons (debug vs optimized)
- Performance regression alerts in PR comments
- Artifact storage of detailed results (30 days)

### Regression Detection

Performance changes are flagged if they exceed these thresholds:

- Config loading: >30μs (target <30μs)
- Task resolution: >30μs (target <30μs)
- Trigger matching: >100ns (target <100ns)
- Binary size: >10MB optimized (target <10MB)

## Profiling Tools

### CPU Profiling

```bash
# Profile a specific command
./scripts/perf-profile.sh cpu ./gh-wm resolve --repo-root . --event-name issues --payload event.json

# View results in terminal
go tool pprof profile.cpu

# View results in web UI
go tool pprof -http=:8080 profile.cpu
```

Best for: Algorithm optimization, identifying hot code paths.

### Memory Profiling

```bash
# Profile memory allocation
./scripts/perf-profile.sh mem ./gh-wm run --task daily-doc-updater --event-name workflow_dispatch

# Analyze allocations
go tool pprof profile.mem
```

Best for: Memory leak detection, allocation optimization.

### Execution Tracing

```bash
# Capture execution trace
./scripts/perf-profile.sh trace ./gh-wm --version

# View trace visualization
go tool trace trace.out
```

Best for: Understanding goroutine behavior, lock contention, GC impact.

## Common Workflows

### Before/After Measurement

```bash
# Establish baseline
git checkout main
go test -bench=BenchmarkConfigLoad -count=5 ./internal/config > baseline.txt

# Test your changes
git checkout feature-branch  
go test -bench=BenchmarkConfigLoad -count=5 ./internal/config > feature.txt

# Compare results
benchcmp baseline.txt feature.txt
```

### Binary Size Tracking

```bash
# Measure impact of changes on binary size
./scripts/bench-compare.sh main

# Or manually compare builds
go build -trimpath -ldflags="-s -w" -o gh-wm-opt .
go build -o gh-wm-debug .
ls -lah gh-wm-*
```

### Performance Investigation

1. **Start broad**: Run full benchmark suite to identify affected areas
2. **Narrow down**: Focus benchmarks on specific components  
3. **Profile deeply**: Use CPU/memory profiling for detailed analysis
4. **Verify fixes**: Use before/after comparison to validate improvements

## Measurement Best Practices

### Benchmark Design

- **Isolate what you're testing**: Use `b.ResetTimer()` to exclude setup
- **Use realistic inputs**: Mirror real-world usage patterns
- **Test memory allocations**: Always include `-benchmem`
- **Run multiple iterations**: Use `-count=3` for stable results

### Environment Considerations

- **CPU scaling**: Disable for consistent results: `sudo cpupower frequency-set --governor performance`
- **Background processes**: Minimize system load during measurements
- **Warm-up**: JIT compilers need warm-up runs; Go is generally stable

### Interpreting Results

- **Variance**: >10% variance between runs suggests unstable measurements
- **Allocations**: Zero allocations in hot paths is ideal
- **Time vs space trade-offs**: Faster code may use more memory
- **Real-world impact**: Microsecond improvements may not affect user experience

## Performance Targets

| Component | Target | Current | Notes |
|-----------|--------|---------|-------|
| Config loading | <30μs | ~22μs | ✅ Well within target |
| Task resolution | <30μs | ~22μs | ✅ Well within target |  
| Trigger matching | <100ns | 15-45ns | ✅ Excellent performance |
| Binary size | <10MB | 9.2MB | ✅ Good compression |
| Test suite | <5s | ~1.5s | ✅ Fast feedback |

## Adding New Benchmarks

When adding new performance-critical code:

1. **Add benchmark tests** in the same package as your code
2. **Follow naming convention**: `BenchmarkFunctionName`
3. **Include memory tracking**: Use `-benchmem` in CI
4. **Set realistic targets**: Based on user-facing impact
5. **Update this documentation**: Add new targets to the table above

Example benchmark:

```go
func BenchmarkMyFunction(b *testing.B) {
    // Setup (excluded from timing)
    input := generateTestInput()
    
    b.ResetTimer() // Start timing here
    for i := 0; i < b.N; i++ {
        MyFunction(input)
    }
}
```