// Package events defines the game's data vocabulary — events, choices,
// conditions and evidence records — and loads it from the embedded YAML.
package events

import (
	"fmt"

	"github.com/funstorm15/survive85/internal/actuary"
	"github.com/funstorm15/survive85/internal/i18n"
	"go.yaml.in/yaml/v3"
)

// Risk is a hidden, multiplicative risk modifier. 1.0 is the population
// baseline from the life table; 2.0 doubles the matching share of mortality.
type Risk string

const (
	RiskCardio Risk = "cardiovascular_risk"
	RiskLiver  Risk = "toxicology_liver_risk"
	RiskTrauma Risk = "trauma_accident_risk"
	RiskLegal  Risk = "legal_prison_risk"
	RiskOnco   Risk = "oncology_risk"
)

// Risks lists every risk modifier in display order.
var Risks = []Risk{RiskCardio, RiskOnco, RiskLiver, RiskTrauma, RiskLegal}

// Valid reports whether r is a known risk modifier.
func (r Risk) Valid() bool {
	for _, x := range Risks {
		if x == r {
			return true
		}
	}
	return false
}

// Cause maps a risk modifier to the cause-of-death group it scales.
// The legal risk has no cause: it drives arrests, not deaths.
func (r Risk) Cause() (actuary.Cause, bool) {
	switch r {
	case RiskCardio:
		return actuary.Cardiovascular, true
	case RiskOnco:
		return actuary.Oncology, true
	case RiskLiver:
		return actuary.Liver, true
	case RiskTrauma:
		return actuary.Trauma, true
	}
	return "", false
}

// RiskFor returns the risk modifier that scales cause c, if any.
func RiskFor(c actuary.Cause) (Risk, bool) {
	for _, r := range Risks {
		if rc, ok := r.Cause(); ok && rc == c {
			return r, true
		}
	}
	return "", false
}

// Category groups events for display.
type Category string

const (
	CatHealth    Category = "health"
	CatChronic   Category = "chronic"
	CatEmergency Category = "emergency"
	CatPharmacy  Category = "pharmacy"
	CatSafety    Category = "safety"
	CatMoney     Category = "money"
	CatLegal     Category = "legal"
	CatWork      Category = "work"
	CatFamily    Category = "family"
	CatMind      Category = "mind"
	CatTravel    Category = "travel"
	CatLife      Category = "life"
)

// Categories lists the valid categories.
var Categories = []Category{CatHealth, CatChronic, CatEmergency, CatPharmacy, CatSafety, CatMoney, CatLegal, CatWork, CatFamily, CatMind, CatTravel, CatLife}

// Tag classifies a choice for end-of-run statistics and the simulator's
// strategies. It never changes the rules.
type Tag string

const (
	TagWise    Tag = "wise"
	TagNeutral Tag = "neutral"
	TagRisky   Tag = "risky"
)

// RiskEffect multiplies a risk modifier, permanently (Years == 0) or for a
// number of years. YAML accepts a bare number (permanent) or {factor, years}.
type RiskEffect struct {
	Factor float64 `yaml:"factor"`
	Years  int     `yaml:"years"`
}

// UnmarshalYAML implements the scalar shorthand.
func (e *RiskEffect) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		var f float64
		if err := node.Decode(&f); err != nil {
			return fmt.Errorf("line %d: risk effect must be a number or {factor, years}", node.Line)
		}
		*e = RiskEffect{Factor: f}
		return nil
	}
	type plain RiskEffect
	return node.Decode((*plain)(e))
}

// Requirements gate an event or a choice on the player's state.
type Requirements struct {
	Flags         []string         `yaml:"flags"`          // all must be set
	NotFlags      []string         `yaml:"not_flags"`      // none may be set
	FlagYears     map[string]int   `yaml:"flag_years"`     // flag set at least N years ago
	Conditions    []string         `yaml:"conditions"`     // all must be active
	NotConditions []string         `yaml:"not_conditions"` // none may be active
	MinWealth     *int64           `yaml:"min_wealth"`
	MaxWealth     *int64           `yaml:"max_wealth"`
	MinHealth     *int             `yaml:"min_health"`
	MaxHealth     *int             `yaml:"max_health"`
	MinSanity     *int             `yaml:"min_sanity"`
	MaxSanity     *int             `yaml:"max_sanity"`
	MinRisk       map[Risk]float64 `yaml:"min_risk"` // modifier at least this high
	Classes       []string         `yaml:"classes"`
	Sex           string           `yaml:"sex"`
}

