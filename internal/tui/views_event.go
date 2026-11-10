package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/funstorm15/survive85/internal/events"
	"github.com/funstorm15/survive85/internal/game"
	"github.com/funstorm15/survive85/internal/i18n"
	"github.com/funstorm15/survive85/internal/tui/components"
)

var categoryIcon = map[events.Category]string{
	events.CatHealth: "✚", events.CatChronic: "♨", events.CatEmergency: "⚠", events.CatPharmacy: "℞",
	events.CatSafety: "⛨", events.CatMoney: "₽", events.CatLegal: "§", events.CatWork: "⚒",
	events.CatFamily: "♥", events.CatMind: "☯", events.CatTravel: "✈", events.CatLife: "☼",
}

// statusBar renders the top line:
// [ВОЗРАСТ: 24/85] | HP: [████████░░] 80% | SANITY: [██████░░░░] 60% | КАПИТАЛ: 154 000 ₽ | ВОЛЯ: ◆◆◆◇◇
func (m *Model) statusBar() string {
	l := m.lang
	p := m.g.Player
	sep := m.st.Dim.Render(" | ")
	empty := m.st.Dim
	build := func(cells int) (string, string) {
		gauge := func(label string, v int) string {
			st := m.st.Level(v)
			return m.st.Bar.Render(label+": ") + "[" + components.Gauge(v, game.MaxStat, cells, st, empty) + "] " + st.Render(fmt.Sprintf("%d%%", v))
		}
		wst := m.st.Text
		if p.Wealth < 0 {
			wst = m.st.Danger
		}
		line := strings.Join([]string{
			m.st.Bar.Render(fmt.Sprintf("[%s: %d/%d]", i18n.S(l, "bar.age"), p.Age, game.VictoryAge)),
			gauge(i18n.S(l, "bar.hp"), p.Health),
			gauge(i18n.S(l, "bar.sanity"), p.Sanity),
			m.st.Bar.Render(i18n.S(l, "bar.wealth")+": ") + wst.Render(i18n.Money(l, p.Wealth)),
		}, sep)
		will := m.st.Bar.Render(i18n.S(l, "bar.will")+": ") + components.Pips(p.Will, game.MaxWill, m.st.Special, empty)
		return line, will
	}
	for _, cells := range []int{10, 6, 4} {
		line, will := build(cells)
		if lipgloss.Width(line+sep+will) <= m.contentWidth() {
			return line + sep + will
		}
	}
	line, will := build(4)
	return line + "\n" + will
}

// eventView renders the year: status bar, the event and its choices.
func (m *Model) eventView() string {
	l := m.lang
	g := m.g
	w := m.contentWidth()
	var b strings.Builder
	b.WriteString(m.statusBar() + "\n" + rule(w, m.st) + "\n")

	ev := g.Current
	if ev == nil {
		b.WriteString(m.st.Heading.Render("☼ "+i18n.S(l, "event.quiet")) + "\n\n")
		b.WriteString(m.st.Text.Render(wrap(i18n.S(l, "event.quiet.text"), w)) + "\n\n")
		b.WriteString(m.st.Cursor.Render(" ▸ "+i18n.S(l, "event.continue")) + "\n")
	} else {
		head := m.st.Heading.Render(categoryIcon[ev.Category] + " " + ev.Title.Get(l))
		tag := m.st.Dim.Render(i18n.S(l, "cat."+string(ev.Category)) + " · " + i18n.F(l, "event.year", g.Player.Age))
		gap := max(1, w-lipgloss.Width(head)-lipgloss.Width(tag))
		b.WriteString(head + strings.Repeat(" ", gap) + tag + "\n\n")
		b.WriteString(m.st.Text.Render(wrap(components.Reflow(ev.Narrative.Get(l)), w)) + "\n\n")

		st := g.State()
		for n, i := range g.VisibleChoices() {
			ch := &ev.Choices[i]
			avail := ch.Availability(st)
			prefix := fmt.Sprintf("%d. ", n+1)
			text := wrap(ch.Text.Get(l), w-6)
			text = strings.ReplaceAll(text, "\n", "\n      ")
			var line string
			switch {
			case avail != events.Available:
				line = "   " + m.st.Disabled.Render(prefix+text)
			case n == m.cursor:
				line = m.st.Cursor.Render(" ▸ "+prefix) + m.st.Selected.Render(text)
			default:
				line = "   " + m.st.Text.Render(prefix+text)
			}
			b.WriteString(line + "\n")
			if cost := m.costLine(ch, avail); cost != "" {
				b.WriteString("      " + cost + "\n")
			}
		}
	}

	b.WriteString(rule(w, m.st) + "\n")
	b.WriteString(m.conditionsLine(w) + "\n")
	b.WriteString(m.st.Footer.Render(wrap(i18n.S(l, "event.keys"), w)))
	return b.String()
}

// costLine summarises what a choice costs.
func (m *Model) costLine(ch *events.Choice, avail events.Unavailable) string {
	l := m.lang
	var parts []string
	if ch.CostMoney > 0 {
		parts = append(parts, m.st.Warn.Render("−"+i18n.Money(l, ch.CostMoney)))
	}
	if ch.CostWill > 0 {
		parts = append(parts, m.st.Special.Render(i18n.S(l, "event.cost.will")+" "+strings.Repeat("◆", ch.CostWill)))
	}
	if ch.CostTime > 0 {
		parts = append(parts, m.st.Dim.Render(i18n.F(l, "event.cost.time", strconv.FormatFloat(ch.CostTime, 'f', -1, 64))))
	}
	switch avail {
	case events.NoMoney:
		parts = append(parts, m.st.Danger.UnsetBold().Render("["+i18n.S(l, "event.nomoney")+"]"))
	case events.NoWill:
		parts = append(parts, m.st.Danger.UnsetBold().Render("["+i18n.S(l, "event.nowill")+"]"))
	}
	return strings.Join(parts, m.st.Dim.Render(" · "))
}

func (m *Model) unavailableText(i int) string {
	ch := &m.g.Current.Choices[i]
	switch ch.Availability(m.g.State()) {
	case events.NoMoney:
		return i18n.S(m.lang, "event.nomoney")
	case events.NoWill:
		return i18n.S(m.lang, "event.nowill")
	}
	return ""
}

// conditionsLine lists the player's visible conditions.
func (m *Model) conditionsLine(w int) string {
	l := m.lang
	var parts []string
	for _, c := range m.g.Player.Conditions {
		def := m.cfg.DB.Conditions[c.ID]
		if def == nil || def.Hidden {
			continue
		}
		st := m.st.Text
		switch def.Kind {
		case "good":
			st = m.st.Good
		case "bad":
			st = m.st.Danger.UnsetBold()
		}
		parts = append(parts, st.Render(strings.TrimSpace(def.Icon+" "+def.Name.Get(l))))
	}
	label := m.st.Dim.Render(i18n.S(l, "event.conditions") + " ")
	if len(parts) == 0 {
		return label + m.st.Dim.Render(i18n.S(l, "event.none"))
	}
	return wrap(label+strings.Join(parts, m.st.Dim.Render(" · ")), w)
}
