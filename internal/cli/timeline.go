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
	"github.com/machuz/engineering-impact-score/internal/cache"
	"github.com/machuz/engineering-impact-score/internal/config"
	"github.com/machuz/engineering-impact-score/internal/domain"
	"github.com/machuz/engineering-impact-score/internal/git"
	"github.com/machuz/engineering-impact-score/internal/metric"
	"github.com/machuz/engineering-impact-score/internal/output"
	"github.com/machuz/engineering-impact-score/internal/scorer"
	"github.com/machuz/engineering-impact-score/internal/team"
	"github.com/machuz/engineering-impact-score/internal/timeline"
)

// TimeWindow represents a time period for timeline analysis.
type TimeWindow struct {
	Label string
	Start time.Time
	End   time.Time
}

func runTimeline(args []string) error {
	fs := flag.NewFlagSet("timeline", flag.ExitOnError)
	configPath := fs.String("config", "eis.yaml", "Config file path")
	spanFlag := fs.String("span", "3m", "Period span: 3m, 6m, 1y")
	periodsFlag := fs.Int("periods", 4, "Number of periods to show (0=all)")
	sinceFlag := fs.String("since", "", "Start date (e.g. 2024-01-01, overrides --periods)")
	formatFlag := fs.String("format", "table", "Output format: table, csv, json, ascii, html, svg")
	outputFlag := fs.String("output", "", "Output file/directory path (html: file path, svg: directory; defaults: eis-timeline.html / current dir)")
	recursive := fs.Bool("recursive", false, "Recursively find git repos under given paths")
	maxDepth := fs.Int("depth", 2, "Max directory depth for recursive search")
	domainFilter := fs.String("domain", "", "Only analyze specific domain")
	authorFilter := fs.String("author", "", "Filter to specific author(s), comma-separated")
	workers := fs.Int("workers", 4, "Number of concurrent blame workers")
	sampleSize := fs.Int("sample", 0, "Max files to blame per repo (overrides config)")
	tau := fs.Float64("tau", 0, "Survival decay parameter (overrides config)")
	activeDays := fs.Int("active-days", 0, "Days to consider author active (overrides config)")
	pressureMode := fs.String("pressure-mode", "include", "Change pressure mode: include or ignore")
	verbose := fs.Bool("verbose", false, "Show detailed debug output")
	noCache := fs.Bool("no-cache", false, "Skip disk cache")

	flagArgs, pathArgs := separateArgs(args, fs)
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	repoPaths := pathArgs
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

	// Recursive: find git repos
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

	quiet := *formatFlag == "json" || *formatFlag == "csv" || *formatFlag == "html" || *formatFlag == "svg"
	spinnerQuiet = quiet

	if !quiet && len(cfg.Aliases) > 0 {
		fmt.Fprintf(os.Stderr, "Loaded %d author aliases from config\n", len(cfg.Aliases))
	}

	now := time.Now()

	// Parse span
	spanMonths, err := parseSpan(*spanFlag)
	if err != nil {
		return err
	}

	// Parse since
	var sinceDate time.Time
	if *sinceFlag != "" {
		sinceDate, err = time.Parse("2006-01-02", *sinceFlag)
		if err != nil {
			return fmt.Errorf("invalid --since date: %w", err)
		}
	}

	// Build time windows
	windows := buildPeriods(spanMonths, *periodsFlag, sinceDate, now)
	if len(windows) == 0 {
		return fmt.Errorf("no periods to analyze")
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "Timeline: %d periods (%s span)\n", len(windows), *spanFlag)
		for _, w := range windows {
			fmt.Fprintf(os.Stderr, "  %s: %s → %s\n", w.Label, w.Start.Format("2006-01-02"), w.End.Format("2006-01-02"))
		}
		fmt.Fprintln(os.Stderr)
	}

	// Parse author filter
	var authorList []string
	if *authorFilter != "" {
		for _, a := range strings.Split(*authorFilter, ",") {
			a = strings.TrimSpace(a)
			if a != "" {
				authorList = append(authorList, a)
			}
		}
	}

	// Remove specified authors from exclude list so they appear in results.
	// Resolves aliases: --author kakki matches both "kakki" and any alias
	// that resolves to "kakki" (e.g. "syunto07ka" -> "kakki").
	if len(authorList) > 0 {
		// Build set of canonical names to include
		includeCanonical := make(map[string]bool)
		for _, a := range authorList {
			includeCanonical[strings.ToLower(cfg.ResolveAuthor(a))] = true
		}

		var newExclude []string
		for _, exc := range cfg.ExcludeAuthors {
			canonical := strings.ToLower(cfg.ResolveAuthor(exc))
			if !includeCanonical[strings.ToLower(exc)] && !includeCanonical[canonical] {
				newExclude = append(newExclude, exc)
			}
		}
		cfg.ExcludeAuthors = newExclude
	}

	ctx := context.Background()
	wk := *workers
	if wk == 0 {
		wk = 4
	}

	// Initialize cache
	cacheStore := cache.New(!*noCache)

	// Deduplicate repos
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
		}
	}
	repoPaths = dedupedPaths

	// Group repos by domain, collecting all commits and repo metadata
	type repoInfo struct {
		path    string
		name    string
		domain  domain.Domain
		commits []git.Commit
		merges  []git.Commit
	}

	var repos []repoInfo
	for _, repoPath := range repoPaths {
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
			continue
		}
		repoName := filepath.Base(repoPath)
		if cfg.IsExcludedRepo(repoName) {
			continue
		}

		repoDomain := resolveRepoDomain(ctx, repoPath, repoName, cfg)
		if *domainFilter != "" && !strings.EqualFold(string(repoDomain), *domainFilter) {
			continue
		}

		if !quiet {
			domainLabel := color.New(color.FgCyan).Sprintf("[%s]", repoDomain)
			color.New(color.Bold).Printf("Loading: %s %s\n", repoName, domainLabel)
		}

		// Get HEAD hash for cache keys
		headHash, _ := git.HeadHash(ctx, repoPath)

		// Parse all commits once (with cache)
		var commits []git.Commit
		logCacheKey := cache.LogKey(repoPath, headHash)
		if headHash != "" && cacheStore.Get(logCacheKey, &commits) {
			if !quiet {
				fmt.Fprintf(os.Stderr, "  %s git log (cached)\n", color.New(color.FgGreen).Sprint("✓"))
			}
		} else {
			spin := spinner("[1/1] Parsing git log...")
			commits, err = git.ParseLog(ctx, repoPath)
			spin.Stop()
			if err != nil {
				return fmt.Errorf("parse log %s: %w", repoName, err)
			}
			if headHash != "" {
				cacheStore.Set(logCacheKey, commits)
			}
		}
		commits = filterCommits(commits, cfg)
		commits = filterFileStats(commits, cfg.ExcludeFilePatterns)

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

		repos = append(repos, repoInfo{
			path:    repoPath,
			name:    repoName,
			domain:  repoDomain,
			commits: commits,
			merges:  mergeCommits,
		})
	}

	// Group by domain
	domainRepos := make(map[domain.Domain][]repoInfo)
	for _, r := range repos {
		domainRepos[r.domain] = append(domainRepos[r.domain], r)
	}

	// For each domain, run analysis per period
	type domainTimeline struct {
		domain  domain.Domain
		periods []timeline.PeriodResult
	}

	var domainTimelines []domainTimeline

	allDomains := domain.AllDomains()
	if _, ok := domainRepos[domain.Unknown]; ok {
		allDomains = append(allDomains, domain.Unknown)
	}

	for _, d := range allDomains {
		drepos, ok := domainRepos[d]
		if !ok {
			continue
		}

		if !quiet {
			fmt.Fprintln(os.Stderr)
			color.New(color.FgHiCyan, color.Bold).Fprintf(os.Stderr, "═══ %s Timeline ═══\n", d)
		}

		var periodResults []timeline.PeriodResult

		for pi, window := range windows {
			if !quiet {
				fmt.Fprintf(os.Stderr, "\n[Period %d/%d] %s\n", pi+1, len(windows), window.Label)
			}

			acc := newDomainAccumulator()

			for _, repo := range drepos {
				// Filter commits to this period
				var periodCommits []git.Commit
				for _, c := range repo.commits {
					if !c.Date.Before(window.Start) && c.Date.Before(window.End) {
						periodCommits = append(periodCommits, c)
					}
				}

				var periodMerges []git.Commit
				for _, c := range repo.merges {
					if !c.Date.Before(window.Start) && c.Date.Before(window.End) {
						periodMerges = append(periodMerges, c)
					}
				}

				acc.repoCount++

				// Production
				prod := metric.CalcProduction(periodCommits, cfg.ExcludeFilePatterns)
				mergeMap(acc.raw.Production, prod)

				// Quality
				allCommits := make([]git.Commit, len(periodCommits), len(periodCommits)+len(periodMerges))
				copy(allCommits, periodCommits)
				allCommits = append(allCommits, periodMerges...)
				qual := metric.CalcQuality(allCommits)
				mergeMapAvg(acc.raw.Quality, qual, acc.qualityCounts)

				// Design
				design := metric.CalcDesign(periodCommits, cfg.ArchitecturePatterns)
				mergeMap(acc.raw.Design, design)

				// Track breadth and date ranges
				for _, c := range periodCommits {
					if _, ok := acc.authorRepoCommits[c.Author]; !ok {
						acc.authorRepoCommits[c.Author] = make(map[string]int)
					}
					acc.authorRepoCommits[c.Author][repo.name]++
					acc.raw.TotalCommits[c.Author]++

					if first, ok := acc.authorFirstDate[c.Author]; !ok || c.Date.Before(first) {
						acc.authorFirstDate[c.Author] = c.Date
					}
					if last, ok := acc.authorLastDate[c.Author]; !ok || c.Date.After(last) {
						acc.authorLastDate[c.Author] = c.Date
					}
				}

				// Blame: find the latest commit hash at period end for this repo
				var blameVerbose func(string)
				if *verbose {
					blameVerbose = func(msg string) {
						fmt.Fprintf(os.Stderr, "\n%s", msg)
					}
				}

				boundaryCommit, err := git.FindCommitAtDate(ctx, repo.path, window.End)
				if err != nil {
					// No commits before this period end — skip blame
					continue
				}

				files, err := git.ListFilesAtCommit(ctx, repo.path, boundaryCommit, cfg.BlameExtensions)
				if err != nil {
					continue
				}
				files = filterFiles(files, cfg.ExcludeFilePatterns)

				if len(files) == 0 {
					continue
				}

				var blameLines []git.BlameLine
				blameCacheKey := cache.BlameAtCommitKey(repo.path, boundaryCommit, files, cfg.SampleSize)
				if cacheStore.Get(blameCacheKey, &blameLines) {
					if !quiet {
						fmt.Fprintf(os.Stderr, "  %s Blame %s (cached)\n", color.New(color.FgGreen).Sprint("✓"), repo.name)
					}
				} else {
					if !quiet {
						spin := spinner(fmt.Sprintf("  Blame %s @ %s...", repo.name, window.Label))
						spin.Clear()
					}
					blameProg := newLiveProgress(fmt.Sprintf("  Blame %s", repo.name))
					blameLines, err = git.ConcurrentBlameFilesAtCommit(ctx, repo.path, boundaryCommit, files, cfg.SampleSize, wk,
						func(done, total int) {
							blameProg.Update(done, total)
						}, blameVerbose)
					blameProg.Stop()
					if err != nil {
						fmt.Fprintf(os.Stderr, "  Warning: blame error: %v\n", err)
					}
					if len(blameLines) > 0 {
						cacheStore.Set(blameCacheKey, blameLines)
					}
				}

				// Apply aliases
				for i := range blameLines {
					blameLines[i].Author = cfg.ResolveAuthor(blameLines[i].Author)
				}
				blameLines = filterBlameLines(blameLines, cfg)

				// Survival
				if *pressureMode == "include" {
					repoPressure := metric.CalcChangePressure(periodCommits, blameLines)
					for mod, p := range repoPressure {
						key := repo.name + "/" + mod
						acc.changePressure[key] = p
					}

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
						pressureThreshold = math.Inf(1)
					}
					survResult := metric.CalcSurvivalWithPressure(blameLines, cfg.Tau, window.End, repoPressure, pressureThreshold)
					mergeMap(acc.raw.Survival, survResult.Decayed)
					mergeMap(acc.raw.RawSurvival, survResult.Raw)
					mergeMap(acc.raw.RobustSurvival, survResult.Robust)
					mergeMap(acc.raw.DormantSurvival, survResult.Dormant)
				} else {
					survResult := metric.CalcSurvival(blameLines, cfg.Tau, window.End)
					mergeMap(acc.raw.Survival, survResult.Decayed)
					mergeMap(acc.raw.RawSurvival, survResult.Raw)
				}

				// Indispensability
				indisp, _ := metric.CalcIndispensability(blameLines, cfg.BusFactor.Critical, cfg.BusFactor.High)
				mergeMap(acc.raw.Indispensability, indisp)

				// Debt
				fixCommits := metric.GetFixCommits(periodCommits)
				if len(fixCommits) > 0 {
					var debt map[string]float64
					var fixHashes []string
					for _, fc := range fixCommits {
						fixHashes = append(fixHashes, fc.Hash)
					}
					debtCacheKey := cache.DebtKey(repo.path, fixHashes)
					if cacheStore.Get(debtCacheKey, &debt) {
						// cached
					} else {
						debt, _ = metric.CalcDebt(ctx, repo.path, fixCommits, 50, cfg.DebtThreshold, cfg.BlameTimeout, cfg.ResolveAuthor, nil, nil)
						if len(debt) > 0 {
							cacheStore.Set(debtCacheKey, debt)
						}
					}
					mergeMapAvg(acc.raw.DebtCleanup, debt, acc.debtCounts)
				}
			}

			// Score this period
			// Breadth
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

			// Convert production to per-day rate
			for author, total := range acc.raw.Production {
				first := acc.authorFirstDate[author]
				last := acc.authorLastDate[author]
				days := last.Sub(first).Hours() / 24
				if days < 1 {
					days = 1
				}
				acc.raw.Production[author] = total / days
			}

			// For timeline, anyone who committed in the period is "recently active".
			// Override ActiveDays to cover the full period span so team.Aggregate
			// includes them as core members.
			periodCfg := *cfg
			periodDays := int(window.End.Sub(window.Start).Hours()/24) + 1
			if periodDays > periodCfg.ActiveDays {
				periodCfg.ActiveDays = periodDays
			}
			scored := scorer.ScoreAt(acc.raw, &periodCfg, acc.authorLastDate, window.End)

			// Filter excluded authors
			var filtered []scorer.Result
			for _, r := range scored {
				if !cfg.IsExcludedAuthor(r.Author) {
					filtered = append(filtered, r)
				}
			}

			periodResults = append(periodResults, timeline.PeriodResult{
				Label:   window.Label,
				Start:   window.Start.Format("2006-01-02"),
				End:     window.End.Format("2006-01-02"),
				Members: filtered,
			})
		}

		if len(periodResults) > 0 {
			domainTimelines = append(domainTimelines, domainTimeline{
				domain:  d,
				periods: periodResults,
			})
		}
	}

	// Build per-domain author timelines (filtered)
	type builtDomainTimeline struct {
		domain    domain.Domain
		span      string
		periods   []timeline.PeriodResult
		timelines []timeline.AuthorTimeline
	}

	var builtTimelines []builtDomainTimeline
	for _, dt := range domainTimelines {
		timelines := timeline.BuildTimeline(dt.periods)

		// Apply author filter
		if len(authorList) > 0 {
			var filtered []timeline.AuthorTimeline
			for _, tl := range timelines {
				for _, a := range authorList {
					if strings.EqualFold(tl.Author, a) {
						filtered = append(filtered, tl)
						break
					}
				}
			}
			timelines = filtered
		}

		builtTimelines = append(builtTimelines, builtDomainTimeline{
			domain:    dt.domain,
			span:      *spanFlag,
			periods:   dt.periods,
			timelines: timelines,
		})
	}

	// Build team timelines
	var teamTimelines []timeline.TeamTimeline
	for _, dt := range domainTimelines {
		teamPeriodResults := buildTeamPeriodResults(dt.domain, dt.periods, cfg)

		for teamName, periods := range teamPeriodResults {
			domainStr := string(dt.domain)
			tl := timeline.BuildTeamTimeline(teamName, domainStr, periods)
			teamTimelines = append(teamTimelines, tl)
		}
	}

	// Output individual timelines
	if *formatFlag == "html" || *formatFlag == "svg" {
		// Collect all domain data (shared by html and svg)
		var htmlDomains []output.DomainTimelineData
		for _, bt := range builtTimelines {
			htmlDomains = append(htmlDomains, output.DomainTimelineData{
				DomainName: string(bt.domain),
				Span:       bt.span,
				Timelines:  bt.timelines,
			})
		}

		if *formatFlag == "html" {
			outPath := *outputFlag
			if outPath == "" {
				outPath = "eis-timeline.html"
			}

			f, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("create output file: %w", err)
			}
			defer f.Close()

			if err := output.WriteTimelineHTML(f, htmlDomains, teamTimelines); err != nil {
				return fmt.Errorf("write HTML: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Timeline HTML written to %s\n", outPath)
		} else {
			// SVG
			outDir := *outputFlag
			if outDir == "" {
				outDir = "."
			}

			files, err := output.WriteTimelineSVG(outDir, htmlDomains, teamTimelines)
			if err != nil {
				return fmt.Errorf("write SVG: %w", err)
			}

			for _, f := range files {
				fmt.Fprintf(os.Stderr, "SVG written: %s\n", f)
			}
			fmt.Fprintf(os.Stderr, "Generated %d SVG files\n", len(files))
		}
	} else {
		for _, bt := range builtTimelines {
			switch *formatFlag {
			case "json":
				output.PrintTimelineJSON(string(bt.domain), *spanFlag, bt.periods, bt.timelines)
			case "csv":
				output.PrintTimelineCSV(string(bt.domain), bt.timelines)
			case "ascii":
				output.PrintTimelineASCII(string(bt.domain), *spanFlag, bt.timelines)
			default:
				output.PrintTimelineTable(string(bt.domain), *spanFlag, bt.timelines)
			}
		}

		if len(teamTimelines) > 0 {
			switch *formatFlag {
			case "json":
				output.PrintTeamTimelineJSON(teamTimelines)
			case "csv":
				output.PrintTeamTimelineCSV(teamTimelines)
			case "ascii":
				for _, tl := range teamTimelines {
					output.PrintTeamTimelineASCII(tl)
				}
			default:
				for _, tl := range teamTimelines {
					output.PrintTeamTimelineTable(tl)
				}
			}
		}
	}

	if !quiet {
		color.New(color.FgHiBlack).Fprintf(os.Stderr, "\nTimeline analysis complete\n")
	}

	return nil
}