// Effects are the stat changes a choice, a check or a condition applies.
type Effects struct {
	DeltaHealth      int                 `yaml:"delta_health"`
	DeltaSanity      int                 `yaml:"delta_sanity"`
	DeltaWealth      int64               `yaml:"delta_wealth"`
	DeltaWill        int                 `yaml:"delta_will"`
	StatEffects      map[Risk]RiskEffect `yaml:"stat_effects"`
	AddConditions    []string            `yaml:"add_conditions"`
	RemoveConditions []string            `yaml:"remove_conditions"`
	SetFlags         []string            `yaml:"set_flags"`
	ClearFlags       []string            `yaml:"clear_flags"`
}

// DeathCheck is a chance of dying on the spot right after a choice.
type DeathCheck struct {
	Chance      float64       `yaml:"chance"`
	ScaleBy     []Risk        `yaml:"scale_by"` // chance *= current modifier(s)
	Cause       actuary.Cause `yaml:"cause"`
	DeathReason i18n.Text     `yaml:"death_reason"`
	Evidence    string        `yaml:"evidence"`
}

// CatastropheKind is a non-fatal disaster.
type CatastropheKind string

const (
	KindPrison CatastropheKind = "prison"
	KindInjury CatastropheKind = "injury"
	KindRuin   CatastropheKind = "ruin"
	KindLoss   CatastropheKind = "loss" // someone close dies or is harmed
)

// CatastropheCheck is a chance of a severe non-fatal outcome after a choice.
type CatastropheCheck struct {
	Chance         float64         `yaml:"chance"`
	ScaleBy        []Risk          `yaml:"scale_by"`
	Kind           CatastropheKind `yaml:"kind"`
	Years          int             `yaml:"years"`           // prison term
	WealthFraction float64         `yaml:"wealth_fraction"` // share of positive wealth lost
	Effects        `yaml:",inline"`
	Text           i18n.Text `yaml:"text"`
	Evidence       string    `yaml:"evidence"`
}

// Choice is one option of an event.
type Choice struct {
	ID                string       `yaml:"id"`
	Text              i18n.Text    `yaml:"text"`
	Tag               Tag          `yaml:"tag"`
	CostMoney         int64        `yaml:"cost_money"`
	CostWill          int          `yaml:"cost_will"`
	CostTime          float64      `yaml:"cost_time"` // hours over the year, for statistics
	Requires          Requirements `yaml:"requires"`
	Effects           `yaml:",inline"`
	InstantDeathCheck *DeathCheck       `yaml:"instant_death_check"`
	CatastropheCheck  *CatastropheCheck `yaml:"catastrophe_check"`
	Outcome           i18n.Text         `yaml:"outcome"`
	Evidence          string            `yaml:"evidence"`
	// DeathReasons names the death this choice would be blamed for, per cause,
	// when it turns out to be the decisive risk factor.
	DeathReasons map[actuary.Cause]i18n.Text `yaml:"death_reasons"`
}

// Event is a situation offered to the player in one year of life.
type Event struct {
	ID         string           `yaml:"id"`
	Title      i18n.Text        `yaml:"title"`
	MinAge     int              `yaml:"min_age"`
	MaxAge     int              `yaml:"max_age"`
	Category   Category         `yaml:"category"`
	Weight     float64          `yaml:"weight"`
	WeightBy   map[Risk]float64 `yaml:"weight_by"` // weight *= modifier^exponent
	Repeatable bool             `yaml:"repeatable"`
	Cooldown   int              `yaml:"cooldown"` // years before a repeatable event can return
	Requires   Requirements     `yaml:"requires"`
	Narrative  i18n.Text        `yaml:"narrative"`
	Choices    []Choice         `yaml:"choices"`
	Evidence   string           `yaml:"evidence"`

	// Source is the file the event was loaded from, for error messages.
	Source string `yaml:"-"`
}

