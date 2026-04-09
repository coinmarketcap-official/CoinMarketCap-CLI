package cmd

import (
	"encoding/json"
	"net/http"
	"regexp"
	"testing"

	"github.com/openCMC/CoinMarketCap-CLI/internal/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrending_DryRun(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call in dry-run mode")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "trending", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Contains(t, out.URL, "/v1/cryptocurrency/trending/latest")
	assert.Equal(t, "50", out.Params["limit"])
	assert.Equal(t, "1", out.Params["start"])
	assert.Equal(t, "USD", out.Params["convert"])
}

func TestTrending_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/cryptocurrency/trending/latest", r.URL.Path)
		assert.Equal(t, "50", r.URL.Query().Get("limit"))
		assert.Equal(t, "1", r.URL.Query().Get("start"))
		assert.Equal(t, "USD", r.URL.Query().Get("convert"))
		resp := api.ListingsLatestResponse{
			Data: []api.ListingCoin{
				{
					ID:      1,
					Name:    "Bitcoin",
					Symbol:  "BTC",
					Slug:    "bitcoin",
					CMCRank: 1,
					Quote: map[string]api.Quote2{
						"USD": {
							Price:            65000,
							Volume24h:        45000000000,
							MarketCap:        1300000000000,
							PercentChange24h: 2.5,
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "trending", "-o", "json")
	require.NoError(t, err)

	var result []api.ListingCoin
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result, 1)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, "Bitcoin", result[0].Name)
}

func TestTrending_LimitValidation(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call when validation fails")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, _, err := executeCommand(t, "trending", "--limit", "0", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be between 1 and 50")

	_, _, err = executeCommand(t, "trending", "--limit", "51", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be between 1 and 50")
}

func TestTrending_CustomLimit(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/cryptocurrency/trending/latest", r.URL.Path)
		assert.Equal(t, "3", r.URL.Query().Get("limit"))
		assert.Equal(t, "1", r.URL.Query().Get("start"))
		resp := api.ListingsLatestResponse{Data: []api.ListingCoin{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, _, err := executeCommand(t, "trending", "--limit", "3", "-o", "json")
	require.NoError(t, err)
}

func TestTrending_TableOutputUsesTrendingPosition(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/cryptocurrency/trending/latest", r.URL.Path)
		resp := api.ListingsLatestResponse{
			Data: []api.ListingCoin{
				{
					ID:      1,
					Name:    "Bitcoin",
					Symbol:  "BTC",
					Slug:    "bitcoin",
					CMCRank: 99,
					Quote: map[string]api.Quote2{
						"USD": {
							Price:            65000,
							Volume24h:        45000000000,
							MarketCap:        1300000000000,
							PercentChange24h: 2.5,
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "trending")
	require.NoError(t, err)
	require.Regexp(t, regexp.MustCompile(`(?m)^\s+1\s+Bitcoin`), stdout)
}

func TestTrending_CommandsCatalogMetadata(t *testing.T) {
	stdout, _, err := executeCommand(t, "commands", "-o", "json")
	require.NoError(t, err)

	var catalog commandCatalog
	require.NoError(t, json.Unmarshal([]byte(stdout), &catalog))

	var trending commandInfo
	found := false
	for _, info := range catalog.Commands {
		if info.Name == "trending" {
			trending = info
			found = true
			break
		}
	}
	require.True(t, found, "trending command should be present in catalog")
	require.Equal(t, "/v1/cryptocurrency/trending/latest", trending.APIEndpoint)
	require.Equal(t, "getV1CryptocurrencyTrendingLatest", trending.OASOperationID)
	assert.Contains(t, trending.Examples, "cmc trending --limit 3")
}
