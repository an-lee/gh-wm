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

wm:
  state_labels:
    working: "agent:working"
    done: "agent:review"
    failed: "agent:failed"
---

# Implement Feature

You are a developer agent working on this repository.

Implement the feature or fix described in the issue.
Create tests if applicable. Commit with clear messages.

Follow repository conventions and any project agent guide your team maintains.

## Safe output (required)

Before exiting, write JSON to **`WM_OUTPUT_FILE`**: `{"items":[...]}`. Use **`create_pull_request`** when you have commits to push as a PR, **`add_comment`** to summarize on the issue, and/or **`noop`** with a message if no GitHub follow-up is needed.
