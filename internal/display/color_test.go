package display

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVisibleWidth(t *testing.T) {
	tests := []struct {
		input  string
		expect int
	}{
		{"hello", 5},
		{"", 0},
		{"\033[32mgreen\033[0m", 5},
		{"\033[38;2;75;204;0mbrand\033[0m", 5},
		{"no color", 8},
		{"\033[31m-3.10%\033[0m", 6},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expect, VisibleWidth(tt.input), "input=%q", tt.input)
	}
}

func TestColorPercent(t *testing.T) {
	pos := ColorPercent(5.25)
	assert.Contains(t, pos, "5.25%")

	neg := ColorPercent(-3.1)
	assert.Contains(t, neg, "-3.10%")

	zero := ColorPercent(0)
	assert.Equal(t, "0.00%", zero)
}
