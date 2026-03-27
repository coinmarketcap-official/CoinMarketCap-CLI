package api

import (
	"context"
	"sort"
	"strconv"
	"strings"
)

const (
	dexSearchConvertID = "2781"
	dexSearchPageLimit = 100
	dexSearchMaxPages  = 5
)

type SearchResult struct {
	Kind      string  `json:"kind"`
	Chain     string  `json:"chain"`
	ID        int64   `json:"id"`
	Slug      string  `json:"slug"`
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	DEX       string  `json:"dex"`
	Pair      string  `json:"pair"`
	Liquidity float64 `json:"liquidity"`
	Volume24h float64 `json:"volume_24h"`
	Rank      int     `json:"rank"`
}

type rankedSearchResult struct {
	result        SearchResult
	matchStrength int
}

func (c *Client) SearchAssets(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, ErrInvalidInput
	}
	if limit <= 0 {
		return nil, ErrInvalidInput
	}

	assets, err := c.resolveMany(ctx, map[string]string{})
	if err != nil {
		return nil, err
	}

	lowerQuery := strings.ToLower(query)
	ranked := make([]rankedSearchResult, 0, len(assets))
	for _, asset := range assets {
		matchStrength := assetMatchStrength(lowerQuery, asset)
		if matchStrength == -1 {
			continue
		}
		ranked = append(ranked, rankedSearchResult{
			matchStrength: matchStrength,
			result: SearchResult{
				Kind:      "asset",
				Chain:     "",
				ID:        asset.ID,
				Slug:      asset.Slug,
				Symbol:    asset.Symbol,
				Name:      asset.Name,
				Address:   "",
				DEX:       "",
				Pair:      "",
				Liquidity: 0,
				Volume24h: 0,
				Rank:      asset.Rank,
			},
		})
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].matchStrength != ranked[j].matchStrength {
			return ranked[i].matchStrength < ranked[j].matchStrength
		}
		if ranked[i].result.Rank == ranked[j].result.Rank {
			return ranked[i].result.Name < ranked[j].result.Name
		}
		if ranked[i].result.Rank == 0 {
			return false
		}
		if ranked[j].result.Rank == 0 {
			return true
		}
		return ranked[i].result.Rank < ranked[j].result.Rank
	})

	if len(ranked) > limit {
		ranked = ranked[:limit]
	}

	results := make([]SearchResult, len(ranked))
	for i, item := range ranked {
		results[i] = item.result
	}
	return results, nil
}

func (c *Client) SearchByAddress(ctx context.Context, chainValue, address string, limit int) ([]SearchResult, error) {
	chainValue = strings.TrimSpace(chainValue)
	address = strings.TrimSpace(address)
	if chainValue == "" || address == "" {
		return nil, ErrInvalidInput
	}
	if limit <= 0 {
		return nil, ErrInvalidInput
	}

	selector, err := ResolveDEXNetworkSelector(chainValue)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, limit)
	seen := map[string]struct{}{}

	pairs, err := c.DEXPairQuotesLatest(ctx, DEXPairLookupRequest{
		NetworkID:       selector.NetworkID,
		NetworkSlug:     selector.NetworkSlug,
		ContractAddress: address,
		ConvertID:       dexSearchConvertID,
	})
	if err != nil {
		return nil, err
	}
	for _, pair := range pairs {
		result := dexPairToSearchResult(pair)
		key := "pair:" + strings.ToLower(result.Address)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		results = append(results, result)
	}

	if len(results) >= limit {
		return results[:limit], nil
	}

	tokenResults := make([]SearchResult, 0, limit)
	for _, dexSlug := range DEXSearchCandidateSlugs(selector.NetworkSlug) {
		if len(tokenResults) >= limit {
			break
		}
		scrollID := ""
		for page := 0; page < dexSearchMaxPages && len(tokenResults) < limit; page++ {
			resp, err := c.DEXSpotPairsLatest(ctx, DEXSpotPairsLatestRequest{
				NetworkID:   selector.NetworkID,
				NetworkSlug: selector.NetworkSlug,
				DEXSlug:     dexSlug,
				ScrollID:    scrollID,
				Limit:       dexSearchPageLimit,
				Sort:        "liquidity",
				SortDir:     "desc",
				ConvertID:   dexSearchConvertID,
			})
			if err != nil {
				break
			}
			if resp == nil || len(resp.Data) == 0 {
				break
			}
			for _, pair := range resp.Data {
				matchedAsset, ok := dexPairMatchesTokenAddress(pair, address)
				if !ok {
					continue
				}
				result := dexTokenToSearchResult(pair, matchedAsset)
				key := "token:" + strings.ToLower(result.Pair) + ":" + strings.ToLower(result.Address)
				if _, exists := seen[key]; exists {
					continue
				}
				seen[key] = struct{}{}
				tokenResults = append(tokenResults, result)
			}
			nextScroll := resp.ScrollID
			if nextScroll == "" || nextScroll == scrollID {
				break
			}
			scrollID = nextScroll
		}
	}

	sort.SliceStable(tokenResults, func(i, j int) bool {
		if tokenResults[i].Liquidity != tokenResults[j].Liquidity {
			return tokenResults[i].Liquidity > tokenResults[j].Liquidity
		}
		if tokenResults[i].Volume24h != tokenResults[j].Volume24h {
			return tokenResults[i].Volume24h > tokenResults[j].Volume24h
		}
		if tokenResults[i].Rank != tokenResults[j].Rank {
			if tokenResults[i].Rank == 0 {
				return false
			}
			if tokenResults[j].Rank == 0 {
				return true
			}
			return tokenResults[i].Rank < tokenResults[j].Rank
		}
		return tokenResults[i].Pair < tokenResults[j].Pair
	})

	results = append(results, tokenResults...)
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func DEXSearchCandidateSlugs(networkSlug string) []string {
	switch normalizeDEXNetworkSlug(networkSlug) {
	case "ethereum":
		return []string{"uniswap-v3", "uniswap-v2", "sushiswap-v3", "curve", "balancer-v2"}
	case "base":
		return []string{"aerodrome", "uniswap-v3", "uniswap-v2"}
	case "bsc", "binance-smart-chain":
		return []string{"pancakeswap-v3", "pancakeswap-v2", "uniswap-v3"}
	case "arbitrum":
		return []string{"uniswap-v3", "camelot", "sushiswap-v3"}
	case "optimism":
		return []string{"uniswap-v3", "velodrome-v2", "sushiswap-v3"}
	case "polygon":
		return []string{"quickswap-v3", "uniswap-v3", "sushiswap-v3"}
	case "solana":
		return []string{"raydium", "orca", "meteora"}
	default:
		return []string{"uniswap-v3", "pancakeswap-v3", "aerodrome", "raydium", "orca"}
	}
}

