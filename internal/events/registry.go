package events

import "math"

// State is the read-only view of a player that requirements and event
// selection need. The game package implements it.
type State interface {
	Age() int
	Health() int
	Sanity() int
	Wealth() int64
	Will() int
	Profession() string
	Sex() string
	FlagAge(flag string) (age int, ok bool)
	HasCondition(id string) bool
	Modifier(r Risk) float64
	TimesSeen(eventID string) int
	LastSeen(eventID string) (age int, ok bool)
	LastCategory() Category
}

// Met reports whether s satisfies the requirements.
func (r Requirements) Met(s State) bool {
	for _, f := range r.Flags {
		if _, ok := s.FlagAge(f); !ok {
			return false
		}
	}
	for _, f := range r.NotFlags {
		if _, ok := s.FlagAge(f); ok {
			return false
		}
	}
	for f, years := range r.FlagYears {
		at, ok := s.FlagAge(f)
		if !ok || s.Age()-at < years {
			return false
		}
	}
	for _, c := range r.Conditions {
		if !s.HasCondition(c) {
			return false
		}
	}
	for _, c := range r.NotConditions {
		if s.HasCondition(c) {
			return false
		}
	}
	if r.MinWealth != nil && s.Wealth() < *r.MinWealth {
		return false
	}
	if r.MaxWealth != nil && s.Wealth() > *r.MaxWealth {
		return false
	}
	if r.MinHealth != nil && s.Health() < *r.MinHealth {
		return false
	}
	if r.MaxHealth != nil && s.Health() > *r.MaxHealth {
		return false
	}
	if r.MinSanity != nil && s.Sanity() < *r.MinSanity {
		return false
	}
	if r.MaxSanity != nil && s.Sanity() > *r.MaxSanity {
		return false
	}
	for risk, min := range r.MinRisk {
		if s.Modifier(risk) < min {
			return false
		}
	}
	if len(r.Classes) > 0 && !contains(r.Classes, s.Profession()) {
		return false
	}
	if r.Sex != "" && r.Sex != s.Sex() {
		return false
	}
	return true
}

// Unavailable explains why a choice cannot be taken.
type Unavailable int

const (
	Available Unavailable = iota
	NoMoney
	NoWill
	NotEligible
)

// Availability reports whether s can take the choice.
func (c *Choice) Availability(s State) Unavailable {
	switch {
	case !c.Requires.Met(s):
		return NotEligible
	case c.CostMoney > 0 && s.Wealth() < c.CostMoney:
		return NoMoney
	case c.CostWill > s.Will():
		return NoWill
	}
	return Available
}

// Visible reports whether a choice should be listed at all. Choices whose
// requirements are unmet are hidden; unaffordable ones are shown disabled.
func (c *Choice) Visible(s State) bool { return c.Requires.Met(s) }

// Eligible reports whether the event may be offered to s this year.
func (ev *Event) Eligible(s State) bool {
	age := s.Age()
	if age < ev.MinAge || age > ev.MaxAge {
		return false
	}
	if n := s.TimesSeen(ev.ID); n > 0 {
		if !ev.Repeatable {
			return false
		}
		if last, ok := s.LastSeen(ev.ID); ok && age-last < max(ev.Cooldown, 1) {
			return false
		}
	}
	if !ev.Requires.Met(s) {
		return false
	}
	// At least one choice must be affordable, otherwise the event is a dead end.
	for i := range ev.Choices {
		if ev.Choices[i].Availability(s) == Available {
			return true
		}
	}
	return false
}

// Weight is the relative chance of the event being picked for s. Risk-driven
// events (a heart attack) become likelier as the matching modifier grows;
// categories that just appeared are damped for variety; events already seen
// (repeatable ones) are damped so fresh situations come first.
func (ev *Event) WeightFor(s State) float64 {
	w := ev.Weight
	for r, exp := range ev.WeightBy {
		w *= math.Pow(s.Modifier(r), exp)
	}
	if ev.Category == s.LastCategory() {
		w *= 0.5
	}
	if n := s.TimesSeen(ev.ID); n > 0 {
		w /= float64(1 + n)
	}
	return w
}

// Pick chooses this year's event. A forced event (from a condition such as a
// prison term) wins; otherwise one eligible event is drawn by weight. roll
// must be uniform in [0, 1). It returns nil when nothing is eligible.
func (db *DB) Pick(s State, forced string, roll float64) *Event {
	if forced != "" {
		if ev, ok := db.byID[forced]; ok {
			return ev
		}
	}
	var pool []*Event
	var weights []float64
	total := 0.0
	for _, ev := range db.Events {
		if db.isForcedTarget(ev.ID) || !ev.Eligible(s) {
			continue
		}
		w := ev.WeightFor(s)
		if w <= 0 {
			continue
		}
		pool = append(pool, ev)
		weights = append(weights, w)
		total += w
	}
	if len(pool) == 0 {
		return nil
	}
	x := roll * total
	for i, w := range weights {
		if x < w {
			return pool[i]
		}
		x -= w
	}
	return pool[len(pool)-1]
}

// EligibleCount is the number of events s could be offered now (for tests
// and the content-coverage report).
func (db *DB) EligibleCount(s State) int {
	n := 0
	for _, ev := range db.Events {
		if !db.isForcedTarget(ev.ID) && ev.Eligible(s) {
			n++
		}
	}
	return n
}

func (db *DB) isForcedTarget(id string) bool {
	for _, c := range db.Conditions {
		if c.ForcesEvent == id {
			return true
		}
	}
	return false
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
