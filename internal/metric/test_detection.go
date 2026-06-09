package metric

import (
	"path/filepath"
	"strings"
)

// IsTestFile reports whether the path matches common test-file naming conventions
// across Go, TS/JS, Python, Ruby, Java/Kotlin/Scala, Rust. It only inspects the
// file path — no file content is read — so the classification is cheap and
// deterministic from git's perspective.
func IsTestFile(path string) bool {
	p := filepath.ToSlash(path)
	base := filepath.Base(p)

	// Go
	if strings.HasSuffix(base, "_test.go") {
		return true
	}
	// TS/JS: *.test.{ts,tsx,js,jsx,mjs,cjs} and *.spec.{same}
	for _, marker := range []string{".test.", ".spec."} {
		if strings.Contains(base, marker) {
			for _, ext := range []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"} {
				if strings.HasSuffix(base, ext) {
					return true
				}
			}
		}
	}
	// Python
	if strings.HasSuffix(base, ".py") && (strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py")) {
		return true
	}
	// Ruby
	if strings.HasSuffix(base, "_spec.rb") || strings.HasSuffix(base, "_test.rb") {
		return true
	}
	// Java / Kotlin / Scala: typical suffixes and prefixes
	for _, ext := range []string{".java", ".kt", ".scala"} {
		if strings.HasSuffix(base, ext) {
			stem := strings.TrimSuffix(base, ext)
			if strings.HasPrefix(stem, "Test") || strings.HasSuffix(stem, "Test") ||
				strings.HasSuffix(stem, "Tests") || strings.HasSuffix(stem, "Spec") ||
				strings.HasSuffix(stem, "IT") { // Integration Test by Maven/Failsafe convention
				return true
			}
		}
	}
	// Rust, Elixir, etc.: tests/ and test/ directories
	if strings.HasPrefix(p, "tests/") || strings.Contains(p, "/tests/") {
		return true
	}
	if strings.HasPrefix(p, "test/") || strings.Contains(p, "/test/") {
		return true
	}
	// JS/TS: __tests__ folder
	if strings.Contains(p, "/__tests__/") || strings.HasPrefix(p, "__tests__/") {
		return true
	}
	// Ruby / general: spec/ at any depth
	if strings.HasPrefix(p, "spec/") || strings.Contains(p, "/spec/") {
		return true
	}
	// Java / Gradle / Maven layouts
	if strings.Contains(p, "/src/test/") || strings.HasPrefix(p, "src/test/") {
		return true
	}

	return false
}

// TestedSet answers "is this file guarded by a test?" using nothing but the
// file manifest — no content analysis, no coverage tooling.
//
// Rules (in order):
//  1. Test files themselves count as tested (they protect other files; the
//     author deserves credit rather than being treated as untested noise).
//  2. Sibling pair: `foo.go` in a dir with `foo_test.go` → tested. Uses
//     language-aware sibling naming patterns.
//  3. Module fallback: if the file's directory contains ANY test file, every
//     non-test file in that directory counts as tested.
//
// Files failing all three checks are considered untested. The TestFileRatio
// (test files / total files) is exposed for downstream observability.
type TestedSet struct {
	tested         map[string]bool
	TestFileRatio  float64
	TotalFiles     int
	TotalTestFiles int

	// Per-module file tallies — used by ScoreModules to compute Vitality=Fragile
	// (a module where survival exists only because nothing touches it and
	// nothing guards it).
	moduleTotalFiles map[string]int
	moduleTestFiles  map[string]int
}

// ModuleTestRatio returns the test-file ratio for the given module
// (test files / total files within that module). Returns (0, false) when
// the module has no tracked files.
func (ts *TestedSet) ModuleTestRatio(module string) (float64, bool) {
	if ts == nil {
		return 0, false
	}
	total := ts.moduleTotalFiles[module]
	if total == 0 {
		return 0, false
	}
	return float64(ts.moduleTestFiles[module]) / float64(total), true
}

// ModuleTestFileCounts returns (total, test) file counts for a module.
// Enables downstream callers to weight averages by module size when
// aggregating across repos.
func (ts *TestedSet) ModuleTestFileCounts(module string) (total, test int) {
	if ts == nil {
		return 0, 0
	}
	return ts.moduleTotalFiles[module], ts.moduleTestFiles[module]
}

