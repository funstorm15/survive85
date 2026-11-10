package i18n

import (
	"strconv"
	"strings"
)

// RublesPerDollar converts in-game money (kept in rubles) to the dollar
// amounts shown in the English UI. It is a rough purchasing-power rate chosen
// so that prices and salaries feel plausible in both languages; it is a game
// convention, not an exchange-rate claim.
const RublesPerDollar = 25

// Money formats an amount of in-game rubles for display in l:
// "154 000 ₽" in Russian and "$6,160" in English.
func Money(l Lang, rubles int64) string {
	if l == RU {
		if rubles < 0 {
			return "−" + Money(l, -rubles)
		}
		return group(rubles, ' ') + " ₽"
	}
	d := rubles / RublesPerDollar
	if d < 0 {
		return "−$" + group(-d, ',')
	}
	return "$" + group(d, ',')
}

// SignedMoney is Money with an explicit sign for deltas ("+12 000 ₽").
func SignedMoney(l Lang, rubles int64) string {
	if rubles > 0 {
		return "+" + Money(l, rubles)
	}
	return Money(l, rubles)
}

// Number groups the thousands of n using the convention of l.
func Number(l Lang, n int64) string {
	if l == RU {
		return group(n, ' ')
	}
	return group(n, ',')
}

func group(n int64, sep rune) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	lead := len(s) % 3
	if lead > 0 {
		b.WriteString(s[:lead])
	}
	for i := lead; i < len(s); i += 3 {
		if b.Len() > 0 && (!neg || b.Len() != 1) {
			b.WriteRune(sep)
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// Percent formats a probability (0..1) as a percentage with precision that
// keeps small risks readable: 0.00042 -> "0.042%", 0.153 -> "15.3%".
func Percent(p float64) string {
	pct := p * 100
	switch {
	case pct == 0:
		return "0%"
	case pct < 0.1:
		return strconv.FormatFloat(pct, 'f', 3, 64) + "%"
	case pct < 10:
		return strconv.FormatFloat(pct, 'f', 2, 64) + "%"
	default:
		return strconv.FormatFloat(pct, 'f', 1, 64) + "%"
	}
}
