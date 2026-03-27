package api

// GlobalMetricsResponse is the raw response from /v1/global-metrics/quotes/latest.
type GlobalMetricsResponse struct {
	Data GlobalMetricsData `json:"data"`
}

// GlobalMetricsData holds the main data payload from the global metrics endpoint.
type GlobalMetricsData struct {
	ActiveCryptocurrencies int                          `json:"active_cryptocurrencies"`
	ActiveExchanges        int                          `json:"active_exchanges"`
	ActiveMarketPairs      int                          `json:"active_market_pairs"`
	TotalCryptocurrencies  int                          `json:"total_cryptocurrencies"`
	BTCDominance           float64                      `json:"btc_dominance"`
	ETHDominance           float64                      `json:"eth_dominance"`
	DeFiVolume24h          float64                      `json:"defi_volume_24h"`
	DeFiVolume24hReported  float64                      `json:"defi_volume_24h_reported"`
	DeFiMarketCap          float64                      `json:"defi_market_cap"`
	DeFi24hPercentChange   float64                      `json:"defi_24h_percentage_change"`
	StablecoinVolume24h    float64                      `json:"stablecoin_volume_24h"`
	StablecoinVolume24hRep float64                      `json:"stablecoin_volume_24h_reported"`
	StablecoinMarketCap    float64                      `json:"stablecoin_market_cap"`
	Stablecoin24hPctChange float64                      `json:"stablecoin_24h_percentage_change"`
	DerivativesVolume24h   float64                      `json:"derivatives_volume_24h"`
	Derivatives24hPctChg   float64                      `json:"derivatives_24h_percentage_change"`
	Quote                  map[string]GlobalMetricsQuote `json:"quote"`
	LastUpdated            string                       `json:"last_updated"`
}

// GlobalMetricsQuote holds convert-scoped totals for the global metrics.
type GlobalMetricsQuote struct {
	TotalMarketCap         float64 `json:"total_market_cap"`
	TotalVolume24h         float64 `json:"total_volume_24h"`
	TotalVolume24hReported float64 `json:"total_volume_24h_reported"`
	AltcoinVolume24h       float64 `json:"altcoin_volume_24h"`
	AltcoinMarketCap       float64 `json:"altcoin_market_cap"`
	LastUpdated            string  `json:"last_updated"`
}

// GlobalMetricsCompact is the stable, compact JSON output shape for agent consumption.
type GlobalMetricsCompact struct {
	ActiveCryptocurrencies int     `json:"active_cryptocurrencies"`
	ActiveExchanges        int     `json:"active_exchanges"`
	ActiveMarketPairs      int     `json:"active_market_pairs"`
	BTCDominance           float64 `json:"btc_dominance"`
	ETHDominance           float64 `json:"eth_dominance"`
	TotalMarketCap         float64 `json:"total_market_cap"`
	TotalVolume24h         float64 `json:"total_volume_24h"`
	AltcoinVolume24h       float64 `json:"altcoin_volume_24h"`
	AltcoinMarketCap       float64 `json:"altcoin_market_cap"`
	DeFiVolume24h          float64 `json:"defi_volume_24h,omitempty"`
	DeFiMarketCap          float64 `json:"defi_market_cap,omitempty"`
	StablecoinVolume24h    float64 `json:"stablecoin_volume_24h,omitempty"`
	StablecoinMarketCap    float64 `json:"stablecoin_market_cap,omitempty"`
	DerivativesVolume24h   float64 `json:"derivatives_volume_24h,omitempty"`
	Convert                string  `json:"convert"`
	LastUpdated            string  `json:"last_updated,omitempty"`
}
