# Task file format

Tasks live under **`.wm/tasks/`** as `*.md` files. The **filename without `.md`** is the **task name** (e.g. `implement.md` → `implement`).

## Layout

```text
.wm/
  config.yml           # Global defaults
  tasks/
    implement.md
    code-review.md
    …
```

## Frontmatter + body

- **YAML frontmatter** is required: first line `---`, closing `---`, then the **markdown body**.
- The **body** is the **agent prompt** (plus optional files appended per `.wm/config.yml`—see below).
- Files named `*.md.disabled` are skipped ([`LoadTasksDir`](../../internal/config/frontmatter.go)).

### Minimal example

```markdown
---
description: Short summary for humans and status output.

on:
  issues:
    types: [labeled]

engine: claude
---

# Do the work

Instructions for the agent…
```

## `.wm/config.yml` (global)

Loaded by [`config.Load`](../../internal/config/config.go). Struct: [`GlobalConfig`](../../internal/config/types.go). **Machine-readable schema:** [global-schema.json](global-schema.json) (subset; extra keys allowed).

| Field | Purpose |
|-------|---------|
| `version` | Schema version (conventionally `1`). |
| `engine` | Default agent backend name (e.g. `claude`); task can override with `engine:`. |
| `model` | Passed to the default **`claude`** / **`codex`** CLI as **`--model`** when set ([`agentCLIArgs`](../../internal/engine/agent.go)). |
| `max_turns` | Passed as **`--max-turns`** when non-zero (default **100** in [`DefaultGlobal`](../../internal/config/config.go)). |
| `claude_output_format` | Optional: **`text`** (default), **`json`**, or **`stream-json`**. For the built-in **`claude`** invocation only (not **`WM_AGENT_CMD`** / **codex**), selects **`claude -p --output-format`** and the run-dir file (**`agent-stdout.log`** vs **`conversation.json`** / **`conversation.jsonl`**). Overridden by **`WM_CLAUDE_OUTPUT_FORMAT`**. |
| `workflow.runs_on` | YAML list of GitHub Actions runner labels baked into generated `wm-agent.yml` as the reusable workflow `runs_on` input (JSON array). If omitted or empty, defaults to `ubuntu-latest`. Use e.g. `self-hosted` plus OS labels for self-hosted runners. |
| `workflow.install_claude_code` | Optional boolean (default **true**). When **true**, generated workflows run the official **Claude Code** install step and put `~/.local/bin` on `PATH` before `gh-wm run` so the default **`claude`** engine is available on minimal self-hosted runners. Set **`false`** for **codex-only** setups or when **`claude`** is already installed on the runner. |
| `workflow.gh_wm_extension_version` | Optional string. When set, generated CI workflows install **`gh-wm`** with **`gh extension install owner/repo --pin <ref>`** (release tag or commit per **`gh help extension install`**). When unset or empty, the install uses the latest. Use this to pin CI to a tag or commit of the extension repo. |
| `workflow.pre_steps` | Optional list of GitHub Actions job steps (`name`, `uses`, `run`, `with`, `env`, `if`) run before installing `gh-wm` and the task. When non-empty, `wm-agent.yml` embeds an **inline** `run` job instead of calling reusable [`agent-run.yml`](../../.github/workflows/agent-run.yml). See [`cli-reference.md`](cli-reference.md) **`init`**. |
| `context.files` | Paths **relative to repo root** read and **appended** to the prompt ([`engine/agent.go`](../../internal/engine/agent.go)). Omit **`CLAUDE.md`** when using Claude Code: it loads that file from the repo on its own; listing it here duplicates it in the prompt. |
| `pr.draft`, `pr.reviewers` | Defaults merged with `safe-outputs.create-pull-request` for `gh pr create`. |

Starter template: [`internal/templates/data/config.yml`](../../internal/templates/data/config.yml).

**Machine-readable schema:** [task-schema.json](task-schema.json) (subset of interpreted fields; extra keys are allowed for gh-aw compatibility).

