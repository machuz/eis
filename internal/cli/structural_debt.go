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

	"github.com/machuz/eis/v2/internal/metric"
	"github.com/machuz/eis/v2/internal/output"
	"github.com/machuz/eis/v2/internal/scorer"
)

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
	}

	// Reuse the shared pipeline — excludePatterns (vendored/generated) are already
	// applied upstream, so the modules we see are post-exclude. Non-graduated
	// modules are already folded into PeripheralModule by the liveness gate.
	domainResults, _, _, err := RunAnalyzePipeline(opts, pathArgs)
	if err != nil {
		return err
	}

	var reports []output.StructuralDebtReport
	for _, dr := range domainResults {
		reports = append(reports, computeStructuralDebt(
			string(dr.Domain), dr.ModuleScores, ownerNames(dr.Ownership), *includePeripheral, *topN))
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
func computeStructuralDebt(domain string, mods []scorer.ModuleScore, ownerNames map[string]string, includePeripheral bool, topN int) output.StructuralDebtReport {
	var total, dead, orphaned, atRisk, classifiedCount int
	var debtMods []output.DebtModule

	for _, m := range mods {
		// The PeripheralModule sentinel is folded/unclassified mass — never
		// "debt", and excluded from the denominator by default. (scorer.ScoreModules
		// already drops it, so this rarely fires; kept as a defensive guard.)
		if m.Module == metric.PeripheralModule && !includePeripheral {
			continue
		}
		mass := m.BlameLines
		if mass <= 0 {
			continue
		}
		total += mass
		classifiedCount++

		isDebt := m.Vitality == "Dead" || m.Ownership == "Orphaned"
		switch {
		case m.Vitality == "Dead":
			dead += mass
		case m.Ownership == "Orphaned":
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
