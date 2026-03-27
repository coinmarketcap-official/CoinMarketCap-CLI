package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/coinmarketcap/coinmarketcap-cli/internal/api"
	"github.com/coinmarketcap/coinmarketcap-cli/internal/display"

	tea "github.com/charmbracelet/bubbletea"
)

type marketsState int

const (
	marketsLoading marketsState = iota
	marketsLoaded
	marketsDetail
)

type closeMode int

const (
	closeQuit closeMode = iota
	closeToLanding
)

type MarketsModel struct {
	client          *api.Client
	convert         string
	total           int
	category        string
	coins           []api.ListingCoin
	cursor          int
	state           marketsState
	detail          DetailModel
	close           closeMode
	ReturnToLanding bool
	err             error
	status          string
	width           int
	height          int
}

type coinsLoadedMsg struct {
	coins []api.ListingCoin
	err   error
}

func NewMarketsModel(client *api.Client, convert string, total int, mode ...closeMode) MarketsModel {
	return NewMarketsModelWithCategory(client, convert, total, "", mode...)
}

func NewMarketsModelWithCategory(client *api.Client, convert string, total int, category string, mode ...closeMode) MarketsModel {
	exitMode := closeQuit
	if len(mode) > 0 {
		exitMode = mode[0]
	}
	return MarketsModel{
		client:   client,
		convert:  strings.ToUpper(convert),
		total:    total,
		category: strings.TrimSpace(category),
		close:    exitMode,
		state:    marketsLoading,
		status:   "Loading markets…",
	}
}

func (m MarketsModel) Init() tea.Cmd {
	return tea.Batch(m.fetchCoins(), refreshTick())
}

func (m MarketsModel) fetchCoins() tea.Cmd {
	return func() tea.Msg {
		coins, err := m.client.ListingsLatestWithCategory(context.Background(), 1, m.total, m.convert, "market_cap", "desc", m.category)
		return coinsLoadedMsg{coins: coins, err: err}
	}
}

func (m MarketsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case coinsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			if len(m.coins) == 0 {
				m.state = marketsLoaded
			}
			m.status = fmt.Sprintf("Refresh failed: %v", msg.err)
			return m, nil
		}
		m.err = nil
		m.coins = msg.coins
		m.state = marketsLoaded
		m.status = fmt.Sprintf("Updated %d markets", len(msg.coins))
		if m.cursor >= len(m.coins) && len(m.coins) > 0 {
			m.cursor = len(m.coins) - 1
		}
		return m, nil

	case coinInfoMsg, quoteDetailMsg, ohlcMsg:
		if m.state == marketsDetail {
			updated, cmd := m.detail.Update(msg)
			m.detail = updated.(DetailModel)
			return m, cmd
		}

	case refreshTickMsg:
		if m.state == marketsDetail {
			updated, cmd := m.detail.Update(msg)
			m.detail = updated.(DetailModel)
			return m, cmd
		}
		return m, tea.Batch(m.fetchCoins(), refreshTick())

	case tea.KeyMsg:
		if m.state == marketsDetail {
			updated, cmd := m.detail.Update(msg)
			detail := updated.(DetailModel)
			if detail.Done {
				if m.close == closeToLanding {
					m.ReturnToLanding = true
					return m, nil
				}
				m.state = marketsLoaded
				m.detail = detail
				return m, refreshTick()
			}
			m.detail = detail
			return m, cmd
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc", "backspace":
			if m.close == closeToLanding {
				m.ReturnToLanding = true
				return m, nil
			}
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.coins)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "r":
			m.status = "Refreshing markets…"
			return m, m.fetchCoins()
		case "enter":
			if m.state == marketsLoaded && len(m.coins) > 0 {
				coin := m.coins[m.cursor]
				width, height := normalizeViewSize(m.width, m.height)
				m.detail = NewDetailModel(m.client, strconv.FormatInt(coin.ID, 10), m.convert, width, height)
				m.state = marketsDetail
				return m, m.detail.Init()
			}
		}
	}

	return m, nil
}

func (m MarketsModel) View() string {
	width, height := normalizeViewSize(m.width, m.height)

	if m.state == marketsLoading && len(m.coins) == 0 {
		return renderLoading(fmt.Sprintf("Fetching top %d CMC listings%s…", m.total, marketsCategorySuffix(m.category)), width, height)
	}
	if m.state == marketsDetail {
		return m.detail.View()
	}
	if len(m.coins) == 0 {
		body := "No market data available."
		if m.err != nil {
			body = fmt.Sprintf("Error: %v", m.err)
		}
		return renderPlaceholder(width, height, "Markets", body)
	}

	var b strings.Builder
	b.WriteString(BrandTitle(fmt.Sprintf("TUI — Top %d Markets%s", len(m.coins), marketsCategorySuffix(m.category))))
	b.WriteString("\n\n")

	header := fmt.Sprintf("  %-4s %-20s %-10s %14s %12s %12s %10s",
		"#", "Name", "Symbol", "Price", "Market Cap", "Volume", "24h")
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	visibleRows := listVisibleRows(height)
	start, end := listWindow(len(m.coins), visibleRows, m.cursor)

	for i := start; i < end; i++ {
		coin := m.coins[i]
		quote := coin.Quote[m.convert]
		pctStr := fmt.Sprintf("%10s", display.FormatPercent(quote.PercentChange24h))
		pctStr = ColorPercent(quote.PercentChange24h, pctStr)

		row := fmt.Sprintf("%-4d %-20s %-10s %14s %12s %12s %s",
			coin.CMCRank,
			truncate(display.SanitizeCell(coin.Name), 20),
			truncate(display.FormatSymbol(coin.Symbol), 10),
			display.FormatPrice(quote.Price, m.convert),
			display.FormatLargeNumber(quote.MarketCap, m.convert),
			display.FormatLargeNumber(quote.Volume24h, m.convert),
			pctStr,
		)
		if i == m.cursor {
			row = SelectedStyle.Render(HighlightSymbol + row)
		} else {
			row = "  " + row
		}
		b.WriteString(row)
		b.WriteString("\n")
	}

	status := HelpStyle.Render(listHelpText + "    r  refresh")
	if m.status != "" {
		status += "\n" + DimStyle.Render("  "+m.status)
	}
	if m.category != "" {
		status += "\n" + DimStyle.Render("  category: "+m.category)
	}
	content := b.String() + "\n" + status
	return renderFrame(width, height, content)
}

func marketsCategorySuffix(category string) string {
	category = strings.TrimSpace(category)
	if category == "" {
		return ""
	}
	return " — " + category
}
