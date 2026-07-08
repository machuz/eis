package cli

// eis structural-debt — Tier-1 Structural Debt Meter (AI-agnostic).
//
// Headline number a CTO reads instantly:
//
//	"38% of your classified source mass is structural debt (Dead/Orphaned),
//	 and 15% more is one departure away (bus-factor 1)."
//
// Reuses the existing analysis pipeline + per-module classification
// (scorer.ModuleScore.{Vitality, Ownership, OwnerActive, BlameLines}). Only NEW
// logic here is the mass-weighted aggregation into SDR.
//
// Deliberately NOT in v0 (tracked as TODOs):
//   - trend / time-series  -> SaaS stores per-run snapshots (cheaper than making
//     the CLI re-run windowed analysis). See TODO(trend) in output.
//   - AI-authorship lens    -> tier-2 (needs authorship tagging). This command is
//     intentionally AI-agnostic: it opens the door on fear, no thesis needed.
//   - significant-lines     -> --strip-blank / --strip-comments. v0 uses
//     BlameLines as-is. See TODO(significant-lines) below.
//
// NOTE: distinct from internal/metric/debt.go, which is the author-level "負債
// 清掃" (DebtBalance / debt cleanup) axis — a different concept entirely.

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/machuz/eis/v2/internal/config"
	"github.com/machuz/eis/v2/internal/metric"
	"github.com/machuz/eis/v2/internal/output"
	"github.com/machuz/eis/v2/internal/scorer"
)

// debtOwnerGoneDaysDefault is the horizon at which an Orphaned module's absent
// owner counts as "gone" for structural debt — deliberately DECOUPLED from the
// general activeDays (~30d), which is a "who's around this sprint" notion. A
// codebase owner idle for a month is on holiday; idle for half a year has left.
// 180d avoids the threshold artifact where a still-active maintainer whose last
// commit happens to be 35 days before the analysis window mints false debt.
const debtOwnerGoneDaysDefault = 180

// debtStaleDaysDefault is the horizon past which a module counts as stale (no
// one has touched it). Debt requires BOTH the owner to be gone AND the module
// itself to be stale — a departed original author whose code is still actively
// maintained by others is inherited-but-alive, not abandoned. This mirrors the
// RobustSurvival idea that others-contested code is not debt: immutable-js's
// __tests__ (owner left 1506d ago, but touched 8 days ago) is NOT structural
// debt, it is living code with a new steward.
const debtStaleDaysDefault = 180

// defaultNonSourceDirs are top-level directory names whose modules are NOT
// production source and so are excluded from BOTH the debt numerator and the
// classified-mass denominator by default. Tests are deliberately absent — test
// debt is a legitimate concern. Overridden when the user declares an
// architecture (eis.yaml module_patterns) or passes --no-default-excludes.
var defaultNonSourceDirs = map[string]bool{
	"examples": true, "example": true,
	"docs": true, "doc": true,
	"website": true,
	"samples": true, "sample": true,
	"fixtures": true,
}