// Condition is a lasting state — a habit, an illness, a safety measure, a
// prison term — that keeps modifying risks and stats while it is active.
type Condition struct {
	ID             string                      `yaml:"id"`
	Name           i18n.Text                   `yaml:"name"`
	Description    i18n.Text                   `yaml:"description"`
	Icon           string                      `yaml:"icon"`
	Kind           string                      `yaml:"kind"` // good | bad | neutral, for colouring
	Hidden         bool                        `yaml:"hidden"`
	Years          int                         `yaml:"years"` // 0 = until removed
	Modifiers      map[Risk]float64            `yaml:"modifiers"`
	PerYear        Effects                     `yaml:"per_year"`
	SuppressIncome bool                        `yaml:"suppress_income"`
	ForcesEvent    string                      `yaml:"forces_event"`
	Evidence       string                      `yaml:"evidence"`
	DeathReasons   map[actuary.Cause]i18n.Text `yaml:"death_reasons"`
}

// Number is a statistic quoted from a study.
type Number struct {
	Kind    string    `yaml:"kind"` // HR, aHR, OR, aOR, RR, IRR, percent, absolute, other
	Value   float64   `yaml:"value"`
	CILow   *float64  `yaml:"ci_low"`
	CIHigh  *float64  `yaml:"ci_high"`
	Outcome i18n.Text `yaml:"outcome"`
}

// Citation is one source as written in the book.
type Citation struct {
	Citation string `yaml:"citation"`
	URL      string `yaml:"url"`
}

// Evidence is one item of the book "HowToLiveBetter", the scientific basis
// for a choice, a condition or an instant check.
type Evidence struct {
	ID           string     `yaml:"id"` // chNN_iMM
	Chapter      int        `yaml:"chapter"`
	Item         int        `yaml:"item"`
	ChapterTitle i18n.Text  `yaml:"chapter_title"`
	Title        i18n.Text  `yaml:"title"`
	Grade        string     `yaml:"grade"` // A, B or C
	Contested    bool       `yaml:"contested"`
	Measure      string     `yaml:"measure"` // mortality | money | time | freedom
	Plain        i18n.Text  `yaml:"plain"`
	Benefit      i18n.Text  `yaml:"benefit"`
	Numbers      []Number   `yaml:"numbers"`
	Sources      []Citation `yaml:"sources"`
	DOI          string     `yaml:"doi"`
	PubMed       string     `yaml:"pubmed"`
}

// OriginalURL deep-links to the item on the book's official search page.
func (e *Evidence) OriginalURL() string {
	return fmt.Sprintf("https://eternity4719.github.io/HowToLiveBetter/#e-%d-%d", e.Chapter, e.Item)
}

// TranslationURL deep-links to the item in the community translation for l.
func (e *Evidence) TranslationURL(l i18n.Lang) string {
	return fmt.Sprintf("https://book.dlgrv.com/%s/#e-%d-%d", l, e.Chapter, e.Item)
}

// PrimaryLink returns the best link for the evidence: DOI, PubMed, then the
// first source URL.
func (e *Evidence) PrimaryLink() string {
	switch {
	case e.DOI != "":
		return "https://doi.org/" + e.DOI
	case e.PubMed != "":
		return e.PubMed
	}
	for _, s := range e.Sources {
		if s.URL != "" {
			return s.URL
		}
	}
	return ""
}

// Profession is a starting class.
type Profession struct {
	ID          string    `yaml:"id"`
	Name        i18n.Text `yaml:"name"`
	Description i18n.Text `yaml:"description"`
	Icon        string    `yaml:"icon"`
	StartWealth int64     `yaml:"start_wealth"`
	StartHealth int       `yaml:"start_health"`
	StartSanity int       `yaml:"start_sanity"`
	StartWill   int       `yaml:"start_will"`
	// Income is the yearly amount left after living costs, before the career
	// curve. StudyUntil delays it (students live on 0 until that age).
	Income       int64            `yaml:"income"`
	IncomeSpread float64          `yaml:"income_spread"` // ± fraction of yearly randomness
	StudyUntil   int              `yaml:"study_until"`
	Modifiers    map[Risk]float64 `yaml:"modifiers"`
	PerYear      Effects          `yaml:"per_year"`
	Conditions   []string         `yaml:"conditions"`
	Flags        []string         `yaml:"flags"`
	Evidence     string           `yaml:"evidence"`
}
