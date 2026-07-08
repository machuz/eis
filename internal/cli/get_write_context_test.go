package cli

import (
	"strings"
	"testing"

	"github.com/machuz/eis/v2/internal/output"
)

// gwcFixtureIndex reuses the same shape as the precheck fixture: orphaned, dead,
// at-risk, and clean modules. Empty module_patterns ⇒ default 2-component
// resolution matches the keys.
func gwcFixtureIndex() output.WriteIndex {
	return output.WriteIndex{
		ModulePatterns: nil,
		Modules: map[string]output.WriteIndexModule{
			"svc/auth":    {DebtTier: "Orphaned", OwnerLeftDays: 243, UntouchedDays: 243, OwnershipConcentration: 0.9, Recommendation: "orphaned_module"},
			"legacy/vm":   {DebtTier: "Dead", UntouchedDays: 900, Recommendation: "dead_module"},
			"core/engine": {AtRisk: true, OwnerActive: true, OwnershipConcentration: 0.88, Recommendation: "bus_factor_1"},
			"core/api":    {OwnerActive: true, OwnershipConcentration: 0.3, Recommendation: ""}, // clean
		},
	}
}

func entryByPath(wc output.WriteContext, path string) (output.WriteContextEntry, bool) {
	for _, e := range wc.Modules {
		if e.Path == path {
			return e, true
		}
	}
	return output.WriteContextEntry{}, false
}

func TestBuildWriteContext_MultiPathAllCases(t *testing.T) {
	idx := gwcFixtureIndex()
	paths := []string{
		"svc/auth/login.go",    // orphaned
		"legacy/vm/run.go",     // dead
		"core/engine/core.go",  // at-risk
		"core/api/handler.go",  // clean (still returned)
		"brand/new/feature.go", // unknown module -> omitted
	}
	wc := buildWriteContext(paths, "", idx)

	// Clean IS returned; unknown is omitted.
	if len(wc.Modules) != 4 {
		t.Fatalf("modules = %d, want 4 (clean returned, unknown omitted)", len(wc.Modules))
	}
	if _, ok := entryByPath(wc, "brand/new/feature.go"); ok {
		t.Error("unknown module must be omitted")
	}

	orph, _ := entryByPath(wc, "svc/auth/login.go")
	if orph.Module != "svc/auth" || orph.DebtTier != "Orphaned" || orph.Recommendation != "orphaned_module" {
		t.Errorf("orphaned entry wrong: %+v", orph)
	}
	if orph.OwnerLeftDays != 243 || orph.UntouchedDays != 243 || orph.OwnershipConcentration != 0.9 {
		t.Errorf("orphaned facts wrong: %+v", orph)
	}
	if len(orph.Directives) != 4 || orph.Directives[0] != "Prefer a minimal diff" {
		t.Errorf("orphaned directives wrong: %v", orph.Directives)
	}
	if orph.SurvivingIdiomDigest != nil {
		t.Error("surviving_idiom_digest must be reserved (null)")
	}

	dead, _ := entryByPath(wc, "legacy/vm/run.go")
	if dead.DebtTier != "Dead" || len(dead.Directives) != 2 || dead.Directives[0] != "Consider deletion over extension" {
		t.Errorf("dead entry wrong: %+v", dead)
	}

	risk, _ := entryByPath(wc, "core/engine/core.go")
	if !risk.AtRisk || risk.Recommendation != "bus_factor_1" || len(risk.Directives) == 0 {
		t.Errorf("at-risk entry wrong: %+v", risk)
	}

	// Clean module: returned with empty tier + empty (non-nil) directives.
	clean, ok := entryByPath(wc, "core/api/handler.go")
	if !ok {
		t.Fatal("clean module must be returned (query API)")
	}
	if clean.DebtTier != "" || clean.Recommendation != "" {
		t.Errorf("clean entry should be empty debt: %+v", clean)
	}
	if clean.Directives == nil || len(clean.Directives) != 0 {
		t.Errorf("clean directives must be empty non-nil slice, got %v", clean.Directives)
	}
}

// Firewall: the serialized response must carry no owner identity.
func TestBuildWriteContext_NoOwnerNames(t *testing.T) {
	idx := gwcFixtureIndex()
	wc := buildWriteContext([]string{"svc/auth/x.go"}, "", idx)
	var buf strings.Builder
	if err := output.EncodeWriteContext(&buf, wc); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, tok := range []string{"tanaka", "Mark Erikson", "Lee Byron", "top_author", "last_owner", "owner_name", "@"} {
		if strings.Contains(s, tok) {
			t.Errorf("response leaked identity token %q:\n%s", tok, s)
		}
	}
	// Empty directives / null digest must serialize as [] / null.
	if !strings.Contains(s, `"directives": [`) {
		t.Errorf("expected directives array in output:\n%s", s)
	}
	if !strings.Contains(s, `"surviving_idiom_digest": null`) {
		t.Errorf("expected reserved null digest:\n%s", s)
	}
}

// Missing/empty index -> fail-open: every path omitted, empty modules list, no
// error. EncodeWriteContext must still emit "modules": [].
func TestBuildWriteContext_FailOpenEmptyIndex(t *testing.T) {
	wc := buildWriteContext([]string{"svc/auth/x.go", "core/api/y.go"}, "", output.WriteIndex{})
	if len(wc.Modules) != 0 {
		t.Errorf("empty index must yield no modules, got %d", len(wc.Modules))
	}
	var buf strings.Builder
	if err := output.EncodeWriteContext(&buf, wc); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"modules": []`) {
		t.Errorf("expected empty modules array, got: %s", buf.String())
	}
}
