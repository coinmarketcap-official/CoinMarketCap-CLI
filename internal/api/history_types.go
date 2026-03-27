package api

import "encoding/json"

type QuotesHistoricalResponse struct {
	Data map[string]HistoricalQuoteAsset `json:"-"`
}

func (r *QuotesHistoricalResponse) UnmarshalJSON(b []byte) error {
	r.Data = make(map[string]HistoricalQuoteAsset)
	return unmarshalHistoricalData(b, r.Data)
}

type HistoricalQuoteAsset struct {
	ID     int64                  `json:"id"`
	Name   string                 `json:"name"`
	Symbol string                 `json:"symbol"`
	Slug   string                 `json:"slug"`
	Quotes []HistoricalQuotePoint `json:"quotes"`
}

type HistoricalQuotePoint struct {
	Timestamp string                       `json:"timestamp"`
	Quote     map[string]HistoricalQuoteUSD `json:"quote"`
}

type HistoricalQuoteUSD struct {
	Price     float64 `json:"price"`
	MarketCap float64 `json:"market_cap"`
	Volume24h float64 `json:"volume_24h"`
}

type OHLCVHistoricalResponse struct {
	Data map[string]HistoricalOHLCVAsset `json:"-"`
}

func (r *OHLCVHistoricalResponse) UnmarshalJSON(b []byte) error {
	r.Data = make(map[string]HistoricalOHLCVAsset)
	return unmarshalHistoricalData(b, r.Data)
}

type HistoricalOHLCVAsset struct {
	ID     int64                  `json:"id"`
	Name   string                 `json:"name"`
	Symbol string                 `json:"symbol"`
	Slug   string                 `json:"slug"`
	Quotes []HistoricalOHLCVPoint `json:"quotes"`
}

type HistoricalOHLCVPoint struct {
	TimeOpen  string                        `json:"time_open"`
	TimeClose string                        `json:"time_close"`
	Quote     map[string]HistoricalOHLCVUSD `json:"quote"`
}

type HistoricalOHLCVUSD struct {
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
	MarketCap float64 `json:"market_cap"`
	Timestamp string  `json:"timestamp"`
}

// unmarshalHistoricalData handles CMC's polymorphic data field.
// CMC returns either:
//   - map keyed by string (slug or ID): {"data": {"bitcoin": {...}}} or {"data": {"1": {...}}}
//   - single flat object: {"data": {"id": 1, "name": "Bitcoin", ...}}
//
// We detect the shape by checking for the "id" key with a numeric value.
func unmarshalHistoricalData[T any](envelope []byte, out map[string]T) error {
	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(envelope, &wrapper); err != nil {
		return err
	}

	// Probe data shape: decode into map[string]json.RawMessage to inspect keys.
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(wrapper.Data, &probe); err != nil {
		return err
	}

	// If data has an "id" key whose value is a number, it's a single asset object.
	if raw, ok := probe["id"]; ok {
		var n json.Number
		if json.Unmarshal(raw, &n) == nil {
			// Single asset object — unmarshal the whole data blob as T.
			var single T
			if err := json.Unmarshal(wrapper.Data, &single); err != nil {
				return err
			}
			out["_single"] = single
			return nil
		}
	}

	// Map format — unmarshal directly.
	return json.Unmarshal(wrapper.Data, &out)
}
