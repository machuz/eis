package cli

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/machuz/eis/v2/internal/cache"
	"github.com/machuz/eis/v2/internal/config"
	"github.com/machuz/eis/v2/internal/domain"
	"github.com/machuz/eis/v2/internal/git"
	"github.com/machuz/eis/v2/internal/metric"
	"github.com/machuz/eis/v2/internal/output"
	"github.com/machuz/eis/v2/internal/scorer"
)

// AnalyzeOptions holds CLI flags for the analysis pipeline.
type AnalyzeOptions struct {
	ConfigPath     string
	ExplicitConfig bool // true when --config was explicitly passed
	Tau            float64
	SampleSize     int
	Workers        int
	Recursive      bool
	MaxDepth       int
	Format         string
	PressureMode   string
	ActiveDays     int
	DomainFilter   string
	Verbose        bool
	NoCache        bool
	PerRepo      bool
}

// DomainResults holds scored results for a single domain.
type DomainResults struct {
	Domain    domain.Domain
	Results   []scorer.Result
	Risks     []metric.ModuleRisk
	RepoCount int
	PerRepo   []RepoResult // per-repo breakdown (only when --per-repo is set)

	// Test coverage summary across all repos in this domain.
	TotalFiles     int     // total code files
	TotalTestFiles int     // how many of those look like tests
	TestFileRatio  float64 // convenience: TotalTestFiles / TotalFiles

	// Module Science Phase 1: direct structural measurement
	Cochange  []metric.CochangeResult    // per-repo co-change coupling (DSM)
	Ownership []metric.ModuleOwnership   // accumulated ownership fragmentation

	// Module Science Phase 2: 3-axis module topology
	ModuleScores []scorer.ModuleScore
}

// RepoResult holds scored results for a single repository.
type RepoResult struct {
	RepoName string
	Domain   domain.Domain
	Results  []scorer.Result
}

// domainAccumulator holds per-domain scoring state
type domainAccumulator struct {
	raw               *metric.RawScores
	qualityCounts     map[string]int
	debtCounts        map[string]int
	authorRepoCommits map[string]map[string]int // author -> repo -> commit count
	authorFirstDate   map[string]time.Time      // earliest commit date per author
	authorLastDate    map[string]time.Time      // latest commit date per author
	repoCount         int
	risks             []metric.ModuleRisk   // accumulated bus factor risks
	changePressure    metric.ChangePressure // accumulated change pressure across repos

	// Module Science Phase 1
	cochangeResults []metric.CochangeResult    // per-repo co-change coupling
	ownership       []metric.ModuleOwnership   // accumulated ownership fragmentation

	// Module Science Phase 2
	moduleSurvival      map[string]float64    // per-module survival rate (0-1)
	modulePressure      metric.ChangePressure // per-module change pressure (without repo prefix)
	modulePressureCounts map[string]int       // count for averaging across repos

	// Test coverage observability (populated from each repo's TestedSet)
	totalFiles     int // sum of code files across all repos in this domain
	totalTestFiles int // sum of test files across all repos in this domain

	// Per-module file counts across all repos in this domain — used by
	// ScoreModules to compute Vitality=Fragile. Keyed on ModuleOf(path).
	moduleAllFiles  map[string]int
	moduleTestFiles map[string]int
}

func newDomainAccumulator() *domainAccumulator {
	return &domainAccumulator{
		raw:               metric.NewRawScores(),
		qualityCounts:     make(map[string]int),
		debtCounts:        make(map[string]int),
		authorRepoCommits: make(map[string]map[string]int),
		authorFirstDate:   make(map[string]time.Time),
		authorLastDate:    make(map[string]time.Time),
		changePressure:       make(metric.ChangePressure),
		moduleSurvival:       make(map[string]float64),
		modulePressure:       make(metric.ChangePressure),
		modulePressureCounts: make(map[string]int),
		moduleAllFiles:       make(map[string]int),
		moduleTestFiles:      make(map[string]int),
	}
}

func runAnalyze(args []string) error {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	configPath := fs.String("config", "", "Config file path")
	tau := fs.Float64("tau", 0, "Survival decay parameter (overrides config)")
	sampleSize := fs.Int("sample", 0, "Max files to blame per repo (overrides config)")
	workers := fs.Int("workers", 4, "Number of concurrent blame workers")
	recursive := fs.Bool("recursive", false, "Recursively find git repos under given paths")
	maxDepth := fs.Int("depth", 2, "Max directory depth for recursive search")
	formatFlag := fs.String("format", "table", "Output format: table, csv, json")
	pressureMode := fs.String("pressure-mode", "include", "Change pressure mode: include (split robust/dormant) or ignore (classic survival)")
	activeDays := fs.Int("active-days", 0, "Days to consider author active (overrides config, default 30)")
	domainFilter := fs.String("domain", "", "Only analyze repos in this domain (e.g. Backend, Frontend, Firmware)")
	verbose := fs.Bool("verbose", false, "Show detailed debug output (file-level timing)")
	noCache := fs.Bool("no-cache", false, "Skip disk cache")
	perRepo := fs.Bool("per-repo", false, "Show per-repository breakdown (requires --recursive)")

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
		PerRepo:      *perRepo,
	}

	domainResults, cfg, err := RunAnalyzePipeline(opts, pathArgs)
	if err != nil {
		return err
	}

	return outputAnalyzeResults(domainResults, cfg, opts.Format)
}

