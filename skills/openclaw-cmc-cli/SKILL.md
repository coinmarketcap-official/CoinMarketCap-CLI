---
name: openclaw-cmc-cli
description: Use when preparing OpenClaw prompts or workflows that should leverage the CMC CLI skill.
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

# OpenClaw + CMC CLI

Use this as a thin entrypoint for OpenClaw. Keep it aligned with the core [cmc-cli](../cmc-cli/SKILL.md) skill rather than duplicating command guidance.

## Prerequisites

This skill requires:
- `cmc` CLI installed and available on PATH — see [CoinMarketCap CLI](https://github.com/openCMC/CoinMarketCap-CLI) for installation options
- `CMC_API_KEY` environment variable set with a valid CoinMarketCap API key
- Authentication completed via `cmc auth`

If either dependency is missing, the skill will not function.

## Guidance

- Frame requests in terms of the user intent, not the endpoint.
- Prefer shipped `cmc` commands over ad hoc API calls.
- Use compact JSON for automation and `-o table` for human review.
- Read [cmc-cli](../cmc-cli/SKILL.md) for the authoritative command map.
