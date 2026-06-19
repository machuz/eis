package metric

import (
	"math"
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

// commitOn: author touches a file at a given date.
func commitOn(author, file string, date time.Time) git.Commit {
	return git.Commit{
		Author:    author,
		Date:      date,
		FileStats: []git.FileStat{{Filename: file, Insertions: 10}},
	}
}

func blameOn(author, file string, t time.Time) git.BlameLine {
	return git.BlameLine{Filename: file, Author: author, CommitterTime: t}
}

// alice lays a foundation that still survives, bob builds on it later and also
// survives → alice is the catalyst, bob (no one came after him) is not.
func TestCatalysis_CreditsSurvivingFoundationOthersBuiltOn(t *testing.T) {
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	tau := 180.0
	commits := []git.Commit{
		commitOn("alice", "f.go", now.AddDate(0, -3, 0)), // alice first
		commitOn("bob", "f.go", now.AddDate(0, -1, 0)),   // bob later
	}
	blame := []git.BlameLine{
		blameOn("alice", "f.go", now.AddDate(0, 0, -20)), // alice's foundation survives
		blameOn("bob", "f.go", now.AddDate(0, 0, -10)),
		blameOn("bob", "f.go", now.AddDate(0, 0, -10)),
	}
	s := CalcCatalysis(commits, blame, tau, now)
	if s["alice"] <= 0 {
		t.Errorf("alice's foundation survives and bob built on it later — expected alice > 0, got %f", s["alice"])
	}
	if s["bob"] != 0 {
		t.Errorf("no one arrived after bob — expected bob = 0, got %f", s["bob"])
	}
}

// THE KEY CASE: alice creates the file but bob rewrites it entirely (alice's
// lines vanish). alice is NOT the foundation — bob is. When carol later builds
// on bob's surviving rewrite, bob gets the catalysis, not the file-creator alice.
func TestCatalysis_RewriterBecomesFoundationNotFileCreator(t *testing.T) {
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	tau := 180.0
	commits := []git.Commit{
		commitOn("alice", "f.go", now.AddDate(0, -6, 0)), // alice created it first
		commitOn("bob", "f.go", now.AddDate(0, -4, 0)),   // bob rewrote it
		commitOn("carol", "f.go", now.AddDate(0, -1, 0)), // carol built on bob's rewrite
	}
	// alice's lines are GONE (fully rewritten). bob's and carol's survive.
	blame := []git.BlameLine{
		blameOn("bob", "f.go", now.AddDate(0, 0, -30)),
		blameOn("bob", "f.go", now.AddDate(0, 0, -30)),
		blameOn("carol", "f.go", now.AddDate(0, 0, -5)),
	}
	s := CalcCatalysis(commits, blame, tau, now)
	if s["alice"] != 0 {
		t.Errorf("alice's foundation was rewritten away — expected alice = 0, got %f", s["alice"])
	}
	if s["bob"] <= 0 {
		t.Errorf("bob's rewrite survives and carol built on it — bob should be the foundation, got %f", s["bob"])
	}
	if s["carol"] != 0 {
		t.Errorf("no one arrived after carol — expected carol = 0, got %f", s["carol"])
	}
}

// THE PARTIAL-REWRITE CASE (the gameable one): alice creates the file but bob
// rewrites most of it, leaving alice only a surviving remnant. alice must NOT be
// credited for bob's whole surviving rewrite just because she touched the file
// first — her credit is scaled down to her shrunken surviving share. (Old
// behaviour: alice got bob's full 9-line mass; now she gets ~0.9.)
func TestCatalysis_PartialRewriteScalesDownFileCreator(t *testing.T) {
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	tau := 180.0
	commits := []git.Commit{
		commitOn("alice", "f.go", now.AddDate(0, -6, 0)), // alice created it
		commitOn("bob", "f.go", now.AddDate(0, -3, 0)),   // bob rewrote most of it
	}
	// alice keeps 1 surviving line; bob's rewrite leaves 9 surviving lines.
	blame := []git.BlameLine{blameOn("alice", "f.go", now.AddDate(0, 0, -5))}
	for i := 0; i < 9; i++ {
		blame = append(blame, blameOn("bob", "f.go", now.AddDate(0, 0, -5)))
	}
	s := CalcCatalysis(commits, blame, tau, now)
	// Far below bob's 9-line surviving mass: alice is a remnant, not the foundation.
	if s["alice"] >= 2 {
		t.Errorf("remnant file-creator must not collect the rewriter's full mass, got %f", s["alice"])
	}
	// But not zero — alice's remnant did survive and bob built around it.
	if s["alice"] <= 0 {
		t.Errorf("alice still has a surviving remnant, expected >0, got %f", s["alice"])
	}
}

// A solo file (only the author's lines, no later contributors) yields zero —
// catalysis requires others to build on you.
func TestCatalysis_SoloFileScoresZero(t *testing.T) {
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	commits := []git.Commit{commitOn("alice", "solo.go", now.AddDate(0, -1, 0))}
	blame := []git.BlameLine{
		blameOn("alice", "solo.go", now.AddDate(0, 0, -3)),
		blameOn("alice", "solo.go", now.AddDate(0, 0, -3)),
	}
	s := CalcCatalysis(commits, blame, 180, now)
	if s["alice"] != 0 {
		t.Errorf("solo authorship should score 0 Catalysis, got %f", s["alice"])
	}
}

// A blamed file with no commit history (e.g. vendored) has unknown precedence
// and must not produce credit.
func TestCatalysis_UnknownPrecedenceSkipped(t *testing.T) {
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	blame := []git.BlameLine{blameOn("bob", "vendor/lib.go", now.AddDate(0, 0, -1))}
	s := CalcCatalysis(nil, blame, 180, now)
	if len(s) != 0 {
		t.Errorf("unknown-precedence lines must not produce any score, got %v", s)
	}
}

// Later contributors' RECENT surviving work weighs more than old surviving work.
func TestCatalysis_DecaysWithAge(t *testing.T) {
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	tau := 180.0
	commits := []git.Commit{
		commitOn("alice", "f.go", now.AddDate(0, -6, 0)),
		commitOn("bob", "f.go", now.AddDate(0, -2, 0)),
	}
	aliceLine := blameOn("alice", "f.go", now.AddDate(0, 0, -1)) // alice foundation alive
	fresh := CalcCatalysis(commits, []git.BlameLine{aliceLine, blameOn("bob", "f.go", now.AddDate(0, 0, -1))}, tau, now)
	old := CalcCatalysis(commits, []git.BlameLine{aliceLine, blameOn("bob", "f.go", now.AddDate(0, -5, 0))}, tau, now)
	if !(fresh["alice"] > old["alice"]) {
		t.Errorf("recent build-on should weigh more than old: fresh=%f old=%f", fresh["alice"], old["alice"])
	}
}

// tau <= 0 must not divide-by-zero / NaN; it falls back to no decay (factor 1).
func TestCatalysis_NonPositiveTauNoDecay(t *testing.T) {
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	commits := []git.Commit{
		commitOn("alice", "f.go", now.AddDate(0, -3, 0)),
		commitOn("bob", "f.go", now.AddDate(0, -1, 0)),
	}
	blame := []git.BlameLine{
		blameOn("alice", "f.go", now.AddDate(0, 0, -20)),
		blameOn("bob", "f.go", now.AddDate(0, 0, -10)),
	}
	s := CalcCatalysis(commits, blame, 0, now)
	if math.IsNaN(s["alice"]) || math.IsInf(s["alice"], 0) {
		t.Fatalf("tau=0 produced non-finite Catalysis: %f", s["alice"])
	}
	// alice 1 surviving line, bob 1 surviving line → alice's share is 1/2, and
	// bob's one later surviving line (factor 1, no decay) is credited × that
	// share = 0.5.
	if s["alice"] != 0.5 {
		t.Errorf("tau=0, equal shares: expected 0.5 (bSurv 1 × share 0.5), got %f", s["alice"])
	}
}

// Precedence is by date, independent of commit slice order (determinism).
func TestCatalysis_PrecedenceByDateNotSliceOrder(t *testing.T) {
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	commits := []git.Commit{
		commitOn("bob", "x.go", now.AddDate(0, -1, 0)),   // listed first, dated later
		commitOn("alice", "x.go", now.AddDate(0, -3, 0)), // dated earlier → foundation
	}
	blame := []git.BlameLine{
		blameOn("alice", "x.go", now.AddDate(0, 0, -10)),
		blameOn("bob", "x.go", now.AddDate(0, 0, -5)),
	}
	s := CalcCatalysis(commits, blame, 180, now)
	if s["alice"] <= 0 {
		t.Errorf("alice (earlier date) is the foundation — expected alice > 0, got %f", s["alice"])
	}
	if s["bob"] != 0 {
		t.Errorf("bob arrived later with no one after him — expected 0, got %f", s["bob"])
	}
}
