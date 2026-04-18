---
description: |
  Implements features or fixes from issues. Triggered by label, slash command, or schedule.

on:
  issues:
    types: [labeled]
  slash_command:
    name: implement
  schedule: "0 22 * * 1-5"

timeout-minutes: 30

safe-outputs:
  create-pull-request:
    draft: true
    labels: [agent]
  add-comment:
    max: 5

engine: claude
---

# Implement feature

You are a developer agent working on this repository.

Follow repository conventions and any project agent guide your team maintains (e.g. `CLAUDE.md`, `AGENTS.md`).

## 1. Context

- Read the issue title, body, and thread so you understand acceptance criteria and constraints.
- Explore the codebase (search, read relevant files) before changing code.

## 2. Implement

- Implement the feature or fix with minimal, focused changes.
- Add or update tests when the repo expects them for the area you touch.

## 3. Validate

- Run the project’s usual checks (tests, build, linters) before you finish.
- Do **not** leave the branch in a broken state: no failing tests or obvious build errors from your commits.
- Commit with clear messages when you have real changes.

## 4. Safe output (required)

Record follow-ups with **`gh wm emit`** (not raw `gh`). The runner sets `WM_REPO_ROOT`, `WM_TASK`, and `WM_SAFE_OUTPUT_FILE` for you.

- **Commits to ship:** after validation passes, if you have commits that should become a PR:

  `gh wm emit create-pull-request --title "…" --body "…"`

  Use `--draft` / `--labels` only if you need to override task defaults; task policy may merge labels.

- **Blocked or no PR:** explain on the issue with:

  `gh wm emit add-comment --body "…"`

- **Nothing to do on GitHub** (e.g. duplicate issue, out of scope): record explicitly:

  `gh wm emit noop --message "…"`

You may emit **one** primary outcome (PR **or** issue comment **or** noop); use `add-comment` for short updates when policy allows multiple lines.
