package cli

// eis write-index — emit the per-module "read before you write" index an AI
// coding agent (via the pre-write hook) consults locally to resolve a target
// module → structural facts, without hitting a remote every time.
//
// Reuses the LEAN debt pipeline (#355): fast, and near-free on a warm cache. It
// covers EVERY module (an agent may write anywhere), unlike structural-debt's
// top-N human drill. Debt classification is the SAME shared logic
// (classifyModuleDebt), so write-index and structural-debt can never disagree.
//
// FIREWALL: this is an agent-facing artifact, so it carries NO owner names — only
// structural facts. (structural-debt's human table keeps last_owner; this does
// not.)

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/machuz/eis/v2/internal/config"
	"github.com/machuz/eis/v2/internal/git"
	"github.com/machuz/eis/v2/internal/output"
)

const defaultWriteIndexPath = ".eis/write-index.json"

func runWriteIndex(args []string) error {
	fs := flag.NewFlagSet("write-index", flag.ExitOnError)
	configPath := fs.String("config", "", "Config file path")
	sampleSize := fs.Int("sample", 0, "Max files to blame per repo (overrides config)")
	workers := fs.Int("workers", 4, "Number of concurrent blame workers")
	recursive := fs.Bool("recursive", false, "Recursively find git repos under given paths")
	maxDepth := fs.Int("depth", 2, "Max directory depth for recursive search")
	activeDays := fs.Int("active-days", 0, "Days to consider an owner active (overrides config)")
	debtOwnerGoneDays := fs.Int("debt-owner-gone-days", debtOwnerGoneDaysDefault, "Days the module owner must be idle before an Orphaned module counts as debt")
	debtStaleDays := fs.Int("debt-stale-days", debtStaleDaysDefault, "Days a module must be untouched before it can count as debt")
	outputPath := fs.String("output", defaultWriteIndexPath, "Where to write the index; \"-\" for stdout")
	verbose := fs.Bool("verbose", false, "Show detailed debug output")
	noCache := fs.Bool("no-cache", false, "Skip disk cache")
	// Idiom-propagation payloads (Build2) — OFF by default so the base index stays
	// lean and near-free. The index-refresh job opts in; the pre-write hook then
	// serves "write toward these" (anchors) / "don't reintroduce this" (graveyard)
	// from the same local lookup. --anchors rides the lean pipeline (it needs only
	// the blame the index already computes); --graveyard adds a separate numstat
	// death walk, so it is the more expensive of the two.
	withAnchors := fs.Bool("anchors", false, "Embed per-module surviving-exemplar digests (propagation anchors)")
	withGraveyard := fs.Bool("graveyard", false, "Embed per-module dead-pattern hotspots (anchors' complement)")
	anchorsTop := fs.Int("anchors-top", 3, "Anchors to embed per module (with --anchors)")
	graveyardTop := fs.Int("graveyard-top", 5, "Graveyard hotspots to embed per module (with --graveyard)")

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
		// json keeps the shared pipeline quiet (no progress spam on stderr).
		Format:     "json",
		ActiveDays: *activeDays,
		Verbose:    *verbose,
		NoCache:    *noCache,
		// Same lean, blame-move-off configuration as structural-debt: only what
		// debt reads, and identical ownership attribution.
		LeanDebt:           true,
		BlameMoveDetection: "off",
		// Anchor capture folds per-file surviving mass from the blame the lean
		// pipeline already computes (metric.CalcFileSurvival) — it stays lean, no
		// heavy science is switched on. Only set when --anchors is requested.
		CaptureAnchors: *withAnchors,
	}

	domainResults, cfg, _, err := RunAnalyzePipeline(opts, pathArgs)
	if err != nil {
		return err
	}

	// HEAD sha of the first analyzed repo stamps the index lineage (best-effort).
	commit := ""
	repoPath := "."
	if len(pathArgs) > 0 {
		repoPath = pathArgs[0]
	}
	if abs, aerr := filepath.Abs(repoPath); aerr == nil {
		if sha, herr := git.HeadHash(context.Background(), abs); herr == nil {
			commit = sha
		}
	}

	// Default excludes match anchors/graveyard: drop non-source modules (unless the
	// user declared an architecture) and test files, so the agent is never told to
	// imitate — or avoid — a config/example/test.
	applyModuleBlocklist := !architectureDeclared(cfg)
	excludeTests := true

	// --anchors: surviving exemplars, from the per-file stats the lean pipeline just
	// captured. nil map when off ⇒ modules omit the field.
	var anchorsByModule map[string][]output.Anchor
	if *withAnchors {
		var stats []AnchorStat
		for _, dr := range domainResults {
			stats = append(stats, dr.Anchors...)
		}
		report := buildAnchors(stats, *anchorsTop, applyModuleBlocklist, excludeTests, readDigest)
		anchorsByModule = make(map[string][]output.Anchor, len(report.Modules))
		for _, m := range report.Modules {
			anchorsByModule[m.Module] = m.Anchors
		}
	}

	// --graveyard: dead patterns, from a separate cheap numstat death walk (no
	// blame). The extra log parse is why this is opt-in, not part of the lean base.
	//
	// The payload is attached (below, in buildWriteIndex) only to modules the
	// pipeline actually scored. The death walk resolves modules over ALL history,
	// so it can surface a module the topology pipeline buckets into the peripheral
	// sentinel (tiny / low-liveness) and never emits as an index row — that
	// module's graveyard is intentionally dropped here so the index stays keyed to
	// its own denoised module set (i.e. `eis graveyard` may list a module this
	// index does not; that is the standalone view, not a disagreement).
	var graveyardByModule map[string]output.WriteIndexGraveyard
	if *withGraveyard {
		repoPaths, rerr := resolveRepoPaths(pathArgs, *recursive, *maxDepth)
		if rerr != nil {
			return rerr
		}
		graves, currentFiles := walkGraveyard(context.Background(), repoPaths, cfg, resolveWorkers(*workers))
		report := buildGraveyard(graves, currentFiles, *graveyardTop, applyModuleBlocklist, excludeTests)
		graveyardByModule = make(map[string]output.WriteIndexGraveyard, len(report.Modules))
		for _, m := range report.Modules {
			graveyardByModule[m.Module] = output.WriteIndexGraveyard{
				DeathIntensity: m.DeathIntensity,
				DeathEvents:    m.DeathEvents,
				Hotspots:       m.Hotspots,
			}
		}
	}

	idx := buildWriteIndex(domainResults, cfg, commit, time.Now().UTC(),
		float64(*debtOwnerGoneDays), float64(*debtStaleDays), anchorsByModule, graveyardByModule)

	if *outputPath == "-" {
		return output.EncodeWriteIndex(os.Stdout, idx)
	}
	if dir := filepath.Dir(*outputPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	f, err := os.Create(*outputPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", *outputPath, err)
	}
	defer f.Close()
	if err := output.EncodeWriteIndex(f, idx); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "✦ wrote %s (%d modules)\n", *outputPath, len(idx.Modules))
	return nil
}

