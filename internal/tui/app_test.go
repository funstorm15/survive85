package tui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/funstorm15/survive85/data"
	"github.com/funstorm15/survive85/internal/actuary"
	"github.com/funstorm15/survive85/internal/events"
	"github.com/funstorm15/survive85/internal/game"
	"github.com/funstorm15/survive85/internal/i18n"
)

func newModel(t *testing.T, lang i18n.Lang, seed uint64) *Model {
	t.Helper()
	db, err := events.Load(data.FS)
	if err != nil {
		t.Fatal(err)
	}
	book, err := actuary.Load(data.FS, data.LifeTables)
	if err != nil {
		t.Fatal(err)
	}
	m := New(Config{DB: db, Book: book, Lang: lang, Seed: seed, FixedSeed: true, Version: "test",
		HistoryPath: filepath.Join(t.TempDir(), "history.jsonl")})
	m.Update(tea.WindowSizeMsg{Width: 110, Height: 40})
	return m
}

func press(m *Model, keys ...string) {
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "up":
			msg = tea.KeyMsg{Type: tea.KeyUp}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		m.Update(msg)
	}
}

func view(m *Model) string { return ansi.Strip(m.View()) }

func TestTitleScreenBothLanguages(t *testing.T) {
	m := newModel(t, i18n.RU, 1)
	v := view(m)
	for _, want := range []string{"Новая жизнь", "Кладбище", "Выход"} {
		if !strings.Contains(v, want) {
			t.Errorf("RU title lacks %q:\n%s", want, v)
		}
	}
	press(m, "L")
	if v := view(m); !strings.Contains(v, "New life") || !strings.Contains(v, "Graveyard") {
		t.Errorf("L must switch to English:\n%s", v)
	}
}

func TestPlayThroughToTheEnd(t *testing.T) {
	m := newModel(t, i18n.RU, 7)
	press(m, "enter") // title → setup
	if m.scr != scrSetup || !strings.Contains(view(m), "НОВАЯ ЖИЗНЬ") {
		t.Fatalf("expected setup, got screen %d:\n%s", m.scr, view(m))
	}
	if !strings.Contains(view(m), "Шанс дожить с 18 до 85") {
		t.Error("setup must show the life-table baseline")
	}
	press(m, "enter") // start
	if m.scr != scrEvent || m.g == nil {
		t.Fatalf("expected event screen, got %d", m.scr)
	}
	v := view(m)
	for _, want := range []string{"[ВОЗРАСТ: 18/85]", "HP: [", "SANITY: [", "КАПИТАЛ:", "ВОЛЯ:"} {
		if !strings.Contains(v, want) {
			t.Errorf("status bar lacks %q:\n%s", want, v)
		}
	}

	switched := false
	for steps := 0; steps < 400 && m.scr != scrMorgue && m.scr != scrVictory; steps++ {
		switch m.scr {
		case scrEvent:
			// Pick the first affordable choice, preferring the last one listed
			// (often the risky one) so the run ends sooner.
			vis := m.g.VisibleChoices()
			pick := 0
			for n := len(vis) - 1; n >= 0; n-- {
				if m.g.Current == nil || m.g.Current.Choices[vis[n]].Availability(m.g.State()) == events.Available {
					pick = n
					break
				}
			}
			press(m, string(rune('1'+pick)))
			if m.scr == scrEvent && m.g.Current == nil {
				press(m, "enter")
			}
		case scrOutcome:
			if !strings.Contains(view(m), "ИТОГИ ГОДА") && m.g.Phase != game.PhaseDead {
				t.Fatalf("outcome must show the year report:\n%s", view(m))
			}
			if !switched && m.g.Player.Age > 25 {
				press(m, "L")
				if !strings.Contains(view(m), "THE YEAR IN REVIEW") {
					t.Fatalf("live language switch failed:\n%s", view(m))
				}
				press(m, "L")
				switched = true
			}
			press(m, "enter")
		}
	}
	switch m.scr {
	case scrMorgue:
		v := view(m)
		for _, want := range []string{"ПРОТОКОЛ ВСКРЫТИЯ", "Причина смерти:", "ПОКОЙСЯ С МИРОМ", "ВСКРЫТЫЕ МОДИФИКАТОРЫ", "ИТОГОВЫЙ СЧЁТ"} {
			if !strings.Contains(v+m.vp.View(), want) && !strings.Contains(m.scrollContent(), want) {
				t.Errorf("morgue lacks %q", want)
			}
		}
	case scrVictory:
		if !strings.Contains(m.scrollContent(), "85") {
			t.Error("victory screen lacks 85")
		}
	default:
		t.Fatalf("run did not finish (screen %d, age %d)", m.scr, m.g.Player.Age)
	}
	if len(LoadHistory(m.historyPath())) != 1 {
		t.Fatal("a finished run must be saved to history")
	}

	// r replays the same seed; Enter goes back to setup.
	seed := m.g.Seed
	press(m, "r")
	if m.scr != scrEvent || m.g.Seed != seed || m.g.Player.Age != game.StartAge {
		t.Fatalf("replay must restart the same seed at 18")
	}
}

func TestTinyTerminalAndHelp(t *testing.T) {
	m := newModel(t, i18n.EN, 1)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	if !strings.Contains(view(m), "too small") {
		t.Fatalf("tiny window must show a notice:\n%s", view(m))
	}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	press(m, "?")
	if !strings.Contains(view(m), "CONTROLS") {
		t.Fatal("? must open help")
	}
	press(m, "x")
	if strings.Contains(view(m), "CONTROLS") {
		t.Fatal("any key must close help")
	}
}

func TestEventFitsEightyColumns(t *testing.T) {
	m := newModel(t, i18n.RU, 3)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	press(m, "enter", "enter")
	for _, line := range strings.Split(view(m), "\n") {
		if w := ansi.StringWidth(line); w > 80 {
			t.Fatalf("line wider than 80 columns (%d): %q", w, line)
		}
	}
}

func TestQuitKeys(t *testing.T) {
	m := newModel(t, i18n.EN, 1)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("q must quit")
	}
	if msg := cmd(); msg != (tea.QuitMsg{}) {
		t.Fatalf("q produced %T", msg)
	}
}

// Every event, with all of its choices, must fit an 80×24 terminal.
func TestEveryEventFitsSmallTerminal(t *testing.T) {
	for _, lang := range i18n.All {
		m := newModel(t, lang, 5)
		m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		press(m, "enter", "enter")
		for _, ev := range m.cfg.DB.Events {
			m.g.Current = ev
			v := view(m)
			if h := strings.Count(v, "\n") + 1; h > 24 {
				t.Errorf("%s/%s: event screen is %d lines tall", lang, ev.ID, h)
			}
			for _, line := range strings.Split(v, "\n") {
				if w := ansi.StringWidth(line); w > 80 {
					t.Errorf("%s/%s: line %d columns wide", lang, ev.ID, w)
				}
			}
		}
	}
}

func TestFormatNumbersKeepsPrecision(t *testing.T) {
	lo, hi := 0.5, 0.68
	got := formatNumbers([]events.Number{{Kind: "OR", Value: 0.58, CILow: &lo, CIHigh: &hi}}, i18n.EN)
	if got != "OR 0.58 (95% CI 0.50–0.68)" {
		t.Fatalf("formatNumbers = %q", got)
	}
}
