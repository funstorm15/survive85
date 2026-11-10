// Package actuary provides the baseline mortality the game multiplies with a
// player's risk modifiers: single-year death probabilities, remaining life
// expectancy and the share of deaths by cause group, per region and sex.
//
// The numbers are generated from WHO data by tools/genlifetables and embedded
// as YAML; see data/lifetables/README.md for provenance.
package actuary

import (
	"fmt"
	"io/fs"
	"math"
	"sort"

	"github.com/funstorm15/survive85/internal/i18n"
	"go.yaml.in/yaml/v3"
)

// Cause is a cause-of-death group the game can attribute a death to.
type Cause string

const (
	Cardiovascular Cause = "cardiovascular" // WHO GHE: cardiovascular diseases + diabetes
	Oncology       Cause = "oncology"       // malignant neoplasms
	Liver          Cause = "liver"          // digestive diseases, alcohol/drug use disorders, poisonings
	Trauma         Cause = "trauma"         // injuries except poisonings and self-harm
	SelfHarm       Cause = "self_harm"      // self-harm
	Other          Cause = "other"          // everything else (infections, respiratory, neurological, ...)
)

// Causes lists every cause group in a stable order.
var Causes = []Cause{Cardiovascular, Oncology, Liver, Trauma, SelfHarm, Other}

// Valid reports whether c is a known cause group.
func (c Cause) Valid() bool {
	for _, x := range Causes {
		if x == c {
			return true
		}
	}
	return false
}

// Source describes where a dataset comes from.
type Source struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// Band holds cause-of-death shares for a five-year age band [From, To].
type Band struct {
	From      int               `yaml:"from"`
	To        int               `yaml:"to"`
	Fractions map[Cause]float64 `yaml:"fractions"`
}

// Table is a life table for one region and sex, covering ages FirstAge..LastAge.
type Table struct {
	ID     string    `yaml:"id"`
	Region string    `yaml:"region"`
	Sex    string    `yaml:"sex"`
	Name   i18n.Text `yaml:"name"`
	// Q[i] is the probability of dying between exact ages FirstAge+i and FirstAge+i+1.
	Q []float64 `yaml:"q"`
	// EX[i] is the remaining life expectancy at exact age FirstAge+i.
	EX    []float64 `yaml:"ex"`
	Bands []Band    `yaml:"bands"`
}

// Book is the full set of embedded life tables.
type Book struct {
	Year     int      `yaml:"year"`
	FirstAge int      `yaml:"first_age"`
	LastAge  int      `yaml:"last_age"`
	Sources  []Source `yaml:"sources"`
	Tables   []Table  `yaml:"tables"`

	byID map[string]*Table
}

// Load reads and validates the life tables at path inside fsys.
func Load(fsys fs.FS, path string) (*Book, error) {
	raw, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}
	var b Book
	if err := yaml.Unmarshal(raw, &b); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if err := b.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &b, nil
}

func (b *Book) validate() error {
	if b.FirstAge <= 0 || b.LastAge <= b.FirstAge {
		return fmt.Errorf("bad age range %d..%d", b.FirstAge, b.LastAge)
	}
	if len(b.Tables) == 0 {
		return fmt.Errorf("no tables")
	}
	n := b.LastAge - b.FirstAge + 1
	b.byID = make(map[string]*Table, len(b.Tables))
	for i := range b.Tables {
		t := &b.Tables[i]
		if _, dup := b.byID[t.ID]; dup {
			return fmt.Errorf("duplicate table %q", t.ID)
		}
		b.byID[t.ID] = t
		if len(t.Q) != n || len(t.EX) != n {
			return fmt.Errorf("table %s: want %d ages, got q=%d ex=%d", t.ID, n, len(t.Q), len(t.EX))
		}
		for i, q := range t.Q {
			if !(q > 0 && q < 1) {
				return fmt.Errorf("table %s: q at age %d out of range: %v", t.ID, b.FirstAge+i, q)
			}
		}
		if !t.Name.Complete() {
			return fmt.Errorf("table %s: name must have ru and en", t.ID)
		}
		sort.Slice(t.Bands, func(i, j int) bool { return t.Bands[i].From < t.Bands[j].From })
		for age := b.FirstAge; age <= b.LastAge; age++ {
			band := t.band(age)
			if band == nil {
				return fmt.Errorf("table %s: no cause band covers age %d", t.ID, age)
			}
			sum := 0.0
			for _, c := range Causes {
				f := band.Fractions[c]
				if f < 0 {
					return fmt.Errorf("table %s: negative fraction for %s", t.ID, c)
				}
				sum += f
			}
			if math.Abs(sum-1) > 0.01 {
				return fmt.Errorf("table %s band %d-%d: fractions sum to %.3f", t.ID, band.From, band.To, sum)
			}
		}
	}
	return nil
}

// Table returns the table with the given id.
func (b *Book) Table(id string) (*Table, bool) {
	t, ok := b.byID[id]
	return t, ok
}

// TableFor returns the table for a region and sex, e.g. ("rus", "male").
func (b *Book) TableFor(region, sex string) (*Table, bool) {
	return b.Table(region + "_" + sex)
}

// Regions lists region codes in file order without duplicates.
func (b *Book) Regions() []string {
	var out []string
	seen := map[string]bool{}
	for _, t := range b.Tables {
		if !seen[t.Region] {
			seen[t.Region] = true
			out = append(out, t.Region)
		}
	}
	return out
}

// RegionName returns the localized name of a region.
func (b *Book) RegionName(region string) i18n.Text {
	for _, t := range b.Tables {
		if t.Region == region {
			return t.Name
		}
	}
	return i18n.T(region, region)
}

func (t *Table) band(age int) *Band {
	for i := range t.Bands {
		if age >= t.Bands[i].From && age <= t.Bands[i].To {
			return &t.Bands[i]
		}
	}
	return nil
}

// QAt is the baseline probability of dying within the year that starts at
// exact age age. Ages outside the table are clamped to its ends.
func (b *Book) QAt(t *Table, age int) float64 {
	return t.Q[b.index(age)]
}

// EXAt is the baseline remaining life expectancy at exact age age.
func (b *Book) EXAt(t *Table, age int) float64 {
	return t.EX[b.index(age)]
}

// Fraction is the share of baseline deaths at age attributed to cause c.
func (b *Book) Fraction(t *Table, age int, c Cause) float64 {
	age = b.FirstAge + b.index(age)
	if band := t.band(age); band != nil {
		return band.Fractions[c]
	}
	return 0
}

// Survival is the baseline probability of surviving from exact age from to
// exact age to (to > from), using the single-year probabilities.
func (b *Book) Survival(t *Table, from, to int) float64 {
	s := 1.0
	for a := from; a < to; a++ {
		s *= 1 - b.QAt(t, a)
	}
	return s
}

func (b *Book) index(age int) int {
	switch {
	case age < b.FirstAge:
		return 0
	case age > b.LastAge:
		return b.LastAge - b.FirstAge
	default:
		return age - b.FirstAge
	}
}
