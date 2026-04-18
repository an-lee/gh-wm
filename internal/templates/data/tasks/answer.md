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

Answer the question or fulfill the request in the comment thread. Be concise and cite code or paths when you reference behavior.

## 1. Research

- Read the triggering comment and thread context.
- Use search and file reads to ground answers in **this** repository (avoid generic guesses).

## 2. Formulate

- Write a clear answer: what, why, and where in the repo (if applicable).
- If something is ambiguous or missing, say what you would need to know.

## 3. Safe output (required)

Post your reply with **`gh wm emit`**:

`gh wm emit add-comment --body "…"`

If nothing should be posted to GitHub (e.g. empty or invalid trigger): `gh wm emit noop --message "…"`.
