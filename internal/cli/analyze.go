package cli

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/machuz/engineering-impact-score/internal/config"
	"github.com/machuz/engineering-impact-score/internal/domain"
	"github.com/machuz/engineering-impact-score/internal/git"
	"github.com/machuz/engineering-impact-score/internal/metric"
	"github.com/machuz/engineering-impact-score/internal/output"
	"github.com/machuz/engineering-impact-score/internal/scorer"
)

// AnalyzeOptions holds CLI flags for the analysis pipeline.
type AnalyzeOptions struct {
	ConfigPath   string
	Tau          float64
	SampleSize   int
	Workers      int
	Recursive    bool
	MaxDepth     int
	Format       string
	PressureMode string
	ActiveDays   int
	DomainFilter string
}

// DomainResults holds scored results for a single domain.
type DomainResults struct {
	Domain    domain.Domain
	Results   []scorer.Result
	Risks     []metric.ModuleRisk
	RepoCount int
}

// domainAccumulator holds per-domain scoring state
type domainAccumulator struct {
	raw               *metric.RawScores
	qualityCounts     map[string]int
	debtCounts        map[string]int
	authorRepoCommits map[string]map[string]int // author -> repo -> commit count
	authorFirstDate   map[string]time.Time      // earliest commit date per author
	authorLastDate    map[string]time.Time       // latest commit date per author
	repoCount         int
	risks             []metric.ModuleRisk        // accumulated bus factor risks
	changePressure    metric.ChangePressure      // accumulated change pressure across repos
}

func newDomainAccumulator() *domainAccumulator {
	return &domainAccumulator{
		raw:               metric.NewRawScores(),
		qualityCounts:     make(map[string]int),
		debtCounts:        make(map[string]int),
		authorRepoCommits: make(map[string]map[string]int),
		authorFirstDate:   make(map[string]time.Time),
		authorLastDate:    make(map[string]time.Time),
		changePressure:    make(metric.ChangePressure),
	}
}

