package display

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		price  float64
		expect string
	}{
		{0, "$0.00"},
		{1.5, "$1.50"},
		{42000.123, "$42,000.12"},
		{0.05, "$0.0500"},
		{0.00001234, "$0.00001234"},
		{-1.5, "-$1.50"},
		{1000000, "$1,000,000.00"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expect, FormatPrice(tt.price), "price=%v", tt.price)
	}
}

func TestFormatPercent(t *testing.T) {
	assert.Equal(t, "5.25%", FormatPercent(5.25))
	assert.Equal(t, "-3.10%", FormatPercent(-3.1))
	assert.Equal(t, "0.00%", FormatPercent(0))
}

func TestFormatLargeNumber(t *testing.T) {
	tests := []struct {
		n      float64
		expect string
	}{
		{0, "$0.00"},
		{500, "$500.00"},
		{1500, "$1.50K"},
		{1500000, "$1.50M"},
		{2500000000, "$2.50B"},
		{1200000000000, "$1.20T"},
		{-5000000, "-$5.00M"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expect, FormatLargeNumber(tt.n), "n=%v", tt.n)
	}
}

func TestFormatSupply(t *testing.T) {
	assert.Equal(t, "21.00M", FormatSupply(21000000))
	assert.Equal(t, "1.50B", FormatSupply(1500000000))
	assert.Equal(t, "100", FormatSupply(100))
}
