package scorer

import (
	"math"
	"testing"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// TestGravityScore_ShapeWeights pins the shape composition. With both gates open
// (Catalysis = RobustSurvival = 100) gravity == shape, so each axis alone reveals
// its weight.
func TestGravityScore_ShapeWeights(t *testing.T) {
	open := func(field string, v float64) Result {
		r := Result{Catalysis: 100, RobustSurvival: 100}
		switch field {
		case "design":
			r.Design = v
		case "breadth":
			r.Breadth = v
		case "indisp":
			r.Indispensability = v
		}
		return r
	}
	for _, c := range []struct {
		f    string
		want float64
	}{{"design", 45}, {"breadth", 25}, {"indisp", 30}} {
		if got := gravityScore(open(c.f, 100)); !approx(got, c.want) {
			t.Errorf("%s alone: gravity = %v, want %v", c.f, got, c.want)
		}
	}
}

// TestGravityScore_SurvivalGate: RobustSurvival (others-contested survival) is a
// necessary gate. With shape and Catalysis maxed, gravity scales with
// RobustSurvival 0/50/100 → 15/57.5/100.
func TestGravityScore_SurvivalGate(t *testing.T) {
	full := func(s float64) Result {
		return Result{Catalysis: 100, RobustSurvival: s, Design: 100, Breadth: 100, Indispensability: 100}
	}
	for _, c := range []struct{ surv, want float64 }{{0, 15}, {50, 57.5}, {100, 100}} {
		if got := gravityScore(full(c.surv)); !approx(got, c.want) {
			t.Errorf("RobustSurvival %v: gravity = %v, want %v", c.surv, got, c.want)
		}
	}
}

// TestGravityScore_CatalysisGate: Catalysis is the other necessary gate.
func TestGravityScore_CatalysisGate(t *testing.T) {
	full := func(cat float64) Result {
		return Result{Catalysis: cat, RobustSurvival: 100, Design: 100, Breadth: 100, Indispensability: 100}
	}
	for _, c := range []struct{ cat, want float64 }{{0, 15}, {50, 57.5}, {100, 100}} {
		if got := gravityScore(full(c.cat)); !approx(got, c.want) {
			t.Errorf("Catalysis %v: gravity = %v, want %v", c.cat, got, c.want)
		}
	}
}

// TestGravityScore_RewrittenFoundationCollapses is the core property: a heavily
// shaped "foundation" (high Design + sole-ownership history) that did NOT survive
// under others' pressure — rewritten away, or sitting in a dead/self-churned
// corner, RobustSurvival → 0 — must collapse, while the same shape with code that
// survived where others work scores high. Gravity decays with the code, and is
// not minted by mere untouched persistence.
func TestGravityScore_RewrittenFoundationCollapses(t *testing.T) {
	rewritten := Result{Catalysis: 73, RobustSurvival: 0, Design: 100, Breadth: 20, Indispensability: 100}
	survived := rewritten
	survived.RobustSurvival = 100

	if g := gravityScore(rewritten); g >= 12 {
		t.Fatalf("a rewritten/uncontested foundation must collapse, got %v", g)
	}
	if g := gravityScore(survived); g <= 55 {
		t.Fatalf("the same shape that survived should score high, got %v", g)
	}
}

// TestGravityScore_RequiresBoth: gravity needs others-contested survival AND
// catalysis. Either one alone leaves it near the floor; only both together reach
// the top.
func TestGravityScore_RequiresBoth(t *testing.T) {
	shape := Result{Design: 100, Breadth: 100, Indispensability: 100} // shape = 100
	survOnly := shape
	survOnly.RobustSurvival = 100 // catalysis 0
	catOnly := shape
	catOnly.Catalysis = 100 // robust survival 0
	both := shape
	both.RobustSurvival, both.Catalysis = 100, 100

	gSurv, gCat, gBoth := gravityScore(survOnly), gravityScore(catOnly), gravityScore(both)
	if gSurv > 20 || gCat > 20 {
		t.Fatalf("one gate alone must stay low: survOnly=%v catOnly=%v", gSurv, gCat)
	}
	if gBoth <= 4*gSurv || gBoth <= 4*gCat {
		t.Fatalf("both gates open must dominate either alone: both=%v survOnly=%v catOnly=%v", gBoth, gSurv, gCat)
	}
}

// TestLifetimeGravity_ShapeWeights pins the indispensability-leaned shape. With
// both gates open (Catalysis = Survival = 100) lifetime == shape.
func TestLifetimeGravity_ShapeWeights(t *testing.T) {
	open := func(field string, v float64) Result {
		r := Result{Catalysis: 100, Survival: 100}
		switch field {
		case "design":
			r.Design = v
		case "breadth":
			r.Breadth = v
		case "indisp":
			r.Indispensability = v
		}
		return r
	}
	for _, c := range []struct {
		f    string
		want float64
	}{{"design", 35}, {"breadth", 20}, {"indisp", 45}} {
		if got := lifetimeGravityScore(open(c.f, 100)); !approx(got, c.want) {
			t.Errorf("%s alone: lifetime = %v, want %v", c.f, got, c.want)
		}
	}
}

// TestLifetimeGravity_SettledFoundationStands is the whole point: a foundation
// that everyone built on and that still STANDS, but which no one currently
// presses on (RobustSurvival → 0 while Survival stays high), reads quiet on
// Gravity yet high on LifetimeGravity. The React/Markbåge case in miniature.
func TestLifetimeGravity_SettledFoundationStands(t *testing.T) {
	markbage := Result{Catalysis: 100, Survival: 100, RobustSurvival: 0, Design: 24, Breadth: 100, Indispensability: 100}
	if g := gravityScore(markbage); g >= 12 {
		t.Fatalf("live Gravity should read quiet for a settled foundation, got %v", g)
	}
	if l := lifetimeGravityScore(markbage); l <= 60 {
		t.Fatalf("LifetimeGravity should surface the durable foundation, got %v", l)
	}
}

// TestLifetimeGravity_SelfChurnStillGated: dropping to raw Survival must NOT
// re-open the self-churn hole — a dead corner the author alone keeps alive has
// high Survival but no one builds on it (low Catalysis), so catGate keeps
// LifetimeGravity near the floor.
func TestLifetimeGravity_SelfChurnStillGated(t *testing.T) {
	selfChurn := Result{Catalysis: 5, Survival: 100, RobustSurvival: 0, Design: 100, Breadth: 100, Indispensability: 100}
	if l := lifetimeGravityScore(selfChurn); l >= 25 {
		t.Fatalf("self-churn (high survival, no catalysis) must stay gated, got %v", l)
	}
}
