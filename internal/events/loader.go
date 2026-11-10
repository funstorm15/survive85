package events

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/funstorm15/survive85/internal/actuary"
	"github.com/funstorm15/survive85/internal/i18n"
	"go.yaml.in/yaml/v3"
)

// Paths inside the data filesystem.
const (
	EventsGlob      = "events/*.yaml"
	EvidenceGlob    = "evidence/*.yaml"
	ConditionsGlob  = "conditions/*.yaml"
	ProfessionsFile = "professions.yaml"
	CausesFile      = "causes.yaml"
)

// Game limits the data must respect.
const (
	FirstAge   = 18
	LastAge    = 84
	MaxChoices = 4
	MinChoices = 2
	MaxWill    = 10
	MaxFactor  = 10.0
)

// CauseInfo is the display data for a cause of death.
type CauseInfo struct {
	Name    i18n.Text   `yaml:"name"`
	Reasons []i18n.Text `yaml:"reasons"` // generic death reasons, one is picked at random
}

// DB is the validated game data.
type DB struct {
	Events      []*Event
	Evidence    map[string]*Evidence
	Conditions  map[string]*Condition
	Professions []*Profession
	Causes      map[actuary.Cause]*CauseInfo

	byID         map[string]*Event
	professionID map[string]*Profession
}

// Event returns the event with the given id.
func (db *DB) Event(id string) (*Event, bool) {
	e, ok := db.byID[id]
	return e, ok
}

// Profession returns the profession with the given id.
func (db *DB) Profession(id string) (*Profession, bool) {
	p, ok := db.professionID[id]
	return p, ok
}

// Load reads every data file from fsys and validates the whole graph.
// All problems are reported together.
func Load(fsys fs.FS) (*DB, error) {
	db := &DB{
		Evidence:     map[string]*Evidence{},
		Conditions:   map[string]*Condition{},
		Causes:       map[actuary.Cause]*CauseInfo{},
		byID:         map[string]*Event{},
		professionID: map[string]*Profession{},
	}
	var errs []error
	add := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	evFiles, err := fs.Glob(fsys, EvidenceGlob)
	add(err)
	for _, f := range evFiles {
		var list []*Evidence
		if err := decodeFile(fsys, f, &list); err != nil {
			add(err)
			continue
		}
		for _, e := range list {
			if _, dup := db.Evidence[e.ID]; dup {
				add(fmt.Errorf("%s: duplicate evidence %q", f, e.ID))
			}
			db.Evidence[e.ID] = e
		}
	}

	condFiles, err := fs.Glob(fsys, ConditionsGlob)
	add(err)
	for _, f := range condFiles {
		var conds []*Condition
		if err := decodeFile(fsys, f, &conds); err != nil {
			add(err)
			continue
		}
		for _, c := range conds {
			if _, dup := db.Conditions[c.ID]; dup {
				add(fmt.Errorf("%s: duplicate condition %q", f, c.ID))
			}
			db.Conditions[c.ID] = c
		}
	}

	add(decodeFile(fsys, ProfessionsFile, &db.Professions))
	for _, p := range db.Professions {
		if _, dup := db.professionID[p.ID]; dup {
			add(fmt.Errorf("%s: duplicate profession %q", ProfessionsFile, p.ID))
		}
		db.professionID[p.ID] = p
	}

	causes := map[actuary.Cause]*CauseInfo{}
	add(decodeFile(fsys, CausesFile, &causes))
	db.Causes = causes

	files, err := fs.Glob(fsys, EventsGlob)
	add(err)
	sort.Strings(files)
	for _, f := range files {
		var list []*Event
		if err := decodeFile(fsys, f, &list); err != nil {
			add(err)
			continue
		}
		for _, e := range list {
			e.Source = path.Base(f)
			if e.Weight == 0 {
				e.Weight = 1
			}
			if _, dup := db.byID[e.ID]; dup {
				add(fmt.Errorf("%s: duplicate event id %q", f, e.ID))
				continue
			}
			db.byID[e.ID] = e
			db.Events = append(db.Events, e)
		}
	}

	errs = append(errs, db.validate()...)
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return db, nil
}

