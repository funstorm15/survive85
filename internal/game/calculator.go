package game

import (
	"math"
	"sort"

	"github.com/funstorm15/survive85/internal/actuary"
	"github.com/funstorm15/survive85/internal/events"
	"github.com/funstorm15/survive85/internal/i18n"
)

// Bounds on a risk modifier. Effects stack multiplicatively, but real risk
// factors overlap, so the product is clamped: no lifestyle removes more than
// three quarters of a cause's mortality, and none multiplies it beyond 25×.
const (
	MinModifier = 0.25
	MaxModifier = 25.0

	// MaxInstantChance caps any single instant check.
	MaxInstantChance = 0.95
	// ArrestPerExcessRisk is the yearly arrest probability per unit of legal
	// risk above the baseline (legal risk 2.0 → 1.2% a year).
	ArrestPerExcessRisk = 0.012
	MaxArrestChance     = 0.25
)

// Overlap correction. Risk factors are not independent — obesity acts partly
// through blood pressure, smoking and drinking travel together — so beyond
// OverlapKnee a combined modifier grows only as a power OverlapExponent of
// the excess, on either side of 1.
const (
	OverlapKnee     = 2.0
	OverlapExponent = 0.6
)

// CombineModifier applies the overlap correction and the clamp to a raw
// product of factors.
func CombineModifier(v float64) float64 {
	switch {
	case v > OverlapKnee:
		v = OverlapKnee * math.Pow(v/OverlapKnee, OverlapExponent)
	case v < 1/OverlapKnee:
		v = (1 / OverlapKnee) * math.Pow(v*OverlapKnee, OverlapExponent)
	}
	return ClampModifier(v)
}

// ClampModifier keeps a modifier within [MinModifier, MaxModifier].
func ClampModifier(v float64) float64 {
	return math.Max(MinModifier, math.Min(MaxModifier, v))
}

// ExpectedHealth is the Health of a typical person at age: the reference at
// which the life table applies unchanged. It is flat at 80 until 50 and falls
// by one point a year after that.
func ExpectedHealth(age int) int {
	if age <= 50 {
		return 80
	}
	return 80 - (age - 50)
}

// HealthFactor scales all non-injury mortality by how the player's Health
// compares with ExpectedHealth for their age: 1.0 at the expected value,
// halving every 30 HP above and doubling every 30 HP below, within [0.6, 4].
// It is a game rule that ties the visible stat to the actuarial model.
func HealthFactor(health, age int) float64 {
	return math.Max(0.6, math.Min(4, math.Pow(2, float64(ExpectedHealth(age)-health)/30)))
}

// SelfHarmFactor scales the self-harm share of mortality by Sanity: 1.0 at
// 50, doubling every 20 points below and halving every 20 above, within
// [0.4, 6].
func SelfHarmFactor(sanity int) float64 {
	return math.Max(0.4, math.Min(6, math.Pow(2, float64(50-sanity)/20)))
}

// CauseRisk is one cause's part of the annual mortality check.
type CauseRisk struct {
	Cause    actuary.Cause
	Base     float64 // baseline yearly probability for this cause
	Modifier float64 // combined multiplier applied to the baseline hazard
	Hazard   float64 // adjusted force of mortality
	Share    float64 // share of the adjusted total (0..1)
}

// Mortality is the full annual mortality computation for one year of life.
type Mortality struct {
	Age      int
	BaseQ    float64 // life-table probability of dying this year
	Total    float64 // adjusted probability after modifiers
	Causes   []CauseRisk
	Relative float64 // Total / BaseQ
}

