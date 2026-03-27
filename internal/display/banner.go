package display

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// Brand color: CoinMarketCap blue.
const (
	brandBlue  = "\033[38;2;23;101;255m"
	dimColor   = "\033[2m"
	cyanColor  = "\033[36m"
	yellowBold = "\033[1;33m"
	boxWidth   = 78 // inner width of the welcome box
)

var asciiLogo = []string{
	"   CoinMarketCap CLI",
}

// PrintLogo prints the full ASCII art CoinMarketCap logo in brand blue to stderr.
func PrintLogo() {
	if !ColorEnabled() {
		for _, line := range asciiLogo {
			_, _ = fmt.Fprintln(os.Stderr, line)
		}
		_, _ = fmt.Fprintln(os.Stderr)
		return
	}
	_, _ = fmt.Fprintln(os.Stderr)
	for _, line := range asciiLogo {
		_, _ = fmt.Fprintf(os.Stderr, "%s%s%s\n", brandBlue, line, colorReset)
	}
	_, _ = fmt.Fprintln(os.Stderr)
}

// PrintWelcomeBox prints a bordered quick-start box to stderr.
func PrintWelcomeBox() {
	w := os.Stderr
	top := "+" + strings.Repeat("-", boxWidth) + "+"
	blank := "|" + strings.Repeat(" ", boxWidth) + "|"
	sep := boxRow(w, dimColor+strings.Repeat("-", boxWidth-2)+colorReset, boxWidth-2)

	_, _ = fmt.Fprintln(w, top)
	_, _ = fmt.Fprintln(w, blank)
	printColoredRow(w, yellowBold+"Official API Command Line Interface"+colorReset, 35)
	_, _ = fmt.Fprintln(w, blank)
	_, _ = fmt.Fprintln(w, sep)
	_, _ = fmt.Fprintln(w, blank)
	printPlainRow(w, "  Quick Start")
	_, _ = fmt.Fprintln(w, blank)
	printCmdRow(w, "cmc auth", "# Set up your API key")
	printCmdRow(w, "cmc price --id 1", "# Get BTC price")
	printCmdRow(w, "cmc markets --limit 100", "# Top 100 by mkt cap")
	printCmdRow(w, "cmc resolve --symbol ETH", "# Resolve an asset")
	printCmdRow(w, "cmc trending", "# Trending tokens")
	printCmdRow(w, "cmc history --id 1 --days 30", "# 30-day price history")
	printCmdRow(w, "cmc top-gainers-losers", "# Top gainers")
	printCmdRow(w, "cmc monitor --id 1", "# Poll latest price")
	printCmdRow(w, "cmc tui markets", "# Interactive TUI")
	_, _ = fmt.Fprintln(w, blank)
	_, _ = fmt.Fprintln(w, sep)
	_, _ = fmt.Fprintln(w, blank)
	printColoredRow(w, "  "+dimColor+"Docs: "+colorReset+cyanColor+"https://coinmarketcap.com/api/documentation/v1/"+colorReset, 62)
	_, _ = fmt.Fprintln(w, blank)
	_, _ = fmt.Fprintln(w, top)
	_, _ = fmt.Fprintln(w)
}

func printPlainRow(w *os.File, text string) {
	pad := boxWidth - 2 - len(text)
	if pad < 0 {
		pad = 0
	}
	_, _ = fmt.Fprintf(w, "| %s%s |\n", text, strings.Repeat(" ", pad))
}

func printColoredRow(w *os.File, content string, visible int) {
	pad := boxWidth - 2 - visible
	if pad < 0 {
		pad = 0
	}
	if !ColorEnabled() {
		plain := ansiRegex.ReplaceAllString(content, "")
		plainPad := boxWidth - 2 - len(plain)
		if plainPad < 0 {
			plainPad = 0
		}
		_, _ = fmt.Fprintf(w, "| %s%s |\n", plain, strings.Repeat(" ", plainPad))
		return
	}
	_, _ = fmt.Fprintf(w, "| %s%s |\n", content, strings.Repeat(" ", pad))
}

func printCmdRow(w *os.File, cmd, comment string) {
	// Layout: "| " + "  " + "$" + " " + cmd(30) + " " + comment + pad + " |"
	cmdPad := 30 - len(cmd)
	if cmdPad < 0 {
		cmdPad = 0
	}
	commentPad := 41 - len(comment)
	if commentPad < 0 {
		commentPad = 0
	}
	if ColorEnabled() {
		_, _ = fmt.Fprintf(w, "|   %s$%s %s%s %s%s%s |\n",
			brandBlue, colorReset,
			cmd, strings.Repeat(" ", cmdPad),
			dimColor, comment, colorReset+strings.Repeat(" ", commentPad))
	} else {
		_, _ = fmt.Fprintf(w, "|   $ %s%s %s%s |\n",
			cmd, strings.Repeat(" ", cmdPad),
			comment, strings.Repeat(" ", commentPad))
	}
}

func boxRow(w *os.File, content string, visible int) string {
	pad := boxWidth - 2 - visible
	if pad < 0 {
		pad = 0
	}
	return fmt.Sprintf("| %s%s |", content, strings.Repeat(" ", pad))
}

// BannerLines is the number of terminal rows a banner-style render occupies
// when it includes the leading \n + text + trailing \n\n layout.
const BannerLines = 3

// FprintBanner writes a compact one-line CoinMarketCap banner to w.
// Color is determined by the writer's fd (not the global ColorEnabled check),
// so writing to stdout vs stderr each gets the right color decision.
func FprintBanner(w io.Writer) {
	colored := false
	if os.Getenv("NO_COLOR") == "" {
		if f, ok := w.(*os.File); ok {
			colored = term.IsTerminal(int(f.Fd()))
		}
	}
	if !colored {
		_, _ = fmt.Fprint(w, "\n  CoinMarketCap CLI  —  Crypto market data\n\n")
		return
	}
	_, _ = fmt.Fprintf(w, "\n  %s◆ CoinMarketCap%s %sCLI  —  Crypto market data%s\n\n",
		brandBlue, colorReset, dimColor, colorReset)
}

// PrintBanner prints a compact one-line CoinMarketCap banner to stderr.
// Writing to stderr keeps stdout clean for piped data.
func PrintBanner() {
	FprintBanner(os.Stderr)
}
