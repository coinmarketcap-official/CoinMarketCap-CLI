package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coinmarketcap/coinmarketcap-cli/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	cfg := &config.Config{APIKey: "test-key", Tier: config.TierBasic}
	c := NewClient(cfg)
	c.SetBaseURL(srv.URL)
	return c, srv
}

func TestUserAgentHeader(t *testing.T) {
	var gotUA string
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{}"))
	})
	defer srv.Close()

	c.UserAgent = "coinmarketcap-cli/v1.2.3"
	var result map[string]any
	_ = c.get(context.Background(), "/test", &result)
	assert.Equal(t, "coinmarketcap-cli/v1.2.3", gotUA)
}

func TestAuthHeadersSent(t *testing.T) {
	var gotHeader string
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-CMC_PRO_API_KEY")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{}"))
	})
	defer srv.Close()

	var result map[string]any
	_ = c.get(context.Background(), "/test", &result)
	assert.Equal(t, "test-key", gotHeader)
}

func TestProAuthHeaders(t *testing.T) {
	var gotHeader string
	cfg := &config.Config{APIKey: "hobbyist-key", Tier: config.TierHobbyist}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-CMC_PRO_API_KEY")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{}"))
	}))
	defer srv.Close()

	c := NewClient(cfg)
	c.SetBaseURL(srv.URL)
	var result map[string]any
	_ = c.get(context.Background(), "/test", &result)
	assert.Equal(t, "hobbyist-key", gotHeader)
}

func TestError401InvalidKey(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"status":{"error_code":1001,"error_message":"This API Key is invalid"}}`))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAPIKey)
	assert.Contains(t, err.Error(), "API Key is invalid")
}

func TestError401EmptyBody(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAPIKey)
}

func TestError401APIKeyMissing(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"status":{"error_code":1002,"error_message":"API key missing"}}`))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAPIKey)
}

func TestError400InvalidInput(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"status":{"error_code":400,"error_message":"Invalid symbol parameter"}}`))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidInput)
	assert.Contains(t, err.Error(), "Invalid symbol")
}

func TestError404AssetNotFound(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"status":{"error_code":404,"error_message":"Cryptocurrency not found"}}`))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAssetNotFound)
	assert.Contains(t, err.Error(), "not found")
}

func TestError404GenericNotAssetNotFound(t *testing.T) {
	// 404 responses that don't indicate "asset not found" should NOT return ErrAssetNotFound.
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"status":{"error_code":404,"error_message":"Endpoint does not exist"}}`))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrAssetNotFound, "generic 404 should not be classified as asset_not_found")
	assert.Contains(t, err.Error(), "Endpoint does not exist")
}

func TestError403(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"status":{"error_code":1003,"error_message":"Your API plan doesn't include access to this endpoint"}}`))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPlanRestricted)
	assert.Contains(t, err.Error(), "API plan")
}

func TestError403EmptyBody(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPlanRestricted)
}

func TestError429(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(429)
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "retry after 30 seconds")
}

func TestError429NoRetryAfter(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRateLimited)
}

func TestError429WithRateLimitReset(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-ratelimit-reset", "2026-03-09 03:28:00 +0000")
		w.WriteHeader(429)
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRateLimited)

	var rle *RateLimitError
	require.ErrorAs(t, err, &rle)
	assert.Equal(t, 0, rle.RetryAfter) // no Retry-After header
	assert.False(t, rle.ResetAt.IsZero())
	assert.Equal(t, 2026, rle.ResetAt.Year())
	assert.Equal(t, time.Month(3), rle.ResetAt.Month())
	assert.Equal(t, 9, rle.ResetAt.Day())
	assert.Equal(t, 3, rle.ResetAt.Hour())
	assert.Equal(t, 28, rle.ResetAt.Minute())
	assert.Contains(t, err.Error(), "resets at 03:28:00 UTC")
}

func TestError429RetryAfterTakesPrecedence(t *testing.T) {
	// When both headers are present, RetryAfter should be set (used first by retry logic)
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "15")
		w.Header().Set("x-ratelimit-reset", "2026-03-09 03:28:00 +0000")
		w.WriteHeader(429)
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)

	var rle *RateLimitError
	require.ErrorAs(t, err, &rle)
	assert.Equal(t, 15, rle.RetryAfter)
	assert.False(t, rle.ResetAt.IsZero())
	assert.Contains(t, err.Error(), "retry after 15 seconds") // RetryAfter wins in message
}

func TestErrorUnknownStatusWithBody(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"status":{"error_code":500,"error_message":"Internal server error"}}`))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error 500")
	assert.Contains(t, err.Error(), "Internal server error")
}

func TestErrorUnknownStatusRawBody(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		_, _ = w.Write([]byte("Bad Gateway"))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error 502")
	assert.Contains(t, err.Error(), "Bad Gateway")
}

func TestSuccessResponseDecodes(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"bitcoin":{"usd":50000}}`))
	})
	defer srv.Close()

	var result PriceResponse
	err := c.get(context.Background(), "/test", &result)
	require.NoError(t, err)
	assert.Equal(t, float64(50000), result["bitcoin"]["usd"])
}

func TestRetryAfterInvalidFallback(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "invalid-value")
		w.WriteHeader(429)
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRateLimited)
}

func TestRequirePaid(t *testing.T) {
	cfg := &config.Config{Tier: config.TierBasic}
	c := NewClient(cfg)
	assert.ErrorIs(t, c.requirePaid(), ErrPlanRestricted)

	cfg2 := &config.Config{Tier: config.TierHobbyist}
	c2 := NewClient(cfg2)
	assert.NoError(t, c2.requirePaid())
}
