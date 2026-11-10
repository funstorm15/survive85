package game

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/funstorm15/survive85/internal/actuary"
	"github.com/funstorm15/survive85/internal/events"
)

const testCauses = `
cardiovascular: {name: {ru: С, en: C}, reasons: [{ru: инфаркт, en: infarction}]}
oncology: {name: {ru: О, en: O}, reasons: [{ru: рак, en: cancer}]}
liver: {name: {ru: П, en: L}, reasons: [{ru: цирроз, en: cirrhosis}]}
trauma: {name: {ru: Т, en: T}, reasons: [{ru: ДТП, en: crash}]}
self_harm: {name: {ru: С, en: S}, reasons: [{ru: суицид, en: suicide}]}
other: {name: {ru: Пр, en: Ot}, reasons: [{ru: пневмония, en: pneumonia}]}
`

const testProfessions = `
- id: tester
  name: {ru: Тестировщик, en: Tester}
  description: {ru: тест, en: test}
  start_wealth: 1000
  income: 100000
`

const testConditions = `
- id: in_prison
  name: {ru: Тюрьма, en: Prison}
  description: {ru: т, en: p}
  kind: bad
  suppress_income: true
  forces_event: prison_year
- id: criminal_record
  name: {ru: Судимость, en: Record}
  description: {ru: с, en: r}
  kind: bad
- id: smoker
  name: {ru: Курильщик, en: Smoker}
  description: {ru: к, en: s}
  kind: bad
  modifiers: {oncology_risk: 2.5}
  death_reasons: {oncology: {ru: Рак лёгкого, en: Lung cancer}}
`

const testEvidence = `
- id: ch34_i3
  chapter: 34
  item: 3
  chapter_title: {ru: Лекарства, en: Medicines}
  title: {ru: Парацетамол, en: Paracetamol}
  grade: A
  measure: mortality
  plain: {ru: п, en: p}
  benefit: {ru: э, en: b}
  sources: [{citation: "FDA", url: "https://www.fda.gov"}]
`

const testEvents = `
- id: prison_year
  title: {ru: Тюрьма, en: Prison}
  min_age: 18
  max_age: 84
  category: legal
  repeatable: true
  narrative: {ru: н, en: n}
  choices:
    - {id: endure, text: {ru: т, en: t}, tag: neutral, outcome: {ru: о, en: o}}
- id: filler
  title: {ru: Год, en: Year}
  min_age: 18
  max_age: 84
  category: life
  repeatable: true
  cooldown: 1
  narrative: {ru: н, en: n}
  choices:
    - {id: a, text: {ru: а, en: a}, tag: wise, outcome: {ru: о, en: o}}
    - {id: b, text: {ru: б, en: b}, tag: risky, outcome: {ru: о, en: o}, cost_will: 10}
- id: poison
  title: {ru: Яд, en: Poison}
  min_age: 18
  max_age: 84
  category: pharmacy
  weight: 1000
  narrative: {ru: н, en: n}
  choices:
    - id: drink
      text: {ru: пить, en: drink}
      tag: risky
      stat_effects: {toxicology_liver_risk: 3.5, oncology_risk: {factor: 2, years: 2}}
      add_conditions: [smoker]
      outcome: {ru: о, en: o}
      instant_death_check:
        chance: 1
        cause: liver
        death_reason: {ru: Печень, en: Liver}
        evidence: ch34_i3
    - id: crime
      text: {ru: украсть, en: steal}
      tag: risky
      cost_money: 500
      outcome: {ru: о, en: o}
      catastrophe_check: {chance: 1, kind: prison, years: 2, delta_sanity: -10, text: {ru: т, en: t}}
    - id: rich
      text: {ru: богатство, en: riches}
      tag: neutral
      cost_money: 1000000
      outcome: {ru: о, en: o}
`

