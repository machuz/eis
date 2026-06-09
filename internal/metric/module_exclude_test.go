package metric

import (
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

// ExcludeModules drops the resolved module to "" — the user-reported case:
// date-stamped migration directories at the repo root (matched by "20*")
// that the 2-component fallback otherwise turns into modules.
func TestModuleResolver_ExcludeModules(t *testing.T) {
	r := NewModuleResolverWithExcludes(nil, []string{"20*", "migrations"})
	cases := []struct {
		path string
		want string
	}{
		// Excluded: date dirs at root (single- and multi-component fallback).
		{"20240403/up.sql", ""},
		{"20250417/update_store_plans/0001.sql", ""},
		{"migrations/0001_init/up.sql", ""},
		// Not excluded: real modules pass through unchanged.
		{"services/ace/backend/main.go", "services/ace"},
		{"app/domain/x.go", "app/domain"},
		// "20*" is a prefix glob on the first component — it matches ANY module
		// whose first dir starts with "20" (e.g. a hypothetical "2024-roadmap").
		// Users who want strictly 8-digit date dirs should use the char-class
		// pattern "[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]" instead.
		{"2024-roadmap/x.go", ""},
	}
	for _, c := range cases {
		if got := r.ModuleOf(c.path); got != c.want {
			t.Errorf("ModuleOf(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

// Without excludes the resolver is unchanged (the date dir still resolves to a
// module) — exclusion is strictly opt-in.
func TestModuleResolver_NoExcludes_Unchanged(t *testing.T) {
	r := NewModuleResolver(nil)
	if got := r.ModuleOf("20240403/up.sql"); got != "20240403" {
		t.Errorf("without excludes, ModuleOf = %q, want 20240403", got)
	}
	r2 := NewModuleResolverWithExcludes(nil, nil)
	if got := r2.ModuleOf("20240403/up.sql"); got != "20240403" {
		t.Errorf("nil excludes, ModuleOf = %q, want 20240403", got)
	}
}

// Excluded modules must not appear in the per-module / per-author survival maps
// that feed Breadth (ComputeBreadth's Hill number) — the decisive reason the
// fix lives in EIS rather than a SaaS read-time filter.
func TestExcludeModules_KeptOutOfSurvivalAndBreadth(t *testing.T) {
	now := time.Now()
	blame := []git.BlameLine{
		{Author: "alice", CommitterTime: now, Filename: "20240403/up.sql"},
		{Author: "alice", CommitterTime: now, Filename: "20250417/seed/x.sql"},
		{Author: "alice", CommitterTime: now, Filename: "src/app/handler.go"},
		{Author: "alice", CommitterTime: now, Filename: "src/core/util.go"},
	}
	r := NewModuleResolverWithExcludes(nil, []string{"20*"})

	mod := CalcModuleSurvival(blame, 180, now, r)
	for m := range mod {
		if m == "20240403" || m == "20250417/seed" {
			t.Errorf("excluded module %q leaked into CalcModuleSurvival", m)
		}
	}

	msba := CalcModuleSurvivalByAuthor(blame, 180, now, r)
	if _, ok := msba["20240403"]; ok {
		t.Errorf("excluded module leaked into CalcModuleSurvivalByAuthor")
	}

	// Breadth = effective module count for alice. Only src/app and src/core
	// survive the filter → 2 distinct equal-mass modules → Hill number 2,
	// not 4 (which the date dirs would have inflated it to).
	breadth := ComputeBreadth(msba)
	if got := breadth["alice"]; got < 1.99 || got > 2.01 {
		t.Errorf("Breadth(alice) = %v, want ~2 (date dirs must not inflate it)", got)
	}
}
