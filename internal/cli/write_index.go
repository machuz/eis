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

	idx := buildWriteIndex(domainResults, cfg, commit, time.Now().UTC(),
		float64(*debtOwnerGoneDays), float64(*debtStaleDays))

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
func buildWriteIndex(domainResults []DomainResults, cfg *config.Config, commit string, generatedAt time.Time, debtOwnerGoneDays, debtStaleDays float64) output.WriteIndex {
	modules := make(map[string]output.WriteIndexModule)
	for _, dr := range domainResults {
		for _, m := range dr.ModuleScores {
			// Same debt logic as structural-debt (Dead > Orphaned > AtRisk).
			v := classifyModuleDebt(m, debtOwnerGoneDays, debtStaleDays)
			modules[m.Module] = output.WriteIndexModule{
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
