// Command validatedata validates the game data, optionally restricted to the
// core files plus the files of some chapters, so that one chapter can be
// checked while others are still being written.
//
//	go run ./tools/validatedata                 # everything (same as go test ./data/)
//	go run ./tools/validatedata -only 02,ch02   # core + files whose name starts with 02 or ch02
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing/fstest"

	"github.com/funstorm15/survive85/internal/events"
)

func main() {
	dir := flag.String("dir", "data", "data directory")
	only := flag.String("only", "", "comma-separated file-name prefixes to include besides the core files")
	flag.Parse()

	src := os.DirFS(*dir)
	fsys := src
	if *only != "" {
		prefixes := strings.Split(*only, ",")
		m := fstest.MapFS{}
		err := fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".yaml") {
				return err
			}
			base := path.Base(p)
			keep := !strings.Contains(p, "/") || strings.HasPrefix(base, "00_")
			for _, pre := range prefixes {
				keep = keep || strings.HasPrefix(base, strings.TrimSpace(pre))
			}
			if keep {
				b, err := fs.ReadFile(src, p)
				if err != nil {
					return err
				}
				m[p] = &fstest.MapFile{Data: b}
			}
			return nil
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fsys = m
	}
	db, err := events.Load(fsys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "INVALID:\n%v\n", err)
		os.Exit(1)
	}
	choices := 0
	for _, e := range db.Events {
		choices += len(e.Choices)
	}
	fmt.Printf("ok: %d events, %d choices, %d evidence, %d conditions\n", len(db.Events), choices, len(db.Evidence), len(db.Conditions))
}
