package metric

import (
	"math"
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

func TestCalcFileSurvival(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tau := 180.0
	mr := NewModuleResolver([]string{"svc/*"})

	// 2 lines in svc/order/a.go, one fresh (age 0 -> weight 1) one old (age=tau
	// -> weight e^-1), two distinct authors. 1 line elsewhere, single author.
	lines := []git.BlameLine{
		{Filename: "svc/order/a.go", Author: "x", CommitterTime: now},
		{Filename: "svc/order/a.go", Author: "y", CommitterTime: now.AddDate(0, 0, -180)},
		{Filename: "svc/pay/b.go", Author: "x", CommitterTime: now},
	}
	got := CalcFileSurvival(lines, tau, now, mr)

	a := got["svc/order/a.go"]
	if a == nil {
		t.Fatal("missing svc/order/a.go")
	}
	if a.Module != "svc/order" {
		t.Errorf("module = %q, want svc/order", a.Module)
	}
	if a.Lines != 2 {
		t.Errorf("lines = %d, want 2", a.Lines)
	}
	if a.Contributors != 2 {
		t.Errorf("contributors = %d, want 2", a.Contributors)
	}
	wantMass := 1.0 + math.Exp(-1) // fresh + one-tau-old
	if math.Abs(a.DecayedMass-wantMass) > 1e-9 {
		t.Errorf("decayed mass = %v, want %v", a.DecayedMass, wantMass)
	}

	b := got["svc/pay/b.go"]
	if b == nil || b.Contributors != 1 || b.Lines != 1 {
		t.Errorf("svc/pay/b.go wrong: %+v", b)
	}
}

// Co-authored lines count each contributor toward the (anonymous) count.
func TestCalcFileSurvival_CoAuthorsCounted(t *testing.T) {
	now := time.Now()
	mr := NewModuleResolver(nil)
	lines := []git.BlameLine{
		{Filename: "a/b/f.go", Author: "x", Authors: []string{"x", "y", "z"}, CommitterTime: now},
	}
	got := CalcFileSurvival(lines, 180, now, mr)
	if got["a/b/f.go"].Contributors != 3 {
		t.Errorf("contributors = %d, want 3 (co-authors)", got["a/b/f.go"].Contributors)
	}
}
