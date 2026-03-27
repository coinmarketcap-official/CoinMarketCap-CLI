package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const oasRepo = "https://coinmarketcap.com/api/documentation/v1/"

type commandAnnotation struct {
	APIEndpoint     string
	APIEndpoints    map[string]string
	OASOperationID  string
	OASOperationIDs map[string]string
	OASSpec         string
	Transport       string // "rest" (default) or "websocket"
	PaidOnly        bool
	RequiresAuth    bool
}

var commandMeta = map[string]commandAnnotation{
	"price": {
		APIEndpoint: "/v2/cryptocurrency/quotes/latest",
		APIEndpoints: map[string]string{
			"explicit":              "/v2/cryptocurrency/quotes/latest",
			"with-info":             "/v2/cryptocurrency/info",
			"with-chain-stats":      "/v1/blockchain/statistics/latest",
			"positional_slug_first": "/v2/cryptocurrency/info",
			"positional_symbol":     "/v1/cryptocurrency/map",
		},
		OASOperationID: "getV2CryptocurrencyQuotesLatest",
		OASOperationIDs: map[string]string{
			"explicit":              "getV2CryptocurrencyQuotesLatest",
			"with-info":             "getV2CryptocurrencyInfo",
			"with-chain-stats":      "getV1BlockchainStatisticsLatest",
			"positional_slug_first": "getV2CryptocurrencyInfo",
			"positional_symbol":     "cryptocurrency-map",
		},
		OASSpec:      "coinmarketcap-v1",
		RequiresAuth: true,
	},
	"metrics": {
		APIEndpoint:    "/v1/global-metrics/quotes/latest",
		OASOperationID: "cmc_get_global_metrics_latest",
		OASSpec:        "coinmarketcap-v1",
		RequiresAuth:   true,
	},
	"markets": {
		APIEndpoint: "/v1/cryptocurrency/listings/latest",
		APIEndpoints: map[string]string{
			"default":         "/v1/cryptocurrency/listings/latest",
			"category_lookup": "/v1/cryptocurrency/categories",
			"category":        "/v1/cryptocurrency/category",
		},
		OASOperationID: "getV1CryptocurrencyListingsLatest",
		OASSpec:        "coinmarketcap-v1",
		RequiresAuth:   true,
	},
	"resolve": {
		APIEndpoints: map[string]string{
			"id":     "/v2/cryptocurrency/info",
			"slug":   "/v2/cryptocurrency/info",
			"symbol": "/v1/cryptocurrency/map",
		},
		OASOperationIDs: map[string]string{
			"id":     "getV2CryptocurrencyInfo",
			"slug":   "getV2CryptocurrencyInfo",
			"symbol": "cryptocurrency-map",
		},
		OASSpec:      "coinmarketcap-v1",
		RequiresAuth: true,
	},
	"trending": {
		APIEndpoint:    "/v1/cryptocurrency/trending/latest",
		OASOperationID: "getV1CryptocurrencyTrendingLatest",
		OASSpec:        "coinmarketcap-v1",
		Transport:      "rest",
		RequiresAuth:   true,
	},
	"news": {
		APIEndpoint:    "/v1/content/latest",
		OASOperationID: "getV1ContentLatest",
		OASSpec:        "coinmarketcap-v1",
		RequiresAuth:   true,
	},
	"history": {
		APIEndpoints: map[string]string{
			"--date":             "/v1/cryptocurrency/quotes/historical",
			"--days":             "/v1/cryptocurrency/quotes/historical",
			"--date --ohlc":      "/v2/cryptocurrency/ohlcv/historical",
			"--days --ohlc":      "/v2/cryptocurrency/ohlcv/historical",
			"--from/--to":        "/v1/cryptocurrency/quotes/historical",
			"--from/--to --ohlc": "/v2/cryptocurrency/ohlcv/historical",
		},
		OASOperationIDs: map[string]string{
			"--date":             "getV1CryptocurrencyQuotesHistorical",
			"--days":             "getV1CryptocurrencyQuotesHistorical",
			"--date --ohlc":      "getV2CryptocurrencyOhlcvHistorical",
			"--days --ohlc":      "getV2CryptocurrencyOhlcvHistorical",
			"--from/--to":        "getV1CryptocurrencyQuotesHistorical",
			"--from/--to --ohlc": "getV2CryptocurrencyOhlcvHistorical",
		},
		OASSpec:      "coinmarketcap-v1",
		RequiresAuth: true,
	},
	"top-gainers-losers": {
		APIEndpoint:    "/v1/cryptocurrency/trending/gainers-losers",
		OASOperationID: "getV1CryptocurrencyTrendingGainersLosers",
		OASSpec:        "coinmarketcap-v1",
		Transport:      "rest",
		RequiresAuth:   true,
	},
	"pairs": {
		APIEndpoint:    "/v1/cryptocurrency/market-pairs/latest",
		OASOperationID: "getV1CryptocurrencyMarketPairsLatest",
		OASSpec:        "coinmarketcap-v1",
		RequiresAuth:   true,
	},
	"monitor": {
		APIEndpoint:  "/v2/cryptocurrency/quotes/latest",
		Transport:    "rest",
		RequiresAuth: true,
	},
	"search": {
		APIEndpoints: map[string]string{
			"asset":         "/v1/cryptocurrency/map",
			"address_pair":  "/v4/dex/pairs/quotes/latest",
			"address_token": "/v4/dex/spot-pairs/latest",
		},
		OASSpec:      "coinmarketcap-v1",
		Transport:    "rest",
		RequiresAuth: true,
	},
}

