package cmd

import (
	"fmt"
	"strings"

	"github.com/openCMC/CoinMarketCap-CLI/internal/api"
	"github.com/openCMC/CoinMarketCap-CLI/internal/display"

	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Show global cryptocurrency market metrics",
	Long: `Fetch the latest global cryptocurrency market metrics from CoinMarketCap.

Includes active cryptocurrencies, exchanges, market pairs, BTC/ETH dominance,
total market cap, 24h volume, and altcoin volume.`,
	Example: `  cmc metrics
  cmc metrics --convert EUR
  cmc metrics --dry-run -o json
  cmc metrics -o table`,
	Args: cobra.NoArgs,
	RunE: runMetrics,
}

func init() {
	metricsCmd.Flags().String("convert", "USD", "Quote currency for market totals")
	addOutputFlag(metricsCmd)
	addDryRunFlag(metricsCmd)
	rootCmd.AddCommand(metricsCmd)
}

func runMetrics(cmd *cobra.Command, args []string) error {
	convert, _ := cmd.Flags().GetString("convert")
	convert = strings.ToUpper(strings.TrimSpace(convert))
	jsonOut := outputJSON(cmd)

	if !jsonOut {
		display.PrintBanner()
	}

	if convert == "" {
		return fmt.Errorf("--convert must not be empty")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if isDryRun(cmd) {
		params := map[string]string{
			"convert": convert,
		}
		return printDryRun(cfg, "metrics", "/v1/global-metrics/quotes/latest", params, nil)
	}

	client := newAPIClient(cfg)
	ctx := cmd.Context()

	data, err := client.GlobalMetricsLatest(ctx, convert)
	if err != nil {
		return err
	}

	quote, err := metricsQuoteForConvert(data, convert)
	if err != nil {
		return err
	}

	if jsonOut {
		compact := toMetricsCompact(data, convert, quote)
		return printJSONRaw(compact)
	}

	printMetricsTable(data, convert, quote)
	return nil
}

func toMetricsCompact(d *api.GlobalMetricsData, convert string, q api.GlobalMetricsQuote) api.GlobalMetricsCompact {
	c := api.GlobalMetricsCompact{
		ActiveCryptocurrencies: d.ActiveCryptocurrencies,
		ActiveExchanges:        d.ActiveExchanges,
		ActiveMarketPairs:      d.ActiveMarketPairs,
		BTCDominance:           d.BTCDominance,
		ETHDominance:           d.ETHDominance,
		DeFiVolume24h:          d.DeFiVolume24h,
		DeFiMarketCap:          d.DeFiMarketCap,
		StablecoinVolume24h:    d.StablecoinVolume24h,
		StablecoinMarketCap:    d.StablecoinMarketCap,
		DerivativesVolume24h:   d.DerivativesVolume24h,
		Convert:                convert,
		LastUpdated:            d.LastUpdated,
	}
	c.TotalMarketCap = q.TotalMarketCap
	c.TotalVolume24h = q.TotalVolume24h
	c.AltcoinVolume24h = q.AltcoinVolume24h
	c.AltcoinMarketCap = q.AltcoinMarketCap
	return c
}

func printMetricsTable(d *api.GlobalMetricsData, convert string, q api.GlobalMetricsQuote) {
	fmt.Printf("Global Market Metrics (convert %s)\n\n", convert)

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Active Cryptocurrencies", formatIntCommas(d.ActiveCryptocurrencies)},
		{"Active Exchanges", formatIntCommas(d.ActiveExchanges)},
		{"Active Market Pairs", formatIntCommas(d.ActiveMarketPairs)},
		{"BTC Dominance", display.FormatPercent(d.BTCDominance)},
		{"ETH Dominance", display.FormatPercent(d.ETHDominance)},
		{"Total Market Cap", display.FormatLargeNumber(q.TotalMarketCap, convert)},
		{"Total Volume 24h", display.FormatLargeNumber(q.TotalVolume24h, convert)},
		{"Altcoin Volume 24h", display.FormatLargeNumber(q.AltcoinVolume24h, convert)},
		{"Altcoin Market Cap", display.FormatLargeNumber(q.AltcoinMarketCap, convert)},
	}

	display.PrintTable(headers, rows)
}

func metricsQuoteForConvert(d *api.GlobalMetricsData, convert string) (api.GlobalMetricsQuote, error) {
	q, ok := d.Quote[convert]
	if !ok {
		return api.GlobalMetricsQuote{}, fmt.Errorf("quote data for convert %s not found", convert)
	}
	return q, nil
}

func formatIntCommas(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i := 0; i < len(s); i++ {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, s[i])
	}
	return string(result)
}
