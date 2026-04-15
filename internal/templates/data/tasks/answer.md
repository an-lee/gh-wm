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
