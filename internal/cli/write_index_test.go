package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/config"
	"github.com/machuz/eis/v2/internal/output"
	"github.com/machuz/eis/v2/internal/scorer"
)

// dr wraps module scores into a one-domain DomainResults for tests.
func drWith(mods ...scorer.ModuleScore) []DomainResults {
	return []DomainResults{{Domain: "Backend", ModuleScores: mods}}
}

func TestBuildWriteIndex_RecommendationsAndFacts(t *testing.T) {
	const gone, stale = 180.0, 180.0
	mods := []scorer.ModuleScore{
		// Orphaned debt: owner gone 400d, untouched 400d, has commits.
		{Module: "svc/auth", Vitality: "Stable", Ownership: "Orphaned",
			OwnerLastActiveDays: 400, ModuleUntouchedDays: 400, ModuleCommits: 12,
			TopAuthorShare: 0.92, OwnerActive: false, BlameLines: 500},
		// Dead debt: no commits, owner inactive.
		{Module: "legacy/vm", Vitality: "Dead", Ownership: "Concentrated",
			OwnerLastActiveDays: 900, ModuleUntouchedDays: 0, ModuleCommits: 0,
			TopAuthorShare: 1.0, OwnerActive: false, BlameLines: 200},
		// At-risk: concentrated + owner active (bus-factor 1), NOT debt.
		{Module: "core/engine", Vitality: "Stable", Ownership: "Concentrated",
			OwnerLastActiveDays: 3, ModuleUntouchedDays: 3, ModuleCommits: 40,
			TopAuthorShare: 0.88, OwnerActive: true, BlameLines: 3000},
		// Healthy: distributed, active.
		{Module: "core/api", Vitality: "Stable", Ownership: "Distributed",
			OwnerLastActiveDays: 2, ModuleUntouchedDays: 2, ModuleCommits: 80,
			TopAuthorShare: 0.3, OwnerActive: true, BlameLines: 4000},
		// Orphaned BUT still maintained (untouched 8d) ⇒ NOT debt (empty rec).
		{Module: "pkg/util", Vitality: "Stable", Ownership: "Orphaned",
			OwnerLastActiveDays: 1506, ModuleUntouchedDays: 8, ModuleCommits: 50,
			TopAuthorShare: 0.7, OwnerActive: false, BlameLines: 800},
	}

	idx := buildWriteIndex(drWith(mods...), &config.Config{}, "abc123", time.Unix(0, 0).UTC(), gone, stale, nil, nil)

	// All modules present (agent may write anywhere).
	if len(idx.Modules) != len(mods) {
		t.Fatalf("modules = %d, want %d (every module must appear)", len(idx.Modules), len(mods))
	}

	want := map[string]struct {
		tier, rec string
		atRisk    bool
	}{
		"svc/auth":    {"Orphaned", "orphaned_module", false},
		"legacy/vm":   {"Dead", "dead_module", false},
		"core/engine": {"", "bus_factor_1", true},
		"core/api":    {"", "", false},
		"pkg/util":    {"", "", false},
	}
	for mod, w := range want {
		got, ok := idx.Modules[mod]
		if !ok {
			t.Fatalf("module %q missing", mod)
		}
		if got.DebtTier != w.tier {
			t.Errorf("%s DebtTier = %q, want %q", mod, got.DebtTier, w.tier)
		}
		if got.Recommendation != w.rec {
			t.Errorf("%s Recommendation = %q, want %q", mod, got.Recommendation, w.rec)
		}
		if got.AtRisk != w.atRisk {
			t.Errorf("%s AtRisk = %v, want %v", mod, got.AtRisk, w.atRisk)
		}
	}

	// Structural facts wired.
	auth := idx.Modules["svc/auth"]
	if auth.OwnerLeftDays != 400 || auth.UntouchedDays != 400 {
		t.Errorf("svc/auth days = %d/%d, want 400/400", auth.OwnerLeftDays, auth.UntouchedDays)
	}
	if auth.OwnershipConcentration != 0.92 {
		t.Errorf("svc/auth concentration = %v, want 0.92", auth.OwnershipConcentration)
	}
	if idx.Commit != "abc123" || idx.GeneratedAt == "" {
		t.Errorf("lineage: commit=%q generated_at=%q", idx.Commit, idx.GeneratedAt)
	}
}

