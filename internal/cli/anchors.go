package cli

// eis anchors — return each module's surviving EXEMPLAR code (propagation
// anchors / positive examples): the regions that scored high on
// RobustSurvival × gravity, i.e. code that survived AND that others built on.
// The agent receives "write toward this surviving pattern".
//
// Uses the FULL pipeline (survival/gravity are needed; lean drops them). Selection
// reuses the pipeline's per-file surviving-mass capture (metric.CalcFileSurvival,
// same decay as survival) — no new heavy computation is added.
//
// FIREWALL: anchors are code + weights, never authorship. contested_by_n is an
// anonymous count; no owner name is emitted.
//
// TODO(wiring): a follow-up wires these digests into write-index's reserved
// `surviving_idiom_digest` slot. v0 is a standalone command.

import (
	"bufio"
	"flag"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/machuz/eis/v2/internal/metric"
	"github.com/machuz/eis/v2/internal/output"
)

// testDirSegments are directory names that hold test code — a file under any of
// them is excluded from anchors (test code is not an exemplar to imitate). This
// catches conventions metric.IsTestFile (which only reads the filename) misses,
// e.g. immutable-js's `type-definitions/ts-tests/`.
var testDirSegments = map[string]bool{
	"test": true, "tests": true, "__tests__": true,
	"spec": true, "specs": true, "e2e": true,
	"testdata": true, "ts-tests": true,
}

// isTestPath reports whether a path is test code, by filename convention OR by
// living under a test directory.
func isTestPath(path string) bool {
	if metric.IsTestFile(path) {
		return true
	}
	for _, seg := range strings.Split(filepath.ToSlash(path), "/") {
		if testDirSegments[strings.ToLower(seg)] {
			return true
		}
	}
	return false
}

const (
	anchorDigestMaxLines  = 12  // compact: token budget (rides to the agent).
	anchorDigestMaxChars  = 600 // hard cap on the excerpt.
	anchorMinContributors = 2   // robust gate: others must have built on it.
)

func runAnchors(args []string) error {
	fs := flag.NewFlagSet("anchors", flag.ExitOnError)
	configPath := fs.String("config", "", "Config file path")
	sampleSize := fs.Int("sample", 0, "Max files to blame per repo (overrides config)")
	workers := fs.Int("workers", 4, "Number of concurrent blame workers")
	recursive := fs.Bool("recursive", false, "Recursively find git repos under given paths")
	maxDepth := fs.Int("depth", 2, "Max directory depth for recursive search")
	tau := fs.Float64("tau", 0, "Survival decay parameter (overrides config)")
	topN := fs.Int("top", 3, "Anchors to return per module")
	formatFlag := fs.String("format", "json", "Output format: json, table")
	// Exemplars must be CORE source: exclude non-source modules (examples/docs/
	// website/samples/fixtures/root) and test files by default. A declared
	// architecture (module_patterns) suppresses the module blocklist, same as
	// structural-debt. --no-default-excludes turns both off.
	noDefaultExcludes := fs.Bool("no-default-excludes", false, "Do not exclude non-source modules and test files from anchors")
	verbose := fs.Bool("verbose", false, "Show detailed debug output")
	noCache := fs.Bool("no-cache", false, "Skip disk cache")

	flagArgs, pathArgs := separateArgs(args, fs)
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	explicitConfig := *configPath != ""
	if !explicitConfig {
		*configPath = "eis.yaml"
	}

	opts := AnalyzeOptions{
		ConfigPath:     *configPath,
		ExplicitConfig: explicitConfig,
		SampleSize:     *sampleSize,
		Workers:        *workers,
		Recursive:      *recursive,
		MaxDepth:       *maxDepth,
		Tau:            *tau,
		// json keeps the shared pipeline quiet. anchors needs the FULL pipeline
		// (survival/gravity), so LeanDebt stays false.
		Format:         "json",
		Verbose:        *verbose,
		NoCache:        *noCache,
		CaptureAnchors: true,
	}

	domainResults, cfg, _, err := RunAnalyzePipeline(opts, pathArgs)
	if err != nil {
		return err
	}

	// Module blocklist follows structural-debt (suppressed when the user declared
	// an architecture); test-file exclusion is anchors-specific (test code is not
	// an exemplar to imitate). --no-default-excludes turns both off.
	applyModuleBlocklist := !*noDefaultExcludes && !architectureDeclared(cfg)
	excludeTests := !*noDefaultExcludes

	var stats []AnchorStat
	for _, dr := range domainResults {
		stats = append(stats, dr.Anchors...)
	}
	report := buildAnchors(stats, *topN, applyModuleBlocklist, excludeTests, readDigest)

	switch *formatFlag {
	case "table":
		output.PrintAnchorsTable(report)
		return nil
	default:
		return output.EncodeAnchorsJSON(os.Stdout, report)
	}
}

