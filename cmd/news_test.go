package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/openCMC/CoinMarketCap-CLI/internal/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNews_DryRun(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call in dry-run mode")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "news", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Contains(t, out.URL, "/v1/content/latest")
	assert.Equal(t, "1", out.Params["start"])
	assert.Equal(t, "20", out.Params["limit"])
	assert.Equal(t, "en", out.Params["language"])
	assert.Equal(t, "news", out.Params["news_type"])
}

func TestNews_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/content/latest", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("start"))
		assert.Equal(t, "20", r.URL.Query().Get("limit"))
		assert.Equal(t, "en", r.URL.Query().Get("language"))
		assert.Equal(t, "news", r.URL.Query().Get("news_type"))
		resp := api.NewsLatestResponse{
			Data: []api.NewsArticle{
				{
					Title:      "Bitcoin hits new milestone",
					Subtitle:   "Market commentary",
					SourceName: "CoinMarketCap",
					SourceURL:  "https://coinmarketcap.com",
					ReleasedAt: "2026-03-24T01:02:03Z",
					CreatedAt:  "2026-03-24T00:59:59Z",
					Assets:     []api.NewsAssetRef{{Symbol: "BTC"}, {Symbol: "ETH"}},
					Type:       "article",
					NewsType:   "news",
					Cover:      "https://example.com/cover.jpg",
					Language:   "en",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "news", "-o", "json")
	require.NoError(t, err)

	var items []newsItem
	require.NoError(t, json.Unmarshal([]byte(stdout), &items))
	require.Len(t, items, 1)
	assert.Equal(t, "Bitcoin hits new milestone", items[0].Title)
	assert.Equal(t, []string{"BTC", "ETH"}, items[0].Assets)
	assert.Equal(t, "CoinMarketCap", items[0].SourceName)
}

func TestNews_LimitValidation(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call when validation fails")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, _, err := executeCommand(t, "news", "--limit", "0", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be between 1 and 100")

	_, _, err = executeCommand(t, "news", "--limit", "101", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be between 1 and 100")
}

func TestNews_StartValidation(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call when validation fails")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, _, err := executeCommand(t, "news", "--start", "0", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--start must be at least 1")
}

func TestNews_TableOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := api.NewsLatestResponse{
			Data: []api.NewsArticle{
				{
					Title:      "Bitcoin hits new milestone",
					SourceName: "CoinMarketCap",
					ReleasedAt: "2026-03-24T01:02:03Z",
					Assets:     []api.NewsAssetRef{{Symbol: "BTC"}, {Symbol: "ETH"}},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "news")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Title")
	assert.Contains(t, stdout, "Source")
	assert.Contains(t, stdout, "Released At")
	assert.Contains(t, stdout, "BTC, ETH")
}
