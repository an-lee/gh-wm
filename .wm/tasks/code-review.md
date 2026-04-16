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