type flagInfo struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Default     string   `json:"default"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// Flag enums are exposed in the command catalog for validation and docs.
var flagEnums = map[string]map[string][]string{
	"history": {
		"interval": {"5m", "hourly", "daily"},
	},
	"networks": {
		"sort-dir": {"asc", "desc"},
	},
	"top-gainers-losers": {
		"time-period": {"1h", "24h", "7d", "30d"},
	},
	"markets": {
		"sort":     {"market_cap", "price", "volume_24h"},
		"sort-dir": {"asc", "desc"},
	},
	"pairs": {
		"sort":     {"volume_24h", "liquidity", "no_of_transactions_24h"},
		"sort-dir": {"asc", "desc"},
	},
	"ohlcv": {
		"window": {"1h", "24h", "30d"},
	},
}

type commandInfo struct {
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Flags           []flagInfo        `json:"flags"`
	Examples        []string          `json:"examples,omitempty"`
	OutputFormats   []string          `json:"output_formats"`
	RequiresAuth    bool              `json:"requires_auth"`
	PaidOnly        bool              `json:"paid_only"`
	Transport       string            `json:"transport,omitempty"`
	APIEndpoint     string            `json:"api_endpoint,omitempty"`
	APIEndpoints    map[string]string `json:"api_endpoints,omitempty"`
	OASOperationID  string            `json:"oas_operation_id,omitempty"`
	OASOperationIDs map[string]string `json:"oas_operation_ids,omitempty"`
}

type commandCatalog struct {
	Version  string        `json:"version"`
	OASRepo  string        `json:"oas_repo"`
	Commands []commandInfo `json:"commands"`
}

var commandsCmd = &cobra.Command{
	Use:    "commands",
	Short:  "Output machine-readable command catalog (JSON)",
	Hidden: true,
	RunE:   runCommands,
}

func init() {
	addOutputFlag(commandsCmd)
	rootCmd.AddCommand(commandsCmd)
}

func runCommands(cmd *cobra.Command, args []string) error {
	catalog := commandCatalog{
		Version: version,
		OASRepo: oasRepo,
	}

	for _, c := range rootCmd.Commands() {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}

		// Skip non-data commands.
		if c.Name() == "auth" || c.Name() == "status" || c.Name() == "version" {
			info := commandInfo{
				Name:          c.Name(),
				Description:   c.Short,
				Flags:         extractFlags(c),
				Examples:      splitExamples(c.Example),
				OutputFormats: []string{},
			}
			if c.Name() == "status" {
				info.OutputFormats = []string{"json", "table"}
			}
			if c.Name() == "version" {
				info.OutputFormats = []string{"table", "json"}
			}
			catalog.Commands = append(catalog.Commands, info)
			continue
		}

		// Handle subcommands (tui markets, tui trending, dex subcommands).
		if c.HasSubCommands() {
			for _, sub := range c.Commands() {
				if sub.Hidden {
					continue
				}
				outputFormats := []string{"json", "table"}
				if c.Name() == "tui" {
					outputFormats = []string{"tui"}
				}
				if c.Name() == "dex" && sub.Name() == "pair" {
					outputFormats = []string{"json"}
				}
				info := commandInfo{
					Name:          c.Name() + " " + sub.Name(),
					Description:   sub.Short,
					Flags:         extractFlags(sub),
					Examples:      splitExamples(sub.Example),
					OutputFormats: outputFormats,
					RequiresAuth:  true,
				}
				if meta, ok := commandMeta[info.Name]; ok {
					info.PaidOnly = meta.PaidOnly
					info.RequiresAuth = meta.RequiresAuth
					info.Transport = meta.Transport
					info.APIEndpoint = meta.APIEndpoint
					info.APIEndpoints = meta.APIEndpoints
					info.OASOperationID = meta.OASOperationID
					info.OASOperationIDs = meta.OASOperationIDs
				}
				catalog.Commands = append(catalog.Commands, info)
			}
			continue
		}

		info := commandInfo{
			Name:          c.Name(),
			Description:   c.Short,
			Flags:         extractFlags(c),
			Examples:      splitExamples(c.Example),
			OutputFormats: []string{"json", "table"},
			RequiresAuth:  true,
		}

		if meta, ok := commandMeta[c.Name()]; ok {
			info.PaidOnly = meta.PaidOnly
			info.RequiresAuth = meta.RequiresAuth
			info.Transport = meta.Transport
			info.APIEndpoint = meta.APIEndpoint
			info.APIEndpoints = meta.APIEndpoints
			info.OASOperationID = meta.OASOperationID
			info.OASOperationIDs = meta.OASOperationIDs
		}

		catalog.Commands = append(catalog.Commands, info)
	}

	return printJSONRaw(catalog)
}

func extractFlags(cmd *cobra.Command) []flagInfo {
	var flags []flagInfo
	cmd.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		fi := flagInfo{
			Name:        f.Name,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Description: f.Usage,
		}
		if enums, ok := flagEnums[cmd.Name()]; ok {
			if e, ok := enums[f.Name]; ok {
				fi.Enum = e
			}
		}
		flags = append(flags, fi)
	})
	return flags
}

func splitExamples(s string) []string {
	if s == "" {
		return nil
	}
	var examples []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimLeft(line, " ")
		if line != "" {
			examples = append(examples, line)
		}
	}
	return examples
}
