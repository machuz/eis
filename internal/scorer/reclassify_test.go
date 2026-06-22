package scorer

import (
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/config"
)

// baseActiveResult is a stored Result whose only strong State signal is
// recency: low debt/raw/indispensability so the "Active" rule (RecentlyActive
// → 0.80) wins while recent, and loses once the calendar passes active_days.
func baseActiveResult(lastActive time.Time) Result {
	return Result{
		Author:           "alice",
		Production:       50,
		Catalysis:        20,
		Survival:         50,
		Design:           20,
		Breadth:          40,
		DebtCleanup:      50,
		Indispensability: 20,
		Gravity:          42,
		Impact:           33,
		TotalCommits:     50,
		Role:             "Producer",
		RoleConf:         0.7,
		Style:            "Builder",
		StyleConf:        0.6,
		LastActiveDate:   lastActive,
	}
}

// Within active_days the engineer is RecentlyActive and State is Active; once
// refTime advances past active_days the same stored Result reclassifies to a
// non-Active state — the daily-moving outputs the cheap cadence layer refreshes
// without touching git.
func TestReclassifyState_CrossesActiveThreshold(t *testing.T) {
	cfg := config.Default() // ActiveDays default 30
	lastActive := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	recent := ReclassifyState(baseActiveResult(lastActive), cfg, lastActive.AddDate(0, 0, 10))
	if !recent.RecentlyActive {
		t.Fatalf("10 days after last commit (active_days=30): want RecentlyActive=true")
	}
	if recent.State != "Active" {
		t.Fatalf("recent engineer: want State=Active, got %q", recent.State)
	}

	stale := ReclassifyState(baseActiveResult(lastActive), cfg, lastActive.AddDate(0, 0, 40))
	if stale.RecentlyActive {
		t.Fatalf("40 days after last commit (active_days=30): want RecentlyActive=false")
	}
	if stale.State == "Active" {
		t.Fatalf("stale engineer: State must no longer be Active, got %q", stale.State)
	}
}

// ReclassifyState touches only RecentlyActive / State / StateConf; every other
// field of the stored Result passes through unchanged. Role and Style are
// HEAD-derived and must not move with the calendar.
func TestReclassifyState_PassesThroughInvariantFields(t *testing.T) {
	cfg := config.Default()
	lastActive := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	in := baseActiveResult(lastActive)

	out := ReclassifyState(in, cfg, lastActive.AddDate(0, 0, 200))

	if out.Author != in.Author || out.Role != in.Role || out.RoleConf != in.RoleConf ||
		out.Style != in.Style || out.StyleConf != in.StyleConf ||
		out.Gravity != in.Gravity || out.Impact != in.Impact ||
		out.Survival != in.Survival || out.Production != in.Production ||
		out.TotalCommits != in.TotalCommits || !out.LastActiveDate.Equal(in.LastActiveDate) {
		t.Fatalf("invariant field mutated by ReclassifyState:\n in=%+v\nout=%+v", in, out)
	}
}

// Pure and deterministic (W-02): identical (r, cfg, refTime) → identical output,
// and refTime is explicit (no time.Now()).
func TestReclassifyState_Deterministic(t *testing.T) {
	cfg := config.Default()
	lastActive := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ref := lastActive.AddDate(0, 0, 45)

	a := ReclassifyState(baseActiveResult(lastActive), cfg, ref)
	b := ReclassifyState(baseActiveResult(lastActive), cfg, ref)
	if a != b {
		t.Fatalf("non-deterministic:\n a=%+v\n b=%+v", a, b)
	}
}

// A zero LastActiveDate (last-commit date unavailable) is treated as not recent
// rather than as the Unix epoch being "0 days ago"; the classifier must not flip
// to Active on missing data.
func TestReclassifyState_ZeroLastActiveDateIsNotRecent(t *testing.T) {
	cfg := config.Default()
	in := baseActiveResult(time.Time{}) // zero value
	out := ReclassifyState(in, cfg, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	if out.RecentlyActive {
		t.Fatalf("zero LastActiveDate: want RecentlyActive=false")
	}
}