func runStructuralDebt(args []string) error {
	fs := flag.NewFlagSet("structural-debt", flag.ExitOnError)
	configPath := fs.String("config", "", "Config file path")
	tau := fs.Float64("tau", 0, "Survival decay parameter (overrides config)")
	sampleSize := fs.Int("sample", 0, "Max files to blame per repo (overrides config)")
	workers := fs.Int("workers", 4, "Number of concurrent blame workers")
	recursive := fs.Bool("recursive", false, "Recursively find git repos under given paths")
	maxDepth := fs.Int("depth", 2, "Max directory depth for recursive search")
	formatFlag := fs.String("format", "table", "Output format: table, json, csv")
	pressureMode := fs.String("pressure-mode", "include", "Change pressure mode: include or ignore")
	activeDays := fs.Int("active-days", 0, "Days to consider an owner active (overrides config)")
	domainFilter := fs.String("domain", "", "Only analyze repos in this domain")
	topN := fs.Int("top", 10, "How many worst debt modules to list")
	// Denominator policy for the PeripheralModule sentinel (folded, unclassified
	// mass). Default excludes it from BOTH numerator and denominator so SDR reads
	// as "debt / classified mass". NOTE: scorer.ScoreModules already drops the
	// PeripheralModule sentinel from its output, so in practice the sentinel never
	// reaches computeStructuralDebt and this flag is a defensive guard only —
	// kept so the denominator policy is explicit and future-proof.
	includePeripheral := fs.Bool("include-peripheral", false, "Count folded/peripheral mass in the denominator")
	// Debt owner-gone horizon: an Orphaned module counts as debt only once its
	// owner has been idle at least this long. Separate from --active-days so the
	// debt bar ("owner has left") is stricter than the general activity bar.
	debtOwnerGoneDays := fs.Int("debt-owner-gone-days", debtOwnerGoneDaysDefault, "Days the module owner must be idle before an Orphaned module counts as structural debt")
	// A module counts as debt only if it is ALSO stale (untouched by anyone for
	// this many days). Symmetric with --debt-owner-gone-days: debt = owner gone
	// AND module stale. Guards against flagging actively-maintained inherited code.
	debtStaleDays := fs.Int("debt-stale-days", debtStaleDaysDefault, "Days a module must be untouched by anyone before it can count as structural debt")
	// Non-source modules (examples/docs/website/root catch-all) are excluded from
	// both numerator and denominator by default; --no-default-excludes disables
	// that, and a declared architecture (eis.yaml module_patterns) always wins.
	noDefaultExcludes := fs.Bool("no-default-excludes", false, "Do not apply the built-in non-source module blocklist (examples/docs/website/...)")
	// structural-debt runs the LEAN pipeline by default: it computes only what
	// debt consumes (ownership, per-module commits + last-touch, author dates) and
	// skips the heavy science (survival/gravity, co-change pressure, breadth,
	// design, catalysis, debt-cleanup re-blame) so it scales to large repos and
	// costs less. --full runs the complete analysis pipeline instead; SDR and
	// top_debt_modules are identical either way (verified in tests).
	full := fs.Bool("full", false, "Run the full analysis pipeline instead of the lean debt-only path (same SDR, slower)")
	// TODO(significant-lines): blank/comment exclusion. v0 uses BlameLines as-is.
	// Add --strip-blank (cheap) then --strip-comments (needs a light per-lang lexer).
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
		Tau:            *tau,
		SampleSize:     *sampleSize,
		Workers:        *workers,
		Recursive:      *recursive,
		MaxDepth:       *maxDepth,
		Format:         *formatFlag,
		PressureMode:   *pressureMode,
		ActiveDays:     *activeDays,
		DomainFilter:   *domainFilter,
		Verbose:        *verbose,
		NoCache:        *noCache,
		// Lean by default: skip the heavy science debt never reads. --full opts
		// back into the complete pipeline (identical SDR, just slower).
		LeanDebt: !*full,
		// -M move detection isn't needed for debt-grade ownership and is a big
		// share of blame time. Off for BOTH lean and --full so they stay
		// byte-equivalent (only the science differs).
		BlameMoveDetection: "off",
	}

	// Reuse the shared pipeline — excludePatterns (vendored/generated) are already
	// applied upstream, so the modules we see are post-exclude. Non-graduated
	// modules are already folded into PeripheralModule by the liveness gate.
	domainResults, cfg, _, err := RunAnalyzePipeline(opts, pathArgs)
	if err != nil {
		return err
	}

	// Default non-source excludes apply only when the user has NOT declared an
	// architecture (module_patterns) — "core is decided by intent" — and did not
	// opt out with --no-default-excludes.
	applyDefaultExcludes := !*noDefaultExcludes && !architectureDeclared(cfg)

	var reports []output.StructuralDebtReport
	for _, dr := range domainResults {
		reports = append(reports, computeStructuralDebt(
			string(dr.Domain), dr.ModuleScores, ownerNames(dr.Ownership),
			*includePeripheral, applyDefaultExcludes,
			float64(*debtOwnerGoneDays), float64(*debtStaleDays), *topN))
	}
	if len(reports) == 0 {
		return fmt.Errorf("no modules to analyze")
	}

	switch opts.Format {
	case "json":
		return output.PrintStructuralDebtJSON(reports)
	case "csv":
		output.PrintStructuralDebtCSV(reports)
	default:
		output.PrintStructuralDebtTable(reports)
	}
	return nil
}

