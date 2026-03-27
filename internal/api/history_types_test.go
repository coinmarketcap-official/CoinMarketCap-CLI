package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BUG-004 regression: CMC returns data as a single object when querying by numeric ID.

func TestOHLCVHistoricalResponse_UnmarshalSingleObject(t *testing.T) {
	raw := `{
		"data": {
			"id": 1839,
			"name": "BNB",
			"symbol": "BNB",
			"quotes": [
				{
					"time_open": "2026-03-15T00:00:00.000Z",
					"time_close": "2026-03-15T23:59:59.999Z",
					"quote": {
						"USD": {
							"open": 600.0,
							"high": 610.0,
							"low": 595.0,
							"close": 605.0,
							"volume": 1000000.0,
							"market_cap": 90000000000.0,
							"timestamp": "2026-03-15T23:59:59.999Z"
						}
					}
				}
			]
		}
	}`

	var resp OHLCVHistoricalResponse
	err := json.Unmarshal([]byte(raw), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Data, 1)

	var asset HistoricalOHLCVAsset
	for _, v := range resp.Data {
		asset = v
	}
	assert.Equal(t, int64(1839), asset.ID)
	assert.Equal(t, "BNB", asset.Name)
	require.Len(t, asset.Quotes, 1)
	assert.Equal(t, "2026-03-15T00:00:00.000Z", asset.Quotes[0].TimeOpen)
	assert.Equal(t, 600.0, asset.Quotes[0].Quote["USD"].Open)
	assert.Equal(t, 605.0, asset.Quotes[0].Quote["USD"].Close)
	assert.Equal(t, 1000000.0, asset.Quotes[0].Quote["USD"].Volume)
}

func TestOHLCVHistoricalResponse_UnmarshalMapFormat(t *testing.T) {
	raw := `{
		"data": {
			"bitcoin": {
				"id": 1,
				"name": "Bitcoin",
				"symbol": "BTC",
				"slug": "bitcoin",
				"quotes": [
					{
						"time_open": "2026-03-15T00:00:00.000Z",
						"time_close": "2026-03-15T23:59:59.999Z",
						"quote": {
							"USD": {
								"open": 69000.0,
								"high": 70000.0,
								"low": 68000.0,
								"close": 69500.0,
								"volume": 50000000.0,
								"market_cap": 1300000000000.0,
								"timestamp": "2026-03-15T23:59:59.999Z"
							}
						}
					}
				]
			}
		}
	}`

	var resp OHLCVHistoricalResponse
	err := json.Unmarshal([]byte(raw), &resp)
	require.NoError(t, err)

	asset, ok := resp.Data["bitcoin"]
	require.True(t, ok)
	assert.Equal(t, int64(1), asset.ID)
	assert.Equal(t, "Bitcoin", asset.Name)
}

func TestQuotesHistoricalResponse_UnmarshalSingleObject(t *testing.T) {
	raw := `{
		"data": {
			"id": 1,
			"name": "Bitcoin",
			"symbol": "BTC",
			"slug": "bitcoin",
			"quotes": [
				{
					"timestamp": "2026-03-19T00:00:00.000Z",
					"quote": {
						"USD": {
							"price": 69800.0,
							"market_cap": 1300000000000.0,
							"volume_24h": 30000000000.0
						}
					}
				},
				{
					"timestamp": "2026-03-20T00:00:00.000Z",
					"quote": {
						"USD": {
							"price": 70200.0,
							"market_cap": 1310000000000.0,
							"volume_24h": 31000000000.0
						}
					}
				}
			]
		}
	}`

	var resp QuotesHistoricalResponse
	err := json.Unmarshal([]byte(raw), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Data, 1)

	var asset HistoricalQuoteAsset
	for _, v := range resp.Data {
		asset = v
	}
	assert.Equal(t, int64(1), asset.ID)
	assert.Equal(t, "Bitcoin", asset.Name)
	require.Len(t, asset.Quotes, 2)
	assert.Equal(t, 69800.0, asset.Quotes[0].Quote["USD"].Price)
	assert.Equal(t, 70200.0, asset.Quotes[1].Quote["USD"].Price)
}

func TestQuotesHistoricalResponse_UnmarshalMapFormat(t *testing.T) {
	raw := `{
		"data": {
			"1": {
				"id": 1,
				"name": "Bitcoin",
				"symbol": "BTC",
				"slug": "bitcoin",
				"quotes": [
					{
						"timestamp": "2026-03-19T00:00:00.000Z",
						"quote": {
							"USD": {
								"price": 69800.0,
								"market_cap": 1300000000000.0,
								"volume_24h": 30000000000.0
							}
						}
					}
				]
			}
		}
	}`

	var resp QuotesHistoricalResponse
	err := json.Unmarshal([]byte(raw), &resp)
	require.NoError(t, err)

	asset, ok := resp.Data["1"]
	require.True(t, ok)
	assert.Equal(t, int64(1), asset.ID)
}
