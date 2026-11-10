package game

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/funstorm15/survive85/internal/actuary"
	"github.com/funstorm15/survive85/internal/events"
	"github.com/funstorm15/survive85/internal/i18n"
)

// PrisonCondition is the condition id the engine applies on arrest; the data
// must define it (it normally forces a "year in prison" event).
const PrisonCondition = "in_prison"

// RecordCondition is added together with a prison term, if the data defines it.
const RecordCondition = "criminal_record"

// Economy rules (game conventions, not data from the book).
const (
	DebtInterest     = 0.20 // yearly interest on negative wealth
	DebtSanityCost   = 5    // sanity lost per year in debt
	ArrestPrisonTerm = 3    // years served after an arrest by the annual legal check
	ArrestWealthLoss = 0.3  // share of positive wealth lost on arrest
	ArrestSanityCost = 25
	PensionAge       = 65
)

// Phase is where the run currently is.
type Phase int

const (
	PhaseEvent   Phase = iota // waiting for a choice
	PhaseOutcome              // showing the result of the year
	PhaseDead
	PhaseVictory
)

// Options configure a new run.
type Options struct {
	Region     string
	Sex        string
	Profession string
	Seed       uint64
}

// Check is the result of a random check.
type Check struct {
	Chance    float64
	Triggered bool
}

// Outcome is everything that happened after a choice, including the end of
// the year.
type Outcome struct {
	Age          int
	Event        *events.Event
	Choice       *events.Choice
	Deltas       Deltas
	Factors      []RiskFactor
	Added        []string
	Removed      []string
	Death        *Check // instant death check, if any
	Disaster     *Check // catastrophe check, if any
	DisasterKind events.CatastropheKind
	Year         *YearEnd
}

// Deltas are visible stat changes.
type Deltas struct {
	Health int
	Sanity int
	Wealth int64
	Will   int
}

func (d *Deltas) add(o Deltas) {
	d.Health += o.Health
	d.Sanity += o.Sanity
	d.Wealth += o.Wealth
	d.Will += o.Will
}

// YearEnd is the annual report: economy, upkeep and the mortality check.
type YearEnd struct {
	Age       int
	Income    int64
	Interest  int64
	Upkeep    Deltas // conditions, profession and ageing
	Mortality Mortality
	Survived  bool
	Arrest    Check
	Expired   []string
}

// Death describes how a run ended.
type Death struct {
	Age      int
	Cause    actuary.Cause
	Reason   i18n.Text
	Instant  bool // died on the spot after a choice
	Collapse bool // Health reached zero
	EventID  string
	ChoiceID string
	Evidence string
	// Chance is the probability of the check that killed the player.
	Chance float64
	// Factor is the decisive risk factor, if one raised this cause's risk.
	Factor    *Contribution
	Mortality *Mortality
	// YLL is the remaining life expectancy at the age of death per the life
	// table: the years of life lost.
	YLL float64
	// CohortDiedBefore is the share of peers who would have died earlier.
	CohortDiedBefore float64
}

// Game is one run.
type Game struct {
	DB    *events.DB
	Book  *actuary.Book
	Table *actuary.Table
	Prof  *events.Profession

	Player  *Player
	Seed    uint64
	Phase   Phase
	Current *events.Event
	Last    *Outcome
	Death   *Death

	rng          *rand.Rand
	lastCategory events.Category
}

// ErrUnavailable is returned when a choice cannot be taken.
var ErrUnavailable = errors.New("choice unavailable")

// New starts a run on the 18th birthday and draws the first event.
func New(db *events.DB, book *actuary.Book, o Options) (*Game, error) {
	t, ok := book.TableFor(o.Region, o.Sex)
	if !ok {
		return nil, fmt.Errorf("no life table for region %q sex %q", o.Region, o.Sex)
	}
	prof, ok := db.Profession(o.Profession)
	if !ok {
		return nil, fmt.Errorf("unknown profession %q", o.Profession)
	}
	g := &Game{
		DB: db, Book: book, Table: t, Prof: prof,
		Player: NewPlayer(prof, o.Region, o.Sex),
		Seed:   o.Seed,
		rng:    rand.New(rand.NewPCG(o.Seed, o.Seed^0x9e3779b97f4a7c15)),
	}
	for r, m := range prof.Modifiers {
		g.Player.Factors = append(g.Player.Factors, RiskFactor{
			Risk: r, Factor: m, Age: StartAge, Label: prof.Name, Evidence: prof.Evidence,
		})
	}
	for _, c := range prof.Conditions {
		g.addCondition(c, nil)
	}
	g.startYear()
	return g, nil
}

