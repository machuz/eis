package scorer

import "testing"

// TestGravityScore_Formula pins the weighted composition so a future reweight is
// a deliberate, reviewed change rather than an accident.
func TestGravityScore_Formula(t *testing.T) {
	r := Result{Catalysis: 80, RobustSurvival: 60, Design: 50, Breadth: 40, Indispensability: 20}
	// 80*.30 + 60*.25 + 50*.20 + 40*.15 + 20*.10 = 24+15+10+6+2 = 57
	if got := gravityScore(r, true); got != 57 {
		t.Fatalf("gravityScore = %v, want 57", got)
	}
}

// TestGravityScore_FallsBackToSurvivalWithoutPressureData: when an analysis has
// no robust/dormant split, RobustSurvival is 0, so the survival term reads total
// Survival instead — otherwise gravity would silently lose its whole survival
// component (the Impact formula falls back the same way).
func TestGravityScore_FallsBackToSurvivalWithoutPressureData(t *testing.T) {
	r := Result{Catalysis: 80, Survival: 60, RobustSurvival: 0, Design: 50, Breadth: 40, Indispensability: 20}
	// With no pressure data the 0.25 term uses Survival (60): same 57 as above.
	if got := gravityScore(r, false); got != 57 {
		t.Fatalf("fallback gravityScore = %v, want 57 (Survival used for the survival term)", got)
	}
	// With pressure data the same row scores lower — RobustSurvival (0) is used.
	if got := gravityScore(r, true); got != 42 {
		t.Fatalf("with pressure data gravityScore = %v, want 42 (RobustSurvival=0)", got)
	}
}

// TestGravityScore_DormantSurvivalExcluded: gravity reads RobustSurvival only.
// DormantSurvival (code resting in quiet modules) must not move gravity at all —
// durable, but it exerts no pull on what others build.
func TestGravityScore_DormantSurvivalExcluded(t *testing.T) {
	base := Result{Catalysis: 50, RobustSurvival: 40, Design: 30, Breadth: 20, Indispensability: 10}
	withDormant := base
	withDormant.DormantSurvival = 100
	if gravityScore(base, true) != gravityScore(withDormant, true) {
		t.Fatal("DormantSurvival must not contribute to gravity")
	}
}

// TestGravityScore_IsRelational: gravity cannot be observed from one person
// alone. A solo author who maxes every axis they CAN reach by themselves
// (Design / Breadth / Indispensability) but has no collaborators (Catalysis = 0)
// and no code proven under shared change pressure (RobustSurvival = 0) must land
// far below the ceiling — and below a genuine catalyst who scores zero on those
// solo-inflatable axes.
func TestGravityScore_IsRelational(t *testing.T) {
	soloOwner := Result{Catalysis: 0, RobustSurvival: 0, Design: 100, Breadth: 100, Indispensability: 100}
	catalyst := Result{Catalysis: 100, RobustSurvival: 100, Design: 0, Breadth: 0, Indispensability: 0}

	if g := gravityScore(soloOwner, true); g >= 50 {
		t.Fatalf("a solo owner should be capped well below the ceiling, got %v", g)
	}
	if gravityScore(catalyst, true) <= gravityScore(soloOwner, true) {
		t.Fatalf("a catalyst (%v) must out-gravitate a solo owner (%v)",
			gravityScore(catalyst, true), gravityScore(soloOwner, true))
	}
}

// TestGravityScore_IndispensabilityNoLongerDominates: owning everything alone
// used to mint 0.40 of gravity; it is now 0.10, and is dwarfed by Catalysis
// (0.30). This is the fix for one-person projects scoring unduly high gravity.
func TestGravityScore_IndispensabilityNoLongerDominates(t *testing.T) {
	onlyIndisp := Result{Indispensability: 100}
	onlyCatalysis := Result{Catalysis: 100}
	if gravityScore(onlyIndisp, true) != 10 {
		t.Fatalf("Indispensability lever = %v, want 10", gravityScore(onlyIndisp, true))
	}
	if gravityScore(onlyCatalysis, true) <= gravityScore(onlyIndisp, true) {
		t.Fatal("Catalysis must outweigh Indispensability as a gravity lever")
	}
}