func outputAnalyzeResults(domainResults []DomainResults, cfg *config.Config, format string) error {
	var jsonWriter *output.JSONWriter
	if format == "json" {
		jsonWriter = output.NewJSONWriter()
	}

	csvHeaderWritten := false

	for _, dr := range domainResults {
		switch format {
		case "json":
			jsonWriter.AddDomain(string(dr.Domain), dr.RepoCount, dr.Results, dr.Risks)
			jsonWriter.AddTestCoverage(string(dr.Domain), dr.TotalFiles, dr.TotalTestFiles, dr.TestFileRatio)
			jsonWriter.AddModuleScience(string(dr.Domain), dr.Cochange, dr.Ownership)
			jsonWriter.AddModuleScores(string(dr.Domain), dr.ModuleScores)
			for _, rr := range dr.PerRepo {
				jsonWriter.AddPerRepo(string(dr.Domain), rr.RepoName, rr.Results)
			}
		case "csv":
			output.PrintRankingsCSV(string(dr.Domain), dr.Results, !csvHeaderWritten)
			csvHeaderWritten = true
			for _, rr := range dr.PerRepo {
				output.PrintRankingsCSV(string(dr.Domain)+"/"+rr.RepoName, rr.Results, false)
			}
		default:
			fmt.Println()
			color.New(color.FgHiCyan, color.Bold).Printf("═══ %s ═══\n", dr.Domain)
			output.PrintSummary(dr.Results, dr.RepoCount)
			output.PrintRankings(dr.Results)

			if len(dr.ModuleScores) > 0 {
				output.PrintModuleArchetypes(dr.ModuleScores)
			}

			if len(dr.PerRepo) > 0 {
				perRepoData := make([]output.PerRepoData, len(dr.PerRepo))
				for i, rr := range dr.PerRepo {
					perRepoData[i] = output.PerRepoData{
						RepoName: rr.RepoName,
						Results:  rr.Results,
					}
				}
				output.PrintPerRepoComparison(string(dr.Domain), perRepoData, dr.Results)
			}
		}
	}

	if format == "json" {
		if err := jsonWriter.Flush(); err != nil {
			return fmt.Errorf("json output: %w", err)
		}
	}

	return nil
}

