package metric

import (
	"path/filepath"
	"strings"
)

// DefaultModulePatterns is the built-in set of glob patterns used by module
// resolution when no caller-supplied pattern list is provided. It mirrors
// the historical "convention dirs" set (services, packages, apps, modules,
// libs), expressed as globs so users can extend / refine it.
//
// The canonical default lives in package config (config.DefaultModulePatterns).
// This copy exists so the metric layer remains import-cycle-free; the two
// MUST stay in lockstep. A test in internal/metric/module_test.go asserts
// that intent at the value level.
var DefaultModulePatterns = []string{
	"services/*",
	"packages/*",
	"apps/*",
	"modules/*",
	"libs/*",
}

// parsedPattern is a pre-compiled glob pattern. Each component is either a
// literal path-component (Wildcard == false) or a single-component wildcard
// (`*`, Wildcard == true). Multi-component wildcards (`**`) are not
// supported — a `*` matches exactly ONE path component.
type parsedPattern struct {
	components []patternComponent
}

type patternComponent struct {
	Literal  string
	Wildcard bool
}

// ModuleResolver maps a file path to a module identifier.
//
// Resolution is pure and deterministic (W-02 / W-03): the result depends
// only on (path, pattern list). The resolver carries no mutable state — it
// is a value, constructed once for a given pattern list and threaded down
// to the metric functions that need it.
//
// Resolution rule:
//   - For each pattern, walk the path's components against the pattern's
//     components. If every pattern component matches (literal-equal or
//     wildcard), the pattern HAS MATCHED and the module identifier is the
//     first N components of the path (N = len(pattern.components)).
//   - The winner across all matching patterns is the LONGEST module
//     identifier; ties are broken by pattern order (first wins).
//   - If NO pattern matches, fall back to the conservative 2-component
//     default (e.g. "a/b/c/d.go" -> "a/b"). A path with fewer than 2
//     components collapses to whatever components exist (the whole dir,
//     or "." for a root-level file).
type ModuleResolver struct {
	patterns []parsedPattern
	// excludes is the compiled ExcludeModules list — each entry is a pattern
	// split into path components. A resolved module is dropped (ModuleOf
	// returns "") when any exclude matches the FRONT of its components via
	// filepath.Match. Empty (the default) means no exclusion.
	excludes [][]string
}

// NewModuleResolver builds a resolver from a glob-pattern list. When the
// list is empty (nil or len 0), DefaultModulePatterns is used. Per-repo
// overrides are not handled here — the caller picks the right pattern list
// (see config.PatternsForRepo) and passes it in. This keeps the resolver
// value scoped to (repo, patterns) instead of baking a full override map
// into a single resolver instance.
//
// Empty pattern strings and patterns consisting entirely of separators
// are dropped silently — they can't match anything meaningful and would
// otherwise produce ambiguous "match at depth 0" hits.
func NewModuleResolver(patterns []string) ModuleResolver {
	src := patterns
	if len(src) == 0 {
		src = DefaultModulePatterns
	}
	out := make([]parsedPattern, 0, len(src))
	for _, p := range src {
		pp, ok := compilePattern(p)
		if !ok {
			continue
		}
		out = append(out, pp)
	}
	return ModuleResolver{patterns: out}
}

// NewModuleResolverWithExcludes is NewModuleResolver plus an ExcludeModules
// list (see config.ExcludeModules). Excluded modules resolve to "" so every
// module-topology aggregation drops them. Excludes are matched component-wise
// with filepath.Match against the front of a resolved module id.
func NewModuleResolverWithExcludes(patterns, excludes []string) ModuleResolver {
	r := NewModuleResolver(patterns)
	for _, e := range excludes {
		clean := strings.Trim(filepath.ToSlash(e), "/")
		if clean == "" {
			continue
		}
		var comps []string
		for _, p := range strings.Split(clean, "/") {
			if p != "" {
				comps = append(comps, p)
			}
		}
		if len(comps) > 0 {
			r.excludes = append(r.excludes, comps)
		}
	}
	return r
}

// isExcluded reports whether a resolved module id matches any ExcludeModules
// pattern. A pattern matches when, for its k components, filepath.Match
// succeeds against the module's first k components (a prefix match — so a
// 1-component pattern like "20*" drops both "20240403" and "20240403/sub").
func (r ModuleResolver) isExcluded(module string) bool {
	if len(r.excludes) == 0 || module == "" {
		return false
	}
	parts := strings.Split(module, "/")
	for _, pat := range r.excludes {
		if len(pat) > len(parts) {
			continue
		}
		matched := true
		for i, comp := range pat {
			if ok, _ := filepath.Match(comp, parts[i]); !ok {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

// compilePattern splits a glob pattern into components. Returns (_, false)
// when the pattern has zero meaningful components (e.g. "", "/", "///").
//
// Consecutive `*` ("**") is treated as TWO single-component wildcards,
// matching exactly two path components — not "any depth". The contract is
// "`*` = a single path component"; "**" is not a separate operator.
func compilePattern(s string) (parsedPattern, bool) {
	clean := strings.Trim(filepath.ToSlash(s), "/")
	if clean == "" {
		return parsedPattern{}, false
	}
	parts := strings.Split(clean, "/")
	comps := make([]patternComponent, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			// Skip empty segments from accidental "//" in a pattern.
			continue
		}
		if p == "*" {
			comps = append(comps, patternComponent{Wildcard: true})
		} else {
			comps = append(comps, patternComponent{Literal: p})
		}
	}
	if len(comps) == 0 {
		return parsedPattern{}, false
	}
	return parsedPattern{components: comps}, true
}

// ModuleOf returns the module identifier for a file path under this
// resolver's pattern set, or "" when the resolved module is excluded
// (ExcludeModules). Callers that aggregate by module MUST skip the "" result
// so excluded modules stay out of every module-topology metric. See
// ModuleResolver for the resolution rule.
func (r ModuleResolver) ModuleOf(path string) string {
	module := r.resolve(path)
	if r.isExcluded(module) {
		return ""
	}
	return module
}

// resolve maps a file path to a module identifier without applying excludes.
func (r ModuleResolver) resolve(path string) string {
	dir := filepath.Dir(path)
	parts := strings.Split(filepath.ToSlash(dir), "/")

	// Try each pattern; keep the longest module identifier (most specific
	// match wins). Ties go to the pattern listed first — we only replace
	// `best` when strictly longer.
	bestLen := 0
	var best string
	for _, pat := range r.patterns {
		n := len(pat.components)
		if n == 0 || n > len(parts) {
			continue
		}
		matched := true
		for i, c := range pat.components {
			if c.Wildcard {
				continue
			}
			if parts[i] != c.Literal {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		if n > bestLen {
			bestLen = n
			best = strings.Join(parts[:n], "/")
		}
	}
	if bestLen > 0 {
		return best
	}

	// No pattern matched: conservative 2-component default.
	if len(parts) > 2 {
		parts = parts[:2]
	}
	return strings.Join(parts, "/")
}