// Conditions returns the condition definitions.
func (g *Game) Conditions() map[string]*events.Condition { return g.DB.Conditions }

// Modifiers returns the player's current hidden risk modifiers.
func (g *Game) Modifiers() map[events.Risk]float64 {
	return g.Player.HiddenModifiers(g.DB.Conditions)
}

// State returns the requirements view of the player.
func (g *Game) State() events.State { return view{g, g.Modifiers()} }

// Mortality is the annual mortality the player faces at the current age.
func (g *Game) Mortality() Mortality {
	p := g.Player
	return AnnualMortality(g.Book, g.Table, p.AgeInt(), g.Modifiers(), p.Health, p.Sanity)
}

// VisibleChoices returns the indexes of the current event's choices the
// player may see.
func (g *Game) VisibleChoices() []int {
	if g.Current == nil {
		return nil
	}
	s := g.State()
	var out []int
	for i := range g.Current.Choices {
		if g.Current.Choices[i].Visible(s) {
			out = append(out, i)
		}
	}
	return out
}

func (g *Game) startYear() {
	g.Phase = PhaseEvent
	forced := ""
	for _, c := range g.Player.Conditions {
		if def := g.DB.Conditions[c.ID]; def != nil && def.ForcesEvent != "" {
			forced = def.ForcesEvent
			break
		}
	}
	g.Current = g.DB.Pick(g.State(), forced, g.rng.Float64())
	if g.Current != nil {
		g.Player.Seen[g.Current.ID]++
		g.Player.LastSeen[g.Current.ID] = g.Player.AgeInt()
		g.lastCategory = g.Current.Category
	}
}

// Choose answers the current event with the choice at index i (an index
// into Current.Choices). With no current event (a quiet year) i is ignored.
func (g *Game) Choose(i int) (*Outcome, error) {
	if g.Phase != PhaseEvent {
		return nil, fmt.Errorf("not waiting for a choice")
	}
	p := g.Player
	o := &Outcome{Age: p.AgeInt(), Event: g.Current}
	if g.Current != nil {
		if i < 0 || i >= len(g.Current.Choices) {
			return nil, fmt.Errorf("choice %d out of range", i)
		}
		ch := &g.Current.Choices[i]
		if ch.Availability(g.State()) != events.Available {
			return nil, ErrUnavailable
		}
		o.Choice = ch
		g.applyChoice(o, ch)
		if g.Phase == PhaseDead {
			g.Last = o
			return o, nil
		}
	}
	o.Year = g.endYear(o)
	g.Last = o
	return o, nil
}

func (g *Game) applyChoice(o *Outcome, ch *events.Choice) {
	p := g.Player
	ev := g.Current
	age := p.AgeInt()
	p.Will -= ch.CostWill
	p.Wealth -= ch.CostMoney
	p.MoneySpent += ch.CostMoney
	p.HoursInvested += ch.CostTime
	o.Deltas.add(Deltas{Wealth: -ch.CostMoney, Will: -ch.CostWill})
	p.Decisions = append(p.Decisions, Decision{Age: age, EventID: ev.ID, ChoiceID: ch.ID, Tag: ch.Tag, Evidence: ch.Evidence})

	label := ev.Title
	g.applyEffects(o, ch.Effects, label, ev.ID, ch.ID, ch.Evidence, ch.DeathReasons)

	if d := ch.InstantDeathCheck; d != nil {
		chance := ScaledChance(d.Chance, g.Modifiers(), d.ScaleBy)
		hit := g.rng.Float64() < chance
		o.Death = &Check{Chance: chance, Triggered: hit}
		if hit {
			g.die(&Death{Age: age, Cause: d.Cause, Reason: d.DeathReason, Instant: true,
				EventID: ev.ID, ChoiceID: ch.ID, Evidence: d.Evidence, Chance: chance})
			return
		}
	}
	if k := ch.CatastropheCheck; k != nil {
		chance := ScaledChance(k.Chance, g.Modifiers(), k.ScaleBy)
		hit := g.rng.Float64() < chance
		o.Disaster = &Check{Chance: chance, Triggered: hit}
		o.DisasterKind = k.Kind
		if hit {
			if k.WealthFraction > 0 && p.Wealth > 0 {
				loss := int64(math.Round(float64(p.Wealth) * k.WealthFraction))
				p.Wealth -= loss
				o.Deltas.Wealth -= loss
			}
			g.applyEffects(o, k.Effects, label, ev.ID, ch.ID, k.Evidence, ch.DeathReasons)
			if k.Kind == events.KindPrison {
				g.imprison(o, k.Years)
			}
		}
	}
	if p.Health <= 0 {
		g.collapse(ev.ID, ch.ID, ch.Evidence)
	}
}

