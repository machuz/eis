package metric

import (
	"math"
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

func TestCalcModuleSurvivalByAuthor(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tau := 90.0
	mr := NewModuleResolver([]string{"services/*"})

	day := 24 * time.Hour
	lines := []git.BlameLine{
		{Author: "alice", Filename: "services/auth/main.go", CommitterTime: now.Add(-10 * day)},
		{Author: "alice", Filename: "services/auth/handler.go", CommitterTime: now.Add(-10 * day)},
		{Author: "bob", Filename: "services/auth/util.go", CommitterTime: now.Add(-200 * day)},
		{Author: "alice", Filename: "services/billing/charge.go", CommitterTime: now.Add(-30 * day)},
	}

	got := CalcModuleSurvivalByAuthor(lines, tau, now, mr)

	// Both modules and both authors are present.
	if _, ok := got["services/auth"]["alice"]; !ok {
		t.Fatalf("expected alice in services/auth")
	}
	if _, ok := got["services/auth"]["bob"]; !ok {
		t.Fatalf("expected bob in services/auth")
	}
	if _, ok := got["services/billing"]["alice"]; !ok {
		t.Fatalf("expected alice in services/billing")
	}

	// Mass is time-decayed: alice's two recent (10d) lines outweigh bob's one
	// old (200d) line in the same module.
	if got["services/auth"]["alice"] <= got["services/auth"]["bob"] {
		t.Errorf("recent holder should outweigh old holder: alice=%v bob=%v",
			got["services/auth"]["alice"], got["services/auth"]["bob"])
	}

	// Consistency with CalcModuleSurvival: summing per-author mass over a module
	// must equal that module's rate × its line count (both decompose the same
	// decayed-blame sum). Guards against the two functions drifting (W-02).
	rate := CalcModuleSurvival(lines, tau, now, mr)
	authMass := got["services/auth"]["alice"] + got["services/auth"]["bob"]
	wantAuthMass := rate["services/auth"] * 3 // 3 lines in services/auth
	if math.Abs(authMass-wantAuthMass) > 1e-9 {
		t.Errorf("per-author mass sum %v != rate*count %v", authMass, wantAuthMass)
	}
}
