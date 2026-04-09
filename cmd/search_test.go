package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/openCMC/CoinMarketCap-CLI/internal/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch_DryRun(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call in dry-run mode")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "resolve", "--symbol", "BTC", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Equal(t, "GET", out.Method)
	assert.Equal(t, "BTC", out.Params["symbol"])
	assert.Contains(t, out.URL, "/v1/cryptocurrency/map")
}

func TestResolve_BySlug_DryRun_UsesInfoEndpoint(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call in dry-run mode")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "resolve", "--slug", "bitcoin", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Equal(t, "GET", out.Method)
	assert.Equal(t, "bitcoin", out.Params["slug"])
	assert.Contains(t, out.URL, "/v2/cryptocurrency/info")
	assert.Equal(t, "getV2CryptocurrencyInfo", out.OASOperationID)
}

func TestResolve_ByID_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/cryptocurrency/info", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("id"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"1": map[string]any{
					"id":   1,
					"name": "Bitcoin",
					"symbol": "BTC",
					"slug": "bitcoin",
				},
			},
		})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "resolve", "--id", "1", "-o", "json")
	require.NoError(t, err)

	var assets []api.ResolvedAsset
	require.NoError(t, json.Unmarshal([]byte(stdout), &assets))
	require.Len(t, assets, 1)
	assert.Equal(t, int64(1), assets[0].ID)
	assert.Equal(t, "bitcoin", assets[0].Slug)
}

func TestResolve_BySlug_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/cryptocurrency/info", r.URL.Path)
		assert.Equal(t, "bitcoin", r.URL.Query().Get("slug"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"1": map[string]any{
					"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin", "category": "coin",
				},
			},
		})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "resolve", "--slug", "bitcoin", "-o", "json")
	require.NoError(t, err)

	var assets []api.ResolvedAsset
	require.NoError(t, json.Unmarshal([]byte(stdout), &assets))
	require.Len(t, assets, 1)
	assert.Equal(t, int64(1), assets[0].ID)
	assert.Equal(t, "bitcoin", assets[0].Slug)
	assert.Equal(t, "BTC", assets[0].Symbol)
	assert.Equal(t, "Bitcoin", assets[0].Name)
	assert.Zero(t, assets[0].Rank)
	assert.False(t, assets[0].IsActive)
}

func TestSearch_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/cryptocurrency/map", r.URL.Path)
		assert.Equal(t, "btc", r.URL.Query().Get("symbol"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin", "rank": 1, "is_active": 1},
			},
		})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "resolve", "--symbol", "btc", "-o", "json")
	require.NoError(t, err)

	var assets []api.ResolvedAsset
	require.NoError(t, json.Unmarshal([]byte(stdout), &assets))
	assert.Len(t, assets, 1)
	assert.Equal(t, int64(1), assets[0].ID)
	assert.Equal(t, "bitcoin", assets[0].Slug)
}

func TestResolve_AmbiguousSymbol(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin", "rank": 1, "is_active": 1},
				{"id": 2, "name": "Bitcoin Token", "symbol": "BTC", "slug": "bitcoin-token", "rank": 200, "is_active": 1},
			},
		})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, _, err := executeCommand(t, "resolve", "--symbol", "btc", "-o", "json")
	require.Error(t, err)
	assert.ErrorIs(t, err, api.ErrResolverAmbiguous)
	assert.Contains(t, err.Error(), "ambiguous")
}

func TestSearch_MissingArg(t *testing.T) {
	_, _, err := executeCommand(t, "resolve")
	require.Error(t, err)
}

func TestSearch_APIError(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = fmt.Fprint(w, `{"status":{"error_code":500,"error_message":"Server error"}}`)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, _, err := executeCommand(t, "resolve", "--symbol", "btc", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestResolve_AmbiguousSymbol_JSONMode_CandidatesInStderr(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin", "rank": 1, "is_active": 1},
				{"id": 2, "name": "Bitcoin Token", "symbol": "BTC", "slug": "bitcoin-token", "rank": 200, "is_active": 1},
			},
		})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, stderr, err := executeCommand(t, "resolve", "--symbol", "BTC", "-o", "json")
	require.Error(t, err)

	// JSON mode: stderr must contain structured error with candidates.
	var cliErr CLIError
	require.NoError(t, json.Unmarshal([]byte(stderr), &cliErr), "stderr should contain valid JSON: %s", stderr)
	assert.Equal(t, "resolver_ambiguous", cliErr.Error)
	require.NotNil(t, cliErr.Candidates, "candidates must be present in JSON stderr")
	assert.Len(t, cliErr.Candidates, 2)
	assert.Equal(t, int64(1), cliErr.Candidates[0].ID)
	assert.Equal(t, "bitcoin", cliErr.Candidates[0].Slug)
}

