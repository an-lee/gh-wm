# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# gh-wm (this repository)

You are working on the **gh-wm** CLI: a Go `gh` extension that resolves GitHub events to `.wm/tasks/*.md` tasks and runs an agent subprocess (`claude -p` by default, or `WM_AGENT_CMD`).

## Commands

```bash
# Build
go build -o gh-wm .

# Test all packages
go test ./...

# Run a single test
go test ./internal/config/... -run TestSplitFrontmatter

# Run resolve manually
./gh-wm resolve --repo-root . --event-name issues --payload /path/to/event.json --json

# Run a task manually (requires ANTHROPIC_API_KEY or WM_AGENT_CMD)
./gh-wm run --repo-root . --task <task-name> --event-name issues --payload /path/to/event.json
```

## Architecture

The core pipeline is **event → resolve → run agent**:

1. **`cmd/`** — Cobra CLI; thin wrappers that delegate to `internal/`.
2. **`internal/engine/resolver.go`** — `ResolveMatchingTasks`: loads all `.wm/tasks/*.md`, calls `trigger.MatchOnOR`, returns matching task names as JSON.
3. **`internal/trigger/match.go`** — `MatchOnOR`: OR-semantics over `on:` frontmatter keys (`issues`, `issue_comment`, `pull_request`, `slash_command`, `schedule`, `workflow_dispatch`).
4. **`internal/engine/runner.go` + `agent.go`** — `RunTask` → `runAgent`: builds a text prompt from the task body + `context.files`, then execs `WM_AGENT_CMD` or `claude -p <prompt>`. The `run` command has a **45-minute** hard timeout.
5. **`internal/config/`** — loads `.wm/config.yml` (GlobalConfig) and parses frontmatter from `.wm/tasks/*.md` (Task). Frontmatter is `map[string]any`; add typed accessors on `Task` when a field becomes first-class.

**Agents receive the task body as a plain-text prompt** via `exec.Cmd`. There is no structured output contract — the agent is expected to use Git/`gh` directly. There is no `internal/output` package yet; it is planned but not wired.

## Non-obvious design constraints

- **Binary name duality**: In CI, `go install` produces `gh-wm` (used directly). When installed as a `gh` extension, commands are `gh wm …`. Both call the same binary.
- **`wm-agent.yml` is generated**: Consumer repos get this file from `gh wm init` / `gh wm upgrade` (via `internal/gen/wmagent.go`). It calls into **this** repo's reusable workflows (`agent-resolve.yml`, `agent-run.yml`). Do not hand-edit generated caller files; change the template in `internal/gen/wmagent.go`.
- **Schedule cron filtering**: At resolve time, all tasks with `on.schedule` match. The `WM_SCHEDULE_CRON` env var (set to the workflow's cron string) further filters so only the right task fires.
- **`engine:` frontmatter field** is parsed but not acted on yet — `WM_AGENT_CMD` is the only override mechanism today.
- **`internal/checkpoint/`** exists (encode/decode `<!-- wm-checkpoint: … -->` comments) but is not yet called from `RunTask`.

## Before changing behavior

1. Read **[`docs/README.md`](docs/README.md)** for the mental model.
2. For code changes, read **[`docs/architecture.md`](docs/architecture.md)** and **[`docs/development.md`](docs/development.md)**.
3. For task markdown / frontmatter, use **[`docs/task-format.md`](docs/task-format.md)**.
4. For CLI flags and env vars, use **[`docs/cli-reference.md`](docs/cli-reference.md)**.

## Accuracy

Keep docs in **`docs/`** aligned with code (`internal/engine`, `internal/trigger`, `cmd/`). If you add flags, matchers, or workflow contract changes, update the relevant doc in the same change.

## Templates vs this repo

`gh wm init` writes embedded templates from [`internal/templates/data/`](internal/templates/data/) into **user repos**. That is separate from the documentation in **`docs/`**, which describes **this** project.