func runAnalyze(args []string) error {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	configPath := fs.String("config", "eis.yaml", "Config file path")
	tau := fs.Float64("tau", 0, "Survival decay parameter (overrides config)")
	sampleSize := fs.Int("sample", 0, "Max files to blame per repo (overrides config)")
	workers := fs.Int("workers", 4, "Number of concurrent blame workers")
	recursive := fs.Bool("recursive", false, "Recursively find git repos under given paths")
	maxDepth := fs.Int("depth", 2, "Max directory depth for recursive search")
	formatFlag := fs.String("format", "table", "Output format: table, csv, json")
	pressureMode := fs.String("pressure-mode", "include", "Change pressure mode: include (split robust/dormant) or ignore (classic survival)")
	activeDays := fs.Int("active-days", 0, "Days to consider author active (overrides config, default 30)")
	domainFilter := fs.String("domain", "", "Only analyze repos in this domain (e.g. Backend, Frontend, Firmware)")

	flagArgs, pathArgs := separateArgs(args, fs)
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	opts := AnalyzeOptions{
		ConfigPath:   *configPath,
		Tau:          *tau,
		SampleSize:   *sampleSize,
		Workers:      *workers,
		Recursive:    *recursive,
		MaxDepth:     *maxDepth,
		Format:       *formatFlag,
		PressureMode: *pressureMode,
		ActiveDays:   *activeDays,
		DomainFilter: *domainFilter,
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
		case "csv":
			output.PrintRankingsCSV(string(dr.Domain), dr.Results, !csvHeaderWritten)
			csvHeaderWritten = true
		default:
			fmt.Println()
			color.New(color.FgHiCyan, color.Bold).Printf("═══ %s ═══\n", dr.Domain)
			output.PrintSummary(dr.Results, dr.RepoCount)
			output.PrintRankings(dr.Results)
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
	cfg, err := config.Load(opts.ConfigPath)
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

	// Per-domain accumulators
	accumulators := make(map[domain.Domain]*domainAccumulator)

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
		repoDomain := resolveRepoDomain(ctx, repoPath, repoName, cfg)

		// Skip repos outside the requested domain
		if opts.DomainFilter != "" && !strings.EqualFold(string(repoDomain), opts.DomainFilter) {
			continue
		}

		bold := color.New(color.Bold)
		domainLabel := color.New(color.FgCyan).Sprintf("[%s]", repoDomain)
		bold.Printf("Analyzing: %s %s\n", repoName, domainLabel)

		// Get or create accumulator for this domain
		acc, ok := accumulators[repoDomain]
		if !ok {
			acc = newDomainAccumulator()
			accumulators[repoDomain] = acc
		}
		acc.repoCount++
		totalAnalyzed++

		// Step 1: Parse git log (feeds Production, Quality, Design)
		fmt.Fprintf(os.Stderr, "  [1/4] Parsing git log...\n")
		commits, err := git.ParseLog(ctx, repoPath)
		if err != nil {
			return nil, nil, fmt.Errorf("parse log %s: %w", repoName, err)
		}

		// Apply author aliases, filter excluded authors, and strip excluded file patterns
		commits = filterCommits(commits, cfg)
		commits = filterFileStats(commits, cfg.ExcludeFilePatterns)

		// Production
		prod := metric.CalcProduction(commits, cfg.ExcludeFilePatterns)
		mergeMap(acc.raw.Production, prod)

		// Quality
		qual := metric.CalcQuality(commits)
		mergeMapAvg(acc.raw.Quality, qual, acc.qualityCounts)

		// Design
		design := metric.CalcDesign(commits, cfg.ArchitecturePatterns)
		mergeMap(acc.raw.Design, design)

		// Track breadth with commit counts per repo, and date ranges for production rate
		for _, c := range commits {
			if _, ok := acc.authorRepoCommits[c.Author]; !ok {
				acc.authorRepoCommits[c.Author] = make(map[string]int)
			}
			acc.authorRepoCommits[c.Author][repoName]++
			acc.raw.TotalCommits[c.Author]++

			// Track earliest and latest commit dates per author
			if first, ok := acc.authorFirstDate[c.Author]; !ok || c.Date.Before(first) {
				acc.authorFirstDate[c.Author] = c.Date
			}
			if last, ok := acc.authorLastDate[c.Author]; !ok || c.Date.After(last) {
				acc.authorLastDate[c.Author] = c.Date
			}
		}

		// Step 2: Blame analysis (feeds Survival, Indispensability)
		fmt.Fprintf(os.Stderr, "  [2/4] Blame analysis...\n")
		files, err := git.ListFiles(ctx, repoPath, cfg.BlameExtensions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: could not list files: %v\n", err)
			continue
		}

		// Filter out excluded file patterns from blame targets
		files = filterFiles(files, cfg.ExcludeFilePatterns)

		blameLines, err := git.ConcurrentBlameFiles(ctx, repoPath, files, cfg.SampleSize, workers,
			func(done, total int) {
				fmt.Fprintf(os.Stderr, "%s\r", progressBar("[2/4] Blame", done, total))
			})
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: blame error: %v\n", err)
		}
		fmt.Fprintln(os.Stderr)

		// Apply aliases to blame lines
		for i := range blameLines {
			blameLines[i].Author = cfg.ResolveAuthor(blameLines[i].Author)
		}
		blameLines = filterBlameLines(blameLines, cfg)

		// Survival: split by change pressure or use classic mode
		if opts.PressureMode == "include" {
			repoPressure := metric.CalcChangePressure(commits, blameLines)
			for mod, p := range repoPressure {
				key := repoName + "/" + mod
				acc.changePressure[key] = p
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
			survResult := metric.CalcSurvivalWithPressure(blameLines, cfg.Tau, start, repoPressure, pressureThreshold)
			mergeMap(acc.raw.Survival, survResult.Decayed)
			mergeMap(acc.raw.RawSurvival, survResult.Raw)
			mergeMap(acc.raw.RobustSurvival, survResult.Robust)
			mergeMap(acc.raw.DormantSurvival, survResult.Dormant)
		} else {
			// Classic mode: single survival score, no pressure split
			survResult := metric.CalcSurvival(blameLines, cfg.Tau, start)
			mergeMap(acc.raw.Survival, survResult.Decayed)
			mergeMap(acc.raw.RawSurvival, survResult.Raw)
		}

		// Indispensability
		indisp, risks := metric.CalcIndispensability(blameLines, cfg.BusFactor.Critical, cfg.BusFactor.High)
		mergeMap(acc.raw.Indispensability, indisp)

		// Step 3: Debt cleanup
		fixCommits := metric.GetFixCommits(commits)
		fmt.Fprintf(os.Stderr, "  [3/4] Debt analysis (%d fix commits)...\n", len(fixCommits))
		debt, _ := metric.CalcDebt(ctx, repoPath, fixCommits, 50, cfg.DebtThreshold, cfg.ResolveAuthor,
			func(done, total int) {
				fmt.Fprintf(os.Stderr, "%s\r", progressBar("[3/4] Debt", done, total))
			})
		fmt.Fprintln(os.Stderr)
		mergeMapAvg(acc.raw.DebtCleanup, debt, acc.debtCounts)

		// Step 4: Accumulate bus factor risks per domain; print immediately for table format
		acc.risks = append(acc.risks, risks...)
		if opts.Format == "table" && len(risks) > 0 {
			output.PrintBusFactorRisks(risks)
		}
	}

	// Score per domain
	domains := domain.AllDomains()
	if _, ok := accumulators[domain.Unknown]; ok {
		domains = append(domains, domain.Unknown)
	}

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

		// Filter out excluded authors from results
		var filtered []scorer.Result
		for _, r := range scored {
			if !cfg.IsExcludedAuthor(r.Author) {
				filtered = append(filtered, r)
			}
		}

		if len(filtered) == 0 {
			continue
		}

		results = append(results, DomainResults{
			Domain:    d,
			Results:   filtered,
			Risks:     acc.risks,
			RepoCount: acc.repoCount,
		})
	}

	if opts.Format == "table" {
		elapsed := time.Since(start)
		color.New(color.FgHiBlack).Printf("Completed in %s (%d repos total)\n", elapsed.Round(time.Second), totalAnalyzed)
	}

	return results, cfg, nil
}

