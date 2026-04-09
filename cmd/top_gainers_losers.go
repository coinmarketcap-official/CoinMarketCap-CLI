package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/openCMC/CoinMarketCap-CLI/internal/api"
	"github.com/openCMC/CoinMarketCap-CLI/internal/display"

	"github.com/spf13/cobra"
)

var topGainersLosersCmd = &cobra.Command{
	Use:   "top-gainers-losers",
	Short: "Show trending gainers and losers from CMC",
	Long: `Fetch CoinMarketCap's trending gainers/losers list.

This command uses CMC's trending endpoint directly and supports offset/limit
pagination, a constrained time-period selector, JSON output, and CSV export.`,
	Example: `  cmc top-gainers-losers
  cmc top-gainers-losers --time-period 1h
  cmc top-gainers-losers --limit 50 --time-period 7d
  cmc top-gainers-losers --start 10 --convert EUR
  cmc top-gainers-losers --export gainers.csv`,
	RunE: runTopGainersLosers,
}

func init() {
	topGainersLosersCmd.Flags().Int("start", 1, "Offset for pagination")
	topGainersLosersCmd.Flags().Int("limit", 100, "Number of results (1-200)")
	topGainersLosersCmd.Flags().String("time-period", "24h", "Time period (1h, 24h, 7d, 30d)")
	topGainersLosersCmd.Flags().String("convert", "USD", "Target fiat currency")
	topGainersLosersCmd.Flags().String("export", "", "Export to CSV file path")
	addOutputFlag(topGainersLosersCmd)
	addDryRunFlag(topGainersLosersCmd)
	rootCmd.AddCommand(topGainersLosersCmd)
}

func runTopGainersLosers(cmd *cobra.Command, args []string) error {
	start, _ := cmd.Flags().GetInt("start")
	limit, _ := cmd.Flags().GetInt("limit")
	timePeriod, _ := cmd.Flags().GetString("time-period")
	convert, _ := cmd.Flags().GetString("convert")
	exportPath, _ := cmd.Flags().GetString("export")
	jsonOut := outputJSON(cmd)

	if !jsonOut {
		display.PrintBanner()
	}

	if start < 1 {
		return fmt.Errorf("--start must be at least 1")
	}
	if limit < 1 || limit > 200 {
		return fmt.Errorf("--limit must be between 1 and 200")
	}
	if !isValidTopMoversPeriod(timePeriod) {
		return fmt.Errorf("invalid --time-period %q — must be one of: 1h, 24h, 7d, 30d", timePeriod)
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
			"start":       fmt.Sprintf("%d", start),
			"limit":       fmt.Sprintf("%d", limit),
			"time_period": timePeriod,
			"convert":     convert,
		}
		return printDryRun(cfg, "top-gainers-losers", "/v1/cryptocurrency/trending/gainers-losers", params, nil)
	}

	client := newAPIClient(cfg)
	ctx := cmd.Context()

	coins, err := client.TrendingGainersLosers(ctx, start, limit, timePeriod, convert)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(bucketTopMoversJSON(coins, timePeriod, convert))
	}

	displayCurrency := strings.ToLower(convert)
	fmt.Printf("Trending Gainers/Losers (%s, convert %s, start %d, limit %d)\n\n", timePeriod, convert, start, limit)

	headers := []string{"#", "Name", "Symbol", "Price", strings.ToLower(timePeriod) + " %", "Volume 24h", "Market Cap"}
	rows := make([][]string, 0, len(coins))
	var csvRows [][]string
	if exportPath != "" {
		csvRows = make([][]string, 0, len(coins))
	}
	for i := range coins {
		idx := fmt.Sprintf("%d", i+1)
		name := display.SanitizeCell(coins[i].Name)
		symbol := display.FormatSymbol(coins[i].Symbol)
		quote := coins[i].Quote[convert]
		periodChange := topMoverPercentChange(quote, timePeriod)
		rows = append(rows, []string{
			idx, name, symbol,
			display.FormatPrice(quote.Price, displayCurrency),
			display.ColorPercent(periodChange),
			display.FormatLargeNumber(quote.Volume24h, displayCurrency),
			display.FormatLargeNumber(quote.MarketCap, displayCurrency),
		})
		if exportPath != "" {
			csvRows = append(csvRows, []string{
				idx, name, symbol,
				fmt.Sprintf("%.8f", quote.Price),
				fmt.Sprintf("%.2f", periodChange),
				fmt.Sprintf("%.2f", quote.Volume24h),
				fmt.Sprintf("%.2f", quote.MarketCap),
			})
		}
	}

	display.PrintTable(headers, rows)

	if exportPath != "" {
		if err := exportCSV(exportPath, headers, csvRows); err != nil {
			return err
		}
	}

	return nil
}

func isValidTopMoversPeriod(period string) bool {
	switch period {
	case "1h", "24h", "7d", "30d":
		return true
	default:
		return false
	}
}

type topGainersLosersJSON struct {
	TimePeriod string              `json:"time_period"`
	Gainers    []api.TrendingCoin2 `json:"gainers"`
	Losers     []api.TrendingCoin2 `json:"losers"`
}

func bucketTopMoversJSON(coins []api.TrendingCoin2, timePeriod, convert string) topGainersLosersJSON {
	gainers := make([]api.TrendingCoin2, 0, len(coins))
	losers := make([]api.TrendingCoin2, 0, len(coins))

	for _, coin := range coins {
		quote, ok := coin.Quote[convert]
		if !ok {
			continue
		}
		if topMoverPercentChange(quote, timePeriod) < 0 {
			losers = append(losers, coin)
			continue
		}
		gainers = append(gainers, coin)
	}

	sort.SliceStable(gainers, func(i, j int) bool {
		return topMoverPercentChange(gainers[i].Quote[convert], timePeriod) > topMoverPercentChange(gainers[j].Quote[convert], timePeriod)
	})
	sort.SliceStable(losers, func(i, j int) bool {
		return topMoverPercentChange(losers[i].Quote[convert], timePeriod) < topMoverPercentChange(losers[j].Quote[convert], timePeriod)
	})

	return topGainersLosersJSON{
		TimePeriod: timePeriod,
		Gainers:    gainers,
		Losers:     losers,
	}
}

func topMoverPercentChange(quote api.Quote2, timePeriod string) float64 {
	switch timePeriod {
	case "1h":
		return quote.PercentChange1h
	case "7d":
		return quote.PercentChange7d
	case "30d":
		return quote.PercentChange30d
	default:
		return quote.PercentChange24h
	}
}
