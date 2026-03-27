package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDEXNetworksList(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v4/dex/networks/list", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, "1", q.Get("start"))
		assert.Equal(t, "50", q.Get("limit"))
		assert.Equal(t, "id", q.Get("sort"))
		assert.Equal(t, "asc", q.Get("sort_dir"))

		_, _ = w.Write([]byte(`{
			"data":[
				{"id":1,"name":"Ethereum","slug":"ethereum"},
				{"id":8453,"name":"Base","slug":"base"}
			]
		}`))
	})
	defer srv.Close()

	result, err := c.DEXNetworksList(context.Background(), 1, 50, "id", "asc")
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, "ethereum", result[0].Slug)
}

func TestDEXListingsQuotes(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v4/dex/listings/quotes", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, "1", q.Get("start"))
		assert.Equal(t, "25", q.Get("limit"))
		assert.Equal(t, "volume_24h", q.Get("sort"))
		assert.Equal(t, "desc", q.Get("sort_dir"))
		assert.Equal(t, "all", q.Get("type"))

		_, _ = w.Write([]byte(`{
			"data":[
				{"id":1,"name":"Uniswap V3","slug":"uniswap-v3","num_markets":1200,"market_share":11.2,"quote":{"2781":{"volume_24h":500000000}}}
			]
		}`))
	})
	defer srv.Close()

	result, err := c.DEXListingsQuotes(context.Background(), 1, 25, "volume_24h", "desc", "all", "2781")
	require.NoError(t, err)
	require.Len(t, result.Data, 1)
	assert.Equal(t, "Uniswap V3", result.Data[0].Name)
	assert.Equal(t, 500000000.0, result.Data[0].Quote["2781"].Volume24h)
}

func TestDEXSpotPairsLatest(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v4/dex/spot-pairs/latest", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, "1", q.Get("network_id"))
		assert.Equal(t, "50", q.Get("limit"))
		assert.Equal(t, "volume_24h", q.Get("sort"))
		assert.Equal(t, "desc", q.Get("sort_dir"))
		assert.Equal(t, "2781", q.Get("convert_id"))

		_, _ = w.Write([]byte(`{
			"data":[
				{
					"contract_address":"0xpair1",
					"network_id":"1",
					"network_name":"Ethereum",
					"network_slug":"Ethereum",
					"dex_id":"1348",
					"dex_name":"Uniswap V3",
					"dex_slug":"uniswap-v3",
					"base_asset_name":"USD Coin",
					"base_asset_symbol":"USDC",
					"base_asset_contract_address":"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
					"quote_asset_name":"Wrapped Ether",
					"quote_asset_symbol":"WETH",
					"quote_asset_contract_address":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
					"quote":[{"convert_id":"2781","price":3500,"volume_24h":1000000,"liquidity":5000000,"percent_change_24h":1.2,"no_of_transactions_24h":120}],
					"scroll_id":"scroll-next"
				}
			]
		}`))
	})
	defer srv.Close()

	result, err := c.DEXSpotPairsLatest(context.Background(), DEXSpotPairsLatestRequest{
		NetworkID: "1",
		Limit:     50,
		Sort:      "volume_24h",
		SortDir:   "desc",
		ConvertID: "2781",
	})
	require.NoError(t, err)
	require.Len(t, result.Data, 1)
	assert.Equal(t, "0xpair1", result.Data[0].ContractAddress)
	assert.Equal(t, "USDC", result.Data[0].BaseAsset.Symbol)
	assert.Equal(t, "WETH", result.Data[0].QuoteAsset.Symbol)
	assert.Equal(t, 5000000.0, result.Data[0].QuoteFor("2781").Liquidity)
	assert.Equal(t, "scroll-next", result.ScrollID)
}

func TestDEXPairQuotesLatest(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v4/dex/pairs/quotes/latest", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, "1", q.Get("network_id"))
		assert.Equal(t, "0xpair1", q.Get("contract_address"))
		assert.Equal(t, "2781", q.Get("convert_id"))

		_, _ = w.Write([]byte(`{
			"data":[
				{
					"contract_address":"0xpair1",
					"network_id":1,
					"network_name":"Ethereum",
					"network_slug":"ethereum",
					"dex_id":10,
					"dex_name":"Uniswap V3",
					"dex_slug":"uniswap-v3",
					"base_asset":{"id":1027,"symbol":"WETH","name":"Wrapped Ether","contract_address":"0xbase"},
					"quote_asset":{"id":3408,"symbol":"USDC","name":"USD Coin","contract_address":"0xquote"},
					"quote":{"2781":{"price":3500,"volume_24h":1000000,"liquidity":5000000,"percent_change_24h":1.2,"no_of_transactions_24h":120}}
				}
			]
		}`))
	})
	defer srv.Close()

	result, err := c.DEXPairQuotesLatest(context.Background(), DEXPairLookupRequest{
		NetworkID:       "1",
		ContractAddress: "0xpair1",
		ConvertID:       "2781",
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Uniswap V3", result[0].DEXName)
}

func TestDEXPairsOHLCVHistorical(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v4/dex/pairs/ohlcv/historical", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, "1", q.Get("network_id"))
		assert.Equal(t, "0xpair1", q.Get("contract_address"))
		assert.Equal(t, "1h", q.Get("interval"))
		assert.Equal(t, "hourly", q.Get("time_period"))
		assert.Equal(t, "2026-03-20T00:00:00Z", q.Get("time_start"))
		assert.Equal(t, "2026-03-21T00:00:00Z", q.Get("time_end"))
		assert.Equal(t, "24", q.Get("count"))
		assert.Equal(t, "2781", q.Get("convert_id"))

		_, _ = w.Write([]byte(`{
			"data":[
				{
					"time_open":"2026-03-20T00:00:00Z",
					"time_close":"2026-03-20T00:59:59Z",
					"quote":{"2781":{"open":3400,"high":3500,"low":3300,"close":3450,"volume_24h":100000,"liquidity":2000000}}
				}
			]
		}`))
	})
	defer srv.Close()

	result, err := c.DEXPairsOHLCVHistorical(context.Background(), DEXOHLCVHistoricalRequest{
		NetworkID:       "1",
		ContractAddress: "0xpair1",
		TimePeriod:      "hourly",
		Interval:        "1h",
		TimeStart:       "2026-03-20T00:00:00Z",
		TimeEnd:         "2026-03-21T00:00:00Z",
		Count:           24,
		ConvertID:       "2781",
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, 3450.0, result[0].Quote["2781"].Close)
}

func TestDEXPairsTradeLatest(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v4/dex/pairs/trade/latest", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, "1", q.Get("network_id"))
		assert.Equal(t, "0xpair1", q.Get("contract_address"))
		assert.Equal(t, "25", q.Get("limit"))
		assert.Equal(t, "2781", q.Get("convert_id"))

		_, _ = w.Write([]byte(`{
			"data":[
				{
					"trade_timestamp":"2026-03-21T01:02:03Z",
					"type":"buy",
					"transaction_hash":"0xhash",
					"quote":{"2781":{"price":3451.23,"volume_24h":1000}}
				}
			]
		}`))
	})
	defer srv.Close()

	result, err := c.DEXPairsTradeLatest(context.Background(), DEXTradeLatestRequest{
		NetworkID:       "1",
		ContractAddress: "0xpair1",
		Limit:           25,
		ConvertID:       "2781",
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "buy", result[0].Type)
	assert.Equal(t, "0xhash", result[0].TransactionHash)
}
