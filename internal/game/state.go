// Package game implements the rules of survive85: the player's state, the
// risk calculator and the year-by-year engine. It has no UI dependencies.
package game

import (
	"sort"

	"github.com/funstorm15/survive85/internal/actuary"
	"github.com/funstorm15/survive85/internal/events"
	"github.com/funstorm15/survive85/internal/i18n"
)

// Game constants. Ages are exact ages: a run starts on the 18th birthday and
// is won on the 85th.
const (
	StartAge   = 18
	VictoryAge = 85

	MaxStat = 100
	MaxWill = 10
)

// RiskFactor records where a part of a risk modifier came from, so the
// autopsy can name the decision that mattered.
type RiskFactor struct {
	Risk     events.Risk
	Factor   float64
	Age      int    // age when it was acquired
	Until    int    // last age it applies to; 0 = permanent
	EventID  string // empty for profession or condition factors
	ChoiceID string
	Evidence string
	// Label names the decision or condition in every language.
	Label        i18n.Text
	ConditionID  string
	DeathReasons map[actuary.Cause]i18n.Text
}

// Active reports whether the factor applies at age.
func (f RiskFactor) Active(age int) bool { return f.Until == 0 || age <= f.Until }

// ActiveCondition is a condition the player currently has.
type ActiveCondition struct {
	ID    string
	Since int // age acquired
	Until int // last age it applies to; 0 = until removed
}

// Decision is one answered event, kept for the run summary.
type Decision struct {
	Age      int
	EventID  string
	ChoiceID string
	Tag      events.Tag
	Evidence string
}

// Player is the complete state of a run. The exported fields are the
// "Player State" of the design document; HiddenModifiers is derived on demand
// from Factors and conditions.
type Player struct {
	Age    uint8
	Health int // 0..100
	Sanity int // 0..100
	Wealth int64
	Will   int // 0..MaxWill, spent on disciplined choices

	Profession string
	Region     string
	Sex        string

	// Factors are every permanent or temporary multiplier picked up so far.
	Factors    []RiskFactor
	Conditions []ActiveCondition
	// Flags maps a flag to the age it was set.
	Flags map[string]int

	Decisions []Decision
	// Seen counts how many times each event was offered; LastSeen is the age.
	Seen     map[string]int
	LastSeen map[string]int

	HoursInvested float64
	MoneySpent    int64
	YearsInPrison int
	Arrests       int
	PeakWealth    int64
}

// NewPlayer creates a player at StartAge for a profession.
func NewPlayer(p *events.Profession, region, sex string) *Player {
	pl := &Player{
		Age:        StartAge,
		Health:     clamp(orDefault(p.StartHealth, 85), 1, MaxStat),
		Sanity:     clamp(orDefault(p.StartSanity, 70), 1, MaxStat),
		Wealth:     p.StartWealth,
		Will:       clamp(orDefault(p.StartWill, 5), 0, MaxWill),
		Profession: p.ID,
		Region:     region,
		Sex:        sex,
		Flags:      map[string]int{},
		Seen:       map[string]int{},
		LastSeen:   map[string]int{},
		PeakWealth: p.StartWealth,
	}
	for _, f := range p.Flags {
		pl.Flags[f] = StartAge
	}
	return pl
}

// AgeInt is Age as an int, for arithmetic.
func (p *Player) AgeInt() int { return int(p.Age) }

// HasFlag reports whether a flag is set.
func (p *Player) HasFlag(f string) bool {
	_, ok := p.Flags[f]
	return ok
}

// HasCondition reports whether a condition is active.
func (p *Player) HasCondition(id string) bool {
	for _, c := range p.Conditions {
		if c.ID == id {
			return true
		}
	}
	return false
}

// HiddenModifiers returns the current value of every risk modifier, combining
// decision factors and active conditions, after the overlap correction and
// the clamp (see CombineModifier).
func (p *Player) HiddenModifiers(conds map[string]*events.Condition) map[events.Risk]float64 {
	out := make(map[events.Risk]float64, len(events.Risks))
	for _, r := range events.Risks {
		out[r] = 1
	}
	age := p.AgeInt()
	for _, f := range p.Factors {
		if f.Active(age) {
			out[f.Risk] *= f.Factor
		}
	}
	for _, c := range p.Conditions {
		if def := conds[c.ID]; def != nil {
			for r, m := range def.Modifiers {
				out[r] *= m
			}
		}
	}
	for r, v := range out {
		out[r] = CombineModifier(v)
	}
	return out
}

// TagCounts tallies decisions by tag.
func (p *Player) TagCounts() map[events.Tag]int {
	out := map[events.Tag]int{}
	for _, d := range p.Decisions {
		out[d.Tag]++
	}
	return out
}

// EvidenceUsed returns the distinct evidence ids behind the player's
// decisions, in first-seen order.
func (p *Player) EvidenceUsed() []string {
	var out []string
	seen := map[string]bool{}
	for _, d := range p.Decisions {
		if d.Evidence != "" && !seen[d.Evidence] {
			seen[d.Evidence] = true
			out = append(out, d.Evidence)
		}
	}
	return out
}

// TopFactors returns active factors sorted by how far they move their risk
// away from 1, strongest first.
func (p *Player) TopFactors(n int) []RiskFactor {
	age := p.AgeInt()
	var fs []RiskFactor
	for _, f := range p.Factors {
		if f.Active(age) && f.Factor != 1 {
			fs = append(fs, f)
		}
	}
	sort.SliceStable(fs, func(i, j int) bool { return logAbs(fs[i].Factor) > logAbs(fs[j].Factor) })
	if n > 0 && len(fs) > n {
		fs = fs[:n]
	}
	return fs
}

func orDefault(v, d int) int {
	if v == 0 {
		return d
	}
	return v
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
