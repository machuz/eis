package scorer

import "testing"

// baseFossil returns a Result that satisfies the gating conditions for Fragile
// (solo owner, low production). Tests vary the dormant / untested signals to
// see how the Fragile confidence moves.
func baseFossil() Result {
	return Result{
		Indispensability: 80,
		Production:       10,
		Quality:          40,
		Survival:         70,
		RawRobustSurv:    0,
		RawDormantSurv:   0,
		RawTestedSurv:    0,
		RawUntestedSurv:  0,
	}
}

func classifyFragile(r Result) AxisMatch {
	return classifyState(r)
}

// Pure dormant with no coverage signal → classic Fragile band (0.85–0.95).
func TestFragile_DormantOnly(t *testing.T) {
	r := baseFossil()
	r.RawDormantSurv = 95 // 95% dormant
	r.RawRobustSurv = 5
	m := classifyFragile(r)
	if m.Name != "Fragile" {
		t.Fatalf("expected Fragile, got %q (conf=%v)", m.Name, m.Confidence)
	}
	if m.Confidence < 0.85 || m.Confidence > 0.95 {
		t.Errorf("dormant-only confidence = %v, want within [0.85, 0.95]", m.Confidence)
	}
}

// Dormant + strong untested signal → boosted band (>=0.90).
func TestFragile_DormantAndUntested(t *testing.T) {
	r := baseFossil()
	r.RawDormantSurv = 95
	r.RawRobustSurv = 5
	r.RawUntestedSurv = 90 // 90% untested
	r.RawTestedSurv = 10
	m := classifyFragile(r)
	if m.Name != "Fragile" {
		t.Fatalf("expected Fragile, got %q", m.Name)
	}
	if m.Confidence < 0.90 {
		t.Errorf("dormant+untested confidence = %v, want ≥ 0.90", m.Confidence)
	}
}

// No pressure data but high untested ratio + high survival → still Fragile,
// slightly softer band than dormant-confirmed.
func TestFragile_UntestedOnly(t *testing.T) {
	r := baseFossil()
	r.Survival = 80
	r.RawUntestedSurv = 90
	r.RawTestedSurv = 10
	m := classifyFragile(r)
	if m.Name != "Fragile" {
		t.Fatalf("expected Fragile, got %q (conf=%v)", m.Name, m.Confidence)
	}
	if m.Confidence < 0.80 || m.Confidence >= 0.90 {
		t.Errorf("untested-only confidence = %v, want within [0.80, 0.90)", m.Confidence)
	}
}

// Mostly-tested code in a dormant module is NOT Fragile: tests protect it.
// With high coverage, Fragile shouldn't fire even though the module sleeps.
// It should fall back to the permissive fallback path (Survival high, Prod low)
// but the user remains potentially classifiable as some State.
func TestFragile_DormantButWellTested(t *testing.T) {
	r := baseFossil()
	r.RawDormantSurv = 95
	r.RawRobustSurv = 5
	r.RawUntestedSurv = 10 // only 10% untested
	r.RawTestedSurv = 90
	m := classifyFragile(r)
	// Still flagged Fragile because dormant ratio alone is enough; however
	// its confidence sits on the dormant-only band (≤0.95), not the boosted
	// dormant+untested band (≥0.96). This keeps behaviour backwards-compatible
	// while making room for stronger signals.
	if m.Name != "Fragile" {
		t.Fatalf("expected Fragile (dormant suffices), got %q", m.Name)
	}
	if m.Confidence > 0.95 {
		t.Errorf("dormant+tested must NOT reach boosted band: got %v", m.Confidence)
	}
}

// Active producer is never Fragile regardless of dormant or untested ratio.
func TestFragile_ActiveProducerGated(t *testing.T) {
	r := baseFossil()
	r.Production = 80 // active builder
	r.RawDormantSurv = 95
	r.RawRobustSurv = 5
	r.RawUntestedSurv = 95
	m := classifyFragile(r)
	if m.Name == "Fragile" {
		t.Errorf("active producer must not be Fragile, got %q (conf=%v)", m.Name, m.Confidence)
	}
}

// Solo-ownership gate: a low-indispensability author shouldn't be Fragile
// even with dormant+untested — those code bits aren't theirs alone.
func TestFragile_SharedOwnershipGated(t *testing.T) {
	r := baseFossil()
	r.Indispensability = 30 // shared owner
	r.RawDormantSurv = 95
	r.RawRobustSurv = 5
	r.RawUntestedSurv = 95
	r.RawTestedSurv = 5
	m := classifyFragile(r)
	if m.Name == "Fragile" && m.Confidence >= 0.80 {
		t.Errorf("shared ownership must not produce high-confidence Fragile: %v", m.Confidence)
	}
}

// Boosted band must be capped at 0.97 even when both ratios are 100%.
// (PR #19 review — reviewer noticed the formula could reach 1.0.)
func TestFragile_BoostedConfidenceCappedAt097(t *testing.T) {
	r := baseFossil()
	r.RawDormantSurv = 100 // 100% dormant
	r.RawRobustSurv = 0
	r.RawUntestedSurv = 100 // 100% untested
	r.RawTestedSurv = 0
	m := classifyFragile(r)
	if m.Name != "Fragile" {
		t.Fatalf("expected Fragile, got %q", m.Name)
	}
	if m.Confidence > 0.97 {
		t.Errorf("boosted confidence exceeded cap: %v > 0.97", m.Confidence)
	}
	if m.Confidence < 0.90 {
		t.Errorf("boosted band must be ≥ 0.90, got %v", m.Confidence)
	}
}

// Untested-only fallback must ONLY fire when pressure data is absent.
// Previously this path fired even with pressure present + non-dormant —
// producing false positives for active untested modules.
// (PR #19 review — high-priority issue.)
//
// We quality-gate (Quality ≥ 70) so the final fallback returns 0, isolating
// the behaviour of the three Fragile bands. With the fix in place, none of
// the three bands should fire because the untested-only path is now
// !hasPressure-gated.
func TestFragile_ActiveUntestedNotFragile(t *testing.T) {
	r := baseFossil()
	r.Quality = 70 // neutralise the permissive final fallback
	// Active module: mostly robust (change pressure present, code is not dormant).
	r.RawRobustSurv = 80
	r.RawDormantSurv = 20 // dormant ratio = 20%, well below 80%
	r.RawUntestedSurv = 90
	r.RawTestedSurv = 10
	r.Survival = 80
	m := classifyFragile(r)
	if m.Name == "Fragile" {
		t.Errorf("active-but-untested module wrongly flagged Fragile with conf %v", m.Confidence)
	}
}

func TestComputeUntestedRatio(t *testing.T) {
	cases := []struct {
		tested, untested float64
		wantRatio        float64
		wantOk           bool
	}{
		{0, 0, 0, false},
		{50, 50, 50, true},
		{100, 0, 0, true},
		{0, 100, 100, true},
		{30, 70, 70, true},
	}
	for _, c := range cases {
		r := Result{RawTestedSurv: c.tested, RawUntestedSurv: c.untested}
		got, ok := computeUntestedRatio(r)
		if ok != c.wantOk {
			t.Errorf("tested=%v untested=%v: ok=%v, want %v", c.tested, c.untested, ok, c.wantOk)
			continue
		}
		if ok && got != c.wantRatio {
			t.Errorf("tested=%v untested=%v: ratio=%v, want %v", c.tested, c.untested, got, c.wantRatio)
		}
	}
}
