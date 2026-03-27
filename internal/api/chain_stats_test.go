package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinmarketcap/coinmarketcap-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newChainStatsClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(handler)
	cfg := &config.Config{APIKey: "test-key", Tier: config.TierBasic}
	client := NewClient(cfg)
	client.SetBaseURL(srv.URL)
	return client, srv
}

func TestBlockchainStatisticsLatestByIDs(t *testing.T) {
	client, srv := newChainStatsClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/blockchain/statistics/latest", r.URL.Path)
		assert.Equal(t, "1,1027", r.URL.Query().Get("id"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"1": map[string]any{
					"id":                    1,
					"slug":                  "bitcoin",
					"symbol":                "BTC",
					"block_reward_static":   "3.125",
					"consensus_mechanism":   "proof-of-work",
					"difficulty":            "11890594958796",
					"hashrate_24h":          "85116194130018810000",
					"pending_transactions":  "1177",
					"reduction_rate":        "50%",
					"total_blocks":          "595165",
					"total_transactions":    "455738994",
					"tps_24h":               "3.808090277777778",
					"first_block_timestamp": "2009-01-09T02:54:25.000Z",
				},
				"1027": map[string]any{
					"id":                    1027,
					"slug":                  "ethereum",
					"symbol":                "ETH",
					"block_reward_static":   "2.0",
					"consensus_mechanism":   "proof-of-stake",
					"difficulty":            "0",
					"hashrate_24h":          "0",
					"pending_transactions":  "42",
					"reduction_rate":        "0%",
					"total_blocks":          "22000000",
					"total_transactions":    "2000000000",
					"tps_24h":               "12.5",
					"first_block_timestamp": "2015-07-30T15:26:13.000Z",
				},
			},
		})
	})
	defer srv.Close()

	stats, err := client.BlockchainStatisticsLatestByIDs(context.Background(), []string{"1", "1027"})
	require.NoError(t, err)
	require.Len(t, stats, 2)
	assert.Equal(t, "proof-of-work", stats["1"].ConsensusMechanism)
	assert.Equal(t, "proof-of-stake", stats["1027"].ConsensusMechanism)
}
