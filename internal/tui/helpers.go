package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	defaultViewWidth  = 120
	defaultViewHeight = 40
	minListRows       = 5
)

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

func renderLoading(msg string, width, height int) string {
	width, height = normalizeViewSize(width, height)

	content := BrandTitle("Loading…") + "\n\n"
	content += lipgloss.Place(
		width-4, height-6,
		lipgloss.Center, lipgloss.Center,
		DimStyle.Render(msg),
	)
	return renderFrame(width, height, content)
}

func normalizeViewSize(width, height int) (int, int) {
	if width <= 0 {
		width = defaultViewWidth
	}
	if height <= 0 {
		height = defaultViewHeight
	}
	return width, height
}

func listVisibleRows(height int) int {
	_, height = normalizeViewSize(0, height)

	visible := height - 8
	if visible < minListRows {
		return minListRows
	}
	return visible
}

func listWindow(total, visibleRows, cursor int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	if visibleRows <= 0 {
		visibleRows = 1
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}
	if visibleRows >= total {
		return 0, total
	}

	start := cursor - visibleRows + 1
	if start < 0 {
		start = 0
	}
	end := start + visibleRows
	if end > total {
		end = total
		start = end - visibleRows
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

func landingLogoView(width int) string {
	width, _ = normalizeViewSize(width, 0)

	if width < 94 {
		return LandingLogoStyle.Width(width).Render("COINMARKETCAP")
	}

	lines := landingLogoLines()
	block := strings.Join(lines, "\n")
	return LandingLogoStyle.Width(width).Render(block)
}

var landingLogoGlyphs = map[rune][]string{
	'C': {
		" ███ ",
		"█   █",
		"█    ",
		"█   █",
		" ███ ",
	},
	'O': {
		" ███ ",
		"█   █",
		"█   █",
		"█   █",
		" ███ ",
	},
	'I': {
		"█████",
		"  █  ",
		"  █  ",
		"  █  ",
		"█████",
	},
	'N': {
		"█   █",
		"██  █",
		"█ █ █",
		"█  ██",
		"█   █",
	},
	'M': {
		"█   █",
		"██ ██",
		"█ █ █",
		"█   █",
		"█   █",
	},
	'A': {
		" ███ ",
		"█   █",
		"█████",
		"█   █",
		"█   █",
	},
	'R': {
		"████ ",
		"█   █",
		"████ ",
		"█  █ ",
		"█   █",
	},
	'K': {
		"█   █",
		"█  █ ",
		"███  ",
		"█  █ ",
		"█   █",
	},
	'E': {
		"█████",
		"█    ",
		"████ ",
		"█    ",
		"█████",
	},
	'T': {
		"█████",
		"  █  ",
		"  █  ",
		"  █  ",
		"  █  ",
	},
	'P': {
		"████ ",
		"█   █",
		"████ ",
		"█    ",
		"█    ",
	},
}

func landingLogoLines() []string {
	word := "COINMARKETCAP"
	rows := make([]string, 5)
	for _, r := range word {
		glyph, ok := landingLogoGlyphs[r]
		if !ok {
			continue
		}
		for i := range rows {
			if rows[i] != "" {
				rows[i] += " "
			}
			rows[i] += glyph[i]
		}
	}

	return rows
}
