package api

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// MarketPairsLatest fetches the latest market pairs for a cryptocurrency asset.
// https://coinmarketcap.com/api/documentation/v1/#operation/getV1CryptocurrencyMarketPairsLatest
func (c *Client) MarketPairsLatest(ctx context.Context, req MarketPairsLatestRequest) ([]MarketPair, error) {
	req.ID = strings.TrimSpace(req.ID)
	req.Category = strings.TrimSpace(req.Category)
	req.Convert = strings.ToUpper(strings.TrimSpace(req.Convert))
	if req.ID == "" {
		return nil, ErrInvalidInput
	}
	if req.Limit <= 0 {
		return nil, ErrInvalidInput
	}
	if req.Category == "" {
		req.Category = "all"
	}

	params := url.Values{
		"id":       {req.ID},
		"category": {req.Category},
		"limit":    {strconv.Itoa(req.Limit)},
		"convert":  {req.Convert},
	}

	var result MarketPairsLatestResponse
	if err := c.get(ctx, "/v1/cryptocurrency/market-pairs/latest?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	return result.Data.MarketPairs, nil
}