// buildTeamPeriodResults aggregates per-period scored results into TeamPeriodResults.
// Returns map[teamName][]TeamPeriodResult.
func buildTeamPeriodResults(d domain.Domain, periods []timeline.PeriodResult, cfg *config.Config) map[string][]timeline.TeamPeriodResult {
	result := make(map[string][]timeline.TeamPeriodResult)

	if len(cfg.Teams) > 0 {
		// Use configured teams
		for teamName, entry := range cfg.Teams {
			if !strings.EqualFold(entry.Domain, string(d)) {
				continue
			}
			for _, p := range periods {
				tr := team.Aggregate(teamName, entry.Domain, 0, p.Members, entry.Members)
				result[teamName] = append(result[teamName], timeline.TeamPeriodResult{
					Label:      p.Label,
					Start:      p.Start,
					End:        p.End,
					TeamResult: tr,
				})
			}
		}
	} else {
		// Each domain = one team
		teamName := string(d)
		for _, p := range periods {
			tr := team.Aggregate(teamName, string(d), 0, p.Members, nil)
			result[teamName] = append(result[teamName], timeline.TeamPeriodResult{
				Label:      p.Label,
				Start:      p.Start,
				End:        p.End,
				TeamResult: tr,
			})
		}
	}

	return result
}

// parseSpan converts span string to months.
func parseSpan(s string) (int, error) {
	switch s {
	case "3m":
		return 3, nil
	case "6m":
		return 6, nil
	case "1y":
		return 12, nil
	default:
		return 0, fmt.Errorf("invalid span %q (use 3m, 6m, or 1y)", s)
	}
}

