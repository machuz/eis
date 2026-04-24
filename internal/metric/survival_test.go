package metric

import (
	"math"
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

func bl(author, file string, committerDaysAgo int, ref time.Time) git.BlameLine {
	return git.BlameLine{
		Author:        author,
		Filename:      file,
		CommitterTime: ref.AddDate(0, 0, -committerDaysAgo),
	}
}

// No test coverage info → behaves exactly like the old CalcSurvival.
func TestCalcSurvival_LegacyMatchesFull(t *testing.T) {
	ref := time.Now()
	lines := []git.BlameLine{
		bl("A", "foo.go", 0, ref),
		bl("A", "bar.go", 90, ref),
		bl("B", "baz.go", 10, ref),
	}
	legacy := CalcSurvival(lines, 180, ref)
	full := CalcSurvivalFull(lines, 180, ref, nil, 0, nil, 1.0)
	for author, v := range legacy.Decayed {
		if math.Abs(full.Decayed[author]-v) > 1e-9 {
			t.Errorf("author %s: Decayed mismatch legacy=%v full=%v", author, v, full.Decayed[author])
		}
	}
}

// Untested lines contribute to Decayed at α rate; Tested/Untested maps expose
// the pre-α split for observability.
// Tested and untested files MUST live in separate directories — module fallback
// otherwise marks everything in a dir that contains any test as tested.
func TestCalcSurvivalFull_AlphaWeighting(t *testing.T) {
	ref := time.Now()
	lines := []git.BlameLine{
		bl("A", "covered/tested.go", 0, ref),
		bl("A", "covered/tested.go", 0, ref),
		bl("A", "legacy/orphan.go", 0, ref),
		bl("A", "legacy/orphan.go", 0, ref),
	}
	testedSet := BuildTestedSet([]string{
		"covered/tested.go",
		"covered/tested_test.go",
		"legacy/orphan.go",
	})
	if !testedSet.IsTested("covered/tested.go") {
		t.Fatal("covered/tested.go must be marked tested (sibling pair)")
	}
	if testedSet.IsTested("legacy/orphan.go") {
		t.Fatal("legacy/orphan.go must be untested (no test in legacy/)")
	}

	result := CalcSurvivalFull(lines, 180, ref, nil, 0, testedSet, 0.5)
	// Decayed(A) = 2 tested × 1.0 + 2 untested × 0.5 = 3.0
	if got := result.Decayed["A"]; math.Abs(got-3.0) > 1e-6 {
		t.Errorf("Decayed = %v, want 3.0", got)
	}
	if got := result.Tested["A"]; math.Abs(got-2.0) > 1e-6 {
		t.Errorf("Tested = %v, want 2.0", got)
	}
	if got := result.Untested["A"]; math.Abs(got-2.0) > 1e-6 {
		t.Errorf("Untested = %v, want 2.0", got)
	}
}

// α=0 makes untested lines disappear entirely from Decayed.
func TestCalcSurvivalFull_AlphaZero(t *testing.T) {
	ref := time.Now()
	lines := []git.BlameLine{
		bl("A", "covered/tested.go", 0, ref),
		bl("A", "legacy/orphan.go", 0, ref),
	}
	testedSet := BuildTestedSet([]string{
		"covered/tested.go",
		"covered/tested_test.go",
		"legacy/orphan.go",
	})
	result := CalcSurvivalFull(lines, 180, ref, nil, 0, testedSet, 0)
	if got := result.Decayed["A"]; math.Abs(got-1.0) > 1e-6 {
		t.Errorf("α=0: Decayed = %v, want 1.0 (only tested line counts)", got)
	}
}

// α=1.0 must match classic behaviour.
func TestCalcSurvivalFull_AlphaOneMatchesClassic(t *testing.T) {
	ref := time.Now()
	lines := []git.BlameLine{
		bl("A", "covered/tested.go", 0, ref),
		bl("A", "legacy/orphan.go", 0, ref),
	}
	testedSet := BuildTestedSet([]string{
		"covered/tested.go",
		"covered/tested_test.go",
		"legacy/orphan.go",
	})
	classic := CalcSurvival(lines, 180, ref)
	full := CalcSurvivalFull(lines, 180, ref, nil, 0, testedSet, 1.0)
	if math.Abs(classic.Decayed["A"]-full.Decayed["A"]) > 1e-9 {
		t.Errorf("α=1.0 should match classic: classic=%v full=%v", classic.Decayed["A"], full.Decayed["A"])
	}
}

// α weighting also affects the Robust/Dormant split when pressure data is present.
// Each pressure bucket has its own tested/untested subdir so module fallback
// doesn't bleed coverage.
func TestCalcSurvivalFull_PressureAndCoverage(t *testing.T) {
	ref := time.Now()
	lines := []git.BlameLine{
		bl("A", "high/covered/tested.go", 0, ref),
		bl("A", "high/legacy/orphan.go", 0, ref),
		bl("A", "low/covered/tested.go", 0, ref),
		bl("A", "low/legacy/orphan.go", 0, ref),
	}
	// ChangePressure is keyed by ModuleOf() which uses up to 3 path components.
	pressure := ChangePressure{
		"high/covered": 10.0,
		"high/legacy":  10.0,
		"low/covered":  0.1,
		"low/legacy":   0.1,
	}
	testedSet := BuildTestedSet([]string{
		"high/covered/tested.go", "high/covered/tested_test.go", "high/legacy/orphan.go",
		"low/covered/tested.go", "low/covered/tested_test.go", "low/legacy/orphan.go",
	})
	if testedSet.IsTested("high/legacy/orphan.go") {
		t.Fatal("high/legacy/orphan.go must be untested")
	}
	result := CalcSurvivalFull(lines, 180, ref, pressure, 1.0, testedSet, 0.5)

	// high bucket: 1.0 (tested) + 0.5 (untested) = 1.5
	if got := result.Robust["A"]; math.Abs(got-1.5) > 1e-6 {
		t.Errorf("Robust = %v, want 1.5", got)
	}
	if got := result.Dormant["A"]; math.Abs(got-1.5) > 1e-6 {
		t.Errorf("Dormant = %v, want 1.5", got)
	}
}
