---
name: openclaw-cmc-market-report
description: Use for OpenClaw-only crypto market report requests such as market report, market snapshot, crypto morning brief, how's crypto today, or today's crypto market summary; produce the fixed report format using shipped cmc commands only.
---

# OpenClaw CMC Market Report

Use this skill only for OpenClaw-driven market report workflows. Keep it aligned with [cmc-cli](../cmc-cli/SKILL.md) and [openclaw-cmc-cli](../openclaw-cmc-cli/SKILL.md); do not restate the command encyclopedia here.

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

## Trigger

Apply when the request matches or clearly implies:
- market report
- market snapshot
- crypto morning brief
- how's crypto today
- today's crypto market summary

## Scope

- Use only shipped `cmc` commands for this report.
- Default commands:
  - `metrics`
  - `price --id 1,1027`
  - `trending`
  - `top-gainers-losers --time-period 24h`
  - `news --limit 5`
- Use `pairs <id> --category derivatives` only when the user explicitly asks for derivatives or an asset-specific derivatives extension.
- Do not switch to MCP or raw Pro API inside this skill.
- Treat `resolve` and `history` as higher-caution, non-default dependencies for this report. Use them only if the request truly needs identity disambiguation or historical context.

## Report Format

Always output these sections in this order:

1. Market Snapshot
2. BTC & ETH
3. Momentum
4. News Flow
5. Risks / Caveats

## Section Rules

- Market Snapshot: summarize the broad tape from `metrics`, `trending`, and `top-gainers-losers`.
- BTC & ETH: anchor on `price --id 1,1027`; keep it tight and comparative.
- Momentum: focus on the strongest movers and what the 24h setup suggests.
- News Flow: use only the top 5 news items and keep it signal-first.
- Risks / Caveats: call out data gaps, stale prints, noisy headlines, and any derivatives caveat if `pairs` was used.

## Failure Handling

- If a command fails, return the affected section with a short partial note instead of dropping the whole report.
- If a command returns incomplete data, say what is missing, use the available subset, and continue to the next section.
- If BTC/ETH data is unavailable, state that explicitly and still complete the remaining sections.

## Style

- Be concise and instruction-focused.
- Prefer report-ready output over explanation.
- Keep the tone suitable for OpenClaw task execution, not general assistant chatter.
