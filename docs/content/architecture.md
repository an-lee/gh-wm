# Architecture

## Goals (design intent)

- **gh-aw format compatibility**: Task files use Markdown + YAML frontmatter like [Agentic Workflows (gh-aw)](https://github.github.io/gh-aw/); you can drop community workflows into `.wm/tasks/`.
- **No compile step**: No `.lock.yml`, no `gh aw compile`.
- **Go + `go-gh`**: GitHub auth follows `gh auth login` (see [`internal/ghclient`](../../internal/ghclient/) for API usage from commands like `assign`).
- **Thin coordination on GitHub**: Issues, labels, Actions, PRs—no extra control plane.

## High-level pipeline

Each task follows **trigger → resolve → run agent → optional outputs**. After a **successful** agent run, [`internal/output`](../../internal/output/) runs steps implied by `safe-outputs:` keys (hints only; not enforced like gh-aw). If `wm.state_labels` is set, [`internal/engine/state.go`](../../internal/engine/state.go) updates labels around the run.

```mermaid
flowchart LR
  subgraph triggers [Triggers from on]
    Issues[issues types]
    IC[issue_comment]
    PR[pull_request]
    Slash[slash_command]
    Sched[schedule]
    WD[workflow_dispatch]
  end

  subgraph engine [Go engine]
    Resolve[ResolveMatchingTasks]
    Run[RunTask]
    Out[RunSuccessOutputs]
  end

  subgraph agent [Agent subprocess]
    Claude[claude -p or WM_AGENT_CMD]
  end

  Issues --> Resolve
  IC --> Resolve
  PR --> Resolve
  Slash --> Resolve
  Sched --> Resolve
  WD --> Resolve
  Resolve --> Run
  Run --> Claude
  Claude --> Out
```

Optional **checkpoints** ([`internal/checkpoint`](../../internal/checkpoint/checkpoint.go)): when `WM_CHECKPOINT=1`, the runner loads the latest checkpoint from issue comments into the prompt before the agent, and posts a new checkpoint comment after a successful run (before/after also tied to outputs and state labels—see [`internal/engine/runner.go`](../../internal/engine/runner.go)).

## Code map

| Concern | Location | Role |
|---------|----------|------|
| CLI entry | [`cmd/`](../../cmd/) | Cobra commands: `init`, `upgrade`, `assign`, `resolve`, `run`, `status`, `logs`, `add`. |
| Config + tasks | [`internal/config/`](../../internal/config/) | Load `.wm/config.yml`, parse `.wm/tasks/*.md` frontmatter ([`frontmatter.go`](../../internal/config/frontmatter.go)). |
| Event → task names | [`internal/trigger/match.go`](../../internal/trigger/match.go) | `MatchOnOR`: implements `on:` OR-semantics against [`types.GitHubEvent`](../../internal/types/types.go). |
| Orchestration | [`internal/engine/`](../../internal/engine/) | `ResolveMatchingTasks` ([`resolver.go`](../../internal/engine/resolver.go)), `RunTask` ([`runner.go`](../../internal/engine/runner.go)), `runAgent` ([`agent.go`](../../internal/engine/agent.go)), state labels ([`state.go`](../../internal/engine/state.go)). |
| Post-agent steps | [`internal/output/`](../../internal/output/) | `RunSuccessOutputs`: `create-pull-request`, `add-labels`, `add-comment` when keys exist under `safe-outputs:`. |
| `wm-agent.yml` generation | [`internal/gen/wmagent.go`](../../internal/gen/wmagent.go), [`schedules.go`](../../internal/gen/schedules.go) | Union of `on.schedule` strings; writes caller workflow. |
| Embedded templates | [`internal/templates/`](../../internal/templates/) | Starters for `gh wm init` (`config.yml`, tasks, `CLAUDE.md`). |
| GitHub API helpers | [`internal/ghclient/`](../../internal/ghclient/) | Labels, issue comments (`gh api`). |

## GitHub Actions: two reusable workflows

Business repos use an **auto-generated** `wm-agent.yml` (from `gh wm init` / `gh wm upgrade`) that calls into **this** repository’s reusable workflows.

```mermaid
flowchart LR
  Caller[wm-agent.yml caller]
  ResolveJob[agent-resolve.yml]
  RunJob[agent-run.yml]

  Caller --> ResolveJob
  ResolveJob -->|"outputs.tasks JSON array"| RunJob
```

1. **`agent-resolve.yml`** ([`.github/workflows/agent-resolve.yml`](../../.github/workflows/agent-resolve.yml))  
   - Checks out the repo, installs `gh-wm` (`go install`), writes `event.json`, runs:
   - `gh-wm resolve --repo-root . --event-name "$EVENT_NAME" --payload event.json --json`  
   - Exposes the printed JSON array as job output `tasks`.

2. **`agent-run.yml`** ([`.github/workflows/agent-run.yml`](../../.github/workflows/agent-run.yml))  
   - Matrix over `fromJSON(needs.resolve.outputs.tasks)` with `fail-fast: false`.  
   - Runs `gh-wm run --repo-root . --task "$TASK_NAME" --event-name "$EVENT_NAME" --payload event.json` with `ANTHROPIC_API_KEY` for the agent.

**Note:** In CI, the installed binary name is `gh-wm`. When installed as a `gh` extension, the same commands are available as `gh wm …`.

## Resolve behavior details

- [`engine.ResolveMatchingTasks`](../../internal/engine/resolver.go) loads all tasks and keeps those where `trigger.MatchOnOR(event, task.OnMap())` is true.
- **Schedule events**: For `event_name == schedule`, every task that includes `on.schedule` matches at resolve time. Optional filter: if `WM_SCHEDULE_CRON` is set (e.g. to the workflow’s cron string), tasks are further filtered with `trigger.ScheduleCronMatches` so only the intended task runs for that cron.
- **Payload**: Event JSON is read from `--payload` or `GITHUB_EVENT_PATH`; name from `--event-name` or `GITHUB_EVENT_NAME`.

## Run behavior details

- [`engine.RunTask`](../../internal/engine/runner.go) loads config + tasks, builds [`TaskContext`](../../internal/types/types.go), optionally loads checkpoint text, applies **working** label if `wm.state_labels` is set, runs `runAgent`, then on success runs **`output.RunSuccessOutputs`** (PR / labels / comment), posts a checkpoint comment if `WM_CHECKPOINT=1`, then applies **done** labels; on failure it applies **failed** labels.
- [`runAgent`](../../internal/engine/agent.go) builds the prompt from the task body + `context.files` + optional checkpoint hint; sets `WM_TASK_TOOLS` when `tools:` is present; selects CLI via `WM_AGENT_CMD` or `engine:` (`claude`, `codex`, `copilot` requires `WM_AGENT_CMD`).
- **Timeout**: [`cmd/run`](../../cmd/run.go) uses `timeout-minutes` from task frontmatter (default 45, max 480).

## Security posture (minimal)

- No sandbox or `safe-outputs` enforcement in-process; workflow permissions and branch protection apply.
- Draft PR defaults in `safe-outputs` / `.wm/config.yml` feed `gh pr create` when `create-pull-request` is listed.