func testDB(t *testing.T) *events.DB {
	t.Helper()
	fsys := fstest.MapFS{
		"causes.yaml":          {Data: []byte(testCauses)},
		"professions.yaml":     {Data: []byte(testProfessions)},
		"conditions/core.yaml": {Data: []byte(testConditions)},
		"evidence/ch34.yaml":   {Data: []byte(testEvidence)},
		"events/test.yaml":     {Data: []byte(testEvents)},
	}
	db, err := events.Load(fsys)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func newTestGame(t *testing.T, seed uint64) *Game {
	t.Helper()
	g, err := New(testDB(t), loadBook(t), Options{Region: "world", Sex: "female", Profession: "tester", Seed: seed})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func choiceIndex(t *testing.T, g *Game, id string) int {
	t.Helper()
	for i, c := range g.Current.Choices {
		if c.ID == id {
			return i
		}
	}
	t.Fatalf("event %s has no choice %s", g.Current.ID, id)
	return -1
}

func TestNewGameStartsAt18WithHeavyEvent(t *testing.T) {
	g := newTestGame(t, 1)
	if g.Player.Age != StartAge || g.Phase != PhaseEvent {
		t.Fatalf("age %d phase %d", g.Player.Age, g.Phase)
	}
	if g.Current == nil || g.Current.ID != "poison" {
		t.Fatalf("weight 1000 event should be drawn first, got %v", g.Current)
	}
}

func TestInstantDeathRecordsEvidenceAndFactors(t *testing.T) {
	g := newTestGame(t, 1)
	o, err := g.Choose(choiceIndex(t, g, "drink"))
	if err != nil {
		t.Fatal(err)
	}
	if g.Phase != PhaseDead || g.Death == nil || !g.Death.Instant {
		t.Fatalf("chance 1 must kill: phase=%d", g.Phase)
	}
	d := g.Death
	if d.Cause != actuary.Liver || d.Evidence != "ch34_i3" || d.Reason.Get("en") != "Liver" || d.Age != 18 {
		t.Fatalf("bad death record: %+v", d)
	}
	if d.YLL <= 50 || d.CohortDiedBefore != 0 {
		t.Fatalf("YLL %v cohort %v", d.YLL, d.CohortDiedBefore)
	}
	if o.Death == nil || !o.Death.Triggered || len(o.Factors) != 2 {
		t.Fatalf("outcome must report the check and both factors: %+v", o)
	}
	mods := g.Modifiers()
	if mods[events.RiskLiver] != CombineModifier(3.5) || mods[events.RiskOnco] != CombineModifier(2*2.5) {
		t.Fatalf("modifiers = %v", mods)
	}
	if _, err := g.Choose(0); err == nil {
		t.Fatal("dead players cannot choose")
	}
}

func TestTemporaryFactorExpires(t *testing.T) {
	g := newTestGame(t, 1)
	g.Player.Factors = append(g.Player.Factors, RiskFactor{Risk: events.RiskOnco, Factor: 2, Age: 18, Until: 19})
	if g.Modifiers()[events.RiskOnco] != 2 {
		t.Fatal("factor must apply at 18")
	}
	g.Player.Age = 20
	if g.Modifiers()[events.RiskOnco] != 1 {
		t.Fatal("factor must expire after 19")
	}
}

func TestPrisonForcesEventsAndSuppressesIncome(t *testing.T) {
	g := newTestGame(t, 3)
	if _, err := g.Choose(choiceIndex(t, g, "crime")); err != nil {
		t.Fatal(err)
	}
	if g.Phase == PhaseDead {
		t.Skip("died in the annual check; seed-dependent")
	}
	if !g.Player.HasCondition(PrisonCondition) || !g.Player.HasCondition(RecordCondition) || g.Player.Arrests != 1 {
		t.Fatalf("prison and record expected: %+v", g.Player.Conditions)
	}
	for year := 0; year < 2; year++ {
		g.Next()
		if g.Current == nil || g.Current.ID != "prison_year" {
			t.Fatalf("year %d in prison: got %v", year, g.Current)
		}
		w := g.Player.Wealth
		o, err := g.Choose(0)
		if err != nil {
			t.Fatal(err)
		}
		if g.Phase == PhaseDead {
			t.Skip("died in prison; seed-dependent")
		}
		if o.Year.Income != 0 || g.Player.Wealth < w-1 && g.Player.Wealth >= 0 {
			t.Fatalf("no income in prison: %+v", o.Year)
		}
	}
	if g.Player.HasCondition(PrisonCondition) || g.Player.YearsInPrison != 2 {
		t.Fatalf("a 2-year term must end after 2 years: conds=%v years=%d", g.Player.Conditions, g.Player.YearsInPrison)
	}
	g.Next()
	if g.Current != nil && g.Current.ID == "prison_year" {
		t.Fatal("free players are not forced into prison events")
	}
}

func TestUnaffordableChoiceIsRejected(t *testing.T) {
	g := newTestGame(t, 1)
	if _, err := g.Choose(choiceIndex(t, g, "rich")); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("want ErrUnavailable, got %v", err)
	}
	if g.Phase != PhaseEvent || len(g.Player.Decisions) != 0 {
		t.Fatal("a rejected choice must not change state")
	}
}

// The same seed and choices must give the same life.
func TestDeterminism(t *testing.T) {
	play := func() (int, actuary.Cause, int64) {
		g := newTestGame(t, 42)
		g.Player.Seen["poison"] = 1 // skip the deadly event
		g.startYear()
		for g.Phase != PhaseDead && g.Phase != PhaseVictory {
			if g.Phase == PhaseOutcome {
				g.Next()
				continue
			}
			if _, err := g.Choose(0); err != nil {
				t.Fatal(err)
			}
		}
		cause := actuary.Cause("")
		if g.Death != nil {
			cause = g.Death.Cause
		}
		return g.Player.AgeInt(), cause, g.Player.Wealth
	}
	a1, c1, w1 := play()
	a2, c2, w2 := play()
	if a1 != a2 || c1 != c2 || w1 != w2 {
		t.Fatalf("runs differ: %d/%s/%d vs %d/%s/%d", a1, c1, w1, a2, c2, w2)
	}
}

// Over many seeds, a careful filler-only life must reach 85 about as often
// as the life table says (health and sanity drift around the reference).
func TestVictoryRateTracksLifeTable(t *testing.T) {
	if testing.Short() {
		t.Skip("statistical")
	}
	db := testDB(t)
	book := loadBook(t)
	tab, _ := book.TableFor("world", "female")
	baseline := book.Survival(tab, StartAge, VictoryAge)
	wins, runs := 0, 3000
	for s := 0; s < runs; s++ {
		g, _ := New(db, book, Options{Region: "world", Sex: "female", Profession: "tester", Seed: uint64(s)})
		g.Player.Seen["poison"] = 1
		g.startYear()
		for g.Phase != PhaseDead && g.Phase != PhaseVictory {
			if g.Phase == PhaseOutcome {
				g.Next()
				continue
			}
			g.Choose(0)
		}
		if g.Phase == PhaseVictory {
			wins++
		}
	}
	rate := float64(wins) / float64(runs)
	if rate < baseline*0.6 || rate > baseline*1.6 {
		t.Fatalf("victory rate %.3f too far from life-table baseline %.3f", rate, baseline)
	}
	t.Logf("victory rate %.3f vs baseline %.3f", rate, baseline)
}
