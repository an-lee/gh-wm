# gh-wm documentation

**gh-wm** is a Go [`gh` CLI extension](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions) that runs **gh-aw–compatible** task files (Markdown + YAML frontmatter) from `.wm/tasks/` **without** compiling to lockfiles or AWF. When `safe-outputs:` is set, the agent records outputs via **`gh wm emit`** (NDJSON into **`output.jsonl`**); gh-wm validates and applies them (`max:`, allowlists, etc.).

This documentation is the **canonical reference** for how the project works and how to extend it.

**Published site:** [https://gh-wm.github.io/gh-wm/](https://gh-wm.github.io/gh-wm/) (same Markdown, rendered with Hugo).

## Who this is for

| Reader | Start here |
|--------|------------|
| **Quick Q&A** | [FAQ](faq.md) |
| **Humans** | [Architecture](architecture.md) → [Architecture review](architecture-review.md) → [v2 changes](v2.md) → [CLI reference](cli-reference.md) → [Task format](task-format.md) |
| **AI coding agents** | Read [Architecture](architecture.md), [Architecture review](architecture-review.md), and [Development](development.md) before changing behavior. Use [Task format](task-format.md) when editing `.wm/tasks/*.md` or explaining compatibility with [gh-aw](https://github.github.io/gh-aw/). |

## Contents

| Doc | Purpose |
|-----|---------|
| [faq.md](faq.md) | Frequently asked questions: how gh-wm works, why safe-outputs and CI split, workflows, agents—links to deeper pages. |
| [architecture.md](architecture.md) | End-to-end flow: GitHub Actions → `resolve` → matrix `run`, Go packages, and **RunTask pipeline (detailed reference)** (reads/writes per phase). |
| [architecture-review.md](architecture-review.md) | Design review, target module layout, phased refactor roadmap (v2). |
| [v2.md](v2.md) | Consolidated v2 breaking / deprecation notes (engine, safe-outputs, timeout, artifacts, GitHub transport). |
| [task-format.md](task-format.md) | `.wm/config.yml`, `.wm/tasks/<name>.md` frontmatter, `on:` trigger semantics, `safe-outputs:`. |
| [cli-reference.md](cli-reference.md) | Every `gh wm` / `gh-wm` subcommand, flags, and environment variables. |
| [development.md](development.md) | Repo layout, extension points, build/test, and conventions for contributors. |

## One-sentence mental model

**GitHub delivers an event → `gh wm resolve` lists matching task names → Actions runs `gh wm run --task <name>` per match; each run runs the agent (default: `claude -p`), then applies validated safe outputs from **`gh wm emit`** when `safe-outputs:` is configured.**

For install and a minimal user quick start, see the repository [README](../../README.md).