// resolveRepoDomain determines the domain for a repo.
// Config overrides take priority, then auto-detection from file extensions.
func resolveRepoDomain(ctx context.Context, repoPath, repoName string, cfg *config.Config) domain.Domain {
	// Check config overrides first
	if domain.MatchRepoPattern(repoName, cfg.Domains.Backend) {
		return domain.Backend
	}
	if domain.MatchRepoPattern(repoName, cfg.Domains.Frontend) {
		return domain.Frontend
	}
	if domain.MatchRepoPattern(repoName, cfg.Domains.Infra) {
		return domain.Infra
	}
	if domain.MatchRepoPattern(repoName, cfg.Domains.Firmware) {
		return domain.Firmware
	}

	// Auto-detect from file extensions
	files, err := git.ListAllFiles(ctx, repoPath)
	if err != nil || len(files) == 0 {
		return domain.Unknown
	}

	return domain.DetectFromFiles(files)
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

// progressBar renders a compact progress bar with color: [████░░░░░░] 12/50
func progressBar(label string, done, total int) string {
	const barWidth = 20
	pct := float64(done) / float64(total)
	filled := int(pct * barWidth)
	if filled > barWidth {
		filled = barWidth
	}
	cyan := color.New(color.FgCyan)
	dim := color.New(color.FgHiBlack)
	green := color.New(color.FgGreen)
	filledBar := cyan.Sprint(strings.Repeat("█", filled))
	emptyBar := dim.Sprint(strings.Repeat("░", barWidth-filled))
	count := green.Sprintf("%d/%d", done, total)
	return fmt.Sprintf("  %s [%s%s] %s", label, filledBar, emptyBar, count)
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
