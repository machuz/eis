package cli

// eis graveyard — surface where past attempts DIED: code people wrote here and
// OTHERS tore out, repeatedly, in files that still exist. The complement of
// anchors (surviving exemplars) — the negative examples / hard spots.
//
// death (v0, structural):
//   - others-contested: author B (≠ A) deleted/rewrote A's lines (self-rewrite
//     A=A is healthy refactoring, excluded).
//   - repeated: >= 2 such deaths in the same file (a single one is a healthy
//     one-off, excluded).
//   - still exists: the file is present at HEAD (whole-file deletion = product
//     decision, excluded).
//
// Cheap: reconstructs a line genealogy from `git log --numstat` alone (no blame).
//
// FIREWALL: death_intensity / death_events / contested_by_n only — no author
// names.
//
// TODO(semantic): a follow-up (LLM) can say WHICH approach died and WHAT replaced
// it. v0 surfaces the structural fact only.

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/machuz/eis/v2/internal/config"
	"github.com/machuz/eis/v2/internal/git"
	"github.com/machuz/eis/v2/internal/metric"
	"github.com/machuz/eis/v2/internal/output"
)

// graveyardMinDeaths is the "repeated" threshold: a file needs at least this many
// others-contested deaths to be a hotspot (1 death = healthy one-off refactor).
const graveyardMinDeaths = 2

func runGraveyard(args []string) error {
	fs := flag.NewFlagSet("graveyard", flag.ExitOnError)
	configPath := fs.String("config", "", "Config file path")
	workers := fs.Int("workers", 0, "Concurrent workers for git log (0 = auto)")
	recursive := fs.Bool("recursive", false, "Recursively find git repos under given paths")
	maxDepth := fs.Int("depth", 2, "Max directory depth for recursive search")
	topN := fs.Int("top", 5, "Hotspot files to list per module")
	formatFlag := fs.String("format", "json", "Output format: json, table")
	// Graveyard is anchors' complement, so it uses the SAME code universe: exclude
	// non-source modules (examples/docs/website/...) and test files by default so
	// it surfaces CODE dead-ends, not doc/config churn. A declared architecture
	// suppresses the module blocklist; --no-default-excludes turns both off.
	noDefaultExcludes := fs.Bool("no-default-excludes", false, "Do not exclude non-source modules and test files")

	flagArgs, pathArgs := separateArgs(args, fs)
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	explicitConfig := *configPath != ""
	if !explicitConfig {
		*configPath = "eis.yaml"
	}
	cfg, err := config.Load(*configPath, explicitConfig)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx := context.Background()
	w := resolveWorkers(*workers)

	repoPaths := pathArgs
	if len(repoPaths) == 0 {
		repoPaths = []string{"."}
	}
	for i, p := range repoPaths {
		abs, aerr := filepath.Abs(p)
		if aerr != nil {
			return fmt.Errorf("resolve %s: %w", p, aerr)
		}
		repoPaths[i] = abs
	}
	if *recursive {
		var discovered []string
		for _, root := range repoPaths {
			repos, derr := findGitRepos(root, *maxDepth)
			if derr != nil {
				return fmt.Errorf("scan %s: %w", root, derr)
			}
			discovered = append(discovered, repos...)
		}
		if len(discovered) == 0 {
			return fmt.Errorf("no git repos found under %v", repoPaths)
		}
		repoPaths = discovered
	}

	var graves []metric.FileGrave
	currentFiles := make(map[string]bool)

	for _, repoPath := range repoPaths {
		if _, serr := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(serr) {
			fmt.Fprintf(os.Stderr, "SKIP: %s (not a git repo)\n", repoPath)
			continue
		}
		repoName := filepath.Base(repoPath)
		if cfg.IsExcludedRepo(repoName) {
			continue
		}
		repoCfg := config.EffectiveConfigForRepo(cfg, repoName)

		// numstat-only log (no -p, no blame): cheap. Reverse-chronological, so
		// reverse to chronological for the inventory walk.
		commits, perr := git.ParseLogParallel(ctx, repoPath, w, false)
		if perr != nil {
			fmt.Fprintf(os.Stderr, "SKIP: %s (log: %v)\n", repoName, perr)
			continue
		}
		// Same commit canonicalization/filtering the analysis pipeline applies, so
		// B≠A is accurate (alias-collapsed) and excluded authors/files are dropped.
		idmap := git.BuildIdentityMap(commits)
		git.CanonicalizeAuthors(commits, nil, idmap)
		commits = filterCommits(commits, cfg)
		excl := metric.NewExcluder(effectiveExcludes(ctx, repoPath, repoCfg))
		commits = filterFileStats(commits, excl)
		reverseCommits(commits)

		resolver := metric.NewModuleResolverWithExcludes(
			config.PatternsForRepo(cfg, repoName), config.ExcludesForRepo(cfg, repoName))

		for _, g := range metric.CalcGraveyard(commits, resolver) {
			graves = append(graves, *g)
		}

		// Files present at HEAD — the "still exists" gate.
		if files, ferr := git.ListAllFiles(ctx, repoPath); ferr == nil {
			for _, f := range files {
				currentFiles[f] = true
			}
		}
	}

	// Module blocklist follows structural-debt/anchors (suppressed when an
	// architecture is declared); test exclusion is on unless opted out.
	applyModuleBlocklist := !*noDefaultExcludes && !architectureDeclared(cfg)
	excludeTests := !*noDefaultExcludes

	report := buildGraveyard(graves, currentFiles, *topN, applyModuleBlocklist, excludeTests)

	switch *formatFlag {
	case "table":
		output.PrintGraveyardTable(report)
		return nil
	default:
		return output.EncodeGraveyardJSON(os.Stdout, report)
	}
}

