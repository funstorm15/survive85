// Command survive85 is a turn-based terminal roguelike: live from 18 to 85,
// one year per turn, with every risk taken from the evidence-graded book
// "HowToLiveBetter" and WHO life tables.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/funstorm15/survive85/data"
	"github.com/funstorm15/survive85/internal/actuary"
	"github.com/funstorm15/survive85/internal/events"
	"github.com/funstorm15/survive85/internal/i18n"
	"github.com/funstorm15/survive85/internal/sim"
	"github.com/funstorm15/survive85/internal/tui"
)

// Set by GoReleaser through -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	cmd := "play"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd, args = args[0], args[1:]
	}
	switch cmd {
	case "play":
		return play(args, stderr)
	case "sim":
		return simulate(args, stdout, stderr)
	case "validate":
		return validate(stdout, stderr)
	case "version":
		fmt.Fprintf(stdout, "survive85 %s (commit %s, built %s)\n", version, commit, date)
		return 0
	case "help":
		usage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "survive85: unknown command %q\n\n", cmd)
		usage(stderr)
		return 2
	}
}

func usage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  survive85 [flags]            play (default)
  survive85 sim [flags]        simulate many runs and print a balance report
  survive85 validate           check the embedded game data
  survive85 version            print the version

Play flags:
  -lang ru|en        interface language (default: from $SURVIVE85_LANG / $LANG)
  -seed N            seed for a reproducible life (default: random)
  -daily             today's seed: everyone gets the same life today
  -theme NAME        nord (default), cyberpunk or monokai

Keys: ↑/↓ or j/k move · Enter/Space/l choose · h/Esc back · 1-4 quick pick
      L language · T theme · ? help · q quit
`)
}

func loadAll() (*events.DB, *actuary.Book, error) {
	db, err := events.Load(data.FS)
	if err != nil {
		return nil, nil, err
	}
	book, err := actuary.Load(data.FS, data.LifeTables)
	if err != nil {
		return nil, nil, err
	}
	return db, book, nil
}

func play(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("survive85", flag.ContinueOnError)
	fs.SetOutput(stderr)
	lang := fs.String("lang", "", "interface language: ru or en")
	seed := fs.Uint64("seed", 0, "seed for a reproducible run")
	daily := fs.Bool("daily", false, "use today's seed")
	theme := fs.String("theme", "nord", "colour theme: nord, cyberpunk, monokai")
	fs.Usage = func() { usage(stderr) }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	l := i18n.Detect()
	if *lang != "" {
		var ok bool
		if l, ok = i18n.Parse(*lang); !ok {
			fmt.Fprintf(stderr, "survive85: unsupported language %q (use ru or en)\n", *lang)
			return 2
		}
	}
	db, book, err := loadAll()
	if err != nil {
		fmt.Fprintf(stderr, "survive85: broken game data:\n%v\n", err)
		return 1
	}
	s := *seed
	switch {
	case *daily:
		y, m, d := time.Now().Date()
		s = uint64(y*10000 + int(m)*100 + d)
	case s == 0:
		s = uint64(time.Now().UnixNano())
	}
	cfg := tui.Config{DB: db, Book: book, Lang: l, Seed: s, FixedSeed: *daily || *seed != 0, Theme: *theme, Version: version}
	if err := tui.Run(cfg); err != nil {
		fmt.Fprintf(stderr, "survive85: %v\n", err)
		return 1
	}
	return 0
}

func simulate(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("sim", flag.ContinueOnError)
	fs.SetOutput(stderr)
	runs := fs.Int("runs", 2000, "number of runs per strategy")
	strategy := fs.String("strategy", "all", "careful, random, reckless or all")
	region := fs.String("region", "world", "life table region: world, rus, usa")
	sex := fs.String("sex", "male", "male or female")
	prof := fs.String("class", "", "profession id (default: rotate through all)")
	seed := fs.Uint64("seed", 1, "base seed")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	db, book, err := loadAll()
	if err != nil {
		fmt.Fprintf(stderr, "survive85: broken game data:\n%v\n", err)
		return 1
	}
	var strategies []sim.Strategy
	if *strategy == "all" {
		strategies = sim.Strategies
	} else {
		strategies = []sim.Strategy{sim.Strategy(*strategy)}
	}
	for _, st := range strategies {
		switch st {
		case sim.Careful, sim.Random, sim.Reckless:
		default:
			fmt.Fprintf(stderr, "survive85: unknown strategy %q\n", st)
			return 2
		}
		rep, err := sim.Run(db, book, sim.Config{Runs: *runs, Seed: *seed, Strategy: st, Region: *region, Sex: *sex, Profession: *prof})
		if err != nil {
			fmt.Fprintf(stderr, "survive85: %v\n", err)
			return 1
		}
		rep.Print(stdout)
	}
	return 0
}

func validate(stdout, stderr io.Writer) int {
	db, book, err := loadAll()
	if err != nil {
		fmt.Fprintf(stderr, "invalid game data:\n%v\n", err)
		return 1
	}
	grades := map[string]int{}
	for _, e := range db.Evidence {
		grades[e.Grade]++
	}
	choices := 0
	for _, e := range db.Events {
		choices += len(e.Choices)
	}
	fmt.Fprintf(stdout, "ok: %d events (%d choices), %d evidence items (A=%d B=%d C=%d), %d conditions, %d professions, %d life tables\n",
		len(db.Events), choices, len(db.Evidence), grades["A"], grades["B"], grades["C"], len(db.Conditions), len(db.Professions), len(book.Tables))
	return 0
}
