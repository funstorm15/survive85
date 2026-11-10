package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/funstorm15/survive85/internal/game"
	"github.com/funstorm15/survive85/internal/i18n"
	"github.com/funstorm15/survive85/internal/tui/components"
)

// titleView is the start screen: logo, menu, best score and recent deaths.
func (m *Model) titleView() string {
	l := m.lang
	w := m.contentWidth()
	var b strings.Builder
	b.WriteString(components.RenderLogo(w, m.h, m.st.Title, m.st.Danger.UnsetBold()))
	b.WriteString("\n\n")
	b.WriteString(m.st.Text.Render(wrap(i18n.S(l, "title.tagline"), w)))
	b.WriteString("\n\n")

	items := []string{"title.new", "title.graveyard", "title.about", "title.lang", "title.quit"}
	for i, key := range items {
		if i == m.cursor {
			b.WriteString(m.st.Cursor.Render(" ▸ " + i18n.S(l, key)))
		} else {
			b.WriteString(m.st.Dim.Render("   " + i18n.S(l, key)))
		}
		b.WriteByte('\n')
	}

	if best, ok := BestRecord(m.history); ok {
		b.WriteString("\n" + m.st.Good.Render(i18n.F(l, "title.best", best.Score, best.Age)) + "\n")
	}
	if dead := RecentDeaths(m.history, 3); len(dead) > 0 && m.h >= 30 {
		b.WriteString(m.st.Dim.Render(i18n.S(l, "title.lastdead")) + "\n")
		for _, r := range dead {
			b.WriteString(m.st.Danger.UnsetBold().Render("  ✝ ") + m.st.Dim.Render(truncate(fmt.Sprintf("%d — %s", r.Age, r.Reason.Get(l)), w-4)) + "\n")
		}
	}
	b.WriteString("\n" + m.st.Footer.Render(wrap(i18n.S(l, "title.based"), w)))
	b.WriteString("\n" + m.st.Footer.Render("v"+m.cfg.Version+" · ? "+strings.ToLower(i18n.S(l, "help.title"))+" · L · T · q"))
	return b.String()
}

// setupView chooses the life table and profession for a new run.
func (m *Model) setupView() string {
	l := m.lang
	w := m.contentWidth()
	book := m.cfg.Book
	regions := book.Regions()
	region := regions[m.region]
	prof := m.cfg.DB.Professions[m.prof]

	row := func(field int, label, value string) string {
		cursor, st := "   ", m.st.Text
		if m.field == field {
			cursor, st = m.st.Cursor.Render(" ▸ "), m.st.Selected
		}
		return cursor + m.st.Dim.Render(fmt.Sprintf("%-28s", label)) + st.Render("‹ "+value+" ›")
	}

	var b strings.Builder
	b.WriteString(m.st.Title.Render(i18n.S(l, "setup.title")) + "\n\n")
	b.WriteString(row(fieldRegion, i18n.S(l, "setup.region"), book.RegionName(region).Get(l)) + "\n")
	b.WriteString(row(fieldSex, i18n.S(l, "setup.sex"), i18n.S(l, "sex."+sexes[m.sex])) + "\n")
	b.WriteString(row(fieldProfession, i18n.S(l, "setup.profession"), prof.Icon+" "+prof.Name.Get(l)) + "\n\n")

	desc := m.st.Panel.Width(min(w, 90)).Render(m.st.Text.Render(wrap(components.Reflow(prof.Description.Get(l)), min(w, 90)-4)))
	b.WriteString(desc + "\n\n")

	if t, ok := book.TableFor(region, sexes[m.sex]); ok {
		p := book.Survival(t, game.StartAge, game.VictoryAge)
		b.WriteString(m.st.Warn.Render(i18n.F(l, "setup.baseline", i18n.Percent(p))) + "\n")
	}
	seed := fmt.Sprint(m.seed)
	if m.cfg.FixedSeed {
		seed = i18n.F(l, "setup.seed.fixed", m.seed)
	}
	b.WriteString(m.st.Dim.Render(i18n.S(l, "setup.seed")+": "+seed) + "\n\n")

	start := "   " + i18n.S(l, "setup.start")
	if m.field == fieldStart {
		start = m.st.Cursor.Render(" ▸ " + i18n.S(l, "setup.start"))
	} else {
		start = m.st.Dim.Render(start)
	}
	b.WriteString(start + "\n\n")
	b.WriteString(m.st.Footer.Render(wrap(i18n.S(l, "setup.hint"), w)))
	return b.String()
}

// helpView is the keyboard reference overlay.
func (m *Model) helpView() string {
	l := m.lang
	body := m.st.Heading.Render(i18n.S(l, "help.title")) + "\n\n" +
		m.st.Text.Render(wrap(i18n.S(l, "help.body"), min(m.contentWidth(), 76))) + "\n\n" +
		m.st.Footer.Render(i18n.S(l, "help.close"))
	return m.st.Panel.Render(body)
}

// graveyardView lists finished runs, newest first.
func (m *Model) graveyardView() string {
	l := m.lang
	if len(m.history) == 0 {
		return m.st.Dim.Render(i18n.S(l, "grave.empty"))
	}
	var b strings.Builder
	for i := len(m.history) - 1; i >= 0; i-- {
		r := m.history[i]
		mark, st := "✝", m.st.Danger.UnsetBold()
		what := r.Reason.Get(l)
		if r.Victory {
			mark, st, what = "★", m.st.Good, i18n.S(l, "grave.victory")
		}
		region := m.cfg.Book.RegionName(r.Region).Get(l)
		prof := r.Profession
		if p, ok := m.cfg.DB.Profession(r.Profession); ok {
			prof = p.Name.Get(l)
		}
		b.WriteString(st.Render(fmt.Sprintf("%s %2d ", mark, r.Age)) +
			m.st.Text.Render(truncate(what, m.contentWidth()-6)) + "\n" +
			m.st.Dim.Render(fmt.Sprintf("     %s · %s · %s · %s · %d · seed %d",
				r.Time.Format("2006-01-02"), region, i18n.S(l, "sex."+r.Sex), prof, r.Score, r.Seed)) + "\n")
	}
	return b.String()
}

// truncate shortens s to w display columns with an ellipsis.
func truncate(s string, w int) string {
	if w <= 1 || lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r))+1 > w {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

func rule(w int, st Styles) string { return components.Rule(w, st.Dim) }
