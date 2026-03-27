package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gainerCoinJSON = `{
	"id": "bitcoin",
	"symbol": "btc",
	"name": "Bitcoin",
	"image": "https://example.com/btc.png",
	"market_cap_rank": 1,
	"usd": 50000.0,
	"usd_24h_change": 2.5
}`

func TestGainerCoin_UnmarshalJSON(t *testing.T) {
	var coin GainerCoin
	err := json.Unmarshal([]byte(gainerCoinJSON), &coin)
	require.NoError(t, err)

	// Known fields are extracted into struct fields.
	assert.Equal(t, "bitcoin", coin.ID)
	assert.Equal(t, "btc", coin.Symbol)
	assert.Equal(t, "Bitcoin", coin.Name)
	assert.Equal(t, "https://example.com/btc.png", coin.Image)
	assert.Equal(t, 1, coin.MarketCapRank)

	// Dynamic keys are preserved in Extra.
	require.NotNil(t, coin.Extra)
	assert.Equal(t, 50000.0, coin.Extra["usd"])
	assert.Equal(t, 2.5, coin.Extra["usd_24h_change"])

	// Known keys also remain in Extra (it stores the full raw map).
	assert.Equal(t, "bitcoin", coin.Extra["id"])
}

func TestGainerCoin_Price(t *testing.T) {
	var coin GainerCoin
	err := json.Unmarshal([]byte(gainerCoinJSON), &coin)
	require.NoError(t, err)

	assert.Equal(t, 50000.0, coin.Price("usd"))

	// Non-existent currency returns zero value.
	assert.Equal(t, 0.0, coin.Price("eur"))
}

func TestGainerCoin_PriceChange(t *testing.T) {
	var coin GainerCoin
	err := json.Unmarshal([]byte(gainerCoinJSON), &coin)
	require.NoError(t, err)

	assert.Equal(t, 2.5, coin.PriceChange("usd"))

	// Non-existent currency returns zero value.
	assert.Equal(t, 0.0, coin.PriceChange("eur"))
}

func TestGainerCoin_MarshalJSON_WithExtra(t *testing.T) {
	var coin GainerCoin
	err := json.Unmarshal([]byte(gainerCoinJSON), &coin)
	require.NoError(t, err)

	data, err := json.Marshal(coin)
	require.NoError(t, err)

	// Re-parse the output and verify dynamic keys survived the round-trip.
	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Equal(t, "bitcoin", raw["id"])
	assert.Equal(t, "btc", raw["symbol"])
	assert.Equal(t, 50000.0, raw["usd"])
	assert.Equal(t, 2.5, raw["usd_24h_change"])
}

func TestGainerCoin_MarshalJSON_NilExtra(t *testing.T) {
	coin := GainerCoin{
		ID:            "ethereum",
		Symbol:        "eth",
		Name:          "Ethereum",
		Image:         "https://example.com/eth.png",
		MarketCapRank: 2,
	}

	data, err := json.Marshal(coin)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Equal(t, "ethereum", raw["id"])
	assert.Equal(t, "eth", raw["symbol"])
	assert.Equal(t, "Ethereum", raw["name"])
	assert.Equal(t, "https://example.com/eth.png", raw["image"])
	assert.Equal(t, float64(2), raw["market_cap_rank"])
}

func TestGainerCoin_RoundTrip(t *testing.T) {
	// Unmarshal from JSON, then marshal back, and verify the two JSON
	// representations are semantically equivalent.
	var coin GainerCoin
	err := json.Unmarshal([]byte(gainerCoinJSON), &coin)
	require.NoError(t, err)

	data, err := json.Marshal(coin)
	require.NoError(t, err)

	// Parse both original and round-tripped JSON into generic maps.
	var original, roundTripped map[string]interface{}
	err = json.Unmarshal([]byte(gainerCoinJSON), &original)
	require.NoError(t, err)
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	assert.Equal(t, original, roundTripped)
}
