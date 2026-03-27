package api

import (
	"context"
	"net/url"
)

// GlobalMetricsLatest fetches /v1/global-metrics/quotes/latest.
func (c *Client) GlobalMetricsLatest(ctx context.Context, convert string) (*GlobalMetricsData, error) {
	params := url.Values{
		"convert": {convert},
	}
	var result GlobalMetricsResponse
	err := c.get(ctx, "/v1/global-metrics/quotes/latest?"+params.Encode(), &result)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}