// ownerNames builds a module -> top-blame-author lookup from the accumulated
// ownership records. Mirrors scorer.ScoreModules' own ownershipMap build (last
// record wins when a module appears across repos), so the owner name shown for a
// debt module matches the one that drove its classification.
func ownerNames(ownership []metric.ModuleOwnership) map[string]string {
	if len(ownership) == 0 {
		return nil
	}
	names := make(map[string]string, len(ownership))
	for _, o := range ownership {
		names[o.Module] = o.TopAuthor
	}
	return names
}

// computeStructuralDebt is the pure, testable core. Mass = BlameLines (blame-
// derived surviving lines — the SAME source that determines ownership, so mass
// and orphan status can never disagree). The debt set and its severity tiers
// come straight from the existing 3-axis classification.
//
// ownerNames maps module -> last owner name for the drill rows; pass nil to skip
// owner enrichment (unit tests do).
//
// applyDefaultExcludes drops non-source modules (examples/docs/website/root) from
// BOTH numerator and denominator. debtOwnerGoneDays is the owner-idle horizon an
// Orphaned module must clear to count as debt (decoupled from activeDays);
// debtStaleDays is the module-untouched horizon a module (either tier) must clear
// — debt requires BOTH owner-gone AND module-stale.
func computeStructuralDebt(domain string, mods []scorer.ModuleScore, ownerNames map[string]string, includePeripheral bool, applyDefaultExcludes bool, debtOwnerGoneDays, debtStaleDays float64, topN int) output.StructuralDebtReport {
	var total, dead, orphaned, atRisk, classifiedCount int
	var debtMods []output.DebtModule

	for _, m := range mods {
		// The PeripheralModule sentinel is folded/unclassified mass — never
		// "debt", and excluded from the denominator by default. (scorer.ScoreModules
		// already drops it, so this rarely fires; kept as a defensive guard.)
		if m.Module == metric.PeripheralModule && !includePeripheral {
			continue
		}
		// Non-source modules (examples/docs/website/root catch-all) are not
		// production code — exclude from BOTH numerator and denominator so a
		// counter demo can't masquerade as core debt.
		if applyDefaultExcludes && isNonSourceModule(m.Module) {
			continue
		}
		mass := m.BlameLines
		if mass <= 0 {
			continue
		}
		total += mass
		classifiedCount++

		// Debt requires the module ITSELF to be stale, not just its owner absent.
		// moduleStale is true when nobody has touched it for debtStaleDays, OR
		// when it has no commits at all in the window (ModuleCommits==0 ⇒ no
		// last-touch date ⇒ maximally stale, the Dead case). A module still being
		// edited by others is inherited-but-alive, not abandoned — this mirrors
		// RobustSurvival (others-contested code is not debt). immutable-js's
		// __tests__ (owner left 1506d, but untouched only 8d) is correctly spared.
		moduleStale := m.ModuleUntouchedDays >= debtStaleDays || m.ModuleCommits == 0

		// Orphaned counts as debt only once the owner has been idle past the debt
		// horizon AND the module is stale. A still-active maintainer whose last
		// commit merely predates the window by a few weeks is NOT gone — that was
		// the redux "Erikson left 35d" false positive.
		deadDebt := m.Vitality == "Dead" && moduleStale
		orphanedDebt := m.Ownership == "Orphaned" && m.OwnerLastActiveDays >= debtOwnerGoneDays && moduleStale
		isDebt := deadDebt || orphanedDebt
		switch {
		case deadDebt:
			dead += mass
		case orphanedDebt:
			orphaned += mass
		case m.Ownership == "Concentrated" && m.OwnerActive:
			// Leading indicator: NOT counted in SDR (SDR = already-unowned mass).
			atRisk += mass
		}

		if isDebt {
			debtMods = append(debtMods, output.DebtModule{
				Module:    m.Module,
				Mass:      mass,
				Tier:      debtTier(m), // "Dead" | "Orphaned"
				LastOwner: ownerNames[m.Module],
				// owner_left_days: days since the module's owner last committed
				// (for a departed owner, "left N days ago"). untouched_days: days
				// since ANY commit touched the module. Both come straight from the
				// classifier (0 when the underlying date wasn't available).
				OwnerLeftDays: int(m.OwnerLastActiveDays),
				UntouchedDays: int(m.ModuleUntouchedDays),
			})
		}
	}

	sortDescByMass(debtMods)
	// Concentration transparency: surface (don't hide) if one module dominates.
	conc := concentration(debtMods, dead+orphaned)
	if topN >= 0 && len(debtMods) > topN {
		debtMods = debtMods[:topN]
	}

	return output.StructuralDebtReport{
		Domain:         domain,
		SDR:            ratio(dead+orphaned, total), // headline
		DeadRatio:      ratio(dead, total),
		OrphanedRatio:  ratio(orphaned, total),
		AtRiskRatio:    ratio(atRisk, total), // leading indicator
		ClassifiedMass: total,
		DebtMass:       dead + orphaned,
		ModuleCount:    classifiedCount,
		TopDebtModules: debtMods,
		Concentration:  conc,
		// No classified source mass ⇒ SDR is undefined, not a healthy 0. Flag it
		// so a consumer never reads "no data" as "clean".
		InsufficientData: total == 0,
		// TODO(gravity-variant): a secondary SDR weighted by survival mass
		// (metric.CalcModuleSurvival) could carry the thesis story — keep it OFF
		// the headline.
	}
}

