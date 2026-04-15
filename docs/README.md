# gh-wm documentation

**gh-wm** is a Go [`gh` CLI extension](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions) that runs **gh-aw–compatible** task files (Markdown + YAML frontmatter) from `.wm/tasks/` **without** compiling to lockfiles, without AWF, and without enforcing gh-aw **limits** in `safe-outputs` (those keys still **select** optional post-agent steps such as PR creation and comments).

This folder is the **canonical reference** for how the project works and how to extend it.

## Who this is for

| Reader | Start here |
|--------|------------|
| **Humans** | [Architecture](architecture.md) → [CLI reference](cli-reference.md) → [Task format](task-format.md) |
| **AI coding agents** | Read [Architecture](architecture.md) and [Development](development.md) before changing behavior. Use [Task format](task-format.md) when editing `.wm/tasks/*.md` or explaining compatibility with [gh-aw](https://github.github.io/gh-aw/). |

## Contents

| Doc | Purpose |
|-----|---------|
| [architecture.md](architecture.md) | End-to-end flow: GitHub Actions → `resolve` → matrix `run`, and how that maps to Go packages. |
| [task-format.md](task-format.md) | `.wm/config.yml`, `.wm/tasks/<name>.md` frontmatter, `on:` trigger semantics, `wm:` extensions. |
| [cli-reference.md](cli-reference.md) | Every `gh wm` / `gh-wm` subcommand, flags, and environment variables. |
| [development.md](development.md) | Repo layout, extension points, build/test, and conventions for contributors. |

## One-sentence mental model

**GitHub delivers an event → `gh wm resolve` lists matching task names → Actions runs `gh wm run --task <name>` per match; each run runs the agent (default: `claude -p`), then optional `safe-outputs` steps (`gh pr create`, labels, issue/PR comment) and `wm.state_labels` updates.**

For install and a minimal user quick start, see the repository [README](../README.md).
