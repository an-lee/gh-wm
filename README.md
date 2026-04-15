# gh-wm — Workflow Manager for personal agentic development

`gh-wm` is a **Go** [`gh` CLI extension](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions) that runs **gh-aw–compatible** task files (Markdown + YAML frontmatter) from `.wm/tasks/` without compiling to lockfiles, without AWF, and without enforcing `safe-outputs` (those fields are treated as hints).

## Documentation

Full docs for humans and AI agents live in **[`docs/`](docs/README.md)**:

| Doc | Contents |
|-----|----------|
| [docs/README.md](docs/README.md) | Index and mental model |
| [docs/architecture.md](docs/architecture.md) | Pipelines, code map, GitHub Actions |
| [docs/task-format.md](docs/task-format.md) | `.wm/config.yml`, `on:` semantics, gh-aw notes |
| [docs/cli-reference.md](docs/cli-reference.md) | Commands, flags, environment variables |
| [docs/development.md](docs/development.md) | Contributing and extending the Go codebase |

## Install

```bash
go install github.com/gh-wm/gh-wm@latest
# or build from this repo
go build -o gh-wm .
```

As a `gh` extension (after publishing releases):

```bash
gh extension install gh-wm/gh-wm
```

## Quick start (in a repository)

```bash
gh wm init
```

This creates:

- `.wm/config.yml` — global defaults (`engine`, `context.files`, …)
- `.wm/tasks/*.md` — starter tasks (gh-aw–style frontmatter + markdown body = agent prompt)
- `.github/workflows/wm-agent.yml` — **auto-generated** caller workflow (do not edit by hand)
- `CLAUDE.md` if missing

Set where reusable workflows live (default `gh-wm/gh-wm`):

```bash
export GH_WM_REPO=your-org/gh-wm
gh wm upgrade
```

## Commands

| Command | Purpose |
|--------|---------|
| `gh wm init` | Scaffold `.wm/`, tasks, and `wm-agent.yml` |
| `gh wm upgrade` | Regenerate `wm-agent.yml` (union of schedules from tasks) |
| `gh wm assign <n>` | Add label (default `agent`) to issue `#n` |
| `gh wm resolve` | List task names matching `GITHUB_EVENT` / payload |
| `gh wm run --task <name>` | Run one task for the current event |
| `gh wm status` | List issues with agent-related labels |
| `gh wm logs <n>` | List recent `wm-agent` workflow runs |

### CI entrypoints

- **`gh wm resolve`** — reads `--payload` (or `GITHUB_EVENT_PATH`), prints JSON array of matching task names.
- **`gh wm run --task …`** — runs the agent (default: `claude -p` with task body + `CLAUDE.md`). Override with `WM_AGENT_CMD`.

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

Details: [docs/architecture.md](docs/architecture.md).

## License

MIT — see [LICENSE](LICENSE).