// RunAnalyzePipeline runs the full analysis pipeline and returns per-domain results.
// This is the shared core used by both `eis analyze` and `eis team`.
func RunAnalyzePipeline(opts AnalyzeOptions, paths []string) ([]DomainResults, *config.Config, error) {
	repoPaths := paths
	if len(repoPaths) == 0 {
		repoPaths = []string{"."}
	}

	// Resolve to absolute paths
	for i, p := range repoPaths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, nil, fmt.Errorf("resolve path %s: %w", p, err)
		}
		repoPaths[i] = abs
	}

	// Recursive: find git repos under given paths
	if opts.Recursive {
		var discovered []string
		for _, root := range repoPaths {
			repos, err := findGitRepos(root, opts.MaxDepth)
			if err != nil {
				return nil, nil, fmt.Errorf("scan %s: %w", root, err)
			}
			discovered = append(discovered, repos...)
		}
		if len(discovered) == 0 {
			return nil, nil, fmt.Errorf("no git repos found under %v (depth=%d)", repoPaths, opts.MaxDepth)
		}
		repoPaths = discovered
		fmt.Fprintf(os.Stderr, "Found %d git repos\n\n", len(repoPaths))
	}

	// Load config
	cfg, err := config.Load(opts.ConfigPath, opts.ExplicitConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}
	if opts.Tau > 0 {
		cfg.Tau = opts.Tau
	}
	if opts.SampleSize > 0 {
		cfg.SampleSize = opts.SampleSize
	}
	if opts.ActiveDays > 0 {
		cfg.ActiveDays = opts.ActiveDays
	}

	// Quiet mode for structured output (suppress progress to stderr)
	quiet := opts.Format == "json" || opts.Format == "csv"
	spinnerQuiet = quiet

	// Print alias info if configured
	if !quiet && len(cfg.Aliases) > 0 {
		fmt.Fprintf(os.Stderr, "Loaded %d author aliases from config\n", len(cfg.Aliases))
	}

	ctx := context.Background()
	start := time.Now()
	workers := opts.Workers
	if workers == 0 {
		workers = 4
	}

	// Initialize cache
	cacheStore := cache.New(!opts.NoCache)

	// Build extension-to-domain map from config + defaults
	extMap := domain.BuildExtMap(cfg.CustomExtensions(), cfg.UseDefaultDomains())

	// Per-domain accumulators
	accumulators := make(map[domain.Domain]*domainAccumulator)

	// Per-repo accumulators (only when --per-repo)
	type repoAccState struct {
		acc             *domainAccumulator
		repoName        string
		domain          domain.Domain
		qualityCounts   map[string]int
		debtCounts      map[string]int
		authorFirstDate map[string]time.Time
		authorLastDate  map[string]time.Time
	}
	var repoAccumulators []repoAccState

	// Deduplicate repos by resolving to real paths
	seen := make(map[string]bool)
	var dedupedPaths []string
	for _, p := range repoPaths {
		real, err := filepath.EvalSymlinks(p)
		if err != nil {
			real = p
		}
		if !seen[real] {
			seen[real] = true
			dedupedPaths = append(dedupedPaths, p)
		} else {
			fmt.Fprintf(os.Stderr, "SKIP: %s (duplicate of already queued repo)\n", filepath.Base(p))
		}
	}
	repoPaths = dedupedPaths

	totalAnalyzed := 0
	for _, repoPath := range repoPaths {
		// Verify it's a git repo
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "SKIP: %s (not a git repo)\n", repoPath)
			continue
		}

		repoName := filepath.Base(repoPath)

		// Skip excluded repos
		if cfg.IsExcludedRepo(repoName) {
			fmt.Fprintf(os.Stderr, "SKIP: %s (excluded in config)\n", repoName)
			continue
		}

		// Determine domain: config override first, then auto-detect
		repoDomain := resolveRepoDomain(ctx, repoPath, repoName, cfg, extMap)

		// Skip repos outside the requested domain
		if opts.DomainFilter != "" {
			filterDomain := domain.NormalizeName(opts.DomainFilter)
			if repoDomain != filterDomain {
				continue
			}
		}

		bold := color.New(color.Bold)
		domainLabel := color.New(color.FgCyan).Sprintf("[%s]", repoDomain)
		bold.Printf("Analyzing: %s %s\n", repoName, domainLabel)

		// Shallow clone warning
		if git.IsShallowRepo(ctx, repoPath) {
			warn := color.New(color.FgYellow)
			warn.Fprintf(os.Stderr, "  ⚠ WARNING: shallow clone detected — git blame may hang or produce inaccurate results\n")
			warn.Fprintf(os.Stderr, "    Run: git fetch --unshallow\n")
		}

		// Get or create accumulator for this domain
		acc, ok := accumulators[repoDomain]
		if !ok {
			acc = newDomainAccumulator()
			accumulators[repoDomain] = acc
		}
		acc.repoCount++
		totalAnalyzed++

		// Get HEAD hash for cache keys
		headHash, _ := git.HeadHash(ctx, repoPath)

		// Step 1: Parse git log (feeds Production, Quality, Design)
		spin := spinner("[1/4] Parsing git log...")
		var commits []git.Commit
		logCacheKey := cache.LogKey(repoPath, headHash)
		if headHash != "" && cacheStore.Get(logCacheKey, &commits) {
			spin.Stop()
			if !quiet {
				fmt.Fprintf(os.Stderr, "  (cached)\n")
			}
		} else {
			commits, err = git.ParseLog(ctx, repoPath)
			spin.Stop()
			if err != nil {
				return nil, nil, fmt.Errorf("parse log %s: %w", repoName, err)
			}
			if headHash != "" {
				cacheStore.Set(logCacheKey, commits)
			}
		}

		// Apply author aliases, filter excluded authors, and strip excluded file patterns
		commits = filterCommits(commits, cfg)
		commits = filterFileStats(commits, cfg.ExcludeFilePatterns)

		// Detect and exclude reverted commits (both originals and revert commits).
		// This ensures that code merged then reverted doesn't inflate metrics.
		revertedHashes, _ := git.FindRevertedCommits(ctx, repoPath)
		if len(revertedHashes) > 0 {
			before := len(commits)
			commits = filterRevertedCommits(commits, revertedHashes)
			excluded := before - len(commits)
			if excluded > 0 && !quiet {
				fmt.Fprintf(os.Stderr, "  Excluded %d reverted commits\n", excluded)
			}
		}

		// Also fetch merge commits for fix detection in Quality
		var mergeCommits []git.Commit
		mergeCacheKey := cache.MergeLogKey(repoPath, headHash)
		if headHash != "" && cacheStore.Get(mergeCacheKey, &mergeCommits) {
			// cached
		} else {
			mergeCommits, _ = git.ParseMergeCommits(ctx, repoPath)
			if headHash != "" {
				cacheStore.Set(mergeCacheKey, mergeCommits)
			}
		}
		mergeCommits = filterCommits(mergeCommits, cfg)
		// Also exclude reverted merge commits from Quality calculation
		if len(revertedHashes) > 0 {
			mergeCommits = filterRevertedCommits(mergeCommits, revertedHashes)
		}

		// Production (non-merge only)
		prod := metric.CalcProduction(commits, cfg.ExcludeFilePatterns)
		mergeMap(acc.raw.Production, prod)

		// Lines added/deleted
		added, deleted := metric.CalcLines(commits, cfg.ExcludeFilePatterns)
		mergeMapInt(acc.raw.LinesAdded, added)
		mergeMapInt(acc.raw.LinesDeleted, deleted)

		// Quality: include merge commits so fix subjects in merge messages are counted
		// Use make+copy+append to avoid slice backing array corruption
		allCommits := make([]git.Commit, len(commits), len(commits)+len(mergeCommits))
		copy(allCommits, commits)
		allCommits = append(allCommits, mergeCommits...)
		qual := metric.CalcQuality(allCommits)
		mergeMapAvg(acc.raw.Quality, qual, acc.qualityCounts)

		// Design (non-merge only — uses numstat)
		design := metric.CalcDesign(commits, cfg.ArchitecturePatterns)
		mergeMap(acc.raw.Design, design)

		// Track breadth with commit counts per repo, and date ranges for production rate
		for _, c := range commits {
			if _, ok := acc.authorRepoCommits[c.Author]; !ok {
				acc.authorRepoCommits[c.Author] = make(map[string]int)
			}
			acc.authorRepoCommits[c.Author][repoName]++
			acc.raw.TotalCommits[c.Author]++

			if first, ok := acc.authorFirstDate[c.Author]; !ok || c.Date.Before(first) {
				acc.authorFirstDate[c.Author] = c.Date
			}
			if last, ok := acc.authorLastDate[c.Author]; !ok || c.Date.After(last) {
				acc.authorLastDate[c.Author] = c.Date
			}
		}
		// Also track activity dates from merge commits
		for _, c := range mergeCommits {
			if first, ok := acc.authorFirstDate[c.Author]; !ok || c.Date.Before(first) {
				acc.authorFirstDate[c.Author] = c.Date
			}
			if last, ok := acc.authorLastDate[c.Author]; !ok || c.Date.After(last) {
				acc.authorLastDate[c.Author] = c.Date
			}
		}

		// Module Science: Co-change Coupling (uses commit data, no extra git calls)
		cochange := metric.CalcCochange(commits)
		acc.cochangeResults = append(acc.cochangeResults, cochange)

		// Step 2: Blame analysis (feeds Survival, Indispensability)
		spin = spinner("[2/4] Blame analysis...")
		files, err := git.ListFiles(ctx, repoPath, cfg.BlameExtensions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: could not list files: %v\n", err)
			continue
		}

		// Filter out excluded file patterns from blame targets
		files = filterFiles(files, cfg.ExcludeFilePatterns)

		var blameVerbose func(string)
		if opts.Verbose {
			blameVerbose = func(msg string) {
				fmt.Fprintf(os.Stderr, "\n%s", msg)
			}
		}
		spin.Clear()

		var blameLines []git.BlameLine
		blameCacheKey := cache.BlameKey(repoPath, headHash, files, cfg.SampleSize)
		if headHash != "" && cacheStore.Get(blameCacheKey, &blameLines) {
			if !quiet {
				fmt.Fprintf(os.Stderr, "  %s [2/4] Blame (cached)\n", color.New(color.FgGreen).Sprint("✓"))
			}
		} else {
			blameProg := newLiveProgress("[2/4] Blame")
			blameLines, err = git.ConcurrentBlameFiles(ctx, repoPath, files, cfg.SampleSize, workers,
				func(done, total int) {
					blameProg.Update(done, total)
				}, blameVerbose)
			blameProg.Stop()
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: blame error: %v\n", err)
			}
			if headHash != "" && len(blameLines) > 0 {
				cacheStore.Set(blameCacheKey, blameLines)
			}
		}

		// Apply aliases to blame lines
		for i := range blameLines {
			blameLines[i].Author = cfg.ResolveAuthor(blameLines[i].Author)
		}
		blameLines = filterBlameLines(blameLines, cfg)

		// Build the test-coverage lookup for this repo. Uses the filtered blame
		// file list (already in-scope) as the manifest — test files and prod
		// files share extensions so the lookup is accurate for code files.
		testedSet := metric.BuildTestedSet(files)
		acc.totalFiles += testedSet.TotalFiles
		acc.totalTestFiles += testedSet.TotalTestFiles
		testedSet.ForEachModule(func(mod string, total, test int) {
			acc.moduleAllFiles[mod] += total
			acc.moduleTestFiles[mod] += test
		})

		// Survival: split by change pressure or use classic mode
		// Keep per-repo survival maps for --per-repo reuse
		var repoSurvDecayed, repoSurvRaw, repoSurvRobust, repoSurvDormant map[string]float64
		var repoSurvTested, repoSurvUntested map[string]float64
		if opts.PressureMode == "include" {
			repoPressure := metric.CalcChangePressure(commits, blameLines)
			for mod, p := range repoPressure {
				key := repoName + "/" + mod
				acc.changePressure[key] = p
				// Module Science Phase 2: accumulate pressure without repo prefix
				n := acc.modulePressureCounts[mod]
				if n > 0 {
					acc.modulePressure[mod] = (acc.modulePressure[mod]*float64(n) + p) / float64(n+1)
				} else {
					acc.modulePressure[mod] = p
				}
				acc.modulePressureCounts[mod] = n + 1
			}

			// Need at least 2 authors with ≥10% AND ≥1000 blame lines for
			// pressure split to be meaningful. This filters out solo-dominated
			// repos while allowing smaller multi-person repos to have pressure.
			blameByAuthor := make(map[string]int)
			for _, bl := range blameLines {
				blameByAuthor[cfg.ResolveAuthor(bl.Author)]++
			}
			minShare := float64(len(blameLines)) * 0.10
			substantialAuthors := 0
			for _, count := range blameByAuthor {
				if float64(count) >= minShare && count >= 1000 {
					substantialAuthors++
				}
			}
			pressureThreshold := repoPressure.MedianPressure()
			if substantialAuthors < 2 {
				pressureThreshold = math.Inf(1) // everything becomes dormant
			}
			survResult := metric.CalcSurvivalFull(blameLines, cfg.Tau, start, repoPressure, pressureThreshold, testedSet, cfg.UntestedSurvivalWeight)
			repoSurvDecayed = survResult.Decayed
			repoSurvRaw = survResult.Raw
			repoSurvRobust = survResult.Robust
			repoSurvDormant = survResult.Dormant
			repoSurvTested = survResult.Tested
			repoSurvUntested = survResult.Untested
			mergeMap(acc.raw.Survival, repoSurvDecayed)
			mergeMap(acc.raw.RawSurvival, repoSurvRaw)
			mergeMap(acc.raw.RobustSurvival, repoSurvRobust)
			mergeMap(acc.raw.DormantSurvival, repoSurvDormant)
			mergeMap(acc.raw.TestedSurvival, repoSurvTested)
			mergeMap(acc.raw.UntestedSurvival, repoSurvUntested)
		} else {
			// Classic mode: no pressure split, but still apply the tested-weighting
			// so comment-era repos still benefit from gaming resistance.
			survResult := metric.CalcSurvivalFull(blameLines, cfg.Tau, start, nil, 0, testedSet, cfg.UntestedSurvivalWeight)
			repoSurvDecayed = survResult.Decayed
			repoSurvRaw = survResult.Raw
			repoSurvTested = survResult.Tested
			repoSurvUntested = survResult.Untested
			mergeMap(acc.raw.Survival, repoSurvDecayed)
			mergeMap(acc.raw.RawSurvival, repoSurvRaw)
			mergeMap(acc.raw.TestedSurvival, repoSurvTested)
			mergeMap(acc.raw.UntestedSurvival, repoSurvUntested)
		}

		// Indispensability
		indisp, risks := metric.CalcIndispensability(blameLines, cfg.BusFactor.Critical, cfg.BusFactor.High)
		mergeMap(acc.raw.Indispensability, indisp)

		// Step 3: Debt cleanup
		fixCommits := metric.GetFixCommits(commits)
		spin = spinner(fmt.Sprintf("[3/4] Debt analysis (%d fix commits)...", len(fixCommits)))
		var debtVerbose metric.VerboseFunc
		if opts.Verbose {
			debtVerbose = func(msg string) {
				fmt.Fprintf(os.Stderr, "\n%s", msg)
			}
		}
		spin.Clear()

		var debt map[string]float64
		var fixHashes []string
		for _, fc := range fixCommits {
			fixHashes = append(fixHashes, fc.Hash)
		}
		debtCacheKey := cache.DebtKey(repoPath, fixHashes)
		if headHash != "" && cacheStore.Get(debtCacheKey, &debt) {
			if !quiet {
				fmt.Fprintf(os.Stderr, "  %s [3/4] Debt (cached)\n", color.New(color.FgGreen).Sprint("✓"))
			}
		} else {
			debtProg := newLiveProgress("[3/4] Debt")
			debt, _ = metric.CalcDebt(ctx, repoPath, fixCommits, 50, cfg.DebtThreshold, cfg.BlameTimeout, cfg.ResolveAuthor,
				func(done, total int) {
					debtProg.Update(done, total)
				}, debtVerbose)
			debtProg.Stop()
			if headHash != "" && len(debt) > 0 {
				cacheStore.Set(debtCacheKey, debt)
			}
		}
		mergeMapAvg(acc.raw.DebtCleanup, debt, acc.debtCounts)

		// Module Science: Ownership Fragmentation (uses blame data)
		ownership := metric.CalcOwnershipFragmentation(blameLines)
		acc.ownership = append(acc.ownership, ownership...)

		// Module Science Phase 2: Per-module survival rate
		repoModSurv := metric.CalcModuleSurvival(blameLines, cfg.Tau, start)
		for mod, surv := range repoModSurv {
			if existing, ok := acc.moduleSurvival[mod]; ok {
				acc.moduleSurvival[mod] = (existing + surv) / 2
			} else {
				acc.moduleSurvival[mod] = surv
			}
		}

		// Step 4: Accumulate bus factor risks per domain; print immediately for table format
		acc.risks = append(acc.risks, risks...)
		if opts.Format == "table" && len(risks) > 0 {
			output.PrintBusFactorRisks(risks)
		}

		// Print module science results inline for table format
		if opts.Format == "table" {
			output.PrintCochangeCoupling(repoName, cochange)
			output.PrintOwnershipFragmentation(repoName, ownership)
		}

		// Per-repo: build independent raw scores for this repo
		if opts.PerRepo {
			repoRaw := metric.NewRawScores()
			mergeMap(repoRaw.Production, prod)
			mergeMap(repoRaw.Quality, qual)
			mergeMap(repoRaw.Design, design)
			mergeMap(repoRaw.Indispensability, indisp)
			mergeMap(repoRaw.DebtCleanup, debt)
			// Reuse already-computed survival data
			mergeMap(repoRaw.Survival, repoSurvDecayed)
			mergeMap(repoRaw.RawSurvival, repoSurvRaw)
			if repoSurvRobust != nil {
				mergeMap(repoRaw.RobustSurvival, repoSurvRobust)
			}
			if repoSurvDormant != nil {
				mergeMap(repoRaw.DormantSurvival, repoSurvDormant)
			}
			if repoSurvTested != nil {
				mergeMap(repoRaw.TestedSurvival, repoSurvTested)
			}
			if repoSurvUntested != nil {
				mergeMap(repoRaw.UntestedSurvival, repoSurvUntested)
			}
			// Track commit counts and dates per author for this repo
			repoFirstDate := make(map[string]time.Time)
			repoLastDate := make(map[string]time.Time)
			for _, c := range commits {
				repoRaw.TotalCommits[c.Author]++
				if first, ok := repoFirstDate[c.Author]; !ok || c.Date.Before(first) {
					repoFirstDate[c.Author] = c.Date
				}
				if last, ok := repoLastDate[c.Author]; !ok || c.Date.After(last) {
					repoLastDate[c.Author] = c.Date
				}
			}
			for _, c := range mergeCommits {
				if first, ok := repoFirstDate[c.Author]; !ok || c.Date.Before(first) {
					repoFirstDate[c.Author] = c.Date
				}
				if last, ok := repoLastDate[c.Author]; !ok || c.Date.After(last) {
					repoLastDate[c.Author] = c.Date
				}
			}
			repoAccumulators = append(repoAccumulators, repoAccState{
				acc:             &domainAccumulator{raw: repoRaw},
				repoName:        repoName,
				domain:          repoDomain,
				authorFirstDate: repoFirstDate,
				authorLastDate:  repoLastDate,
			})
		}
	}

	// Score per domain (stable order: built-in first, then custom alphabetically, Unknown last)
	var domainKeys []domain.Domain
	for d := range accumulators {
		domainKeys = append(domainKeys, d)
	}
	domains := domain.SortDomains(domainKeys)

	var results []DomainResults

	for _, d := range domains {
		acc, ok := accumulators[d]
		if !ok {
			continue
		}

		// Breadth: count repos where author has >= 3 commits
		const minCommitsForBreadth = 3
		for author, repos := range acc.authorRepoCommits {
			count := 0
			for _, commits := range repos {
				if commits >= minCommitsForBreadth {
					count++
				}
			}
			if count > 0 {
				acc.raw.Breadth[author] = float64(count)
			}
		}

		// Convert production total to per-day rate for absolute scoring
		for author, total := range acc.raw.Production {
			first := acc.authorFirstDate[author]
			last := acc.authorLastDate[author]
			days := last.Sub(first).Hours() / 24
			if days < 1 {
				days = 1
			}
			acc.raw.Production[author] = total / days
		}

		// Score and rank
		scored := scorer.Score(acc.raw, cfg, acc.authorLastDate)

		// Filter out excluded authors and ghost entries (0 commits, 0 total)
		var filtered []scorer.Result
		for _, r := range scored {
			if cfg.IsExcludedAuthor(r.Author) {
				continue
			}
			if r.TotalCommits == 0 && r.Impact == 0 {
				continue
			}
			filtered = append(filtered, r)
		}

		if len(filtered) == 0 {
			continue
		}

		// Module Science Phase 2: Score and classify modules
		// Aggregate per-module test ratio across repos (weighted by module size).
		moduleTestRatio := make(map[string]float64)
		for mod, total := range acc.moduleAllFiles {
			if total == 0 {
				continue
			}
			moduleTestRatio[mod] = float64(acc.moduleTestFiles[mod]) / float64(total)
		}

		moduleScores := scorer.ScoreModules(
			acc.modulePressure,
			acc.cochangeResults,
			acc.ownership,
			acc.moduleSurvival,
			acc.authorLastDate,
			cfg.ActiveDays,
			moduleTestRatio,
		)

		dr := DomainResults{
			Domain:         d,
			Results:        filtered,
			Risks:          acc.risks,
			RepoCount:      acc.repoCount,
			Cochange:       acc.cochangeResults,
			Ownership:      acc.ownership,
			ModuleScores:   moduleScores,
			TotalFiles:     acc.totalFiles,
			TotalTestFiles: acc.totalTestFiles,
		}
		if acc.totalFiles > 0 {
			dr.TestFileRatio = float64(acc.totalTestFiles) / float64(acc.totalFiles)
		}

		// Score per-repo results for this domain
		if opts.PerRepo {
			for _, ra := range repoAccumulators {
				if ra.domain != d {
					continue
				}
				// Convert production to per-day rate
				for author, total := range ra.acc.raw.Production {
					first := ra.authorFirstDate[author]
					last := ra.authorLastDate[author]
					days := last.Sub(first).Hours() / 24
					if days < 1 {
						days = 1
					}
					ra.acc.raw.Production[author] = total / days
				}
				// Breadth is 1 for single repo
				for author := range ra.acc.raw.TotalCommits {
					ra.acc.raw.Breadth[author] = 1
				}
				scored := scorer.Score(ra.acc.raw, cfg, ra.authorLastDate)
				var repoFiltered []scorer.Result
				for _, r := range scored {
					if !cfg.IsExcludedAuthor(r.Author) {
						repoFiltered = append(repoFiltered, r)
					}
				}
				if len(repoFiltered) > 0 {
					dr.PerRepo = append(dr.PerRepo, RepoResult{
						RepoName: ra.repoName,
						Domain:   ra.domain,
						Results:  repoFiltered,
					})
				}
			}
		}

		results = append(results, dr)
	}

	if opts.Format == "table" {
		elapsed := time.Since(start)
		color.New(color.FgHiBlack).Printf("Completed in %s (%d repos total)\n", elapsed.Round(time.Second), totalAnalyzed)
	}

	return results, cfg, nil
}

