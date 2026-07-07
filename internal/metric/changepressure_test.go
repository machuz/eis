package metric

import (
	"math"
	"testing"

	"github.com/machuz/eis/v2/internal/git"
)

func commitTouching(author string, files ...string) git.Commit {
	fs := make([]git.FileStat, len(files))
	for i, f := range files {
		fs[i] = git.FileStat{Filename: f, Insertions: 1}
	}
	return git.Commit{Author: author, FileStats: fs}
}

// TestOthersPressure_ExcludesSelf is the heart of others-contested survival: the
// pressure a module sees must EXCLUDE the queried author's own commits. A solo
// author who churns their own module alone produces zero others-pressure there,
// so their surviving lines read as dormant (not robust) — exactly the signal that
// stops untouched/self-churned code from minting gravity.
func TestOthersPressure_ExcludesSelf(t *testing.T) {
	// core: alice 5 commits, bob 2 commits; 10 surviving blame lines.
	commits := []git.Commit{}
	for i := 0; i < 5; i++ {
		commits = append(commits, commitTouching("alice", "core/a.go"))
	}
	for i := 0; i < 2; i++ {
		commits = append(commits, commitTouching("bob", "core/b.go"))
	}
	blame := make([]git.BlameLine, 10)
	for i := range blame {
		blame[i] = git.BlameLine{Filename: "core/a.go"}
	}

	o := CalcOthersPressure(commits, blame, ModuleResolver{})

	cases := []struct {
		author string
		want   float64
	}{
		{"alice", float64(7-5) / 10}, // others = bob's 2 → 0.2
		{"bob", float64(7-2) / 10},   // others = alice's 5 → 0.5
		{"carol", float64(7-0) / 10}, // non-contributor sees all 7 → 0.7
	}
	for _, c := range cases {
		if got := o.For("core", c.author); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("For(core, %s) = %v, want %v", c.author, got, c.want)
		}
	}
}

// TestOthersPressure_SoloAuthorZero: a module touched only by one author yields
// zero others-pressure for that author, no matter how many times they churn it.
func TestOthersPressure_SoloAuthorZero(t *testing.T) {
	commits := []git.Commit{}
	for i := 0; i < 20; i++ {
		commits = append(commits, commitTouching("solo", "lib/x.go"))
	}
	blame := make([]git.BlameLine, 4)
	for i := range blame {
		blame[i] = git.BlameLine{Filename: "lib/x.go"}
	}

	o := CalcOthersPressure(commits, blame, ModuleResolver{})
	if got := o.For("lib", "solo"); got != 0 {
		t.Errorf("solo self-churn must be zero others-pressure, got %v", got)
	}
	// A different author sees the full churn as pressure: 20/4 = 5.
	if got := o.For("lib", "other"); math.Abs(got-5) > 1e-9 {
		t.Errorf("For(lib, other) = %v, want 5", got)
	}
}

// cosmeticTouch is a commit whose files register ZERO substantive changed lines
// (Insertions+Deletions == 0) — the shape git.FileStat takes for a pure rename,
// a comment/doc-only edit, or a blank-only change after comment-exclusion.
func cosmeticTouch(author string, files ...string) git.Commit {
	fs := make([]git.FileStat, len(files))
	for i, f := range files {
		fs[i] = git.FileStat{Filename: f, Insertions: 0, Deletions: 0}
	}
	return git.Commit{Author: author, FileStats: fs}
}

