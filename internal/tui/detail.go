package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openCMC/CoinMarketCap-CLI/internal/api"
	"github.com/openCMC/CoinMarketCap-CLI/internal/display"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DetailModel struct {
	client  *api.Client
	coinID  string
	vs      string
	window  chartWindow
	info    *api.CoinInfo
	quote   *api.QuoteCoin
	ohlc    api.OHLCData
	loading int
	Done    bool
	err     error
	status  string
	width   int
	height  int
}

type coinInfoMsg struct {
	info *api.CoinInfo
	err  error
}

type quoteDetailMsg struct {
	quote *api.QuoteCoin
	err   error
}

type ohlcMsg struct {
	data   api.OHLCData
	err    error
	status string
}

func NewDetailModel(client *api.Client, coinID, vs string, width, height int) DetailModel {
	return DetailModel{
		client:  client,
		coinID:  coinID,
		vs:      strings.ToUpper(vs),
		window:  chartWindow1H,
		loading: 3,
		status:  "Loading asset detail…",
		width:   width,
		height:  height,
	}
}

func (m DetailModel) Init() tea.Cmd {
	return tea.Batch(m.fetchInfo(), m.fetchQuote(), m.fetchOHLC(), refreshTick())
}

func (m DetailModel) fetchInfo() tea.Cmd {
	return func() tea.Msg {
		info, err := m.client.InfoByID(context.Background(), m.coinID)
		return coinInfoMsg{info: info, err: err}
	}
}

func (m DetailModel) fetchQuote() tea.Cmd {
	return func() tea.Msg {
		quotes, err := m.client.QuotesLatestByID(context.Background(), []string{m.coinID}, m.vs)
		if err != nil {
			return quoteDetailMsg{err: err}
		}
		quote, ok := quotes[m.coinID]
		if !ok {
			return quoteDetailMsg{err: api.ErrAssetNotFound}
		}
		return quoteDetailMsg{quote: &quote}
	}
}

func (m DetailModel) fetchOHLC() tea.Cmd {
	return func() tea.Msg {
		data, status, err := m.loadOHLC(context.Background(), m.window)
		return ohlcMsg{data: data, status: status, err: err}
	}
}

func (m DetailModel) loadOHLC(ctx context.Context, window chartWindow) (api.OHLCData, string, error) {
	now := time.Now().UTC()
	attempts := window.fetchAttempts(now)

	var firstErr error
	for _, attempt := range attempts {
		asset, err := m.client.OHLCVHistoricalByID(ctx, m.coinID, m.vs, attempt.timePeriod, attempt.start, attempt.end, attempt.count, attempt.interval)
		if err == nil {
			rows := convertOHLCVAsset(asset, m.vs)
			if len(rows) > 0 {
				return rows, window.RefreshTextWithNote(attempt.fallbackNote), nil
			}
		} else if firstErr == nil {
			firstErr = err
		}

		quoteAsset, qErr := m.client.QuotesHistoricalByID(ctx, m.coinID, m.vs, attempt.start, attempt.end, attempt.count, attempt.interval)
		if qErr == nil {
			rows := quoteHistoryToOHLC(quoteAsset, m.vs)
			if len(rows) > 0 {
				return rows, window.RefreshTextWithNote(attempt.fallbackNote), nil
			}
		} else if firstErr == nil {
			firstErr = qErr
		}
	}

	if firstErr == nil {
		firstErr = api.ErrAssetNotFound
	}
	return nil, "", firstErr
}

func convertOHLCVAsset(asset *api.HistoricalOHLCVAsset, vs string) api.OHLCData {
	if asset == nil {
		return nil
	}
	rows := make(api.OHLCData, 0, len(asset.Quotes))
	for _, point := range asset.Quotes {
		values, ok := point.Quote[vs]
		if !ok {
			continue
		}
		ts, parseErr := time.Parse(time.RFC3339, point.TimeOpen)
		if parseErr != nil {
			continue
		}
		rows = append(rows, []float64{
			float64(ts.UnixMilli()),
			values.Open,
			values.High,
			values.Low,
			values.Close,
		})
	}
	return rows
}

func (m DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case coinInfoMsg:
		if m.loading > 0 {
			m.loading--
		}
		if msg.err != nil {
			m.err = msg.err
			m.status = fmt.Sprintf("Info refresh failed: %v", msg.err)
		} else {
			m.info = msg.info
		}

	case quoteDetailMsg:
		if m.loading > 0 {
			m.loading--
		}
		if msg.err != nil {
			m.err = msg.err
			m.status = fmt.Sprintf("Quote refresh failed: %v", msg.err)
		} else {
			m.quote = msg.quote
		}

	case ohlcMsg:
		if m.loading > 0 {
			m.loading--
		}
		if msg.err != nil {
			m.err = msg.err
			m.status = fmt.Sprintf("Chart refresh failed: %v", msg.err)
		} else {
			m.err = nil
			m.ohlc = msg.data
			if msg.status != "" {
				m.status = msg.status
			} else {
				m.status = m.window.RefreshText()
			}
		}

	case refreshTickMsg:
		m.loading = 3
		m.status = "Refreshing asset detail…"
		return m, tea.Batch(m.fetchInfo(), m.fetchQuote(), m.fetchOHLC(), refreshTick())

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			m.Done = true
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = 3
			m.status = "Refreshing asset detail…"
			return m, tea.Batch(m.fetchInfo(), m.fetchQuote(), m.fetchOHLC())
		case "p":
			m.window = m.window.Next()
			m.ohlc = nil
			m.err = nil
			m.status = m.window.LoadingText()
			return m, m.fetchOHLC()
		}
	}
	return m, nil
}

