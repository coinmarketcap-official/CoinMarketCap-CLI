---
name: coin-research
description: Single cryptocurrency research workflow using cmc-cli. Use when the user asks to research, investigate, or get a comprehensive overview of a specific crypto asset. Produces a structured research report covering fundamentals, price action, market structure, sentiment, and bull/bear assessment.
metadata: {"openclaw":{"requires":{"bins":["cmc"],"env":["CMC_API_KEY"]},"primaryEnv":"CMC_API_KEY","homepage":"https://github.com/coinmarketcap-official/CoinMarketCap-CLI"}}
---

# Coin Research (cmc-cli)

Structured single-coin research using the `cmc` CLI. Derives its analytical framework from a multi-model institutional analysis pipeline, compressed into a single-pass workflow that any Claude Code user can run.

## Prerequisites

This skill requires:
- `cmc` CLI installed and available on PATH — see [CoinMarketCap CLI](https://github.com/coinmarketcap-official/CoinMarketCap-CLI) for installation options
- `CMC_API_KEY` available to Claude via the session environment

If either dependency is missing, stop and report the missing requirement.

## When to Use

- User says "research BTC", "look into SOL", "tell me about ETH", "coin overview"
- User wants a structured asset profile for a specific crypto
- User asks "what's happening with X" for a specific coin

## When NOT to Use

- Market-wide scan (no specific coin) → use `cmc markets` / `cmc trending` directly
- Stock analysis → not supported by cmc-cli
- User wants a quick price check only → use `cmc price <coin>` directly

## Research Pipeline

```
Step 1: Identity Resolution
    ↓
Step 2: Move Profile Analysis          ← compares coin vs BTC/ETH
    ↓
Step 3: Price + Fundamentals + Chain    ← single enriched call
    ↓
Step 4: Historical Price Action         ← 7d hourly + 30d daily OHLCV
    ↓
Step 5: Market Structure                ← trading pairs & liquidity
    ↓
Step 6: Market Context                  ← global metrics + top movers
    ↓
Step 7: News & Sentiment                ← news + trending
    ↓
Synthesis: Structured Research Report
```

## Execution

### Step 1: Identity Resolution

```bash
# Resolve symbol to CMC ID for deterministic subsequent calls
cmc resolve --symbol <SYMBOL> -o json

# If ambiguous or unknown, search first
cmc search <query> -o json
```

**Extract**: `cmc_id`, `slug`, `name`, `symbol`, `rank`

All subsequent commands MUST use `--id <cmc_id>` (not `--symbol`) for determinism.

If identity resolution fails or remains ambiguous, stop and report that the target asset could not be resolved cleanly.

### Step 2: Move Profile Analysis

This step determines whether the coin's recent movement is driven by the broader market or by coin-specific factors.

```bash
# Fetch BTC, ETH, and target coin prices in one call
cmc price btc eth --id <cmc_id> -o json
```

**Analysis logic** — compare `percent_change_24h` across BTC, ETH, and the target:

| Condition | Classification | Implication |
|-----------|---------------|-------------|
| Target moves in same direction as BTC/ETH, within ±2× magnitude | **Market Beta** | Price driven by macro/market forces. Focus synthesis on market context. |
| Target diverges from BTC/ETH direction OR magnitude differs >2× | **Idiosyncratic** | Coin-specific catalyst likely. Focus synthesis on news, fundamentals, project events. |

Record the classification — it guides the emphasis of the final synthesis.

### Step 3: Price + Fundamentals + Chain Stats

```bash
# Single enriched call — quotes + project info + blockchain stats
cmc price --id <cmc_id> --with-info --with-chain-stats -o json
```

**Extract from quotes**: price, market_cap, volume_24h, percent_change_1h/24h/7d/30d/90d, circulating_supply, total_supply, max_supply

**Extract from info**: description, tags (→ sector/category), date_added, urls (website, explorer, github, twitter, reddit)

**Extract from chain_stats**: consensus_mechanism, hashrate_24h, tps_24h, total_transactions, total_blocks, block_reward

Note: `--with-chain-stats` may return empty for some tokens (especially ERC-20s). Fall back gracefully.

### Step 4: Historical Price Action

```bash
# 7-day hourly candles (short-term structure)
cmc history --id <cmc_id> --days 7 --ohlc --interval hourly -o json

# 30-day daily candles (medium-term trend)
cmc history --id <cmc_id> --days 30 --ohlc -o json
```

**Analysis framework** (derived from technical analysis methodology):

1. **Trend Structure**: Identify higher highs / higher lows (uptrend) or lower highs / lower lows (downtrend). Note trend reversals.
2. **Key Levels**: Identify support (recent lows, consolidation floors) and resistance (recent highs, rejection points) from OHLCV data.
3. **Volume-Price Relationship**: Volume increasing with price movement = confirmation. Volume diverging from price = potential reversal signal.
4. **Volatility Assessment**: Compare recent daily range (high-low) to 30-day average. Rate as high / medium / low relative to the asset's own history.

### Step 5: Market Structure

```bash
# Top trading pairs across spot + derivatives
cmc pairs <slug> --category all --limit 20 -o json
```

**Analyze**:
- **Liquidity distribution**: Which exchanges hold most volume?
- **Spot vs derivatives ratio**: High derivatives ratio = speculative interest dominant
- **Pair concentration**: Volume concentrated on 1-2 exchanges = liquidity risk

### Step 6: Market Context

```bash
# Global metrics (BTC dominance, total market cap, volume)
cmc metrics -o json

# Top movers — is our asset among them?
cmc top-gainers-losers --time-period 24h --limit 20 -o json
```

**Analyze**:
- **Asset vs market**: Is the coin outperforming, underperforming, or tracking the market?
- **BTC dominance trend**: Rising dominance = risk-off (capital flowing to BTC). Falling = risk-on (alt season signal).
- **Total market cap trajectory**: Expanding or contracting liquidity environment?
- **Sector momentum**: Cross-reference the coin's tags with top movers to assess sector rotation.

### Step 7: News & Sentiment

```bash
# Latest crypto news
cmc news --limit 10 -o json

# Community trending assets
cmc trending --limit 20 -o json
```

**Analyze**:
- **Catalyst identification**: Any news directly mentioning the target asset? Upcoming events (upgrades, listings, partnerships)?
- **Narrative alignment**: Does the coin fit the current market narrative (e.g., AI, RWA, L2, meme)?
- **Social momentum**: Is the asset in the trending list? Rising or falling rank?

## Synthesis Framework

After collecting all data, synthesize using these analytical lenses (each derived from institutional analysis methodology):

### Lens 1: Fundamentals Snapshot
Evaluate the project's intrinsic value drivers:
- **What problem does it solve?** Is the solution differentiated?
- **Tokenomics health**: Supply dynamics (inflationary/deflationary), value accrual mechanism (burn, staking, fee sharing)
- **Ecosystem vitality**: Developer activity signals (GitHub links), DApp count, community size
- **Competitive position**: Where does it rank within its sector/category? Who are the main competitors?

### Lens 2: Technical State
From OHLCV data, provide a concise market structure read:
- Current trend direction and strength
- Nearest support and resistance levels with price values
- Volume trend (confirming or diverging from price?)
- Short-term vs medium-term trend alignment or conflict

### Lens 3: Money Flow & Sentiment
Based on Move Profile result + market context:
- **If Market Beta**: Emphasize macro environment, BTC dominance trend, total market health
- **If Idiosyncratic**: Emphasize project-specific catalysts, news events, social signals
- Volume/market-cap ratio as activity indicator
- Position among top gainers/losers

### Lens 4: Catalyst Scan
Identify actionable information signals:
- Specific news events or announcements
- Upcoming protocol upgrades or token unlocks (if mentioned in info/news)
- Trending narrative fit or misalignment
- Social momentum direction

### Lens 5: Bull Case / Bear Case
Structured assessment with **specific evidence** from data:
- **Bull Case** (3 points): strongest arguments for upside, each citing data
- **Bear Case** (3 points): strongest arguments for downside, each citing data
- **Verdict**: Based on evidence weight, where does the balance tilt? (Bullish / Neutral / Bearish)

## Output Format

Present findings as a structured report:

```markdown
# <Name> (<Symbol>) — Research Report

> Generated via cmc-cli | <date>

## Overview
| Metric | Value |
|--------|-------|
| **Rank** | #X |
| **Price** | $X |
| **Market Cap** | $X |
| **24h Volume** | $X (Vol/MCap: X%) |
| **24h / 7d / 30d Change** | X% / X% / X% |
| **Circulating / Max Supply** | X / X |

## Move Profile
**Classification**: Market Beta / Idiosyncratic
- Target 24h: X% | BTC 24h: X% | ETH 24h: X%
- <interpretation of what's driving the current move>

## Fundamentals
- **Description**: <1-2 sentences>
- **Sector/Tags**: <from CMC tags>
- **Tokenomics**: <supply model, value accrual, notable features>
- **Chain Stats**: <consensus, hashrate/TPS if available>
- **Key Links**: website | explorer | github | twitter

## Price Action
### Short-term (7d)
- **Trend**: <direction + key levels>
- **Volume pattern**: <confirming/diverging>

### Medium-term (30d)
- **Trend**: <direction + key levels>
- **Volatility**: <high/medium/low vs own history>

## Market Structure
- **Top exchanges**: <top 3 by volume>
- **Spot/Derivatives split**: <ratio>
- **Liquidity assessment**: <concentrated/distributed>

## Market Context
- **BTC Dominance**: X% (<trending direction>)
- **Total Market Cap**: $X (<expanding/contracting>)
- **Asset vs Market**: <outperforming/underperforming/in-line>

## News & Sentiment
- **Key headlines**: <top 2-3 relevant items>
- **Trending status**: <in trending list? rank?>
- **Narrative fit**: <aligned with current market theme?>
- **Catalyst watch**: <upcoming events if any>

## Assessment

### Bull Case
1. <point with evidence>
2. <point with evidence>
3. <point with evidence>

### Bear Case
1. <point with evidence>
2. <point with evidence>
3. <point with evidence>

### Verdict
<Bullish / Neutral / Bearish> — <1-2 sentence summary of evidence balance>

---
> ⚠️ This report is for informational purposes only and does not constitute investment advice. Cryptocurrency markets are highly volatile — always conduct your own due diligence.
```

## Parallel Execution Guide

After Step 1 (identity resolution), Steps 2-7 have no dependencies between them. Run all CLI commands in parallel for faster execution:

```
Step 1 (sequential — needed for cmc_id)
    ↓
Steps 2-7 (all in parallel)
    ↓
Synthesis (sequential — needs all data)
```

## Tips

- Always use `--id` over `--symbol` after resolution — avoids symbol ambiguity
- Use `-o json` for all data fetching — structured data for better analysis
- If `--with-chain-stats` returns empty, skip Chain Stats section gracefully
- Keep the synthesis concise and evidence-based — cite specific data points, don't make unsupported claims
- Move Profile is the most important analytical step — it determines the lens through which everything else is interpreted
- For coins not in the pre-mapped list (BTC, ETH, SOL, etc.), `cmc search` + `cmc resolve` will find them
