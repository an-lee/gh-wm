# Development guide

How to build, test, and extend **gh-wm** safely. Pair this with [architecture.md](architecture.md).

## Prerequisites

- **Go** (see [`go.mod`](../go.mod) for version).
- **`gh`** CLI for local commands that shell out (`assign`, `status`, `logs`).
- Optional: **`claude`** CLI for default `run` behavior, or set **`WM_AGENT_CMD`**.

## Build and run

```bash
# From repo root
go build -o gh-wm .

# Invoke (same as gh extension binary name in CI)
./gh-wm resolve --payload /path/to/event.json --event-name issues --json
./gh-wm run --task implement --payload /path/to/event.json --event-name issues
```

Install via module path:

```bash
go install github.com/gh-wm/gh-wm@latest
```

## Repository layout

| Path | Role |
|------|------|
| [`main.go`](../main.go) | Calls `cmd.Execute()`. |
| [`cmd/`](../cmd/) | Cobra commands; keep thin—delegate to `internal/`. |
| [`internal/config/`](../internal/config/) | YAML + markdown frontmatter loading. |
| [`internal/engine/`](../internal/engine/) | Resolve + run + agent subprocess. |
| [`internal/trigger/`](../internal/trigger/) | `on:` matching only (`match.go`). |
| [`internal/types/`](../internal/types/) | `GitHubEvent`, `TaskContext`, `AgentResult`. |
| [`internal/gen/`](../internal/gen/) | `wm-agent.yml` and schedule collection. |
| [`internal/templates/`](../internal/templates/) | Embedded files for `gh wm init`. |
| [`internal/ghclient/`](../internal/ghclient/) | Small helpers (`assign`). |
| [`internal/checkpoint/`](../internal/checkpoint/) | Checkpoint comment encode/decode (optional). |
| [`.github/workflows/`](../.github/workflows/) | Reusable Actions + release. |

## Extending `on:` matching

1. Edit [`internal/trigger/match.go`](../internal/trigger/match.go).
2. Add a branch inside `MatchOnOR` **or** extend an existing matcher (`matchIssues`, `matchSlashCommand`, etc.).
3. Add **tests**: prefer table-driven tests in a new `match_test.go` or extend [`internal/config/frontmatter_test.go`](../internal/config/frontmatter_test.go) if loading is involved.
4. Document new syntax in [task-format.md](task-format.md).

**Convention:** GitHub delivers `GITHUB_EVENT_NAME` as the key used in GitHub’s webhook docs (`issues`, `issue_comment`, …). Keep names consistent.

## Extending configuration

- **Global config**: Extend [`GlobalConfig`](../internal/config/types.go) and document fields in [task-format.md](task-format.md).
- **Tasks**: Frontmatter is `map[string]any`—add accessors on [`Task`](../internal/config/types.go) when a field becomes first-class.

## Agent backend selection

[`runAgent`](../internal/engine/agent.go) reads `task.Engine()` and global `engine` but does **not** branch yet—override behavior with **`WM_AGENT_CMD`** for experiments.

Future work: switch on `engine` to invoke Copilot/Codex/etc.

## Outputs (PR, comments, labels)

The original design described pluggable **Output** interfaces. The current **`RunTask`** path only runs the agent subprocess; there is **no** `internal/output` package yet. If you add outputs:

1. Define small interfaces in `internal/types` or `internal/output`.
2. Call them from `RunTask` after a successful agent run (or as orchestrated steps).
3. Update [architecture.md](architecture.md) and this file.

## Workflows and releases

- **Caller** `wm-agent.yml` is **generated**—do not hand-edit in consumer repos; use `gh wm upgrade`.
- Reusable workflows live in this repo: [`agent-resolve.yml`](../.github/workflows/agent-resolve.yml), [`agent-run.yml`](../.github/workflows/agent-run.yml). CI installs **`gh-wm`** with `go install` and invokes **`gh-wm resolve` / `gh-wm run`** (not `gh wm`).

When changing reusable workflow inputs/outputs, update the template in [`internal/gen/wmagent.go`](../internal/gen/wmagent.go) and verify consumers regenerate `wm-agent.yml`.

## Tests

```bash
go test ./...
```

Add tests for parsers and `trigger.MatchOnOR` whenever behavior changes.

## Style

- Keep CLI flags and env var names stable; document in [cli-reference.md](cli-reference.md).
- Prefer explicit errors over silent skips (e.g. missing payload path).
