package metric

import (
	"testing"

	"github.com/machuz/eis/v2/internal/git"
)

// This file pre-registers ADVERSARIAL fixtures against the contest detector —
// others-contested survival, the single load-bearing gate that stops churn from
// minting gravity. Each test encodes a specific attack and the outcome the gate
// MUST produce. The point is to VERIFY, not assume, that "shipped" means "not
// gamed": the DORA-death thesis applies to our own metric — whatever we measure,
// someone optimizes, and the day a sludge/durability number is sold to a CTO is
// the day their own AI agents start attacking it.
//
// The gate has two independent teeth, both exercised here:
//  1. self-exclusion — For(mod, author) counts only OTHER authors' commits, so
//     churning your own code (or an AI refreshing its own output) never contests it.
//  2. substance gate — a cosmetic touch (0 substantive lines: pure rename, blank-
//     or comment-only) never marks a module contested (MinSubstantiveLines).
//
// Layer note: these run at the metric layer on synthetic FileStats, which is
// exactly where self-exclusion + the substance gate live. Whitespace-only churn
// that git still reports as >=1 changed line is NOT caught here — that defense is
// parse-time (comment filter / numstat) and must be verified there, not faked at
// this layer. See TestContest_OneLineFloor_IsThePinnedResidual.

// cosmeticTouch (0 insertions + 0 deletions: pure rename / file move / blank- or
// comment-only edit) is defined in changepressure_test.go and reused here.

func blameLines(file string, n int) []git.BlameLine {
	b := make([]git.BlameLine, n)
	for i := range b {
		b[i] = git.BlameLine{Filename: file}
	}
	return b
}

// Attack 1 — AI self-regeneration loop. An agent repeatedly rewrites its OWN
// module (substantively) to keep its blame fresh and its survival young. With no
// other author ever contesting it, self-exclusion must read it as dormant:
// refreshing your own code — human or AI — can never mint robust survival. This
// is the AI-native attack, and the one most likely to arrive at scale.
func TestContest_AISelfRegeneration_StaysDormant(t *testing.T) {
	var commits []git.Commit
	for i := 0; i < 200; i++ {
		commits = append(commits, commitTouching("ai-agent", "gen/module.go"))
	}
	o := CalcOthersPressure(commits, blameLines("gen/module.go", 500), ModuleResolver{})
	if got := o.For("gen", "ai-agent"); got != 0 {
		t.Fatalf("AI self-regeneration minted others-pressure %v; want 0 (self-churn is never contested)", got)
	}
}

// Attack 2 — cosmetic collusion. The attacker solo-churns a module substantively,
// then a colluder rubber-stamps it with many 0-line cosmetic touches to fake
// "others contested this". The substance gate must drop every cosmetic touch, so
// the attacker's survival stays dormant — you cannot buy contest with rename PRs.
// The second half proves the gate blocks cosmetic touches, not the colleague:
// one REAL substantive touch by the same buddy does legitimately register.
func TestContest_CosmeticCollusion_DoesNotFakeContest(t *testing.T) {
	var commits []git.Commit
	for i := 0; i < 20; i++ {
		commits = append(commits, commitTouching("attacker", "feature/x.go"))
	}
	for i := 0; i < 100; i++ {
		commits = append(commits, cosmeticTouch("buddy", "feature/x.go"))
	}
	blame := blameLines("feature/x.go", 50)

	o := CalcOthersPressure(commits, blame, ModuleResolver{})
	if got := o.For("feature", "attacker"); got != 0 {
		t.Fatalf("cosmetic collusion faked others-pressure %v; want 0 (cosmetic touches are not contest)", got)
	}

	commits = append(commits, commitTouching("buddy", "feature/x.go"))
	o = CalcOthersPressure(commits, blame, ModuleResolver{})
	if got := o.For("feature", "attacker"); got == 0 {
		t.Fatal("one real substantive contest by buddy should register; the gate is over-blocking, not just anti-cosmetic")
	}
}

// (Pure rename / file-move contest is already covered by
// TestOthersPressure_CosmeticTouchDoesNotContest in changepressure_test.go —
// cosmeticTouch is exactly a rename/blank/comment-only edit.)

// Boundary — the substance floor is MinSubstantiveLines (== 1). This pins the
// load-bearing weak point: a colluder making many *1-line* commits DOES generate
// others-pressure. That is intended at this layer (a 1-line change is
// substantive), but it is the cheapest remaining contest-forgery — so it is
// pinned here explicitly. If MinSubstantiveLines is ever raised, or a
// whitespace/comment normalization changes what counts as 1 line, this test
// forces a conscious update rather than a silent shift in the gate's strength.
func TestContest_OneLineFloor_IsThePinnedResidual(t *testing.T) {
	if MinSubstantiveLines != 1 {
		t.Fatalf("MinSubstantiveLines = %d; this pinned residual assumed 1 — re-derive the attack cost before changing it", MinSubstantiveLines)
	}
	var commits []git.Commit
	for i := 0; i < 20; i++ {
		commits = append(commits, commitTouching("attacker", "mod/x.go"))
	}
	for i := 0; i < 5; i++ {
		commits = append(commits, commitTouching("colluder", "mod/x.go")) // 1-line each
	}
	o := CalcOthersPressure(commits, blameLines("mod/x.go", 50), ModuleResolver{})
	if got := o.For("mod", "attacker"); got == 0 {
		t.Fatal("the 1-line substance floor should admit real (if tiny) contest; behavior changed — the gate got stricter")
	}
}
