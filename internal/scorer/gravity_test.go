package scorer

import (
	"math"
	"testing"
)

// approx compares gravity values, tolerating the float rounding the gate
// multiply introduces (e.g. 0.6×100 = 60.000000000001).
func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// TestGravityScore_FootprintWeights pins the footprint composition. With
// Catalysis = 100 the gate is 1.0, so gravity == footprint and each axis alone
// reveals its weight.
func TestGravityScore_FootprintWeights(t *testing.T) {
	cases := []struct {
		name string
		r    Result
		want float64
	}{
		{"robust", Result{Catalysis: 100, RobustSurvival: 100}, 35},
		{"design", Result{Catalysis: 100, Design: 100}, 25},
		{"breadth", Result{Catalysis: 100, Breadth: 100}, 20},
		{"indisp", Result{Catalysis: 100, Indispensability: 100}, 20},
	}
	for _, c := range cases {
		if got := gravityScore(c.r, true); !approx(got, c.want) {
			t.Errorf("%s: gravityScore = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestGravityScore_GateRampsWithCatalysis: Catalysis gates the footprint. With a
// maxed footprint (100), gravity scales 20 → 60 → 100 across Catalysis 0/50/100.
func TestGravityScore_GateRampsWithCatalysis(t *testing.T) {
	full := func(cat float64) Result {
		return Result{Catalysis: cat, RobustSurvival: 100, Design: 100, Breadth: 100, Indispensability: 100}
	}
	cases := []struct {
		cat, want float64
	}{{0, 20}, {50, 60}, {100, 100}}
	for _, c := range cases {
		if got := gravityScore(full(c.cat), true); !approx(got, c.want) {
			t.Errorf("Catalysis %v: gravity = %v, want %v", c.cat, got, c.want)
		}
	}
}

// TestGravityScore_SoloIsGated: gravity cannot be observed from one person alone.
// A solo author who maxes every footprint axis they can reach by themselves but
// has no one building on them (Catalysis = 0) is capped at the gate floor — far
// below the same footprint once collaborators build on it.
func TestGravityScore_SoloIsGated(t *testing.T) {
	footprint := Result{RobustSurvival: 100, Design: 100, Breadth: 100, Indispensability: 100}

	solo := footprint // Catalysis 0
	withCollaborators := footprint
	withCollaborators.Catalysis = 100

	if g := gravityScore(solo, true); g > 20 {
		t.Fatalf("a solo owner must cap at the gate floor (~20), got %v", g)
	}
	if gravityScore(withCollaborators, true) <= gravityScore(solo, true) {
		t.Fatalf("the same footprint with collaborators (%v) must out-gravitate solo (%v)",
			gravityScore(withCollaborators, true), gravityScore(solo, true))
	}
}

// TestGravityScore_DormantSurvivalExcluded: the footprint reads RobustSurvival
// only. DormantSurvival (code resting in quiet modules) must not move gravity.
func TestGravityScore_DormantSurvivalExcluded(t *testing.T) {
	base := Result{Catalysis: 100, RobustSurvival: 40, Design: 30, Breadth: 20, Indispensability: 10}
	withDormant := base
	withDormant.DormantSurvival = 100
	if gravityScore(base, true) != gravityScore(withDormant, true) {
		t.Fatal("DormantSurvival must not contribute to gravity")
	}
}

// TestGravityScore_UnprovenSurvivalIsStronglyDamped: without the robust/dormant
// split the survival term uses total Survival at a steep discount, never face
// value — survival we cannot confirm is "under pull" must not read as robust.
func TestGravityScore_UnprovenSurvivalIsStronglyDamped(t *testing.T) {
	// Catalysis 100 (gate 1.0) isolates the survival term. Survival 100 only.
	r := Result{Catalysis: 100, Survival: 100, RobustSurvival: 0}
	// No pressure data: footprint = (100*0.3)*0.35 = 30*0.35 = 10.5.
	if got := gravityScore(r, false); !approx(got, 10.5) {
		t.Fatalf("damped gravity = %v, want 10.5 (Survival 100 discounted to 30)", got)
	}
	// Face value would have been 35 — the damping must cost real points.
	if gravityScore(r, false) >= 35 {
		t.Fatal("unproven survival must be discounted below its robust-equivalent")
	}
	// With pressure data RobustSurvival (0) is used → footprint 0 → gravity 0.
	if got := gravityScore(r, true); got != 0 {
		t.Fatalf("with pressure data gravity = %v, want 0 (RobustSurvival=0)", got)
	}
}

// TestGravityScore_IndispensabilityDoesNotInflateSolo: sole ownership carries a
// real footprint weight (0.20), but because the gate suppresses solo work, owning
// everything alone (Catalysis 0) can no longer mint high gravity — the old failure
// mode where a one-person project scored top gravity from ownership alone.
func TestGravityScore_IndispensabilityDoesNotInflateSolo(t *testing.T) {
	soloOwner := Result{Indispensability: 100}                 // Catalysis 0
	if got := gravityScore(soloOwner, true); !approx(got, 4) { // 0.2 gate * (100*0.20) = 0.2*20 = 4
		t.Fatalf("solo sole-owner gravity = %v, want 4", got)
	}
}
