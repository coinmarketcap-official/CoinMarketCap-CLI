---
name: openclaw-cmc-cli
description: Use when preparing OpenClaw prompts or workflows that should leverage the CMC CLI skill.
---

# OpenClaw + CMC CLI

Use this as a thin entrypoint for OpenClaw. Keep it aligned with the core [cmc-cli](../cmc-cli/SKILL.md) skill rather than duplicating command guidance.

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

- Frame requests in terms of the user intent, not the endpoint.
- Prefer shipped `cmc` commands over ad hoc API calls.
- Use compact JSON for automation and `-o table` for human review.
- Read [cmc-cli](../cmc-cli/SKILL.md) for the authoritative command map.
