#!/bin/bash
#
# Benchmark comparison helper script for gh-wm
# Compares current performance against a baseline (default: main branch)
#
# Usage: ./scripts/bench-compare.sh [baseline-ref]
#
# Examples:
#   ./scripts/bench-compare.sh                    # Compare against main
#   ./scripts/bench-compare.sh v1.0.0             # Compare against tag
#   ./scripts/bench-compare.sh HEAD~5             # Compare against 5 commits ago

set -euo pipefail

baseline=${1:-main}
current_branch=$(git rev-parse --abbrev-ref HEAD)

echo "📊 Benchmark Comparison"
echo "  Baseline: $baseline"
echo "  Current:  $current_branch"
echo ""

# Ensure we have a clean state
if [ -n "$(git status --porcelain)" ]; then
    echo "⚠️  Working directory has uncommitted changes"
    echo "   Results may not be representative"
    echo ""
fi

# Create temp directory for results
temp_dir=$(mktemp -d)
baseline_file="$temp_dir/baseline.txt"
current_file="$temp_dir/current.txt"

echo "🔄 Running baseline benchmarks ($baseline)..."
git stash -u > /dev/null 2>&1 || true
git checkout "$baseline" > /dev/null 2>&1
go test -bench=. -benchmem -count=3 ./... > "$baseline_file" 2>/dev/null || {
    echo "❌ Failed to run benchmarks on $baseline"
    git checkout "$current_branch" > /dev/null 2>&1
    git stash pop > /dev/null 2>&1 || true
    rm -rf "$temp_dir"
    exit 1
}

echo "🔄 Running current benchmarks ($current_branch)..."
git checkout "$current_branch" > /dev/null 2>&1
git stash pop > /dev/null 2>&1 || true
go test -bench=. -benchmem -count=3 ./... > "$current_file" 2>/dev/null || {
    echo "❌ Failed to run current benchmarks"
    rm -rf "$temp_dir"
    exit 1
}

echo ""
echo "📈 Results:"
echo ""

# Extract key benchmark results and compare
echo "| Benchmark | Baseline | Current | Change |"
echo "|-----------|----------|---------|--------|"

# Parse benchmark results for comparison
grep "^Benchmark" "$baseline_file" | grep -E "(ConfigLoad|ParseGlobal|ResolveMatchingTasks|MatchOnOR)" | while read -r line; do
    benchmark=$(echo "$line" | awk '{print $1}')
    baseline_ns=$(echo "$line" | awk '{print $3}')

    current_line=$(grep "^$benchmark" "$current_file" || echo "")
    if [ -n "$current_line" ]; then
        current_ns=$(echo "$current_line" | awk '{print $3}')

        # Calculate percentage change
        if [ "$baseline_ns" != "0" ]; then
            change_pct=$(echo "scale=1; (($current_ns - $baseline_ns) * 100) / $baseline_ns" | bc -l 2>/dev/null || echo "N/A")
            if [[ "$change_pct" =~ ^-?[0-9]+\.?[0-9]*$ ]]; then
                if (( $(echo "$change_pct > 0" | bc -l) )); then
                    change_str="+${change_pct}%"
                else
                    change_str="${change_pct}%"
                fi
            else
                change_str="N/A"
            fi
        else
            change_str="N/A"
        fi

        echo "| $benchmark | ${baseline_ns} ns | ${current_ns} ns | $change_str |"
    fi
done

# Check binary size difference
echo ""
echo "📦 Binary Size Comparison:"
echo ""

git checkout "$baseline" > /dev/null 2>&1
go build -trimpath -ldflags="-s -w" -o gh-wm-baseline . 2>/dev/null
baseline_size=$(wc -c < gh-wm-baseline)

git checkout "$current_branch" > /dev/null 2>&1
go build -trimpath -ldflags="-s -w" -o gh-wm-current . 2>/dev/null
current_size=$(wc -c < gh-wm-current)

size_change_pct=$(echo "scale=1; (($current_size - $baseline_size) * 100) / $baseline_size" | bc -l 2>/dev/null || echo "N/A")

echo "| Metric | Baseline | Current | Change |"
echo "|--------|----------|---------|--------|"
echo "| Binary Size | $(numfmt --to=iec $baseline_size) | $(numfmt --to=iec $current_size) | ${size_change_pct}% |"

# Cleanup
rm -f gh-wm-baseline gh-wm-current
rm -rf "$temp_dir"

echo ""
echo "✅ Comparison complete"