// applyEffects applies an Effects block and records risk factors.
func (g *Game) applyEffects(o *Outcome, e events.Effects, label i18n.Text, eventID, choiceID, evidence string, reasons map[actuary.Cause]i18n.Text) {
	p := g.Player
	age := p.AgeInt()
	before := Deltas{Health: p.Health, Sanity: p.Sanity, Wealth: p.Wealth, Will: p.Will}
	p.Health = clamp(p.Health+e.DeltaHealth, 0, MaxStat)
	p.Sanity = clamp(p.Sanity+e.DeltaSanity, 0, MaxStat)
	p.Wealth += e.DeltaWealth
	p.Will = clamp(p.Will+e.DeltaWill, 0, MaxWill)
	o.Deltas.add(Deltas{Health: p.Health - before.Health, Sanity: p.Sanity - before.Sanity,
		Wealth: p.Wealth - before.Wealth, Will: p.Will - before.Will})
	if p.Wealth > p.PeakWealth {
		p.PeakWealth = p.Wealth
	}
	for _, r := range events.Risks { // stable order
		re, ok := e.StatEffects[r]
		if !ok {
			continue
		}
		f := RiskFactor{Risk: r, Factor: re.Factor, Age: age, EventID: eventID, ChoiceID: choiceID,
			Evidence: evidence, Label: label, DeathReasons: reasons}
		if re.Years > 0 {
			f.Until = age + re.Years - 1
		}
		p.Factors = append(p.Factors, f)
		o.Factors = append(o.Factors, f)
	}
	for _, id := range e.RemoveConditions {
		if g.removeCondition(id) {
			o.Removed = append(o.Removed, id)
		}
	}
	for _, id := range e.AddConditions {
		if g.addCondition(id, nil) {
			o.Added = append(o.Added, id)
		}
	}
	for _, f := range e.SetFlags {
		if _, ok := p.Flags[f]; !ok {
			p.Flags[f] = age
		}
	}
	for _, f := range e.ClearFlags {
		delete(p.Flags, f)
	}
}

func (g *Game) addCondition(id string, years *int) bool {
	p := g.Player
	def := g.DB.Conditions[id]
	if def == nil {
		return false
	}
	n := def.Years
	if years != nil {
		n = *years
	}
	until := 0
	if n > 0 {
		until = p.AgeInt() + n - 1
	}
	for i := range p.Conditions {
		if p.Conditions[i].ID == id {
			// Re-acquiring refreshes the duration.
			if until == 0 || (p.Conditions[i].Until != 0 && until > p.Conditions[i].Until) {
				p.Conditions[i].Until = until
			}
			return false
		}
	}
	p.Conditions = append(p.Conditions, ActiveCondition{ID: id, Since: p.AgeInt(), Until: until})
	return true
}

func (g *Game) removeCondition(id string) bool {
	p := g.Player
	for i, c := range p.Conditions {
		if c.ID == id {
			p.Conditions = append(p.Conditions[:i], p.Conditions[i+1:]...)
			return true
		}
	}
	return false
}

// imprison starts a prison term. The term begins next year: the current
// year is the one in which the arrest happened.
func (g *Game) imprison(o *Outcome, years int) {
	p := g.Player
	p.Arrests++
	until := p.AgeInt() + years
	def := g.DB.Conditions[PrisonCondition]
	if def == nil {
		return
	}
	for i := range p.Conditions {
		if p.Conditions[i].ID == PrisonCondition {
			p.Conditions[i].Until = max(p.Conditions[i].Until, until)
			return
		}
	}
	p.Conditions = append(p.Conditions, ActiveCondition{ID: PrisonCondition, Since: p.AgeInt() + 1, Until: until})
	if o != nil {
		o.Added = append(o.Added, PrisonCondition)
	}
	if g.addCondition(RecordCondition, nil) && o != nil {
		o.Added = append(o.Added, RecordCondition)
	}
	// Serving the sentence settles the legal risks taken so far.
	for i := range p.Factors {
		if p.Factors[i].Risk == events.RiskLegal && p.Factors[i].Factor > 1 && p.Factors[i].Until == 0 {
			p.Factors[i].Until = p.AgeInt()
		}
	}
}