func decodeFile(fsys fs.FS, name string, v any) error {
	raw, err := fs.ReadFile(fsys, name)
	if err != nil {
		return err
	}
	dec := yaml.NewDecoder(strings.NewReader(string(raw)))
	dec.KnownFields(true)
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

var (
	evidenceID = regexp.MustCompile(`^ch(\d{2})_i(\d{1,2})$`)
	slugID     = regexp.MustCompile(`^[a-z0-9][a-z0-9_]*$`)
	doiRe      = regexp.MustCompile(`^10\.\d{4,9}/\S+$`)
)

// validate checks references, ranges and translations across all data.
func (db *DB) validate() []error {
	var errs []error
	errf := func(format string, a ...any) { errs = append(errs, fmt.Errorf(format, a...)) }

	setFlags := map[string]bool{}
	for _, p := range db.Professions {
		for _, f := range p.Flags {
			setFlags[f] = true
		}
	}
	collect := func(e Effects) {
		for _, f := range e.SetFlags {
			setFlags[f] = true
		}
	}
	for _, ev := range db.Events {
		for _, c := range ev.Choices {
			collect(c.Effects)
			if c.CatastropheCheck != nil {
				collect(c.CatastropheCheck.Effects)
			}
		}
	}
	for _, c := range db.Conditions {
		collect(c.PerYear)
	}

	refEvidence := func(where, id string) {
		if id == "" {
			return
		}
		if _, ok := db.Evidence[id]; !ok {
			errf("%s: unknown evidence %q", where, id)
		}
	}
	refCondition := func(where, id string) {
		if _, ok := db.Conditions[id]; !ok {
			errf("%s: unknown condition %q", where, id)
		}
	}
	checkText := func(where, field string, t i18n.Text, required bool) {
		if t.IsZero() {
			if required {
				errf("%s: %s is required", where, field)
			}
			return
		}
		if !t.Complete() {
			errf("%s: %s needs both ru and en", where, field)
		}
	}
	checkEffects := func(where string, e Effects) {
		for r, re := range e.StatEffects {
			if !r.Valid() {
				errf("%s: unknown risk %q", where, r)
			}
			if !(re.Factor > 0 && re.Factor <= MaxFactor) {
				errf("%s: %s factor %v outside (0, %v]", where, r, re.Factor, MaxFactor)
			}
			if re.Years < 0 {
				errf("%s: %s years must be >= 0", where, r)
			}
		}
		for _, id := range e.AddConditions {
			refCondition(where, id)
		}
		for _, id := range e.RemoveConditions {
			refCondition(where, id)
		}
		if e.DeltaHealth < -100 || e.DeltaHealth > 100 || e.DeltaSanity < -100 || e.DeltaSanity > 100 {
			errf("%s: health/sanity deltas must be within ±100", where)
		}
	}
	checkReq := func(where string, r Requirements) {
		for _, f := range append(append(append([]string{}, r.Flags...), r.NotFlags...), keys(r.FlagYears)...) {
			if !setFlags[f] {
				errf("%s: flag %q is required but nothing ever sets it", where, f)
			}
		}
		for _, id := range append(append([]string{}, r.Conditions...), r.NotConditions...) {
			refCondition(where, id)
		}
		for rk := range r.MinRisk {
			if !rk.Valid() {
				errf("%s: unknown risk %q", where, rk)
			}
		}
		for _, c := range r.Classes {
			if _, ok := db.professionID[c]; !ok {
				errf("%s: unknown profession %q", where, c)
			}
		}
		if r.Sex != "" && r.Sex != "male" && r.Sex != "female" {
			errf("%s: sex must be male or female", where)
		}
	}
	checkReasons := func(where string, m map[actuary.Cause]i18n.Text) {
		for c, t := range m {
			if !c.Valid() {
				errf("%s: unknown cause %q", where, c)
			}
			checkText(where, "death_reasons."+string(c), t, true)
		}
	}

	for id, e := range db.Evidence {
		where := "evidence " + id
		m := evidenceID.FindStringSubmatch(id)
		if m == nil {
			errf("%s: id must look like ch01_i02", where)
		} else if fmt.Sprintf("ch%02d_i%d", e.Chapter, e.Item) != id {
			errf("%s: id does not match chapter %d item %d", where, e.Chapter, e.Item)
		}
		if e.Chapter < 1 || e.Chapter > 34 || e.Item < 1 {
			errf("%s: chapter/item out of range", where)
		}
		switch e.Grade {
		case "A", "B", "C":
		default:
			errf("%s: grade must be A, B or C", where)
		}
		checkText(where, "title", e.Title, true)
		checkText(where, "chapter_title", e.ChapterTitle, true)
		checkText(where, "plain", e.Plain, true)
		checkText(where, "benefit", e.Benefit, true)
		if len(e.Sources) == 0 {
			errf("%s: at least one source is required", where)
		}
		if e.DOI != "" && !doiRe.MatchString(e.DOI) {
			errf("%s: malformed DOI %q", where, e.DOI)
		}
		for i, n := range e.Numbers {
			checkText(fmt.Sprintf("%s number %d", where, i), "outcome", n.Outcome, false)
		}
	}

	for id, c := range db.Conditions {
		where := "condition " + id
		if !slugID.MatchString(id) {
			errf("%s: id must be snake_case", where)
		}
		checkText(where, "name", c.Name, true)
		checkText(where, "description", c.Description, true)
		switch c.Kind {
		case "good", "bad", "neutral":
		default:
			errf("%s: kind must be good, bad or neutral", where)
		}
		for r, m := range c.Modifiers {
			if !r.Valid() {
				errf("%s: unknown risk %q", where, r)
			}
			if !(m > 0 && m <= MaxFactor) {
				errf("%s: modifier %v outside (0, %v]", where, m, MaxFactor)
			}
		}
		checkEffects(where+" per_year", c.PerYear)
		if c.ForcesEvent != "" {
			if _, ok := db.byID[c.ForcesEvent]; !ok {
				errf("%s: forces unknown event %q", where, c.ForcesEvent)
			}
		}
		refEvidence(where, c.Evidence)
		checkReasons(where, c.DeathReasons)
	}

	if len(db.Professions) == 0 {
		errf("%s: no professions", ProfessionsFile)
	}
	for _, p := range db.Professions {
		where := "profession " + p.ID
		checkText(where, "name", p.Name, true)
		checkText(where, "description", p.Description, true)
		for r, m := range p.Modifiers {
			if !r.Valid() || !(m > 0 && m <= MaxFactor) {
				errf("%s: bad modifier %s=%v", where, r, m)
			}
		}
		for _, c := range p.Conditions {
			refCondition(where, c)
		}
		checkEffects(where+" per_year", p.PerYear)
		refEvidence(where, p.Evidence)
	}

	for _, c := range actuary.Causes {
		info, ok := db.Causes[c]
		if !ok {
			errf("%s: missing cause %q", CausesFile, c)
			continue
		}
		checkText(CausesFile+" "+string(c), "name", info.Name, true)
		if len(info.Reasons) == 0 {
			errf("%s %s: needs at least one reason", CausesFile, c)
		}
		for _, r := range info.Reasons {
			checkText(CausesFile+" "+string(c), "reason", r, true)
		}
	}

	for _, ev := range db.Events {
		where := ev.Source + ": event " + ev.ID
		if !slugID.MatchString(ev.ID) {
			errf("%s: id must be snake_case", where)
		}
		checkText(where, "title", ev.Title, true)
		checkText(where, "narrative", ev.Narrative, true)
		if ev.MinAge < FirstAge || ev.MaxAge > LastAge || ev.MinAge > ev.MaxAge {
			errf("%s: ages %d..%d must lie within %d..%d", where, ev.MinAge, ev.MaxAge, FirstAge, LastAge)
		}
		if !validCategory(ev.Category) {
			errf("%s: unknown category %q", where, ev.Category)
		}
		if ev.Weight < 0 || ev.Cooldown < 0 {
			errf("%s: weight and cooldown must be >= 0", where)
		}
		for r := range ev.WeightBy {
			if !r.Valid() {
				errf("%s: weight_by unknown risk %q", where, r)
			}
		}
		refEvidence(where, ev.Evidence)
		checkReq(where, ev.Requires)
		if n := len(ev.Choices); n < MinChoices || n > MaxChoices {
			if n != 1 || !ev.Repeatable || !ev.forcedOnly(db) {
				errf("%s: needs %d-%d choices, has %d", where, MinChoices, MaxChoices, n)
			}
		}
		ids := map[string]bool{}
		for _, ch := range ev.Choices {
			cw := where + " choice " + ch.ID
			if ch.ID == "" || ids[ch.ID] {
				errf("%s: choice id empty or duplicate", cw)
			}
			ids[ch.ID] = true
			checkText(cw, "text", ch.Text, true)
			checkText(cw, "outcome", ch.Outcome, true)
			switch ch.Tag {
			case TagWise, TagNeutral, TagRisky:
			default:
				errf("%s: tag must be wise, neutral or risky", cw)
			}
			if ch.CostMoney < 0 || ch.CostWill < 0 || ch.CostWill > MaxWill || ch.CostTime < 0 {
				errf("%s: costs must be >= 0 and will <= %d", cw, MaxWill)
			}
			checkReq(cw, ch.Requires)
			checkEffects(cw, ch.Effects)
			refEvidence(cw, ch.Evidence)
			checkReasons(cw, ch.DeathReasons)
			if d := ch.InstantDeathCheck; d != nil {
				if !(d.Chance > 0 && d.Chance <= 1) {
					errf("%s: instant_death_check chance must be in (0, 1]", cw)
				}
				if !d.Cause.Valid() {
					errf("%s: instant_death_check cause %q unknown", cw, d.Cause)
				}
				checkText(cw, "death_reason", d.DeathReason, true)
				if d.Evidence == "" {
					errf("%s: instant_death_check needs evidence", cw)
				}
				refEvidence(cw, d.Evidence)
				for _, r := range d.ScaleBy {
					if !r.Valid() {
						errf("%s: scale_by unknown risk %q", cw, r)
					}
				}
			}
			if k := ch.CatastropheCheck; k != nil {
				if !(k.Chance > 0 && k.Chance <= 1) {
					errf("%s: catastrophe_check chance must be in (0, 1]", cw)
				}
				switch k.Kind {
				case KindPrison:
					if k.Years <= 0 {
						errf("%s: prison needs years > 0", cw)
					}
				case KindInjury, KindRuin, KindLoss:
				default:
					errf("%s: catastrophe kind %q unknown", cw, k.Kind)
				}
				if k.WealthFraction < 0 || k.WealthFraction > 1 {
					errf("%s: wealth_fraction must be in [0, 1]", cw)
				}
				checkText(cw, "catastrophe text", k.Text, true)
				checkEffects(cw+" catastrophe", k.Effects)
				refEvidence(cw, k.Evidence)
				for _, r := range k.ScaleBy {
					if !r.Valid() {
						errf("%s: scale_by unknown risk %q", cw, r)
					}
				}
			}
		}
	}
	return errs
}

// forcedOnly reports whether some condition forces this event, which allows
// it to have a single "carry on" choice (a year in prison, for instance).
func (ev *Event) forcedOnly(db *DB) bool {
	for _, c := range db.Conditions {
		if c.ForcesEvent == ev.ID {
			return true
		}
	}
	return false
}

func validCategory(c Category) bool {
	for _, x := range Categories {
		if x == c {
			return true
		}
	}
	return false
}

func keys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
