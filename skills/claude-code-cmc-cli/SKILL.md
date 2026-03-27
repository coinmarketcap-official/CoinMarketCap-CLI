---
name: claude-code-cmc-cli
description: Use when preparing Claude Code prompts or workflows that should leverage the CMC CLI skill.
---

# Claude Code + CMC CLI

Use this as a thin entrypoint for Claude Code. Do not duplicate the CMC CLI reference; route command choice, flags, and output expectations to [cmc-cli](../cmc-cli/SKILL.md).

## Prerequisites

Before running any `cmc` command, ensure the CLI is installed and authenticated.

```sh
# Install cmc from source (skips if already present)
if ! command -v cmc &>/dev/null; then
  go install github.com/coinmarketcap/coinmarketcap-cli@latest
  mv "$(go env GOPATH)/bin/coinmarketcap-cli" "$(go env GOPATH)/bin/cmc"
fi

# Authenticate (requires CMC_API_KEY in environment)
cmc auth
```

If `CMC_API_KEY` is not set in the environment, ask the user to provide it before proceeding.

## Guidance

- Prefer short, explicit prompts that name the user intent.
- Prefer the smallest `cmc` command that answers the request.
- Prefer compact JSON for machine use and table output for human review.
- When unsure, read the core `cmc-cli` skill first.
