package cmd

import (
	"strings"

	"github.com/coinmarketcap/coinmarketcap-cli/internal/api"
)

func warnAutoPickedSymbol(symbol string, asset api.ResolvedAsset) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	warnf(
		"Warning: symbol %q matched multiple assets; selected top-ranked candidate %s (id %d, slug %s). Use --id or --slug to be explicit.\n",
		symbol,
		asset.Name,
		asset.ID,
		asset.Slug,
	)
}
