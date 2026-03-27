---
name: claude-code-cmc-cli
description: Use when preparing Claude Code prompts or workflows that should leverage the CMC CLI skill.
---

# Claude Code + CMC CLI

Use this as a thin entrypoint for Claude Code. Do not duplicate the CMC CLI reference; route command choice, flags, and output expectations to [cmc-cli](../cmc-cli/SKILL.md).

## Guidance

- Prefer short, explicit prompts that name the user intent.
- Prefer the smallest `cmc` command that answers the request.
- Prefer compact JSON for machine use and table output for human review.
- When unsure, read the core `cmc-cli` skill first.
