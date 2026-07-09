package cli

import (
	"strings"
	"testing"

	"github.com/machuz/eis/v2/internal/output"
)

// stubDigest returns a fixed digest so buildAnchors tests need no filesystem.
func stubDigest(abs string) (string, [2]int) {
	return "package x\nfunc H() {}", [2]int{1, 2}
}

func TestBuildAnchors_RanksBySurvivalTimesGravityAndGatesContested(t *testing.T) {
	stats := []AnchorStat{
		// module svc/order:
		// high survival + high mass, contested -> top anchor.
		{Module: "svc/order", File: "svc/order/order.go", AbsPath: "/r/svc/order/order.go", DecayedMass: 90, Lines: 100, Contributors: 4},
		// lower mass -> ranked below.
		{Module: "svc/order", File: "svc/order/util.go", AbsPath: "/r/svc/order/util.go", DecayedMass: 20, Lines: 40, Contributors: 3},
		// single-contributor -> excluded by the robust gate even though heavy.
		{Module: "svc/order", File: "svc/order/solo.go", AbsPath: "/r/svc/order/solo.go", DecayedMass: 200, Lines: 210, Contributors: 1},
		// another module, contested -> its own anchor list.
		{Module: "svc/pay", File: "svc/pay/pay.go", AbsPath: "/r/svc/pay/pay.go", DecayedMass: 50, Lines: 60, Contributors: 2},
	}

	r := buildAnchors(stats, 3, stubDigest)

	if len(r.Modules) != 2 {
		t.Fatalf("modules = %d, want 2 (svc/order, svc/pay)", len(r.Modules))
	}
	// Modules sorted by name.
	if r.Modules[0].Module != "svc/order" || r.Modules[1].Module != "svc/pay" {
		t.Fatalf("module order = %v", []string{r.Modules[0].Module, r.Modules[1].Module})
	}

	order := r.Modules[0]
	// solo.go must be gated out (single contributor); 2 anchors remain.
	if len(order.Anchors) != 2 {
		t.Fatalf("svc/order anchors = %d, want 2 (solo.go gated)", len(order.Anchors))
	}
	if order.Anchors[0].File != "svc/order/order.go" {
		t.Errorf("top anchor = %s, want svc/order/order.go (higher survival×gravity)", order.Anchors[0].File)
	}
	for _, a := range order.Anchors {
		if a.File == "svc/order/solo.go" {
			t.Error("single-contributor file must not be an anchor")
		}
	}

	top := order.Anchors[0]
	if top.ContestedByN != 4 {
		t.Errorf("contested_by_n = %d, want 4", top.ContestedByN)
	}
	if top.Survival != 0.9 { // 90/100
		t.Errorf("survival = %v, want 0.9", top.Survival)
	}
	if top.Gravity != 90 {
		t.Errorf("gravity = %v, want 90", top.Gravity)
	}
	if top.LineRange != [2]int{1, 2} || top.Digest == "" {
		t.Errorf("digest/line_range not wired: %+v", top)
	}
}

func TestBuildAnchors_TopNCap(t *testing.T) {
	var stats []AnchorStat
	for _, f := range []string{"a", "b", "c", "d", "e"} {
		stats = append(stats, AnchorStat{Module: "m", File: "m/" + f + ".go", DecayedMass: 10, Lines: 10, Contributors: 2})
	}
	r := buildAnchors(stats, 2, stubDigest)
	if len(r.Modules) != 1 || len(r.Modules[0].Anchors) != 2 {
		t.Fatalf("top=2 not honored: %+v", r.Modules)
	}
}

// Firewall: no author identity in the serialized anchors report.
func TestBuildAnchors_NoOwnerNames(t *testing.T) {
	stats := []AnchorStat{
		{Module: "svc/order", File: "svc/order/order.go", DecayedMass: 90, Lines: 100, Contributors: 4},
	}
	r := buildAnchors(stats, 3, func(abs string) (string, [2]int) {
		// Even a digest with a name-shaped token: firewall is about the STRUCT
		// fields, and contested_by_n is a count. (Digest is code, not authorship.)
		return "func H() {}", [2]int{1, 1}
	})
	var buf strings.Builder
	if err := output.EncodeAnchorsJSON(&buf, r); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, tok := range []string{"author", "owner", "top_author", "contributor\"", "\"Alice\"", "email"} {
		if strings.Contains(s, tok) {
			t.Errorf("anchors leaked identity-shaped field %q:\n%s", tok, s)
		}
	}
	// It DOES carry the anonymous count + weights.
	if !strings.Contains(s, `"contested_by_n": 4`) || !strings.Contains(s, `"gravity"`) {
		t.Errorf("expected anonymous count + weights:\n%s", s)
	}
}

func TestBuildAnchors_EmptyIsWellFormed(t *testing.T) {
	r := buildAnchors(nil, 3, stubDigest)
	var buf strings.Builder
	if err := output.EncodeAnchorsJSON(&buf, r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"modules": []`) {
		t.Errorf("empty report should serialize modules: [], got %s", buf.String())
	}
}
