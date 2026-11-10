// Package sim plays many headless runs with a fixed strategy. It backs the
// `survive85 sim` balance report and the engine's statistical tests.
package sim

import (
	"fmt"
	"io"
	"math/rand/v2"
	"sort"

	"github.com/funstorm15/survive85/internal/actuary"
	"github.com/funstorm15/survive85/internal/events"
	"github.com/funstorm15/survive85/internal/game"
)

// Strategy picks a choice index among the available ones.
type Strategy string

const (
	Careful  Strategy = "careful"  // prefer wise, then neutral, then risky
	Reckless Strategy = "reckless" // prefer risky, then neutral, then wise
	Random   Strategy = "random"
)

// Strategies lists the built-in strategies.
var Strategies = []Strategy{Careful, Random, Reckless}

// Config describes a batch of runs.
type Config struct {
	Runs       int
	Seed       uint64
	Strategy   Strategy
	Region     string
	Sex        string
	Profession string // empty = rotate through all professions
}

// Report aggregates a batch.
type Report struct {
	Config     Config
	Runs       int
	Victories  int
	MeanAge    float64 // mean age at the end of the run
	MedianAge  int
	Causes     map[actuary.Cause]int
	Instant    int // deaths from instant checks
	Collapses  int
	Arrests    int
	MeanScore  float64
	EventsSeen map[string]int
	// Killers counts deaths by the event that caused them (instant checks),
	// "collapse" (Health reached zero) or "annual:<cause>" (yearly check).
	Killers      map[string]int
	NoEventYears int
	// Baseline is the life-table probability of reaching 85 from 18.
	Baseline float64
}

// VictoryRate is the share of runs that reached 85.
func (r *Report) VictoryRate() float64 {
	if r.Runs == 0 {
		return 0
	}
	return float64(r.Victories) / float64(r.Runs)
}

// Run plays the batch.
func Run(db *events.DB, book *actuary.Book, cfg Config) (*Report, error) {
	t, ok := book.TableFor(cfg.Region, cfg.Sex)
	if !ok {
		return nil, fmt.Errorf("no life table for %s/%s", cfg.Region, cfg.Sex)
	}
	rep := &Report{Config: cfg, Causes: map[actuary.Cause]int{}, EventsSeen: map[string]int{}, Killers: map[string]int{},
		Baseline: book.Survival(t, game.StartAge, game.VictoryAge)}
	pick := rand.New(rand.NewPCG(cfg.Seed, 0xC0FFEE))
	var ages []int
	totalAge, totalScore := 0, 0
	for i := 0; i < cfg.Runs; i++ {
		prof := cfg.Profession
		if prof == "" {
			prof = db.Professions[i%len(db.Professions)].ID
		}
		g, err := game.New(db, book, game.Options{Region: cfg.Region, Sex: cfg.Sex, Profession: prof, Seed: cfg.Seed + uint64(i)*7919})
		if err != nil {
			return nil, err
		}
		for steps := 0; g.Phase != game.PhaseDead && g.Phase != game.PhaseVictory; steps++ {
			if steps > 1000 {
				return nil, fmt.Errorf("run %d did not finish", i)
			}
			if g.Phase == game.PhaseOutcome {
				g.Next()
				continue
			}
			if g.Current == nil {
				rep.NoEventYears++
				if _, err := g.Choose(0); err != nil {
					return nil, err
				}
				continue
			}
			rep.EventsSeen[g.Current.ID]++
			idx := Choose(g, cfg.Strategy, pick)
			if _, err := g.Choose(idx); err != nil {
				return nil, fmt.Errorf("run %d age %d event %s: %w", i, g.Player.Age, g.Current.ID, err)
			}
		}
		rep.Runs++
		age := g.Player.AgeInt()
		if g.Phase == game.PhaseVictory {
			rep.Victories++
		} else {
			rep.Causes[g.Death.Cause]++
			switch {
			case g.Death.Instant:
				rep.Instant++
				rep.Killers[g.Death.EventID]++
			case g.Death.Collapse:
				rep.Collapses++
				rep.Killers["collapse"]++
			default:
				rep.Killers["annual:"+string(g.Death.Cause)]++
			}
		}
		rep.Arrests += g.Player.Arrests
		ages = append(ages, age)
		totalAge += age
		totalScore += g.Score()
	}
	if rep.Runs > 0 {
		rep.MeanAge = float64(totalAge) / float64(rep.Runs)
		rep.MeanScore = float64(totalScore) / float64(rep.Runs)
		sort.Ints(ages)
		rep.MedianAge = ages[len(ages)/2]
	}
	return rep, nil
}

// Choose returns the index of the choice the strategy takes for the current
// event. It only returns available choices; if none is available it returns
// the first visible one (the engine will reject it and the caller fails).
func Choose(g *game.Game, s Strategy, rng *rand.Rand) int {
	st := g.State()
	var avail []int
	for _, i := range g.VisibleChoices() {
		if g.Current.Choices[i].Availability(st) == events.Available {
			avail = append(avail, i)
		}
	}
	if len(avail) == 0 {
		if v := g.VisibleChoices(); len(v) > 0 {
			return v[0]
		}
		return 0
	}
	if s == Random {
		return avail[rng.IntN(len(avail))]
	}
	order := map[events.Tag]int{events.TagWise: 0, events.TagNeutral: 1, events.TagRisky: 2}
	if s == Reckless {
		order = map[events.Tag]int{events.TagRisky: 0, events.TagNeutral: 1, events.TagWise: 2}
	}
	best := avail[0]
	for _, i := range avail[1:] {
		if order[g.Current.Choices[i].Tag] < order[g.Current.Choices[best].Tag] {
			best = i
		}
	}
	return best
}

// Print writes a human-readable report.
func (r *Report) Print(w io.Writer) {
	c := r.Config
	fmt.Fprintf(w, "strategy=%s region=%s sex=%s profession=%q runs=%d seed=%d\n", c.Strategy, c.Region, c.Sex, c.Profession, r.Runs, c.Seed)
	fmt.Fprintf(w, "  reached 85: %5.1f%%   (life-table baseline %.1f%%)\n", 100*r.VictoryRate(), 100*r.Baseline)
	fmt.Fprintf(w, "  mean age at end: %.1f   median: %d   mean score: %.0f\n", r.MeanAge, r.MedianAge, r.MeanScore)
	fmt.Fprintf(w, "  instant deaths: %d   health collapses: %d   arrests: %d   quiet years: %d\n", r.Instant, r.Collapses, r.Arrests, r.NoEventYears)
	deaths := r.Runs - r.Victories
	for _, cause := range actuary.Causes {
		if n := r.Causes[cause]; n > 0 {
			fmt.Fprintf(w, "  %-15s %5.1f%% of deaths\n", cause, 100*float64(n)/float64(max(deaths, 1)))
		}
	}
	type kv struct {
		k string
		n int
	}
	var ks []kv
	for k, n := range r.Killers {
		ks = append(ks, kv{k, n})
	}
	sort.Slice(ks, func(i, j int) bool { return ks[i].n > ks[j].n || ks[i].n == ks[j].n && ks[i].k < ks[j].k })
	if len(ks) > 6 {
		ks = ks[:6]
	}
	fmt.Fprint(w, "  top killers:")
	for _, x := range ks {
		fmt.Fprintf(w, " %s=%d", x.k, x.n)
	}
	fmt.Fprintln(w)
}
