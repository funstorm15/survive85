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

// outcomeView shows what the decision did, the evidence behind it and the
// year-end report.
func (m *Model) outcomeView() string {
	l := m.lang
	o := m.g.Last
	w := m.contentWidth()
	if o == nil {
		return ""
	}
	var b strings.Builder
	if o.Event == nil {
		b.WriteString(m.st.Heading.Render("☼ "+i18n.S(l, "event.quiet")) + "\n")
	} else {
		b.WriteString(m.st.Heading.Render(categoryIcon[o.Event.Category]+" "+o.Event.Title.Get(l)) +
			m.st.Dim.Render(" · "+i18n.F(l, "event.year", o.Age)) + "\n\n")
	}
	if ch := o.Choice; ch != nil {
		b.WriteString(m.st.Dim.Render(i18n.S(l, "out.chose")+" ") + m.st.Selected.Render(wrap(ch.Text.Get(l), w-12)) + "\n\n")
		b.WriteString(m.st.Text.Render(wrap(components.Reflow(ch.Outcome.Get(l)), w)) + "\n\n")
	}

	if d := o.Deltas; d != (game.Deltas{}) {
		b.WriteString(m.st.Dim.Render(i18n.S(l, "out.changes")+" ") + m.deltas(d) + "\n")
	}
	if len(o.Factors) > 0 {
		var parts []string
		for _, f := range o.Factors {
			parts = append(parts, m.riskArrow(f.Risk, f.Factor, f.Until, f.Age))
		}
		b.WriteString(m.st.Dim.Render(i18n.S(l, "out.risks")+" ") + strings.Join(parts, m.st.Dim.Render(" · ")) + "\n")
	}
	if names := m.conditionNames(o.Added); names != "" {
		b.WriteString(m.st.Dim.Render(i18n.S(l, "out.added")+" ") + names + "\n")
	}
	if names := m.conditionNames(o.Removed); names != "" {
		b.WriteString(m.st.Dim.Render(i18n.S(l, "out.removed")+" ") + names + "\n")
	}

	if c := o.Death; c != nil {
		if c.Triggered {
			b.WriteString("\n" + m.st.Danger.Render("✝ "+m.g.Death.Reason.Get(l)) + "\n")
		} else {
			b.WriteString("\n" + m.st.Good.Render("✓ "+i18n.F(l, "out.survived", i18n.Percent(c.Chance))) + "\n")
		}
	}
	if c := o.Disaster; c != nil && o.Choice != nil && o.Choice.CatastropheCheck != nil {
		if c.Triggered {
			b.WriteString("\n" + m.st.Danger.Render("✗ "+i18n.F(l, "out.disaster", i18n.Percent(c.Chance))) + "\n" +
				m.st.Warn.Render(wrap(components.Reflow(o.Choice.CatastropheCheck.Text.Get(l)), w)) + "\n")
		} else {
			b.WriteString("\n" + m.st.Good.Render("✓ "+i18n.F(l, "out.dodged", i18n.Percent(c.Chance))) + "\n")
		}
	}

	// The evidence card: the choice's source, else the event's.
	evID := ""
	if o.Choice != nil {
		evID = o.Choice.Evidence
	}
	if evID == "" && o.Event != nil {
		evID = o.Event.Evidence
	}
	if ev := m.cfg.DB.Evidence[evID]; ev != nil {
		b.WriteString("\n" + m.evidenceCard(ev, w) + "\n")
	}

	if y := o.Year; y != nil {
		b.WriteString("\n" + m.yearEnd(y, w))
	} else if m.g.Phase == game.PhaseDead {
		b.WriteString("\n" + m.st.Danger.Render("✝ "+i18n.S(l, "morgue.rip")) + "\n")
	}
	return b.String()
}

func (m *Model) deltas(d game.Deltas) string {
	l := m.lang
	var parts []string
	add := func(key string, v int64, text string) {
		if v != 0 {
			parts = append(parts, m.st.Dim.Render(i18n.S(l, key)+" ")+m.st.Signed(v, text))
		}
	}
	add("stat.health", int64(d.Health), signed(d.Health))
	add("stat.sanity", int64(d.Sanity), signed(d.Sanity))
	add("stat.wealth", d.Wealth, i18n.SignedMoney(l, d.Wealth))
	add("stat.will", int64(d.Will), signed(d.Will))
	return strings.Join(parts, m.st.Dim.Render(" · "))
}

func signed(v int) string {
	switch {
	case v > 0:
		return "+" + strconv.Itoa(v)
	case v < 0:
		return "−" + strconv.Itoa(-v)
	}
	return "0"
}

// shortCitation drops the long verbatim excerpts the book appends to some
// citations after "：「" so the card shows the reference itself.
func shortCitation(c string) string {
	for _, sep := range []string{"：「", ":「", "「"} {
		if i := strings.Index(c, sep); i > 0 {
			return strings.TrimSpace(c[:i])
		}
	}
	return c
}

// riskArrow tells which hidden risk moved and roughly how much, without
// revealing the number.
func (m *Model) riskArrow(r events.Risk, f float64, until, age int) string {
	l := m.lang
	name := i18n.S(l, "risk."+string(r))
	var arrow string
	switch {
	case f >= 1.5:
		arrow = m.st.Danger.Render("↑↑")
	case f > 1:
		arrow = m.st.Danger.UnsetBold().Render("↑")
	case f <= 0.75:
		arrow = m.st.Good.Bold(true).Render("↓↓")
	default:
		arrow = m.st.Good.Render("↓")
	}
	s := m.st.Text.Render(name) + " " + arrow
	if until > 0 {
		s += m.st.Dim.Render(" (" + i18n.F(l, "out.years", until-age+1) + ")")
	}
	return s
}

