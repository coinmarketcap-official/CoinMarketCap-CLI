package tui

import "github.com/charmbracelet/lipgloss"

// Brand colors matching CoinMarketCap identity.
var (
	CMCBlue = lipgloss.Color("#1765FF")
	Gold    = lipgloss.Color("#FFE866")
)

var (
	HeaderStyle          = lipgloss.NewStyle().Bold(true).Foreground(Gold)
	SelectedStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(CMCBlue)
	HelpStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	GreenStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	RedStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	DimStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	TitleStyle           = lipgloss.NewStyle().Bold(true).Foreground(CMCBlue)
	LabelStyle           = lipgloss.NewStyle().Bold(true).Foreground(Gold)
	LandingLogoStyle     = lipgloss.NewStyle().Bold(true).Foreground(CMCBlue).Align(lipgloss.Center)
	LandingTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(CMCBlue).Align(lipgloss.Center)
	LandingSubheadStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Align(lipgloss.Center)
	LandingCardStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(CMCBlue).Padding(1, 2)
	LandingEntryStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
	LandingSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(CMCBlue)
)

const HighlightSymbol = "▶ "

const listHelpText = "  ↑/k  ↓/j  navigate    Enter  detail    q/Esc  quit"

func ColorPercent(pct float64, s string) string {
	if pct > 0 {
		return GreenStyle.Render(s)
	} else if pct < 0 {
		return RedStyle.Render(s)
	}
	return s
}

// BrandTitle returns the branded title line: ◆ CoinMarketCap <subtitle>
func BrandTitle(subtitle string) string {
	return TitleStyle.Render(" ◆ CoinMarketCap") + " " + DimStyle.Render(subtitle)
}

// FrameStyle returns a bordered lipgloss style for the TUI outer frame.
func FrameStyle(width, height int) lipgloss.Style {
	width, height = normalizeViewSize(width, height)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(CMCBlue).
		Width(width - 2).
		Height(height - 2).
		PaddingLeft(1).
		PaddingRight(1)
}

// renderFrame wraps content in a branded frame and places it in the terminal.
func renderFrame(width, height int, content string) string {
	width, height = normalizeViewSize(width, height)

	frame := FrameStyle(width, height)
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, frame.Render(content))
}

// renderPlaceholder renders a branded frame with a title and message body.
func renderPlaceholder(width, height int, title, body string) string {
	width, height = normalizeViewSize(width, height)

	content := BrandTitle(title) + "\n\n" + body + "\n\nPress esc to go back."
	return renderFrame(width, height, content)
}
