# CLI reference

The extension is invoked as **`gh wm <subcommand>`** when installed via `gh extension install`. Building from source produces a binary named **`gh-wm`** (same commands as the root Cobra app in [`cmd/root.go`](../../cmd/root.go)).

## Global

- **Root use string**: `gh-wm` (binary name).
- **Short description**: GitHub Workflow Manager — run gh-aw-style task markdown in CI.

---

## `version`

**Purpose:** Print the installed `gh-wm` version (and optional Git commit line for release builds).

**Usage:**

- `gh wm version` — prints `gh-wm <version>` on one line; if the binary was built with a commit SHA (release workflow), prints `commit: <short-sha>` on a second line.
- `gh wm --version` / `gh wm -v` — same version string via the root Cobra flag.

Local and CI builds without linker flags report **`dev`**. Release assets built from a git tag embed the tag (without the leading `v`) and a short commit hash. See [`cmd/version.go`](../../cmd/version.go) and [`.github/workflows/release.yml`](../../.github/workflows/release.yml).

---

## `init`

**Purpose:** Create `.wm/` layout, starter tasks, and generate `.github/workflows/wm-agent.yml`.

**Usage:** `gh wm init`

**Steps (see [`cmd/init.go`](../../cmd/init.go)):**

1. Create `.wm/tasks/`.
2. Write embedded `config.yml` and starter tasks ([`internal/templates`](../../internal/templates/)).
3. Collect schedules from `.wm/tasks` and generate `wm-agent.yml` via [`gen.WriteWMAgent`](../../internal/gen/wmagent.go), including **`workflow.runs_on`** from `.wm/config.yml` (default `ubuntu-latest` if unset).
4. Ensure **`.wm/.gitignore`** contains **`runs/`** (per-run artifact dirs from `gh wm run`) — creates or appends that file when needed.

