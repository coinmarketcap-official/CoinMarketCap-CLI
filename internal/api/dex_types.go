package api

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

type DEXNetwork struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type DEXNetworksResponse struct {
	Data []DEXNetwork `json:"data"`
}

type DEXListingQuote struct {
	Volume24h   float64 `json:"volume_24h"`
	MarketShare float64 `json:"market_share"`
}

type DEXListing struct {
	ID          int64                      `json:"id"`
	Name        string                     `json:"name"`
	Slug        string                     `json:"slug"`
	NumMarkets  int                        `json:"num_markets"`
	MarketShare float64                    `json:"market_share"`
	Quote       map[string]DEXListingQuote `json:"quote"`
}

type DEXListingsQuotesResponse struct {
	Data []DEXListing `json:"data"`
}

type DEXAsset struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	Slug            string `json:"slug"`
	ContractAddress string `json:"contract_address"`
}

type DEXPairQuote struct {
	Price               float64 `json:"price"`
	Volume24h           float64 `json:"volume_24h"`
	Liquidity           float64 `json:"liquidity"`
	PercentChange24h    float64 `json:"percent_change_24h"`
	NoOfTransactions24h float64 `json:"no_of_transactions_24h"`
	Open                float64 `json:"open"`
	High                float64 `json:"high"`
	Low                 float64 `json:"low"`
	Close               float64 `json:"close"`
	MarketCap           float64 `json:"market_cap"`
	FullyDilutedValue   float64 `json:"fully_diluted_value"`
}

type DEXPair struct {
	ContractAddress string                  `json:"contract_address"`
	NetworkID       int64                   `json:"network_id"`
	NetworkName     string                  `json:"network_name"`
	NetworkSlug     string                  `json:"network_slug"`
	ScrollID        string                  `json:"scroll_id"`
	DEXID           int64                   `json:"dex_id"`
	DEXName         string                  `json:"dex_name"`
	DEXSlug         string                  `json:"dex_slug"`
	BaseAsset       DEXAsset                `json:"base_asset"`
	QuoteAsset      DEXAsset                `json:"quote_asset"`
	Quote           map[string]DEXPairQuote `json:"quote"`
}

func (p *DEXPair) UnmarshalJSON(data []byte) error {
	type wirePair struct {
		ContractAddress string          `json:"contract_address"`
		NetworkID       json.RawMessage `json:"network_id"`
		NetworkName     string          `json:"network_name"`
		NetworkSlug     string          `json:"network_slug"`
		ScrollID        string          `json:"scroll_id"`
		DEXID           json.RawMessage `json:"dex_id"`
		DEXName         string          `json:"dex_name"`
		DEXSlug         string          `json:"dex_slug"`
		BaseAsset       json.RawMessage `json:"base_asset"`
		QuoteAsset      json.RawMessage `json:"quote_asset"`
		Quote           json.RawMessage `json:"quote"`

		BaseAssetID              json.RawMessage `json:"base_asset_id"`
		BaseAssetUCID            json.RawMessage `json:"base_asset_ucid"`
		BaseAssetName            string          `json:"base_asset_name"`
		BaseAssetSymbol          string          `json:"base_asset_symbol"`
		BaseAssetContractAddress string          `json:"base_asset_contract_address"`

		QuoteAssetID              json.RawMessage `json:"quote_asset_id"`
		QuoteAssetUCID            json.RawMessage `json:"quote_asset_ucid"`
		QuoteAssetName            string          `json:"quote_asset_name"`
		QuoteAssetSymbol          string          `json:"quote_asset_symbol"`
		QuoteAssetContractAddress string          `json:"quote_asset_contract_address"`
	}

	var wire wirePair
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	p.ContractAddress = wire.ContractAddress
	p.NetworkID = parseJSONInt64(wire.NetworkID)
	p.NetworkName = wire.NetworkName
	p.NetworkSlug = wire.NetworkSlug
	p.ScrollID = wire.ScrollID
	p.DEXID = parseJSONInt64(wire.DEXID)
	p.DEXName = wire.DEXName
	p.DEXSlug = wire.DEXSlug

	if len(bytes.TrimSpace(wire.BaseAsset)) > 0 && !bytes.Equal(bytes.TrimSpace(wire.BaseAsset), []byte("null")) {
		if err := json.Unmarshal(wire.BaseAsset, &p.BaseAsset); err != nil {
			return err
		}
	}
	if len(bytes.TrimSpace(wire.QuoteAsset)) > 0 && !bytes.Equal(bytes.TrimSpace(wire.QuoteAsset), []byte("null")) {
		if err := json.Unmarshal(wire.QuoteAsset, &p.QuoteAsset); err != nil {
			return err
		}
	}

	if p.BaseAsset == (DEXAsset{}) {
		p.BaseAsset = DEXAsset{
			ID:              parseJSONInt64(wire.BaseAssetID),
			Name:            wire.BaseAssetName,
			Symbol:          wire.BaseAssetSymbol,
			ContractAddress: wire.BaseAssetContractAddress,
		}
	}
	if p.QuoteAsset == (DEXAsset{}) {
		p.QuoteAsset = DEXAsset{
			ID:              parseJSONInt64(wire.QuoteAssetID),
			Name:            wire.QuoteAssetName,
			Symbol:          wire.QuoteAssetSymbol,
			ContractAddress: wire.QuoteAssetContractAddress,
		}
	}

	if len(bytes.TrimSpace(wire.Quote)) > 0 && !bytes.Equal(bytes.TrimSpace(wire.Quote), []byte("null")) {
		quotes, err := parseDEXPairQuoteMap(wire.Quote)
		if err != nil {
			return err
		}
		p.Quote = quotes
	}
	return nil
}

