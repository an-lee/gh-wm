# CLI reference

The extension is invoked as **`gh wm <subcommand>`** when installed via `gh extension install`. Building from source produces a binary named **`gh-wm`** (same commands as the root Cobra app in [`cmd/root.go`](../cmd/root.go)).

## Global

- **Root use string**: `gh-wm` (binary name).
- **Short description**: GitHub Workflow Manager — run gh-aw-style task markdown in CI.

---

## `init`

**Purpose:** Create `.wm/` layout, starter tasks, optional `CLAUDE.md`, and generate `.github/workflows/wm-agent.yml`.

**Usage:** `gh wm init`

**Steps (see [`cmd/init.go`](../cmd/init.go)):**

1. Create `.wm/tasks/`.
2. Write embedded `config.yml` and starter tasks ([`internal/templates`](../internal/templates/)).
3. Write `CLAUDE.md` in repo root if missing (from template).
4. Collect schedules from `.wm/tasks` and generate `wm-agent.yml` via [`gen.WriteWMAgent`](../internal/gen/wmagent.go).

**Environment:**

| Variable | Default | Meaning |
|----------|---------|---------|
| `GH_WM_REPO` | `gh-wm/gh-wm` | `owner/repo` for **reusable workflow** `uses:` in generated `wm-agent.yml`. |

---

## `upgrade`

**Purpose:** Regenerate **only** `.github/workflows/wm-agent.yml` from current tasks (schedule union) and `GH_WM_REPO`.

**Usage:** `gh wm upgrade`

---

## `assign`

**Purpose:** Add a label to an issue (local `gh` auth).

**Usage:** `gh wm assign <issue-number>`

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--label` | `agent` | Label name to add |

**Implementation:** [`ghclient.AddIssueLabel`](../internal/ghclient/ghclient.go) via `gh api`.

---

## `resolve`

**Purpose:** Print matching **task names** for a GitHub event (JSON array by default).

**Usage:** `gh wm resolve`

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--repo-root` | `.` | Repository root containing `.wm/` |
| `--event-name` | `$GITHUB_EVENT_NAME` | GitHub event name |
| `--payload` | `$GITHUB_EVENT_PATH` | Path to event JSON file |
| `--json` | `true` | If true, print JSON array; if false, one name per line |

**Requires:** `--payload` or `GITHUB_EVENT_PATH` set.

---

## `run`

**Purpose:** Execute **one** task: load `.wm/tasks/<task>.md`, build prompt, run agent subprocess.

**Usage:** `gh wm run --task <name>`

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--repo-root` | `.` | Repository root |
| `--task` | _(required)_ | Task name (filename without `.md`) |
| `--event-name` | `$GITHUB_EVENT_NAME` | Event name |
| `--payload` | `$GITHUB_EVENT_PATH` | Path to event JSON |

**Behavior:** 45-minute [`context.WithTimeout`](../cmd/run.go) around [`engine.RunTask`](../internal/engine/runner.go). Stdout/stderr from agent printed to stderr.

**Agent invocation ([`internal/engine/agent.go`](../internal/engine/agent.go)):**

| Variable | Meaning |
|----------|---------|
| `WM_AGENT_CMD` | If set, split on whitespace: first token = executable, rest + prompt = args. |
| _(default)_ | `claude -p <prompt>` |

Environment passed to subprocess includes `GITHUB_REPOSITORY` (from env) and `WM_TASK=<task name>`.

**Secrets (CI):** `ANTHROPIC_API_KEY` is expected by reusable workflow for Claude Code; ensure the agent you invoke uses it as required.

---

## `status`

**Purpose:** List open issues with common agent labels (wrapper around `gh`).

**Usage:** `gh wm status`

Runs approximately: `gh issue list --label agent,agent:working,agent:review …` ([`cmd/status.go`](../cmd/status.go)).

---

## `logs`

**Purpose:** Show recent **`wm-agent`** workflow runs (not fully issue-scoped yet).

**Usage:** `gh wm logs <issue-number>`

**Note:** Issue number is validated but listing uses `gh run list --workflow wm-agent.yml` ([`cmd/logs.go`](../cmd/logs.go))—treat as a quick visibility helper.

---

## CI-related environment (summary)

| Variable | Used by |
|----------|---------|
| `GITHUB_EVENT_NAME`, `GITHUB_EVENT_PATH` | `resolve`, `run` when flags omitted |
| `GITHUB_REPOSITORY` | Passed to agent; should be set in Actions |
| `WM_SCHEDULE_CRON` | `resolve` schedule narrowing ([`resolver.go`](../internal/engine/resolver.go)) |
| `WM_AGENT_CMD` | Override agent command ([`agent.go`](../internal/engine/agent.go)) |
| `GH_WM_REPO` | `init`, `upgrade` for reusable workflow owner/repo |
