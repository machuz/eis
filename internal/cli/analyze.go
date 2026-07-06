package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

// defaultWorkers picks the concurrency for the log/blame/debt git fan-out when
// --workers is 0 (auto). The work is git-subprocess-bound, so it scales with
// cores; leave two for the main goroutine + git's own threads, floor at 4 (never
// slower than the old fixed default), cap at 16 (diminishing returns past there,
// and the blame pool caps there too).
func defaultWorkers() int {
	n := runtime.NumCPU() - 2
	if n < 4 {
		n = 4
	}
	if n > 16 {
		n = 16
	}
	return n
}

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
	PerRepo        bool
	FastLog        bool // skip `git log -p` comment filtering (numstat-only) for speed

	// AnalysisTime is the W-02 envelope clock: the single wall-clock reference
	// against which time-decay (survival/catalysis) and lifecycle classification
	// (State/RecentlyActive) are computed. Pin it to make a run reproducible —
	// the same (git_sha, AnalysisTime) MUST always produce the same scores.
	// Zero value means "use time.Now().UTC()" (the normal interactive default).
	AnalysisTime time.Time
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
	Cochange  []metric.CochangeResult  // per-repo co-change coupling (DSM)
	Ownership []metric.ModuleOwnership // accumulated ownership fragmentation

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
	raw                 *metric.RawScores
	debtCounts          map[string]int
	authorRepoCommits   map[string]map[string]int // author -> repo -> commit count
	authorModuleCommits map[string]map[string]int // author -> module -> commit count
	// authorModuleSurvival accumulates per-(module, author) surviving gravity
	// mass across repos — the input to the survival-weighted, structure-
	// neutral Breadth (Hill number). Keyed module -> author -> mass.
	authorModuleSurvival map[string]map[string]float64
	authorFirstDate      map[string]time.Time // earliest commit date per author
	authorLastDate       map[string]time.Time // latest commit date per author
	repoCount            int
	risks                []metric.ModuleRisk   // accumulated bus factor risks
	changePressure       metric.ChangePressure // accumulated change pressure across repos

	// Module Science Phase 1
	cochangeResults []metric.CochangeResult  // per-repo co-change coupling
	ownership       []metric.ModuleOwnership // accumulated ownership fragmentation

	// Module Science Phase 2
	moduleSurvival       map[string]float64    // per-module survival rate (0-1)
	modulePressure       metric.ChangePressure // per-module change pressure (without repo prefix)
	modulePressureCounts map[string]int        // count for averaging across repos

	// Test coverage observability (populated from each repo's TestedSet)
	totalFiles     int // sum of code files across all repos in this domain
	totalTestFiles int // sum of test files across all repos in this domain

	// Per-module file counts across all repos in this domain — used by
	// ScoreModules to compute Vitality=Fragile. Keyed on the module id
	// from the convention-aware metric.ModuleResolver.
	moduleAllFiles  map[string]int
	moduleTestFiles map[string]int
}