// buildGraveyard is the pure, testable core: keep only still-existing files, roll
// per-file deaths up to modules, and surface modules that have at least one
// REPEATED-death hotspot (>= graveyardMinDeaths). death_intensity = others-
// contested dead lines in hotspots / total writes in the module.
func buildGraveyard(graves []metric.FileGrave, currentFiles map[string]bool, topN int, applyModuleBlocklist, excludeTests bool) output.GraveyardReport {
	type modAgg struct {
		totalAdds     int
		hotspotDead   int
		hotspotDeaths int
		hotspots      []output.GraveyardHotspot
	}
	mods := make(map[string]*modAgg)
	get := func(m string) *modAgg {
		a := mods[m]
		if a == nil {
			a = &modAgg{}
			mods[m] = a
		}
		return a
	}

	for _, g := range graves {
		if g.Module == "" || g.Module == metric.PeripheralModule {
			continue
		}
		// Same code universe as anchors: drop non-source modules and test files so
		// graveyard surfaces code dead-ends, not doc/config/test churn.
		if applyModuleBlocklist && isNonSourceModule(g.Module) {
			continue
		}
		if excludeTests && isTestPath(g.File) {
			continue
		}
		// "still exists": whole-file deletions (not at HEAD) are product decisions,
		// not graveyards. When currentFiles is nil, keep everything (test convenience).
		if currentFiles != nil && !currentFiles[g.File] {
			continue
		}
		a := get(g.Module)
		a.totalAdds += g.TotalAdds // denominator = all production in surviving files
		if g.Deaths >= graveyardMinDeaths {
			a.hotspotDead += g.OthersDeadLines
			a.hotspotDeaths += g.Deaths
			a.hotspots = append(a.hotspots, output.GraveyardHotspot{
				File:         g.File,
				Deaths:       g.Deaths,
				ContestedByN: g.Contributors,
			})
		}
	}

	names := make([]string, 0, len(mods))
	for m := range mods {
		names = append(names, m)
	}
	sort.Strings(names)

	report := output.GraveyardReport{Modules: []output.GraveyardModule{}}
	for _, m := range names {
		a := mods[m]
		if len(a.hotspots) == 0 {
			continue // no repeated death -> not a graveyard.
		}
		sort.Slice(a.hotspots, func(i, j int) bool {
			if a.hotspots[i].Deaths != a.hotspots[j].Deaths {
				return a.hotspots[i].Deaths > a.hotspots[j].Deaths
			}
			return a.hotspots[i].File < a.hotspots[j].File
		})
		if topN >= 0 && len(a.hotspots) > topN {
			a.hotspots = a.hotspots[:topN]
		}
		intensity := 0.0
		if a.totalAdds > 0 {
			intensity = float64(a.hotspotDead) / float64(a.totalAdds)
			if intensity > 1 {
				intensity = 1
			}
		}
		report.Modules = append(report.Modules, output.GraveyardModule{
			Module:         m,
			DeathIntensity: round2(intensity),
			DeathEvents:    a.hotspotDeaths,
			Hotspots:       a.hotspots,
		})
	}
	return report
}

func reverseCommits(c []git.Commit) {
	for i, j := 0, len(c)-1; i < j; i, j = i+1, j-1 {
		c[i], c[j] = c[j], c[i]
	}
}