// endYear runs the yearly upkeep and the annual mortality check for the
// current age, then ages the player by one year.
func (g *Game) endYear(o *Outcome) *YearEnd {
	p := g.Player
	age := p.AgeInt()
	y := &YearEnd{Age: age}
	before := Deltas{Health: p.Health, Sanity: p.Sanity, Wealth: p.Wealth, Will: p.Will}

	// Upkeep from conditions and the profession.
	upkeep := func(e events.Effects) {
		p.Health = clamp(p.Health+e.DeltaHealth, 0, MaxStat)
		p.Sanity = clamp(p.Sanity+e.DeltaSanity, 0, MaxStat)
		p.Wealth += e.DeltaWealth
		p.Will = clamp(p.Will+e.DeltaWill, 0, MaxWill)
	}
	suppress := false
	inPrison := false
	for _, c := range p.Conditions {
		if def := g.DB.Conditions[c.ID]; def != nil && c.Since <= age {
			upkeep(def.PerYear)
			suppress = suppress || def.SuppressIncome
			inPrison = inPrison || c.ID == PrisonCondition
		}
	}
	upkeep(g.Prof.PerYear)
	if inPrison {
		p.YearsInPrison++
	}

	// Economy.
	if !suppress {
		y.Income = g.income(age)
		p.Wealth += y.Income
	}
	if p.Wealth < 0 {
		y.Interest = int64(math.Round(float64(-p.Wealth) * DebtInterest))
		p.Wealth -= y.Interest
		p.Sanity = clamp(p.Sanity-DebtSanityCost, 0, MaxStat)
	}
	if p.Wealth > p.PeakWealth {
		p.PeakWealth = p.Wealth
	}

	// Ageing and recovery.
	p.Health = clamp(p.Health+naturalHealth(age, p.Health), 0, MaxStat)
	p.Sanity = clamp(p.Sanity+int(math.Round(float64(65-p.Sanity)*0.15)), 0, MaxStat)
	p.Will = clamp(p.Will+1+p.Sanity/40, 0, MaxWill)
	y.Upkeep = Deltas{Health: p.Health - before.Health, Sanity: p.Sanity - before.Sanity,
		Wealth: p.Wealth - before.Wealth - y.Income + y.Interest, Will: p.Will - before.Will}

	if p.Health <= 0 {
		g.collapse("", "", "")
		y.Mortality = g.Mortality()
		return y
	}

	// Past legal risks may catch up.
	mods := g.Modifiers()
	if !inPrison {
		// The profession's own legal exposure scales catastrophe checks but
		// never gets anyone arrested by itself: only decisions do.
		base := 1.0
		if m, ok := g.Prof.Modifiers[events.RiskLegal]; ok && m > 0 {
			base = m
		}
		y.Arrest.Chance = ArrestChance(mods[events.RiskLegal] / base)
		if y.Arrest.Chance > 0 && g.rng.Float64() < y.Arrest.Chance {
			y.Arrest.Triggered = true
			if p.Wealth > 0 {
				p.Wealth -= int64(math.Round(float64(p.Wealth) * ArrestWealthLoss))
			}
			p.Sanity = clamp(p.Sanity-ArrestSanityCost, 0, MaxStat)
			g.imprison(o, ArrestPrisonTerm)
		}
	}

	// Annual mortality check.
	y.Mortality = AnnualMortality(g.Book, g.Table, age, mods, p.Health, p.Sanity)
	if g.rng.Float64() < y.Mortality.Total {
		cause := y.Mortality.PickCause(g.rng.Float64())
		m := y.Mortality
		g.die(&Death{Age: age, Cause: cause, Chance: m.Total, Mortality: &m})
		return y
	}
	y.Survived = true

	// Expire conditions that ran their course this year.
	kept := p.Conditions[:0]
	for _, c := range p.Conditions {
		if c.Until != 0 && c.Until <= age {
			y.Expired = append(y.Expired, c.ID)
			continue
		}
		kept = append(kept, c)
	}
	p.Conditions = kept

	p.Age++
	if p.AgeInt() >= VictoryAge {
		g.Phase = PhaseVictory
		g.Current = nil
	} else {
		g.Phase = PhaseOutcome
	}
	return y
}

