package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics_DryRun(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call in dry-run mode")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommandCLI(t, "metrics", "--convert", "EUR", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Equal(t, "GET", out.Method)
	assert.Contains(t, out.URL, "/v1/global-metrics/quotes/latest")
	assert.Equal(t, "EUR", out.Params["convert"])
}

func TestMetrics_DefaultJSONOutput_CompactShape(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/global-metrics/quotes/latest", r.URL.Path)
		assert.Equal(t, "USD", r.URL.Query().Get("convert"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"active_cryptocurrencies": 13042,
				"active_exchanges":        812,
				"active_market_pairs":     99234,
				"btc_dominance":           54.21,
				"eth_dominance":           16.48,
				"last_updated":            "2026-03-24T08:00:00.000Z",
				"quote": map[string]any{
					"USD": map[string]any{
						"total_market_cap":    3123456789012.34,
						"total_volume_24h":    189234567890.12,
						"altcoin_volume_24h":  92345678901.23,
						"total_volume_24h_yesterday": 180000000000.0,
					},
				},
			},
		})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommandCLI(t, "metrics")
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Equal(t, "USD", out["convert"])
	assert.Equal(t, float64(13042), out["active_cryptocurrencies"])
	assert.Equal(t, float64(812), out["active_exchanges"])
	assert.Equal(t, float64(99234), out["active_market_pairs"])
	assert.Equal(t, 54.21, out["btc_dominance"])
	assert.Equal(t, 16.48, out["eth_dominance"])
	assert.Equal(t, 3123456789012.34, out["total_market_cap"])
	assert.Equal(t, 189234567890.12, out["total_volume_24h"])
	assert.Equal(t, 92345678901.23, out["altcoin_volume_24h"])
	assert.Equal(t, "2026-03-24T08:00:00.000Z", out["last_updated"])
}

func TestMetrics_TableOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"active_cryptocurrencies": 13042,
				"active_exchanges":        812,
				"active_market_pairs":     99234,
				"btc_dominance":           54.21,
				"eth_dominance":           16.48,
				"last_updated":            "2026-03-24T08:00:00.000Z",
				"quote": map[string]any{
					"USD": map[string]any{
						"total_market_cap":   3123456789012.34,
						"total_volume_24h":   189234567890.12,
						"altcoin_volume_24h": 92345678901.23,
					},
				},
			},
		})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "metrics")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Active Cryptocurrencies")
	assert.Contains(t, stdout, "Active Exchanges")
	assert.Contains(t, stdout, "Active Market Pairs")
	assert.Contains(t, stdout, "BTC Dominance")
	assert.Contains(t, stdout, "ETH Dominance")
	assert.Contains(t, stdout, "Total Market Cap")
	assert.Contains(t, stdout, "Total Volume 24h")
	assert.Contains(t, stdout, "Altcoin Volume 24h")
}

func TestMetrics_MissingQuoteForConvertFailsClosed(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "EUR", r.URL.Query().Get("convert"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"active_cryptocurrencies": 13042,
				"active_exchanges":        812,
				"active_market_pairs":     99234,
				"btc_dominance":           54.21,
				"eth_dominance":           16.48,
				"last_updated":            "2026-03-24T08:00:00.000Z",
				"quote": map[string]any{
					"USD": map[string]any{
						"total_market_cap":   3123456789012.34,
						"total_volume_24h":   189234567890.12,
						"altcoin_volume_24h": 92345678901.23,
					},
				},
			},
		})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, _, err := executeCommandCLI(t, "metrics", "--convert", "EUR", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quote data for convert EUR not found")
}

func TestMetrics_NoPositionalArgs(t *testing.T) {
	_, _, err := executeCommandCLI(t, "metrics", "extra", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}
