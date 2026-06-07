package metric

import (
	"math"
	"testing"
)

func approx(t *testing.T, label string, got, want, eps float64) {
	t.Helper()
	if math.Abs(got-want) > eps {
		t.Errorf("%s = %v, want ~%v", label, got, want)
	}
}

// Even spread across N modules → effective module count ≈ N.
func TestComputeBreadth_EvenSpread(t *testing.T) {
	in := map[string]map[string]float64{
		"m1": {"A": 1}, "m2": {"A": 1}, "m3": {"A": 1}, "m4": {"A": 1},
	}
	got := ComputeBreadth(in)
	approx(t, "A even-spread breadth", got["A"], 4.0, 1e-9)
}

// All gravity concentrated in one module → effective ≈ 1 (a specialist).
func TestComputeBreadth_Concentrated(t *testing.T) {
	in := map[string]map[string]float64{
		"m1": {"A": 42},
	}
	got := ComputeBreadth(in)
	approx(t, "A concentrated breadth", got["A"], 1.0, 1e-9)
}

// Survival-weighted: 80% in one module plus thin scraps elsewhere does NOT
// read as broad — the Hill number discounts the thin shares (this is the
// "broad but shallow" failure mode dissolving by construction).
func TestComputeBreadth_SkewedDiscountsThin(t *testing.T) {
	in := map[string]map[string]float64{
		"big":   {"A": 8},
		"scrap": {"A": 2},
	}
	// shares 0.8/0.2 → exp(-(0.8 ln0.8 + 0.2 ln0.2)) ≈ 1.649.
	got := ComputeBreadth(in)
	approx(t, "A skewed breadth", got["A"], 1.649, 1e-3)
	if got["A"] >= 2.0 {
		t.Errorf("skewed breadth should stay well under 2, got %v", got["A"])
	}
}

// A near-zero-mass module barely moves the effective count (diversity index,
// not a raw count) — adding a 0.01-share touch keeps breadth ≈ its prior value.
func TestComputeBreadth_TinyTouchBarelyCounts(t *testing.T) {
	base := ComputeBreadth(map[string]map[string]float64{
		"m1": {"A": 1}, "m2": {"A": 1},
	})
	withTouch := ComputeBreadth(map[string]map[string]float64{
		"m1": {"A": 1}, "m2": {"A": 1}, "m3": {"A": 0.01},
	})
	approx(t, "base breadth", base["A"], 2.0, 1e-9)
	if math.Abs(withTouch["A"]-2.0) > 0.1 {
		t.Errorf("a 0.01 touch should barely move breadth, got %v (base 2.0)", withTouch["A"])
	}
}

// Authors are independent; zero-gravity authors are omitted entirely.
func TestComputeBreadth_MultiAuthorAndZeroOmitted(t *testing.T) {
	in := map[string]map[string]float64{
		"m1": {"A": 1, "B": 5},
		"m2": {"A": 1},
		"m3": {"C": 0}, // zero mass → C never appears
	}
	got := ComputeBreadth(in)
	approx(t, "A breadth", got["A"], 2.0, 1e-9) // even across m1,m2
	approx(t, "B breadth", got["B"], 1.0, 1e-9) // only m1
	if _, present := got["C"]; present {
		t.Errorf("C should be omitted (zero gravity), got %v", got["C"])
	}
}
