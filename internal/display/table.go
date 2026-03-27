package display

import (
	"fmt"
	"strings"
	"unicode"
)

// FormatSymbol sanitizes and uppercases a coin symbol from API data.
func FormatSymbol(s string) string {
	return strings.ToUpper(SanitizeCell(s))
}

// FormatRank formats a market cap rank, returning "-" for zero/unranked.
func FormatRank(rank int) string {
	if rank > 0 {
		return fmt.Sprintf("%d", rank)
	}
	return "-"
}

func PrintTable(headers []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Println("No data to display.")
		return
	}

	cols := len(headers)
	// Compute visible widths for all cells in a single pass.
	headerWidths := make([]int, cols)
	for i, h := range headers {
		headerWidths[i] = VisibleWidth(h)
	}
	cellWidths := make([][]int, len(rows))
	colMax := make([]int, cols)
	copy(colMax, headerWidths)
	for r, row := range rows {
		cellWidths[r] = make([]int, len(row))
		for i, cell := range row {
			w := VisibleWidth(cell)
			cellWidths[r][i] = w
			if i < cols && w > colMax[i] {
				colMax[i] = w
			}
		}
	}

	coloredHeaders := make([]string, len(headers))
	for i, h := range headers {
		coloredHeaders[i] = ColorHeader(h)
	}
	printRowWithWidths(coloredHeaders, headerWidths, colMax)
	printSeparator(colMax)
	for r, row := range rows {
		printRowWithWidths(row, cellWidths[r], colMax)
	}
}

func printRowWithWidths(cells []string, visWidths []int, colMax []int) {
	parts := make([]string, len(colMax))
	for i := range colMax {
		cell := ""
		visible := 0
		if i < len(cells) {
			cell = cells[i]
			visible = visWidths[i]
		}
		pad := colMax[i] - visible
		if pad < 0 {
			pad = 0
		}
		parts[i] = cell + strings.Repeat(" ", pad)
	}
	fmt.Println("  " + strings.Join(parts, "  "))
}

func printSeparator(widths []int) {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("─", w)
	}
	fmt.Println("  " + strings.Join(parts, "  "))
}

// SanitizeCell strips all ANSI escape sequences and non-printable control
// characters from untrusted API data. Call this on API-sourced strings (names,
// symbols) before passing to PrintTable. Internal color formatting (e.g.
// ColorPercent) operates on numeric values and should NOT be sanitized.
func SanitizeCell(s string) string {
	// Fast path: check if sanitization is needed before allocating.
	needsSanitize := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' || (s[i] < 0x20 && s[i] != '\t') {
			needsSanitize = true
			break
		}
	}
	if !needsSanitize {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\033' {
			// Skip entire escape sequence (ESC [ ... final_byte).
			j := i + 1
			if j < len(runes) && runes[j] == '[' {
				k := j + 1
				for k < len(runes) && runes[k] >= 0x20 && runes[k] <= 0x3F {
					k++
				}
				if k < len(runes) && runes[k] >= 0x40 && runes[k] <= 0x7E {
					i = k
				}
			}
			continue
		}
		if unicode.IsControl(r) && r != '\t' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

