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
- Files named `*.md.disabled` are skipped ([`LoadTasksDir`](../internal/config/frontmatter.go)).

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

Loaded by [`config.Load`](../internal/config/config.go). Struct: [`GlobalConfig`](../internal/config/types.go).

| Field | Purpose |
|-------|---------|
| `version` | Schema version (conventionally `1`). |
| `engine` | Default agent backend name (e.g. `claude`); task can override with `engine:`. |
| `model` | Reserved for agent configuration (not consumed by `runAgent` today). |
| `max_turns` | Reserved (defaulted in [`DefaultGlobal`](../internal/config/config.go)). |
| `context.files` | Paths **relative to repo root** read and **appended** to the prompt ([`engine/agent.go`](../internal/engine/agent.go)). |
| `pr.draft`, `pr.reviewers` | Intended for future PR automation; not applied by the runner today. |

Starter template: [`internal/templates/data/config.yml`](../internal/templates/data/config.yml).

## `on:` block — what gh-wm implements

Matching is implemented in [`internal/trigger/match.go`](../internal/trigger/match.go) as **OR across keys**: if **any** supported block matches the incoming event, the task matches.

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

In frontmatter, `on.schedule` is a **string** (see [`Task.ScheduleString`](../internal/config/types.go)). Recognized aliases when normalizing for `wm-agent.yml` and cron comparison:

| Value | Normalized cron (typical) |
|-------|---------------------------|
| `daily` | `0 0 * * *` |
| `weekly` | `0 0 * * 0` |
| `hourly` | `0 * * * *` |
| other | used as-is (must be valid cron for GitHub Actions) |

## gh-aw fields — expectations

| Field | In gh-wm today |
|-------|------------------|
| `on:` | **Used** for matching (see table above). |
| `description:` | Stored in frontmatter; useful for humans/tools. |
| `engine:` | Read for future use; default agent still `claude -p` unless `WM_AGENT_CMD` is set. |
| `safe-outputs:` | **Not enforced**. Treat as **hints** for authors and agents. |
| `timeout-minutes:` | **Not wired** to `run` timeout (CLI uses a fixed long timeout). |
| `permissions:`, `network:`, `imports:` | Not interpreted; consider documenting in your task if migrating from gh-aw. |

## `wm:` extension (gh-wm–specific)

Parsed as YAML into frontmatter; intended for options **ignored by gh-aw**, e.g.:

```yaml
wm:
  state_labels:
    working: "agent:working"
    done: "agent:review"
    failed: "agent:failed"
```

[`WMExtension`](../internal/config/types.go) defines `state_labels`. **Applying** these labels during runs is not implemented in the engine yet—this is reserved for stateful workflows.

## Checkpoint comments (optional future use)

Format supported by [`checkpoint.Encode`](../internal/checkpoint/checkpoint.go):

```html
<!-- wm-checkpoint: {"branch":"…","sha":"…","step":"…",…} -->
```

The engine does not automatically read or write these yet; integrations can use the package directly.
