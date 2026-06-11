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
// The two trailing anchors `**/domain/*` and `**/usecase/*` capture
// sub-domain modules at ANY depth: a `domain` or `usecase` directory marks a
// business-knowledge boundary (the part of the tree that carries domain
// meaning), so the module identifier extends down to the immediate child of
// that directory regardless of how deep the monorepo nests it (e.g.
// "services/ace/backend/app/domain/star"). Infrastructure and presentation
// directories are deliberately NOT anchored — they carry framework/plumbing,
// not business knowledge, and anchoring them would fragment the topology
// without adding domain signal. See docs/eis/module-recognition.md (in the
// orbit repo) for the rationale.
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
	"**/domain/*",
	"**/usecase/*",
}

// PeripheralModule is the sentinel module id into which fallback-derived
// modules that fail the liveness gate are FOLDED (not dropped — observation
// is preserved, D-07). A fallback module earns module-hood by being touched in
// >= ModuleLivenessMinMonths distinct calendar months; otherwise its
// survival / gravity / etc. aggregate under this id. Pattern-declared modules
// never fold. See docs/eis/module-recognition.md (orbit repo), ADR step 2.
const PeripheralModule = "peripheral"

// parsedPattern is a pre-compiled glob pattern. Each component is a literal
// path-component (Literal), a single-component wildcard (`*`, Wildcard ==
// true, matches exactly ONE path component), or an any-depth wildcard (`**`,
// DoubleWild == true, matches ZERO OR MORE consecutive path components).
type parsedPattern struct {
	components []patternComponent
}

type patternComponent struct {
	Literal string
	// Wildcard is a single-component wildcard (`*`): matches exactly one path
	// component.
	Wildcard bool
	// DoubleWild is an any-depth wildcard (`**`): matches zero or more
	// consecutive path components.
	DoubleWild bool
}

