package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/funstorm15/survive85/internal/game"
	"github.com/funstorm15/survive85/internal/i18n"
	"github.com/funstorm15/survive85/internal/tui/components"
)

// victoryView celebrates reaching 85.
func (m *Model) victoryView() string {
	l := m.lang
	g := m.g
	w := m.contentWidth()
	var b strings.Builder
	b.WriteString(m.st.Warn.Render(components.Cake) + "\n\n")
	b.WriteString(m.st.Good.Bold(true).Render(i18n.S(l, "victory.epitaph")) + "\n\n")
	died := game.CohortDiedBefore(g.Book, g.Table, game.VictoryAge)
	base := g.Book.Survival(g.Table, game.StartAge, game.VictoryAge)
	b.WriteString(m.st.Text.Render(wrap(i18n.F(l, "victory.text", i18n.Percent(died), i18n.Percent(base)), w)) + "\n\n")

	if top := g.Player.TopFactors(5); len(top) > 0 {
		b.WriteString(m.st.Heading.Render(i18n.S(l, "victory.top")) + "\n")
		for _, f := range top {
			b.WriteString("  " + m.riskArrow(f.Risk, f.Factor, 0, 0) + m.st.Dim.Render(" — "+truncate(f.Label.Get(l), w-30)+" ("+i18n.F(l, "morgue.at", f.Age)+")") + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(m.modifiersBlock(w) + "\n")
	b.WriteString(m.statsBlock() + "\n")
	b.WriteString(m.st.Good.Bold(true).Render(i18n.F(l, "morgue.score", g.Score())) + "\n")
	b.WriteString(m.st.Dim.Render(i18n.F(l, "morgue.seed", g.Seed, g.Seed)))
	return lipgloss.NewStyle().Width(w).Render(b.String())
}
