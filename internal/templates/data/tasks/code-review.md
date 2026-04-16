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
    max: 10

engine: claude
---

# Code Review

Review this pull request. Check for bugs, edge cases, style, and tests.

Provide actionable feedback as review comments.

## Safe output (required)

Before exiting, write JSON to **`WM_OUTPUT_FILE`**: `{"items":[{"type":"add_comment","body":"…"}]}` with your review summary (or **`noop`** if nothing to post).
