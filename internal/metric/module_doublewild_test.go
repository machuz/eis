package metric

import (
	"reflect"
	"testing"

	"github.com/machuz/eis/v2/internal/config"
)

// The `**` any-depth glob anchors sub-domain modules (domain/usecase) at any
// nesting depth. These are the worked examples from the module-recognition
// ADR (step 1): a `**` binds to the SHORTEST prefix, so we anchor at the
// FIRST occurrence of the literal that follows it, and the module identifier
// runs down to (and including) the immediate child captured by the trailing
// `*`. A `domain`/`usecase` dir with no child after it (a loose file directly
// under it) does NOT match the anchor and falls back to the conservative
// 2-component default.
func TestResolve_DoubleWildAnchors(t *testing.T) {
	cases := []struct {
		name     string
		patterns []string // nil = default set (which now includes the anchors)
		path     string
		want     string
	}{
		{
			name: "domain anchor, shallow",
			path: "app/domain/star/model.go",
			want: "app/domain/star",
		},
		{
			name: "domain anchor, deep monorepo nesting",
			path: "services/ace/backend/app/domain/star/x.go",
			want: "services/ace/backend/app/domain/star",
		},
		{
			name: "loose file directly under domain -> no anchor, 2-comp fallback",
			path: "app/domain/errors.go",
			want: "app/domain",
		},
		{
			name: "usecase anchor",
			path: "app/usecase/billing/svc.go",
			want: "app/usecase/billing",
		},
		{
			// services/* consumes 2 (services/ace); **/domain/* consumes 6 —
			// the deeper, most-specific domain module wins (longest consumed).
			name: "longest-wins between services/* and **/domain/*",
			path: "services/ace/backend/app/domain/star/x.go",
			want: "services/ace/backend/app/domain/star",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := NewModuleResolver(c.patterns)
			if got := r.ModuleOf(c.path); got != c.want {
				t.Errorf("ModuleOf(%q) = %q, want %q", c.path, got, c.want)
			}
		})
	}
}

// Explicit longest-wins check with the exact two-pattern set from the ADR
// worked example: [services/*, **/domain/*]. services/* would consume 2,
// **/domain/* consumes 6 — the deeper module wins.
func TestResolve_DoubleWild_LongestWins(t *testing.T) {
	r := NewModuleResolver([]string{"services/*", "**/domain/*"})
	const path = "services/ace/backend/app/domain/star/x.go"
	if got := r.ModuleOf(path); got != "services/ace/backend/app/domain/star" {
		t.Errorf("ModuleOf(%q) = %q, want services/ace/backend/app/domain/star", path, got)
	}
}

// Backward compatibility: a fixed-length pattern with no `**` behaves exactly
// as before — `*` matches exactly one component, and the module is the
// matched-length prefix.
func TestResolve_DoubleWild_BackwardCompat(t *testing.T) {
	r := NewModuleResolver([]string{"services/*"})
	if got := r.ModuleOf("services/api/foo.go"); got != "services/api" {
		t.Errorf("ModuleOf(services/api/foo.go) = %q, want services/api", got)
	}
}

// Excludes still work with the new resolver: a date-stamped fallback module is
// dropped by "20*", and deep `**`-anchored modules are unaffected (isExcluded
// is prefix-based, and "20*" never matches an "app/.../domain/..." prefix).
func TestResolve_DoubleWild_ExcludesPreserved(t *testing.T) {
	r := NewModuleResolverWithExcludes(nil, []string{"20*"})
	if got := r.ModuleOf("20241030/up.sql"); got != "" {
		t.Errorf("ModuleOf(20241030/up.sql) = %q, want \"\" (excluded)", got)
	}
	if got := r.ModuleOf("app/domain/star/model.go"); got != "app/domain/star" {
		t.Errorf("deep domain module must survive excludes: got %q, want app/domain/star", got)
	}
}

// Lockstep: the metric-layer DefaultModulePatterns copy MUST equal the
// canonical config.DefaultModulePatterns value-for-value. The two live in
// separate packages to keep the metric layer import-cycle-free; this test is
// the contract that keeps them identical (referenced by both doc comments).
func TestDefaultModulePatterns_LockstepWithConfig(t *testing.T) {
	if !reflect.DeepEqual(DefaultModulePatterns, config.DefaultModulePatterns) {
		t.Errorf("metric.DefaultModulePatterns %v != config.DefaultModulePatterns %v",
			DefaultModulePatterns, config.DefaultModulePatterns)
	}
}
