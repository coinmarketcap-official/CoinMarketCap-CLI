---
name: openclaw-cmc-cli
description: Use when preparing OpenClaw prompts or workflows that should leverage the CMC CLI skill.
---

# OpenClaw + CMC CLI

Use this as a thin entrypoint for OpenClaw. Keep it aligned with the core [cmc-cli](../cmc-cli/SKILL.md) skill rather than duplicating command guidance.

## Guidance

- Frame requests in terms of the user intent, not the endpoint.
- Prefer shipped `cmc` commands over ad hoc API calls.
- Use compact JSON for automation and `-o table` for human review.
- Read [cmc-cli](../cmc-cli/SKILL.md) for the authoritative command map.
