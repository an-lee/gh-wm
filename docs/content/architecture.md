# Architecture

## Goals (design intent)

- **gh-aw format compatibility**: Task files use Markdown + YAML frontmatter like [Agentic Workflows (gh-aw)](https://github.github.io/gh-aw/); you can drop community workflows into `.wm/tasks/`.
- **No compile step**: No `.lock.yml`, no `gh aw compile`.
- **Go + `go-gh`**: GitHub auth follows `gh auth login` (see [`internal/ghclient`](../../internal/ghclient/) for API usage from commands like `assign`).
- **Thin coordination on GitHub**: Issues, labels, Actions, PRs—no extra control plane.

## High-level pipeline

Each task follows **trigger → resolve → run** (`RunTask`). The run is a **five-phase pipeline** in-process (no gh-aw-style compile): **activation** (event/task validation, optional `wm.state_labels` working label, feature branch for PR mode, **per-run artifact dir** under `.wm/runs/` or `WM_RUN_DIR`), **agent** (`runAgent` subprocess), **validation** (successful exit and output size bound; context deadline surfaced as timeout), **safe-outputs** ([`internal/output`](../../internal/output/) when keys exist under `safe-outputs:` — hints only; not enforced like gh-aw), and **conclusion** (`defer`: done/failed labels, checkpoint comment, branch rollback on failure, **`result.json`**). [`RunTask`](../../internal/engine/runner.go) returns a [`types.RunResult`](../../internal/types/types.go) with `Phase`, `Success`, `Errors`, `RunDir`, and `AgentResult`; `wm run` logs `phase=` and `artifacts=` on stderr when a run directory is created.

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
| CLI entry | [`cmd/`](../../cmd/) | Cobra commands: `init`, `upgrade`, `update`, `assign`, `resolve`, `run`, `status`, `logs`, `add`. |
| Config + tasks | [`internal/config/`](../../internal/config/) | Load `.wm/config.yml`, parse `.wm/tasks/*.md` frontmatter ([`frontmatter.go`](../../internal/config/frontmatter.go)). |
| Event → task names | [`internal/trigger/match.go`](../../internal/trigger/match.go) | `MatchOnOR`: implements `on:` OR-semantics against [`types.GitHubEvent`](../../internal/types/types.go). |
| Orchestration | [`internal/engine/`](../../internal/engine/) | `ResolveMatchingTasks` and `ResolveForcedTask` ([`resolver.go`](../../internal/engine/resolver.go)) — forced resolve pins one task by filename without evaluating `on:` (matches local `gh wm run`); `RunTask` ([`runner.go`](../../internal/engine/runner.go)), per-run dirs ([`rundir.go`](../../internal/engine/rundir.go)), activation checks ([`activation.go`](../../internal/engine/activation.go)), output validation ([`validation.go`](../../internal/engine/validation.go)), conclusion/defer ([`conclusion.go`](../../internal/engine/conclusion.go)), `runAgent` ([`agent.go`](../../internal/engine/agent.go)), state labels ([`state.go`](../../internal/engine/state.go)). |
| Post-agent steps | [`internal/output/`](../../internal/output/) | `RunSuccessOutputs`: `create-pull-request`, `add-labels`, `add-comment` when keys exist under `safe-outputs:`. |
| `wm-agent.yml` generation | [`internal/gen/wmagent.go`](../../internal/gen/wmagent.go), [`schedules.go`](../../internal/gen/schedules.go) | Union of `on.schedule` strings; writes caller workflow. |
| Embedded templates | [`internal/templates/`](../../internal/templates/) | Starters for `gh wm init` (`config.yml`, tasks). |
| GitHub API helpers | [`internal/ghclient/`](../../internal/ghclient/) | Labels, issue comments (`gh api`). |
| Feature branch before PR | [`internal/gitbranch/`](../../internal/gitbranch/) | When `safe-outputs` includes `create-pull-request`, create `wm/<task>-…` on the default branch so the agent does not commit directly to `main`. |

## GitHub Actions: reusable workflows and generated `wm-agent.yml`

Business repos use an **auto-generated** `wm-agent.yml` (from `gh wm init` / `gh wm upgrade`). Runner labels come from **`workflow.runs_on`** in [`.wm/config.yml`](task-format.md); optional **`workflow.pre_steps`** lists prerequisite Actions steps (toolchains, deps); `upgrade` rewrites `wm-agent.yml` when you change them.

- **Resolve** always uses reusable **`agent-resolve.yml`**.
- **Run** uses reusable **`agent-run.yml`** when **`workflow.pre_steps` is empty**. If **`workflow.pre_steps` is set**, the generator embeds the same checkout → pre-steps → `gh-wm` install → `gh-wm run` sequence **inline** in `wm-agent.yml` (reusable workflows cannot take arbitrary step YAML as inputs).

```mermaid
flowchart LR
  Caller[wm-agent.yml caller]
  ResolveJob[agent-resolve.yml]
  RunJob[agent-run.yml or inline run job]

  Caller --> ResolveJob
  ResolveJob -->|"outputs.tasks JSON array"| RunJob
```

1. **`agent-resolve.yml`** ([`.github/workflows/agent-resolve.yml`](../../.github/workflows/agent-resolve.yml))  
   - `runs-on` is driven by the **`runs_on` workflow input** (JSON array of labels), with default `["ubuntu-latest"]`; generated `wm-agent.yml` passes labels from `.wm/config.yml`.
   - Checks out the repo, installs `gh-wm` (`go install`), writes the GitHub event JSON to **`.wm/runs/github-event.json`** (under the ignored `runs/` tree; see **`.wm/.gitignore`**) so `git status` stays clean for `gh-wm run`’s working-tree check, then runs:
   - `gh-wm resolve --repo-root . --event-name "$EVENT_NAME" --payload .wm/runs/github-event.json --json`  
   - Exposes the printed JSON array as job output `tasks`.

2. **`agent-run.yml`** ([`.github/workflows/agent-run.yml`](../../.github/workflows/agent-run.yml)) — **when `workflow.pre_steps` is unset**  
   - Matrix over `fromJSON(needs.resolve.outputs.tasks)` with `fail-fast: false`.  
   - Writes the same **`.wm/runs/github-event.json`** payload and runs `gh-wm run --repo-root . --task "$TASK_NAME" --event-name "$EVENT_NAME" --payload .wm/runs/github-event.json` with `ANTHROPIC_API_KEY` for the agent.

3. **Inline `run` job** — **when `workflow.pre_steps` is set**  
   - Same matrix and `gh-wm run` invocation (payload under **`.wm/runs/github-event.json`** as above); steps include **`workflow.pre_steps`** after checkout and before installing `gh-wm`.

**Note:** In CI, the installed binary name is `gh-wm`. When installed as a `gh` extension, the same commands are available as `gh wm …`.

## Resolve behavior details

- [`engine.ResolveMatchingTasks`](../../internal/engine/resolver.go) loads all tasks and keeps those where `trigger.MatchOnOR(event, task.OnMap())` is true.
- **Schedule events**: For `event_name == schedule`, every task that includes `on.schedule` matches at resolve time. Optional filter: if `WM_SCHEDULE_CRON` is set (e.g. to the workflow’s cron string), tasks are further filtered with `trigger.ScheduleCronMatches` (recomputes the same fuzzy cron as `gen.FuzzyNormalizeSchedule` for that task path) so only the intended task runs for that cron.
- **Payload**: Event JSON is read from `--payload` or `GITHUB_EVENT_PATH` when set; if both are unset, the payload defaults to `{}`. Event name comes from `--event-name` or `GITHUB_EVENT_NAME`.

## Run behavior details

- [`engine.RunTask`](../../internal/engine/runner.go) returns a [`RunResult`](../../internal/types/types.go) with phase, accumulated errors, timing, and **`RunDir`**. It validates the event and engine, builds [`TaskContext`](../../internal/types/types.go), creates a **per-run directory** ([`NewRunDir`](../../internal/engine/rundir.go): `.wm/runs/<id>/` or `WM_RUN_DIR/<id>/`), optionally loads checkpoint text, applies **working** label if `wm.state_labels` is set, optionally creates a **feature branch** via [`internal/gitbranch`](../../internal/gitbranch/) when `safe-outputs` includes `create-pull-request` (see CLI reference), runs `runAgent` (writes **`prompt.md`**, streams combined stdout/stderr to a per-run **agent log file** — default **`agent-stdout.log`**, or structured **`conversation.json`** / **`conversation.jsonl`** when print-mode JSON is enabled for the built-in **`claude`** CLI; **SIGTERM** then kill on Unix when the run context is canceled), validates agent output size (from log file stat when present) and success, then on success runs **`output.RunSuccessOutputs`** (PR / labels / comment). A **deferred conclusion** always runs: on success, checkpoint comment if `WM_CHECKPOINT=1` and **done** labels; on failure, **failed** labels and **checkout** of the previous branch if a feature branch was created; finally **`result.json`** and **`meta.json`** (phase **conclusion**).
- [`runAgent`](../../internal/engine/agent.go) builds the prompt from the task body + `context.files` + optional checkpoint hint; sets `WM_TASK_TOOLS` when `tools:` is present; selects CLI via `WM_AGENT_CMD` or `engine:` (`claude`, `codex`, `copilot` requires `WM_AGENT_CMD`). Default **`claude`** uses **stdin** for the prompt, **`--dangerously-skip-permissions`**, and optional **`--model`** / **`--max-turns`** from global config so the agent can run tools (including **`gh`**) non-interactively. When **`claude_output_format`** / **`WM_CLAUDE_OUTPUT_FORMAT`** request **`json`** or **`stream-json`**, the runner also passes **`--output-format`** (built-in **`claude`** only; **`WM_AGENT_CMD`**, **codex**, and **copilot** keep plain-text capture). In-memory **`Stdout`/`Summary`** hold a **64 KiB tail** of the transcript when a run dir is used (full text is on disk).
- **Timeout**: [`cmd/run`](../../cmd/run.go) uses `timeout-minutes` from task frontmatter (default 45, max 480).

## RunTask pipeline (detailed reference)

Implementation: [`RunTask`](../../internal/engine/runner.go), [`rundir.go`](../../internal/engine/rundir.go), [`activation.go`](../../internal/engine/activation.go), [`validation.go`](../../internal/engine/validation.go), [`conclusion.go`](../../internal/engine/conclusion.go), [`state.go`](../../internal/engine/state.go).

**Contract:** One `gh-wm run` / `gh wm run` process executes the pipeline below. The primary API result is [`types.RunResult`](../../internal/types/types.go) (`Phase`, `Success`, `AgentResult`, `Errors`, `Duration`, `RunDir`) plus a Go `error`. **Conclusion** (labels, checkpoint, branch rollback, **`result.json`**) runs in a **`defer`** after `task` and `tc` are set; if the run fails earlier (e.g. config load, missing task, invalid event), `tc` may be nil and **conclusion does nothing** (and no run dir is created if failure is before `NewRunDir`).

### Phase 1 — Activation (`PhaseActivation`)

| Reads | Purpose |
|-------|---------|
| Disk: `.wm/config.yml`, `.wm/tasks/*.md` | `config.Load` → global config + tasks |
| In-memory: `*GitHubEvent` | Must be non-nil; `Payload` non-nil; `Name` non-empty (except `unknown` for local empty-event runs) |
| Env: `GITHUB_REPOSITORY`, `WM_AGENT_CMD`, task `engine:` / global `engine` | Engine validation |
| Env: `WM_CHECKPOINT=1` (optional) | Enables checkpoint **read** below |
| Env: `WM_RUN_DIR` (optional) | Base path for per-run dirs instead of `<repo>/.wm/runs/` |
| Disk: `claude_output_format` in `.wm/config.yml`; env: `WM_CLAUDE_OUTPUT_FORMAT` (optional) | Overrides config when set: **`text`** (default), **`json`**, or **`stream-json`** for built-in **`claude`** — chooses run-dir filename and **`--output-format`** |
| GitHub API: `ghclient.ListIssueCommentBodies` (optional) | Only with checkpoint mode + `GITHUB_REPOSITORY` + issue/PR number: load comment bodies to find latest `<!-- wm-checkpoint: … -->` |

| In-memory outputs | |
|-------------------|--|
| `TaskContext` | Task name, `RepoPath`, event, issue/PR numbers from payload (`extractNumbers`) |
| `CheckpointHint` | Latest checkpoint summary text for the agent prompt |
| `wm` extension | `wm.state_labels` from task frontmatter |

| Writes / side effects | Where |
|----------------------|--------|
| Optional: **working** label | GitHub (if `wm.state_labels.working` and repo + issue/PR number). Errors are logged and appended to `RunResult.Errors`; run **continues**. |
| Optional: feature branch | Local git repo (`gitbranch.PrepareFeatureForPR`) when `safe-outputs` includes `create-pull-request` |
| **Per-run directory** | **`<repo>/.wm/runs/<id>/`** or **`WM_RUN_DIR/<id>/`**: `meta.json` (phase **activation**); **`PruneRunDirs`** drops dirs older than 7 days under `.wm/runs` (and under `WM_RUN_DIR` when set) |

### Phase 2 — Agent (`PhaseAgent`)

| Reads | Purpose |
|-------|---------|
| Task body, global `context.files` | Prompt in [`runAgent`](../../internal/engine/agent.go) |
| `CheckpointHint` | Appended to prompt |
| Repo working tree | Agent subprocess `Dir` = `--repo-root`; agent may edit files / run git |

| Outputs | |
|---------|--|
| `AgentResult` | Combined transcript: full agent log on disk (**`agent-stdout.log`**, or **`conversation.json`** / **`conversation.jsonl`** when structured print-mode output is enabled for built-in **`claude`**); **`Stdout`/`Summary`** hold a **64 KiB tail** when a run dir exists (for checkpoints/comments); `Success`, `ExitCode`, **`TimedOut`** if context deadline exceeded |
| Optional stream | Tee to `RunOptions.LogWriter` (CLI uses stderr) and to the same per-run log file |

### Phase 3 — Validation (`PhaseValidation`)

| Reads | Purpose |
|-------|---------|
| `AgentResult`, run `context` | In-process checks; deadline → **timeout** error |

| Checks | |
|--------|--|
| [`validateAgentOutputErr`](../../internal/engine/validation.go) | Non-nil result, `Success`, not timed out; size from the on-disk agent log path when set, else in-memory text length ≤ 12 MiB. **Empty** successful output is allowed. |

### Phase 4 — Safe outputs (`PhaseOutputs`)

| Reads | Purpose |
|-------|---------|
| `AgentResult`, `TaskContext`, `safe-outputs:` keys | [`RunSuccessOutputs`](../../internal/output/output.go) |

| Writes (if configured) | Persistence |
|---------------------------|-------------|
| `create-pull-request` | `git push`, `gh pr create` → GitHub |
| `add-labels` | GitHub API → labels |
| `add-comment` | `gh issue comment` / `gh pr comment` → GitHub |

### Phase 5 — Conclusion (deferred)

Runs in `defer` via [`concludeRun`](../../internal/engine/conclusion.go) only when **`task` and `tc` are non-nil**.

**On success (`runSucceeded`):**

| Action | Reads | Writes |
|--------|-------|--------|
| Checkpoint | `WM_CHECKPOINT=1`, `AgentResult` text | New issue comment (`ghclient.PostIssueComment`) |
| State **done** | `wm.state_labels` | Remove working / add done label on GitHub |

**On failure:**

| Action | Reads | Writes |
|--------|-------|--------|
| Branch rollback | `branchCreated`, `prevBranch` | `git checkout` previous branch on disk (if applicable) |
| State **failed** | `wm.state_labels` | Remove working / add failed label on GitHub (removing **working** treats **404** / missing label as non-fatal) |
| **Artifacts** | `RunResult` | **`meta.json`** (phase **conclusion**), **`result.json`** (serialized outcome) |

Checkpoint or label failures are appended to `RunResult.Errors` and do not always change the primary returned `error` from an earlier phase.

### What persists where

| Kind | Where |
|------|--------|
| `RunResult` / errors | In-memory for the process; CLI prints `phase=`, **`artifacts=`**, and `failure phase:` on **stderr** |
| Per-run artifacts | **`.wm/runs/<id>/`** (or **`WM_RUN_DIR/<id>/`**): `prompt.md`; combined agent stdout/stderr (**`agent-stdout.log`** by default, or **`conversation.json`** / **`conversation.jsonl`** when **`claude_output_format`** / **`WM_CLAUDE_OUTPUT_FORMAT`** is **`json`** / **`stream-json`** for built-in **`claude`**); `meta.json` (phase updates); `result.json` (final snapshot). Ignore **`runs/`** under **`.wm/`** via **`.wm/.gitignore`** (`gh wm init` / `gh wm upgrade` ensure that file). |
| Agent tail in memory | Last **64 KiB** of combined output in `AgentResult` when a run dir is used (full output remains in the per-run agent log file above) |
| Repo state | Whatever git / the agent wrote under `--repo-root` |
| Coordination | GitHub: labels, issue/PR comments, PRs — the main external persistence |
| Checkpoints | Issue comments when `WM_CHECKPOINT=1`, encoded in [`internal/checkpoint`](../../internal/checkpoint/checkpoint.go) |

**Note:** `RunResult.Phase` is the last phase reached or where failure occurred; it is **not** set to a separate `conclusion` value after the defer. There is no collaborator/actor permission gate in the current implementation.

## Security posture (minimal)

- No sandbox or `safe-outputs` enforcement in-process; workflow permissions and branch protection apply.
- Draft PR defaults in `safe-outputs` / `.wm/config.yml` feed `gh pr create` when `create-pull-request` is listed.
