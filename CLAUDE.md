# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# gh-wm (this repository)

You are working on the **gh-wm** CLI: a Go `gh` extension that resolves GitHub events to `.wm/tasks/*.md` tasks and runs an agent subprocess (default: **`claude -p`** with **`--dangerously-skip-permissions`**, prompt on stdin, optional `--model` / `--max-turns` from `.wm/config.yml`, or **`WM_AGENT_CMD`**).

## Commands

```bash
# Build
go build -o gh-wm .

# Print version (default `dev`; override at build time with -ldflags)
./gh-wm version
# go build -o gh-wm -ldflags "-X github.com/an-lee/gh-wm/cmd.Version=1.0.0 -X github.com/an-lee/gh-wm/cmd.Commit=local" .

# Test all packages
go test ./...

# Run a single test
go test ./internal/config/... -run TestSplitFrontmatter

# Run resolve manually
./gh-wm resolve --repo-root . --event-name issues --payload /path/to/event.json --json

# Omit --payload (and GITHUB_EVENT_PATH) to use an empty event `{}` for quick local runs
./gh-wm run --repo-root . --task <task-name> --event-name workflow_dispatch

# Run a task manually (requires ANTHROPIC_API_KEY or WM_AGENT_CMD)
# `run` requires a clean git working tree unless you pass --allow-dirty
./gh-wm run --repo-root . --task <task-name> --event-name issues --payload /path/to/event.json

# Dispatch wm-agent.yml on GitHub (gh CLI; regenerate wm-agent.yml with gh wm upgrade after upgrading gh-wm)
./gh-wm run --repo-root . --task <task-name> --remote
```

## Architecture

The core pipeline is **event → resolve → run agent**:

1. **`cmd/`** — Cobra CLI; thin wrappers that delegate to `internal/`.
2. **`internal/engine/resolver.go`** — `ResolveMatchingTasks`: loads all `.wm/tasks/*.md`, calls `trigger.MatchOnOR`, returns matching task names as JSON.
3. **`internal/trigger/match.go`** — `MatchOnOR`: OR-semantics over `on:` frontmatter keys (`issues`, `issue_comment`, `pull_request`, `slash_command`, `schedule`, `workflow_dispatch`).
4. **`internal/engine/runner.go` + `agent.go` + `rundir.go`** — `RunTask` → `runAgent`: builds a text prompt from the task body + `context.files` + optional safe-output instructions, writes **`prompt.md`** under **`.wm/runs/<id>/`** (or **`WM_RUN_DIR`**), sets **`WM_OUTPUT_FILE`** to **`output.json`** in that dir, streams output to **`agent-stdout.log`**, then execs `WM_AGENT_CMD` or `claude -p <prompt>`. The `run` command uses **`timeout-minutes`** from the task (default **45**). Post-agent **`internal/output/`** applies agent-written **`output.json`** **`items`** when `safe-outputs:` is non-empty (use **`noop`** if no GitHub actions).
5. **`internal/config/`** — loads `.wm/config.yml` (GlobalConfig) and parses frontmatter from `.wm/tasks/*.md` (Task). Frontmatter is `map[string]any`; add typed accessors on `Task` when a field becomes first-class.

**Agents receive the task body as a plain-text prompt** via `exec.Cmd` and must write structured **`items`** to **`WM_OUTPUT_FILE`** (`output.json`) when `safe-outputs:` is set. **`internal/output/`** validates and executes those items against `safe-outputs:` policy.

## Non-obvious design constraints

- **Binary name duality**: In CI, `go install` produces `gh-wm` (used directly). When installed as a `gh` extension, commands are `gh wm …`. Both call the same binary.
- **`wm-agent.yml` is generated**: Consumer repos get this file from `gh wm init` / `gh wm upgrade` (via `internal/gen/wmagent.go`). `gh wm upgrade` also runs **best-effort** `gh extension upgrade an-lee/gh-wm` before regenerating the workflow. **`gh wm update`** re-fetches `.wm/tasks/*.md` files that have a `source:` (https URL or `owner/repo/path` shorthand, set by `gh wm add <url | owner/repo/task | path>`). It calls into **this** repo's reusable workflows (`agent-resolve.yml`, `agent-run.yml`). Do not hand-edit generated caller files; change the template in `internal/gen/wmagent.go`.
- **Schedule cron filtering**: At resolve time, all tasks with `on.schedule` match. The `WM_SCHEDULE_CRON` env var (set to the workflow's cron string) further filters so only the right task fires.
- **`engine:` frontmatter field** selects the default agent CLI when `WM_AGENT_CMD` is unset (`claude`, `codex`; `copilot` requires `WM_AGENT_CMD`).
- **`internal/checkpoint/`** — when `WM_CHECKPOINT=1`, the runner loads the latest checkpoint from issue comments into the prompt and posts a new checkpoint comment after success.

## Before changing behavior

1. Read **[`docs/content/_index.md`](docs/content/_index.md)** for the mental model (see also **[`docs/README.md`](docs/README.md)**).
2. For code changes, read **[`docs/content/architecture.md`](docs/content/architecture.md)** and **[`docs/content/development.md`](docs/content/development.md)**.
3. For task markdown / frontmatter, use **[`docs/content/task-format.md`](docs/content/task-format.md)**.
4. For CLI flags and env vars, use **[`docs/content/cli-reference.md`](docs/content/cli-reference.md)**.

## Accuracy

Keep docs in **`docs/`** aligned with code (`internal/engine`, `internal/trigger`, `cmd/`). If you add flags, matchers, or workflow contract changes, update the relevant doc in the same change.

## Templates vs this repo

`gh wm init` writes embedded templates from [`internal/templates/data/`](internal/templates/data/) into **user repos**. That is separate from the documentation in **`docs/`**, which describes **this** project.
