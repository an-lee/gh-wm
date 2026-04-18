---
description: |
  Implements features or fixes from issues. Triggered by label, slash command, or schedule.

on:
  issues:
    types: [labeled]
    labels: [implement]
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

# Implement Feature

You are a developer agent working on this repository.

Implement the feature or fix described in the issue.
Create tests if applicable. Commit with clear messages.

Read CLAUDE.md first for project-specific conventions.
