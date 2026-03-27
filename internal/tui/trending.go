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

type trendingState int

const (
	trendingLoading trendingState = iota
	trendingLoaded
	trendingDetail
)

type trendingCoin struct {
	ID               int64
	Slug             string
	Symbol           string
	Name             string
	Price            float64
	PercentChange24h float64
}

type TrendingModel struct {
	client          *api.Client
	vs              string
	limit           int
	coins           []trendingCoin
	cursor          int
	state           trendingState
	detail          DetailModel
	close           closeMode
	ReturnToLanding bool
	err             error
	status          string
	width           int
	height          int
}

type trendingLoadedMsg struct {
	coins []trendingCoin
	err   error
}

func NewTrendingModel(client *api.Client, vs string, mode ...closeMode) TrendingModel {
	exitMode := closeQuit
	if len(mode) > 0 {
		exitMode = mode[0]
	}
	return TrendingModel{
		client: client,
		vs:     strings.ToUpper(vs),
		limit:  50,
		close:  exitMode,
		state:  trendingLoading,
		status: "Loading trending assets…",
	}
}

func (m TrendingModel) Init() tea.Cmd {
	return tea.Batch(m.fetchTrending(), refreshTick())
}

func (m TrendingModel) fetchTrending() tea.Cmd {
	return func() tea.Msg {
		coins, err := m.client.TrendingLatest(context.Background(), 1, m.limit, m.vs)
		if err != nil {
			return trendingLoadedMsg{err: err}
		}

		trendingCoins := make([]trendingCoin, 0, len(coins))
		for _, coin := range coins {
			quote, ok := coin.Quote[m.vs]
			if !ok {
				continue
			}
			trendingCoins = append(trendingCoins, trendingCoin{
				ID:               coin.ID,
				Slug:             coin.Slug,
				Symbol:           coin.Symbol,
				Name:             coin.Name,
				Price:            quote.Price,
				PercentChange24h: quote.PercentChange24h,
			})
		}
		return trendingLoadedMsg{coins: trendingCoins}
	}
}

func (m TrendingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case trendingLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = trendingLoaded
			m.status = fmt.Sprintf("Refresh failed: %v", msg.err)
			return m, nil
		}
		m.err = nil
		m.coins = msg.coins
		m.state = trendingLoaded
		m.status = fmt.Sprintf("Updated %d trending assets", len(msg.coins))
		if m.cursor >= len(m.coins) && len(m.coins) > 0 {
			m.cursor = len(m.coins) - 1
		}
		return m, nil

	case coinInfoMsg, quoteDetailMsg, ohlcMsg:
		if m.state == trendingDetail {
			updated, cmd := m.detail.Update(msg)
			m.detail = updated.(DetailModel)
			return m, cmd
		}

	case refreshTickMsg:
		if m.state == trendingDetail {
			updated, cmd := m.detail.Update(msg)
			m.detail = updated.(DetailModel)
			return m, cmd
		}
		return m, tea.Batch(m.fetchTrending(), refreshTick())

	case tea.KeyMsg:
		if m.state == trendingDetail {
			updated, cmd := m.detail.Update(msg)
			detail := updated.(DetailModel)
			if detail.Done {
				if m.close == closeToLanding {
					m.ReturnToLanding = true
					return m, nil
				}
				m.state = trendingLoaded
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
			m.status = "Refreshing trending assets…"
			return m, m.fetchTrending()
		case "enter":
			if len(m.coins) > 0 {
				coin := m.coins[m.cursor]
				width, height := normalizeViewSize(m.width, m.height)
				m.detail = NewDetailModel(m.client, strconv.FormatInt(coin.ID, 10), m.vs, width, height)
				m.state = trendingDetail
				return m, m.detail.Init()
			}
		}
	}

	return m, nil
}

func (m TrendingModel) View() string {
	width, height := normalizeViewSize(m.width, m.height)

	if m.state == trendingLoading && len(m.coins) == 0 {
		return renderLoading("Fetching trending assets…", width, height)
	}
	if m.state == trendingDetail {
		return m.detail.View()
	}
	if len(m.coins) == 0 {
		body := "No trending data available."
		if m.err != nil {
			body = fmt.Sprintf("Error: %v", m.err)
		}
		return renderPlaceholder(width, height, "Trending", body)
	}

	var b strings.Builder
	b.WriteString(BrandTitle(fmt.Sprintf("TUI — Top %d Trending Assets", m.limit)))
	b.WriteString("\n\n")

	header := fmt.Sprintf("  %-6s %-20s %-10s %14s %10s", "Trend", "Name", "Symbol", "Price", "24h")
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	visibleRows := listVisibleRows(height)
	start, end := listWindow(len(m.coins), visibleRows, m.cursor)

	for i := start; i < end; i++ {
		coin := m.coins[i]
		change := fmt.Sprintf("%10s", display.FormatPercent(coin.PercentChange24h))
		change = ColorPercent(coin.PercentChange24h, change)
		row := fmt.Sprintf("%-6s %-20s %-10s %14s %s",
			fmt.Sprintf("#%d", i+1),
			truncate(display.SanitizeCell(coin.Name), 20),
			display.FormatSymbol(coin.Symbol),
			display.FormatPrice(coin.Price, m.vs),
			change,
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
	content := b.String() + "\n" + status
	return renderFrame(width, height, content)
}
