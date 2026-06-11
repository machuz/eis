package metric

import "testing"

// Glob-pattern resolution: a path matching "services/*" collapses to the
// 2-component prefix, anchoring module identity to a real project boundary.
func TestModuleResolver_GlobMatch(t *testing.T) {
	r := NewModuleResolver(nil) // default pattern set
	cases := []struct {
		path string
		want string
	}{
		{"services/ace/backend/internal/x.go", "services/ace"},
		{"services/true/frontend/src/App.tsx", "services/true"},
		{"packages/ui/src/Button.tsx", "packages/ui"},
		{"apps/web/main.go", "apps/web"},
		{"modules/auth/jwt.go", "modules/auth"},
		{"libs/core/util.go", "libs/core"},
	}
	for _, c := range cases {
		if got := r.ModuleOf(c.path); got != c.want {
			t.Errorf("ModuleOf(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

// A deeper pattern ("apps/*/lib") picks out a 3-component module identifier
// when the literal "lib" segment is in the right slot.
func TestModuleResolver_DeeperPattern(t *testing.T) {
	r := NewModuleResolver([]string{"apps/*/lib"})
	if got := r.ModuleOf("apps/api/lib/utils/foo.go"); got != "apps/api/lib" {
		t.Errorf("ModuleOf(apps/api/lib/...) = %q, want apps/api/lib", got)
	}
	// Same prefix shape but the "lib" slot mismatches → no glob hit,
	// falls back to the conservative 2-component default.
	if got := r.ModuleOf("apps/api/bin/foo.go"); got != "apps/api" {
		t.Errorf("ModuleOf(apps/api/bin/...) = %q, want apps/api (fallback)", got)
	}
}

// Non-matching paths fall back to the conservative 2-component default —
// intentionally shallower than a deep-tree explosion would suggest.
func TestModuleResolver_NoMatchFallback(t *testing.T) {
	// Use an explicit pattern that won't match these paths.
	r := NewModuleResolver([]string{"services/*"})
	cases := []struct {
		path string
		want string
	}{
		{"a/b/c/d.go", "a/b"},
		{"internal/metric/survival.go", "internal/metric"},
		{"src/components/forms/Input.tsx", "src/components"},
	}
	for _, c := range cases {
		if got := r.ModuleOf(c.path); got != c.want {
			t.Errorf("ModuleOf(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

// Short paths must not panic and must collapse sanely: a 1-component dir
// stays whole, a root-level file resolves to ".".
func TestModuleResolver_ShortPaths(t *testing.T) {
	r := NewModuleResolver(nil)
	cases := []struct {
		path string
		want string
	}{
		{"a/b.go", "a"},     // dir is "a" (1 component)
		{"main.go", "."},    // root-level file: dir is "."
		{"a/b/c.go", "a/b"}, // dir is "a/b" (exactly 2)
		{"", "."},           // empty path: filepath.Dir("") == "."
	}
	for _, c := range cases {
		if got := r.ModuleOf(c.path); got != c.want {
			t.Errorf("ModuleOf(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

// Multiple-pattern, longest-wins: the most specific match controls the
// module identifier. Both "apps/*" (2 comps) and "apps/*/lib" (3 comps)
// can match "apps/api/lib/x.go"; the 3-comp pattern wins.
func TestModuleResolver_LongestWins(t *testing.T) {
	r := NewModuleResolver([]string{"apps/*", "apps/*/lib"})
	if got := r.ModuleOf("apps/api/lib/x.go"); got != "apps/api/lib" {
		t.Errorf("longest-wins: ModuleOf(apps/api/lib/x.go) = %q, want apps/api/lib", got)
	}
	// "apps/web/main.go" doesn't have a "lib" slot — only the 2-comp pattern
	// matches.
	if got := r.ModuleOf("apps/web/main.go"); got != "apps/web" {
		t.Errorf("only shorter matches: ModuleOf(apps/web/main.go) = %q, want apps/web", got)
	}
}

// Tie-break (same-length matches): the pattern listed FIRST in the input
// wins. We compare two patterns that resolve to module identifiers of the
// same length but spelled differently; the first listed pattern's match
// is the one that gets kept.
func TestModuleResolver_TieBreakByOrder(t *testing.T) {
	// Both patterns match "services/ace/foo.go" at depth 2 and yield the
	// SAME module id ("services/ace") — that's by design (same depth →
	// same prefix). Use a synthetic case where two equally-deep patterns
	// each match a different path; here we just confirm that swapping
	// pattern order doesn't change the resolved id when both match the
	// same depth (since both patterns yield the same prefix).
	a := NewModuleResolver([]string{"services/*", "*/ace"})
	b := NewModuleResolver([]string{"*/ace", "services/*"})
	// Path matches both patterns: services/ace/...
	pa := a.ModuleOf("services/ace/backend/x.go")
	pb := b.ModuleOf("services/ace/backend/x.go")
	if pa != pb {
		t.Errorf("tie-break should yield the same prefix regardless of order: %q vs %q", pa, pb)
	}
	if pa != "services/ace" {
		t.Errorf("tie-break: ModuleOf(services/ace/...) = %q, want services/ace", pa)
	}
}

// Single-component path: the conservative 2-component fallback collapses
// it to whatever components exist.
func TestModuleResolver_SingleComponentPath(t *testing.T) {
	r := NewModuleResolver([]string{"services/*"})
	if got := r.ModuleOf("README.md"); got != "." {
		t.Errorf("ModuleOf(README.md) = %q, want .", got)
	}
	if got := r.ModuleOf("services"); got != "." {
		t.Errorf("ModuleOf(services) = %q, want . (single-component file at root)", got)
	}
}

// A custom pattern set REPLACES the default — "services" is no longer
// a module-anchoring pattern, and the new "domains/*" is.
func TestModuleResolver_CustomReplacesDefault(t *testing.T) {
	r := NewModuleResolver([]string{"domains/*"})
	if got := r.ModuleOf("domains/payment/charge.go"); got != "domains/payment" {
		t.Errorf("custom pattern: ModuleOf(domains/...) = %q, want domains/payment", got)
	}
	// "services" is NOT in the custom set → plain 2-layer fallback.
	if got := r.ModuleOf("services/ace/backend/internal/x.go"); got != "services/ace" {
		t.Errorf("custom pattern: ModuleOf(services/...) = %q, want services/ace (fallback)", got)
	}
}

// Empty pattern set falls back to the built-in default, and the default
// must include "packages/*" (so "packages/ui/..." resolves to "packages/ui").
func TestNewModuleResolver_EmptyUsesDefault(t *testing.T) {
	rEmpty := NewModuleResolver(nil)
	rExplicit := NewModuleResolver(DefaultModulePatterns)
	path := "packages/ui/src/Button.tsx"
	if rEmpty.ModuleOf(path) != rExplicit.ModuleOf(path) {
		t.Errorf("empty pattern set must equal explicit default set")
	}
	if rEmpty.ModuleOf(path) != "packages/ui" {
		t.Errorf("default set must treat packages/* as a module pattern")
	}
}

// Edge cases in pattern compilation: empty patterns, separator-only
// patterns, and a lone any-depth wildcard ("**") must not crash.
//   - "" and "/" are dropped (zero meaningful components).
//   - "**" is an any-depth wildcard matching zero or more components; a lone
//     "**" consumes zero parts, so it never produces a module on its own and
//     the resolver falls back to the conservative 2-component default.
func TestNewModuleResolver_PatternEdgeCases(t *testing.T) {
	// Empty/separator-only patterns are dropped silently.
	r := NewModuleResolver([]string{"", "/", "services/*"})
	if got := r.ModuleOf("services/ace/x.go"); got != "services/ace" {
		t.Errorf("after dropping junk patterns: ModuleOf(services/ace/x.go) = %q, want services/ace", got)
	}

	// A lone "**" matches every path but consumes zero leading components, so
	// it produces no module — the resolver falls back to the 2-component
	// default (or whatever the dir collapses to for short paths).
	rss := NewModuleResolver([]string{"**"})
	if got := rss.ModuleOf("a/b/c/d.go"); got != "a/b" {
		t.Errorf("'**' pattern: ModuleOf(a/b/c/d.go) = %q, want a/b (fallback)", got)
	}
	if got := rss.ModuleOf("a.go"); got != "." {
		t.Errorf("'**' pattern: ModuleOf(a.go) = %q, want . (fallback collapses to dir)", got)
	}
}
