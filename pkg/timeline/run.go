// Package timeline provides a library-friendly timeline analysis pipeline.
// This is the public API for timeline analysis without CLI dependencies
// (no spinner, color, cache, flag parsing).
package timeline

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/machuz/eis/v2/internal/cache"
	"github.com/machuz/eis/v2/internal/config"
	"github.com/machuz/eis/v2/internal/domain"
	"github.com/machuz/eis/v2/internal/git"
	"github.com/machuz/eis/v2/internal/metric"
	"github.com/machuz/eis/v2/internal/scorer"
	"github.com/machuz/eis/v2/internal/team"
	"github.com/machuz/eis/v2/internal/timeline"
)

// Options controls timeline analysis behavior.
type Options struct {
	Span         string // "1w", "1m", "3m", "6m", "1y"
	Periods      int    // number of periods to show (0 = all history)
	Since        string // ISO date (e.g. "2024-01-01"), overrides Periods
	Workers      int
	DomainFilter string
	PressureMode string // "include" or "ignore"
	Tau          float64
	SampleSize   int
	ActiveDays   int
	CacheEnabled bool // enable disk cache for blame/log results
	// PerRepo, when true, computes per-(repo, domain, period) scored
	// results in addition to the merged per-domain rollup. The result
	// lands in PeriodResult.PerRepo. Default false to keep existing CLI
	// callers byte-stable; SaaS callers (which persist per-period
	// snapshots and need per-repo breakdowns for StarDetail) opt in.
	PerRepo bool
}

// Callbacks for progress reporting during timeline analysis.
type Callbacks struct {
	OnRepoStart     func(repoName string, domain string)
	OnRepoSkipped   func(repoName, reason string)
	OnPeriodStart   func(periodLabel string, index, total int)
	OnBlameProgress func(repoName string, done, total int)
	OnVerbose       func(msg string)

	// OnPeriodComplete fires once per timeline window after every domain
	// in that window has produced its scored result. The map is keyed
	// by domain name and mirrors the per-domain PeriodResult that ends
	// up in the final []DomainTimeline returned by Run.
	//
	// SaaS callers use this to persist each period as it completes —
	// pkgtimeline still returns the full slice on success, but a
	// streaming consumer no longer has to wait for every period to
	// land. If the host process is killed mid-run (worker timeout,
	// OOM, deploy), the previously emitted periods are already on
	// disk and the next run only has to fill in the missing tail.
	OnPeriodComplete func(domains map[string]PeriodResult)
}

// DomainTimeline holds timeline results for one domain.
type DomainTimeline struct {
	Domain  string
	Periods []PeriodResult
}

// TimeWindow represents a time period for timeline analysis.
type TimeWindow struct {
	Label string
	Start time.Time
	End   time.Time
}

// ParseSpan converts a span string to a SpanDuration.
// Supported values: "1w", "1m", "3m", "6m", "1y".
func ParseSpan(s string) (months int, days int, err error) {
	switch s {
	case "1w":
		return 0, 7, nil
	case "1m":
		return 1, 0, nil
	case "3m":
		return 3, 0, nil
	case "6m":
		return 6, 0, nil
	case "1y":
		return 12, 0, nil
	default:
		return 0, 0, fmt.Errorf("invalid span %q (use 1w, 1m, 3m, 6m, or 1y)", s)
	}
}

// addSpan advances a time by the given span (months or days).
func addSpan(t time.Time, months, days int) time.Time {
	if months > 0 {
		return t.AddDate(0, months, 0)
	}
	return t.AddDate(0, 0, days)
}

// subSpan subtracts a span from a time.
func subSpan(t time.Time, months, days int) time.Time {
	if months > 0 {
		return t.AddDate(0, -months, 0)
	}
	return t.AddDate(0, 0, -days)
}

// PeriodLabel generates a human-readable label for a period.
func PeriodLabel(start time.Time, spanMonths, spanDays int) string {
	year := start.Year()
	month := start.Month()

	switch {
	case spanDays == 7:
		_, week := start.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	case spanMonths == 1:
		return fmt.Sprintf("%d-%02d", year, month)
	case spanMonths == 3:
		q := (int(month) - 1) / 3
		qLabels := []string{"Q1 (Jan)", "Q2 (Apr)", "Q3 (Jul)", "Q4 (Oct)"}
		return fmt.Sprintf("%d-%s", year, qLabels[q])
	case spanMonths == 6:
		if month <= 6 {
			return fmt.Sprintf("%d-H1", year)
		}
		return fmt.Sprintf("%d-H2", year)
	case spanMonths == 12:
		return fmt.Sprintf("%d", year)
	default:
		return fmt.Sprintf("%d-%02d", year, month)
	}
}