func newDomainAccumulator() *domainAccumulator {
	return &domainAccumulator{
		raw:                  metric.NewRawScores(),
		debtCounts:           make(map[string]int),
		authorRepoCommits:    make(map[string]map[string]int),
		authorModuleCommits:  make(map[string]map[string]int),
		authorModuleSurvival: make(map[string]map[string]float64),
		authorFirstDate:      make(map[string]time.Time),
		authorLastDate:       make(map[string]time.Time),
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
	workers := fs.Int("workers", 0, "Concurrent workers for log/blame/debt (0 = auto, based on CPU count)")
	recursive := fs.Bool("recursive", false, "Recursively find git repos under given paths")
	maxDepth := fs.Int("depth", 2, "Max directory depth for recursive search")
	formatFlag := fs.String("format", "table", "Output format: table, csv, json")
	pressureMode := fs.String("pressure-mode", "include", "Change pressure mode: include (split robust/dormant) or ignore (classic survival)")
	activeDays := fs.Int("active-days", 0, "Days to consider author active (overrides config, default 30)")
	domainFilter := fs.String("domain", "", "Only analyze repos in this domain (e.g. Backend, Frontend, Firmware)")
	verbose := fs.Bool("verbose", false, "Show detailed debug output (file-level timing)")
	noCache := fs.Bool("no-cache", false, "Skip disk cache")
	perRepo := fs.Bool("per-repo", false, "Show per-repository breakdown (requires --recursive)")
	fastLog := fs.Bool("fast-log", false, "Skip git log -p comment filtering (numstat-only) — much faster on large repos; insertion counts then include comment/blank lines (gravity is essentially unaffected)")
	upload := fs.Bool("upload", false, "Upload signals to the OrbitLens observatory (code stays local; only signals are sent)")
	token := fs.String("token", os.Getenv("EIS_TOKEN"), "Upload token (or set EIS_TOKEN); create one in Settings → API tokens")

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
		PerRepo:        *perRepo,
		FastLog:        *fastLog,
	}

	domainResults, cfg, analysisTime, err := RunAnalyzePipeline(opts, pathArgs)
	if err != nil {
		return err
	}

	if err := outputAnalyzeResults(domainResults, cfg, opts.Format, analysisTime); err != nil {
		return err
	}

	if *upload {
		// Resolve the analyzed repo from the first path arg (default ".") — the
		// HEAD sha there stamps the observation lineage. Upload status goes to
		// stderr so it never pollutes --format=json stdout. analysisTime is the
		// SAME envelope clock the scores were computed against — the upload
		// envelope records it verbatim (W-02), never a fresh time.Now().
		repoPath := "."
		if len(pathArgs) > 0 {
			repoPath = pathArgs[0]
		}
		if err := uploadResults(context.Background(), domainResults, repoPath, *token, analysisTime); err != nil {
			return fmt.Errorf("upload: %w", err)
		}
		fmt.Fprintln(os.Stderr, "✦ signals uploaded to the observatory")
	}

	return nil
}

