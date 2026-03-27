package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) BlockchainStatisticsLatestByIDs(ctx context.Context, ids []string) (map[string]BlockchainStatistics, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("blockchain statistics not found (%w)", ErrAssetNotFound)
	}
	return c.blockchainStatisticsByQuery(ctx, url.Values{"id": {strings.Join(ids, ",")}})
}

func (c *Client) blockchainStatisticsByQuery(ctx context.Context, params url.Values) (map[string]BlockchainStatistics, error) {
	var result BlockchainStatisticsLatestResponse
	if err := c.get(ctx, "/v1/blockchain/statistics/latest?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("blockchain statistics not found (%w)", ErrAssetNotFound)
	}
	return result.Data, nil
}
