package tui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openCMC/CoinMarketCap-CLI/internal/api"
	"github.com/openCMC/CoinMarketCap-CLI/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

// makeTrendingInDetail returns a TrendingModel already in trendingDetail state.
func makeTrendingInDetail() TrendingModel {
	m := TrendingModel{
		state:  trendingDetail,
		detail: DetailModel{loading: 3, vs: "USD"},
		width:  120,
		height: 40,
	}
	return m
}

// makeMarketsInDetail returns a MarketsModel already in marketsDetail state.
func makeMarketsInDetail() MarketsModel {
	m := MarketsModel{
		state:  marketsDetail,
		detail: DetailModel{loading: 3, vs: "USD"},
		width:  120,
		height: 40,
	}
	return m
}

// --- Trending detail message forwarding ---

func TestTrendingDetail_ForwardsInfoMsg(t *testing.T) {
	m := makeTrendingInDetail()
	info := &api.CoinInfo{Name: "Bitcoin", Symbol: "BTC", Slug: "bitcoin"}
	updated, _ := m.Update(coinInfoMsg{info: info})
	tm := updated.(TrendingModel)
	if tm.detail.info == nil {
		t.Fatal("coinInfoMsg was not forwarded to detail model")
	}
	if tm.detail.info.Name != "Bitcoin" {
		t.Fatalf("expected info.Name=Bitcoin, got %s", tm.detail.info.Name)
	}
	if tm.detail.loading != 2 {
		t.Fatalf("expected loading=2 after info msg, got %d", tm.detail.loading)
	}
}

func TestTrendingDetail_ForwardsQuoteMsg(t *testing.T) {
	m := makeTrendingInDetail()
	quote := &api.QuoteCoin{Name: "Bitcoin", Quote: map[string]api.Quote{"USD": {Price: 70000}}}
	updated, _ := m.Update(quoteDetailMsg{quote: quote})
	tm := updated.(TrendingModel)
	if tm.detail.quote == nil {
		t.Fatal("quoteDetailMsg was not forwarded to detail model")
	}
	if tm.detail.loading != 2 {
		t.Fatalf("expected loading=2 after quote msg, got %d", tm.detail.loading)
	}
}

func TestTrendingDetail_ForwardsOHLCMsg(t *testing.T) {
	m := makeTrendingInDetail()
	data := api.OHLCData{{1, 2, 3, 4, 5}}
	updated, _ := m.Update(ohlcMsg{data: data})
	tm := updated.(TrendingModel)
	if len(tm.detail.ohlc) != 1 {
		t.Fatalf("ohlcMsg was not forwarded to detail model, got %d rows", len(tm.detail.ohlc))
	}
	if tm.detail.loading != 2 {
		t.Fatalf("expected loading=2 after ohlc msg, got %d", tm.detail.loading)
	}
}

// --- Markets detail message forwarding ---

func TestMarketsDetail_ForwardsInfoMsg(t *testing.T) {
	m := makeMarketsInDetail()
	info := &api.CoinInfo{Name: "Ethereum", Symbol: "ETH", Slug: "ethereum"}
	updated, _ := m.Update(coinInfoMsg{info: info})
	mm := updated.(MarketsModel)
	if mm.detail.info == nil {
		t.Fatal("coinInfoMsg was not forwarded to detail model")
	}
	if mm.detail.info.Name != "Ethereum" {
		t.Fatalf("expected info.Name=Ethereum, got %s", mm.detail.info.Name)
	}
}

func TestMarketsDetail_ForwardsQuoteMsg(t *testing.T) {
	m := makeMarketsInDetail()
	quote := &api.QuoteCoin{Name: "Ethereum", Quote: map[string]api.Quote{"USD": {Price: 2100}}}
	updated, _ := m.Update(quoteDetailMsg{quote: quote})
	mm := updated.(MarketsModel)
	if mm.detail.quote == nil {
		t.Fatal("quoteDetailMsg was not forwarded to detail model")
	}
}

func TestMarketsDetail_ForwardsOHLCMsg(t *testing.T) {
	m := makeMarketsInDetail()
	data := api.OHLCData{{1, 2, 3, 4, 5}, {6, 7, 8, 9, 10}}
	updated, _ := m.Update(ohlcMsg{data: data})
	mm := updated.(MarketsModel)
	if len(mm.detail.ohlc) != 2 {
		t.Fatalf("ohlcMsg was not forwarded to detail model, got %d rows", len(mm.detail.ohlc))
	}
}

// --- Detail refresh state mutation ---

func TestDetailRefresh_MutatesLoadingState(t *testing.T) {
	m := DetailModel{loading: 0, vs: "USD"}
	updated, cmd := m.Update(refreshTickMsg{})
	dm := updated.(DetailModel)
	if dm.loading != 3 {
		t.Fatalf("expected loading=3 after refreshTickMsg, got %d", dm.loading)
	}
	if dm.status != "Refreshing asset detail…" {
		t.Fatalf("expected refresh status message, got %q", dm.status)
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd from refreshTickMsg")
	}
}

func TestDetailRefresh_ManualR_MutatesLoadingState(t *testing.T) {
	m := DetailModel{loading: 0, vs: "USD"}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	dm := updated.(DetailModel)
	if dm.loading != 3 {
		t.Fatalf("expected loading=3 after pressing r, got %d", dm.loading)
	}
	if dm.status != "Refreshing asset detail…" {
		t.Fatalf("expected refresh status message, got %q", dm.status)
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd from r key")
	}
}

