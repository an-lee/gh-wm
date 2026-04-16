#!/bin/bash
#
# Performance profiling helper script for gh-wm
# Usage: ./scripts/perf-profile.sh [cpu|mem|trace] [command args...]
#
# Examples:
#   ./scripts/perf-profile.sh cpu ./gh-wm resolve --repo-root . --event-name issues --payload event.json
#   ./scripts/perf-profile.sh mem ./gh-wm run --task daily-doc-updater --event-name workflow_dispatch
#   ./scripts/perf-profile.sh trace ./gh-wm --version

set -euo pipefail

if [ $# -lt 2 ]; then
    echo "Usage: $0 [cpu|mem|trace] [command args...]"
    echo ""
    echo "Profile types:"
    echo "  cpu   - CPU profiling (saves to profile.cpu)"
    echo "  mem   - Memory profiling (saves to profile.mem)"
    echo "  trace - Execution trace (saves to trace.out)"
    echo ""
    echo "Example:"
    echo "  $0 cpu ./gh-wm resolve --repo-root . --event-name issues --payload event.json"
    exit 1
fi

profile_type="$1"
shift

case "$profile_type" in
    "cpu")
        echo "🔍 Starting CPU profiling..."
        CPUPROFILE=profile.cpu "$@"
        echo "✅ CPU profile saved to profile.cpu"
        echo "   View with: go tool pprof profile.cpu"
        echo "   Web UI: go tool pprof -http=:8080 profile.cpu"
        ;;
    "mem")
        echo "🧠 Starting memory profiling..."
        MEMPROFILE=profile.mem "$@"
        echo "✅ Memory profile saved to profile.mem"
        echo "   View with: go tool pprof profile.mem"
        echo "   Web UI: go tool pprof -http=:8080 profile.mem"
        ;;
    "trace")
        echo "🔗 Starting execution trace..."
        TRACE=trace.out "$@"
        echo "✅ Execution trace saved to trace.out"
        echo "   View with: go tool trace trace.out"
        ;;
    *)
        echo "❌ Unknown profile type: $profile_type"
        echo "   Supported: cpu, mem, trace"
        exit 1
        ;;
esac