**`workflow.pre_steps` (optional):** A list of GitHub Actions job steps (`name`, `uses`, `run`, `with`, `env`, `if`) run **after** checkout and **before** installing `gh-wm` and running the task. Use this for toolchains (e.g. [`jdx/mise-action`](https://github.com/jdx/mise-action)), dependency installs, or installing the agent CLI. When **`pre_steps` is non-empty**, the generated `wm-agent.yml` uses an **inline** `run` job (steps embedded in the file) instead of calling the reusable [`agent-run.yml`](../../.github/workflows/agent-run.yml) workflow, because reusable workflows cannot accept arbitrary step YAML as inputs.

**Environment:**

| Variable     | Default        | Meaning                                                                     |
| ------------ | -------------- | --------------------------------------------------------------------------- |
| `GH_WM_REPO` | `an-lee/gh-wm` | `owner/repo` for **reusable workflow** `uses:` in generated `wm-agent.yml`. |

---

## `upgrade`

**Purpose:** Run **`gh extension upgrade an-lee/gh-wm`** (best-effort: if it fails, e.g. the CLI was not installed as a `gh` extension, a message is printed and the command continues), then regenerate `.github/workflows/wm-agent.yml` from current tasks (schedule union), **`workflow.runs_on`** and **`workflow.pre_steps`** in `.wm/config.yml` (when present), and `GH_WM_REPO`.

**Usage:** `gh wm upgrade`

If `.wm/config.yml` is missing, runner labels default to **`ubuntu-latest`** when generating `wm-agent.yml`. **`workflow.pre_steps`** follows the same rules as under **`init`** above. **`upgrade`** also ensures **`.wm/.gitignore`** lists **`runs/`** (same as step 5 under **`init`**), so older repos pick up the ignore rule without re-running **`init`**.

---

## `update`

**Purpose:** Re-download task files using each task’s **`source:`** frontmatter (same idea as **`gh aw update`** for workflows with a source). **`source:`** may be an **https URL** or an **`owner/repo/path`** shorthand (path under the repo on **`main`**, e.g. `owner/repo/workflows/task.md`).

**Usage:**

- `gh wm update` — update every `.wm/tasks/*.md` that has a non-empty `source:` field.
- `gh wm update <task-name> …` — update only the named tasks (filename without `.md`, or with `.md`).

Tasks created with **`gh wm add`** (URL, **`owner/repo/task`** shorthand, or path) get a **`source:`** when appropriate so **`gh wm update`** can re-fetch. **`gh wm add`** runs **`gh wm upgrade`** automatically after a successful write. After **`gh wm update`**, run **`gh wm upgrade`** to refresh `wm-agent.yml` if schedules or other generator inputs changed.

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

Writes `<cwd>/.wm/tasks/<basename>.md`, then runs **`gh wm upgrade`** (same as the **`upgrade`** command: best-effort extension self-upgrade and regenerate **`wm-agent.yml`**). See [`cmd/add.go`](../../cmd/add.go) and [`cmd/github.go`](../../cmd/github.go).

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
| `--force-task` | _(unset)_            | Pin a single task by name; skips event/`on:` matching (same idea as local `run` picking a task). Used for manual runs and CI when the resolve job should return exactly one task. |

---

## `run`

**Purpose:** Execute **one** task: **activation** (validate event/engine, optional working label, branch prep for PR mode), **agent**, **validation** (exit + output size), **`safe-outputs:`** steps, then **conclusion** (done/failed labels, checkpoint, branch rollback on failure).

**Usage:** `gh wm run --task <name>` (local agent), or `gh wm run --task <name> --remote` to dispatch the **`wm-agent`** workflow on GitHub.

**Git working tree:** For **local** runs (default), `gh wm` requires a **clean** repository at `--repo-root`: `git status --porcelain` must be empty (no modified, staged, or untracked files). CI checkouts from `actions/checkout` usually satisfy this. Use **`--allow-dirty`** to skip the check (e.g. local scripts or tests). **`--remote`** does not require a clean tree (it does not run the agent locally).

**Flags:**

| Flag             | Default              | Description                        |
| ---------------- | -------------------- | ---------------------------------- |
| `--repo-root`    | `.`                  | Repository root                    |
| `--task`         | _(required)_         | Task name (filename without `.md`) |
| `--event-name`   | `$GITHUB_EVENT_NAME` | Event name (local run only)        |
| `--payload`      | `$GITHUB_EVENT_PATH` | Path to event JSON; if `--payload` and `GITHUB_EVENT_PATH` are both unset, payload defaults to `{}` (local run only) |
| `--allow-dirty`  | `false`              | Skip the git clean working tree check (local run only) |
| `--remote`       | `false`              | Run **`gh workflow run`** to trigger **`workflow_dispatch`** on the repo’s **`wm-agent.yml`** with **`-f task_name=<task>`**. Requires the **`gh`** CLI and auth. Repository defaults to **`gh repo view`**; override with **`--repo OWNER/NAME`**. Optional **`--workflow`** (default **`wm-agent.yml`**), **`--ref`** (git ref for the workflow run), and **`--issue`** (passed as **`-f issue_number=`** for the dispatch inputs). After upgrading **`gh-wm`**, run **`gh wm upgrade`** in the target repo so the generated workflow declares the **`task_name`** input; older **`wm-agent.yml`** files may reject unknown **`-f`** fields. **`--remote`** does not send a custom GitHub event payload (the run on Actions sees a normal **`workflow_dispatch`** event, optionally with **`issue_number`**). |

**Timeout:** Uses `timeout-minutes` from task frontmatter (default **45**, max **480**) for **local** runs only. See [`cmd/run.go`](../../cmd/run.go).

**Output:** Before the agent starts, stderr prints a short **banner** (task name, repo path, current git branch, engine). Agent subprocess **stdout and stderr are streamed to stderr** as they are produced. Each run also writes a **per-run artifact directory** under **`.wm/runs/<id>/`** (or **`WM_RUN_DIR/<id>/`** when set): `prompt.md`, combined agent log (default **`agent-stdout.log`**, or **`conversation.json`** / **`conversation.jsonl`** when structured Claude print-mode output is configured — see **`claude_output_format`** / **`WM_CLAUDE_OUTPUT_FORMAT`** in [task-format.md](task-format.md)), `meta.json`, `result.json` (see [architecture.md](architecture.md#what-persists-where)). The summary block on stderr includes a line **`artifacts=<path>`** when that directory was created. After the run, a short **summary line** is printed to stderr (task name, repo path, duration, exit code, success, **`phase=`** — `activation`, `agent`, `validation`, `safe-outputs`, or last phase reached). If the run fails, stderr also prints **`failure phase:`** (for `safe-outputs`, the message still says **safe-outputs (post-agent)**; otherwise the failing **phase** name).

**Branch + PR (`safe-outputs: create-pull-request`):** If the task lists **`create-pull-request`** under `safe-outputs` and the repo is on the **default branch** (or detached `HEAD`), [`internal/gitbranch`](../../internal/gitbranch/) creates and checks out **`wm/<task-slug>-<UTC-timestamp>`** before the agent runs so commits are not on `main`. If you are **already on a non-default branch**, no branch is created. On **agent failure** after a branch was created, the runner checks out the previous branch when possible (skipped when the previous state was detached `HEAD`). After a successful agent run, if **`output.json`** includes a **`create_pull_request`** item (and policy allows it), the safe-output step runs **`git push`** then **`gh pr create --base <default>`** when there are commits ahead of the remote default branch; it **skips** if the current branch is still the default, if **`gh pr list`** already shows a PR for the current head, or if there is nothing to push.

**Agent invocation ([`internal/engine/agent.go`](../../internal/engine/agent.go)):**

| Variable              | Meaning                                                                                                       |
| --------------------- | ------------------------------------------------------------------------------------------------------------- |
| `WM_AGENT_CMD`        | If set: split on whitespace for argv; **prompt** is appended as the last argument unless the string contains **`{prompt}`**, in which case that placeholder is replaced by the prompt (single argument). Overrides `engine:`. |
| `WM_ENGINE_CODEX_CMD` | When `engine: codex` and `WM_AGENT_CMD` unset: same `{prompt}` rule as above (otherwise prompt is appended). Default **`codex`** invocation mirrors **claude** (stdin prompt, `--dangerously-skip-permissions`, `--model` / `--max-turns` from config when set). |
| _(default)_           | **`claude -p --dangerously-skip-permissions`** with the prompt on **stdin**; **`--model`** and **`--max-turns`** come from [`.wm/config.yml`](task-format.md) when set. Optional **`--output-format json`** or **`stream-json`** when **`claude_output_format`** / **`WM_CLAUDE_OUTPUT_FORMAT`** is set (built-in **`claude`** only). |
| `copilot`             | **`engine: copilot`** requires `WM_AGENT_CMD` (no default CLI).                                               |

The default **`claude`** invocation uses **`--dangerously-skip-permissions`** so non-interactive runs can use tools (file edits, **`gh`**, git). Subprocess **env** is the parent environment (`GITHUB_TOKEN` in Actions, `gh auth` locally) plus `GITHUB_REPOSITORY`, `WM_TASK`, **`WM_OUTPUT_FILE`** (path to per-run **`output.json`** when a run directory exists), and **`WM_TASK_TOOLS`** when `tools:` is set in the task frontmatter (JSON for structured values).

**Post-agent:** When **`safe-outputs:`** has at least one key, the agent **must** write non-empty **`items`** JSON to **`WM_OUTPUT_FILE`** ([`internal/output`](../../internal/output/)); missing or empty output fails that phase (use **`noop`** if no GitHub actions are needed). **`WM_CHECKPOINT=1`** enables loading/posting checkpoint comments ([`internal/engine/runner.go`](../../internal/engine/runner.go)).

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
| `WM_OUTPUT_FILE`                         | Set by `run` when a per-run dir exists: path where the agent may write **`output.json`** (`items` → safe outputs). |
| `WM_CHECKPOINT`                          | Set to `1` to enable checkpoint load/post                                         |
| `WM_RUN_DIR`                             | If set, per-run artifacts are written under **`<WM_RUN_DIR>/<run-id>/`** instead of **`<repo>/.wm/runs/<run-id>/`** (useful for CI artifact upload paths). |
| `WM_CLAUDE_OUTPUT_FORMAT`                | Overrides **`claude_output_format`** in [`.wm/config.yml`](task-format.md): **`text`** (default), **`json`**, or **`stream-json`** for built-in **`claude`** (run-dir filename and **`--output-format`**). |
| `GH_WM_REPO`                             | `init`, `upgrade` for reusable workflow owner/repo                                |
