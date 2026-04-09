---
name: claude-code-cmc-cli
description: Use when preparing Claude Code prompts or workflows that should leverage the CMC CLI skill.
version: 0.1.0
metadata:
  requires:
    bins:
      - cmc
    env:
      - CMC_API_KEY
  install: |
    go install github.com/openCMC/CoinMarketCap-CLI@latest
    mv "$(go env GOPATH)/bin/coinmarketcap-cli" "$(go env GOPATH)/bin/cmc"
---

# Claude Code + CMC CLI

Use this as a thin entrypoint for Claude Code. Do not duplicate the CMC CLI reference; route command choice, flags, and output expectations to [cmc-cli](../cmc-cli/SKILL.md).

## Prerequisites

This skill requires:
- `cmc` CLI installed and available on PATH — see [CoinMarketCap CLI](https://github.com/openCMC/CoinMarketCap-CLI) for installation options
- `CMC_API_KEY` environment variable set with a valid CoinMarketCap API key
- Authentication completed via `cmc auth`

If either dependency is missing, the skill will not function.

## Guidance

- Prefer short, explicit prompts that name the user intent.
- Prefer the smallest `cmc` command that answers the request.
- Prefer compact JSON for machine use and table output for human review.
- When unsure, read the core `cmc-cli` skill first.
