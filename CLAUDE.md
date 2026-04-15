# gh-wm (this repository)

You are working on the **gh-wm** CLI: a Go `gh` extension that resolves GitHub events to `.wm/tasks/*.md` tasks and runs an agent subprocess (`claude -p` by default, or `WM_AGENT_CMD`).

## Before changing behavior

1. Read **[`docs/README.md`](docs/README.md)** for the mental model.
2. For code changes, read **[`docs/architecture.md`](docs/architecture.md)** and **[`docs/development.md`](docs/development.md)**.
3. For task markdown / frontmatter, use **[`docs/task-format.md`](docs/task-format.md)**.
4. For CLI flags and env vars, use **[`docs/cli-reference.md`](docs/cli-reference.md)**.

## Accuracy

Keep docs in **`docs/`** aligned with code (`internal/engine`, `internal/trigger`, `cmd/`). If you add flags, matchers, or workflow contract changes, update the relevant doc in the same change.

## Templates vs this repo

`gh wm init` writes embedded templates from [`internal/templates/data/`](internal/templates/data/) into **user repos**. That is separate from the documentation in **`docs/`**, which describes **this** project.