// BuildPeriods creates time windows from now backwards (or from since forward).
func BuildPeriods(spanMonths, spanDays, numPeriods int, since time.Time, now time.Time) []TimeWindow {
	if since.IsZero() && numPeriods == 0 {
		// All history: 10 years back max
		since = now.AddDate(-10, 0, 0)
	}

	var windows []TimeWindow

	if !since.IsZero() {
		// From since to now
		current := since
		for current.Before(now) {
			end := addSpan(current, spanMonths, spanDays)
			if end.After(now) {
				end = now
			}
			windows = append(windows, TimeWindow{
				Label: PeriodLabel(current, spanMonths, spanDays),
				Start: current,
				End:   end,
			})
			current = end
		}
	} else {
		// Work backwards from now
		for i := numPeriods - 1; i >= 0; i-- {
			end := subSpan(now, spanMonths*i, spanDays*i)
			start := subSpan(end, spanMonths, spanDays)
			windows = append(windows, TimeWindow{
				Label: PeriodLabel(start, spanMonths, spanDays),
				Start: start,
				End:   end,
			})
		}
	}

	return windows
}

// Run executes the timeline analysis pipeline on the given repository paths.
// It returns per-domain timeline results without any CLI-specific behavior.
func Run(opts Options, repoPaths []string, cfg *config.Config, cb *Callbacks) ([]DomainTimeline, error) {
	if cb == nil {
		cb = &Callbacks{}
	}

	cacheStore := cache.New(opts.CacheEnabled)

	if opts.Tau > 0 {
		cfg.Tau = opts.Tau
	}
	if opts.SampleSize > 0 {
		cfg.SampleSize = opts.SampleSize
	}
	if opts.ActiveDays > 0 {
		cfg.ActiveDays = opts.ActiveDays
	}

	ctx := context.Background()
	workers := opts.Workers
	if workers == 0 {
		workers = 4
	}

	// Parse span
	spanMonths, spanDays, err := ParseSpan(opts.Span)
	if err != nil {
		return nil, err
	}

	// Parse since
	var sinceDate time.Time
	if opts.Since != "" {
		sinceDate, err = time.Parse("2006-01-02", opts.Since)
		if err != nil {
			return nil, fmt.Errorf("invalid since date: %w", err)
		}
	}

	periods := opts.Periods
	if periods == 0 && opts.Since == "" {
		periods = 0 // all history
	}

	now := time.Now()
	windows := BuildPeriods(spanMonths, spanDays, periods, sinceDate, now)
	if len(windows) == 0 {
		return nil, fmt.Errorf("no periods to analyze")
	}

	// Build extension-to-domain map
	extMap := domain.BuildExtMap(cfg.CustomExtensions(), cfg.UseDefaultDomains())

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

	// Collect repo info
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
			if cb.OnRepoSkipped != nil {
				cb.OnRepoSkipped(repoPath, "not a git repo")
			}
			continue
		}
		repoName := filepath.Base(repoPath)
		if cfg.IsExcludedRepo(repoName) {
			if cb.OnRepoSkipped != nil {
				cb.OnRepoSkipped(repoName, "excluded")
			}
			continue
		}

		repoDomain := resolveRepoDomain(ctx, repoPath, repoName, cfg, extMap)
		if opts.DomainFilter != "" && !strings.EqualFold(string(repoDomain), opts.DomainFilter) {
			continue
		}

		if cb.OnRepoStart != nil {
			cb.OnRepoStart(repoName, string(repoDomain))
		}

		headHash, _ := git.HeadHash(ctx, repoPath)

		var commits []git.Commit
		logCacheKey := cache.LogKey(repoPath, headHash)
		if headHash != "" && cacheStore.Get(logCacheKey, &commits) {
			// cached
		} else {
			commits, err = git.ParseLog(ctx, repoPath)
			if err != nil {
				return nil, fmt.Errorf("parse log %s: %w", repoName, err)
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

	var domainKeys []domain.Domain
	for d := range domainRepos {
		domainKeys = append(domainKeys, d)
	}
	allDomains := domain.SortDomains(domainKeys)

	// domainPeriodMap accumulates per-domain results as the period-outer
	// loop runs, so the original []DomainTimeline shape can be returned
	// on success even though we now compute period-by-period across all
	// domains (the order pkgtimeline used to compute domain-by-period).
	// Flipping the loop is what unlocks OnPeriodComplete: every domain
	// for a window finishes before the next window starts, so the
	// callback fires once per window with all domains' results.
	domainPeriodMap := make(map[string][]PeriodResult, len(allDomains))

	for pi, window := range windows {
		if cb.OnPeriodStart != nil {
			cb.OnPeriodStart(window.Label, pi, len(windows))
		}

		windowDomainResults := make(map[string]PeriodResult, len(allDomains))

		for _, d := range allDomains {
			drepos, ok := domainRepos[d]
			if !ok {
				continue
			}

			acc := newAccumulator()
			// repoAccs holds per-repo accumulators in parallel with the
			// merged `acc` so `Options.PerRepo=true` callers can score
			// each repo independently. We populate both side-by-side
			// inside the same for loop so the per-repo numbers are
			// guaranteed to be derived from the same blame/commit data
			// the merged total saw — no second pass, no drift.
			//
			// Entries are keyed by repo name. nil when PerRepo is off
			// so existing CLI callers don't pay for unused work.
			var repoAccs map[string]*accumulator
			if opts.PerRepo {
				repoAccs = make(map[string]*accumulator, len(drepos))
			}

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

				var racc *accumulator
				if opts.PerRepo {
					racc = newAccumulator()
					racc.repoCount = 1
					repoAccs[repo.name] = racc
				}

				// Production
				prod := metric.CalcProduction(periodCommits, cfg.ExcludeFilePatterns)
				mergeMap(acc.raw.Production, prod)
				if racc != nil {
					mergeMap(racc.raw.Production, prod)
				}

				// Lines
				added, deleted := metric.CalcLines(periodCommits, cfg.ExcludeFilePatterns)
				mergeMapInt(acc.raw.LinesAdded, added)
				mergeMapInt(acc.raw.LinesDeleted, deleted)
				if racc != nil {
					mergeMapInt(racc.raw.LinesAdded, added)
					mergeMapInt(racc.raw.LinesDeleted, deleted)
				}

				// Quality
				allCommits := make([]git.Commit, len(periodCommits), len(periodCommits)+len(periodMerges))
				copy(allCommits, periodCommits)
				allCommits = append(allCommits, periodMerges...)
				qual := metric.CalcQuality(allCommits)
				mergeMapAvg(acc.raw.Quality, qual, acc.qualityCounts)
				if racc != nil {
					mergeMapAvg(racc.raw.Quality, qual, racc.qualityCounts)
				}

				// Design
				design := metric.CalcDesign(periodCommits, cfg.ArchitecturePatterns)
				mergeMap(acc.raw.Design, design)
				if racc != nil {
					mergeMap(racc.raw.Design, design)
				}

				// Breadth + date tracking
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

					if racc != nil {
						if _, ok := racc.authorRepoCommits[c.Author]; !ok {
							racc.authorRepoCommits[c.Author] = make(map[string]int)
						}
						racc.authorRepoCommits[c.Author][repo.name]++
						racc.raw.TotalCommits[c.Author]++
						if first, ok := racc.authorFirstDate[c.Author]; !ok || c.Date.Before(first) {
							racc.authorFirstDate[c.Author] = c.Date
						}
						if last, ok := racc.authorLastDate[c.Author]; !ok || c.Date.After(last) {
							racc.authorLastDate[c.Author] = c.Date
						}
					}
				}

				// Blame at period boundary
				var blameVerbose func(string)
				if cb.OnVerbose != nil {
					blameVerbose = cb.OnVerbose
				}

				boundaryCommit, err := git.FindCommitAtDate(ctx, repo.path, window.End)
				if err != nil {
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

				// Pre-skip files larger than cfg.MaxBlameFileBytes at the
				// boundary commit. The boundary tree can include checked-in
				// dumps (huge SQL bulk inserts, generated assets) that
				// aren't in HEAD anymore — those would otherwise deadlock
				// blame and there's no per-file timeout in the boundary
				// path's parent context.
				files, err = git.FilterFilesBySize(ctx, repo.path, boundaryCommit, files, cfg.MaxBlameFileBytes, blameVerbose)
				if err != nil && blameVerbose != nil {
					blameVerbose(fmt.Sprintf("  [blame] size filter error: %v", err))
				}
				if len(files) == 0 {
					continue
				}

				var blameLines []git.BlameLine
				blameCacheKey := cache.BlameAtCommitKey(repo.path, boundaryCommit, files, cfg.SampleSize)
				if cacheStore.Get(blameCacheKey, &blameLines) {
					// cached
				} else {
					var blameProg func(int, int)
					if cb.OnBlameProgress != nil {
						repoName := repo.name
						blameProg = func(done, total int) {
							cb.OnBlameProgress(repoName, done, total)
						}
					}

					blameLines, err = git.ConcurrentBlameFilesAtCommit(ctx, repo.path, boundaryCommit, files, cfg.SampleSize, workers, cfg.BlameTimeout, blameProg, blameVerbose)
					if err != nil {
						// Non-fatal: continue with whatever blame lines we got
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
				pressureMode := opts.PressureMode
				if pressureMode == "" {
					pressureMode = "include"
				}
				if pressureMode == "include" {
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
					if racc != nil {
						mergeMap(racc.raw.Survival, survResult.Decayed)
						mergeMap(racc.raw.RawSurvival, survResult.Raw)
						mergeMap(racc.raw.RobustSurvival, survResult.Robust)
						mergeMap(racc.raw.DormantSurvival, survResult.Dormant)
					}
				} else {
					survResult := metric.CalcSurvival(blameLines, cfg.Tau, window.End)
					mergeMap(acc.raw.Survival, survResult.Decayed)
					mergeMap(acc.raw.RawSurvival, survResult.Raw)
					if racc != nil {
						mergeMap(racc.raw.Survival, survResult.Decayed)
						mergeMap(racc.raw.RawSurvival, survResult.Raw)
					}
				}

				// Indispensability
				indisp, _ := metric.CalcIndispensability(blameLines, cfg.BusFactor.Critical, cfg.BusFactor.High)
				mergeMap(acc.raw.Indispensability, indisp)
				if racc != nil {
					mergeMap(racc.raw.Indispensability, indisp)
				}

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
					if racc != nil {
						mergeMapAvg(racc.raw.DebtCleanup, debt, racc.debtCounts)
					}
				}
			}

			// Breadth
			const minCommitsForBreadth = 3
			for author, repoCommits := range acc.authorRepoCommits {
				count := 0
				for _, commits := range repoCommits {
					if commits >= minCommitsForBreadth {
						count++
					}
				}
				if count > 0 {
					acc.raw.Breadth[author] = float64(count)
				}
			}
			// Per-repo Breadth: each repo's authors contribute to that
			// repo's own Breadth based on commits-in-this-repo only. With a
			// single repo per accumulator the count saturates at 1 (the
			// only repo the author touched). The CLI/SaaS rendering of
			// per-repo Breadth is therefore always {0, 1}, mirroring
			// analyzer.RepoResult.Results[*].Breadth shape.
			if repoAccs != nil {
				for _, racc := range repoAccs {
					for author, repoCommits := range racc.authorRepoCommits {
						count := 0
						for _, commits := range repoCommits {
							if commits >= minCommitsForBreadth {
								count++
							}
						}
						if count > 0 {
							racc.raw.Breadth[author] = float64(count)
						}
					}
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
			if repoAccs != nil {
				for _, racc := range repoAccs {
					for author, total := range racc.raw.Production {
						first := racc.authorFirstDate[author]
						last := racc.authorLastDate[author]
						days := last.Sub(first).Hours() / 24
						if days < 1 {
							days = 1
						}
						racc.raw.Production[author] = total / days
					}
				}
			}

			// Override ActiveDays to cover the full period
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

			// Score per-repo accumulators under the same window/cfg.
			// Iterate in deterministic repo-name order (W-02) so repeated
			// runs of the same input produce byte-identical PerRepo slices.
			var perRepoOut []RepoPeriodResult
			if repoAccs != nil {
				repoNames := make([]string, 0, len(repoAccs))
				for name := range repoAccs {
					repoNames = append(repoNames, name)
				}
				sort.Strings(repoNames)
				for _, name := range repoNames {
					racc := repoAccs[name]
					rscored := scorer.ScoreAt(racc.raw, &periodCfg, racc.authorLastDate, window.End)
					var rfiltered []scorer.Result
					for _, r := range rscored {
						if !cfg.IsExcludedAuthor(r.Author) {
							rfiltered = append(rfiltered, r)
						}
					}
					if len(rfiltered) == 0 {
						continue
					}
					perRepoOut = append(perRepoOut, RepoPeriodResult{
						RepoName: name,
						Domain:   string(d),
						Members:  rfiltered,
					})
				}
			}

			pr := PeriodResult{
				Label:   window.Label,
				Start:   window.Start.Format("2006-01-02"),
				End:     window.End.Format("2006-01-02"),
				Members: filtered,
				PerRepo: perRepoOut,
			}
			windowDomainResults[string(d)] = pr
			domainPeriodMap[string(d)] = append(domainPeriodMap[string(d)], pr)
		}

		if cb.OnPeriodComplete != nil && len(windowDomainResults) > 0 {
			cb.OnPeriodComplete(windowDomainResults)
		}
	}

	var results []DomainTimeline
	for _, d := range allDomains {
		periods := domainPeriodMap[string(d)]
		if len(periods) > 0 {
			results = append(results, DomainTimeline{
				Domain:  string(d),
				Periods: periods,
			})
		}
	}

	return results, nil
}

// BuildTeamPeriodResults aggregates per-period scored results into TeamPeriodResults.
// Returns map[teamName][]TeamPeriodResult.
func BuildTeamPeriodResults(d string, periods []PeriodResult, cfg *config.Config) map[string][]timeline.TeamPeriodResult {
	result := make(map[string][]timeline.TeamPeriodResult)

	if len(cfg.Teams) > 0 {
		for teamName, entry := range cfg.Teams {
			if !strings.EqualFold(entry.Domain, d) {
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
		teamName := d
		for _, p := range periods {
			tr := team.Aggregate(teamName, d, 0, p.Members, nil)
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

// --- internal helpers (same as pkg/analyzer) ---

type accumulator struct {
	raw               *metric.RawScores
	qualityCounts     map[string]int
	debtCounts        map[string]int
	authorRepoCommits map[string]map[string]int
	authorFirstDate   map[string]time.Time
	authorLastDate    map[string]time.Time
	repoCount         int
	changePressure    metric.ChangePressure
}

func newAccumulator() *accumulator {
	return &accumulator{
		raw:               metric.NewRawScores(),
		qualityCounts:     make(map[string]int),
		debtCounts:        make(map[string]int),
		authorRepoCommits: make(map[string]map[string]int),
		authorFirstDate:   make(map[string]time.Time),
		authorLastDate:    make(map[string]time.Time),
		changePressure:    make(metric.ChangePressure),
	}
}

func resolveRepoDomain(ctx context.Context, repoPath, repoName string, cfg *config.Config, extMap map[string]domain.Domain) domain.Domain {
	for name, entry := range cfg.Domains {
		if len(entry.Repos) > 0 && domain.MatchRepoPattern(repoName, entry.Repos) {
			return domain.NormalizeName(name)
		}
	}
	files, err := git.ListAllFiles(ctx, repoPath)
	if err != nil || len(files) == 0 {
		return domain.Unknown
	}
	return domain.DetectFromFiles(files, extMap)
}

func filterCommits(commits []git.Commit, cfg *config.Config) []git.Commit {
	var r []git.Commit
	for _, c := range commits {
		c.Author = cfg.ResolveAuthor(c.Author)
		if !cfg.IsExcludedAuthor(c.Author) {
			r = append(r, c)
		}
	}
	return r
}

func filterFileStats(commits []git.Commit, patterns []string) []git.Commit {
	if len(patterns) == 0 {
		return commits
	}
	for i := range commits {
		var f []git.FileStat
		for _, fs := range commits[i].FileStats {
			if !metric.IsExcluded(fs.Filename, patterns) {
				f = append(f, fs)
			}
		}
		commits[i].FileStats = f
	}
	return commits
}

func filterFiles(files []string, patterns []string) []string {
	if len(patterns) == 0 {
		return files
	}
	var r []string
	for _, f := range files {
		if !metric.IsExcluded(f, patterns) {
			r = append(r, f)
		}
	}
	return r
}

func filterBlameLines(lines []git.BlameLine, cfg *config.Config) []git.BlameLine {
	var r []git.BlameLine
	for _, bl := range lines {
		if !cfg.IsExcludedAuthor(bl.Author) {
			r = append(r, bl)
		}
	}
	return r
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

func mergeMapAvg(dst, src map[string]float64, counts map[string]int) {
	for k, v := range src {
		n := counts[k]
		if n > 0 {
			dst[k] = (dst[k]*float64(n) + v) / float64(n+1)
		} else {
			dst[k] = v
		}
		counts[k] = n + 1
	}
}
