package cmd

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coinmarketcap/coinmarketcap-cli/internal/api"
	"github.com/coinmarketcap/coinmarketcap-cli/internal/config"
	"github.com/coinmarketcap/coinmarketcap-cli/internal/display"

	"github.com/spf13/cobra"
)

const (
	historyDateLayout = "2006-01-02"
	maxHistoryPoints  = 100
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Get historical quotes or OHLCV data",
	Long: `Fetch historical market data from CoinMarketCap using one of three modes:
  --date YYYY-MM-DD
  --days N
  --from YYYY-MM-DD --to YYYY-MM-DD

Provide exactly one asset identity via --id, --slug, or --symbol.
The --to date is inclusive and covers the full UTC day through 23:59:59.`,
	Example: `  cmc history --id 1 --date 2024-01-01
  cmc history --id 1 --days 1 --interval 5m
  cmc history --slug bitcoin --days 30
  cmc history --symbol BTC --from 2024-01-01 --to 2024-03-01
  cmc history --id 1 --days 7 --ohlc --interval hourly
  cmc history --id 1 --days 30 --export history.csv`,
	Args: cobra.NoArgs,
	RunE: runHistory,
}

func init() {
	historyCmd.Flags().String("id", "", "Resolve an exact CoinMarketCap numeric ID")
	historyCmd.Flags().String("slug", "", "Resolve an exact slug")
	historyCmd.Flags().String("symbol", "", "Resolve a symbol, failing on ambiguity")
	historyCmd.Flags().String("date", "", "Snapshot date (YYYY-MM-DD)")
	historyCmd.Flags().String("days", "", "Data for the last N intervals")
	historyCmd.Flags().String("from", "", "Range start date (YYYY-MM-DD)")
	historyCmd.Flags().String("to", "", "Range end date (YYYY-MM-DD)")
	historyCmd.Flags().String("convert", "USD", "Target currency")
	historyCmd.Flags().String("interval", "daily", "Data interval: 5m, daily, or hourly")
	historyCmd.Flags().Bool("ohlc", false, "Output OHLCV candles instead of quote snapshots")
	historyCmd.Flags().String("export", "", "Export to CSV file path")
	addOutputFlag(historyCmd)
	addDryRunFlag(historyCmd)
	rootCmd.AddCommand(historyCmd)
}

type historyIdentity struct {
	id     string
	slug   string
	symbol string
}

type historyMode struct {
	date     string
	days     int
	maxDays  bool
	fromTime time.Time
	toTime   time.Time
}

type historyRange struct {
	start time.Time
	end   time.Time
	count int
}

type historyQuoteRow struct {
	ID        int64   `json:"id"`
	Slug      string  `json:"slug"`
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Timestamp string  `json:"timestamp"`
	Price     float64 `json:"price"`
	MarketCap float64 `json:"market_cap"`
	Volume24h float64 `json:"volume_24h"`
}