// ModuleResolver maps a file path to a module identifier.
//
// Resolution is pure and deterministic (W-02 / W-03): the result depends
// only on (path, pattern list). The resolver carries no mutable state — it
// is a value, constructed once for a given pattern list and threaded down
// to the metric functions that need it.
//
// Resolution rule:
//   - For each pattern, match the pattern's components against a PREFIX of
//     the path's components. A `*` matches exactly one component; a `**`
//     matches any number of components (including zero); a literal must
//     equal the component. When the pattern matches, it CONSUMES some number
//     of leading path components, and the module identifier is exactly that
//     prefix. For a fixed-length pattern (no `**`) the consumed length is
//     len(pattern.components); with a `**` the length is variable and the
//     module identifier runs UP TO AND INCLUDING the component matched by the
//     LAST (non-`**`) pattern component — the `**` binds to the SHORTEST
//     prefix, so `**/domain/*` anchors at the FIRST `domain` directory.
//   - The winner across all matching patterns is the LONGEST module
//     identifier (most consumed components); ties are broken by pattern order
//     (first wins).
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
	// fold is the set of fallback-derived module ids that failed the liveness
	// gate (touched in fewer than ModuleLivenessMinMonths distinct calendar
	// months). When a resolved, non-excluded module is in this set, ModuleOf
	// returns PeripheralModule instead, folding its mass into the sentinel.
	// nil (the default) means no folding. Populated via WithFold; the fold set
	// itself is computed by ComputeModuleFold. Pattern-declared modules and
	// PeripheralModule are never present here.
	fold map[string]bool
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
// A `**` segment matches any number of path components (including zero);
// a `*` segment matches exactly one. Anything else is a literal path
// component.
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
		switch p {
		case "**":
			comps = append(comps, patternComponent{DoubleWild: true})
		case "*":
			comps = append(comps, patternComponent{Wildcard: true})
		default:
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
// so excluded modules stay out of every module-topology metric.
//
// Precedence: exclude beats fold. A path resolving to an excluded module
// returns "" (it never reaches the fold step). Otherwise, when the resolved
// module is in the fold set (a fallback-derived module that failed the
// liveness gate), PeripheralModule is returned, aggregating its mass into the
// sentinel. PeripheralModule itself is never folded or excluded. See
// ModuleResolver for the resolution rule.
func (r ModuleResolver) ModuleOf(path string) string {
	module := r.resolve(path)
	if r.isExcluded(module) {
		return ""
	}
	if r.fold[module] {
		return PeripheralModule
	}
	return module
}

// ResolveWithOrigin returns the exclude-applied module id for a file path and
// whether it was pattern-DECLARED (a glob pattern matched) as opposed to
// fallback-derived (the conservative 2-component default). An excluded module
// returns ("", false). Folding is NOT applied here — the liveness gate
// (ComputeModuleFold) needs the un-folded origin to decide what folds. Use
// ModuleOf for the bucketing module id (which applies fold).
func (r ModuleResolver) ResolveWithOrigin(path string) (module string, declared bool) {
	module, declared = r.resolveWithOrigin(path)
	if r.isExcluded(module) {
		return "", false
	}
	return module, declared
}

// resolve maps a file path to a module identifier without applying excludes
// or fold.
func (r ModuleResolver) resolve(path string) string {
	module, _ := r.resolveWithOrigin(path)
	return module
}

// resolveWithOrigin maps a file path to a module identifier without applying
// excludes or fold, also reporting whether a glob pattern matched (declared)
// or the conservative 2-component fallback was used (declared == false).
func (r ModuleResolver) resolveWithOrigin(path string) (module string, declared bool) {
	dir := filepath.Dir(path)
	parts := strings.Split(filepath.ToSlash(dir), "/")

	// Try each pattern; keep the longest module identifier (most specific
	// match wins). Ties go to the pattern listed first — we only replace
	// `best` when strictly longer.
	bestLen := 0
	for _, pat := range r.patterns {
		consumed, ok := matchPrefix(pat.components, parts)
		// A pattern that consumes zero parts (e.g. a lone `**`) does not
		// produce a module — require at least one consumed component.
		if !ok || consumed < 1 {
			continue
		}
		if consumed > bestLen {
			bestLen = consumed
		}
	}
	if bestLen > 0 {
		return strings.Join(parts[:bestLen], "/"), true
	}

	// No pattern matched: conservative 2-component default (fallback-derived).
	if len(parts) > 2 {
		parts = parts[:2]
	}
	return strings.Join(parts, "/"), false
}

// WithFold returns a copy of the resolver whose ModuleOf folds the given set
// of (fallback-derived, gate-failing) module ids into PeripheralModule. Value
// semantics: the receiver is unchanged. Passing nil or an empty map disables
// folding. The fold set is produced by ComputeModuleFold.
func (r ModuleResolver) WithFold(fold map[string]bool) ModuleResolver {
	r.fold = fold
	return r
}

// matchPrefix matches a compiled pattern's components against a PREFIX of
// parts and reports (consumed, true) where consumed is the number of leading
// parts the pattern spans, or (0, false) when no prefix match exists. The
// module is always a path prefix, so the pattern need not span the whole
// path — only its leading components.
//
// `*` matches exactly one part; a literal must equal the part; `**` matches
// zero or more consecutive parts. `**` is resolved by backtracking and binds
// to the SHORTEST run that lets the rest of the pattern match (standard glob
// two-pointer with `**` lookahead), so `**/domain/*` anchors at the FIRST
// `domain` directory. `**` may appear anywhere in the pattern, not just at
// the front.
func matchPrefix(comps []patternComponent, parts []string) (int, bool) {
	// Pattern fully consumed: it matched this prefix (the remaining parts are
	// the deeper path below the module). The "must consume >= 1" rule is
	// enforced by the caller (resolve), not here — an empty pattern in the
	// recursion legitimately consumes zero further parts.
	if len(comps) == 0 {
		return 0, true
	}
	c := comps[0]
	if c.DoubleWild {
		// `**` binds to the shortest prefix: try consuming 0, then 1, ...
		// of parts before matching the rest of the pattern. The trailing
		// `**` (no further components) consumes zero — handled by caller's
		// consumed >= 1 gate.
		for skip := 0; skip <= len(parts); skip++ {
			if n, ok := matchPrefix(comps[1:], parts[skip:]); ok {
				return skip + n, true
			}
		}
		return 0, false
	}
	// Single-component (literal or `*`) needs at least one part.
	if len(parts) == 0 {
		return 0, false
	}
	if !c.Wildcard && parts[0] != c.Literal {
		return 0, false
	}
	n, ok := matchPrefix(comps[1:], parts[1:])
	if !ok {
		return 0, false
	}
	return 1 + n, true
}
