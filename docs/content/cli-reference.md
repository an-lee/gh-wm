# CLI reference

The extension is invoked as **`gh wm <subcommand>`** when installed via `gh extension install`. Building from source produces a binary named **`gh-wm`** (same commands as the root Cobra app in [`cmd/root.go`](../../cmd/root.go)).

## Global

- **Root use string**: `gh-wm` (binary name).
- **Short description**: GitHub Workflow Manager — run gh-aw-style task markdown in CI.

---

## `init`

**Purpose:** Create `.wm/` layout, starter tasks, optional `CLAUDE.md`, and generate `.github/workflows/wm-agent.yml`.

**Usage:** `gh wm init`

**Steps (see [`cmd/init.go`](../../cmd/init.go)):**

1. Create `.wm/tasks/`.
2. Write embedded `config.yml` and starter tasks ([`internal/templates`](../../internal/templates/)).
3. Write `CLAUDE.md` in repo root if missing (from template).
4. Collect schedules from `.wm/tasks` and generate `wm-agent.yml` via [`gen.WriteWMAgent`](../../internal/gen/wmagent.go), including **`workflow.runs_on`** from `.wm/config.yml` (default `ubuntu-latest` if unset).

**`workflow.pre_steps` (optional):** A list of GitHub Actions job steps (`name`, `uses`, `run`, `with`, `env`, `if`) run **after** checkout and **before** installing `gh-wm` and running the task. Use this for toolchains (e.g. [`jdx/mise-action`](https://github.com/jdx/mise-action)), dependency installs, or installing the agent CLI. When **`pre_steps` is non-empty**, the generated `wm-agent.yml` uses an **inline** `run` job (steps embedded in the file) instead of calling the reusable [`agent-run.yml`](../../.github/workflows/agent-run.yml) workflow, because reusable workflows cannot accept arbitrary step YAML as inputs.

**Environment:**

| Variable     | Default        | Meaning                                                                     |
| ------------ | -------------- | --------------------------------------------------------------------------- |
| `GH_WM_REPO` | `an-lee/gh-wm` | `owner/repo` for **reusable workflow** `uses:` in generated `wm-agent.yml`. |

---

## `upgrade`

**Purpose:** Run **`gh extension upgrade an-lee/gh-wm`** (best-effort: if it fails, e.g. the CLI was not installed as a `gh` extension, a message is printed and the command continues), then regenerate `.github/workflows/wm-agent.yml` from current tasks (schedule union), **`workflow.runs_on`** and **`workflow.pre_steps`** in `.wm/config.yml` (when present), and `GH_WM_REPO`.

**Usage:** `gh wm upgrade`

If `.wm/config.yml` is missing, runner labels default to **`ubuntu-latest`** when generating `wm-agent.yml`. **`workflow.pre_steps`** follows the same rules as under **`init`** above.

---

## `update`

**Purpose:** Re-download task files using each task’s **`source:`** frontmatter (same idea as **`gh aw update`** for workflows with a source). **`source:`** may be an **https URL** or an **`owner/repo/path`** shorthand (path under the repo on **`main`**, e.g. `owner/repo/workflows/task.md`).

**Usage:**

- `gh wm update` — update every `.wm/tasks/*.md` that has a non-empty `source:` field.
- `gh wm update <task-name> …` — update only the named tasks (filename without `.md`, or with `.md`).

Tasks created with **`gh wm add`** (URL, **`owner/repo/task`** shorthand, or path) get a **`source:`** when appropriate so **`gh wm update`** can re-fetch. After updating tasks, run **`gh wm upgrade`** to refresh `wm-agent.yml` if schedules or other generator inputs changed.

See [`cmd/update.go`](../../cmd/update.go).

---

## `assign`

**Purpose:** Add a label to an issue (local `gh` auth).

**Usage:** `gh wm assign <issue-number>`

**Flags:**

| Flag      | Default | Description       |
| --------- | ------- | ----------------- |
| `--label` | `agent` | Label name to add |

**Implementation:** [`ghclient.AddIssueLabel`](../../internal/ghclient/ghclient.go) via `gh api`.

---

## `add`

**Purpose:** Copy or download a gh-aw–compatible Markdown file into `.wm/tasks/` (validates YAML frontmatter).

**Usage:** `gh wm add <owner/repo/task | url | path>`

- **`owner/repo/task-name`** — Fetches from the default branch (**`main`**), trying **`workflows/<task>.md`** first (gh aw layout), then **`.wm/tasks/<task>.md`**. Records **`source:`** as **`owner/repo/workflows/…`** or **`owner/repo/.wm/tasks/…`** (gh aw-style shorthand), not a raw URL.
- **`https://…` or `http://…`** — Downloads the file; **`source:`** is the same URL (unless already set in the file).
- **Local path** — Copies the file; no **`source:`** is injected unless the file already has one.

Writes `<cwd>/.wm/tasks/<basename>.md` and prints a reminder to run **`gh wm upgrade`**. See [`cmd/add.go`](../../cmd/add.go) and [`cmd/github.go`](../../cmd/github.go).

---

## `resolve`

**Purpose:** Print matching **task names** for a GitHub event (JSON array by default).

**Usage:** `gh wm resolve`

**Flags:**

| Flag           | Default              | Description                                            |
| -------------- | -------------------- | ------------------------------------------------------ |
| `--repo-root`  | `.`                  | Repository root containing `.wm/`                      |
| `--event-name` | `$GITHUB_EVENT_NAME` | GitHub event name                                      |
| `--payload`    | `$GITHUB_EVENT_PATH` | Path to event JSON file; if `--payload` and `GITHUB_EVENT_PATH` are both unset, payload defaults to `{}` |
| `--json`       | `true`               | If true, print JSON array; if false, one name per line |

---

## `run`

**Purpose:** Execute **one** task: load `.wm/tasks/<task>.md`, optional state labels, run agent, then **`safe-outputs:`** steps on success.

**Usage:** `gh wm run --task <name>`

**Git working tree:** Before running, `gh wm` requires a **clean** repository at `--repo-root`: `git status --porcelain` must be empty (no modified, staged, or untracked files). CI checkouts from `actions/checkout` usually satisfy this. Use **`--allow-dirty`** to skip the check (e.g. local scripts or tests).

**Flags:**

| Flag             | Default              | Description                        |
| ---------------- | -------------------- | ---------------------------------- |
| `--repo-root`    | `.`                  | Repository root                    |
| `--task`         | _(required)_         | Task name (filename without `.md`) |
| `--event-name`   | `$GITHUB_EVENT_NAME` | Event name                         |
| `--payload`      | `$GITHUB_EVENT_PATH` | Path to event JSON; if `--payload` and `GITHUB_EVENT_PATH` are both unset, payload defaults to `{}` |
| `--allow-dirty`  | `false`              | Skip the git clean working tree check |

**Timeout:** Uses `timeout-minutes` from task frontmatter (default **45**, max **480**). See [`cmd/run.go`](../../cmd/run.go).

**Output:** Agent subprocess **stdout and stderr are streamed to stderr** as they are produced (full transcript is still captured for `safe-outputs` and checkpoints). After the run, a short **summary line** is printed to stderr (task name, repo path, duration, exit code, success). If the run fails, stderr also indicates whether failure was in the **agent** phase or **`safe-outputs`** (post-agent) phase.

**Agent invocation ([`internal/engine/agent.go`](../../internal/engine/agent.go)):**

| Variable              | Meaning                                                                                                       |
| --------------------- | ------------------------------------------------------------------------------------------------------------- |
| `WM_AGENT_CMD`        | If set, split on whitespace: first token = executable, rest + prompt = args. Overrides `engine:`.             |
| `WM_ENGINE_CODEX_CMD` | When `engine: codex` and `WM_AGENT_CMD` unset: full command prefix before prompt (otherwise runs `codex -p`). |
| _(default)_           | `claude -p <prompt>` when `engine:` is `claude` or empty.                                                     |
| `copilot`             | **`engine: copilot`** requires `WM_AGENT_CMD` (no default CLI).                                               |

Subprocess env includes `GITHUB_REPOSITORY`, `WM_TASK`, and **`WM_TASK_TOOLS`** when `tools:` is set in the task frontmatter (JSON for structured values).

**Post-agent:** `safe-outputs` keys drive [`internal/output`](../../internal/output/) (PR / labels / comment). **`WM_CHECKPOINT=1`** enables loading/posting checkpoint comments ([`internal/engine/runner.go`](../../internal/engine/runner.go)).

**Secrets (CI):** `ANTHROPIC_API_KEY` is expected by reusable workflow for Claude Code; ensure the agent you invoke uses it as required.

---

## `status`

**Purpose:** List issues with agent-related labels.

**Usage:** `gh wm status`

**Flags:**

| Flag    | Default | Description                                                                                        |
| ------- | ------- | -------------------------------------------------------------------------------------------------- |
| `--all` | `false` | Use `gh search issues` across visible repositories instead of `gh issue list` for the current repo |

See [`cmd/status.go`](../../cmd/status.go).

---

## `logs`

**Purpose:** List **`wm-agent`** workflow runs; prefers runs whose **title contains** `#<issue-number>`.

**Usage:** `gh wm logs <issue-number>`

If none match, prints recent runs with a note. See [`cmd/logs.go`](../../cmd/logs.go).

---

## CI-related environment (summary)

| Variable                                 | Used by                                                                           |
| ---------------------------------------- | --------------------------------------------------------------------------------- |
| `GITHUB_EVENT_NAME`, `GITHUB_EVENT_PATH` | `resolve`, `run`: event name / payload file when flags omitted; if `GITHUB_EVENT_PATH` unset, payload is `{}` |
| `GITHUB_REPOSITORY`                      | Agent + `gh` outputs; required for labels/comments                                |
| `WM_SCHEDULE_CRON`                       | `resolve` schedule narrowing ([`resolver.go`](../../internal/engine/resolver.go)) |
| `WM_AGENT_CMD`                           | Override agent command ([`agent.go`](../../internal/engine/agent.go))             |
| `WM_ENGINE_CODEX_CMD`                    | Codex CLI prefix when `engine: codex`                                             |
| `WM_TASK_TOOLS`                          | Set automatically from `tools:` (read by agent)                                   |
| `WM_CHECKPOINT`                          | Set to `1` to enable checkpoint load/post                                         |
| `GH_WM_REPO`                             | `init`, `upgrade` for reusable workflow owner/repo                                |
