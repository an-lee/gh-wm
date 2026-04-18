# CLAUDE.md

**gh-wm** is a Go `gh` CLI extension: resolves GitHub events → `.wm/tasks/*.md` → runs an agent subprocess (default `claude -p`; override with `WM_AGENT_CMD`).

## Commands

```bash
go build -o gh-wm .                  # build
go test ./...                        # test all
go test ./internal/config/... -run TestSplitFrontmatter  # single test

./gh-wm resolve --repo-root . --event-name issues --payload event.json --json
./gh-wm run --repo-root . --task <name> --event-name workflow_dispatch
./gh-wm run --repo-root . --task <name> --agent-only   # CI: agent phase only; then gh wm process-outputs
./gh-wm process-outputs --repo-root . --task <name> --event-name issues --payload event.json
# or: --run-dir .wm/runs/<id>
./gh-wm run --repo-root . --task <name> --remote   # dispatch on GitHub
```

## After editing Go code

Run `make ci` — it mirrors `.github/workflows/ci.yml` (fmt-check → vet → test → build):

```bash
make ci
```

Or individually: `gofmt -l .` (must be empty), `go vet ./...`, `go test ./...`, `go build -v ./...`.

## Architecture

Core pipeline: **event → resolve → run agent**.

| Layer | Key files | Purpose |
|-------|-----------|---------|
| CLI | `cmd/*.go` | Cobra commands: `init`, `upgrade`, `update`, `add`, `assign`, `resolve`, `run`, `process-outputs`, `emit`, `status`, `logs`, `version` |
| Resolver | `internal/engine/resolver.go` | `ResolveMatchingTasks` — loads `.wm/tasks/*.md`, calls `trigger.MatchOnOR`, returns matches |
| Trigger | `internal/trigger/match.go` | `MatchOnOR` — OR-semantics over `on:` frontmatter (`issues`, `issue_comment`, `pull_request`, `slash_command`, `schedule`, `workflow_dispatch`) |
| Runner | `internal/engine/runner.go`, `agent.go`, `rundir.go`, `process_outputs.go` | `RunTask` → `runAgent`; optional **`RunOptions.AgentOnly`** stops before safe-outputs; **`ProcessRunOutputs`** completes safe-outputs + conclusion (CI write token). **`timeout-minutes`** enforced in `RunTask` (default 45) |
| Config | `internal/config/` | `.wm/config.yml` (GlobalConfig), task frontmatter parsing (`map[string]any`; add typed accessors when fields become first-class) |
| Output | `internal/output/` | Reads `WM_SAFE_OUTPUT_FILE` (NDJSON from `gh wm emit`); validates against `safe-outputs:` policy; executes items (noop, comment, label, issue, PR, PR review comments, submit PR review, missing_tool, missing_data); optional `safe-outputs.messages` status comments |
| Checkpoint | `internal/checkpoint/` | When `WM_CHECKPOINT=1`: loads/posts checkpoint state via issue comments |
| Git helpers | `internal/gitstatus/`, `internal/gitbranch/` | Clean-tree check (`run` requires clean unless `--allow-dirty`); feature-branch prep for PR outputs |
| GitHub | `internal/ghclient/`, `internal/gh/` | Default: `gh api` subprocess; set **`GH_WM_REST=1`** for `go-gh` REST ([`internal/gh`](internal/gh/)). `CurrentRepo` still shells out to `gh repo view`. |
| Generator | `internal/gen/wmagent.go` | Generates `wm-agent.yml` (single template: inline vs reusable `agent-run.yml`); also cron scheduling helpers |
| Templates | `internal/templates/data/` | Embedded defaults written by `gh wm init` into user repos |

Agent prompt flow: task body + `context.files` + safe-output reference (user message) → `prompt.md` → stdin to agent. When `safe-outputs:` is set, the built-in **`claude`** engine also passes **`--append-system-prompt`** with enforcement text (use **`gh wm emit`**, not raw **`gh`** for those mutations). Safe outputs are recorded only via NDJSON in `WM_SAFE_OUTPUT_FILE`. If it is empty, the run warns and succeeds (implicit noop).

## Non-obvious constraints

- **Binary name duality**: `go install` produces `gh-wm`; as a `gh` extension it's `gh wm …`. Same binary.
- **`wm-agent.yml` is generated**: Written by `gh wm init` / `gh wm upgrade` (template in `internal/gen/wmagent.go`). Never hand-edit in consumer repos. `upgrade` also best-effort runs `gh extension upgrade an-lee/gh-wm`.
- **`gh wm update`**: Re-fetches tasks with a `source:` field (URL or `owner/repo/path` shorthand, set by `gh wm add`).
- **Schedule cron filtering**: All `on.schedule` tasks match at resolve time; `WM_SCHEDULE_CRON` env var further filters to the correct task.
- **`engine:` frontmatter**: Selects default agent CLI (`claude`, `codex`). The former `copilot` engine name is removed — use `WM_AGENT_CMD` for a custom CLI.
- **Per-run `run.json`**: Written alongside `meta.json` / `result.json` (merged snapshot for tooling).
- **`WM_SAFE_OUTPUT_FILE`**: Per-run `output.jsonl` — `gh wm emit` appends validated lines for safe-outputs. **`WM_REPO_ROOT`**, **`WM_ISSUE_NUMBER`**, **`WM_PR_NUMBER`** assist emit validation.
- **`WM_LOG_FORMAT=json`**: Structured `slog` on stderr for pipeline phases.
- **`install_claude_code`**: `agent-run.yml` input, default `true` from `workflow.install_claude_code` in `.wm/config.yml`.
- **`gh_wm_extension_version`**: Optional `workflow.gh_wm_extension_version` in `.wm/config.yml`; passed to reusable workflows so CI can run `gh extension install owner/repo --pin <ref>` (see `gh help extension install`).
- **CI token sandbox**: `agent-run.yml` runs **`gh wm run --agent-only`** with read-only `GITHUB_TOKEN`, packs the workspace, then **`gh wm process-outputs`** with write permissions so **`gh wm emit`** is the enforced path for GitHub mutations.

## Before changing behavior

1. **Mental model**: [`docs/content/_index.md`](docs/content/_index.md)
2. **Code changes**: [`docs/content/architecture.md`](docs/content/architecture.md), [`docs/content/development.md`](docs/content/development.md)
3. **Task format**: [`docs/content/task-format.md`](docs/content/task-format.md)
4. **CLI flags / env vars**: [`docs/content/cli-reference.md`](docs/content/cli-reference.md)

Keep `docs/` aligned with code. If you add flags, matchers, or workflow changes, update the relevant doc in the same change.

## Templates vs docs

`gh wm init` writes embedded templates from `internal/templates/data/` into **user repos**. The `docs/` directory documents **this project** itself — they are separate.
