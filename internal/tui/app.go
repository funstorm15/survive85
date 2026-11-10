// Package tui is the Bubble Tea interface of survive85: title and setup
// screens, the yearly event, the outcome with its evidence card, the autopsy
// report, victory, the graveyard and help. All text comes from the i18n
// catalog or from bilingual game data, so the language can switch live.
package tui

import (
	"errors"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/funstorm15/survive85/internal/actuary"
	"github.com/funstorm15/survive85/internal/events"
	"github.com/funstorm15/survive85/internal/game"
	"github.com/funstorm15/survive85/internal/i18n"
	"github.com/funstorm15/survive85/internal/tui/components"
)

// Config is everything the interface needs from main.
type Config struct {
	DB        *events.DB
	Book      *actuary.Book
	Lang      i18n.Lang
	Seed      uint64
	FixedSeed bool // seed came from -seed or -daily
	Theme     string
	Version   string
	// HistoryPath overrides where finished runs are stored (tests).
	HistoryPath string
}

// Run starts the interface and blocks until the player quits.
func Run(cfg Config) error {
	if cfg.DB == nil || cfg.Book == nil {
		return errors.New("tui: missing game data")
	}
	p := tea.NewProgram(New(cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Minimum terminal size the layout supports.
const (
	MinWidth  = 60
	MinHeight = 18
	maxWidth  = 100
)

type screen int

const (
	scrTitle screen = iota
	scrSetup
	scrEvent
	scrOutcome
	scrMorgue
	scrVictory
	scrGraveyard
	scrAbout
)

// setup fields, in display order.
const (
	fieldRegion = iota
	fieldSex
	fieldProfession
	fieldStart
	fieldCount
)

var sexes = []string{"male", "female"}

// Model is the root Bubble Tea model.
type Model struct {
	cfg    Config
	lang   i18n.Lang
	theme  int
	st     Styles
	w, h   int
	scr    screen
	back   screen // where Esc returns from graveyard/about
	help   bool
	cursor int // menu cursor or choice cursor

	region, sex, prof int // setup selections
	field             int
	seed              uint64

	g        *game.Game
	vp       viewport.Model
	history  []Record
	recorded bool
	flash    string
}

// New builds the model in its initial state (the title screen).
func New(cfg Config) *Model {
	m := &Model{
		cfg:   cfg,
		lang:  cfg.Lang,
		theme: ThemeIndex(cfg.Theme),
		seed:  cfg.Seed,
		w:     100,
		h:     32,
		vp:    viewport.New(80, 20),
	}
	if m.lang == "" {
		m.lang = i18n.EN
	}
	if m.lang == i18n.RU {
		m.region = indexOf(cfg.Book.Regions(), "rus")
	}
	m.st = NewStyles(Themes[m.theme])
	m.history = LoadHistory(m.historyPath())
	return m
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.refresh(false)
		return m, nil
	case tea.KeyMsg:
		return m.key(msg)
	}
	return m, nil
}

func (m *Model) key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if k == "ctrl+c" {
		return m, tea.Quit
	}
	if m.help {
		m.help = false
		return m, nil
	}
	m.flash = ""
	switch k {
	case "q":
		return m, tea.Quit
	case "?":
		m.help = true
		return m, nil
	case "L":
		m.lang = m.lang.Next()
		m.refresh(false)
		return m, nil
	case "T":
		m.theme = (m.theme + 1) % len(Themes)
		m.st = NewStyles(Themes[m.theme])
		m.flash = i18n.F(m.lang, "theme", Themes[m.theme].Name)
		m.refresh(false)
		return m, nil
	}
	switch m.scr {
	case scrTitle:
		return m, m.keyTitle(k)
	case scrSetup:
		m.keySetup(k)
	case scrEvent:
		m.keyEvent(k)
	case scrOutcome:
		m.keyOutcome(k, msg)
	case scrMorgue, scrVictory:
		m.keyEnd(k, msg)
	case scrGraveyard, scrAbout:
		switch k {
		case "esc", "h", "backspace", "enter":
			m.go2(m.back)
		default:
			m.scroll(msg)
		}
	}
	return m, nil
}

func isUp(k string) bool      { return k == "up" || k == "k" }
func isDown(k string) bool    { return k == "down" || k == "j" }
func isConfirm(k string) bool { return k == "enter" || k == " " || k == "l" }

// Title menu items.
const (
	menuNew = iota
	menuGraveyard
	menuAbout
	menuLang
	menuQuit
	menuCount
)

func (m *Model) keyTitle(k string) tea.Cmd {
	switch {
	case isUp(k):
		m.cursor = (m.cursor + menuCount - 1) % menuCount
	case isDown(k):
		m.cursor = (m.cursor + 1) % menuCount
	case isConfirm(k):
		switch m.cursor {
		case menuNew:
			m.field = fieldStart
			m.go2(scrSetup)
		case menuGraveyard:
			m.back = scrTitle
			m.go2(scrGraveyard)
		case menuAbout:
			m.back = scrTitle
			m.go2(scrAbout)
		case menuLang:
			m.lang = m.lang.Next()
		case menuQuit:
			return tea.Quit
		}
	}
	return nil
}

func (m *Model) keySetup(k string) {
	regions := m.cfg.Book.Regions()
	switch {
	case isUp(k):
		m.field = (m.field + fieldCount - 1) % fieldCount
	case isDown(k), k == "tab":
		m.field = (m.field + 1) % fieldCount
	case k == "h" || k == "left":
		m.change(-1, len(regions))
	case k == "right" || (k == "l" && m.field != fieldStart):
		m.change(+1, len(regions))
	case k == "esc":
		m.go2(scrTitle)
	case isConfirm(k):
		if m.field == fieldStart {
			m.start(false)
		} else {
			m.field++
		}
	}
}

func (m *Model) change(d, regions int) {
	wrap := func(v, n int) int { return ((v+d)%n + n) % n }
	switch m.field {
	case fieldRegion:
		m.region = wrap(m.region, regions)
	case fieldSex:
		m.sex = wrap(m.sex, len(sexes))
	case fieldProfession:
		m.prof = wrap(m.prof, len(m.cfg.DB.Professions))
	}
}

// start begins a run; replay keeps the current seed.
func (m *Model) start(replay bool) {
	if !replay && m.g != nil && !m.cfg.FixedSeed {
		m.seed = nextSeed(m.seed)
	}
	opts := game.Options{
		Region:     m.cfg.Book.Regions()[m.region],
		Sex:        sexes[m.sex],
		Profession: m.cfg.DB.Professions[m.prof].ID,
		Seed:       m.seed,
	}
	g, err := game.New(m.cfg.DB, m.cfg.Book, opts)
	if err != nil {
		m.flash = err.Error()
		return
	}
	m.g = g
	m.recorded = false
	m.cursor = 0
	m.go2(scrEvent)
}

func (m *Model) keyEvent(k string) {
	g := m.g
	vis := g.VisibleChoices()
	if g.Current == nil || len(vis) == 0 {
		if isConfirm(k) && g.Current == nil {
			m.choose(0)
		}
		return
	}
	switch {
	case isUp(k):
		m.cursor = (m.cursor + len(vis) - 1) % len(vis)
	case isDown(k):
		m.cursor = (m.cursor + 1) % len(vis)
	case isConfirm(k):
		m.choose(vis[m.cursor])
	case len(k) == 1 && k[0] >= '1' && k[0] <= '9':
		n := int(k[0] - '1')
		if n < len(vis) {
			m.cursor = n
			m.choose(vis[n])
		}
	}
}

func (m *Model) choose(i int) {
	if _, err := m.g.Choose(i); err != nil {
		if errors.Is(err, game.ErrUnavailable) {
			m.flash = m.unavailableText(i)
		}
		return
	}
	m.cursor = 0
	m.go2(scrOutcome)
}

func (m *Model) keyOutcome(k string, msg tea.KeyMsg) {
	if k != "enter" && k != " " {
		m.scroll(msg)
		return
	}
	switch m.g.Phase {
	case game.PhaseDead:
		m.record()
		m.go2(scrMorgue)
	case game.PhaseVictory:
		m.record()
		m.go2(scrVictory)
	default:
		m.g.Next()
		m.go2(scrEvent)
	}
}

func (m *Model) keyEnd(k string, msg tea.KeyMsg) {
	switch k {
	case "enter":
		m.field = fieldStart
		m.go2(scrSetup)
	case "r":
		m.start(true)
	case "g":
		m.back = m.scr
		m.go2(scrGraveyard)
	default:
		m.scroll(msg)
	}
}

func (m *Model) scroll(msg tea.KeyMsg) {
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	_ = cmd
}

func (m *Model) go2(s screen) {
	m.scr = s
	if s == scrTitle {
		m.cursor = 0
	}
	m.refresh(true)
}

// scrollable reports whether the current screen renders through the viewport.
func (m *Model) scrollable() bool {
	switch m.scr {
	case scrOutcome, scrMorgue, scrVictory, scrGraveyard, scrAbout:
		return true
	}
	return false
}

// refresh re-renders viewport content after a state, size or language change.
func (m *Model) refresh(top bool) {
	if !m.scrollable() {
		return
	}
	header, footer := m.chrome()
	m.vp.Width = m.contentWidth()
	m.vp.Height = max(3, m.h-lipgloss.Height(header)-lipgloss.Height(footer)-1)
	m.vp.SetContent(m.scrollContent())
	if top {
		m.vp.GotoTop()
	}
}

func (m *Model) contentWidth() int {
	return max(MinWidth-4, min(m.w-4, maxWidth))
}

// View implements tea.Model.
func (m *Model) View() string {
	if m.w < MinWidth || m.h < MinHeight {
		return m.st.Warn.Render(i18n.F(m.lang, "small", MinWidth, MinHeight, m.w, m.h))
	}
	if m.help {
		return m.place(m.helpView())
	}
	var body string
	switch m.scr {
	case scrTitle:
		body = m.titleView()
	case scrSetup:
		body = m.setupView()
	case scrEvent:
		body = m.eventView()
	default:
		header, footer := m.chrome()
		body = lipgloss.JoinVertical(lipgloss.Left, header, m.vp.View(), footer)
	}
	if m.flash != "" {
		body += "\n" + m.st.Warn.Render(m.flash)
	}
	return lipgloss.NewStyle().Padding(0, 1).Render(body)
}

// chrome returns the fixed header and footer around a scrollable screen.
func (m *Model) chrome() (string, string) {
	w := m.contentWidth()
	switch m.scr {
	case scrOutcome:
		key := "out.continue"
		switch m.g.Phase {
		case game.PhaseDead:
			key = "out.dead"
		case game.PhaseVictory:
			key = "out.victory"
		}
		return m.statusBar() + "\n" + rule(w, m.st), m.footer(key)
	case scrMorgue:
		return m.st.Danger.Render(i18n.F(m.lang, "morgue.title", m.g.Seed%100000)), m.footer("morgue.keys")
	case scrVictory:
		return m.st.Good.Bold(true).Render(i18n.S(m.lang, "victory.title")), m.footer("victory.keys")
	case scrGraveyard:
		return m.st.Heading.Render(i18n.S(m.lang, "grave.title")), m.footer("grave.keys")
	case scrAbout:
		return m.st.Heading.Render(i18n.S(m.lang, "about.title")), m.footer("about.keys")
	}
	return "", ""
}

func (m *Model) scrollContent() string {
	switch m.scr {
	case scrOutcome:
		return m.outcomeView()
	case scrMorgue:
		return m.morgueView()
	case scrVictory:
		return m.victoryView()
	case scrGraveyard:
		return m.graveyardView()
	case scrAbout:
		return wrap(i18n.S(m.lang, "about.body"), m.contentWidth())
	}
	return ""
}

func (m *Model) footer(key string) string {
	return m.st.Footer.Render(wrap(i18n.S(m.lang, key), m.contentWidth()))
}

func (m *Model) place(s string) string {
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, s)
}

// record appends the finished run to the history once.
func (m *Model) record() {
	if m.recorded || m.g == nil {
		return
	}
	m.recorded = true
	r := NewRecord(m.g, m.cfg.Version, time.Now())
	m.history = append(m.history, r)
	_ = AppendHistory(m.historyPath(), r) // a read-only home must not crash the game
}

func (m *Model) historyPath() string {
	if m.cfg.HistoryPath != "" {
		return m.cfg.HistoryPath
	}
	return DefaultHistoryPath()
}

// nextSeed derives a fresh seed (splitmix64).
func nextSeed(s uint64) uint64 {
	s += 0x9e3779b97f4a7c15
	z := s
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func indexOf(list []string, s string) int {
	for i, x := range list {
		if x == s {
			return i
		}
	}
	return 0
}

func wrap(s string, w int) string { return components.Wrap(s, w) }
