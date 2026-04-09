package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/openCMC/CoinMarketCap-CLI/internal/api"
	"github.com/openCMC/CoinMarketCap-CLI/internal/display"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/graph"
	"github.com/charmbracelet/lipgloss"
)

type chartWindow int

const (
	chartWindow1H chartWindow = iota
	chartWindow24H
	chartWindow30D
	chartWindow7D
)

type chartFetchAttempt struct {
	timePeriod   string
	interval     string
	start        time.Time
	end          time.Time
	count        int
	fallbackNote string
}

func (w chartWindow) String() string {
	switch w {
	case chartWindow24H:
		return "24H"
	case chartWindow30D:
		return "30D"
	case chartWindow7D:
		return "7D"
	default:
		return "1H"
	}
}

func (w chartWindow) Next() chartWindow {
	switch w {
	case chartWindow1H:
		return chartWindow24H
	case chartWindow24H:
		return chartWindow30D
	default:
		return chartWindow1H
	}
}

func (w chartWindow) Title() string {
	return w.String() + " Price"
}

func (w chartWindow) LoadingText() string {
	return "Loading " + w.Title() + "…"
}

func (w chartWindow) RefreshText() string {
	return w.Title() + " updated"
}

func (w chartWindow) RefreshTextWithNote(note string) string {
	if note == "" {
		return w.RefreshText()
	}
	return fmt.Sprintf("%s (%s)", w.RefreshText(), note)
}

func (w chartWindow) AxisLabels() (string, string, string) {
	switch w {
	case chartWindow24H:
		return "0h", "12h", "24h"
	case chartWindow30D:
		return "0d", "15d", "30d"
	case chartWindow7D:
		return "0d", "3d", "7d"
	default:
		return "0m", "30m", "60m"
	}
}

func (w chartWindow) fetchAttempts(now time.Time) []chartFetchAttempt {
	switch w {
	case chartWindow7D:
		return chartFetchAttempts(now, 7*24*time.Hour, []chartFetchAttempt{
			{timePeriod: "daily", interval: "daily"},
		})
	case chartWindow24H:
		return chartFetchAttempts(now, 24*time.Hour, []chartFetchAttempt{
			{timePeriod: "hourly", interval: "5m"},
			{timePeriod: "hourly", interval: "hourly", fallbackNote: "hourly fallback"},
		})
	case chartWindow30D:
		return chartFetchAttempts(now, 30*24*time.Hour, []chartFetchAttempt{
			{timePeriod: "daily", interval: "daily"},
		})
	default:
		return chartFetchAttempts(now, time.Hour, []chartFetchAttempt{
			{timePeriod: "hourly", interval: "5m"},
			{timePeriod: "hourly", interval: "hourly", fallbackNote: "hourly fallback"},
		})
	}
}

func chartFetchAttempts(now time.Time, duration time.Duration, attempts []chartFetchAttempt) []chartFetchAttempt {
	out := make([]chartFetchAttempt, 0, len(attempts))
	for _, attempt := range attempts {
		step := chartStep(attempt.interval)
		count := int(duration/step) + 1
		if count < 2 {
			count = 2
		}
		out = append(out, chartFetchAttempt{
			timePeriod:   attempt.timePeriod,
			interval:     attempt.interval,
			start:        now.Add(-duration).UTC(),
			end:          now.UTC(),
			count:        count,
			fallbackNote: attempt.fallbackNote,
		})
	}
	return out
}

func chartStep(interval string) time.Duration {
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "5m":
		return 5 * time.Minute
	case "hourly":
		return time.Hour
	case "daily":
		return 24 * time.Hour
	default:
		return time.Hour
	}
}

func quoteHistoryToOHLC(asset *api.HistoricalQuoteAsset, vs string) api.OHLCData {
	if asset == nil {
		return nil
	}
	rows := make(api.OHLCData, 0, len(asset.Quotes))
	for _, point := range asset.Quotes {
		values, ok := point.Quote[vs]
		if !ok {
			continue
		}
		ts, parseErr := time.Parse(time.RFC3339, point.Timestamp)
		if parseErr != nil {
			continue
		}
		price := values.Price
		rows = append(rows, []float64{
			float64(ts.UnixMilli()),
			price, price, price, price,
		})
	}
	return rows
}