// resolveRepoDomain determines the domain for a repo.
// Config repo patterns take priority (checked in sorted key order for determinism),
// then auto-detection from file extensions.
func resolveRepoDomain(ctx context.Context, repoPath, repoName string, cfg *config.Config, extMap map[string]domain.Domain) domain.Domain {
	// Check config repo pattern overrides first (sorted for deterministic results)
	names := make([]string, 0, len(cfg.Domains))
	for name := range cfg.Domains {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		entry := cfg.Domains[name]
		if len(entry.Repos) > 0 && domain.MatchRepoPattern(repoName, entry.Repos) {
			return domain.NormalizeName(name)
		}
	}

	// Auto-detect from file extensions
	files, err := git.ListAllFiles(ctx, repoPath)
	if err != nil || len(files) == 0 {
		return domain.Unknown
	}

	return domain.DetectFromFiles(files, extMap)
}

// findGitRepos walks a directory tree up to maxDepth and returns paths containing .git
func findGitRepos(root string, maxDepth int) ([]string, error) {
	var repos []string

	rootDepth := len(splitPath(root))

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}

		if !info.IsDir() {
			return nil
		}

		depth := len(splitPath(path)) - rootDepth
		if depth > maxDepth {
			return filepath.SkipDir
		}

		// Check if this directory is a git repo (not a submodule — submodules have .git as a file, not a dir)
		gitDir := filepath.Join(path, ".git")
		if fi, err := os.Stat(gitDir); err == nil {
			if fi.IsDir() {
				repos = append(repos, path)
			}
			// Whether .git is a dir or file (submodule), don't descend further
			return filepath.SkipDir
		}

		return nil
	})

	return repos, err
}

