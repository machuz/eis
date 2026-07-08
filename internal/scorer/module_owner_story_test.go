package scorer

import (
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/metric"
)

// TestScoreModules_OwnerStoryDates verifies the debt-drill wiring a unit test on
// computeStructuralDebt alone can't see: ScoreModules must derive
// OwnerLastActiveDays from authorLastDate[TopAuthor] and ModuleUntouchedDays
// from moduleLastDate — the "owner left N days ago · untouched M days" story.
func TestScoreModules_OwnerStoryDates(t *testing.T) {
	now := time.Now()
	ownerLeft := now.AddDate(0, 0, -300) // owner last committed 300 days ago
	lastTouch := now.AddDate(0, 0, -200) // module last touched 200 days ago

	ownership := []metric.ModuleOwnership{{
		Module:      "svc/auth",
		TotalLines:  400,
		AuthorCount: 1,
		TopAuthor:   "tanaka",
		TopShare:    0.9,
		Level:       "SOLE_OWNER",
	}}

	scores := ScoreModules(
		metric.ChangePressure{}, // no pressure data
		nil,                     // no co-change results
		ownership,               // one sole-owned module
		map[string]float64{},    // no survival data
		map[string]time.Time{"tanaka": ownerLeft},
		30, // activeDays — owner (300d) is well past this ⇒ Orphaned + inactive
		nil,
		map[string]time.Time{"svc/auth": lastTouch},
	)

	var got *ModuleScore
	for i := range scores {
		if scores[i].Module == "svc/auth" {
			got = &scores[i]
			break
		}
	}
	if got == nil {
		t.Fatalf("svc/auth not found in %d scored modules", len(scores))
	}

	if got.OwnerActive {
		t.Errorf("OwnerActive = true, want false (owner left 300d ago)")
	}
	if got.Ownership != "Orphaned" {
		t.Errorf("Ownership = %q, want Orphaned", got.Ownership)
	}
	// Allow a small slack for wall-clock drift during the test.
	if got.OwnerLastActiveDays < 299 || got.OwnerLastActiveDays > 301 {
		t.Errorf("OwnerLastActiveDays = %.2f, want ~300", got.OwnerLastActiveDays)
	}
	if got.ModuleUntouchedDays < 199 || got.ModuleUntouchedDays > 201 {
		t.Errorf("ModuleUntouchedDays = %.2f, want ~200", got.ModuleUntouchedDays)
	}
}
