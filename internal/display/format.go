package display

import (
	"fmt"
	"math"
	"strings"
)

// CurrencySymbol returns the display symbol for a supported quote currency code.
var currencySymbols = map[string]string{
	"usd": "$", "aed": "د.إ", "ars": "ARS$", "aud": "A$",
	"bdt": "৳", "bhd": "BD", "bmd": "BD$", "brl": "R$",
	"cad": "C$", "chf": "CHF", "clp": "CLP$", "cny": "¥",
	"czk": "Kč", "dkk": "kr", "eur": "€", "gbp": "£",
	"gel": "₾", "hkd": "HK$", "huf": "Ft", "idr": "Rp",
	"ils": "₪", "inr": "₹", "jpy": "¥", "krw": "₩",
	"kwd": "KD", "lkr": "Rs", "mmk": "K", "mxn": "MX$",
	"myr": "RM", "ngn": "₦", "nok": "kr", "nzd": "NZ$",
	"php": "₱", "pkr": "₨", "pln": "zł", "rub": "₽",
	"sar": "﷼", "sek": "kr", "sgd": "S$", "thb": "฿",
	"try": "₺", "twd": "NT$", "uah": "₴", "vef": "Bs.F",
	"vnd": "₫", "zar": "R",
	// Crypto
	"btc": "₿", "eth": "Ξ", "ltc": "Ł", "xrp": "XRP ",
	"dot": "DOT ", "sats": "sats ",
}

func CurrencySymbol(vs string) string {
	if s, ok := currencySymbols[vs]; ok {
		return s
	}
	return strings.ToUpper(vs) + " "
}

func FormatPrice(price float64, vs ...string) string {
	sym := "$"
	if len(vs) > 0 {
		sym = CurrencySymbol(vs[0])
	}
	if price == 0 {
		return sym + "0.00"
	}
	abs := math.Abs(price)
	sign := ""
	if price < 0 {
		sign = "-"
	}

	switch {
	case abs >= 1:
		return sign + sym + formatWithCommas(abs, 2)
	case abs >= 0.01:
		return fmt.Sprintf("%s%s%.4f", sign, sym, abs)
	default:
		return fmt.Sprintf("%s%s%.8f", sign, sym, abs)
	}
}

func FormatPercent(pct float64) string {
	return fmt.Sprintf("%.2f%%", pct)
}

func FormatLargeNumber(n float64, vs ...string) string {
	sym := "$"
	if len(vs) > 0 {
		sym = CurrencySymbol(vs[0])
	}
	abs := math.Abs(n)
	sign := ""
	if n < 0 {
		sign = "-"
	}

	switch {
	case abs >= 1e12:
		return fmt.Sprintf("%s%s%.2fT", sign, sym, abs/1e12)
	case abs >= 1e9:
		return fmt.Sprintf("%s%s%.2fB", sign, sym, abs/1e9)
	case abs >= 1e6:
		return fmt.Sprintf("%s%s%.2fM", sign, sym, abs/1e6)
	case abs >= 1e3:
		return fmt.Sprintf("%s%s%.2fK", sign, sym, abs/1e3)
	default:
		return sign + sym + formatWithCommas(abs, 2)
	}
}

func FormatSupply(n float64) string {
	abs := math.Abs(n)
	switch {
	case abs >= 1e12:
		return fmt.Sprintf("%.2fT", abs/1e12)
	case abs >= 1e9:
		return fmt.Sprintf("%.2fB", abs/1e9)
	case abs >= 1e6:
		return fmt.Sprintf("%.2fM", abs/1e6)
	default:
		return formatWithCommas(abs, 0)
	}
}

func formatWithCommas(n float64, decimals int) string {
	s := fmt.Sprintf("%.*f", decimals, n)
	parts := strings.Split(s, ".")
	intPart := parts[0]

	var result []byte
	for i := 0; i < len(intPart); i++ {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, intPart[i])
	}

	if len(parts) > 1 {
		return string(result) + "." + parts[1]
	}
	return string(result)
}