type historyOHLCRow struct {
	ID        int64   `json:"id"`
	Slug      string  `json:"slug"`
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Timestamp string  `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	MarketCap float64 `json:"market_cap"`
	Volume24h float64 `json:"volume_24h"`
}

func runHistory(cmd *cobra.Command, args []string) error {
	jsonOut := outputJSON(cmd)
	if !jsonOut {
		display.PrintBanner()
	}

	identity, err := parseHistoryIdentity(cmd)
	if err != nil {
		return err
	}
	mode, err := parseHistoryMode(cmd)
	if err != nil {
		return err
	}

	interval, _ := cmd.Flags().GetString("interval")
	interval = strings.ToLower(strings.TrimSpace(interval))
	if interval == "" {
		interval = "daily"
	}
	if interval != "daily" && interval != "hourly" && interval != "5m" {
		return fmt.Errorf("invalid --interval %q — must be 5m, daily or hourly", interval)
	}

	convert, _ := cmd.Flags().GetString("convert")
	convert = strings.ToUpper(strings.TrimSpace(convert))
	if convert == "" {
		convert = "USD"
	}

	ohlc, _ := cmd.Flags().GetBool("ohlc")
	exportPath, _ := cmd.Flags().GetString("export")
	if ohlc && interval == "5m" {
		return fmt.Errorf("--ohlc does not support --interval 5m")
	}
	if mode.maxDays {
		if ohlc {
			return fmt.Errorf("--days max does not support --ohlc")
		}
		if interval != "daily" {
			return fmt.Errorf("--days max only supports --interval daily")
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	ranges := buildHistoryRanges(mode, interval)
	if isDryRun(cmd) {
		return historyDryRun(cfg, identity, convert, interval, ohlc, ranges, mode)
	}
	client := newAPIClient(cfg)
	ctx := cmd.Context()
	id, err := resolveHistoryID(ctx, client, identity)
	if err != nil {
		return err
	}
	if mode.maxDays {
		info, err := client.InfoByID(ctx, id)
		if err != nil {
			return err
		}
		start, err := parseDateAdded(info.DateAdded)
		if err != nil {
			return err
		}
		ranges = splitHistoryRanges(start, time.Now().UTC(), interval)
	}

	if ohlc {
		rows, err := fetchOHLCHistory(ctx, client, id, convert, interval, ranges)
		if err != nil {
			return err
		}
		rows = normalizeOHLCHistoryRows(rows, mode)
		return outputOHLCHistory(rows, convert, exportPath, jsonOut)
	}

	rows, err := fetchQuoteHistory(ctx, client, id, convert, interval, ranges)
	if err != nil {
		return err
	}
	rows = normalizeQuoteHistoryRows(rows, mode)
	return outputQuoteHistory(rows, convert, exportPath, jsonOut)
}

func parseHistoryIdentity(cmd *cobra.Command) (historyIdentity, error) {
	id, _ := cmd.Flags().GetString("id")
	slug, _ := cmd.Flags().GetString("slug")
	symbol, _ := cmd.Flags().GetString("symbol")

	values := []string{id, slug, symbol}
	provided := 0
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			provided++
		}
	}
	if provided != 1 {
		return historyIdentity{}, fmt.Errorf("specify exactly one of --id, --slug, or --symbol")
	}

	return historyIdentity{
		id:     strings.TrimSpace(id),
		slug:   strings.TrimSpace(slug),
		symbol: strings.TrimSpace(symbol),
	}, nil
}

func parseHistoryMode(cmd *cobra.Command) (historyMode, error) {
	dateStr, _ := cmd.Flags().GetString("date")
	daysStr, _ := cmd.Flags().GetString("days")
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")

	modes := 0
	if strings.TrimSpace(dateStr) != "" {
		modes++
	}
	if strings.TrimSpace(daysStr) != "" {
		modes++
	}
	if strings.TrimSpace(fromStr) != "" || strings.TrimSpace(toStr) != "" {
		modes++
	}
	if modes != 1 {
		return historyMode{}, fmt.Errorf("specify exactly one mode: --date, --days, or --from/--to")
	}

	switch {
	case strings.TrimSpace(dateStr) != "":
		t, err := time.Parse(historyDateLayout, dateStr)
		if err != nil {
			return historyMode{}, fmt.Errorf("invalid --date date, use YYYY-MM-DD: %w", err)
		}
		return historyMode{date: dateStr, fromTime: t.UTC(), toTime: t.UTC()}, nil

	case strings.TrimSpace(daysStr) != "":
		if strings.EqualFold(strings.TrimSpace(daysStr), "max") {
			return historyMode{maxDays: true}, nil
		}
		days, err := strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			return historyMode{}, fmt.Errorf("invalid --days %q — must be a positive integer", daysStr)
		}
		return historyMode{days: days}, nil

	default:
		if strings.TrimSpace(fromStr) == "" || strings.TrimSpace(toStr) == "" {
			return historyMode{}, fmt.Errorf("both --from and --to are required for range mode")
		}
		fromTime, err := time.Parse(historyDateLayout, fromStr)
		if err != nil {
			return historyMode{}, fmt.Errorf("invalid --from date, use YYYY-MM-DD: %w", err)
		}
		toTime, err := time.Parse(historyDateLayout, toStr)
		if err != nil {
			return historyMode{}, fmt.Errorf("invalid --to date, use YYYY-MM-DD: %w", err)
		}
		if toTime.Before(fromTime) {
			return historyMode{}, fmt.Errorf("--to must be on or after --from")
		}
		return historyMode{fromTime: fromTime.UTC(), toTime: toTime.UTC()}, nil
	}
}

func buildHistoryRanges(mode historyMode, interval string) []historyRange {
	return buildHistoryRangesAt(mode, interval, time.Now().UTC())
}

func buildHistoryRangesAt(mode historyMode, interval string, now time.Time) []historyRange {
	now = now.UTC()
	step := historyStep(interval)
	alignedNow := now.Truncate(step)
	switch {
	case mode.date != "":
		start := mode.fromTime.UTC().Add(-step)
		end := endOfUTCDay(mode.fromTime)
		return splitHistoryRanges(start, end, interval)

	case mode.days > 0:
		start := alignedNow.Add(-time.Duration(mode.days) * step).UTC()
		return splitHistoryRanges(start, alignedNow, interval)
	case mode.maxDays:
		return nil
	default:
		start := mode.fromTime.UTC().Add(-step)
		end := endOfUTCDay(mode.toTime)
		return splitHistoryRanges(start, end, interval)
	}
}

func splitHistoryRanges(start, end time.Time, interval string) []historyRange {
	step := historyStep(interval)
	if step <= 0 {
		step = 24 * time.Hour
	}

	var ranges []historyRange
	cursor := start.UTC()
	for !cursor.After(end) {
		chunkStart := cursor
		chunkEnd := chunkStart.Add(time.Duration(maxHistoryPoints-1) * step)
		if chunkEnd.After(end) {
			chunkEnd = end
		}
		count := int(chunkEnd.Sub(chunkStart)/step) + 1
		if count < 1 {
			count = 1
		}
		ranges = append(ranges, historyRange{start: chunkStart, end: chunkEnd, count: count})
		cursor = chunkEnd.Add(step)
	}
	return ranges
}

func historyStep(interval string) time.Duration {
	if interval == "hourly" {
		return time.Hour
	}
	if interval == "5m" {
		return 5 * time.Minute
	}
	return 24 * time.Hour
}

func historyDryRun(cfg *config.Config, identity historyIdentity, convert, interval string, ohlc bool, ranges []historyRange, mode historyMode) error {
	paramsKey, paramsValue := historyDryRunIdentity(identity)
	requests := make([]struct {
		opKey    string
		endpoint string
		params   map[string]string
	}, 0, len(ranges))

	endpoint := "/v1/cryptocurrency/quotes/historical"
	opKey := "--days"
	if mode.date != "" {
		opKey = "--date"
	}
	if mode.days == 0 && mode.date == "" {
		opKey = "--from/--to"
	}
	if mode.maxDays {
		params := map[string]string{
			paramsKey:  paramsValue,
			"convert":  convert,
			"days":     "max",
			"interval": interval,
		}
		return printDryRunFull(cfg, "history", "--days", endpoint, params, nil, "time_start will be derived from asset date_added at runtime")
	}
	if ohlc {
		endpoint = "/v2/cryptocurrency/ohlcv/historical"
		switch opKey {
		case "--days":
			opKey = "--days --ohlc"
		case "--date":
			opKey = "--date --ohlc"
		default:
			opKey = "--from/--to --ohlc"
		}
	}

	for _, rng := range ranges {
		params := map[string]string{
			paramsKey:    paramsValue,
			"convert":    convert,
			"time_start": rng.start.UTC().Format(time.RFC3339),
			"time_end":   rng.end.UTC().Format(time.RFC3339),
			"count":      strconv.Itoa(rng.count),
			"interval":   interval,
		}
		if ohlc {
			params["time_period"] = interval
		}
		requests = append(requests, struct {
			opKey    string
			endpoint string
			params   map[string]string
		}{opKey: opKey, endpoint: endpoint, params: params})
	}

	if len(requests) == 1 {
		req := requests[0]
		return printDryRunWithOp(cfg, "history", req.opKey, req.endpoint, req.params, nil)
	}
	return printDryRunMulti(cfg, "history", requests)
}

func historyDryRunIdentity(identity historyIdentity) (string, string) {
	switch {
	case identity.id != "":
		return "id", identity.id
	case identity.slug != "":
		return "slug", identity.slug
	default:
		return "symbol", identity.symbol
	}
}

func resolveHistoryID(ctx context.Context, client *api.Client, identity historyIdentity) (string, error) {
	switch {
	case identity.id != "":
		return identity.id, nil
	case identity.slug != "":
		asset, err := client.ResolveBySlug(ctx, identity.slug)
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(asset.ID, 10), nil
	default:
		asset, err := client.ResolveBySymbol(ctx, identity.symbol)
		if err != nil {
			var ambig *api.ResolverAmbiguityError
			if errors.As(err, &ambig) {
				return "", fmt.Errorf("symbol %q is ambiguous; use --id or --slug instead (%w)", identity.symbol, api.ErrResolverAmbiguous)
			}
			return "", err
		}
		return strconv.FormatInt(asset.ID, 10), nil
	}
}

func fetchQuoteHistory(ctx context.Context, client *api.Client, id, convert, interval string, ranges []historyRange) ([]historyQuoteRow, error) {
	rowsByTimestamp := map[string]historyQuoteRow{}
	for _, rng := range ranges {
		asset, err := withHistoryRetry(ctx, func() (*api.HistoricalQuoteAsset, error) {
			return client.QuotesHistoricalByID(ctx, id, convert, rng.start, rng.end, rng.count, interval)
		})
		if err != nil {
			return nil, err
		}
		for _, quote := range asset.Quotes {
			values, ok := quote.Quote[convert]
			if !ok {
				continue
			}
			rowsByTimestamp[quote.Timestamp] = historyQuoteRow{
				ID:        asset.ID,
				Slug:      asset.Slug,
				Symbol:    asset.Symbol,
				Name:      asset.Name,
				Timestamp: quote.Timestamp,
				Price:     values.Price,
				MarketCap: values.MarketCap,
				Volume24h: values.Volume24h,
			}
		}
	}

	rows := make([]historyQuoteRow, 0, len(rowsByTimestamp))
	for _, row := range rowsByTimestamp {
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Timestamp < rows[j].Timestamp
	})
	return rows, nil
}

func fetchOHLCHistory(ctx context.Context, client *api.Client, id, convert, interval string, ranges []historyRange) ([]historyOHLCRow, error) {
	rowsByTimestamp := map[string]historyOHLCRow{}
	for _, rng := range ranges {
		asset, err := withHistoryRetry(ctx, func() (*api.HistoricalOHLCVAsset, error) {
			return client.OHLCVHistoricalByID(ctx, id, convert, interval, rng.start, rng.end, rng.count, interval)
		})
		if err != nil {
			return nil, err
		}
		for _, quote := range asset.Quotes {
			values, ok := quote.Quote[convert]
			if !ok {
				continue
			}
			rowsByTimestamp[quote.TimeOpen] = historyOHLCRow{
				ID:        asset.ID,
				Slug:      asset.Slug,
				Symbol:    asset.Symbol,
				Name:      asset.Name,
				Timestamp: quote.TimeOpen,
				Open:      values.Open,
				High:      values.High,
				Low:       values.Low,
				Close:     values.Close,
				MarketCap: values.MarketCap,
				Volume24h: values.Volume,
			}
		}
	}

	rows := make([]historyOHLCRow, 0, len(rowsByTimestamp))
	for _, row := range rowsByTimestamp {
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Timestamp < rows[j].Timestamp
	})
	return rows, nil
}

func normalizeQuoteHistoryRows(rows []historyQuoteRow, mode historyMode) []historyQuoteRow {
	switch {
	case mode.date != "":
		return normalizeHistoryDateRows(rows, mode.fromTime)
	case mode.days > 0:
		return normalizeHistoryDaysRows(rows, mode.days)
	case mode.maxDays:
		return dedupeHistoryRows(rows)
	default:
		return normalizeHistoryRangeRows(rows, mode.fromTime, mode.toTime)
	}
}

func normalizeOHLCHistoryRows(rows []historyOHLCRow, mode historyMode) []historyOHLCRow {
	switch {
	case mode.date != "":
		return normalizeHistoryDateRows(rows, mode.fromTime)
	case mode.days > 0:
		return normalizeHistoryDaysRows(rows, mode.days)
	case mode.maxDays:
		return dedupeHistoryRows(rows)
	default:
		return normalizeHistoryRangeRows(rows, mode.fromTime, mode.toTime)
	}
}

func normalizeHistoryDateRows[T any](rows []T, day time.Time) []T {
	filtered := filterHistoryRows(rows, day.UTC(), endOfUTCDay(day))
	if len(filtered) == 0 {
		return []T{}
	}
	return filtered[len(filtered)-1:]
}

func normalizeHistoryRangeRows[T any](rows []T, fromTime, toTime time.Time) []T {
	return filterHistoryRows(rows, fromTime.UTC(), endOfUTCDay(toTime))
}

func normalizeHistoryDaysRows[T any](rows []T, count int) []T {
	filtered := dedupeHistoryRows(rows)
	if count <= 0 || len(filtered) <= count {
		return filtered
	}
	return filtered[len(filtered)-count:]
}

func filterHistoryRows[T any](rows []T, fromTime, toTime time.Time) []T {
	var zero time.Time
	filtered := make([]T, 0, len(rows))
	for _, row := range dedupeHistoryRows(rows) {
		ts, ok := historyRowTimestamp(row)
		if !ok {
			continue
		}
		if !fromTime.Equal(zero) && ts.Before(fromTime) {
			continue
		}
		if !toTime.Equal(zero) && ts.After(toTime) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func dedupeHistoryRows[T any](rows []T) []T {
	type indexedRow struct {
		timestamp time.Time
		index     int
	}

	selected := make(map[string]indexedRow)
	for i, row := range rows {
		ts, ok := historyRowTimestamp(row)
		if !ok {
			continue
		}
		key := ts.UTC().Format(time.RFC3339Nano)
		if existing, found := selected[key]; !found || ts.After(existing.timestamp) {
			selected[key] = indexedRow{timestamp: ts, index: i}
		}
	}
	if len(selected) == 0 {
		return []T{}
	}
	result := make([]T, 0, len(selected))
	for _, entry := range selected {
		result = append(result, rows[entry.index])
	}
	sort.Slice(result, func(i, j int) bool {
		ti, _ := historyRowTimestamp(result[i])
		tj, _ := historyRowTimestamp(result[j])
		return ti.Before(tj)
	})
	return result
}

func historyRowTimestamp(row any) (time.Time, bool) {
	switch v := row.(type) {
	case historyQuoteRow:
		return parseHistoryRowTimestamp(v.Timestamp)
	case historyOHLCRow:
		return parseHistoryRowTimestamp(v.Timestamp)
	default:
		return time.Time{}, false
	}
}

func parseHistoryRowTimestamp(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed.UTC(), true
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC(), true
	}
	return time.Time{}, false
}

func endOfUTCDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)
}

func withHistoryRetry[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	var zero T
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}
		lastErr = err
		var rateErr *api.RateLimitError
		if !errors.As(err, &rateErr) || attempt == 2 {
			return zero, err
		}
		delay := time.Second
		if rateErr.RetryAfter > 0 {
			delay = time.Duration(rateErr.RetryAfter) * time.Second
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, ctx.Err()
		case <-timer.C:
		}
	}
	return zero, lastErr
}

func outputQuoteHistory(rows []historyQuoteRow, convert, exportPath string, jsonOut bool) error {
	if exportPath != "" {
		if err := exportCSV(exportPath, quoteHistoryCSVHeaders(), quoteHistoryCSVRows(rows)); err != nil {
			return err
		}
	}
	if jsonOut {
		return printJSONRaw(rows)
	}

	headers := quoteHistoryTableHeaders()
	tableRows := quoteHistoryTableRows(rows, convert)
	display.PrintTable(headers, tableRows)
	return nil
}

func outputOHLCHistory(rows []historyOHLCRow, convert, exportPath string, jsonOut bool) error {
	if exportPath != "" {
		if err := exportCSV(exportPath, ohlcHistoryCSVHeaders(), ohlcHistoryCSVRows(rows)); err != nil {
			return err
		}
	}
	if jsonOut {
		return printJSONRaw(rows)
	}

	headers := ohlcHistoryTableHeaders()
	tableRows := ohlcHistoryTableRows(rows, convert)
	display.PrintTable(headers, tableRows)
	return nil
}

func quoteHistoryTableHeaders() []string {
	return []string{"Timestamp", "ID", "Name", "Symbol", "Price", "Market Cap", "Volume 24h"}
}

func quoteHistoryTableRows(rows []historyQuoteRow, convert string) [][]string {
	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, []string{
			row.Timestamp,
			strconv.FormatInt(row.ID, 10),
			display.SanitizeCell(row.Name),
			display.FormatSymbol(row.Symbol),
			display.FormatPrice(row.Price, convert),
			display.FormatLargeNumber(row.MarketCap, convert),
			display.FormatLargeNumber(row.Volume24h, convert),
		})
	}
	return tableRows
}

func quoteHistoryCSVHeaders() []string {
	return []string{"Timestamp", "ID", "Name", "Symbol", "Price", "Market Cap", "Volume 24h"}
}

func quoteHistoryCSVRows(rows []historyQuoteRow) [][]string {
	csvRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		csvRows = append(csvRows, []string{
			row.Timestamp,
			strconv.FormatInt(row.ID, 10),
			row.Name,
			row.Symbol,
			strconv.FormatFloat(row.Price, 'f', -1, 64),
			strconv.FormatFloat(row.MarketCap, 'f', -1, 64),
			strconv.FormatFloat(row.Volume24h, 'f', -1, 64),
		})
	}
	return csvRows
}

func ohlcHistoryTableHeaders() []string {
	return []string{"Timestamp", "ID", "Name", "Symbol", "Open", "High", "Low", "Close", "Market Cap", "Volume 24h"}
}

func ohlcHistoryTableRows(rows []historyOHLCRow, convert string) [][]string {
	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, []string{
			row.Timestamp,
			strconv.FormatInt(row.ID, 10),
			display.SanitizeCell(row.Name),
			display.FormatSymbol(row.Symbol),
			display.FormatPrice(row.Open, convert),
			display.FormatPrice(row.High, convert),
			display.FormatPrice(row.Low, convert),
			display.FormatPrice(row.Close, convert),
			display.FormatLargeNumber(row.MarketCap, convert),
			display.FormatLargeNumber(row.Volume24h, convert),
		})
	}
	return tableRows
}

func ohlcHistoryCSVHeaders() []string {
	return []string{"Timestamp", "ID", "Name", "Symbol", "Open", "High", "Low", "Close", "Market Cap", "Volume 24h"}
}

func ohlcHistoryCSVRows(rows []historyOHLCRow) [][]string {
	csvRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		csvRows = append(csvRows, []string{
			row.Timestamp,
			strconv.FormatInt(row.ID, 10),
			row.Name,
			row.Symbol,
			strconv.FormatFloat(row.Open, 'f', -1, 64),
			strconv.FormatFloat(row.High, 'f', -1, 64),
			strconv.FormatFloat(row.Low, 'f', -1, 64),
			strconv.FormatFloat(row.Close, 'f', -1, 64),
			strconv.FormatFloat(row.MarketCap, 'f', -1, 64),
			strconv.FormatFloat(row.Volume24h, 'f', -1, 64),
		})
	}
	return csvRows
}


func parseDateAdded(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, fmt.Errorf("asset is missing date_added metadata required for --days max")
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid asset date_added %q: %w", value, err)
	}
	return parsed.UTC(), nil
}
