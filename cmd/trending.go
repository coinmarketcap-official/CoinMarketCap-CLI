package cmd

import (
	"fmt"
	"strings"

	"github.com/coinmarketcap/coinmarketcap-cli/internal/display"

	"github.com/spf13/cobra"
)

var trendingCmd = &cobra.Command{
	Use:   "trending",
	Short: "Show trending assets from CMC",
	Long: `Fetch the latest CoinMarketCap trending asset surface.

The command returns a ranked list of trending assets from the provider-backed
trending/latest endpoint.`,
	Example: `  cmc trending
  cmc trending --limit 3
  cmc trending --dry-run -o json`,
	RunE: runTrending,
}

func init() {
	trendingCmd.Flags().Int("limit", 50, "Number of trending assets to fetch from CMC (1-50)")
	trendingCmd.Flags().String("convert", "USD", "Target fiat currency")
	addOutputFlag(trendingCmd)
	addDryRunFlag(trendingCmd)
	rootCmd.AddCommand(trendingCmd)
}

func runTrending(cmd *cobra.Command, args []string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	convert, _ := cmd.Flags().GetString("convert")
	jsonOut := outputJSON(cmd)

	if !jsonOut {
		display.PrintBanner()
	}

	if limit < 1 || limit > 50 {
		return fmt.Errorf("--limit must be between 1 and 50")
	}
	convert = strings.ToUpper(strings.TrimSpace(convert))
	if convert == "" {
		return fmt.Errorf("--convert must not be empty")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if isDryRun(cmd) {
		params := map[string]string{
			"start":   "1",
			"limit":   fmt.Sprintf("%d", limit),
			"convert": convert,
		}
		return printDryRun(cfg, "trending", "/v1/cryptocurrency/trending/latest", params, nil)
	}

	client := newAPIClient(cfg)
	ctx := cmd.Context()

	coins, err := client.TrendingLatest(ctx, 1, limit, convert)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(coins)
	}

	fmt.Printf("Trending Assets (CMC, convert %s, limit %d)\n\n", convert, limit)
	headers := []string{"#", "Name", "Symbol", "Price", "Market Cap", "Volume", "24h"}
	rows := make([][]string, 0, len(coins))
	for i, coin := range coins {
		quote, ok := coin.Quote[convert]
		if !ok {
			continue
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			display.SanitizeCell(coin.Name),
			display.FormatSymbol(coin.Symbol),
			display.FormatPrice(quote.Price, convert),
			display.FormatLargeNumber(quote.MarketCap, convert),
			display.FormatLargeNumber(quote.Volume24h, convert),
			display.ColorPercent(quote.PercentChange24h),
		})
	}
	display.PrintTable(headers, rows)

	return nil
}
