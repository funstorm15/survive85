package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/funstorm15/survive85/internal/events"
	"github.com/funstorm15/survive85/internal/game"
	"github.com/funstorm15/survive85/internal/i18n"
	"github.com/funstorm15/survive85/internal/tui/components"
)

// morgueView is the autopsy report shown after death.
func (m *Model) morgueView() string {
	l := m.lang
	g := m.g
	d := g.Death
	p := g.Player
	w := m.contentWidth()
	if d == nil {
		return ""
	}
	inner := w - 6
	field := func(label, value string) string {
		return m.st.Dim.Render(fmt.Sprintf("%-28s", label)) + m.st.Text.Render(wrap(value, inner-28))
	}

	var report strings.Builder
	patient := fmt.Sprintf("%s, %s, %s", g.Prof.Name.Get(l), strings.ToLower(i18n.S(l, "sex."+p.Sex)), m.cfg.Book.RegionName(p.Region).Get(l))
	report.WriteString(field(i18n.S(l, "morgue.patient"), patient) + "\n")
	report.WriteString(field(i18n.S(l, "morgue.age"), i18n.F(l, "morgue.age.value", d.Age)) + "\n")
	report.WriteString(m.st.Dim.Render(fmt.Sprintf("%-28s", i18n.S(l, "morgue.cause"))) + m.st.Danger.Render(wrap(d.Reason.Get(l), inner-28)) + "\n")
	if info := m.cfg.DB.Causes[d.Cause]; info != nil {
		report.WriteString(field(i18n.S(l, "morgue.group"), info.Name.Get(l)) + "\n")
	}
	how := i18n.F(l, "morgue.annual", i18n.Percent(d.Chance))
	switch {
	case d.Instant:
		how = i18n.S(l, "morgue.instant")
	case d.Collapse:
		how = i18n.S(l, "morgue.collapse")
	}
	report.WriteString(field(i18n.S(l, "morgue.how"), how) + "\n")

	// The decision that led here.
	report.WriteString("\n" + m.st.Heading.Render(i18n.S(l, "morgue.decision")) + "\n")
	evID := d.Evidence
	if ev, ok := m.cfg.DB.Event(d.EventID); ok {
		line := "«" + ev.Title.Get(l) + "»"
		for i := range ev.Choices {
			if ev.Choices[i].ID == d.ChoiceID {
				line += " → " + ev.Choices[i].Text.Get(l)
			}
		}
		at := ""
		if d.Factor != nil {
			at = " " + i18n.F(l, "morgue.at", d.Factor.Age)
		}
		report.WriteString(m.st.Warn.Render(wrap(line+at, inner)) + "\n")
		if d.Factor != nil {
			report.WriteString(m.st.Dim.Render(i18n.F(l, "morgue.factor", strconv.FormatFloat(d.Factor.Factor, 'f', 2, 64))) + "\n")
		}
	} else if d.Factor != nil {
		report.WriteString(m.st.Warn.Render(wrap(d.Factor.Label.Get(l)+" "+i18n.F(l, "morgue.at", d.Factor.Age), inner)) + "\n")
		report.WriteString(m.st.Dim.Render(i18n.F(l, "morgue.factor", strconv.FormatFloat(d.Factor.Factor, 'f', 2, 64))) + "\n")
	} else {
		report.WriteString(m.st.Dim.Render(wrap(i18n.S(l, "morgue.natural"), inner)) + "\n")
	}

	// Scientific basis.
	if ev := m.cfg.DB.Evidence[evID]; ev != nil {
		report.WriteString("\n" + m.st.Heading.Render(i18n.S(l, "morgue.science")) + "\n")
		report.WriteString(m.evidenceCard(ev, inner) + "\n")
	} else if len(m.cfg.Book.Sources) > 0 {
		report.WriteString("\n" + m.st.Heading.Render(i18n.S(l, "morgue.science")) + "\n")
		src := m.cfg.Book.Sources[0]
		report.WriteString(m.st.Text.Render(wrap(src.Name, inner)) + "\n" + m.st.Special.Render(src.URL) + "\n")
	}

	// Years lost and peers.
	report.WriteString("\n" + field(i18n.S(l, "morgue.lost"), strconv.FormatFloat(d.YLL, 'f', 1, 64)) + "\n")
	report.WriteString(field(i18n.S(l, "morgue.short"), i18n.F(l, "morgue.years", game.VictoryAge-d.Age)) + "\n")
	report.WriteString(m.st.Text.Render(i18n.F(l, "morgue.peers", i18n.Percent(d.CohortDiedBefore))) + "\n")

	report.WriteString("\n" + m.modifiersBlock(inner))
	report.WriteString("\n" + m.statsBlock())
	report.WriteString("\n" + m.st.Danger.Render(i18n.F(l, "morgue.score", g.Score())) + "\n")
	report.WriteString(m.st.Dim.Render(i18n.F(l, "morgue.seed", g.Seed, g.Seed)))

	stone := components.Tombstone(
		i18n.S(l, "morgue.rip"),
		fmt.Sprintf("%d — %d", game.StartAge, d.Age),
		g.Prof.Name.Get(l),
	)
	return m.st.Dim.Render(stone) + "\n\n" + m.st.Frame.Width(w-2).Render(report.String())
}

// modifiersBlock reveals the hidden risk modifiers.
func (m *Model) modifiersBlock(w int) string {
	l := m.lang
	mods := m.g.Modifiers()
	var b strings.Builder
	b.WriteString(m.st.Heading.Render(i18n.S(l, "morgue.mods")) + "\n")
	for _, r := range events.Risks {
		v := mods[r]
		st := m.st.Text
		switch {
		case v >= 1.5:
			st = m.st.Danger
		case v > 1.0001:
			st = m.st.Warn
		case v < 0.9999:
			st = m.st.Good
		}
		name := i18n.S(l, "risk."+string(r))
		b.WriteString(m.st.Dim.Render(fmt.Sprintf("  %-22s ", name)) + st.Render(fmt.Sprintf("×%.2f", v)) + "\n")
	}
	_ = w
	return b.String()
}

func (m *Model) statsBlock() string {
	l := m.lang
	p := m.g.Player
	tags := p.TagCounts()
	var b strings.Builder
	b.WriteString(m.st.Heading.Render(i18n.S(l, "morgue.stats")) + "\n")
	b.WriteString(m.st.Text.Render(i18n.F(l, "morgue.decisions", tags[events.TagWise], tags[events.TagNeutral], tags[events.TagRisky])) + "\n")
	b.WriteString(m.st.Text.Render(i18n.F(l, "morgue.hours", i18n.Number(l, int64(p.HoursInvested)), i18n.Money(l, p.MoneySpent))) + "\n")
	b.WriteString(m.st.Text.Render(i18n.F(l, "morgue.peak", i18n.Money(l, p.PeakWealth))) + "\n")
	if p.YearsInPrison > 0 || p.Arrests > 0 {
		b.WriteString(m.st.Warn.Render(i18n.F(l, "morgue.prison", p.YearsInPrison, p.Arrests)) + "\n")
	}
	b.WriteString(m.st.Text.Render(i18n.F(l, "morgue.sources", len(p.EvidenceUsed()))) + "\n")
	return b.String()
}
