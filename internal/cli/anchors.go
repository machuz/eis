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
	"sort"
	"strings"

	"github.com/machuz/eis/v2/internal/output"
)

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

	domainResults, _, _, err := RunAnalyzePipeline(opts, pathArgs)
	if err != nil {
		return err
	}

	var stats []AnchorStat
	for _, dr := range domainResults {
		stats = append(stats, dr.Anchors...)
	}
	report := buildAnchors(stats, *topN, readDigest)

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
func buildAnchors(stats []AnchorStat, topN int, readDigest func(absPath string) (string, [2]int)) output.AnchorsReport {
	byModule := make(map[string][]AnchorStat)
	for _, s := range stats {
		if s.Lines <= 0 || s.Contributors < anchorMinContributors {
			continue // not load-bearing / not contested — not an anchor.
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
		if topN >= 0 && len(files) > topN {
			files = files[:topN]
		}
		anchors := make([]output.Anchor, 0, len(files))
		for _, s := range files {
			digest, lineRange := "", [2]int{0, 0}
			if readDigest != nil {
				digest, lineRange = readDigest(s.AbsPath)
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

// readDigest returns a compact representative excerpt of a file plus the 1-based
// line range it covers. It skips leading blank lines and caps by lines and
// chars (token budget). Unreadable files yield ("", [0,0]).
func readDigest(absPath string) (string, [2]int) {
	f, err := os.Open(absPath)
	if err != nil {
		return "", [2]int{0, 0}
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineNo := 0
	first := 0
	var b strings.Builder
	kept := 0
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		if kept == 0 && strings.TrimSpace(line) == "" {
			continue // skip leading blanks
		}
		if kept == 0 {
			first = lineNo
		}
		if b.Len()+len(line)+1 > anchorDigestMaxChars || kept >= anchorDigestMaxLines {
			break
		}
		if kept > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(strings.TrimRight(line, " \t"))
		kept++
	}
	if kept == 0 {
		return "", [2]int{0, 0}
	}
	return b.String(), [2]int{first, first + kept - 1}
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }
func round2(v float64) float64 { return math.Round(v*100) / 100 }
