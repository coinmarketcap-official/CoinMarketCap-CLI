package api

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

func (c *Client) QuotesHistoricalByID(ctx context.Context, id, convert string, timeStart, timeEnd time.Time, count int, interval string) (*HistoricalQuoteAsset, error) {
	params := url.Values{
		"id":         {id},
		"convert":    {convert},
		"time_start": {timeStart.UTC().Format(time.RFC3339)},
		"time_end":   {timeEnd.UTC().Format(time.RFC3339)},
		"count":      {fmt.Sprintf("%d", count)},
		"interval":   {interval},
	}
	var result QuotesHistoricalResponse
	if err := c.get(ctx, "/v1/cryptocurrency/quotes/historical?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	return pickHistoricalQuoteAsset(result.Data)
}

func (c *Client) OHLCVHistoricalByID(ctx context.Context, id, convert, timePeriod string, timeStart, timeEnd time.Time, count int, interval string) (*HistoricalOHLCVAsset, error) {
	params := url.Values{
		"id":          {id},
		"convert":     {convert},
		"time_period": {timePeriod},
		"time_start":  {timeStart.UTC().Format(time.RFC3339)},
		"time_end":    {timeEnd.UTC().Format(time.RFC3339)},
		"count":       {fmt.Sprintf("%d", count)},
		"interval":    {interval},
	}
	var result OHLCVHistoricalResponse
	if err := c.get(ctx, "/v2/cryptocurrency/ohlcv/historical?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	return pickHistoricalOHLCVAsset(result.Data)
}

func pickHistoricalQuoteAsset(data map[string]HistoricalQuoteAsset) (*HistoricalQuoteAsset, error) {
	for _, asset := range data {
		assetCopy := asset
		return &assetCopy, nil
	}
	return nil, ErrAssetNotFound
}

func pickHistoricalOHLCVAsset(data map[string]HistoricalOHLCVAsset) (*HistoricalOHLCVAsset, error) {
	for _, asset := range data {
		assetCopy := asset
		return &assetCopy, nil
	}
	return nil, ErrAssetNotFound
}