## `on:` block — what gh-wm implements

Matching is implemented in [`internal/trigger/match.go`](../../internal/trigger/match.go) as **OR across keys**: if **any** supported block matches the incoming event, the task matches.

GitHub’s `GITHUB_EVENT_NAME` must align with the keys below (e.g. `issues`, not `issue`).

| `on:` key | Expected `GITHUB_EVENT_NAME` | Behavior |
|-----------|------------------------------|----------|
| `issues` | `issues` | Matches `payload.action` against `types:` (e.g. `labeled`, `opened`). Empty `types` → always match. Optional **`labels:`** (list of names): when set, only **`labeled`** actions match, and **`payload.label.name`** must equal one of the listed names (use this to avoid tasks re-firing on unrelated or state-machine labels). |
| `issue_comment` | `issue_comment` | Optionally restricts `types:` (e.g. `created`). |
| `pull_request` | `pull_request` or `pull_request_target` | Matches `payload.action` to `types:` (e.g. `review_requested`). Empty `types` → always match. |
| `slash_command` | `issue_comment` or `pull_request_review_comment` | Body must start with `/name` or `/name …` where `name` comes from `slash_command.name`. |
| `schedule` | `schedule` | At resolve, any task with `on.schedule` matches a schedule event; use `WM_SCHEDULE_CRON` to narrow (see [architecture](architecture.md)). |
| `workflow_dispatch` | `workflow_dispatch` | Presence of key is enough; inputs are not matched per-field yet. |

### Optional `reaction:`

Optional string sibling inside **`on:`** (e.g. `reaction: eyes`). Value must be a GitHub **reaction content** accepted by the API: **`+1`**, **`-1`**, **`laugh`**, **`confused`**, **`heart`**, **`hooray`**, **`rocket`**, **`eyes`**. It does **not** participate in event matching (triggers are still **`issues`** / **`issue_comment`** / … as above).

When **`gh wm run`** executes the task, during **activation** (after context is loaded, **before** the agent subprocess), gh-wm posts that reaction to the **triggering** resource if possible:

- **`issue_comment`** events: reaction on the **`comment`** from the payload when a comment id is present; otherwise no-op.
- Other events with an issue or PR number (e.g. **`issues`**, **`pull_request`**): reaction on that issue/PR.
- No **`GITHUB_REPOSITORY`**, or no applicable issue/comment (e.g. some **`schedule`** runs): skipped silently.

If **`gh api`** fails (including permissions), the error is recorded but the run **continues** (best-effort). Duplicate reactions from the same user are treated as success when the API reports **`already_exists`** or **`Resource already exists`** (GitHub may return either form).

### Generated `wm-agent.yml` triggers

`gh wm init` and `gh wm upgrade` build the workflow **`on:`** block from a **union** over all tasks ([`gen.CollectTriggersFromTasksDir`](../../internal/gen/triggers.go)): **`issues`**, **`issue_comment`**, **`pull_request`**, and **`pull_request_review_comment`** each get a merged **`types:`** list (task-only filters such as **`labels:`** are not copied into the workflow—resolve still enforces them). **`slash_command`** implies **`issue_comment`** with **`types: [created]`** for conversation comments and can also imply **`pull_request_review_comment`** with **`types: [created]`** when configured for PR review comments; **`schedule`** unions normalized crons; **`workflow_dispatch`** is always included for manual runs. Keys with no GitHub Actions workflow equivalent (e.g. **`reaction:`**) are ignored for generation; **`reaction:`** is still applied at run time as described above.

### Schedule strings

In frontmatter, `on.schedule` is a **string** (see [`Task.ScheduleString`](../../internal/config/types.go)). Aliases are expanded with [`gen.FuzzyNormalizeSchedule`](../../internal/gen/schedules.go) (github/gh-aw–compatible: FNV-1a hash of the task file path + weighted time pool) so each task gets a **stable, distinct** cron instead of everyone at midnight:

| Value | Normalized cron |
|-------|-----------------|
| `daily` | One run per day at a deterministic minute/hour (e.g. `43 5 * * *`) |
| `weekly` | One run per week: scattered weekday (0–6) + time from the same pool |
| `weekly on <weekday>` | One run per week on that weekday at a deterministic minute/hour (same UTC time pool as `daily`/`weekly`). Day-of-week follows GitHub cron (0=Sunday … 6=Saturday). Accepts English names (`Monday`, …) and common abbreviations (`mon`, …). Case-insensitive; extra spaces allowed (e.g. `weekly  on  mon`). |
| `hourly` | `M */1 * * *` with scattered minute `M` in 5–54 |
| `every N hours` | For `2 ≤ N ≤ 23`: `M */N * * *` with the same scattered `M` as `hourly`. `every 1 hour` / `every 1 hours` matches `hourly`. `every 24 hours` matches `daily`. Case-insensitive; extra spaces allowed (e.g. `every  3  hours`). |
| other | If it is already a **5-field** cron string, whitespace-normalized and used as-is; otherwise passed through unchanged (must be valid for GitHub Actions if used as cron) |

Out-of-range `every N hours` (`N < 1` or `N > 24`) is passed through unchanged, like an unknown token. A `weekly on …` string with an unrecognized weekday is passed through unchanged.

## `safe-outputs:` — policy + `gh wm emit`

If the task omits **`safe-outputs:`** or the block is **empty**, the post-agent safe-output phase does **nothing**.

