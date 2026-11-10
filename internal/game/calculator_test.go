package game

import (
	"math"
	"testing"

	"github.com/funstorm15/survive85/data"
	"github.com/funstorm15/survive85/internal/actuary"
	"github.com/funstorm15/survive85/internal/events"
)

func baseline() map[events.Risk]float64 {
	m := map[events.Risk]float64{}
	for _, r := range events.Risks {
		m[r] = 1
	}
	return m
}

func loadBook(t *testing.T) *actuary.Book {
	t.Helper()
	b, err := actuary.Load(data.FS, data.LifeTables)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestClampModifier(t *testing.T) {
	for _, tc := range []struct{ in, want float64 }{{0.01, MinModifier}, {1, 1}, {3.5, 3.5}, {1000, MaxModifier}} {
		if got := ClampModifier(tc.in); got != tc.want {
			t.Errorf("ClampModifier(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestCombineModifier(t *testing.T) {
	for _, v := range []float64{0.6, 1, 1.5, 2} {
		if got := CombineModifier(v); got != v {
			t.Errorf("CombineModifier(%v) = %v: values near 1 must pass through", v, got)
		}
	}
	if got := CombineModifier(25); got >= 25 || got <= 2 {
		t.Errorf("CombineModifier(25) = %v: large products must be compressed", got)
	}
	if got := CombineModifier(0.1); got <= 0.1 || got >= 0.5 {
		t.Errorf("CombineModifier(0.1) = %v: small products must be compressed", got)
	}
	prev := 0.0
	for v := 0.05; v < 50; v *= 1.1 {
		got := CombineModifier(v)
		if got < prev {
			t.Fatalf("CombineModifier must be monotonic (v=%v)", v)
		}
		prev = got
	}
}

func TestHealthAndSanityFactors(t *testing.T) {
	if got := HealthFactor(80, 30); got != 1 {
		t.Errorf("HealthFactor(80, 30) = %v, want 1", got)
	}
	if got := HealthFactor(50, 30); math.Abs(got-2) > 1e-9 {
		t.Errorf("HealthFactor(50, 30) = %v, want 2", got)
	}
	if got := HealthFactor(60, 70); got != 1 {
		t.Errorf("at 70 the reference is 60 HP, got %v", got)
	}
	if HealthFactor(100, 70) != 0.6 || HealthFactor(0, 30) != 4 {
		t.Errorf("HealthFactor bounds wrong: %v %v", HealthFactor(100, 70), HealthFactor(0, 30))
	}
	if got := SelfHarmFactor(30); math.Abs(got-2) > 1e-9 {
		t.Errorf("SelfHarmFactor(30) = %v, want 2", got)
	}
	for h := 1; h < 100; h++ {
		if HealthFactor(h, 40) < HealthFactor(h+1, 40) {
			t.Fatalf("HealthFactor must not increase with health (h=%d)", h)
		}
	}
}

// At baseline modifiers, the expected health for the age and sanity 50 the
// adjusted probability must equal the life table exactly.
func TestAnnualMortalityBaselineMatchesLifeTable(t *testing.T) {
	b := loadBook(t)
	for _, tab := range b.Tables {
		for age := b.FirstAge; age <= b.LastAge; age++ {
			m := AnnualMortality(b, &tab, age, baseline(), ExpectedHealth(age), 50)
			if math.Abs(m.Total-m.BaseQ) > 1e-12 {
				t.Fatalf("%s age %d: total %v != q %v", tab.ID, age, m.Total, m.BaseQ)
			}
			share := 0.0
			for _, c := range m.Causes {
				share += c.Share
			}
			if math.Abs(share-1) > 1e-9 {
				t.Fatalf("%s age %d: shares sum to %v", tab.ID, age, share)
			}
		}
	}
}

// A hazard ratio on one cause must scale exactly that cause's hazard:
// doubling cardiovascular risk adds mu*f_cvd to the total hazard.
func TestAnnualMortalityProportionalHazards(t *testing.T) {
	b := loadBook(t)
	tab, _ := b.TableFor("rus", "male")
	age := 55
	h := ExpectedHealth(age)
	base := AnnualMortality(b, tab, age, baseline(), h, 50)
	mods := baseline()
	mods[events.RiskCardio] = 2
	got := AnnualMortality(b, tab, age, mods, h, 50)

	q := b.QAt(tab, age)
	mu := -math.Log(1 - q)
	f := b.Fraction(tab, age, actuary.Cardiovascular)
	want := 1 - math.Exp(-(mu + mu*f))
	if math.Abs(got.Total-want) > 1e-12 {
		t.Fatalf("total = %v, want %v", got.Total, want)
	}
	if got.Total <= base.Total {
		t.Fatalf("doubling CVD risk must raise mortality")
	}
	// Protective modifiers lower it.
	mods[events.RiskCardio] = 0.5
	if AnnualMortality(b, tab, age, mods, h, 50).Total >= base.Total {
		t.Fatalf("halving CVD risk must lower mortality")
	}
}

func TestPickCause(t *testing.T) {
	m := Mortality{Causes: []CauseRisk{{Cause: actuary.Oncology, Share: 0.25}, {Cause: actuary.Trauma, Share: 0.75}}}
	if m.PickCause(0.1) != actuary.Oncology || m.PickCause(0.5) != actuary.Trauma || m.PickCause(0.9999) != actuary.Trauma {
		t.Fatal("PickCause does not follow shares")
	}
}

func TestScaledChance(t *testing.T) {
	mods := baseline()
	mods[events.RiskLiver] = 3
	if got := ScaledChance(0.1, mods, []events.Risk{events.RiskLiver}); math.Abs(got-0.3) > 1e-12 {
		t.Errorf("ScaledChance = %v, want 0.3", got)
	}
	if got := ScaledChance(0.5, mods, []events.Risk{events.RiskLiver}); got != MaxInstantChance {
		t.Errorf("ScaledChance must cap at %v, got %v", MaxInstantChance, got)
	}
	if got := ScaledChance(0.2, mods, nil); got != 0.2 {
		t.Errorf("unscaled chance changed: %v", got)
	}
}

func TestArrestChance(t *testing.T) {
	if ArrestChance(1) != 0 || ArrestChance(0.5) != 0 {
		t.Error("a clean record must never be arrested")
	}
	if got := ArrestChance(2); math.Abs(got-0.012) > 1e-12 {
		t.Errorf("ArrestChance(2) = %v", got)
	}
	if ArrestChance(100) != MaxArrestChance {
		t.Error("arrest chance must be capped")
	}
}

func TestCohortDiedBefore(t *testing.T) {
	b := loadBook(t)
	tab, _ := b.TableFor("world", "female")
	if CohortDiedBefore(b, tab, StartAge) != 0 {
		t.Fatal("nobody dies before the start")
	}
	prev := 0.0
	for a := StartAge + 1; a <= VictoryAge; a++ {
		v := CohortDiedBefore(b, tab, a)
		if v <= prev || v >= 1 {
			t.Fatalf("cohort deaths must grow monotonically in (0,1): age %d = %v", a, v)
		}
		prev = v
	}
}

func TestLifeTablesArePlausible(t *testing.T) {
	b := loadBook(t)
	for _, tab := range b.Tables {
		s := b.Survival(&tab, StartAge, VictoryAge)
		if s < 0.05 || s > 0.8 {
			t.Errorf("%s: survival 18→85 = %.3f is implausible", tab.ID, s)
		}
		for i := 1; i < len(tab.EX); i++ {
			if tab.EX[i] >= tab.EX[i-1] {
				t.Errorf("%s: life expectancy must fall with age", tab.ID)
				break
			}
		}
	}
	rm, _ := b.TableFor("rus", "male")
	rf, _ := b.TableFor("rus", "female")
	if b.Survival(rm, 18, 85) >= b.Survival(rf, 18, 85) {
		t.Error("WHO data: Russian men survive to 85 less often than women")
	}
}