type DEXNetworkSelector struct {
	NetworkID   string
	NetworkSlug string
}

func (s DEXNetworkSelector) ParamKey() string {
	if s.NetworkID != "" {
		return "network_id"
	}
	return "network_slug"
}

func (s DEXNetworkSelector) ParamValue() string {
	if s.NetworkID != "" {
		return s.NetworkID
	}
	return s.NetworkSlug
}

func ResolveDEXNetworkSelector(value string) (DEXNetworkSelector, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return DEXNetworkSelector{}, ErrInvalidInput
	}
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return DEXNetworkSelector{NetworkID: value}, nil
	}
	slug := normalizeDEXNetworkSlug(value)
	if slug == "" {
		return DEXNetworkSelector{}, ErrInvalidInput
	}
	return DEXNetworkSelector{NetworkSlug: slug}, nil
}

func assetMatchStrength(query string, asset ResolvedAsset) int {
	fields := []string{
		strings.ToLower(asset.Symbol),
		strings.ToLower(asset.Slug),
		strings.ToLower(asset.Name),
	}
	for _, field := range fields {
		if field != "" && field == query {
			return 0
		}
	}
	for _, field := range fields {
		if field != "" && strings.HasPrefix(field, query) {
			return 1
		}
	}
	for _, field := range fields {
		if field != "" && strings.Contains(field, query) {
			return 2
		}
	}
	return -1
}

func dexPairMatchesTokenAddress(pair DEXPair, address string) (DEXAsset, bool) {
	switch {
	case strings.EqualFold(pair.BaseAsset.ContractAddress, address):
		return pair.BaseAsset, true
	case strings.EqualFold(pair.QuoteAsset.ContractAddress, address):
		return pair.QuoteAsset, true
	default:
		return DEXAsset{}, false
	}
}

func dexPairToSearchResult(pair DEXPair) SearchResult {
	quote := pair.QuoteFor(dexSearchConvertID)
	return SearchResult{
		Kind:      "pair",
		Chain:     normalizeDEXNetworkSlug(pair.NetworkSlug),
		ID:        0,
		Slug:      "",
		Symbol:    pair.PairLabel(),
		Name:      pair.PairLabel(),
		Address:   pair.ContractAddress,
		DEX:       pair.DEXSlug,
		Pair:      pair.PairLabel(),
		Liquidity: quote.Liquidity,
		Volume24h: quote.Volume24h,
		Rank:      0,
	}
}

func dexTokenToSearchResult(pair DEXPair, asset DEXAsset) SearchResult {
	quote := pair.QuoteFor(dexSearchConvertID)
	return SearchResult{
		Kind:      "token_contract",
		Chain:     normalizeDEXNetworkSlug(pair.NetworkSlug),
		ID:        asset.ID,
		Slug:      asset.Slug,
		Symbol:    asset.Symbol,
		Name:      asset.Name,
		Address:   asset.ContractAddress,
		DEX:       pair.DEXSlug,
		Pair:      pair.PairLabel(),
		Liquidity: quote.Liquidity,
		Volume24h: quote.Volume24h,
		Rank:      0,
	}
}

func normalizeDEXNetworkSlug(value string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if b.Len() > 0 && !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
