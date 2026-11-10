package data_test

import (
	"testing"

	"github.com/funstorm15/survive85/data"
	"github.com/funstorm15/survive85/internal/actuary"
	"github.com/funstorm15/survive85/internal/events"
)

// TestData validates every embedded data file: schema, translations,
// references between events, evidence and conditions, and the life tables.
func TestData(t *testing.T) {
	db, err := events.Load(data.FS)
	if err != nil {
		t.Fatalf("game data is invalid:\n%v", err)
	}
	if _, err := actuary.Load(data.FS, data.LifeTables); err != nil {
		t.Fatalf("life tables are invalid: %v", err)
	}
	t.Logf("%d events, %d evidence items, %d conditions, %d professions",
		len(db.Events), len(db.Evidence), len(db.Conditions), len(db.Professions))
}
