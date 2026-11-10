// Package components holds small rendering helpers shared by the screens:
// gauges, the ASCII logo and tombstone, and text wrapping.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

// Gauge renders a bar like "████████░░" for value in 0..max.
func Gauge(value, max, cells int, fill lipgloss.Style, empty lipgloss.Style) string {
	if max <= 0 || cells <= 0 {
		return ""
	}
	if value < 0 {
		value = 0
	}
	if value > max {
		value = max
	}
	n := (value*cells + max/2) / max
	return fill.Render(strings.Repeat("█", n)) + empty.Render(strings.Repeat("░", cells-n))
}

// Pips renders willpower as filled and empty diamonds: "◆◆◆◇◇".
func Pips(value, max int, fill, empty lipgloss.Style) string {
	if value < 0 {
		value = 0
	}
	if value > max {
		value = max
	}
	return fill.Render(strings.Repeat("◆", value)) + empty.Render(strings.Repeat("◇", max-value))
}

// Wrap word-wraps text to width, keeping paragraph breaks.
func Wrap(text string, width int) string {
	if width < 10 {
		width = 10
	}
	return wordwrap.String(strings.TrimRight(text, "\n"), width)
}

// Reflow joins the hard line breaks of YAML block text into flowing
// paragraphs, keeping blank-line paragraph breaks, so it can be re-wrapped
// to the terminal width.
func Reflow(s string) string {
	paras := strings.Split(strings.TrimSpace(s), "\n\n")
	for i, p := range paras {
		paras[i] = strings.Join(strings.Fields(p), " ")
	}
	return strings.Join(paras, "\n\n")
}

// Logo is the title art. Every line has the same width.
var Logo = []string{
	`███████╗██╗   ██╗██████╗ ██╗   ██╗██╗██╗   ██╗███████╗    █████╗ ███████╗`,
	`██╔════╝██║   ██║██╔══██╗██║   ██║██║██║   ██║██╔════╝██╗██╔══██╗██╔════╝`,
	`███████╗██║   ██║██████╔╝██║   ██║██║██║   ██║█████╗  ╚═╝╚█████╔╝███████╗`,
	`╚════██║██║   ██║██╔══██╗╚██╗ ██╔╝██║╚██╗ ██╔╝██╔══╝  ██╗██╔══██╗╚════██║`,
	`███████║╚██████╔╝██║  ██║ ╚████╔╝ ██║ ╚████╔╝ ███████╗╚═╝╚█████╔╝███████║`,
	`╚══════╝ ╚═════╝ ╚═╝  ╚═╝  ╚═══╝  ╚═╝  ╚═══╝  ╚══════╝    ╚════╝ ╚══════╝`,
}

// SmallLogo fits narrow terminals.
var SmallLogo = []string{
	`╔═╗ ╦ ╦ ╦═╗ ╦  ╦ ╦ ╦  ╦ ╔═╗   ╔═╗ ╔═╗`,
	`╚═╗ ║ ║ ╠╦╝ ╚╗╔╝ ║ ╚╗╔╝ ║╣  • ╠═╣ ╚═╗`,
	`╚═╝ ╚═╝ ╩╚═  ╚╝  ╩  ╚╝  ╚═╝ • ╚═╝ ╚═╝`,
}

// RenderLogo picks the logo that fits width and colours it line by line
// with a vertical gradient between two styles.
func RenderLogo(width, height int, top, bottom lipgloss.Style) string {
	art := Logo
	if lipgloss.Width(Logo[0]) > width || height < 30 {
		art = SmallLogo
	}
	var b strings.Builder
	for i, line := range art {
		st := top
		if i >= len(art)/2 {
			st = bottom
		}
		b.WriteString(st.Render(line))
		if i < len(art)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// Tombstone draws a gravestone with up to three centred lines of text.
func Tombstone(lines ...string) string {
	const inner, margin = 25, 3 // columns between the side walls; left margin
	sp := func(n int) string { return strings.Repeat(" ", n) }
	center := func(s string) string {
		s = truncate(s, inner-2)
		w := lipgloss.Width(s)
		left := (inner - w) / 2
		return sp(left) + s + sp(inner-w-left)
	}
	for len(lines) < 3 {
		lines = append(lines, "")
	}
	wall := func(s string) string { return sp(margin) + "|" + center(s) + "|" }
	art := []string{
		sp(margin+5) + "_." + strings.Repeat("-", inner-12) + "._",
		sp(margin+2) + ".-'" + sp(inner-8) + "'-.",
		sp(margin+1) + "/" + sp(inner-2) + "\\",
		wall(lines[0]),
		wall(lines[1]),
		wall(lines[2]),
		wall(""),
		wall("~ * ~ * ~"),
		sp(margin-1) + "_|" + strings.Repeat("_", inner) + "|_",
		sp(margin-2) + "/" + strings.Repeat("_", inner+4) + "\\",
	}
	return strings.Join(art, "\n")
}

func truncate(s string, w int) string {
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r)) > w {
		r = r[:len(r)-1]
	}
	return string(r)
}

// Cake is the victory art.
const Cake = `          i  i  i  i  i
         |:||:||:||:||:|
       __|:||:||:||:||:|__
      |  ~~~~~ 85 ~~~~~   |
      |___________________|
     (_____________________)`

// Rule is a horizontal line of width w.
func Rule(w int, st lipgloss.Style) string {
	if w < 1 {
		w = 1
	}
	return st.Render(strings.Repeat("─", w))
}