func TestResolve_AmbiguousSymbol_TableMode_CandidatesRendered(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin", "rank": 1, "is_active": 1},
				{"id": 2, "name": "Bitcoin Token", "symbol": "BTC", "slug": "bitcoin-token", "rank": 200, "is_active": 1},
			},
		})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, stderr, err := executeCommand(t, "resolve", "--symbol", "BTC")
	require.Error(t, err)

	// Table mode: stderr must render a readable table with candidates.
	assert.Contains(t, stderr, "Multiple assets match symbol", "stderr should explain ambiguity")
	assert.Contains(t, stderr, "bitcoin", "stderr should list first candidate slug")
	assert.Contains(t, stderr, "bitcoin-token", "stderr should list second candidate slug")
	assert.Contains(t, stderr, "ID", "stderr should render table headers")
}

func TestSearchCommand_TextRanking_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/cryptocurrency/map", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 2, "name": "Bitcoin Cash", "symbol": "BCH", "slug": "bitcoin-cash", "rank": 12, "is_active": 1},
				{"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin", "rank": 1, "is_active": 1},
				{"id": 3, "name": "Wrapped Bitcoin", "symbol": "WBTC", "slug": "wrapped-bitcoin", "rank": 15, "is_active": 1},
			},
		})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "search", "bitcoin", "-o", "json")
	require.NoError(t, err)

	var results []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &results))
	require.Len(t, results, 3)
	assert.Equal(t, "asset", results[0]["kind"])
	assert.Equal(t, "bitcoin", results[0]["slug"])
	assert.Equal(t, "bitcoin-cash", results[1]["slug"])
	assert.Equal(t, "wrapped-bitcoin", results[2]["slug"])
}

func TestSearchCommand_LimitApplied(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/cryptocurrency/map", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin", "rank": 1, "is_active": 1},
				{"id": 2, "name": "Bitcoin Cash", "symbol": "BCH", "slug": "bitcoin-cash", "rank": 12, "is_active": 1},
				{"id": 3, "name": "Wrapped Bitcoin", "symbol": "WBTC", "slug": "wrapped-bitcoin", "rank": 15, "is_active": 1},
			},
		})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "search", "bitcoin", "--limit", "2", "-o", "json")
	require.NoError(t, err)

	var results []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &results))
	require.Len(t, results, 2)
}

func TestSearchCommand_AddressLookup_PairBeforeToken_WithChainName(t *testing.T) {
	var seen []string
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.Path+"?"+r.URL.RawQuery)
		switch r.URL.Path {
		case "/v4/dex/pairs/quotes/latest":
			assert.Equal(t, "ethereum", r.URL.Query().Get("network_slug"))
			assert.Equal(t, "0xPair", r.URL.Query().Get("contract_address"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"contract_address": "0xPair",
						"network_id":       1,
						"network_name":     "Ethereum",
						"network_slug":     "ethereum",
						"dex_id":           101,
						"dex_name":         "Uniswap V3",
						"dex_slug":         "uniswap-v3",
						"base_asset": map[string]any{
							"id":               1,
							"name":             "Siren",
							"symbol":           "SIREN",
							"slug":             "siren",
							"contract_address": "0xToken",
						},
						"quote_asset": map[string]any{
							"id":               825,
							"name":             "Tether USDt",
							"symbol":           "USDT",
							"slug":             "tether",
							"contract_address": "0xQuote",
						},
						"quote": map[string]any{
							"2781": map[string]any{
								"price":                  1.23,
								"volume_24h":             2200,
								"liquidity":              54000,
								"no_of_transactions_24h": 88,
							},
						},
					},
				},
			})
		case "/v4/dex/spot-pairs/latest":
			assert.Equal(t, "ethereum", r.URL.Query().Get("network_slug"))
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":{"error_code":400,"error_message":"Please provide either a dex id or dex slug."}}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "search", "--chain", "ethereum", "--address", " 0xPair ", "-o", "json")
	require.NoError(t, err)

	var results []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &results))
	require.Len(t, results, 1)
	assert.Equal(t, "pair", results[0]["kind"])
	assert.Equal(t, "ethereum", results[0]["chain"])
	assert.Equal(t, "0xPair", results[0]["address"])
	assert.Equal(t, "uniswap-v3", results[0]["dex"])

	require.GreaterOrEqual(t, len(seen), 2)
	assert.True(t, strings.HasPrefix(seen[0], "/v4/dex/pairs/quotes/latest"), "pair lookup must run first")
	assert.True(t, strings.HasPrefix(seen[1], "/v4/dex/spot-pairs/latest"), "token fallback should run after pair lookup")
	assert.Contains(t, seen[1], "dex_slug=uniswap-v3")
}

func TestSearchCommand_AddressLookup_WithChainID(t *testing.T) {
	var seen []string
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.Path+"?"+r.URL.RawQuery)
		switch r.URL.Path {
		case "/v4/dex/pairs/quotes/latest":
			assert.Equal(t, "1", r.URL.Query().Get("network_id"))
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}})
		case "/v4/dex/spot-pairs/latest":
			assert.Equal(t, "1", r.URL.Query().Get("network_id"))
			if len(seen) == 2 {
				assert.Equal(t, "uniswap-v3", r.URL.Query().Get("dex_slug"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "search", "--chain", "1", "--address", "0xToken", "-o", "json")
	require.NoError(t, err)

	var results []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &results))
	assert.Len(t, results, 0)
	require.GreaterOrEqual(t, len(seen), 2)
	assert.Contains(t, seen[1], "dex_slug=uniswap-v3")
}