// Next moves from the outcome screen to the next year's event.
func (g *Game) Next() {
	if g.Phase == PhaseOutcome {
		g.startYear()
	}
}

// income is the money left after living costs in a year of life.
func (g *Game) income(age int) int64 {
	pr := g.Prof
	if age < pr.StudyUntil {
		return 0
	}
	var curve float64
	switch {
	case age < 25:
		curve = 0.7
	case age < 35:
		curve = 1.0
	case age < 55:
		curve = 1.3
	case age < PensionAge:
		curve = 1.1
	default:
		curve = 0.4 // pension
	}
	jitter := 1 + pr.IncomeSpread*(2*g.rng.Float64()-1)
	return int64(math.Round(float64(pr.Income) * curve * jitter))
}

// naturalHealth is the yearly drift of Health with age: the young recover a
// share of what they have lost, the old decline slowly.
func naturalHealth(age, health int) int {
	missing := float64(MaxStat - health)
	switch {
	case age < 30:
		return int(math.Round(missing * 0.15))
	case age < 50:
		return int(math.Round(missing * 0.10))
	case age < 65:
		return int(math.Round(missing * 0.05))
	default:
		return -1
	}
}

// collapse kills a player whose Health reached zero. The cause is the
// largest share of their current mortality.
func (g *Game) collapse(eventID, choiceID, evidence string) {
	m := g.Mortality()
	cause := actuary.Other
	if sc := m.SortedCauses(); len(sc) > 0 {
		cause = sc[0].Cause
	}
	g.die(&Death{Age: g.Player.AgeInt(), Cause: cause, Collapse: true, EventID: eventID,
		ChoiceID: choiceID, Evidence: evidence, Chance: 1, Mortality: &m})
}

func (g *Game) die(d *Death) {
	p := g.Player
	if c, ok := DecisiveFactor(p, g.DB.Conditions, d.Cause); ok && !d.Instant {
		d.Factor = &c
		if d.Evidence == "" {
			d.Evidence = c.Evidence
		}
		if d.EventID == "" {
			d.EventID, d.ChoiceID = c.EventID, c.ChoiceID
		}
		if r, ok := c.DeathReasons[d.Cause]; ok && d.Reason.IsZero() {
			d.Reason = r
		}
	}
	if d.Reason.IsZero() {
		if info := g.DB.Causes[d.Cause]; info != nil && len(info.Reasons) > 0 {
			d.Reason = info.Reasons[g.rng.IntN(len(info.Reasons))]
		}
	}
	d.YLL = g.Book.EXAt(g.Table, d.Age)
	d.CohortDiedBefore = CohortDiedBefore(g.Book, g.Table, d.Age)
	g.Death = d
	g.Phase = PhaseDead
}

// Score rates the run. Every year lived is worth 100 points; reaching 85
// adds 2000; health, sanity, wise decisions and savings add the rest.
func (g *Game) Score() int {
	p := g.Player
	tags := p.TagCounts()
	s := (p.AgeInt()-StartAge)*100 + p.Health + p.Sanity + tags[events.TagWise]*20 - tags[events.TagRisky]*10
	if g.Phase == PhaseVictory {
		s += 2000
	}
	if p.Wealth > 0 {
		s += int(min(p.Wealth/50_000, 500))
	}
	return max(s, 0)
}

// view adapts a Game to events.State.
type view struct {
	g    *Game
	mods map[events.Risk]float64
}

func (v view) Age() int                    { return v.g.Player.AgeInt() }
func (v view) Health() int                 { return v.g.Player.Health }
func (v view) Sanity() int                 { return v.g.Player.Sanity }
func (v view) Wealth() int64               { return v.g.Player.Wealth }
func (v view) Will() int                   { return v.g.Player.Will }
func (v view) Profession() string          { return v.g.Player.Profession }
func (v view) Sex() string                 { return v.g.Player.Sex }
func (v view) HasCondition(id string) bool { return v.g.Player.HasCondition(id) }
func (v view) Modifier(r events.Risk) float64 {
	if m, ok := v.mods[r]; ok {
		return m
	}
	return 1
}
func (v view) FlagAge(f string) (int, bool) {
	a, ok := v.g.Player.Flags[f]
	return a, ok
}
func (v view) TimesSeen(id string) int { return v.g.Player.Seen[id] }
func (v view) LastSeen(id string) (int, bool) {
	a, ok := v.g.Player.LastSeen[id]
	return a, ok
}
func (v view) LastCategory() events.Category { return v.g.lastCategory }
