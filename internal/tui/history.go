package tui

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/funstorm15/survive85/internal/game"
	"github.com/funstorm15/survive85/internal/i18n"
)

// Record is one finished run, stored as a JSON line.
type Record struct {
	Time       time.Time `json:"time"`
	Version    string    `json:"version"`
	Seed       uint64    `json:"seed"`
	Region     string    `json:"region"`
	Sex        string    `json:"sex"`
	Profession string    `json:"profession"`
	Age        int       `json:"age"`
	Victory    bool      `json:"victory"`
	Cause      string    `json:"cause,omitempty"`
	Reason     i18n.Text `json:"reason,omitempty"`
	Score      int       `json:"score"`
}

// NewRecord summarises a finished game.
func NewRecord(g *game.Game, version string, now time.Time) Record {
	p := g.Player
	r := Record{
		Time: now.UTC(), Version: version, Seed: g.Seed, Region: p.Region, Sex: p.Sex,
		Profession: p.Profession, Age: p.AgeInt(), Victory: g.Phase == game.PhaseVictory, Score: g.Score(),
	}
	if d := g.Death; d != nil {
		r.Cause = string(d.Cause)
		r.Reason = d.Reason
	}
	return r
}

// DefaultHistoryPath is $XDG_DATA_HOME/survive85/history.jsonl, falling back
// to ~/.local/share/survive85/history.jsonl.
func DefaultHistoryPath() string {
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dir, "survive85", "history.jsonl")
}

// LoadHistory reads every record it can; a missing or damaged file yields
// what could be read.
func LoadHistory(path string) []Record {
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()
	var out []Record
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var r Record
		if json.Unmarshal(sc.Bytes(), &r) == nil {
			out = append(out, r)
		}
	}
	return out
}

// AppendHistory appends one record, creating the directory if needed.
func AppendHistory(path string, r Record) error {
	if path == "" {
		return os.ErrNotExist
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	b, err := json.Marshal(r)
	if err != nil {
		_ = f.Close()
		return err
	}
	if _, err := f.Write(append(b, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// BestRecord returns the highest-scoring run.
func BestRecord(rs []Record) (Record, bool) {
	var best Record
	ok := false
	for _, r := range rs {
		if !ok || r.Score > best.Score {
			best, ok = r, true
		}
	}
	return best, ok
}

// RecentDeaths returns up to n most recent deaths, newest first.
func RecentDeaths(rs []Record, n int) []Record {
	var out []Record
	for i := len(rs) - 1; i >= 0 && len(out) < n; i-- {
		if !rs[i].Victory {
			out = append(out, rs[i])
		}
	}
	return out
}