func splitPath(p string) []string {
	return strings.Split(filepath.ToSlash(p), "/")
}

func filterCommits(commits []git.Commit, cfg *config.Config) []git.Commit {
	var result []git.Commit
	for _, c := range commits {
		c.Author = cfg.ResolveAuthor(c.Author)
		if cfg.IsExcludedAuthor(c.Author) {
			continue
		}
		result = append(result, c)
	}
	return result
}

// filterFileStats removes excluded file patterns from commit FileStats
func filterFileStats(commits []git.Commit, excludePatterns []string) []git.Commit {
	if len(excludePatterns) == 0 {
		return commits
	}
	for i := range commits {
		var filtered []git.FileStat
		for _, fs := range commits[i].FileStats {
			if !metric.IsExcluded(fs.Filename, excludePatterns) {
				filtered = append(filtered, fs)
			}
		}
		commits[i].FileStats = filtered
	}
	return commits
}

// filterFiles removes excluded file patterns from a file list
func filterFiles(files []string, excludePatterns []string) []string {
	if len(excludePatterns) == 0 {
		return files
	}
	var result []string
	for _, f := range files {
		if !metric.IsExcluded(f, excludePatterns) {
			result = append(result, f)
		}
	}
	return result
}

// filterRevertedCommits removes commits whose hashes are in the reverted set.
func filterRevertedCommits(commits []git.Commit, reverted map[string]bool) []git.Commit {
	if len(reverted) == 0 {
		return commits
	}
	var result []git.Commit
	for _, c := range commits {
		if !reverted[c.Hash] {
			result = append(result, c)
		}
	}
	return result
}

