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

// TestOthersPressure_NilSafe: a nil receiver returns 0 (classic mode passes nil).
func TestOthersPressure_NilSafe(t *testing.T) {
	var o *OthersPressure
	if got := o.For("anything", "anyone"); got != 0 {
		t.Errorf("nil OthersPressure must return 0, got %v", got)
	}
}