func outputAnalyzeResults(domainResults []DomainResults, cfg *config.Config, format string, analysisTime time.Time) error {
	var jsonWriter *output.JSONWriter
	if format == "json" {
		// Stamp the W-02 envelope clock into the JSON envelope so a file-based
		// consumer can capture/replay the exact decay/classification reference
		// the scores were computed against (additive field; shape unchanged).
		jsonWriter = output.NewJSONWriter()
		jsonWriter.SetAnalysisTime(analysisTime)
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

// RunAnalyzePipeline runs the full analysis pipeline and returns per-domain
// results plus the W-02 envelope clock (analysisTime) the scores were actually
// computed against — so callers (upload, JSON output) can record/replay the
// exact decay/classification reference rather than re-stamping a fresh
// time.Now() that would not match the scores. This is the shared core used by
// both `eis analyze` and `eis team`.
func RunAnalyzePipeline(opts AnalyzeOptions, paths []string) ([]DomainResults, *config.Config, time.Time, error) {
	repoPaths := paths
	if len(repoPaths) == 0 {
		repoPaths = []string{"."}
	}

	// Resolve to absolute paths
	for i, p := range repoPaths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, nil, time.Time{}, fmt.Errorf("resolve path %s: %w", p, err)
		}
		repoPaths[i] = abs
	}

	// Recursive: find git repos under given paths
	if opts.Recursive {
		var discovered []string
		for _, root := range repoPaths {
			repos, err := findGitRepos(root, opts.MaxDepth)
			if err != nil {
				return nil, nil, time.Time{}, fmt.Errorf("scan %s: %w", root, err)
			}
			discovered = append(discovered, repos...)
		}
		if len(discovered) == 0 {
			return nil, nil, time.Time{}, fmt.Errorf("no git repos found under %v (depth=%d)", repoPaths, opts.MaxDepth)
		}
		repoPaths = discovered
		fmt.Fprintf(os.Stderr, "Found %d git repos\n\n", len(repoPaths))
	}

	// Load config
	cfg, err := config.Load(opts.ConfigPath, opts.ExplicitConfig)
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("load config: %w", err)
	}
	if opts.FastLog {
		off := false
		cfg.CommentFilter = &off
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

	// analysisTime is the W-02 envelope clock: the SINGLE decay/classification
	// reference threaded through every score-affecting computation below
	// (survival, catalysis, module survival, and scorer.ScoreAt). Pinning it
	// (opts.AnalysisTime) makes the run reproducible — same (git_sha,
	// analysisTime) ⇒ same scores. It is deliberately distinct from `start`,
	// which is ONLY a perf timer for duration logging and must never reach a
	// score. See pkg/timeline/run.go for the reference threading (window.End).
	analysisTime := opts.AnalysisTime
	if analysisTime.IsZero() {
		analysisTime = time.Now().UTC()
	}

	start := time.Now() // perf timer only — NOT a decay/classification reference
	workers := opts.Workers
	if workers == 0 {
		workers = defaultWorkers()
	}

	// Module resolvers are built per repo (see the loop below) so each
	// repo can honor its own RepoOverrides without leaking pattern sets
	// across repos. Resolution remains deterministic per repo (W-02/W-03)
	// because it depends only on the resolved pattern list.

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
		bold.Fprintf(os.Stderr, "Analyzing: %s %s\n", repoName, domainLabel)

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

		// Per-repo module resolver: honors RepoOverrides for this specific
		// repo (lookup key is the repo's base name, matching the CLI's
		// repoName-only identifier convention). Falls back to org-level
		// ModulePatterns and then to DefaultModulePatterns inside
		// PatternsForRepo.
		moduleResolver := metric.NewModuleResolverWithExcludes(config.PatternsForRepo(cfg, repoName), config.ExcludesForRepo(cfg, repoName))

		// Get HEAD hash for cache keys
		headHash, _ := git.HeadHash(ctx, repoPath)

		// Step 1: build the commit aggregate (feeds Production, Catalysis, Design,
		// Breadth, co-change, change pressure). Both strategies fold every filtered
		// commit into the SAME metric.CommitAggregator, so their output is identical
		// (verified byte-for-byte); they differ only in whether the history is held:
		//   - materialized (default): one `git log -p` walk into a []Commit that
		//     feeds the aggregator. Fast; peak RSS is dominated by the blame stage
		//     anyway, so holding the slice is fine up to very large repos.
		//   - streaming (only past streamingCommitThreshold): three lighter passes
		//     that fold each commit and discard it, never materializing []Commit —
		//     lower LIVE heap on Linux-scale histories (1M+ commits) where the slice
		//     genuinely dominates. Empirically this does NOT lower peak RSS on repos
		//     up to ~200k commits (blame + Go arenas dominate), so it is gated behind
		//     the threshold to avoid the extra passes' latency on ordinary repos.
		git.ConfigureBotCoAuthorPatterns(cfg.BotCoAuthorPatterns)

		// Effective exclusions: config patterns + .gitattributes linguist-generated
		// /vendored paths (so generated/vendored files don't inflate gravity). Used
		// by every file filter below, so blame + change-volume agree.
		excludes := effectiveExcludes(ctx, repoPath, cfg)
		// Reverted commits (originals + reverts) are dropped so code merged then
		// reverted doesn't inflate metrics. HEAD-derived set; retains no history.
		revertedHashes, _ := git.FindRevertedCommits(ctx, repoPath)

		spin := spinner("[1/4] Parsing git log...")
		commitCount, _ := git.CountCommits(ctx, repoPath)
		var res *aggResult
		if commitCount >= streamingCommitThreshold {
			res, err = aggregateStreaming(ctx, repoPath, workers, cfg, moduleResolver, excludes, revertedHashes)
		} else {
			res, err = aggregateMaterialized(ctx, repoPath, headHash, workers, cfg, moduleResolver, excludes, revertedHashes, cacheStore)
		}
		spin.Stop()
		if err != nil {
			return nil, nil, time.Time{}, fmt.Errorf("parse log %s: %w", repoName, err)
		}
		ag := res.ag
		idmap := res.idmap
		moduleResolver = res.resolver
		if res.revertedDropped > 0 && !quiet {
			fmt.Fprintf(os.Stderr, "  Excluded %d reverted commits\n", res.revertedDropped)
		}

		// Also fetch merge commits for fix detection in Catalysis
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
		// Also exclude reverted merge commits from Catalysis calculation
		if len(revertedHashes) > 0 {
			mergeCommits = filterRevertedCommits(mergeCommits, revertedHashes)
		}

		// Production + Lines, from the aggregator (streaming equivalents of
		// CalcProduction / CalcLines, folded in history order so the float sum is
		// bit-identical to the serial walk).
		mergeMap(acc.raw.Production, ag.Production)
		mergeMapInt(acc.raw.LinesAdded, ag.LinesAdded)
		mergeMapInt(acc.raw.LinesDeleted, ag.LinesDeleted)

		// Catalysis is computed in the blame stage below (it needs both the
		// commit history — to find each file's originator — and the surviving
		// blame lines — to measure how much of others' work on those files lasts).

		// Design is computed in the blame stage below (survival-weighted: surviving
		// arch-file blame lines, not arch lines changed — needs blame + analysisTime).

		// Breadth inputs (per-repo + per-module commit counts) and activity dates,
		// from the aggregator (the streaming equivalent of the old per-commit loop).
		for author, n := range ag.TotalCommits {
			if _, ok := acc.authorRepoCommits[author]; !ok {
				acc.authorRepoCommits[author] = make(map[string]int)
			}
			acc.authorRepoCommits[author][repoName] += n
			acc.raw.TotalCommits[author] += n
		}
		for author, mods := range ag.AuthorModuleCommits {
			dst := acc.authorModuleCommits[author]
			if dst == nil {
				dst = make(map[string]int)
				acc.authorModuleCommits[author] = dst
			}
			for mod, n := range mods {
				dst[mod] += n
			}
		}
		for author, d := range ag.FirstDate {
			if first, ok := acc.authorFirstDate[author]; !ok || d.Before(first) {
				acc.authorFirstDate[author] = d
			}
		}
		for author, d := range ag.LastDate {
			if last, ok := acc.authorLastDate[author]; !ok || d.After(last) {
				acc.authorLastDate[author] = d
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

		// Module Science: Co-change Coupling (folded by the aggregator during pass 2)
		cochange := ag.Cochange
		acc.cochangeResults = append(acc.cochangeResults, cochange)

		// Step 2: Blame analysis (feeds Survival, Indispensability)
		spin = spinner("[2/4] Blame analysis...")
		files, err := git.ListFiles(ctx, repoPath, cfg.BlameExtensions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: could not list files: %v\n", err)
			continue
		}

		// Filter out excluded file patterns from blame targets
		files = filterFiles(files, excludes)

		var blameVerbose func(string)
		if opts.Verbose {
			blameVerbose = func(msg string) {
				fmt.Fprintf(os.Stderr, "\n%s", msg)
			}
		}
		spin.Clear()

		var blameLines []git.BlameLine
		// Apply the run's blame move/copy detection policy (read by every blame
		// call) and fold it into the cache key so a level change recomputes.
		git.ConfigureBlameMoveDetection(cfg.BlameMoveDetection)
		blameCacheKey := cache.BlameKey(repoPath, headHash, files, cfg.SampleSize, cfg.BlameMoveDetection)
		if headHash != "" && cacheStore.Get(blameCacheKey, &blameLines) {
			if !quiet {
				fmt.Fprintf(os.Stderr, "  %s [2/4] Blame (cached)\n", color.New(color.FgGreen).Sprint("✓"))
			}
		} else {
			// Pre-skip files larger than cfg.MaxBlameFileBytes so a single
			// pathological dump can't drag the whole blame phase down.
			files, err = git.FilterFilesBySize(ctx, repoPath, "HEAD", files, cfg.MaxBlameFileBytes, blameVerbose)
			if err != nil && blameVerbose != nil {
				blameVerbose(fmt.Sprintf("  [blame] size filter error: %v", err))
			}
			blameProg := newLiveProgress("[2/4] Blame")
			blameLines, err = git.ConcurrentBlameFiles(ctx, repoPath, files, cfg.SampleSize, workers, cfg.BlameTimeout,
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

		// Collapse split identities on blame authors (by name, via the commit
		// email map), then apply config aliases — same order as the commit path.
		git.CanonicalizeAuthors(nil, blameLines, idmap)
		// Attach each line's co-author set (by commit SHA) so survival is split
		// across squash/pair co-authors. Built from the canonicalized commits, so
		// the sets carry the same author keys as the blame primaries.
		coMap := ag.CoMap
		for i := range blameLines {
			blameLines[i].Author = cfg.ResolveAuthor(blameLines[i].Author)
			if set, ok := coMap[blameLines[i].Commit]; ok {
				blameLines[i].Authors = set
			}
		}
		// Drop lines a `git subtree --squash` import collapsed onto the
		// integrator (blame is at HEAD, so scan squashes reachable from HEAD).
		// Unlike co-authored lines these have no recoverable author, so they
		// credit no one rather than the squasher.
		if !cfg.AttributeSubtreeSquash {
			if squash, sqErr := git.SubtreeSquashCommits(ctx, repoPath, "HEAD"); sqErr == nil {
				if kept, dropped := git.DropSubtreeSquashBlame(blameLines, squash); dropped > 0 {
					blameLines = kept
					if blameVerbose != nil {
						blameVerbose(fmt.Sprintf("  [blame] subtree-squash: suppressed %d imported line(s) from %d import commit(s)", dropped, len(squash)))
					}
				}
			}
		}
		blameLines = filterBlameLines(blameLines, cfg)

		// Build the test-coverage lookup for this repo. Uses the filtered blame
		// file list (already in-scope) as the manifest — test files and prod
		// files share extensions so the lookup is accurate for code files.
		testedSet := metric.BuildTestedSet(files, moduleResolver)
		acc.totalFiles += testedSet.TotalFiles
		acc.totalTestFiles += testedSet.TotalTestFiles
		testedSet.ForEachModule(func(mod string, total, test int) {
			acc.moduleAllFiles[mod] += total
			acc.moduleTestFiles[mod] += test
		})

		// Design (survival-weighted): surviving, time-decayed arch-file blame lines
		// per author, on the same analysisTime + tau as survival. Arch churn that was
		// itself rewritten has decayed out of the blame, so design measures durable
		// structural ownership rather than raw arch-edit volume.
		designSurv := metric.CalcDesignSurviving(blameLines, cfg.ArchitecturePatterns, cfg.Tau, analysisTime)
		mergeMap(acc.raw.Design, designSurv)

		// Survival: split by change pressure or use classic mode
		// Keep per-repo survival maps for --per-repo reuse
		var repoSurvDecayed, repoSurvRaw, repoSurvRobust, repoSurvDormant map[string]float64
		var repoSurvTested, repoSurvUntested map[string]float64
		if opts.PressureMode == "include" {
			repoPressure := metric.CalcChangePressureFrom(ag.Cochange.ModuleCommits, blameLines, moduleResolver)
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

			// Need at least 2 substantial authors for the pressure split to be
			// meaningful; otherwise everything becomes dormant. See
			// metric.PressureThreshold for why this is an absolute footprint and
			// not a share of the repo.
			// blameLines authors are already alias-resolved in place above, so
			// count on bl.Author directly (matches the survival maps' keys).
			blameByAuthor := make(map[string]int)
			for _, bl := range blameLines {
				blameByAuthor[bl.Author]++
			}
			pressureThreshold := metric.PressureThreshold(repoPressure, blameByAuthor, metric.SubstantialAuthorLines)
			repoOthers := metric.CalcOthersPressureFrom(ag.Cochange.ModuleCommits, ag.ModuleAuthorCommits, blameLines, moduleResolver)
			survResult := metric.CalcSurvivalFull(blameLines, cfg.TauForDomain(string(repoDomain)), analysisTime, repoPressure, pressureThreshold, moduleResolver, testedSet, cfg.UntestedSurvivalWeight, repoOthers)
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
			survResult := metric.CalcSurvivalFull(blameLines, cfg.TauForDomain(string(repoDomain)), analysisTime, nil, 0, moduleResolver, testedSet, cfg.UntestedSurvivalWeight, nil)
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
		indisp, risks := metric.CalcIndispensability(blameLines, moduleResolver, cfg.BusFactor.Critical, cfg.BusFactor.High)
		mergeMap(acc.raw.Indispensability, indisp)

		// Catalysis: surviving mass of others' work on files this author
		// originated. Needs the commit history (originator per file) and the
		// blame lines (surviving mass). Same decay reference (analysisTime) as Survival.
		catalysis := metric.CalcCatalysisFrom(ag.FirstContrib, blameLines, cfg.TauForDomain(string(repoDomain)), analysisTime)
		mergeMap(acc.raw.Catalysis, catalysis)

		// Step 3: Debt cleanup
		fixCommits := ag.FixCommits
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
			debt, _ = metric.CalcDebt(ctx, repoPath, fixCommits, 50, cfg.DebtThreshold, cfg.BlameTimeout, workers, cfg.ResolveAuthor,
				ag.CoMap,
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
		ownership := metric.CalcOwnershipFragmentation(blameLines, moduleResolver)
		acc.ownership = append(acc.ownership, ownership...)

		// Module Science Phase 2: Per-module survival rate
		repoModSurv := metric.CalcModuleSurvival(blameLines, cfg.TauForDomain(string(repoDomain)), analysisTime, moduleResolver)
		for mod, surv := range repoModSurv {
			if existing, ok := acc.moduleSurvival[mod]; ok {
				acc.moduleSurvival[mod] = (existing + surv) / 2
			} else {
				acc.moduleSurvival[mod] = surv
			}
		}

		// Per-(module, author) surviving gravity for Breadth (Hill number).
		repoMSBA := metric.CalcModuleSurvivalByAuthor(blameLines, cfg.TauForDomain(string(repoDomain)), analysisTime, moduleResolver)
		for mod, authors := range repoMSBA {
			dst := acc.authorModuleSurvival[mod]
			if dst == nil {
				dst = make(map[string]float64)
				acc.authorModuleSurvival[mod] = dst
			}
			for author, mass := range authors {
				dst[author] += mass
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
			mergeMap(repoRaw.Production, ag.Production)
			mergeMap(repoRaw.Catalysis, catalysis)
			mergeMap(repoRaw.Design, designSurv)
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
			// Track commit counts and dates per author for this repo (from the
			// aggregator — this repo's non-merge stream).
			repoFirstDate := make(map[string]time.Time)
			repoLastDate := make(map[string]time.Time)
			for author, n := range ag.TotalCommits {
				repoRaw.TotalCommits[author] += n
			}
			for author, d := range ag.FirstDate {
				repoFirstDate[author] = d
			}
			for author, d := range ag.LastDate {
				repoLastDate[author] = d
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

		// Breadth = effective number of modules an author holds surviving
		// gravity in (Hill number over per-module gravity), structure-neutral
		// and survival-weighted — same shared helper as the analyzer and
		// timeline pipelines so the three can't drift (W-02).
		breadth := metric.ComputeBreadth(acc.authorModuleSurvival)
		for author, b := range breadth {
			acc.raw.Breadth[author] = b
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

		// Score and rank. ScoreAt (not Score) so State/RecentlyActive classify
		// against the same W-02 envelope clock the decay metrics used — never a
		// fresh time.Now().
		scored := scorer.ScoreAt(acc.raw, cfg, acc.authorLastDate, analysisTime)

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
				// Per-repo Breadth saturates at 1: a per-repo breakdown is
				// scoped to one repo, so repo-unit Breadth is always {0,1}.
				// Module-unit Breadth is a run-level decision reported on
				// the merged result, not re-derived per repo here.
				for author := range ra.acc.raw.TotalCommits {
					ra.acc.raw.Breadth[author] = 1
				}
				scored := scorer.ScoreAt(ra.acc.raw, cfg, ra.authorLastDate, analysisTime)
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

	return results, cfg, analysisTime, nil
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
		c.CoAuthors = resolveCoAuthors(c.CoAuthors, c.Author, cfg)
		result = append(result, c)
	}
	return result
}

// streamingCommitThreshold is the non-merge commit count at/above which the
// analyzer switches from the materialized log walk to the three-pass streaming
// ingest. It is set high on purpose: measurements show the streaming path lowers
// LIVE heap but NOT peak RSS on repos up to ~200k commits (the blame stage and Go
// arenas dominate there), while costing an extra ~10% wall time for its extra
// passes. Only Linux-scale histories (1M+ commits), where the []Commit slice
// genuinely dominates memory, benefit — so ordinary and even large repos keep the
// faster single-walk path. var (not const) so a test can lower it.
var streamingCommitThreshold = 400000

// aggResult bundles the commit-aggregation outputs the per-repo loop consumes,
// so the materialized and streaming strategies present an identical interface.
type aggResult struct {
	ag              metric.Aggregate
	idmap           map[string]string
	resolver        metric.ModuleResolver
	revertedDropped int
}

// aggregateStreaming builds the aggregate WITHOUT materializing []Commit, via
// three streaming passes (identity map, then the module-liveness fold, then the
// CommitAggregator). Passes 1 and 2 apply keepFilteredCommit per commit — the
// same transform the materialized path applies to the whole slice — so the
// result is identical. For giant (Linux-scale) histories only.
func aggregateStreaming(ctx context.Context, repoPath string, workers int, cfg *config.Config, resolver metric.ModuleResolver, excludes []string, reverted map[string]bool) (*aggResult, error) {
	// Pass 0: identity map over all raw commits (cheap format-only walk).
	idmap, err := git.StreamIdentityMap(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	// Pass 1: module-liveness fold over the filtered stream, then fold the
	// resolver (must precede every module-topology metric, W-02).
	foldAcc := metric.NewModuleFoldAccumulator(resolver, cfg.ModuleLivenessMinMonths)
	if err := git.StreamLogParallel(ctx, repoPath, workers, false, func(c *git.Commit) {
		if keepFilteredCommit(c, cfg, idmap, excludes) && !reverted[c.Hash] {
			foldAcc.Add(c)
		}
	}); err != nil {
		return nil, err
	}
	resolver = resolver.WithFold(foldAcc.Result())
	// Pass 2: fold every commit's contribution into the aggregator.
	agg := metric.NewCommitAggregator(resolver)
	dropped := 0
	if err := git.StreamLogParallel(ctx, repoPath, workers, cfg.CommentFilterEnabled(), func(c *git.Commit) {
		if !keepFilteredCommit(c, cfg, idmap, excludes) {
			return
		}
		if reverted[c.Hash] {
			dropped++
			return
		}
		agg.Fold(c)
	}); err != nil {
		return nil, err
	}
	return &aggResult{ag: agg.Finalize(), idmap: idmap, resolver: resolver, revertedDropped: dropped}, nil
}

// aggregateMaterialized builds the aggregate the classic way: one `git log -p`
// walk into a []Commit (cached under LogKey), filtered in place, then folded into
// the SAME CommitAggregator. Holding the slice is fine here because peak RSS is
// dominated by the blame stage, not the log, up to very large repos.
func aggregateMaterialized(ctx context.Context, repoPath, headHash string, workers int, cfg *config.Config, resolver metric.ModuleResolver, excludes []string, reverted map[string]bool, cacheStore *cache.Store) (*aggResult, error) {
	var commits []git.Commit
	logCacheKey := cache.LogKey(repoPath, headHash, cfg.CommentFilterEnabled())
	if headHash == "" || !cacheStore.Get(logCacheKey, &commits) {
		var err error
		commits, err = git.ParseLogParallel(ctx, repoPath, workers, cfg.CommentFilterEnabled())
		if err != nil {
			return nil, err
		}
		if headHash != "" {
			cacheStore.Set(logCacheKey, commits)
		}
	}
	// Collapse split identities, apply aliases, strip excluded files/authors and
	// reverted commits — the same up-front transform, then feed the aggregator.
	idmap := git.BuildIdentityMap(commits)
	git.CanonicalizeAuthors(commits, nil, idmap)
	commits = filterCommits(commits, cfg)
	commits = filterFileStats(commits, excludes)
	dropped := 0
	if len(reverted) > 0 {
		before := len(commits)
		commits = filterRevertedCommits(commits, reverted)
		dropped = before - len(commits)
	}
	fold := metric.ComputeModuleFold(commits, resolver, cfg.ModuleLivenessMinMonths)
	resolver = resolver.WithFold(fold)
	agg := metric.NewCommitAggregator(resolver)
	for i := range commits {
		agg.Fold(&commits[i])
	}
	return &aggResult{ag: agg.Finalize(), idmap: idmap, resolver: resolver, revertedDropped: dropped}, nil
}

// keepFilteredCommit applies, IN PLACE, the same per-commit transform the
// materialized pipeline applies to the whole slice up front —
// git.CanonicalizeAuthors (idmap) → filterCommits (alias + excluded-author drop)
// → filterFileStats (excluded paths) — and reports whether the commit survives.
// Reverted-commit dropping is intentionally NOT here: callers apply it after this
// (matching the materialized order filterCommits→filterFileStats→filterReverted),
// so a reverted commit can still be counted for the stderr "excluded N" note.
func keepFilteredCommit(c *git.Commit, cfg *config.Config, idmap map[string]string, excludes []string) bool {
	if cn, ok := idmap[c.Author]; ok {
		c.Author = cn
	}
	for j, ca := range c.CoAuthors {
		if cn, ok := idmap[ca]; ok {
			c.CoAuthors[j] = cn
		}
	}
	c.Author = cfg.ResolveAuthor(c.Author)
	if cfg.IsExcludedAuthor(c.Author) {
		return false
	}
	c.CoAuthors = resolveCoAuthors(c.CoAuthors, c.Author, cfg)
	if len(excludes) > 0 {
		kept := c.FileStats[:0]
		for _, fs := range c.FileStats {
			if !metric.IsExcluded(fs.Filename, excludes) {
				kept = append(kept, fs)
			}
		}
		c.FileStats = kept
	}
	return true
}

// resolveCoAuthors aliases each Co-authored-by name, drops excluded authors and
// any that coincide with the primary (so the primary is never double-counted).
func resolveCoAuthors(coAuthors []string, primary string, cfg *config.Config) []string {
	if len(coAuthors) == 0 {
		return nil
	}
	out := coAuthors[:0]
	for _, a := range coAuthors {
		a = cfg.ResolveAuthor(a)
		if a == "" || a == primary || cfg.IsExcludedAuthor(a) {
			continue
		}
		out = append(out, a)
	}
	if len(out) == 0 {
		return nil
	}
	return out
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

// effectiveExcludes returns the file-exclusion patterns for a repo: the config
// patterns plus, when respect_gitattributes is on, the linguist-generated /
// vendored paths (escaped to exact-match globs) so generated/vendored files don't
// inflate gravity. Best-effort — a git error falls back to the config patterns
// (real files are never dropped on an attribute-setup edge).
func effectiveExcludes(ctx context.Context, repoPath string, cfg *config.Config) []string {
	if !cfg.RespectGitattributesEnabled() {
		return cfg.ExcludeFilePatterns
	}
	gen, err := git.LinguistExcluded(ctx, repoPath)
	if err != nil || len(gen) == 0 {
		return cfg.ExcludeFilePatterns
	}
	out := make([]string, 0, len(cfg.ExcludeFilePatterns)+len(gen))
	out = append(out, cfg.ExcludeFilePatterns...)
	for _, p := range gen {
		out = append(out, metric.EscapeGlob(p))
	}
	return out
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
