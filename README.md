# gh-wm — Workflow Manager for personal agentic development

`gh-wm` is a **Go** [`gh` CLI extension](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions) that runs **gh-aw–compatible** task files (Markdown + YAML frontmatter) from `.wm/tasks/` without compiling to lockfiles, without AWF, and without enforcing gh-aw **limits** in `safe-outputs` (keys still **select** optional post-agent steps).

## Documentation

Full docs for humans and AI agents live under **[`docs/`](docs/README.md)** (Markdown in [`docs/content/`](docs/content/), Hugo in the same folder). A browsable HTML version is published at **[https://gh-wm.github.io/gh-wm/](https://gh-wm.github.io/gh-wm/)**.

| Doc                                                            | Contents                                       |
| -------------------------------------------------------------- | ---------------------------------------------- |
| [docs/content/\_index.md](docs/content/_index.md)              | Index and mental model                         |
| [docs/content/architecture.md](docs/content/architecture.md)   | Pipelines, code map, GitHub Actions            |
| [docs/content/task-format.md](docs/content/task-format.md)     | `.wm/config.yml`, `on:` semantics, gh-aw notes |
| [docs/content/cli-reference.md](docs/content/cli-reference.md) | Commands, flags, environment variables         |
| [docs/content/development.md](docs/content/development.md)     | Contributing and extending the Go codebase     |

## Install

```bash
go install github.com/an-lee/gh-wm@latest
# or build from this repo
go build -o gh-wm .
```

As a `gh` extension (after publishing releases):

```bash
gh extension install an-lee/gh-wm
```

## Quick start (in a repository)

```bash
gh wm init
```

This creates:

- `.wm/config.yml` — global defaults (`engine`, `context.files`, …)
- `.wm/tasks/*.md` — starter tasks (gh-aw–style frontmatter + markdown body = agent prompt)
- `.github/workflows/wm-agent.yml` — **auto-generated** caller workflow (do not edit by hand)

Set where reusable workflows live (default `an-lee/gh-wm`):

```bash
export GH_WM_REPO=your-org/gh-wm
gh wm upgrade
```

## Commands

| Command                   | Purpose                                                       |
| ------------------------- | ------------------------------------------------------------- |
| `gh wm init`              | Scaffold `.wm/`, tasks, and `wm-agent.yml`                    |
| `gh wm upgrade`           | `gh extension upgrade` (best-effort) + regenerate `wm-agent.yml` |
| `gh wm update`            | Re-fetch tasks that have `source:` (URL or `owner/repo/path`)   |
| `gh wm add <…>`           | Add a task `.md` (`owner/repo/task`, URL, or path; then `upgrade`) |
| `gh wm assign <n>`        | Add label (default `agent`) to issue `#n`                     |
| `gh wm resolve`           | List task names matching `GITHUB_EVENT` / payload             |
| `gh wm run --task <name>` | Run one task (agent + optional `safe-outputs` / labels)       |
| `gh wm status`            | List issues with agent-related labels (`--all` = `gh search`) |
| `gh wm logs <n>`          | List `wm-agent` runs (best-effort match on `#n` in title)     |

### CI entrypoints

- **`gh wm resolve`** — reads `--payload` or `GITHUB_EVENT_PATH` (if both unset, payload defaults to `{}`), prints JSON array of matching task names.
- **`gh wm run --task …`** — same payload resolution as `resolve`; requires a **clean git working tree** unless **`--allow-dirty`**. Streams agent output to **stderr** and prints a short summary when finished. Runs the agent (default: `claude -p` with the task body plus optional `context.files` from `.wm/config.yml`; `timeout-minutes` from frontmatter). Override with `WM_AGENT_CMD`. On success, runs `safe-outputs` steps (e.g. PR, comment) and optional `wm.state_labels`. Use `WM_CHECKPOINT=1` for checkpoint load/post.

### Secrets

- **`ANTHROPIC_API_KEY`** — for Claude Code in Actions (configure in repo secrets).

## Task format

Tasks are `.wm/tasks/<name>.md` with YAML frontmatter compatible with [GitHub Agentic Workflows](https://github.github.io/gh-aw/) (`on:`, `safe-outputs:`, `engine:`, …) and a markdown body used as the agent prompt.

Optional **`wm:`** block (ignored by gh-aw) for gh-wm-only options, e.g. `state_labels`.

## Architecture (summary)

1. **One caller workflow** (`wm-agent.yml`) listens for broad events and schedule crons.
2. **Resolve job** runs `gh wm resolve` → JSON list of matching tasks.
3. **Matrix job** runs `gh wm run --task <name>` in parallel for each match (`fail-fast: false`).

After changing any task’s `on:` triggers, run **`gh wm upgrade`** to refresh `wm-agent.yml`.

Details: [docs/content/architecture.md](docs/content/architecture.md).

## License

MIT — see [LICENSE](LICENSE).