func (p DEXPair) QuoteFor(convertID string) DEXPairQuote {
	if p.Quote == nil {
		return DEXPairQuote{}
	}
	if quote, ok := p.Quote[convertID]; ok {
		return quote
	}
	for _, quote := range p.Quote {
		return quote
	}
	return DEXPairQuote{}
}

func (p DEXPair) PairLabel() string {
	parts := make([]string, 0, 2)
	if p.BaseAsset.Symbol != "" {
		parts = append(parts, strings.ToUpper(p.BaseAsset.Symbol))
	}
	if p.QuoteAsset.Symbol != "" {
		parts = append(parts, strings.ToUpper(p.QuoteAsset.Symbol))
	}
	return strings.Join(parts, "/")
}

type DEXPairsResponse struct {
	Data []DEXPair `json:"data"`
}

type DEXSpotPairsLatestResponse struct {
	Data     []DEXPair `json:"data"`
	ScrollID string    `json:"scroll_id"`
}

func (r *DEXSpotPairsLatestResponse) UnmarshalJSON(data []byte) error {
	type wireResponse struct {
		Data     []DEXPair `json:"data"`
		ScrollID string    `json:"scroll_id"`
	}
	var wire wireResponse
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	r.Data = wire.Data
	r.ScrollID = wire.ScrollID
	if r.ScrollID == "" {
		for i := len(r.Data) - 1; i >= 0; i-- {
			if r.Data[i].ScrollID != "" {
				r.ScrollID = r.Data[i].ScrollID
				break
			}
		}
	}
	return nil
}

type DEXOHLCVPoint struct {
	TimeOpen  string                  `json:"time_open"`
	TimeClose string                  `json:"time_close"`
	Quote     map[string]DEXPairQuote `json:"quote"`
}

type DEXOHLCVHistoricalResponse struct {
	Data []DEXOHLCVPoint `json:"data"`
}

type DEXTrade struct {
	TradeTimestamp  string                  `json:"trade_timestamp"`
	Type            string                  `json:"type"`
	TransactionHash string                  `json:"transaction_hash"`
	Quote           map[string]DEXPairQuote `json:"quote"`
}

func (t DEXTrade) QuoteFor(convertID string) DEXPairQuote {
	if t.Quote == nil {
		return DEXPairQuote{}
	}
	if quote, ok := t.Quote[convertID]; ok {
		return quote
	}
	for _, quote := range t.Quote {
		return quote
	}
	return DEXPairQuote{}
}

func parseJSONInt64(raw json.RawMessage) int64 {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return 0
	}
	var num json.Number
	if err := json.Unmarshal(raw, &num); err == nil {
		if value, err := num.Int64(); err == nil {
			return value
		}
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if s = strings.TrimSpace(s); s != "" {
			if value, err := strconv.ParseInt(s, 10, 64); err == nil {
				return value
			}
		}
	}
	return 0
}

func parseDEXPairQuoteMap(raw json.RawMessage) (map[string]DEXPairQuote, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}
	if raw[0] == '{' {
		quotes := map[string]DEXPairQuote{}
		if err := json.Unmarshal(raw, &quotes); err != nil {
			return nil, err
		}
		return quotes, nil
	}
	if raw[0] != '[' {
		return nil, nil
	}
	type wireQuote struct {
		ConvertID string `json:"convert_id"`
		DEXPairQuote
	}
	var items []wireQuote
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	quotes := make(map[string]DEXPairQuote, len(items))
	for _, item := range items {
		if item.ConvertID == "" {
			continue
		}
		quotes[item.ConvertID] = item.DEXPairQuote
	}
	return quotes, nil
}

type DEXTradesLatestResponse struct {
	Data []DEXTrade `json:"data"`
}

type DEXSpotPairsLatestRequest struct {
	NetworkID       string
	NetworkSlug     string
	DEXID           string
	DEXSlug         string
	ContractAddress string
	ScrollID        string
	Limit           int
	Sort            string
	SortDir         string
	ConvertID       string
}

type DEXPairLookupRequest struct {
	NetworkID       string
	NetworkSlug     string
	ContractAddress string
	ConvertID       string
}

type DEXOHLCVHistoricalRequest struct {
	NetworkID       string
	NetworkSlug     string
	ContractAddress string
	TimePeriod      string
	Interval        string
	TimeStart       string
	TimeEnd         string
	Count           int
	ConvertID       string
}

type DEXTradeLatestRequest struct {
	NetworkID       string
	NetworkSlug     string
	ContractAddress string
	Limit           int
	ConvertID       string
}
