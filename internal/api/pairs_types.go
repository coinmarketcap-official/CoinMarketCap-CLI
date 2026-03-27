package api

import (
	"strings"
)

type MarketPairsLatestResponse struct {
	Data MarketPairsLatestData `json:"data"`
}

type MarketPairsLatestData struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Symbol      string       `json:"symbol"`
	Slug        string       `json:"slug"`
	MarketPairs []MarketPair `json:"market_pairs"`
	NumMarkets  int          `json:"num_markets,omitempty"`
	MarketShare float64      `json:"market_share,omitempty"`
}

type MarketPair struct {
	MarketPair      string                     `json:"market_pair"`
	ExchangeID      int64                      `json:"exchange_id"`
	ExchangeName    string                     `json:"exchange_name"`
	ExchangeSlug    string                     `json:"exchange_slug"`
	Exchange        MarketPairExchange         `json:"exchange"`
	Category        string                     `json:"category"`
	FeeType         string                     `json:"fee_type"`
	MarketPairBase  MarketPairAsset            `json:"market_pair_base"`
	MarketPairQuote MarketPairAsset            `json:"market_pair_quote"`
	Quote           map[string]MarketPairQuote `json:"quote"`
}

type MarketPairExchange struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type MarketPairAsset struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	Slug            string `json:"slug"`
	ContractAddress string `json:"contract_address"`
}

type MarketPairQuote struct {
	Price     float64 `json:"price"`
	Volume24h float64 `json:"volume_24h"`
}

type MarketPairsLatestRequest struct {
	ID       string
	Category string
	Limit    int
	Convert  string
}

func (p MarketPair) PairLabel() string {
	if label := strings.TrimSpace(p.MarketPair); label != "" {
		return label
	}
	parts := make([]string, 0, 2)
	if p.MarketPairBase.Symbol != "" {
		parts = append(parts, strings.ToUpper(p.MarketPairBase.Symbol))
	}
	if p.MarketPairQuote.Symbol != "" {
		parts = append(parts, strings.ToUpper(p.MarketPairQuote.Symbol))
	}
	return strings.Join(parts, "/")
}

func (p MarketPair) ExchangeLabel() string {
	if label := strings.TrimSpace(p.Exchange.Name); label != "" {
		return label
	}
	if label := strings.TrimSpace(p.Exchange.Slug); label != "" {
		return label
	}
	if label := strings.TrimSpace(p.ExchangeName); label != "" {
		return label
	}
	return strings.TrimSpace(p.ExchangeSlug)
}

func (p MarketPair) QuoteFor(convert string) (MarketPairQuote, bool) {
	if p.Quote == nil {
		return MarketPairQuote{}, false
	}
	convert = strings.ToUpper(strings.TrimSpace(convert))
	if quote, ok := p.Quote[convert]; ok {
		return quote, true
	}
	return MarketPairQuote{}, false
}