// ForEachModule invokes fn for every module tracked by this TestedSet.
// Used by the analyzer to merge per-repo counts into a domain-level tally.
func (ts *TestedSet) ForEachModule(fn func(module string, total, test int)) {
	if ts == nil || fn == nil {
		return
	}
	for mod, total := range ts.moduleTotalFiles {
		fn(mod, total, ts.moduleTestFiles[mod])
	}
}

// BuildTestedSet inspects the full file manifest once and returns a lookup
// that callers can query per blame line.
func BuildTestedSet(allFiles []string, mr ModuleResolver) *TestedSet {
	ts := &TestedSet{
		tested:           make(map[string]bool, len(allFiles)),
		moduleTotalFiles: make(map[string]int),
		moduleTestFiles:  make(map[string]int),
		TotalFiles:       len(allFiles),
	}
	if len(allFiles) == 0 {
		return ts
	}

	// Pass 1: identify test files and the set of dirs that contain at least one.
	// Also tally per-module totals so downstream callers can compute
	// per-module test ratio without re-walking the manifest.
	testFiles := make(map[string]struct{}, len(allFiles)/4)
	dirHasTest := make(map[string]bool)
	for _, f := range allFiles {
		// mod == "" → excluded module (ExcludeModules): keep it out of the
		// per-module test ratios, but still let the file feed the global
		// tested-set (testFiles / dirHasTest) so survival weighting is unchanged.
		mod := mr.ModuleOf(f)
		if mod != "" {
			ts.moduleTotalFiles[mod]++
		}
		if IsTestFile(f) {
			testFiles[f] = struct{}{}
			if mod != "" {
				ts.moduleTestFiles[mod]++
			}
			dir := filepath.Dir(filepath.ToSlash(f))
			dirHasTest[dir] = true
		}
	}
	ts.TotalTestFiles = len(testFiles)
	ts.TestFileRatio = float64(ts.TotalTestFiles) / float64(ts.TotalFiles)

	// Pass 2: classify every file.
	for _, f := range allFiles {
		if _, isTest := testFiles[f]; isTest {
			// Rule 1: test files themselves count as tested.
			ts.tested[f] = true
			continue
		}
		// Rule 2: sibling pair (strongest signal).
		if hasSiblingTestFile(f, testFiles) {
			ts.tested[f] = true
			continue
		}
		// Rule 3: module fallback — any test file in the same directory.
		dir := filepath.Dir(filepath.ToSlash(f))
		if dirHasTest[dir] {
			ts.tested[f] = true
		}
	}
	return ts
}

// IsTested reports whether the given file is covered by a test, per the set's
// precomputed classification. Nil receiver returns false.
func (ts *TestedSet) IsTested(path string) bool {
	if ts == nil {
		return false
	}
	return ts.tested[path]
}

// hasSiblingTestFile looks for a test file in the same directory whose name
// matches the production file via a language-idiomatic rule.
func hasSiblingTestFile(prodPath string, testFiles map[string]struct{}) bool {
	p := filepath.ToSlash(prodPath)
	dir := filepath.Dir(p)
	base := filepath.Base(p)
	ext := filepath.Ext(base)
	if ext == "" {
		return false
	}
	stem := strings.TrimSuffix(base, ext)

	candidates := []string{
		dir + "/" + stem + "_test" + ext, // Go, Python: foo_test.go / foo_test.py
		dir + "/" + stem + ".test" + ext, // JS/TS: foo.test.ts
		dir + "/" + stem + ".spec" + ext, // JS/TS: foo.spec.ts
		dir + "/" + "test_" + stem + ext, // Python: test_foo.py
		dir + "/" + stem + "_spec" + ext, // Ruby: foo_spec.rb (+ other .ext)
		dir + "/" + stem + "Test" + ext,  // Java/Kotlin: FooTest.java
		dir + "/" + stem + "Tests" + ext, // Java variant
		dir + "/" + stem + "Spec" + ext,  // Scala: FooSpec.scala
		dir + "/" + "Test" + stem + ext,  // Java: TestFoo.java (older style)
	}
	for _, c := range candidates {
		if _, ok := testFiles[c]; ok {
			return true
		}
	}
	return false
}