func ratio(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) / float64(d)
}

// architectureDeclared reports whether the user has declared a module
// architecture (eis.yaml module_patterns, org-level or per-repo). When true, the
// default non-source blocklist is suppressed — a declared "examples" module is
// then intentional and the architecture decides what core is.
func architectureDeclared(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	if len(cfg.ModulePatterns) > 0 {
		return true
	}
	for _, ov := range cfg.RepoOverrides {
		if len(ov.ModulePatterns) > 0 {
			return true
		}
	}
	return false
}

// isNonSourceModule reports whether a folded module id is a non-source area
// (examples/docs/website/... or the root "." catch-all) that should not count
// toward structural debt. Matches on the first path segment so nested modules
// like "examples/counter-ts" are caught. Tests are intentionally NOT excluded.
func isNonSourceModule(mod string) bool {
	if mod == "" || mod == "." {
		return true // root catch-all / unrooted miscellany
	}
	seg := mod
	if i := strings.IndexByte(mod, '/'); i >= 0 {
		seg = mod[:i]
	}
	return defaultNonSourceDirs[seg]
}

func debtTier(m scorer.ModuleScore) string {
	if m.Vitality == "Dead" {
		return "Dead"
	}
	return "Orphaned"
}

// sortDescByMass orders debt modules by mass descending, tie-broken by module id
// so equal-mass modules keep a stable, reproducible order.
func sortDescByMass(d []output.DebtModule) {
	sort.Slice(d, func(i, j int) bool {
		if d[i].Mass != d[j].Mass {
			return d[i].Mass > d[j].Mass
		}
		return d[i].Module < d[j].Module
	})
}

// concentration reports the single largest debt module's share of total debt
// mass. Assumes d is already sorted descending by mass.
func concentration(d []output.DebtModule, debtMass int) output.Concentration {
	if len(d) == 0 || debtMass <= 0 {
		return output.Concentration{}
	}
	return output.Concentration{
		TopModule: d[0].Module,
		Share:     float64(d[0].Mass) / float64(debtMass),
	}
}
