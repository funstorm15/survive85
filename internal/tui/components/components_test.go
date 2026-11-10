package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestTombstoneIsRectangularOnTheWalls(t *testing.T) {
	art := Tombstone("ПОКОЙСЯ С МИРОМ", "18 — 47", "Инфаркт миокарда у курильщика со стажем")
	lines := strings.Split(art, "\n")
	want := -1
	for _, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "|") {
			if want < 0 {
				want = lipgloss.Width(l)
			}
			if lipgloss.Width(l) != want {
				t.Fatalf("wall lines differ in width:\n%s", art)
			}
		}
	}
	if !strings.Contains(art, "ПОКОЙСЯ С МИРОМ") {
		t.Fatal("text missing")
	}
	t.Log("\n" + art)
}

func TestGaugeAndPips(t *testing.T) {
	plain := lipgloss.NewStyle()
	if g := Gauge(80, 100, 10, plain, plain); g != "████████░░" {
		t.Errorf("Gauge = %q", g)
	}
	if g := Gauge(-5, 100, 4, plain, plain); g != "░░░░" {
		t.Errorf("Gauge clamps low: %q", g)
	}
	if p := Pips(3, 5, plain, plain); p != "◆◆◆◇◇" {
		t.Errorf("Pips = %q", p)
	}
}

func TestLogoLinesHaveEqualWidth(t *testing.T) {
	for _, art := range [][]string{Logo, SmallLogo} {
		for _, l := range art {
			if lipgloss.Width(l) != lipgloss.Width(art[0]) {
				t.Fatalf("logo line width differs: %q", l)
			}
		}
	}
}