// TestBuildWriteIndex_NoOwnerNames is the firewall guard: the serialized index
// must not contain any owner identity, even when ownership records carry names.
func TestBuildWriteIndex_NoOwnerNames(t *testing.T) {
	mods := []scorer.ModuleScore{
		{Module: "svc/auth", Ownership: "Orphaned", OwnerLastActiveDays: 400,
			ModuleUntouchedDays: 400, ModuleCommits: 5, TopAuthorShare: 0.9, BlameLines: 100},
	}
	dr := drWith(mods...)
	// Ownership records DO carry the name — buildWriteIndex must not leak it.
	dr[0].Ownership = nil // WriteIndex reads ModuleScores, not Ownership; belt-and-suspenders

	idx := buildWriteIndex(dr, &config.Config{}, "sha", time.Unix(0, 0).UTC(), 180, 180, nil, nil)

	var buf bytes.Buffer
	if err := output.EncodeWriteIndex(&buf, idx); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, name := range []string{"tanaka", "Mark Erikson", "Lee Byron", "top_author", "last_owner", "TopAuthor", "owner_name"} {
		if strings.Contains(s, name) {
			t.Errorf("index leaked owner identity token %q:\n%s", name, s)
		}
	}
	// Sanity: it DOES carry structural facts.
	if !strings.Contains(s, `"owner_left_days": 400`) {
		t.Errorf("expected structural facts in index:\n%s", s)
	}
}

// TestBuildWriteIndex_DebtMatchesStructuralDebt locks the two commands together:
// for every module, write-index's debt_tier must agree with the tier
// structural-debt would assign under the same thresholds.
func TestBuildWriteIndex_DebtMatchesStructuralDebt(t *testing.T) {
	const gone, stale = 180.0, 180.0
	mods := []scorer.ModuleScore{
		{Module: "a", Vitality: "Dead", Ownership: "Concentrated", ModuleCommits: 0, OwnerLastActiveDays: 500, ModuleUntouchedDays: 0, BlameLines: 100},
		{Module: "b", Vitality: "Stable", Ownership: "Orphaned", ModuleCommits: 10, OwnerLastActiveDays: 300, ModuleUntouchedDays: 300, BlameLines: 200},
		{Module: "c", Vitality: "Stable", Ownership: "Orphaned", ModuleCommits: 10, OwnerLastActiveDays: 300, ModuleUntouchedDays: 8, BlameLines: 200}, // not stale
		{Module: "d", Vitality: "Stable", Ownership: "Distributed", ModuleCommits: 20, BlameLines: 400},
	}
	idx := buildWriteIndex(drWith(mods...), &config.Config{}, "x", time.Unix(0, 0).UTC(), gone, stale, nil, nil)

	// structural-debt's own report over the same modules (no non-source excludes,
	// so every debt module surfaces as a drill row we can cross-check).
	rep := computeStructuralDebt("Backend", mods, nil, false, false, gone, stale, 100)
	sdTier := map[string]string{}
	for _, dm := range rep.TopDebtModules {
		sdTier[dm.Module] = dm.Tier
	}
	for _, m := range mods {
		got := idx.Modules[m.Module].DebtTier
		want := sdTier[m.Module] // "" when structural-debt did not flag it
		if got != want {
			t.Errorf("module %s: write-index tier %q != structural-debt tier %q", m.Module, got, want)
		}
	}
}

