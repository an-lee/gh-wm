# FAQ

Short answers about **how gh-wm works** and **why it is designed this way**. For full detail, see the linked docs.

## What is this and why use it?

### What is gh-wm in one sentence?

**gh-wm** is a Go [`gh` CLI extension](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions) that reads Markdown tasks under `.wm/tasks/`, matches them to a GitHub event, runs an agent subprocess with the task body as the prompt, then optionally validates and applies **safe outputs** (comments, labels, PRs, etc.) according to each task’s policy. The pipeline is **event → resolve → run**; see [Architecture](architecture.md) and the [index](_index.md) mental model.

### How does gh-wm relate to GitHub Agentic Workflows (gh-aw)?

Task files use **gh-aw–style** Markdown plus YAML frontmatter (`on:`, `safe-outputs:`, `engine:`, …), so many community patterns port directly. gh-wm does **not** use gh-aw’s compile step (no `.lock.yml`, no `gh aw compile`). Enforcement of gh-aw **numeric limits** in `safe-outputs` may differ; keys still **select** which post-agent steps run. See [Task format](task-format.md).

### Can task bodies use `${{ github.event… }}` like gh-aw?

Yes, within a **restricted allowlist**: gh-wm can **validate** those placeholders on **`gh wm upgrade`**, **`gh wm init`**, and **`gh wm validate`**, and **expand** them at **`gh wm run`** when **`compat.gh_aw_expand`** is enabled (default). Prefer canonical **`wm.sanitized.*`** and **`wm.task_name`**; **`steps.sanitized.outputs.*`** is accepted as a gh-aw alias. Details and **`compat.gh_aw_expressions`** (**error** / **warn** / **off**) are in [Task format — expressions](task-format.md).

### Why Go and the `gh` CLI instead of a separate service?

The tool stays a **thin coordinator** on top of GitHub: issues, labels, Actions, and PRs—no extra control plane. Authentication follows **`gh auth login`**. See [Architecture — Goals](architecture.md#goals-design-intent).

## How does a run work end-to-end?

### What is `resolve` vs `run`?

- **`gh wm resolve`** inspects the event (payload or `GITHUB_EVENT_PATH`) and prints a JSON array of **task names** whose `on:` rules match.
- **`gh wm run --task <name>`** runs **one** task: activation, agent, validation, optional safe-outputs, conclusion.

See [CLI reference](cli-reference.md) and [Architecture](architecture.md).

### What are the main phases of a run?

Roughly: **activation** (event/task checks, run directory, feature branch setup for PR outputs) → **agent** (subprocess writing agent log and per-run **`output.jsonl`**) → **validation** (exit status, size/time limits) → **safe-outputs** (apply allowed actions from **`output.jsonl`**) → **conclusion** (checkpoint comment, artifacts). Details: [Architecture — RunTask pipeline](architecture.md).

### What are per-run artifacts?

Each run can use a directory under `.wm/runs/` (or `WM_RUN_DIR`) with files such as **`meta.json`**, **`result.json`**, **`run.json`** (merged snapshot), and agent **`output.jsonl`** (safe-output **requests** via **`gh wm emit`**). See [Architecture](architecture.md).

## Safe outputs and GitHub mutations

### Why `safe-outputs:` and `gh wm emit` instead of raw `gh`?

Declared **`safe-outputs:`** in the task defines **what** the agent may request (comments, labels, PRs, etc.) and **limits** (`max:`, allowlists). The agent records intents via **`gh wm emit`** (NDJSON into **`WM_SAFE_OUTPUT_FILE`**); gh-wm **validates** before touching GitHub. That keeps mutations **policy-bound** and reviewable. See [Task format](task-format.md) and [Architecture](architecture.md).

### What is `WM_SAFE_OUTPUT_FILE`?

**`WM_SAFE_OUTPUT_FILE`** points at per-run **`output.jsonl`**: one JSON object per line (`gh wm emit` appends validated lines). The runner reads that file when applying safe-outputs. Empty output typically **warns and succeeds** (noop). See [CLI reference](cli-reference.md) and [Architecture](architecture.md).

### Why `process-outputs` and `--agent-only` in CI?

In GitHub Actions, the **agent** phase often runs with a **read-only** `GITHUB_TOKEN` so the model cannot directly mutate the repo. The workflow packs the workspace, then runs **`gh wm process-outputs`** with a **write**-capable token to apply safe-outputs. **`gh wm run --agent-only`** stops after the agent phase; **`process-outputs`** completes safe-outputs and conclusion. See [Architecture — GitHub Actions](architecture.md).

## Workflows and repo setup

### Why is `wm-agent.yml` generated? Why `gh wm upgrade` after changing triggers?

**`wm-agent.yml`** is **generated** from your tasks’ union of **`on:`** keys (issue types, schedules, etc.) plus config. Editing triggers in `.wm/tasks/*.md` does not update the caller workflow until you run **`gh wm upgrade`** (or **`init`** on a fresh repo). See [Architecture — wm-agent.yml](architecture.md) and [Task format](task-format.md).

### What do `init`, `add`, and `update` do?

- **`gh wm init`** scaffolds `.wm/config.yml`, starter tasks, and the generated workflow.
- **`gh wm add`** pulls in a task from another repo or URL and runs upgrade afterward.
- **`gh wm update`** re-fetches tasks that declare **`source:`**.

See the repository [README](../../README.md) and [CLI reference](cli-reference.md).

## Agents and local behavior

### Which agent runs by default? How do I override?

The default engine uses **Claude** (`claude -p`) unless you change **`engine:`** or set **`WM_AGENT_CMD`** to another command. The former **`engine: copilot`** name is **removed**; use **`WM_AGENT_CMD`** or **`claude`** / **`codex`**. See [v2.md](v2.md) and [CLI reference](cli-reference.md).

### Why does `run` want a clean git working tree?

So the runner knows exactly what state is being automated (especially for PR-related outputs and branch operations). Use **`--allow-dirty`** when you intentionally run with local changes. See [CLI reference](cli-reference.md).

## Optional features

### What are checkpoints (`WM_CHECKPOINT=1`)?

With checkpoints enabled, the runner can **load** the latest checkpoint from issue comments before the agent and **post** an updated checkpoint after a successful run, helping long-running work resume across runs. See [Architecture](architecture.md).

### What are `schedule` and `WM_SCHEDULE_CRON`?

**`on.schedule`** tasks can match at resolve time; **`WM_SCHEDULE_CRON`** in the environment **filters** which scheduled task should run when multiple crons could match. See [CLI reference](cli-reference.md) and [Task format](task-format.md).

## Where to learn more

- **Overview and doc map:** [_index.md](_index.md) (Contents table).
- **Deep dives:** [Architecture](architecture.md), [Task format](task-format.md), [CLI reference](cli-reference.md).
- **Contributing:** [Development](development.md).

Install and minimal quick start: repository [README](../../README.md).