func filterBlameLines(lines []git.BlameLine, cfg *config.Config) []git.BlameLine {
	var result []git.BlameLine
	for _, bl := range lines {
		if !cfg.IsExcludedAuthor(bl.Author) {
			result = append(result, bl)
		}
	}
	return result
}

func mergeMap(dst, src map[string]float64) {
	for k, v := range src {
		dst[k] += v
	}
}

func mergeMapInt(dst, src map[string]int) {
	for k, v := range src {
		dst[k] += v
	}
}

// separateArgs splits CLI args into flags (--foo, --foo=bar, --foo bar) and
// positional paths. This allows flags to appear after positional arguments,
// which Go's flag package does not support by default.
func separateArgs(args []string, fs *flag.FlagSet) (flags []string, paths []string) {
	knownFlags := make(map[string]bool)
	fs.VisitAll(func(f *flag.Flag) {
		knownFlags[f.Name] = true
	})

	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			flags = append(flags, a)
			// Check if this flag takes a value (not a bool flag)
			name := strings.TrimLeft(a, "-")
			if idx := strings.Index(name, "="); idx >= 0 {
				// --flag=value — already included
				continue
			}
			// Look up if this is a bool flag
			if f := fs.Lookup(name); f != nil {
				if _, ok := f.Value.(interface{ IsBoolFlag() bool }); !ok {
					// Non-bool flag: next arg is the value
					if i+1 < len(args) {
						i++
						flags = append(flags, args[i])
					}
				}
			}
		} else {
			paths = append(paths, a)
		}
	}
	return
}

