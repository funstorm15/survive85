package tui

import "github.com/charmbracelet/lipgloss"

// Palette is a colour theme.
type Palette struct {
	Name    string
	Accent  lipgloss.Color // headings, cursor, frames
	Good    lipgloss.Color // gains, wise, grade A
	Warn    lipgloss.Color // costs, warnings, grade B
	Danger  lipgloss.Color // death, losses, risky
	Text    lipgloss.Color
	Dim     lipgloss.Color
	Subtle  lipgloss.Color // borders
	Special lipgloss.Color // evidence card, links
}

// Themes lists the built-in palettes; the first is the default.
var Themes = []Palette{
	{Name: "nord", Accent: "#88C0D0", Good: "#A3BE8C", Warn: "#EBCB8B", Danger: "#BF616A",
		Text: "#ECEFF4", Dim: "#7B88A1", Subtle: "#4C566A", Special: "#B48EAD"},
	{Name: "cyberpunk", Accent: "#00F0FF", Good: "#39FF14", Warn: "#FCEE0A", Danger: "#FF2A6D",
		Text: "#E0E0FF", Dim: "#8080B0", Subtle: "#3D3D6B", Special: "#D300C5"},
	{Name: "monokai", Accent: "#66D9EF", Good: "#A6E22E", Warn: "#E6DB74", Danger: "#F92672",
		Text: "#F8F8F2", Dim: "#90908A", Subtle: "#49483E", Special: "#AE81FF"},
}

// ThemeIndex returns the index of a theme by name, or 0.
func ThemeIndex(name string) int {
	for i, t := range Themes {
		if t.Name == name {
			return i
		}
	}
	return 0
}

// Styles are the lipgloss styles derived from a palette.
type Styles struct {
	P        Palette
	Title    lipgloss.Style
	Heading  lipgloss.Style
	Text     lipgloss.Style
	Dim      lipgloss.Style
	Good     lipgloss.Style
	Warn     lipgloss.Style
	Danger   lipgloss.Style
	Special  lipgloss.Style
	Cursor   lipgloss.Style
	Selected lipgloss.Style
	Disabled lipgloss.Style
	Panel    lipgloss.Style
	Card     lipgloss.Style
	Frame    lipgloss.Style
	Footer   lipgloss.Style
	Bar      lipgloss.Style
}

// NewStyles builds the styles for palette p.
func NewStyles(p Palette) Styles {
	base := lipgloss.NewStyle().Foreground(p.Text)
	return Styles{
		P:        p,
		Title:    lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		Heading:  lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		Text:     base,
		Dim:      lipgloss.NewStyle().Foreground(p.Dim),
		Good:     lipgloss.NewStyle().Foreground(p.Good),
		Warn:     lipgloss.NewStyle().Foreground(p.Warn),
		Danger:   lipgloss.NewStyle().Foreground(p.Danger).Bold(true),
		Special:  lipgloss.NewStyle().Foreground(p.Special),
		Cursor:   lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		Selected: lipgloss.NewStyle().Foreground(p.Text).Bold(true),
		Disabled: lipgloss.NewStyle().Foreground(p.Subtle).Strikethrough(true),
		Panel:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Subtle).Padding(0, 1),
		Card:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Special).Padding(0, 1),
		Frame:    lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(p.Danger).Padding(0, 2),
		Footer:   lipgloss.NewStyle().Foreground(p.Dim),
		Bar:      lipgloss.NewStyle().Foreground(p.Text).Bold(true),
	}
}

// Level picks good/warn/danger for a 0..100 stat.
func (s Styles) Level(v int) lipgloss.Style {
	switch {
	case v >= 60:
		return s.Good
	case v >= 30:
		return s.Warn
	default:
		return s.Danger
	}
}

// Grade colours an evidence grade.
func (s Styles) Grade(g string) lipgloss.Style {
	switch g {
	case "A":
		return s.Good.Bold(true)
	case "B":
		return s.Warn.Bold(true)
	default:
		return s.Dim.Bold(true)
	}
}

// Signed colours a delta: gains green, losses red.
func (s Styles) Signed(v int64, text string) string {
	switch {
	case v > 0:
		return s.Good.Render(text)
	case v < 0:
		return s.Danger.Render(text)
	default:
		return s.Dim.Render(text)
	}
}