func TestSearchCommand_AddressLookup_LiveSchemaTokenFallback(t *testing.T) {
	var seen []string
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.Path+"?"+r.URL.RawQuery)
		switch r.URL.Path {
		case "/v4/dex/pairs/quotes/latest":
			assert.Equal(t, "ethereum", r.URL.Query().Get("network_slug"))
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}})
		case "/v4/dex/spot-pairs/latest":
			assert.Equal(t, "ethereum", r.URL.Query().Get("network_slug"))
			if len(seen) == 2 {
				assert.Equal(t, "uniswap-v3", r.URL.Query().Get("dex_slug"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"contract_address":             "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640",
						"network_id":                   "1",
						"network_name":                 "Ethereum",
						"network_slug":                 "Ethereum",
						"dex_id":                       "1348",
						"dex_name":                     "Uniswap V3",
						"dex_slug":                     "uniswap-v3",
						"base_asset_name":              "USD Coin",
						"base_asset_symbol":            "USDC",
						"base_asset_contract_address":  "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
						"quote_asset_name":             "Wrapped Ether",
						"quote_asset_symbol":           "WETH",
						"quote_asset_contract_address": "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
						"quote": []map[string]any{
							{
								"convert_id":             "2781",
								"price":                  3500,
								"volume_24h":             1000000,
								"liquidity":              5000000,
								"percent_change_24h":     1.2,
								"no_of_transactions_24h": 120,
							},
						},
						"scroll_id": "MTAw",
					},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "search", "--chain", "ethereum", "--address", "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", "-o", "json")
	require.NoError(t, err)

	var results []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &results))
	require.Len(t, results, 1)
	assert.Equal(t, "token_contract", results[0]["kind"])
	assert.Equal(t, "ethereum", results[0]["chain"])
	assert.Equal(t, "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", results[0]["address"])
	assert.Equal(t, "USDC", results[0]["symbol"])
	assert.Equal(t, "USDC/WETH", results[0]["pair"])
	assert.Equal(t, "uniswap-v3", results[0]["dex"])
	require.GreaterOrEqual(t, len(seen), 2)
	assert.Contains(t, seen[1], "dex_slug=uniswap-v3")
}

func TestSearchCommand_DryRun_AddressSearchShowsMultiRequest(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call in dry-run mode")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "search", "--chain", "ethereum", "--address", "0xPair", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out []dryRunOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	require.Len(t, out, 1+len(api.DEXSearchCandidateSlugs("ethereum")))
	assert.Equal(t, "ethereum", out[0].Params["network_slug"])
	assert.Contains(t, out[0].URL, "/v4/dex/pairs/quotes/latest")
	assert.NotContains(t, out[0].Params, "limit")
	assert.Equal(t, "ethereum", out[1].Params["network_slug"])
	assert.Equal(t, "uniswap-v3", out[1].Params["dex_slug"])
	assert.Contains(t, out[1].URL, "/v4/dex/spot-pairs/latest")
	assert.NotContains(t, out[1].Params, "contract_address")
	assert.Equal(t, "100", out[1].Params["limit"])
	assert.Equal(t, "liquidity", out[1].Params["sort"])
	assert.Equal(t, "desc", out[1].Params["sort_dir"])
	assert.Equal(t, "2781", out[1].Params["convert_id"])
	assert.Contains(t, out[1].Note, "candidate DEX slugs")
}

func TestSearchCommand_DryRun_QuerySearchShowsMapEndpoint(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call in dry-run mode")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "search", "btc", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.NotContains(t, out.Params, "query")
	assert.Contains(t, out.URL, "/v1/cryptocurrency/map")
}

func TestSearchCommand_AddressLookup_MissingChainFails(t *testing.T) {
	_, _, err := executeCommand(t, "search", "--address", "0xPair", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--chain is required with --address")
}

func TestSearchCommand_QueryAndAddressMutuallyExclusive(t *testing.T) {
	_, _, err := executeCommand(t, "search", "btc", "--chain", "ethereum", "--address", "0xPair", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provide either <query> or --address")
}

func TestSearchCommand_AddressTrimmedPreserved(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v4/dex/pairs/quotes/latest":
			assert.Equal(t, "0xAbC123", r.URL.Query().Get("contract_address"))
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}})
		case "/v4/dex/spot-pairs/latest":
			assert.Equal(t, "ethereum", r.URL.Query().Get("network_slug"))
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":{"error_code":400,"error_message":"Please provide either a dex id or dex slug."}}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, _, err := executeCommand(t, "search", "--chain", "ethereum", "--address", " 0xAbC123 ", "-o", "json")
	require.NoError(t, err)
}

func TestSearchCommand_TableOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/cryptocurrency/map", r.URL.Path)
		values, _ := url.ParseQuery(r.URL.RawQuery)
		assert.Empty(t, values)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin", "rank": 1, "is_active": 1},
			},
		})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "search", "btc", "-o", "table")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Kind")
	assert.Contains(t, stdout, "Bitcoin")
	assert.Contains(t, stdout, "asset")
}