func (m *Model) conditionNames(ids []string) string {
	var parts []string
	for _, id := range ids {
		if def := m.cfg.DB.Conditions[id]; def != nil && !def.Hidden {
			st := m.st.Text
			switch def.Kind {
			case "good":
				st = m.st.Good
			case "bad":
				st = m.st.Danger.UnsetBold()
			}
			parts = append(parts, st.Render(strings.TrimSpace(def.Icon+" "+def.Name.Get(m.lang))))
		}
	}
	return strings.Join(parts, m.st.Dim.Render(" · "))
}

// evidenceCard renders one book item: chapter, item, grade, plain text,
// numbers, the first source and links.
func (m *Model) evidenceCard(ev *events.Evidence, w int) string {
	l := m.lang
	inner := w - 4
	var b strings.Builder
	b.WriteString(m.st.Special.Bold(true).Render(i18n.S(l, "out.evidence")) + "\n")
	b.WriteString(m.st.Dim.Render(wrap(i18n.F(l, "out.chapter", ev.Chapter, ev.ChapterTitle.Get(l), ev.Item), inner)) + "\n")
	grade := m.st.Grade(ev.Grade).Render("[" + ev.Grade + "] " + i18n.F(l, "out.grade", ev.Grade))
	if ev.Contested {
		grade += m.st.Warn.Render(" · " + i18n.S(l, "out.contested"))
	}
	b.WriteString(grade + "\n")
	b.WriteString(m.st.Selected.Render(wrap(ev.Title.Get(l), inner)) + "\n\n")
	b.WriteString(m.st.Text.Render(wrap(ev.Plain.Get(l), inner)) + "\n")
	if nums := formatNumbers(ev.Numbers, l); nums != "" {
		b.WriteString("\n" + m.st.Warn.Render(wrap(nums, inner)) + "\n")
	}
	if len(ev.Sources) > 0 {
		b.WriteString("\n" + m.st.Text.Render(wrap(i18n.S(l, "out.source")+" "+shortCitation(ev.Sources[0].Citation), inner)) + "\n")
	}
	if link := ev.PrimaryLink(); link != "" {
		b.WriteString(m.st.Special.Render(link) + "\n")
	}
	b.WriteString(m.st.Dim.Render(i18n.S(l, "out.book")+" ") + m.st.Special.Render(ev.TranslationURL(l)))
	return m.st.Card.Width(w - 2).Render(b.String()) // Width excludes the border
}

// formatNumbers renders study statistics: "OR 0.58 (95% CI 0.50–0.68)".
// The value and its interval share one precision, as in the book.
func formatNumbers(ns []events.Number, l i18n.Lang) string {
	decimals := func(f float64) int {
		s := strconv.FormatFloat(f, 'f', -1, 64)
		if i := strings.IndexByte(s, '.'); i >= 0 {
			return len(s) - i - 1
		}
		return 0
	}
	var parts []string
	for _, n := range ns {
		prec := decimals(n.Value)
		if n.CILow != nil && n.CIHigh != nil {
			prec = max(prec, decimals(*n.CILow), decimals(*n.CIHigh))
		}
		f := func(v float64) string { return strconv.FormatFloat(v, 'f', prec, 64) }
		s := n.Kind + " " + f(n.Value)
		if n.CILow != nil && n.CIHigh != nil {
			s += fmt.Sprintf(" (95%% CI %s–%s)", f(*n.CILow), f(*n.CIHigh))
		}
		if o := n.Outcome.Get(l); o != "" {
			s += " — " + o
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, " · ")
}

// yearEnd renders the annual report.
func (m *Model) yearEnd(y *game.YearEnd, w int) string {
	l := m.lang
	var b strings.Builder
	b.WriteString(m.st.Heading.Render(i18n.S(l, "out.yearend")) + "\n")
	if y.Income != 0 {
		b.WriteString(m.st.Dim.Render(i18n.S(l, "out.income")+" ") + m.st.Signed(y.Income, i18n.SignedMoney(l, y.Income)) + "\n")
	} else {
		b.WriteString(m.st.Dim.Render(i18n.S(l, "out.noincome")) + "\n")
	}
	if y.Interest > 0 {
		b.WriteString(m.st.Dim.Render(i18n.S(l, "out.interest")+" ") + m.st.Danger.Render(i18n.SignedMoney(l, -y.Interest)) + "\n")
	}
	if up := m.deltas(y.Upkeep); up != "" {
		b.WriteString(m.st.Dim.Render(i18n.S(l, "out.upkeep")+" ") + up + "\n")
	}
	mt := y.Mortality
	rel := strconv.FormatFloat(mt.Relative, 'f', 2, 64)
	riskStyle := m.st.Text
	switch {
	case mt.Relative >= 1.5:
		riskStyle = m.st.Danger
	case mt.Relative <= 0.9:
		riskStyle = m.st.Good
	}
	b.WriteString(riskStyle.Render(wrap(i18n.F(l, "out.risk", i18n.Percent(mt.Total), rel), w)) + "\n")
	if y.Arrest.Triggered {
		b.WriteString(m.st.Danger.Render("§ "+i18n.F(l, "out.arrest", i18n.Percent(y.Arrest.Chance))) + "\n")
	}
	if names := m.conditionNames(y.Expired); names != "" {
		b.WriteString(m.st.Dim.Render(i18n.S(l, "out.expired")+" ") + names + "\n")
	}
	if y.Survived {
		b.WriteString(m.st.Good.Render("✓ "+i18n.F(l, "out.passed", y.Age+1)) + "\n")
	} else if m.g.Death != nil {
		b.WriteString(m.st.Danger.Render("✝ "+m.g.Death.Reason.Get(l)) + "\n")
	}
	return b.String()
}