// AnnualMortality applies the player's modifiers to the life table under
// proportional hazards: the baseline force of mortality mu = -ln(1-q) is
// split by cause shares, each cause's hazard is multiplied by its modifier
// (this is exactly how a hazard ratio acts), and the year's probability of
// death is 1 - exp(-sum of hazards).
func AnnualMortality(book *actuary.Book, t *actuary.Table, age int, mods map[events.Risk]float64, health, sanity int) Mortality {
	q := book.QAt(t, age)
	mu := -math.Log(1 - q)
	hf := HealthFactor(health, age)
	m := Mortality{Age: age, BaseQ: q}
	sum := 0.0
	for _, c := range actuary.Causes {
		f := book.Fraction(t, age, c)
		mod := 1.0
		switch c {
		case actuary.SelfHarm:
			mod = SelfHarmFactor(sanity)
		case actuary.Trauma:
			// Injuries do not depend on fitness: only the risk modifier applies.
			mod = mods[events.RiskTrauma]
		default:
			if r, ok := events.RiskFor(c); ok {
				mod = mods[r]
			}
			mod *= hf
		}
		h := mu * f * mod
		sum += h
		m.Causes = append(m.Causes, CauseRisk{Cause: c, Base: 1 - math.Exp(-mu*f), Modifier: mod, Hazard: h})
	}
	m.Total = 1 - math.Exp(-sum)
	if sum > 0 {
		for i := range m.Causes {
			m.Causes[i].Share = m.Causes[i].Hazard / sum
		}
	}
	if q > 0 {
		m.Relative = m.Total / q
	}
	return m
}

// PickCause chooses the cause of a death that happened, weighted by share.
// roll must be uniform in [0, 1).
func (m Mortality) PickCause(roll float64) actuary.Cause {
	acc := 0.0
	for _, c := range m.Causes {
		acc += c.Share
		if roll < acc {
			return c.Cause
		}
	}
	return actuary.Other
}

// SortedCauses returns the causes ordered by share, largest first.
func (m Mortality) SortedCauses() []CauseRisk {
	out := append([]CauseRisk(nil), m.Causes...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Share > out[j].Share })
	return out
}

// ScaledChance multiplies an instant-check chance by the named modifiers.
// Scaling can raise a chance at most to MaxInstantChance; a chance the data
// sets higher (a certainty) is kept.
func ScaledChance(chance float64, mods map[events.Risk]float64, scaleBy []events.Risk) float64 {
	limit := math.Max(MaxInstantChance, math.Min(chance, 1))
	for _, r := range scaleBy {
		if v, ok := mods[r]; ok {
			chance *= v
		}
	}
	return math.Max(0, math.Min(limit, chance))
}

// ArrestChance is the yearly probability that past legal risks catch up.
// A clean record (modifier ≤ 1) is never arrested by this check.
func ArrestChance(legal float64) float64 {
	return math.Max(0, math.Min(MaxArrestChance, ArrestPerExcessRisk*(legal-1)))
}

// CohortDiedBefore is the share of people with the same life table who die
// between StartAge and age: "you outlived N% of your peers".
func CohortDiedBefore(book *actuary.Book, t *actuary.Table, age int) float64 {
	if age <= StartAge {
		return 0
	}
	return 1 - book.Survival(t, StartAge, age)
}

// Contribution is a single source of risk for a cause, used by the autopsy.
type Contribution struct {
	Factor       float64
	Label        i18n.Text
	EventID      string
	ChoiceID     string
	ConditionID  string
	Evidence     string
	Age          int
	DeathReasons map[actuary.Cause]i18n.Text
}

// DecisiveFactor finds the strongest harmful contribution to the risk that
// scales cause c: a decision factor or an active condition. ok is false when
// nothing raised the risk above the baseline.
func DecisiveFactor(p *Player, conds map[string]*events.Condition, c actuary.Cause) (Contribution, bool) {
	r, ok := events.RiskFor(c)
	if !ok {
		return Contribution{}, false
	}
	var best Contribution
	age := p.AgeInt()
	for _, f := range p.Factors {
		if f.Risk == r && f.Active(age) && f.Factor > best.Factor && f.Factor > 1 {
			best = Contribution{Factor: f.Factor, Label: f.Label, EventID: f.EventID, ChoiceID: f.ChoiceID,
				ConditionID: f.ConditionID, Evidence: f.Evidence, Age: f.Age, DeathReasons: f.DeathReasons}
		}
	}
	for _, ac := range p.Conditions {
		def := conds[ac.ID]
		if def == nil {
			continue
		}
		if m := def.Modifiers[r]; m > 1 && m > best.Factor {
			best = Contribution{Factor: m, Label: def.Name, ConditionID: def.ID, Evidence: def.Evidence,
				Age: ac.Since, DeathReasons: def.DeathReasons}
		}
	}
	return best, best.Factor > 1
}

func logAbs(f float64) float64 {
	if f <= 0 {
		return 0
	}
	return math.Abs(math.Log(f))
}