func TestDetailWindowCyclesWithP(t *testing.T) {
	m := DetailModel{window: chartWindow1H, vs: "USD"}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	dm := updated.(DetailModel)
	if dm.window != chartWindow24H {
		t.Fatalf("expected 1H -> 24H, got %s", dm.window.String())
	}
	if dm.status != "Loading 24H Price…" {
		t.Fatalf("expected loading status for 24H, got %q", dm.status)
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd from p key")
	}

	updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	dm = updated.(DetailModel)
	if dm.window != chartWindow30D {
		t.Fatalf("expected 24H -> 30D, got %s", dm.window.String())
	}

	updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	dm = updated.(DetailModel)
	if dm.window != chartWindow1H {
		t.Fatalf("expected 30D -> 1H, got %s", dm.window.String())
	}
}

func TestDetailView_RendersWindowTitle(t *testing.T) {
	base := DetailModel{
		info: &api.CoinInfo{Name: "Bitcoin", Symbol: "BTC", Slug: "bitcoin"},
		quote: &api.QuoteCoin{
			Name:   "Bitcoin",
			Symbol: "BTC",
			Quote: map[string]api.Quote{
				"USD": {Price: 70000, MarketCap: 1.2e12, Volume24h: 3e10, PercentChange24h: 2.4},
			},
		},
		ohlc:   api.OHLCData{{1, 1, 1, 1, 1}, {2, 2, 2, 2, 2}},
		width:  120,
		height: 40,
		vs:     "USD",
	}

	for _, tc := range []struct {
		name   string
		window chartWindow
		expect string
	}{
		{name: "1H", window: chartWindow1H, expect: "1H Price"},
		{name: "24H", window: chartWindow24H, expect: "24H Price"},
		{name: "30D", window: chartWindow30D, expect: "30D Price"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := base
			m.window = tc.window
			view := m.View()
			if !strings.Contains(view, tc.expect) {
				t.Fatalf("expected view to contain %q, got:\n%s", tc.expect, view)
			}
		})
	}
}

// --- OHLCV fallback to quotes/historical ---

func TestFetchOHLC_FallbackToHourlyQuotesHistorical(t *testing.T) {
	var requests []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		interval := r.URL.Query().Get("interval")
		if strings.Contains(r.URL.Path, "/v2/cryptocurrency/ohlcv/historical") {
			requests = append(requests, "ohlcv:"+interval)
			// Simulate 403 plan restricted for 5m and hourly OHLCV.
			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(map[string]any{
				"status": map[string]any{
					"error_code":    1003,
					"error_message": "plan restricted",
				},
			}); err != nil {
				t.Errorf("encode ohlcv restricted response: %v", err)
			}
			return
		}
		if strings.Contains(r.URL.Path, "/v1/cryptocurrency/quotes/historical") {
			requests = append(requests, "quotes:"+interval)
			if interval == "5m" {
				w.WriteHeader(http.StatusForbidden)
				if err := json.NewEncoder(w).Encode(map[string]any{
					"status": map[string]any{
						"error_code":    1003,
						"error_message": "plan restricted",
					},
				}); err != nil {
					t.Errorf("encode quotes restricted response: %v", err)
				}
				return
			}
			// Return valid quote historical data for the hourly fallback.
			if err := json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"1": map[string]any{
						"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin",
						"quotes": []map[string]any{
							{
								"timestamp": "2026-03-15T00:00:00.000Z",
								"quote": map[string]any{
									"USD": map[string]any{"price": 69000.0, "market_cap": 1.3e12, "volume_24h": 3e10},
								},
							},
							{
								"timestamp": "2026-03-16T00:00:00.000Z",
								"quote": map[string]any{
									"USD": map[string]any{"price": 70000.0, "market_cap": 1.35e12, "volume_24h": 3.1e10},
								},
							},
						},
					},
				},
			}); err != nil {
				t.Errorf("encode quotes historical response: %v", err)
			}
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	cfg := &config.Config{APIKey: "test-key", Tier: config.TierBasic}
	client := api.NewClient(cfg)
	client.SetBaseURL(srv.URL)

	m := DetailModel{client: client, coinID: "1", vs: "USD", window: chartWindow1H}
	cmd := m.fetchOHLC()
	msg := cmd()

	ohlc, ok := msg.(ohlcMsg)
	if !ok {
		t.Fatalf("expected ohlcMsg, got %T", msg)
	}
	if ohlc.err != nil {
		t.Fatalf("expected no error with fallback, got %v", ohlc.err)
	}
	if len(ohlc.data) != 2 {
		t.Fatalf("expected 2 data points from fallback, got %d", len(ohlc.data))
	}
	if !strings.Contains(ohlc.status, "hourly fallback") {
		t.Fatalf("expected hourly fallback status, got %q", ohlc.status)
	}
	if len(requests) < 4 {
		t.Fatalf("expected 4 requests (ohlcv 5m, quotes 5m, ohlcv hourly, quotes hourly), got %v", requests)
	}
	if requests[0] != "ohlcv:5m" || requests[1] != "quotes:5m" || requests[2] != "ohlcv:hourly" || requests[3] != "quotes:hourly" {
		t.Fatalf("unexpected request order: %v", requests)
	}
	// Verify synthesized OHLC: open=high=low=close=price
	row := ohlc.data[0]
	if len(row) < 5 {
		t.Fatalf("expected 5 values per row, got %d", len(row))
	}
	if row[1] != 69000.0 || row[2] != 69000.0 || row[3] != 69000.0 || row[4] != 69000.0 {
		t.Fatalf("expected all OHLC=69000 from price fallback, got O=%.0f H=%.0f L=%.0f C=%.0f", row[1], row[2], row[3], row[4])
	}
}
