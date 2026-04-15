# Development guide

How to build, test, and extend **gh-wm**. Pair this with [architecture.md](architecture.md).

## Prerequisites

- **Go** (see [`go.mod`](../../go.mod) for version).
- **`gh`** CLI for local commands that shell out (`assign`, `status`, `logs`, outputs).
- Optional: **`claude`** CLI for default `run` behavior, or set **`WM_AGENT_CMD`**.

## Build and run

```bash
# From repo root
go build -o gh-wm .

./gh-wm resolve --payload /path/to/event.json --event-name issues --json
./gh-wm run --task implement --payload /path/to/event.json --event-name issues
```

Install via module path:

```bash
go install github.com/an-lee/gh-wm@latest
```

## Documentation site (Hugo)

The HTML site at [https://gh-wm.github.io/gh-wm/](https://gh-wm.github.io/gh-wm/) is built from this repo’s [`docs/`](../../docs/) tree (Markdown in [`content/`](../../docs/content/), Hugo config in [`hugo.toml`](../../docs/hugo.toml)). Local preview:

```bash
cd docs && hugo mod get -u && hugo server
```

Deployment runs via [`.github/workflows/pages.yml`](../../.github/workflows/pages.yml) when `docs/` changes on `main`.

## Repository layout

| Path                                                              | Role                                                                                                                                   |
| ----------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| [`main.go`](../../main.go)                                        | Calls `cmd.Execute()`.                                                                                                                 |
| [`cmd/`](../../cmd/)                                              | Cobra commands; keep thin—delegate to `internal/`.                                                                                     |
| [`internal/config/`](../../internal/config/)                      | YAML + markdown frontmatter loading.                                                                                                   |
| [`internal/engine/`](../../internal/engine/)                      | Resolve + run + agent + state labels + checkpoint wiring.                                                                              |
| [`internal/output/`](../../internal/output/)                      | Post-agent steps from `safe-outputs:` keys.                                                                                            |
| [`internal/trigger/`](../../internal/trigger/)                    | `on:` matching (`match.go`).                                                                                                           |
| [`internal/types/`](../../internal/types/)                        | `GitHubEvent`, `TaskContext`, `AgentResult`.                                                                                           |
| [`internal/gen/`](../../internal/gen/)                            | `wm-agent.yml` and schedule collection.                                                                                                |
| [`internal/templates/`](../../internal/templates/)                | Embedded files for `gh wm init`.                                                                                                       |
| [`internal/ghclient/`](../../internal/ghclient/)                  | `gh api` helpers (labels, comments).                                                                                                   |
| [`internal/checkpoint/`](../../internal/checkpoint/checkpoint.go) | Checkpoint HTML comments.                                                                                                              |
| [`docs/`](../../docs/)                                            | Hugo site ([`hugo.toml`](../../docs/hugo.toml), [`content/`](../../docs/content/)) for [GitHub Pages](https://gh-wm.github.io/gh-wm/). |
| [`.github/workflows/`](../../.github/workflows/)                  | CI + reusable workflows + release + Pages.                                                                                             |

## Extending `on:` matching

1. Edit [`internal/trigger/match.go`](../../internal/trigger/match.go).
2. Add a branch inside `MatchOnOR` **or** extend an existing matcher.
3. Add **tests** where practical.
4. Document new syntax in [task-format.md](task-format.md).

**Convention:** `GITHUB_EVENT_NAME` matches GitHub’s webhook names (`issues`, `issue_comment`, …).

## Extending outputs

1. Add a function in [`internal/output/`](../../internal/output/) and call it from [`RunSuccessOutputs`](../../internal/output/output.go) when the right `safe-outputs` key is present.
2. Use [`ghclient`](../../internal/ghclient/) or `exec.Command("gh", …)` with `tc.RepoPath` / `GITHUB_REPOSITORY`.
3. Document the key in [task-format.md](task-format.md).

## Extending configuration

- **Global config**: Extend [`GlobalConfig`](../../internal/config/types.go) and document in [task-format.md](task-format.md).
- **Tasks**: Frontmatter is `map[string]any`—add accessors on [`Task`](../../internal/config/types.go) when a field becomes first-class.

## Agent backend selection

See [`runAgent`](../../internal/engine/agent.go): `WM_AGENT_CMD` overrides everything; otherwise `engine:` selects `claude`, `codex` (+ optional `WM_ENGINE_CODEX_CMD`), or **`copilot`** (must set `WM_AGENT_CMD`).

## Workflows and releases

- **Caller** `wm-agent.yml` is **generated**—use `gh wm upgrade`.
- Reusable workflows live in this repo. CI installs **`gh-wm`** with `go install` and invokes **`gh-wm resolve` / `gh-wm run`**.

When changing reusable workflow inputs/outputs, update [`internal/gen/wmagent.go`](../../internal/gen/wmagent.go).

## Tests

```bash
go test ./...
```

Add tests for parsers and `trigger.MatchOnOR` whenever behavior changes.

## Style

- Keep CLI flags and env var names stable; document in [cli-reference.md](cli-reference.md).
- Prefer explicit errors over silent skips (e.g. missing payload path).
