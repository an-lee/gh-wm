---
description: Answer questions via slash command.

on:
  slash_command:
    name: agent

timeout-minutes: 10

safe-outputs:
  add-comment:
    max: 5

engine: claude
---

# Answer

Answer the question or fulfill the request in the comment thread.

## Safe output (required)

Before exiting, write JSON to **`WM_OUTPUT_FILE`**: `{"items":[{"type":"add_comment","body":"…"}]}` (or **`noop`** if nothing to post).
