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

Loaded by [`config.Load`](../../internal/config/config.go). Struct: [`GlobalConfig`](../../internal/config/types.go).

| Field | Purpose |
|-------|---------|
| `version` | Schema version (conventionally `1`). |
| `engine` | Default agent backend name (e.g. `claude`); task can override with `engine:`. |
| `model` | Reserved for agent configuration (not consumed by `runAgent` today). |
| `max_turns` | Reserved (defaulted in [`DefaultGlobal`](../../internal/config/config.go)). |
| `workflow.runs_on` | YAML list of GitHub Actions runner labels baked into generated `wm-agent.yml` as the reusable workflow `runs_on` input (JSON array). If omitted or empty, defaults to `ubuntu-latest`. Use e.g. `self-hosted` plus OS labels for self-hosted runners. |
| `workflow.pre_steps` | Optional list of GitHub Actions job steps (`name`, `uses`, `run`, `with`, `env`, `if`) run before installing `gh-wm` and the task. When non-empty, `wm-agent.yml` embeds an **inline** `run` job instead of calling reusable [`agent-run.yml`](../../.github/workflows/agent-run.yml). See [`cli-reference.md`](cli-reference.md) **`init`**. |
| `context.files` | Paths **relative to repo root** read and **appended** to the prompt ([`engine/agent.go`](../../internal/engine/agent.go)). |
| `pr.draft`, `pr.reviewers` | Defaults merged with `safe-outputs.create-pull-request` for `gh pr create`. |

Starter template: [`internal/templates/data/config.yml`](../../internal/templates/data/config.yml).

## `on:` block — what gh-wm implements

Matching is implemented in [`internal/trigger/match.go`](../../internal/trigger/match.go) as **OR across keys**: if **any** supported block matches the incoming event, the task matches.

GitHub’s `GITHUB_EVENT_NAME` must align with the keys below (e.g. `issues`, not `issue`).

| `on:` key | Expected `GITHUB_EVENT_NAME` | Behavior |
|-----------|------------------------------|----------|
| `issues` | `issues` | Matches `payload.action` against `types:` (e.g. `labeled`, `opened`). Empty `types` → always match. |
| `issue_comment` | `issue_comment` | Optionally restricts `types:` (e.g. `created`). |
| `pull_request` | `pull_request` or `pull_request_target` | Matches `payload.action` to `types:` (e.g. `review_requested`). Empty `types` → always match. |
| `slash_command` | `issue_comment` | Body must start with `/name` or `/name …` where `name` comes from `slash_command.name`. |
| `schedule` | `schedule` | At resolve, any task with `on.schedule` matches a schedule event; use `WM_SCHEDULE_CRON` to narrow (see [architecture](architecture.md)). |
| `workflow_dispatch` | `workflow_dispatch` | Presence of key is enough; inputs are not matched per-field yet. |

### Schedule strings

In frontmatter, `on.schedule` is a **string** (see [`Task.ScheduleString`](../../internal/config/types.go)). Recognized aliases when normalizing for `wm-agent.yml` and cron comparison:

| Value | Normalized cron (typical) |
|-------|---------------------------|
| `daily` | `0 0 * * *` |
| `weekly` | `0 0 * * 0` |
| `hourly` | `0 * * * *` |
| other | used as-is (must be valid cron for GitHub Actions) |

## `safe-outputs:` — implemented vs hints

Keys under `safe-outputs:` select **post-agent** behavior in [`internal/output`](../../internal/output/). **Limits** (`max:`, etc.) are **not enforced**; they are gh-aw-compatible hints only.

| Key | When present | Behavior |
|-----|----------------|----------|
| `create-pull-request` | Yes | After success, if `git` has commits ahead of `origin/<default-branch>`, `git push` + `gh pr create` (draft/labels from block + global `pr.*`). |
| `add-labels` | Yes | After success, adds each name in `add-labels.labels` via `gh api`. |
| `add-comment` | Yes | After success, posts agent stdout/summary as `gh issue comment` or `gh pr comment`. |

**Order of execution:** create-pull-request → add-labels → add-comment.

Other gh-aw keys (`create-issue`, `create-discussion`, …) are **not** implemented as automated outputs yet.

## Other frontmatter fields

| Field | In gh-wm |
|-------|----------|
| `on:` | **Used** for matching (see table above). |
| `source:` | Optional upstream reference: an **https URL** or **`owner/repo/path`** to the file on **`main`** (e.g. **`owner/repo/workflows/task.md`**, same style as gh aw). Set when adding via **`gh wm add`** (URL or **`owner/repo/task`** shorthand); **`gh wm update`** resolves it and re-fetches the file. |
| `description:` | Stored in frontmatter; useful for humans/tools. |
| `engine:` | Selects backend when `WM_AGENT_CMD` is unset: `claude` (default), `codex` (`codex -p` or `WM_ENGINE_CODEX_CMD`), `copilot` requires `WM_AGENT_CMD`. |
| `timeout-minutes:` | **Used** by [`cmd/run`](../../cmd/run.go) for the context timeout (capped). |
| `tools:` | Serialized to env **`WM_TASK_TOOLS`** for the agent subprocess (JSON if structured). |
| `permissions:`, `network:`, `imports:` | Not interpreted. |

## `wm:` extension (gh-wm–specific)

```yaml
wm:
  state_labels:
    working: "agent:working"
    done: "agent:review"
    failed: "agent:failed"
```

If set, [`engine/state.go`](../../internal/engine/state.go) adds/removes these labels around the run (requires `GITHUB_REPOSITORY` and an issue/PR number in the event).

## Checkpoint comments (optional)

Set **`WM_CHECKPOINT=1`** to:

1. Load the latest `<!-- wm-checkpoint: … -->` from issue comments into the prompt ([`checkpoint.ParseLatest`](../../internal/checkpoint/checkpoint.go)).
2. After a successful run, post a new checkpoint comment with the latest agent summary.

Format is defined in [`checkpoint.Encode`](../../internal/checkpoint/checkpoint.go).
