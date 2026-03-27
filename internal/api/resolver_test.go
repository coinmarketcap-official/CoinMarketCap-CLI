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

func newResolverClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(handler)
	cfg := &config.Config{APIKey: "test-key", Tier: config.TierBasic}
	client := NewClient(cfg)
	client.SetBaseURL(srv.URL)
	return client, srv
}

func TestResolveByID(t *testing.T) {
	client, srv := newResolverClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/cryptocurrency/info", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("id"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"1": map[string]any{
					"id":          1,
					"name":        "Bitcoin",
					"symbol":      "BTC",
					"slug":        "bitcoin",
					"date_added":  "2013-04-28T00:00:00.000Z",
					"description": "",
					"urls": map[string]any{
						"website": []string{},
					},
				},
			},
		})
	})
	defer srv.Close()

	asset, err := client.ResolveByID(context.Background(), "1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), asset.ID)
	assert.Equal(t, "bitcoin", asset.Slug)
	assert.Equal(t, "BTC", asset.Symbol)
	assert.Equal(t, "Bitcoin", asset.Name)

	raw, err := json.Marshal(asset)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"rank":null`)
	assert.Contains(t, string(raw), `"is_active":null`)
}

func TestResolveBySlug(t *testing.T) {
	client, srv := newResolverClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/cryptocurrency/info", r.URL.Path)
		assert.Equal(t, "ethereum", r.URL.Query().Get("slug"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"1027": map[string]any{
					"id": 1027, "name": "Ethereum", "symbol": "ETH", "slug": "ethereum", "category": "coin",
				},
			},
		})
	})
	defer srv.Close()

	asset, err := client.ResolveBySlug(context.Background(), "ethereum")
	require.NoError(t, err)
	assert.Equal(t, int64(1027), asset.ID)
	assert.Equal(t, "ethereum", asset.Slug)
	assert.Equal(t, "ETH", asset.Symbol)
	assert.Equal(t, "Ethereum", asset.Name)

	raw, err := json.Marshal(asset)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"rank":null`)
	assert.Contains(t, string(raw), `"is_active":null`)
}

func TestResolveBySlugEmpty(t *testing.T) {
	client, srv := newResolverClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request to %s?%s", r.URL.Path, r.URL.RawQuery)
	})
	defer srv.Close()

	_, err := client.ResolveBySlug(context.Background(), "   ")
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestResolveBySlugNotFound(t *testing.T) {
	client, srv := newResolverClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/cryptocurrency/info", r.URL.Path)
		assert.Equal(t, "missing-slug", r.URL.Query().Get("slug"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{},
		})
	})
	defer srv.Close()

	_, err := client.ResolveBySlug(context.Background(), "missing-slug")
	require.ErrorIs(t, err, ErrAssetNotFound)
}

func TestResolveBySymbolUnique(t *testing.T) {
	client, srv := newResolverClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "BTC", r.URL.Query().Get("symbol"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin", "rank": 1, "is_active": 1},
			},
		})
	})
	defer srv.Close()

	asset, err := client.ResolveBySymbol(context.Background(), "BTC")
	require.NoError(t, err)
	assert.Equal(t, int64(1), asset.ID)
}

func TestResolveBySymbolAmbiguous(t *testing.T) {
	client, srv := newResolverClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin", "rank": 1, "is_active": 1},
				{"id": 2, "name": "Bitcoin Token", "symbol": "BTC", "slug": "bitcoin-token", "rank": 200, "is_active": 1},
			},
		})
	})
	defer srv.Close()

	_, err := client.ResolveBySymbol(context.Background(), "BTC")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrResolverAmbiguous)

	var amb *ResolverAmbiguityError
	require.ErrorAs(t, err, &amb)
	assert.Len(t, amb.Candidates, 2)
}

func TestResolveBySymbolNotFound(t *testing.T) {
	client, srv := newResolverClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}})
	})
	defer srv.Close()

	_, err := client.ResolveBySymbol(context.Background(), "MISSING")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAssetNotFound)
}

func TestResolveBySymbolOrSlugSymbolFirst(t *testing.T) {
	client, srv := newResolverClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/cryptocurrency/map", r.URL.Path)
		assert.Equal(t, "BTC", r.URL.Query().Get("symbol"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin", "rank": 1, "is_active": 1},
			},
		})
	})
	defer srv.Close()

	asset, err := client.ResolveBySymbolOrSlug(context.Background(), "btc")
	require.NoError(t, err)
	assert.Equal(t, int64(1), asset.ID)
}

func TestResolveBySymbolOrSlugFallsBackToSlug(t *testing.T) {
	client, srv := newResolverClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/cryptocurrency/map":
			assert.Equal(t, "BITCOIN", r.URL.Query().Get("symbol"))
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}})
		case "/v2/cryptocurrency/info":
			assert.Equal(t, "bitcoin", r.URL.Query().Get("slug"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"1": map[string]any{"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin"},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()

	asset, err := client.ResolveBySymbolOrSlug(context.Background(), "BITCOIN")
	require.NoError(t, err)
	assert.Equal(t, int64(1), asset.ID)
	assert.Equal(t, "bitcoin", asset.Slug)
}

func TestResolveBySymbolOrSlugAmbiguityNoSlugFallback(t *testing.T) {
	client, srv := newResolverClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/cryptocurrency/map", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Bitcoin", "symbol": "BTC", "slug": "bitcoin", "rank": 1, "is_active": 1},
				{"id": 2, "name": "Bitcoin Token", "symbol": "BTC", "slug": "bitcoin-token", "rank": 200, "is_active": 1},
			},
		})
	})
	defer srv.Close()

	_, err := client.ResolveBySymbolOrSlug(context.Background(), "BTC")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrResolverAmbiguous)
}
