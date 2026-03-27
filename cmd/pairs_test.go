package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPairs_BySymbol_JSONOutput_DefaultCategory(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/cryptocurrency/map":
			assert.Equal(t, "BTC", r.URL.Query().Get("symbol"))
			_, _ = w.Write([]byte(`{
				"data":[{"id":1,"name":"Bitcoin","symbol":"BTC","slug":"bitcoin","rank":1,"is_active":1}]
			}`))
		case "/v1/cryptocurrency/market-pairs/latest":
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
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "pairs", "btc", "-o", "json")
	require.NoError(t, err)

	var out []pairView
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	require.Len(t, out, 1)
	assert.Equal(t, "BTC/USDT", out[0].Pair)
	assert.Equal(t, "Binance", out[0].Exchange)
	assert.Equal(t, "spot", out[0].Category)
	assert.Equal(t, 68000.12, out[0].Price)
	assert.Equal(t, 123456.78, out[0].Volume24h)
	assert.Equal(t, "percent", out[0].FeeType)
}

func TestPairs_MissingRequestedConvert_FailsClosed(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/cryptocurrency/map":
			assert.Equal(t, "BTC", r.URL.Query().Get("symbol"))
			_, _ = w.Write([]byte(`{
				"data":[{"id":1,"name":"Bitcoin","symbol":"BTC","slug":"bitcoin","rank":1,"is_active":1}]
			}`))
		case "/v1/cryptocurrency/market-pairs/latest":
			q := r.URL.Query()
			assert.Equal(t, "1", q.Get("id"))
			assert.Equal(t, "all", q.Get("category"))
			assert.Equal(t, "20", q.Get("limit"))
			assert.Equal(t, "EUR", q.Get("convert"))
			_, _ = w.Write([]byte(`{
				"data":{
					"id":1,
					"name":"Bitcoin",
					"symbol":"BTC",
					"slug":"bitcoin",
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
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, _, err := executeCommand(t, "pairs", "btc", "--convert", "EUR", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `requested convert "EUR" is not available`)
}

func TestPairs_PositionalSlugFirstDerivatives_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/cryptocurrency/info":
			assert.Equal(t, "bitcoin", r.URL.Query().Get("slug"))
			_, _ = w.Write([]byte(`{
				"data":{"1":{"id":1,"name":"Bitcoin","symbol":"BTC","slug":"bitcoin"}}
			}`))
		case "/v1/cryptocurrency/market-pairs/latest":
			q := r.URL.Query()
			assert.Equal(t, "1", q.Get("id"))
			assert.Equal(t, "derivatives", q.Get("category"))
			assert.Equal(t, "50", q.Get("limit"))
			assert.Equal(t, "USD", q.Get("convert"))
			_, _ = w.Write([]byte(`{
				"data":{
					"id":1,
					"name":"Bitcoin",
					"symbol":"BTC",
					"slug":"bitcoin",
					"market_pairs":[
						{
							"market_pair":"BTC/USDC",
							"exchange":{"name":"Coinbase Derivatives","slug":"coinbase-derivatives"},
							"category":"derivatives",
							"fee_type":"flat",
							"market_pair_base":{"id":1,"name":"Bitcoin","symbol":"BTC","slug":"bitcoin"},
							"market_pair_quote":{"id":2781,"name":"USD Coin","symbol":"USDC","slug":"usd-coin"},
							"quote":{"USD":{"price":68100.5,"volume_24h":987654.32}}
						}
					]
				}
			}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "pairs", "bitcoin", "--category", "derivatives", "--limit", "50", "-o", "json")
	require.NoError(t, err)

	var out []pairView
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	require.Len(t, out, 1)
	assert.Equal(t, "BTC/USDC", out[0].Pair)
	assert.Equal(t, "Coinbase Derivatives", out[0].Exchange)
	assert.Equal(t, "derivatives", out[0].Category)
	assert.Equal(t, 68100.5, out[0].Price)
	assert.Equal(t, 987654.32, out[0].Volume24h)
	assert.Equal(t, "flat", out[0].FeeType)
}

func TestPairs_AmbiguousSymbolShorthandWarnsAndChoosesHighestRanked(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/cryptocurrency/map":
			assert.Equal(t, "BTC", r.URL.Query().Get("symbol"))
			_, _ = w.Write([]byte(`{
				"data":[
					{"id":12345,"name":"Bitcoin Clone","symbol":"BTC","slug":"bitcoin-clone","rank":250,"is_active":1},
					{"id":1,"name":"Bitcoin","symbol":"BTC","slug":"bitcoin","rank":1,"is_active":1}
				]
			}`))
		case "/v1/cryptocurrency/market-pairs/latest":
			assert.Equal(t, "1", r.URL.Query().Get("id"))
			_, _ = w.Write([]byte(`{
				"data":{
					"id":1,
					"name":"Bitcoin",
					"symbol":"BTC",
					"slug":"bitcoin",
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
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, stderr, err := executeCommand(t, "pairs", "btc", "-o", "json")
	require.NoError(t, err)

	var out []pairView
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	require.Len(t, out, 1)
	assert.Equal(t, "BTC/USDT", out[0].Pair)
	assert.Contains(t, stderr, `Warning: symbol "BTC" matched multiple assets`)
	assert.Contains(t, stderr, "selected top-ranked candidate Bitcoin")
}

func TestPairs_TableOutputColumns(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/cryptocurrency/map":
			_, _ = w.Write([]byte(`{
				"data":[{"id":1,"name":"Bitcoin","symbol":"BTC","slug":"bitcoin","rank":1,"is_active":1}]
			}`))
		case "/v1/cryptocurrency/market-pairs/latest":
			_, _ = w.Write([]byte(`{
				"data":{
					"id":1,
					"name":"Bitcoin",
					"symbol":"BTC",
					"slug":"bitcoin",
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
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "pairs", "btc")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Pair")
	assert.Contains(t, stdout, "Exchange")
	assert.Contains(t, stdout, "Category")
	assert.Contains(t, stdout, "Price")
	assert.Contains(t, stdout, "Volume 24h")
	assert.Contains(t, stdout, "Fee Type")
}

func TestPairs_DryRun_PositionalPlan(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("dry-run should not make HTTP calls")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "pairs", "btc", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var plan pairsPositionalDryRunPlan
	require.NoError(t, json.Unmarshal([]byte(stdout), &plan))
	assert.Equal(t, "pairs", plan.Command)
	assert.Equal(t, "positional_shorthand", plan.Mode)
	assert.Equal(t, []string{"btc"}, plan.Inputs)
	assert.Equal(t, "all", plan.Category)
	assert.Equal(t, "USD", plan.Convert)
	assert.Equal(t, 20, plan.Limit)
	require.Len(t, plan.Steps, 3)
	assert.Contains(t, plan.Steps[0].Request.URL, "/v1/cryptocurrency/map")
	assert.Equal(t, "BTC", plan.Steps[0].Request.Params["symbol"])
	assert.Contains(t, plan.Steps[1].Request.URL, "/v2/cryptocurrency/info")
	assert.Equal(t, "btc", plan.Steps[1].Request.Params["slug"])
	assert.Contains(t, plan.Steps[2].Request.URL, "/v1/cryptocurrency/market-pairs/latest")
	assert.Equal(t, "<resolved asset id>", plan.Steps[2].Request.Params["id"])
	assert.Equal(t, "all", plan.Steps[2].Request.Params["category"])
	assert.Equal(t, "20", plan.Steps[2].Request.Params["limit"])
}