func renderBrailleChart(ohlc api.OHLCData, width, height int, vs string, window chartWindow) string {
	xLeft, xMid, xRight := window.AxisLabels()
	return renderBrailleChartWithLabels(ohlc, width, height, vs, xLeft, xMid, xRight)
}

func renderBrailleChartWithLabels(ohlc api.OHLCData, width, height int, vs, xLeft, xMid, xRight string) string {
	if len(ohlc) == 0 || width < 10 || height < 5 {
		return "No data"
	}

	// Extract close prices (index 4)
	prices := make([]float64, 0, len(ohlc))
	for _, d := range ohlc {
		if len(d) >= 5 {
			prices = append(prices, d[4])
		}
	}
	if len(prices) == 0 {
		return "No data"
	}

	minP, maxP := prices[0], prices[0]
	for _, p := range prices {
		if p < minP {
			minP = p
		}
		if p > maxP {
			maxP = p
		}
	}

	if maxP == minP {
		maxP = minP + 1
	}

	// Y-axis labels: high, mid, low
	midP := (minP + maxP) / 2
	yHigh := display.FormatPrice(maxP, vs)
	yMid := display.FormatPrice(midP, vs)
	yLow := display.FormatPrice(minP, vs)

	// Find the widest Y label for padding (display width, not byte length).
	yWidth := max(lipgloss.Width(yHigh), lipgloss.Width(yMid), lipgloss.Width(yLow)) + 1

	// Chart area dimensions (subtract Y-axis width and X-axis row)
	chartW := width - yWidth
	chartH := height - 2 // leave room for x-axis label row
	if chartW < 4 {
		chartW = 4
	}
	if chartH < 3 {
		chartH = 3
	}

	bg := graph.NewBrailleGrid(chartW, chartH,
		0, float64(len(prices)-1),
		minP, maxP,
	)

	for i := 1; i < len(prices); i++ {
		p1 := canvas.Float64Point{X: float64(i - 1), Y: prices[i-1]}
		p2 := canvas.Float64Point{X: float64(i), Y: prices[i]}
		gp1 := bg.GridPoint(p1)
		gp2 := bg.GridPoint(p2)

		points := graph.GetLinePoints(gp1, gp2)
		for _, pt := range points {
			bg.Set(pt)
		}
	}

	patterns := bg.BraillePatterns()

	// Color the chart line based on 7-day performance.
	var chartStyle lipgloss.Style
	if prices[len(prices)-1] >= prices[0] {
		chartStyle = GreenStyle
	} else {
		chartStyle = RedStyle
	}

	// Pre-compute styled Y-axis labels (right-aligned to yWidth).
	pad := strings.Repeat(" ", yWidth)
	styledHigh := DimStyle.Render(fmt.Sprintf("%*s", yWidth, yHigh))
	styledMid := DimStyle.Render(fmt.Sprintf("%*s", yWidth, yMid))
	styledLow := DimStyle.Render(fmt.Sprintf("%*s", yWidth, yLow))

	var b strings.Builder
	for i, row := range patterns {
		switch {
		case i == 0:
			b.WriteString(styledHigh)
		case i == len(patterns)/2:
			b.WriteString(styledMid)
		case i == len(patterns)-1:
			b.WriteString(styledLow)
		default:
			b.WriteString(pad)
		}
		b.WriteString(chartStyle.Render(string(row)))
		b.WriteString("\n")
	}

	gap := chartW - len(xLeft) - len(xMid) - len(xRight)
	if gap < 2 {
		gap = 2
	}
	leftGap := gap / 2
	rightGap := gap - leftGap
	xAxis := strings.Repeat(" ", yWidth) +
		xLeft + strings.Repeat(" ", leftGap) +
		xMid + strings.Repeat(" ", rightGap) +
		xRight
	b.WriteString(DimStyle.Render(xAxis))

	return b.String()
}