// buildPeriods creates time windows from now backwards.
func buildPeriods(spanMonths int, numPeriods int, since time.Time, now time.Time) []TimeWindow {
	if since.IsZero() && numPeriods == 0 {
		// All history: find earliest reasonable start (10 years back max)
		since = now.AddDate(-10, 0, 0)
	}

	var windows []TimeWindow

	if !since.IsZero() {
		// From since to now
		current := since
		for current.Before(now) {
			end := current.AddDate(0, spanMonths, 0)
			if end.After(now) {
				end = now
			}
			windows = append(windows, TimeWindow{
				Label: periodLabel(current, spanMonths),
				Start: current,
				End:   end,
			})
			current = end
		}
	} else {
		// Work backwards from now
		for i := numPeriods - 1; i >= 0; i-- {
			end := now.AddDate(0, -spanMonths*i, 0)
			start := end.AddDate(0, -spanMonths, 0)
			windows = append(windows, TimeWindow{
				Label: periodLabel(start, spanMonths),
				Start: start,
				End:   end,
			})
		}
	}

	return windows
}

// periodLabel generates a human-readable label for a period.
func periodLabel(start time.Time, spanMonths int) string {
	year := start.Year()
	month := start.Month()

	switch spanMonths {
	case 3:
		q := (int(month) - 1) / 3
		qLabels := []string{"Q1 (Jan)", "Q2 (Apr)", "Q3 (Jul)", "Q4 (Oct)"}
		return fmt.Sprintf("%d-%s", year, qLabels[q])
	case 6:
		if month <= 6 {
			return fmt.Sprintf("%d-H1", year)
		}
		return fmt.Sprintf("%d-H2", year)
	case 12:
		return fmt.Sprintf("%d", year)
	default:
		return fmt.Sprintf("%d-%02d", year, month)
	}
}
