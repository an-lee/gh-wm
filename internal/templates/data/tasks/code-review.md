---
description: Reviews pull requests for bugs and style issues.

on:
  slash_command:
    name: review
  pull_request:
    types: [review_requested]

timeout-minutes: 15

safe-outputs:
  add-comment:
    max: 5
  create-pull-request-review-comment:
    max: 10
  submit-pull-request-review:
    max: 1

engine: claude
---

# Code review

Review this pull request for correctness, edge cases, security, style, and tests. Prefer **actionable** feedback.

## 1. Local check

- If practical, run the project’s tests, build, or linters against the PR branch so feedback reflects real failures, not guesses.

## 2. Diff review

- Read the changed files and how they connect to existing behavior.
- Note must-fix issues vs nits; prioritize blocking problems.

## 3. Inline feedback

- For specific lines, use **`gh wm emit`** so comments are tied to the diff:

  `gh wm emit create-pull-request-review-comment --body "…" --commit-id <head-sha> --path <file> --line <n> [--side LEFT|RIGHT]`

- `WM_PR_NUMBER` is set when the event is PR-based; use `--target` if you need another PR number.

## 4. Submit review

- Finish with a single submitted review:

  `gh wm emit submit-pull-request-review --event APPROVE|REQUEST_CHANGES|COMMENT [--body "…"]`

- Use **REQUEST_CHANGES** when there are must-fix items; **COMMENT** for questions or optional notes; **APPROVE** when you are satisfied.

## 5. Safe output (required)

- Prefer **inline comments + one `submit-pull-request-review`** over a wall of issue-style text.
- Optional top-level summary on the PR if helpful: `gh wm emit add-comment --body "…"`.
- If there is nothing to post: `gh wm emit noop --message "…"`.
