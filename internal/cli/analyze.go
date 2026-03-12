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

	if err := fs.Parse(args); err != nil {
		return err
	}

	repoPaths := fs.Args()
	if len(repoPaths) == 0 {
		repoPaths = []string{"."}
	}

	// Resolve to absolute paths
	for i, p := range repoPaths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("resolve path %s: %w", p, err)
		}
		repoPaths[i] = abs
	}

	// Recursive: find git repos under given paths
	if *recursive {
		var discovered []string
		for _, root := range repoPaths {
			repos, err := findGitRepos(root, *maxDepth)
			if err != nil {
				return fmt.Errorf("scan %s: %w", root, err)
			}
			discovered = append(discovered, repos...)
		}
		if len(discovered) == 0 {
			return fmt.Errorf("no git repos found under %v (depth=%d)", repoPaths, *maxDepth)
		}
		repoPaths = discovered
		fmt.Fprintf(os.Stderr, "Found %d git repos\n\n", len(repoPaths))
	}

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if *tau > 0 {
		cfg.Tau = *tau
	}
	if *sampleSize > 0 {
		cfg.SampleSize = *sampleSize
	}
	if *activeDays > 0 {
		cfg.ActiveDays = *activeDays
	}

	// Quiet mode for structured output (suppress progress to stderr)
	quiet := *formatFlag == "json" || *formatFlag == "csv"

	// Print alias info if configured
	if !quiet && len(cfg.Aliases) > 0 {
		fmt.Fprintf(os.Stderr, "Loaded %d author aliases from config\n", len(cfg.Aliases))
	}

	ctx := context.Background()
	start := time.Now()

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
		if *domainFilter != "" && !strings.EqualFold(string(repoDomain), *domainFilter) {
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
			return fmt.Errorf("parse log %s: %w", repoName, err)
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

		blameLines, err := git.ConcurrentBlameFiles(ctx, repoPath, files, cfg.SampleSize, *workers,
			func(done, total int) {
				fmt.Fprintf(os.Stderr, "  [2/4] Blame: %d/%d files\r", done, total)
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
		if *pressureMode == "include" {
			repoPressure := metric.CalcChangePressure(commits, blameLines)
			for mod, p := range repoPressure {
				key := repoName + "/" + mod
				acc.changePressure[key] = p
			}

			// Need at least 2 authors with ≥5000 blame lines for pressure
			// split to be meaningful. Below that, the repo is effectively
			// solo-owned and has no real change pressure.
			blameByAuthor := make(map[string]int)
			for _, bl := range blameLines {
				blameByAuthor[cfg.ResolveAuthor(bl.Author)]++
			}
			substantialAuthors := 0
			for _, count := range blameByAuthor {
				if count >= 5000 {
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
		fmt.Fprintf(os.Stderr, "  [3/4] Debt analysis...\n")
		fixCommits := metric.GetFixCommits(commits)
		debt, _ := metric.CalcDebt(ctx, repoPath, fixCommits, 50, cfg.DebtThreshold, cfg.ResolveAuthor)
		mergeMapAvg(acc.raw.DebtCleanup, debt, acc.debtCounts)

		// Step 4: Accumulate bus factor risks per domain; print immediately for table format
		acc.risks = append(acc.risks, risks...)
		if *formatFlag == "table" && len(risks) > 0 {
			output.PrintBusFactorRisks(risks)
		}
	}

	// Score and output per domain
	domains := domain.AllDomains()
	// Also include Unknown if present
	if _, ok := accumulators[domain.Unknown]; ok {
		domains = append(domains, domain.Unknown)
	}

	// For JSON, accumulate all domains before writing
	var jsonWriter *output.JSONWriter
	if *formatFlag == "json" {
		jsonWriter = output.NewJSONWriter()
	}

	csvHeaderWritten := false

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
		results := scorer.Score(acc.raw, cfg, acc.authorLastDate)

		// Filter out excluded authors from results
		var filtered []scorer.Result
		for _, r := range results {
			if !cfg.IsExcludedAuthor(r.Author) {
				filtered = append(filtered, r)
			}
		}

		if len(filtered) == 0 {
			continue
		}

		switch *formatFlag {
		case "json":
			jsonWriter.AddDomain(string(d), acc.repoCount, filtered, acc.risks)
		case "csv":
			output.PrintRankingsCSV(string(d), filtered, !csvHeaderWritten)
			csvHeaderWritten = true
		default:
			fmt.Println()
			color.New(color.FgHiCyan, color.Bold).Printf("═══ %s ═══\n", d)
			output.PrintSummary(filtered, acc.repoCount)
			output.PrintRankings(filtered)
		}
	}

	if *formatFlag == "json" {
		if err := jsonWriter.Flush(); err != nil {
			return fmt.Errorf("json output: %w", err)
		}
	}

	if *formatFlag == "table" {
		elapsed := time.Since(start)
		color.New(color.FgHiBlack).Printf("Completed in %s (%d repos total)\n", elapsed.Round(time.Second), totalAnalyzed)
	}

	return nil
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