// spinResult holds the stop functions for a spinner.
type spinResult struct {
	// Stop stops the spinner and prints a ✓ completion line.
	Stop func()
	// Clear stops the spinner and clears the line (for transitioning to progress bar).
	Clear func()
}

// spinner runs an animated spinner on stderr until Stop or Clear is called.
// In quiet mode, both functions are no-ops.
var spinnerQuiet bool

func spinner(label string) spinResult {
	if spinnerQuiet {
		noop := func() {}
		return spinResult{Stop: noop, Clear: noop}
	}
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := make(chan struct{})
	stopped := false
	go func() {
		cyan := color.New(color.FgCyan)
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Fprintf(os.Stderr, "  %s %s\r", cyan.Sprint(frames[i%len(frames)]), label)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
	doStop := func() {
		if stopped {
			return
		}
		stopped = true
		close(done)
		time.Sleep(10 * time.Millisecond)
	}
	return spinResult{
		Stop: func() {
			doStop()
			fmt.Fprintf(os.Stderr, "  %s %s\n", color.New(color.FgGreen).Sprint("✓"), label)
		},
		Clear: func() {
			doStop()
			// Clear the spinner line with spaces and return carriage
			fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", len(label)+10))
		},
	}
}

// liveProgress manages a background-animated progress bar.
// The spinner keeps animating even when progress doesn't update.
type liveProgress struct {
	label string
	done  int
	total int
	mu    sync.Mutex
	quit  chan struct{}
}

var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func newLiveProgress(label string) *liveProgress {
	lp := &liveProgress{
		label: label,
		quit:  make(chan struct{}),
	}
	if !spinnerQuiet {
		go lp.run()
	}
	return lp
}

func (lp *liveProgress) Update(done, total int) {
	lp.mu.Lock()
	lp.done = done
	lp.total = total
	lp.mu.Unlock()
}

func (lp *liveProgress) run() {
	cyan := color.New(color.FgCyan)
	dim := color.New(color.FgHiBlack)
	green := color.New(color.FgGreen)
	i := 0
	for {
		select {
		case <-lp.quit:
			return
		default:
			lp.mu.Lock()
			done, total := lp.done, lp.total
			lp.mu.Unlock()

			const barWidth = 20
			var pct float64
			if total > 0 {
				pct = float64(done) / float64(total)
			}
			filled := int(pct * barWidth)
			if filled > barWidth {
				filled = barWidth
			}
			filledBar := cyan.Sprint(strings.Repeat("█", filled))
			emptyBar := dim.Sprint(strings.Repeat("░", barWidth-filled))
			count := green.Sprintf("%d/%d", done, total)
			frame := cyan.Sprint(spinFrames[i%len(spinFrames)])
			fmt.Fprintf(os.Stderr, "  %s %s [%s%s] %s\r", frame, lp.label, filledBar, emptyBar, count)
			i++
			time.Sleep(80 * time.Millisecond)
		}
	}
}

func (lp *liveProgress) Stop() {
	close(lp.quit)
	if spinnerQuiet {
		return
	}
	time.Sleep(10 * time.Millisecond)
	lp.mu.Lock()
	done, total := lp.done, lp.total
	lp.mu.Unlock()

	cyan := color.New(color.FgCyan)
	dim := color.New(color.FgHiBlack)
	green := color.New(color.FgGreen)
	const barWidth = 20
	var pct float64
	if total > 0 {
		pct = float64(done) / float64(total)
	}
	filled := int(pct * barWidth)
	if filled > barWidth {
		filled = barWidth
	}
	filledBar := cyan.Sprint(strings.Repeat("█", filled))
	emptyBar := dim.Sprint(strings.Repeat("░", barWidth-filled))
	count := green.Sprintf("%d/%d", done, total)
	fmt.Fprintf(os.Stderr, "  %s %s [%s%s] %s\n", green.Sprint("✓"), lp.label, filledBar, emptyBar, count)
}

// mergeMapAvg keeps a correct running average for quality scores across repos
func mergeMapAvg(dst, src map[string]float64, counts map[string]int) {
	for k, v := range src {
		n := counts[k]
		if n > 0 {
			// Cumulative average: (oldAvg * n + newValue) / (n + 1)
			dst[k] = (dst[k]*float64(n) + v) / float64(n+1)
		} else {
			dst[k] = v
		}
		counts[k] = n + 1
	}
}
