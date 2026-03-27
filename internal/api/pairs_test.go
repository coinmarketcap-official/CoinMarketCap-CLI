package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarketPairsLatest(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/cryptocurrency/market-pairs/latest", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, "1", q.Get("id"))
		assert.Equal(t, "all", q.Get("category"))
		assert.Equal(t, "20", q.Get("limit"))
		assert.Equal(t, "USD", q.Get("convert"))

		_, _ = w.Write([]byte(`{
			"data":{
				"id":1,
				"name":"Bitcoin",
				"symbol":"BTC",
				"slug":"bitcoin",
				"num_markets":1,
				"market_pairs":[
					{
						"market_pair":"BTC/USDT",
						"exchange":{"name":"Binance","slug":"binance"},
						"category":"spot",
						"fee_type":"percent",
						"market_pair_base":{"id":1,"name":"Bitcoin","symbol":"BTC","slug":"bitcoin"},
						"market_pair_quote":{"id":825,"name":"Tether","symbol":"USDT","slug":"tether"},
						"quote":{"USD":{"price":68000.12,"volume_24h":123456.78}}
					}
				]
			}
		}`))
	})
	defer srv.Close()

	result, err := c.MarketPairsLatest(context.Background(), MarketPairsLatestRequest{
		ID:       "1",
		Category: "all",
		Limit:    20,
		Convert:  "USD",
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "BTC/USDT", result[0].PairLabel())
	assert.Equal(t, "Binance", result[0].ExchangeLabel())
	quote, ok := result[0].QuoteFor("USD")
	require.True(t, ok)
	assert.Equal(t, 68000.12, quote.Price)
	assert.Equal(t, 123456.78, quote.Volume24h)
}

func TestMarketPairExchangeLabel_FallsBackToFlatFields(t *testing.T) {
	pair := MarketPair{
		ExchangeName: "Kraken",
		ExchangeSlug: "kraken",
	}
	assert.Equal(t, "Kraken", pair.ExchangeLabel())
}

func TestMarketPairQuoteFor_MissingConvertFailsClosed(t *testing.T) {
	pair := MarketPair{
		Quote: map[string]MarketPairQuote{
			"USD": {Price: 1.23, Volume24h: 45.6},
		},
	}
	_, ok := pair.QuoteFor("EUR")
	assert.False(t, ok)
}
