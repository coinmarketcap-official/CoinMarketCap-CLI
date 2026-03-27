package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// DEX endpoints are implemented from CoinMarketCap's documented DEX suite
// (Academy launch article + public Postman collection). Live probing was
// unstable in this environment, so request builders intentionally stick to the
// narrow parameter set confirmed there.

func (c *Client) DEXNetworksList(ctx context.Context, start, limit int, sort, sortDir string) ([]DEXNetwork, error) {
	params := url.Values{
		"start": {fmt.Sprintf("%d", start)},
		"limit": {fmt.Sprintf("%d", limit)},
	}
	if sort != "" {
		params.Set("sort", sort)
	}
	if sortDir != "" {
		params.Set("sort_dir", sortDir)
	}

	var result DEXNetworksResponse
	if err := c.get(ctx, "/v4/dex/networks/list?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *Client) DEXListingsQuotes(ctx context.Context, start, limit int, sort, sortDir, dexType, convertID string) (*DEXListingsQuotesResponse, error) {
	params := url.Values{
		"start": {fmt.Sprintf("%d", start)},
		"limit": {fmt.Sprintf("%d", limit)},
	}
	if sort != "" {
		params.Set("sort", sort)
	}
	if sortDir != "" {
		params.Set("sort_dir", sortDir)
	}
	if dexType != "" {
		params.Set("type", dexType)
	}
	if convertID != "" {
		params.Set("convert_id", convertID)
	}

	var result DEXListingsQuotesResponse
	if err := c.get(ctx, "/v4/dex/listings/quotes?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DEXSpotPairsLatest(ctx context.Context, req DEXSpotPairsLatestRequest) (*DEXSpotPairsLatestResponse, error) {
	params := url.Values{}
	if req.NetworkID != "" {
		params.Set("network_id", req.NetworkID)
	}
	if req.NetworkSlug != "" {
		params.Set("network_slug", req.NetworkSlug)
	}
	if req.DEXID != "" {
		params.Set("dex_id", req.DEXID)
	}
	if req.DEXSlug != "" {
		params.Set("dex_slug", req.DEXSlug)
	}
	if req.ContractAddress != "" {
		params.Set("contract_address", req.ContractAddress)
	}
	if req.ScrollID != "" {
		params.Set("scroll_id", req.ScrollID)
	}
	if req.Limit > 0 {
		params.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.Sort != "" {
		params.Set("sort", req.Sort)
	}
	if req.SortDir != "" {
		params.Set("sort_dir", req.SortDir)
	}
	if req.ConvertID != "" {
		params.Set("convert_id", req.ConvertID)
	}

	var result DEXSpotPairsLatestResponse
	if err := c.get(ctx, "/v4/dex/spot-pairs/latest?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DEXPairQuotesLatest(ctx context.Context, req DEXPairLookupRequest) ([]DEXPair, error) {
	params := url.Values{
		"contract_address": {req.ContractAddress},
	}
	if req.NetworkID != "" {
		params.Set("network_id", req.NetworkID)
	}
	if req.NetworkSlug != "" {
		params.Set("network_slug", req.NetworkSlug)
	}
	if req.ConvertID != "" {
		params.Set("convert_id", req.ConvertID)
	}

	var result DEXPairsResponse
	if err := c.get(ctx, "/v4/dex/pairs/quotes/latest?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *Client) DEXPairsOHLCVHistorical(ctx context.Context, req DEXOHLCVHistoricalRequest) ([]DEXOHLCVPoint, error) {
	params := url.Values{
		"contract_address": {req.ContractAddress},
	}
	if req.NetworkID != "" {
		params.Set("network_id", req.NetworkID)
	}
	if req.NetworkSlug != "" {
		params.Set("network_slug", req.NetworkSlug)
	}
	if req.TimePeriod != "" {
		params.Set("time_period", req.TimePeriod)
	}
	if req.Interval != "" {
		params.Set("interval", req.Interval)
	}
	if req.TimeStart != "" {
		params.Set("time_start", req.TimeStart)
	}
	if req.TimeEnd != "" {
		params.Set("time_end", req.TimeEnd)
	}
	if req.Count > 0 {
		params.Set("count", strconv.Itoa(req.Count))
	}
	if req.ConvertID != "" {
		params.Set("convert_id", req.ConvertID)
	}

	var result DEXOHLCVHistoricalResponse
	if err := c.get(ctx, "/v4/dex/pairs/ohlcv/historical?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *Client) DEXPairsTradeLatest(ctx context.Context, req DEXTradeLatestRequest) ([]DEXTrade, error) {
	params := url.Values{
		"contract_address": {req.ContractAddress},
	}
	if req.NetworkID != "" {
		params.Set("network_id", req.NetworkID)
	}
	if req.NetworkSlug != "" {
		params.Set("network_slug", req.NetworkSlug)
	}
	if req.ConvertID != "" {
		params.Set("convert_id", req.ConvertID)
	}
	if req.Limit > 0 {
		params.Set("limit", strconv.Itoa(req.Limit))
	}

	var result DEXTradesLatestResponse
	if err := c.get(ctx, "/v4/dex/pairs/trade/latest?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func JoinDEXIDs(listings []DEXListing) string {
	ids := make([]string, 0, len(listings))
	for _, listing := range listings {
		if listing.ID > 0 {
			ids = append(ids, strconv.FormatInt(listing.ID, 10))
		}
	}
	return strings.Join(ids, ",")
}
