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

## `add`

**Purpose:** Copy or download a gh-aw-style Markdown file into `.wm/tasks/` (validates YAML frontmatter).

**Usage:** `gh wm add <url-or-path>`

Writes `<cwd>/.wm/tasks/<basename>.md` and prints a reminder to run `gh wm upgrade`. See [`cmd/add.go`](../cmd/add.go).

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

**Purpose:** Execute **one** task: load `.wm/tasks/<task>.md`, optional state labels, run agent, then **`safe-outputs:`** steps on success.

**Usage:** `gh wm run --task <name>`

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--repo-root` | `.` | Repository root |
| `--task` | _(required)_ | Task name (filename without `.md`) |
| `--event-name` | `$GITHUB_EVENT_NAME` | Event name |
| `--payload` | `$GITHUB_EVENT_PATH` | Path to event JSON |

**Timeout:** Uses `timeout-minutes` from task frontmatter (default **45**, max **480**). See [`cmd/run.go`](../cmd/run.go).

**Agent invocation ([`internal/engine/agent.go`](../internal/engine/agent.go)):**

| Variable | Meaning |
|----------|---------|
| `WM_AGENT_CMD` | If set, split on whitespace: first token = executable, rest + prompt = args. Overrides `engine:`. |
| `WM_ENGINE_CODEX_CMD` | When `engine: codex` and `WM_AGENT_CMD` unset: full command prefix before prompt (otherwise runs `codex -p`). |
| _(default)_ | `claude -p <prompt>` when `engine:` is `claude` or empty. |
| `copilot` | **`engine: copilot`** requires `WM_AGENT_CMD` (no default CLI). |

Subprocess env includes `GITHUB_REPOSITORY`, `WM_TASK`, and **`WM_TASK_TOOLS`** when `tools:` is set in the task frontmatter (JSON for structured values).

**Post-agent:** `safe-outputs` keys drive [`internal/output`](../internal/output/) (PR / labels / comment). **`WM_CHECKPOINT=1`** enables loading/posting checkpoint comments ([`internal/engine/runner.go`](../internal/engine/runner.go)).

**Secrets (CI):** `ANTHROPIC_API_KEY` is expected by reusable workflow for Claude Code; ensure the agent you invoke uses it as required.

---

## `status`

**Purpose:** List issues with agent-related labels.

**Usage:** `gh wm status`

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--all` | `false` | Use `gh search issues` across visible repositories instead of `gh issue list` for the current repo |

See [`cmd/status.go`](../cmd/status.go).

---

## `logs`

**Purpose:** List **`wm-agent`** workflow runs; prefers runs whose **title contains** `#<issue-number>`.

**Usage:** `gh wm logs <issue-number>`

If none match, prints recent runs with a note. See [`cmd/logs.go`](../cmd/logs.go).

---

## CI-related environment (summary)

| Variable | Used by |
|----------|---------|
| `GITHUB_EVENT_NAME`, `GITHUB_EVENT_PATH` | `resolve`, `run` when flags omitted |
| `GITHUB_REPOSITORY` | Agent + `gh` outputs; required for labels/comments |
| `WM_SCHEDULE_CRON` | `resolve` schedule narrowing ([`resolver.go`](../internal/engine/resolver.go)) |
| `WM_AGENT_CMD` | Override agent command ([`agent.go`](../internal/engine/agent.go)) |
| `WM_ENGINE_CODEX_CMD` | Codex CLI prefix when `engine: codex` |
| `WM_TASK_TOOLS` | Set automatically from `tools:` (read by agent) |
| `WM_CHECKPOINT` | Set to `1` to enable checkpoint load/post |
| `GH_WM_REPO` | `init`, `upgrade` for reusable workflow owner/repo |
