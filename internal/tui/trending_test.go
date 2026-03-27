package tui

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrendingViewUsesVisibleWindow(t *testing.T) {
	client := newLandingTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	})

	coins := make([]trendingCoin, 50)
	for i := range coins {
		n := i + 1
		coins[i] = trendingCoin{
			ID:               int64(n),
			Slug:             "coin-" + strconv.Itoa(n),
			Symbol:           "C" + strconv.Itoa(n),
			Name:             "Coin " + strconv.Itoa(n),
			Price:            float64(n),
			PercentChange24h: float64(n),
		}
	}

	m := NewTrendingModel(client, "USD")
	m.state = trendingLoaded
	m.width = 120
	m.height = 24
	m.coins = coins
	m.cursor = 49

	view := m.View()
	require.Contains(t, view, "Coin 50")
	require.Contains(t, view, "#50")
}

func TestTrendingVisibleRowsUsesAvailableHeight(t *testing.T) {
	require.Greater(t, listVisibleRows(40), 5)
	require.Equal(t, 5, listVisibleRows(10))
}

func TestTrendingWindowAnchorsCursor(t *testing.T) {
	start, end := listWindow(50, 10, 49)
	require.Equal(t, 40, start)
	require.Equal(t, 50, end)
}

func TestTrendingInitialLoadErrorDoesNotStayLoading(t *testing.T) {
	client := newLandingTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	})

	m := NewTrendingModel(client, "USD")
	updated, _ := m.Update(trendingLoadedMsg{err: fmt.Errorf("boom")})
	tm := updated.(TrendingModel)

	require.Equal(t, trendingLoaded, tm.state)
	require.Error(t, tm.err)
	view := tm.View()
	require.NotContains(t, view, "Fetching trending assets…")
	require.Contains(t, view, "Error: boom")
}