If **`safe-outputs:`** contains **at least one key**, record outputs by running **`gh wm emit <subcommand>`** with flags for each follow-up. Each call appends one validated JSON line to **`WM_SAFE_OUTPUT_FILE`** (`output.jsonl` in the [per-run directory](architecture.md#what-persists-where)). The run sets **`WM_REPO_ROOT`**, **`WM_TASK`**, **`WM_SAFE_OUTPUT_FILE`**, **`GITHUB_REPOSITORY`**, and **`WM_ISSUE_NUMBER`** / **`WM_PR_NUMBER`** when applicable. Built-in subcommands **`missing-tool`** and **`missing-data`** are always available.

If there is **no** NDJSON in `output.jsonl`, the safe-output phase **succeeds** with a **warning** (implicit noop). Prefer **`gh wm emit noop --message "…"`** when you want an explicit record.

Keys under **`safe-outputs:`** declare what operations are **allowed**; each item has a **`type`** using **underscores** (gh-aw style): **`create_pull_request`**, **`add_comment`**, **`add_labels`**, **`remove_labels`**, **`create_issue`**, **`update_pull_request`**, **`update_issue`**, **`close_issue`**, **`close_pull_request`**, **`add_reviewer`**, **`create_pull_request_review_comment`**, **`submit_pull_request_review`**, **`reply_to_pull_request_review_comment`**, **`resolve_pull_request_review_thread`**, **`push_to_pull_request_branch`**, **`noop`**, **`missing_tool`**, **`missing_data`**. Dash forms in **`type`** (e.g. `create-pull-request`) are accepted too.

For **issue/PR numbers**, JSON may use **`target`** (preferred) or gh-aw-style aliases: **`issue_number`**, **`pull_request_number`**, **`item_number`**. The first **strictly positive** value wins in a fixed order starting with **`target`**.

**`update_issue` / `update_pull_request` — `operation`:** Optional **`operation`** on the item: **`replace`** (default), **`append`**, **`prepend`**, **`replace-island`** (hyphen or underscore accepted). **`replace`** sets the body to the supplied **`body`** string. **`append`** / **`prepend`** load the current title/body from the API, then concatenate. **`replace-island`** replaces the region between **`<!-- gh-wm:island -->`** and **`<!-- /gh-wm:island -->`** in the existing body with **`body`** (markers must be present and in order).

**`noop`:** In addition to **`{"type":"noop","message":"…"}`**, a gh-aw-style envelope **`{"noop":{"message":"…"}}`** without a top-level **`type`** is accepted and treated as **`noop`**.

**`push_to_pull_request_branch`** — allowed when **`push-to-pull-request-branch`** is set under **`safe-outputs:`**. At execution time the runner resolves the PR, checks optional **`title-prefix`** and required **`labels`** (from frontmatter) against the PR, requires the **current git branch** to equal the PR’s **head** branch, then runs **`git push -u origin HEAD`** in **`WM_REPO_ROOT`**. Same-repo only; no cross-repo routing.

Optional **`messages:`** (sibling of output keys under **`safe-outputs:`**) configures status comments and optional footer text. Supported keys are **`run-started`**, **`run-success`**, **`run-failure`**, and **`footer`**.

```json
{
  "items": [
    { "type": "add_comment", "body": "Done." },
    { "type": "noop", "message": "No PR needed." }
  ]
}
```

- A **`type`** is **rejected** (skipped with a log line) if its corresponding **`safe-outputs:`** key is **not** declared (except **`noop`**, which is always allowed).
- **`max:`** per handler is **enforced** (defaults apply when omitted: e.g. **1** for PR / comment / issue / update / close-issue / submit-review, **10** for **`close_pull_request`** and **`reply_to_pull_request_review_comment`**, **5** for **`create_pull_request_review_comment`** and **`resolve_pull_request_review_thread`**, **3** for label lists and **`add_reviewer`**).
- **`title-prefix`**: enforced for **`create_pull_request`**, **`create_issue`**, **`update_pull_request`**, and **`update_issue`** titles when a non-empty title is supplied (prefix applied when missing).
- **`labels`** under **`create-pull-request`** / **`create_issue`**: merged with agent-supplied labels (deduped).
- **`add-labels`** / **`remove-labels`**: optional **`allowed:`** and **`blocked:`** (glob patterns); **`blocked`** is evaluated first.

The agent prompt includes an **Available Outputs** section whenever `safe-outputs:` is non-empty ([`internal/output/prompt.go`](../../internal/output/prompt.go)).

## Other frontmatter fields

| Field | In gh-wm |
|-------|----------|
| `on:` | **Used** for matching (see table above). |
| `source:` | Optional upstream reference: an **https URL** or **`owner/repo/path`** to the file on **`main`** (e.g. **`owner/repo/workflows/task.md`**, same style as gh aw). Set when adding via **`gh wm add`** (URL or **`owner/repo/task`** shorthand); **`gh wm update`** resolves it and re-fetches the file. |
| `description:` | Stored in frontmatter; useful for humans/tools. |
| `engine:` | Selects backend when `WM_AGENT_CMD` is unset: `claude` (default), `codex` (`codex -p` or `WM_ENGINE_CODEX_CMD`). The former `copilot` engine name is **removed**; use **`WM_AGENT_CMD`** for a custom CLI. |
| `timeout-minutes:` | **Used** by [`cmd/run`](../../cmd/run.go) for the context timeout (capped). |
| `tools:` | Serialized to env **`WM_TASK_TOOLS`** for the agent subprocess (JSON if structured). |
| `permissions:`, `network:`, `imports:` | Not interpreted. |

### Migration: removed `wm.state_labels`

Earlier versions supported optional **`wm.state_labels`** (working / done / failed) for automatic issue labels during runs. That feature was **removed**—label-based run state proved fragile. Use **`on.issues.labels`** (or slash commands / schedule) for precise triggers, and **`gh wm emit`** (e.g. **`add-labels`**) when the agent should change labels explicitly.

## Checkpoint comments (optional)

Set **`WM_CHECKPOINT=1`** to:

1. Load the latest `<!-- wm-checkpoint: … -->` from issue comments into the prompt ([`checkpoint.ParseLatest`](../../internal/checkpoint/checkpoint.go)).
2. After a successful run, post a new checkpoint comment with the latest agent summary.

Format is defined in [`checkpoint.Encode`](../../internal/checkpoint/checkpoint.go).