// buildAnchors is the pure, testable core: group per-file stats by module, keep
// only others-contested files (robust gate), rank by survival × gravity, take
// top-N, and render each with a compact digest. readDigest is injected so tests
// need no filesystem.
func buildAnchors(stats []AnchorStat, topN int, applyModuleBlocklist, excludeTests bool, readDigest func(absPath string) (string, [2]int)) output.AnchorsReport {
	byModule := make(map[string][]AnchorStat)
	for _, s := range stats {
		if s.Lines <= 0 || s.Contributors < anchorMinContributors {
			continue // not load-bearing / not contested — not an anchor.
		}
		// Exemplars must be core source: drop non-source modules and test files
		// so the agent isn't told to imitate a config/example/test.
		if applyModuleBlocklist && isNonSourceModule(s.Module) {
			continue
		}
		if excludeTests && isTestPath(s.File) {
			continue
		}
		byModule[s.Module] = append(byModule[s.Module], s)
	}

	// score = mean survival × surviving gravity = RobustSurvival×gravity proxy.
	score := func(s AnchorStat) float64 {
		return (s.DecayedMass / float64(s.Lines)) * s.DecayedMass
	}

	mods := make([]string, 0, len(byModule))
	for m := range byModule {
		mods = append(mods, m)
	}
	sort.Strings(mods)

	report := output.AnchorsReport{Modules: []output.AnchorModule{}}
	for _, m := range mods {
		files := byModule[m]
		sort.Slice(files, func(i, j int) bool {
			si, sj := score(files[i]), score(files[j])
			if si != sj {
				return si > sj
			}
			return files[i].File < files[j].File // stable tie-break
		})
		// Walk the ranked candidates, reading digests, and keep the first topN
		// that have a real-logic digest. A file with no logic block (e.g. a pure
		// re-export barrel) is NOT an exemplar, so it's skipped rather than
		// emitted with an empty digest.
		anchors := make([]output.Anchor, 0, topN)
		for _, s := range files {
			if topN >= 0 && len(anchors) >= topN {
				break
			}
			digest, lineRange := "", [2]int{0, 0}
			if readDigest != nil {
				digest, lineRange = readDigest(s.AbsPath)
			}
			if digest == "" {
				continue // no representative logic — not a useful anchor.
			}
			anchors = append(anchors, output.Anchor{
				File:         s.File,
				LineRange:    lineRange,
				Survival:     round2(s.DecayedMass / float64(s.Lines)),
				Gravity:      round1(s.DecayedMass),
				ContestedByN: s.Contributors,
				Digest:       digest,
			})
		}
		if len(anchors) > 0 {
			report.Modules = append(report.Modules, output.AnchorModule{Module: m, Anchors: anchors})
		}
	}
	return report
}

// readDigest returns a compact representative excerpt of a file's REAL LOGIC
// (skipping the import/package/comment preamble) plus the 1-based line range it
// covers. Unreadable files yield ("", [0,0]).
func readDigest(absPath string) (string, [2]int) {
	f, err := os.Open(absPath)
	if err != nil {
		return "", [2]int{0, 0}
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var lines []string
	for sc.Scan() && len(lines) < 4000 { // safety cap on huge files
		lines = append(lines, sc.Text())
	}
	return digestFromLines(lines)
}

// digestFromLines picks a compact excerpt beginning at the file's first real
// logic line — skipping the leading package/import/comment preamble so the agent
// sees an actual implementation exemplar, not an import list. Caps at
// anchorDigestMaxLines / anchorDigestMaxChars (token budget). Pure + testable.
func digestFromLines(lines []string) (string, [2]int) {
	start := firstLogicLine(lines) // 0-based index, -1 if none
	if start < 0 {
		return "", [2]int{0, 0}
	}
	var b strings.Builder
	kept := 0
	for i := start; i < len(lines) && kept < anchorDigestMaxLines; i++ {
		line := strings.TrimRight(lines[i], " \t")
		if b.Len()+len(line)+1 > anchorDigestMaxChars && kept > 0 {
			break
		}
		if kept > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line)
		kept++
	}
	if kept == 0 {
		return "", [2]int{0, 0}
	}
	return b.String(), [2]int{start + 1, start + kept} // 1-based inclusive
}

// firstLogicLine returns the 0-based index of the first line that is real logic
// (not blank, not a comment, not a package/import/re-export statement, not inside
// an import block). Returns -1 when the file is all preamble.
func firstLogicLine(lines []string) int {
	inImportBlock := false
	for i, raw := range lines {
		t := strings.TrimSpace(raw)
		if inImportBlock {
			// The block closer starts with ")" (Go) or "}" (TS/JS named import,
			// e.g. "} from './types'") — match on prefix, not suffix, since the
			// TS form ends with the quoted module path.
			if strings.HasPrefix(t, ")") || strings.HasPrefix(t, "}") {
				inImportBlock = false
			}
			continue
		}
		if t == "" || isCommentLine(t) {
			continue
		}
		if isImportLine(t) {
			// A parenthesized/braced import block: skip its body too.
			if strings.HasSuffix(t, "(") || (strings.HasSuffix(t, "{") && !strings.Contains(t, "}")) {
				inImportBlock = true
			}
			continue
		}
		return i // first real logic line
	}
	return -1
}

func isCommentLine(t string) bool {
	for _, p := range []string{"//", "#", "/*", "*/", "*", "<!--", "--", ";;"} {
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	return false
}

func isImportLine(t string) bool {
	// Language-agnostic preamble statements (imports, package decls, re-exports).
	prefixes := []string{
		"package ", "package\t",
		"import ", "import\t", "import(", "import{",
		"from ", "require(", "require ",
		"#include", "#import", "@import",
		"using ", "use ",
		"export {", "export type {", "export * from", "export {}",
		"'use strict'", "\"use strict\"",
		"module ", "namespace ",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	return t == "import" || t == "package"
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }
func round2(v float64) float64 { return math.Round(v*100) / 100 }
