package tui

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/coinmarketcap/coinmarketcap-cli/internal/api"

	"github.com/stretchr/testify/require"
)

func TestMarketsFetchCoinsUsesCategory(t *testing.T) {
	callCount := 0
	client := newLandingTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch callCount {
		case 1:
			require.Equal(t, "/v1/cryptocurrency/categories", r.URL.Path)
			require.Equal(t, "1", r.URL.Query().Get("start"))
			require.Equal(t, "5000", r.URL.Query().Get("limit"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "604f2776ebccdd50cd175fdc", "name": "Layer 2", "title": "Layer 2"},
				},
			})
		case 2:
			require.Equal(t, "/v1/cryptocurrency/category", r.URL.Path)
			require.Equal(t, "604f2776ebccdd50cd175fdc", r.URL.Query().Get("id"))
			resp := api.CategoryResponse{Data: api.CategoryDetail{Coins: []api.ListingCoin{}}}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			t.Fatalf("unexpected request %d", callCount)
		}
	})

	m := NewMarketsModelWithCategory(client, "USD", 50, "layer-2")
	msg := m.fetchCoins()()
	loaded, ok := msg.(coinsLoadedMsg)
	require.True(t, ok)
	require.NoError(t, loaded.err)
	require.Equal(t, 2, callCount)
}

func TestMarketsViewShowsCategoryStatus(t *testing.T) {
	client := newLandingTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	})

	m := NewMarketsModelWithCategory(client, "USD", 50, "layer-2")
	m.state = marketsLoaded
	m.width = 120
	m.height = 40
	m.status = "Updated 1 markets"
	m.coins = []api.ListingCoin{
		{
			ID:      1,
			Name:    "Bitcoin",
			Symbol:  "BTC",
			CMCRank: 1,
			Quote: map[string]api.Quote2{
				"USD": {
					Price:            50000,
					MarketCap:        1e12,
					Volume24h:        4.5e10,
					PercentChange24h: 2.5,
				},
			},
		},
	}

	view := m.View()
	require.Contains(t, view, "layer-2")
	require.Contains(t, view, "Updated 1 markets")
}

func TestListVisibleRowsUsesAvailableHeight(t *testing.T) {
	require.Greater(t, listVisibleRows(40), 5)
	require.Equal(t, 5, listVisibleRows(10))
	require.Greater(t, listVisibleRows(0), 5)
}

func TestMarketsViewScrollsToCursor(t *testing.T) {
	client := newLandingTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	})

	coins := make([]api.ListingCoin, 50)
	for i := range coins {
		n := i + 1
		coins[i] = api.ListingCoin{
			ID:      int64(n),
			Name:    "Coin " + strconv.Itoa(n),
			Symbol:  "C" + strconv.Itoa(n),
			CMCRank: n,
			Quote: map[string]api.Quote2{
				"USD": {
					Price:            float64(n),
					MarketCap:        float64(n) * 10,
					Volume24h:        float64(n) * 5,
					PercentChange24h: float64(n),
				},
			},
		}
	}

	m := NewMarketsModel(client, "USD", 50)
	m.state = marketsLoaded
	m.width = 120
	m.height = 24
	m.coins = coins
	m.cursor = 49

	view := m.View()
	require.Contains(t, view, "Coin 50")
}

func TestListWindowAnchorsCursor(t *testing.T) {
	start, end := listWindow(50, 10, 49)
	require.Equal(t, 40, start)
	require.Equal(t, 50, end)
}