// buildWriteIndex is the pure, testable core: every module across every domain,
// mapped to its structural facts and a shared debt verdict. No owner names.
// anchorsByModule / graveyardByModule are nil unless the caller requested the
// --anchors / --graveyard payloads; modules absent from them omit those fields.
func buildWriteIndex(domainResults []DomainResults, cfg *config.Config, commit string, generatedAt time.Time, debtOwnerGoneDays, debtStaleDays float64, anchorsByModule map[string][]output.Anchor, graveyardByModule map[string]output.WriteIndexGraveyard) output.WriteIndex {
	modules := make(map[string]output.WriteIndexModule)
	for _, dr := range domainResults {
		for _, m := range dr.ModuleScores {
			// Same debt logic as structural-debt (Dead > Orphaned > AtRisk).
			v := classifyModuleDebt(m, debtOwnerGoneDays, debtStaleDays)
			wim := output.WriteIndexModule{
				DebtTier:               debtTierOf(v),
				AtRisk:                 v.AtRisk,
				OwnerActive:            m.OwnerActive,
				OwnerLeftDays:          int(m.OwnerLastActiveDays),
				UntouchedDays:          int(m.ModuleUntouchedDays),
				OwnershipConcentration: m.TopAuthorShare,
				Recommendation:         writeIndexRecommendation(v),
				// NOTE (firewall): m.Module's owner name (metric.ModuleOwnership.
				// TopAuthor) is deliberately NOT emitted here.
			}
			if a := anchorsByModule[m.Module]; len(a) > 0 {
				wim.Anchors = a
			}
			if g, ok := graveyardByModule[m.Module]; ok {
				wim.Graveyard = &g
			}
			modules[m.Module] = wim
		}
	}

	patterns := config.DefaultModulePatterns
	source := "builtin_default"
	if cfg != nil && len(cfg.ModulePatterns) > 0 {
		patterns = cfg.ModulePatterns
		source = "config"
	}

	return output.WriteIndex{
		GeneratedAt:          generatedAt.Format(time.RFC3339),
		Commit:               commit,
		Modules:              modules,
		ModulePatterns:       patterns,
		ModulePatternsSource: source,
	}
}

// debtTierOf maps a verdict to the drill tier string (Dead precedence).
func debtTierOf(v debtVerdict) string {
	switch {
	case v.DeadDebt:
		return "Dead"
	case v.OrphanedDebt:
		return "Orphaned"
	default:
		return ""
	}
}

// writeIndexRecommendation maps a debt verdict to the agent-facing hint.
func writeIndexRecommendation(v debtVerdict) string {
	switch {
	case v.DeadDebt:
		return "dead_module"
	case v.OrphanedDebt:
		return "orphaned_module"
	case v.AtRisk:
		return "bus_factor_1"
	default:
		return ""
	}
}