func (m DetailModel) View() string {
	if m.loading > 0 && m.quote == nil && m.info == nil {
		return renderLoading(fmt.Sprintf("Fetching detail for %s…", m.coinID), m.width, m.height)
	}
	if m.quote == nil && m.info == nil {
		body := "No detail data available."
		if m.err != nil {
			body = fmt.Sprintf("Error: %v", m.err)
		}
		return renderPlaceholder(m.width, m.height, "Detail", body)
	}

	infoName := m.coinID
	infoSymbol := "—"
	infoSlug := "—"
	infoCategory := "—"
	infoDescription := ""
	if m.info != nil {
		if m.info.Name != "" {
			infoName = m.info.Name
		}
		if m.info.Symbol != "" {
			infoSymbol = m.info.Symbol
		}
		if m.info.Slug != "" {
			infoSlug = m.info.Slug
		}
		if m.info.Category != "" {
			infoCategory = m.info.Category
		}
		infoDescription = m.info.Description
	}

	price := "—"
	marketCap := "—"
	volume := "—"
	change := "—"
	if m.quote != nil {
		if q, ok := m.quote.Quote[m.vs]; ok {
			price = display.FormatPrice(q.Price, m.vs)
			marketCap = display.FormatLargeNumber(q.MarketCap, m.vs)
			volume = display.FormatLargeNumber(q.Volume24h, m.vs)
			change = display.ColorPercent(q.PercentChange24h)
		}
	}

	leftWidth := 38
	if m.width > 0 {
		leftWidth = m.width * 32 / 100
		if leftWidth < 34 {
			leftWidth = 34
		}
	}

	var left strings.Builder
	left.WriteString("\n")
	addDetailField(&left, "ID", m.coinID)
	addDetailField(&left, "Name", display.SanitizeCell(infoName))
	addDetailField(&left, "Symbol", display.FormatSymbol(infoSymbol))
	addDetailField(&left, "Slug", display.SanitizeCell(infoSlug))
	addDetailField(&left, "Category", display.SanitizeCell(infoCategory))
	left.WriteString("\n")
	addDetailField(&left, "Price", price)
	addDetailField(&left, "Mkt Cap", marketCap)
	addDetailField(&left, "Vol 24h", volume)
	addDetailField(&left, "24h Chg", change)
	if infoDescription != "" {
		left.WriteString("\n")
		addDetailField(&left, "Summary", truncate(display.SanitizeCell(infoDescription), 140))
	}
	left.WriteString("\n")
	left.WriteString(DimStyle.Render("  p window    r refresh    Esc back"))

	infoBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(CMCBlue).
		Width(leftWidth - 2).
		PaddingLeft(1)
	leftPanel := infoBox.SetString(LabelStyle.Render(" Info ")).Render(left.String())

	var right strings.Builder
	right.WriteString("\n")
	if len(m.ohlc) > 0 {
		chartWidth := m.width - leftWidth - 12
		if chartWidth < 20 {
			chartWidth = 40
		}
		chartHeight := m.height - 12
		if chartHeight < 8 {
			chartHeight = 12
		}
		right.WriteString(renderBrailleChart(m.ohlc, chartWidth, chartHeight, m.vs, m.window))
		right.WriteString("\n")
	} else {
		right.WriteString(DimStyle.Render("  No chart data available"))
		right.WriteString("\n")
	}
	if m.status != "" {
		right.WriteString("\n")
		right.WriteString(DimStyle.Render("  " + m.status))
	}

	chartBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(CMCBlue).
		Width(m.width - leftWidth - 8).
		PaddingLeft(1)
	rightPanel := chartBox.SetString(LabelStyle.Render(" " + m.window.Title() + " ")).Render(right.String())

	title := BrandTitle(fmt.Sprintf("%s (%s) — Detail", display.SanitizeCell(infoName), display.FormatSymbol(infoSymbol)))
	content := title + "\n\n" + lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
	return renderFrame(m.width, m.height, content)
}

func addDetailField(b *strings.Builder, label, value string) {
	fmt.Fprintf(b, " %-12s %s\n", LabelStyle.Render(label), value)
}