// TestOthersPressure_CosmeticTouchDoesNotContest is the gaming-resistance gate:
// a module solo-owned by author A must NOT read as "others-contested" just because
// another author B ran a rename / comment / whitespace-only commit over it. Only a
// SUBSTANTIVE change by B may lift others-pressure above zero — otherwise a solo
// author's untouched lines mint RobustSurvival (and gravity) with no real contest.
func TestOthersPressure_CosmeticTouchDoesNotContest(t *testing.T) {
	// A solo-owns core with 10 surviving lines and 5 real commits.
	base := []git.Commit{}
	for i := 0; i < 5; i++ {
		base = append(base, commitTouching("alice", "core/a.go"))
	}
	blame := make([]git.BlameLine, 10)
	for i := range blame {
		blame[i] = git.BlameLine{Filename: "core/a.go"}
	}

	// Before: B only makes a cosmetic (0/0) touch to core — NOT a contest.
	cosmetic := append(append([]git.Commit{}, base...),
		cosmeticTouch("bob", "core/a.go"),
		cosmeticTouch("bob", "core/a.go"),
	)
	o := CalcOthersPressure(cosmetic, blame, ModuleResolver{})
	if got := o.For("core", "alice"); got != 0 {
		t.Fatalf("cosmetic touch by bob must NOT contest core; For(core, alice) = %v, want 0", got)
	}

	// After: B makes a SUBSTANTIVE commit (real changed lines) to core — a real
	// contest, so others-pressure for alice rises above zero.
	substantive := append(append([]git.Commit{}, base...),
		commitTouching("bob", "core/a.go"),
		commitTouching("bob", "core/a.go"),
	)
	o2 := CalcOthersPressure(substantive, blame, ModuleResolver{})
	if got := o2.For("core", "alice"); got <= 0 {
		t.Fatalf("substantive contest by bob must lift others-pressure; For(core, alice) = %v, want >0", got)
	}
	// Exactly bob's 2 substantive commits over 10 surviving lines.
	if got, want := o2.For("core", "alice"), float64(2)/10; math.Abs(got-want) > 1e-9 {
		t.Errorf("For(core, alice) = %v, want %v", got, want)
	}
}

// TestOthersPressure_NilSafe: a nil receiver returns 0 (classic mode passes nil).
func TestOthersPressure_NilSafe(t *testing.T) {
	var o *OthersPressure
	if got := o.For("anything", "anyone"); got != 0 {
		t.Errorf("nil OthersPressure must return 0, got %v", got)
	}
}

// TestPressureThreshold gates the robust/dormant split on the number of
// SUBSTANTIAL authors (absolute footprint), not on ownership share. The crucial
// case is the distributed repo: many authors, none owning even 10% of the repo,
// must still get a real (non-Inf) threshold — the old share-based gate zeroed
// robust survival for everyone in large collaborative repos.
func TestPressureThreshold(t *testing.T) {
	pressure := ChangePressure{"a": 1, "b": 2, "c": 3} // median 2

	t.Run("solo-dominated → Inf (all dormant)", func(t *testing.T) {
		blame := map[string]int{"solo": 50000, "minor": 50} // only 1 substantial
		if got := PressureThreshold(pressure, blame, SubstantialAuthorLines); !math.IsInf(got, 1) {
			t.Errorf("solo-dominated repo must yield +Inf, got %v", got)
		}
	})

	t.Run("two substantial authors → median", func(t *testing.T) {
		blame := map[string]int{"alice": 5000, "bob": 3000, "drive-by": 10}
		if got := PressureThreshold(pressure, blame, SubstantialAuthorLines); got != 2 {
			t.Errorf("two substantial authors must yield median (2), got %v", got)
		}
	})

	t.Run("distributed repo, nobody owns 10% → median (regression)", func(t *testing.T) {
		// 200 authors each with 1500 surviving lines: total 300k, so each owns
		// 0.5% — far below the old 10% bar — yet the split must be meaningful.
		blame := make(map[string]int, 200)
		for i := 0; i < 200; i++ {
			blame[string(rune('a'+i%26))+string(rune('0'+i/26))] = 1500
		}
		if got := PressureThreshold(pressure, blame, SubstantialAuthorLines); math.IsInf(got, 1) {
			t.Fatalf("distributed repo must NOT be gated to Inf; got Inf")
		}
		if got := PressureThreshold(pressure, blame, SubstantialAuthorLines); got != 2 {
			t.Errorf("distributed repo must yield median (2), got %v", got)
		}
	})
}