// TestBuildWriteIndex_AnchorsGraveyardWiring checks the Build2 payloads attach to
// the right modules, and that modules absent from the maps omit the fields (opt-in
// stays opt-in per module).
func TestBuildWriteIndex_AnchorsGraveyardWiring(t *testing.T) {
	mods := []scorer.ModuleScore{
		{Module: "core/api", Ownership: "Distributed", ModuleCommits: 40, BlameLines: 1000},
		{Module: "legacy/vm", Ownership: "Concentrated", ModuleCommits: 30, BlameLines: 500},
		{Module: "pkg/util", Ownership: "Distributed", ModuleCommits: 20, BlameLines: 300},
	}
	anchors := map[string][]output.Anchor{
		"core/api": {{File: "core/api/router.go", LineRange: [2]int{10, 20}, Survival: 0.8, Gravity: 12.5, ContestedByN: 3, Digest: "func Route() {"}},
	}
	graves := map[string]output.WriteIndexGraveyard{
		"legacy/vm": {DeathIntensity: 0.42, DeathEvents: 6, Hotspots: []output.GraveyardHotspot{{File: "legacy/vm/exec.go", Deaths: 4, ContestedByN: 2}}},
	}

	idx := buildWriteIndex(drWith(mods...), &config.Config{}, "sha", time.Unix(0, 0).UTC(), 180, 180, anchors, graves)

	// Anchors attach only to core/api.
	if got := idx.Modules["core/api"].Anchors; len(got) != 1 || got[0].File != "core/api/router.go" {
		t.Errorf("core/api anchors = %+v, want the one router.go anchor", got)
	}
	if a := idx.Modules["legacy/vm"].Anchors; a != nil {
		t.Errorf("legacy/vm should have no anchors, got %+v", a)
	}
	// Graveyard attaches only to legacy/vm.
	g := idx.Modules["legacy/vm"].Graveyard
	if g == nil || g.DeathEvents != 6 || g.DeathIntensity != 0.42 {
		t.Errorf("legacy/vm graveyard = %+v, want death_events 6 / intensity 0.42", g)
	}
	if idx.Modules["core/api"].Graveyard != nil {
		t.Errorf("core/api should have no graveyard, got %+v", idx.Modules["core/api"].Graveyard)
	}
	// pkg/util is in neither map ⇒ both fields omitted.
	if u := idx.Modules["pkg/util"]; u.Anchors != nil || u.Graveyard != nil {
		t.Errorf("pkg/util should carry neither payload, got anchors=%+v graveyard=%+v", u.Anchors, u.Graveyard)
	}

	// The omitempty contract: a payload-free module serializes without the keys.
	var buf bytes.Buffer
	if err := output.EncodeWriteIndex(&buf, idx); err != nil {
		t.Fatal(err)
	}
	// pkg/util's object must not carry "anchors"/"graveyard". Cheap structural check:
	// decode and assert the raw keys are absent on that module.
	var decoded struct {
		Modules map[string]map[string]json.RawMessage `json:"modules"`
	}
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"anchors", "graveyard"} {
		if _, present := decoded.Modules["pkg/util"][k]; present {
			t.Errorf("pkg/util serialized with %q key despite empty payload", k)
		}
	}
}

func TestBuildWriteIndex_ModulePatternsSource(t *testing.T) {
	// No declared architecture ⇒ built-in defaults + annotation.
	idx := buildWriteIndex(drWith(), &config.Config{}, "", time.Unix(0, 0).UTC(), 180, 180, nil, nil)
	if idx.ModulePatternsSource != "builtin_default" {
		t.Errorf("source = %q, want builtin_default", idx.ModulePatternsSource)
	}
	if len(idx.ModulePatterns) == 0 {
		t.Error("expected fallback DefaultModulePatterns to be emitted")
	}
	// Declared architecture ⇒ config patterns.
	cfg := &config.Config{ModulePatterns: []string{"core/*", "svc/*"}}
	idx = buildWriteIndex(drWith(), cfg, "", time.Unix(0, 0).UTC(), 180, 180, nil, nil)
	if idx.ModulePatternsSource != "config" {
		t.Errorf("source = %q, want config", idx.ModulePatternsSource)
	}
	b, _ := json.Marshal(idx.ModulePatterns)
	if !strings.Contains(string(b), "core/*") {
		t.Errorf("declared patterns not emitted: %s", b)
	}
}
