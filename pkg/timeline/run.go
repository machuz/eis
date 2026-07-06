// Package timeline provides a library-friendly timeline analysis pipeline.
// This is the public API for timeline analysis without CLI dependencies
// (no spinner, color, cache, flag parsing).
package timeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
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
	// PeriodConcurrency bounds how many timeline windows are computed in
	// parallel. Each window is fully independent — it re-filters the shared
	// pre-parsed commits, resolves its own boundary commit at window.End,
	// runs its own blame, and scores against window.End — so parallelism
	// cannot change any value (W-02 determinism holds regardless).
	//
	//   <= 1 → sequential, exactly the historical behavior (DEFAULT). Existing
	//          callers and the CLI stay bit-identical.
	//   >  1 → up to this many windows computed concurrently.
	//
	// The OnPeriodComplete / OnPeriodStart callbacks still fire in strict
	// window order even under concurrency (see the ordered emitter in Run),
	// so streaming consumers (Ace) need no change. Memory caveat: each window
	// runs its own blame worker pool, so peak blame memory scales roughly as
	// PeriodConcurrency × Workers × (hundreds of MB on a large repo). Callers
	// must bound the two jointly.
	PeriodConcurrency int
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

// incrementalBlame reports whether the per-file last-touch blame cache
// (cache.BlameFileAtCommitKey) is sound for this move-detection level. Only
// within-file detection (-M — the default and what observation uses) keeps a
// file's blame a pure function of its own history; -C/-C -C attribute lines to
// copy sources in other files, so a file's blame then depends on other files'
// commits and the per-file key would unify blames that differ. Those levels use
// the whole-boundary key instead.
// autoPeriodConcurrency picks the default number of timeline windows to compute
// in parallel when PeriodConcurrency is unset (<= 0). The cold-run cost is
// concentrated in a handful of HEAVY windows (large boundary blames), and each
// window already fans its blame across `Workers` threads. Measurement showed the
// sweet spot is modest window parallelism WITH full per-window workers —
// overlapping ~4 heavy windows captures most of the win, while pushing window
// parallelism higher and stealing workers per window is slower (heavy windows get
// throttled). Peak blame memory ≈ this × Workers × per-repo blame, so it is
// clamped low (≤ 4) to stay safe for SaaS/ingest runners.
func autoPeriodConcurrency() int {
	n := runtime.NumCPU() / 3
	if n < 1 {
		n = 1
	}
	if n > 4 {
		n = 4
	}
	return n
}

func incrementalBlame(moveDetection string) bool {
	return moveDetection == "" || moveDetection == "file"
}

// lastTouchByFile maps each file path to the SHA of the newest commit (by date)
// that modified it within commits — which the caller passes pre-filtered to a
// window boundary (cumulativeCommits, all commits <= window.End). That SHA is
// the per-file cache coordinate for incremental blame: a file unchanged since
// its last-touch commit has identical -M blame at every later boundary, so the
// last-touch SHA — not the moving window boundary — is the stable key.
//
// "Newest by date" is consistent with the timeline's own date-based windowing
// (windows and cumulativeCommits are themselves date-filtered), so it introduces
// no ordering assumption the model doesn't already make. On equal dates the
// later position in commits wins, which is deterministic for a fixed parsed log.
func lastTouchByFile(commits []git.Commit) map[string]string {
	type touch struct {
		sha  string
		date time.Time
	}
	latest := make(map[string]touch)
	for _, c := range commits {
		for _, fs := range c.FileStats {
			if prev, ok := latest[fs.Filename]; !ok || !c.Date.Before(prev.date) {
				latest[fs.Filename] = touch{c.Hash, c.Date}
			}
		}
	}
	out := make(map[string]string, len(latest))
	for f, t := range latest {
		out[f] = t.sha
	}
	return out
}

// Run executes the timeline analysis pipeline on the given repository paths.
// It returns per-domain timeline results without any CLI-specific behavior.
// Run walks the timeline for repoPaths and returns per-domain period results.
//
// ctx bounds the walk: every git operation (log parse, boundary resolution,
// blame) runs under it, and the window scheduler stops dispatching new windows
// once ctx is cancelled. A caller that time-boxes a long backfill (Ace splits a
// years-long history into bounded jobs) passes a ctx it cancels at the budget;
// the walk then stops MID-window instead of running to the present. Cancellation
// yields a PARTIAL result (the windows completed before the deadline) with a nil
// error — the caller detects the truncation via context.Cause and resumes the
// tail. Pass context.Background() for an uninterrupted, historical-behavior walk
// (the CLI does); that path is bit-identical to before ctx threading (W-02).
func Run(ctx context.Context, opts Options, repoPaths []string, cfg *config.Config, cb *Callbacks) ([]DomainTimeline, error) {
	if ctx == nil {
		ctx = context.Background()
	}
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

	workers := opts.Workers
	if workers == 0 {
		workers = 4
	}

	// Module resolvers are built per repo (see the inner loop) so each
	// repo can honor its own RepoOverrides. Per-repo resolution remains
	// deterministic (W-02/W-03) because it depends only on the resolved
	// pattern list.

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

	// now drives ONLY default-window/period selection (which time windows to
	// analyze) — it is an INPUT, not a decay/classification reference. Each
	// period scores against its own window.End (see CalcSurvival*/ScoreAt
	// below), so wall-clock here never leaks into a score (W-02).
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
		idmap   map[string]string // email-derived identity map, for per-period blame
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
		logCacheKey := cache.LogKey(repoPath, headHash, cfg.CommentFilterEnabled())
		if headHash != "" && cacheStore.Get(logCacheKey, &commits) {
			// cached
		} else {
			git.ConfigureBotCoAuthorPatterns(cfg.BotCoAuthorPatterns)
			commits, err = git.ParseLogParallel(ctx, repoPath, workers, cfg.CommentFilterEnabled())
			if err != nil {
				return nil, fmt.Errorf("parse log %s: %w", repoName, err)
			}
			if headHash != "" {
				cacheStore.Set(logCacheKey, commits)
			}
		}
		// Collapse split identities (one person, several names under one email)
		// before alias resolution; reused to remap per-period blame authors.
		idmap := git.BuildIdentityMap(commits)
		git.CanonicalizeAuthors(commits, nil, idmap)
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
			idmap:   idmap,
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

	// Blame move/copy detection is a cross-cutting git-invocation policy
	// (a package global in internal/git). Set it ONCE here, before any
	// window is computed, rather than inside the per-window body: under
	// PeriodConcurrency > 1, concurrent identical writes to that global
	// would trip the race detector even though the value never differs.
	git.ConfigureBlameMoveDetection(cfg.BlameMoveDetection)

	// computeWindow performs the full, independent work for a single
	// timeline window: filter the shared pre-parsed commits to the window,
	// resolve the boundary commit at window.End, blame, score every domain,
	// and return the per-domain results keyed by domain name. It reads only
	// shared read-only state (repos, cfg, cacheStore) plus its own window,
	// so it is safe to run concurrently for distinct windows — every value
	// is pinned to window.End, so ordering never changes a result (W-02).
	computeWindow := func(window TimeWindow) map[string]PeriodResult {
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
				// Per-repo module resolver: honors RepoOverrides keyed by
				// repo.name (the same identifier the YAML uses).
				moduleResolver := metric.NewModuleResolverWithExcludes(config.PatternsForRepo(cfg, repo.name), config.ExcludesForRepo(cfg, repo.name))

				// Filter commits to this period. cumulativeCommits holds ALL
				// commits up to the period boundary — Catalysis precedence
				// (who first touched a file before whom) must be judged against
				// the full history, not just this window, or recent narrow
				// periods have no precedence structure and collapse to ~0.
				var periodCommits, cumulativeCommits []git.Commit
				for _, c := range repo.commits {
					if c.Date.Before(window.End) {
						cumulativeCommits = append(cumulativeCommits, c)
						if !c.Date.Before(window.Start) {
							periodCommits = append(periodCommits, c)
						}
					}
				}

				var periodMerges []git.Commit
				for _, c := range repo.merges {
					if !c.Date.Before(window.Start) && c.Date.Before(window.End) {
						periodMerges = append(periodMerges, c)
					}
				}

				// Module liveness gate (ADR step 2): fold fallback-derived
				// modules not touched in >= ModuleLivenessMinMonths distinct
				// calendar months WITHIN this window into
				// metric.PeripheralModule. Derived from commit.Date over the
				// in-scope periodCommits (deterministic, W-02); applied before
				// any module-topology metric below.
				fold := metric.ComputeModuleFold(periodCommits, moduleResolver, cfg.ModuleLivenessMinMonths)
				moduleResolver = moduleResolver.WithFold(fold)

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

				// Catalysis is computed in the blame stage below (needs the
				// period commit history for file origin + the period blame lines).

				// Design is computed in the blame stage below (survival-weighted: it
				// scores SURVIVING arch lines from the blame, not arch lines changed,
				// so it needs the period blame + the window boundary as the time basis).

				// Breadth + date tracking
				for _, c := range periodCommits {
					if _, ok := acc.authorRepoCommits[c.Author]; !ok {
						acc.authorRepoCommits[c.Author] = make(map[string]int)
					}
					acc.authorRepoCommits[c.Author][repo.name]++
					// Per-module commit counts feed module-unit Breadth.
					// A commit counts once per distinct module it touches.
					if _, ok := acc.authorModuleCommits[c.Author]; !ok {
						acc.authorModuleCommits[c.Author] = make(map[string]int)
					}
					touchedModules := make(map[string]bool)
					for _, fs := range c.FileStats {
						if mod := moduleResolver.ModuleOf(fs.Filename); mod != "" {
							touchedModules[mod] = true
						}
					}
					for mod := range touchedModules {
						acc.authorModuleCommits[c.Author][mod]++
					}
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
						if _, ok := racc.authorModuleCommits[c.Author]; !ok {
							racc.authorModuleCommits[c.Author] = make(map[string]int)
						}
						for mod := range touchedModules {
							racc.authorModuleCommits[c.Author][mod]++
						}
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

				var blameProg func(int, int)
				if cb.OnBlameProgress != nil {
					repoName := repo.name
					blameProg = func(done, total int) {
						cb.OnBlameProgress(repoName, done, total)
					}
				}

				var blameLines []git.BlameLine
				if incrementalBlame(cfg.BlameMoveDetection) {
					// Incremental per-file blame cache (see
					// cache.BlameFileAtCommitKey). Under -M a file's blame is a
					// pure function of its own history, so blaming it here equals
					// blaming it at the commit that last touched it — key the
					// cache on that last-touch SHA and files unchanged since an
					// earlier window (or run) are reused; only files modified in
					// this window are re-blamed. This collapses the per-window
					// blame from O(all files) to O(files changed in the window),
					// which is the dominant cost of a timeline backfill.
					//
					// Sample first so the CLI's default file sampling is
					// preserved exactly (observation sets SampleSize=all, so its
					// set is unchanged); cache/reblame operate on that subset.
					sampled := git.SampleFiles(files, cfg.SampleSize)
					touch := lastTouchByFile(cumulativeCommits)
					var cold []string
					for _, f := range sampled {
						th, ok := touch[f]
						if !ok {
							// No recorded touch (file present in the tree but not
							// in the parsed/sampled log window) — blame fresh and
							// don't cache, since we have no stable key for it.
							cold = append(cold, f)
							continue
						}
						var fl []git.BlameLine
						if cacheStore.Get(cache.BlameFileAtCommitKey(repo.path, f, th, cfg.SampleSize, cfg.BlameMoveDetection), &fl) {
							blameLines = append(blameLines, fl...)
						} else {
							cold = append(cold, f)
						}
					}
					if len(cold) > 0 {
						// cold is already the sampled subset; blame per input path
						// (grouped) so each file's lines cache under its own key
						// regardless of git's historical Filename reporting.
						fresh := git.ConcurrentBlameFilesAtCommitByFile(ctx, repo.path, boundaryCommit, cold, workers, cfg.BlameTimeout, blameProg, blameVerbose)
						for _, f := range cold {
							fl := fresh[f]
							if len(fl) == 0 {
								continue
							}
							if th, ok := touch[f]; ok {
								cacheStore.Set(cache.BlameFileAtCommitKey(repo.path, f, th, cfg.SampleSize, cfg.BlameMoveDetection), fl)
							}
							blameLines = append(blameLines, fl...)
						}
					}
				} else {
					// -C / -C -C: a line can be attributed to a copy source in
					// another file, so a file's blame depends on other files'
					// commits and the per-file last-touch key is unsound. Fall
					// back to the whole-boundary key (immutable at a fixed commit).
					blameCacheKey := cache.BlameAtCommitKey(repo.path, boundaryCommit, files, cfg.SampleSize, cfg.BlameMoveDetection)
					if !cacheStore.Get(blameCacheKey, &blameLines) {
						blameLines, err = git.ConcurrentBlameFilesAtCommit(ctx, repo.path, boundaryCommit, files, cfg.SampleSize, workers, cfg.BlameTimeout, blameProg, blameVerbose)
						if err != nil {
							// Non-fatal: continue with whatever blame lines we got
						}
						if len(blameLines) > 0 {
							cacheStore.Set(blameCacheKey, blameLines)
						}
					}
				}

				// Collapse split identities, then apply aliases
				git.CanonicalizeAuthors(nil, blameLines, repo.idmap)
				coMap := metric.CoAuthorMap(repo.commits)
				for i := range blameLines {
					blameLines[i].Author = cfg.ResolveAuthor(blameLines[i].Author)
					if set, ok := coMap[blameLines[i].Commit]; ok {
						blameLines[i].Authors = set
					}
				}
				// Drop lines a `git subtree --squash` import collapsed onto the
				// integrator. Blame here is at the period boundary, so scan
				// squashes reachable from boundaryCommit — a squash that lands
				// after this window must not retroactively suppress it.
				if !cfg.AttributeSubtreeSquash {
					if squash, sqErr := git.SubtreeSquashCommits(ctx, repo.path, boundaryCommit); sqErr == nil {
						if kept, dropped := git.DropSubtreeSquashBlame(blameLines, squash); dropped > 0 {
							blameLines = kept
							if blameVerbose != nil {
								blameVerbose(fmt.Sprintf("  [blame] subtree-squash: suppressed %d imported line(s) from %d import commit(s)", dropped, len(squash)))
							}
						}
					}
				}
				blameLines = filterBlameLines(blameLines, cfg)

				// Catalysis: surviving mass others built on this author's still-
				// living foundation, as of the period boundary. Precedence uses
				// cumulativeCommits (full history up to window.End), not just the
				// window; survival uses the period-boundary blame. Decay ref =
				// window.End (matches Survival).
				catalysis := metric.CalcCatalysis(cumulativeCommits, blameLines, cfg.TauForDomain(string(repo.domain)), window.End)
				mergeMap(acc.raw.Catalysis, catalysis)
				if racc != nil {
					mergeMap(racc.raw.Catalysis, catalysis)
				}

				// Design (survival-weighted): surviving, time-decayed arch-file blame
				// lines per author, on the same window boundary + tau the survival stage
				// uses. Churned-away arch edits have decayed out of the blame, so design
				// tracks durable structural ownership, not arch churn volume.
				design := metric.CalcDesignSurviving(blameLines, cfg.ArchitecturePatterns, cfg.Tau, window.End)
				mergeMap(acc.raw.Design, design)
				if racc != nil {
					mergeMap(racc.raw.Design, design)
				}

				// Survival
				pressureMode := opts.PressureMode
				if pressureMode == "" {
					pressureMode = "include"
				}
				if pressureMode == "include" {
					repoPressure := metric.CalcChangePressure(periodCommits, blameLines, moduleResolver)
					for mod, p := range repoPressure {
						key := repo.name + "/" + mod
						acc.changePressure[key] = p
					}

					// blameLines authors are already alias-resolved in place above.
					blameByAuthor := make(map[string]int)
					for _, bl := range blameLines {
						blameByAuthor[bl.Author]++
					}
					pressureThreshold := metric.PressureThreshold(repoPressure, blameByAuthor, metric.SubstantialAuthorLines)
					periodOthers := metric.CalcOthersPressure(periodCommits, blameLines, moduleResolver)
					survResult := metric.CalcSurvivalWithPressure(blameLines, cfg.TauForDomain(string(repo.domain)), window.End, repoPressure, pressureThreshold, moduleResolver, periodOthers)
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
					survResult := metric.CalcSurvival(blameLines, cfg.TauForDomain(string(repo.domain)), window.End)
					mergeMap(acc.raw.Survival, survResult.Decayed)
					mergeMap(acc.raw.RawSurvival, survResult.Raw)
					if racc != nil {
						mergeMap(racc.raw.Survival, survResult.Decayed)
						mergeMap(racc.raw.RawSurvival, survResult.Raw)
					}
				}

				// Per-(module, author) surviving gravity for Breadth (Hill
				// number). Independent of the pressure split above; uses this
				// repo's resolver and the period boundary as the time basis.
				msba := metric.CalcModuleSurvivalByAuthor(blameLines, cfg.TauForDomain(string(repo.domain)), window.End, moduleResolver)
				for mod, authors := range msba {
					dst := acc.authorModuleSurvival[mod]
					if dst == nil {
						dst = make(map[string]float64)
						acc.authorModuleSurvival[mod] = dst
					}
					var rdst map[string]float64
					if racc != nil {
						rdst = racc.authorModuleSurvival[mod]
						if rdst == nil {
							rdst = make(map[string]float64)
							racc.authorModuleSurvival[mod] = rdst
						}
					}
					for author, mass := range authors {
						dst[author] += mass
						if rdst != nil {
							rdst[author] += mass
						}
					}
				}

				// Indispensability
				indisp, _ := metric.CalcIndispensability(blameLines, moduleResolver, cfg.BusFactor.Critical, cfg.BusFactor.High)
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
						debt, _ = metric.CalcDebt(ctx, repo.path, fixCommits, 50, cfg.DebtThreshold, cfg.BlameTimeout, workers, cfg.ResolveAuthor, metric.CoAuthorMap(repo.commits), nil, nil)
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

			// Breadth = effective number of modules an author holds surviving
			// gravity in (Hill number over per-module gravity), structure-
			// neutral and survival-weighted — same shared helper as the
			// analyzer pipeline so the two can't drift (W-02).
			breadth := metric.ComputeBreadth(acc.authorModuleSurvival)
			for author, b := range breadth {
				acc.raw.Breadth[author] = b
			}
			// Per-repo Breadth = effective modules the author reached inside
			// that one repo (Hill number over this repo's per-module gravity).
			if repoAccs != nil {
				for _, racc := range repoAccs {
					rb := metric.ComputeBreadth(racc.authorModuleSurvival)
					for author, b := range rb {
						racc.raw.Breadth[author] = b
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
		}

		return windowDomainResults
	}

	// outputs[pi] holds window pi's per-domain results, written into a
	// per-window INDEXED slot (never a shared appended map) so concurrent
	// windows don't contend on a mutex on the hot path and the final
	// assembled order is deterministic regardless of completion order.
	outputs := make([]map[string]PeriodResult, len(windows))

	// PeriodConcurrency <= 0 means "auto": pick a good default here so every
	// caller — CLI, SaaS, and the ingest jobs — gets window parallelism without
	// having to specify it. Windows run sequentially at 1; the cold-run cost is
	// concentrated in a handful of heavy windows (large boundary blames), and
	// overlapping those is the biggest remaining lever.
	concurrency := opts.PeriodConcurrency
	if concurrency <= 0 {
		concurrency = autoPeriodConcurrency()
	}
	if concurrency > len(windows) {
		concurrency = len(windows)
	}

	if concurrency <= 1 {
		// Sequential: bit-identical to the historical behavior. OnPeriodStart
		// fires BEFORE the window is computed (progress UI depends on that),
		// then OnPeriodComplete after.
		for pi, window := range windows {
			// Stop dispatching new windows once the caller's budget cancels ctx.
			// Windows already emitted stay; the remainder is the caller's to resume.
			if ctx.Err() != nil {
				break
			}
			if cb.OnPeriodStart != nil {
				cb.OnPeriodStart(window.Label, pi, len(windows))
			}
			outputs[pi] = computeWindow(window)
			if cb.OnPeriodComplete != nil && len(outputs[pi]) > 0 {
				cb.OnPeriodComplete(outputs[pi])
			}
		}
	} else {
		// Concurrent computation, ORDERED emission. A bounded worker pool
		// computes windows in whatever order they finish, but an emit cursor
		// (nextEmit) guarded by mu only ever fires OnPeriodStart/OnPeriod
		// Complete for window nextEmit once it AND all earlier windows have
		// landed. Net effect: the callback sequence is identical to sequential
		// mode, so ordered-emission consumers (Ace's dedup) need no change.
		//
		// Note: OnBlameProgress / OnVerbose are best-effort progress hooks and
		// may be invoked concurrently from multiple window goroutines under
		// PeriodConcurrency > 1; the ordered guarantee covers only the paired
		// (OnPeriodStart, OnPeriodComplete) sequence.
		var mu sync.Mutex
		done := make([]bool, len(windows))
		nextEmit := 0
		emitReady := func() {
			for nextEmit < len(windows) && done[nextEmit] {
				if cb.OnPeriodStart != nil {
					cb.OnPeriodStart(windows[nextEmit].Label, nextEmit, len(windows))
				}
				if cb.OnPeriodComplete != nil && len(outputs[nextEmit]) > 0 {
					cb.OnPeriodComplete(outputs[nextEmit])
				}
				nextEmit++
			}
		}

		idxCh := make(chan int)
		var wg sync.WaitGroup
		for w := 0; w < concurrency; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for pi := range idxCh {
					res := computeWindow(windows[pi])
					mu.Lock()
					outputs[pi] = res
					done[pi] = true
					emitReady()
					mu.Unlock()
				}
			}()
		}
		for pi := range windows {
			// Stop feeding new windows once the caller's budget cancels ctx. In-flight
			// windows drain fast (their git ops honor ctx); windows never fed leave
			// done[pi]=false, so the ordered emitter halts at the first unfed window —
			// exactly the truncation frontier the caller resumes from.
			if ctx.Err() != nil {
				break
			}
			idxCh <- pi
		}
		close(idxCh)
		wg.Wait()
	}

	// Assemble domainPeriodMap from the indexed slots IN WINDOW ORDER, so the
	// returned []DomainTimeline is byte-identical to the sequential build:
	// every domain present in a window's result is appended in window order.
	domainPeriodMap := make(map[string][]PeriodResult, len(allDomains))
	for pi := range windows {
		res := outputs[pi]
		for _, d := range allDomains {
			if pr, ok := res[string(d)]; ok {
				domainPeriodMap[string(d)] = append(domainPeriodMap[string(d)], pr)
			}
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
	raw                 *metric.RawScores
	debtCounts          map[string]int
	authorRepoCommits   map[string]map[string]int
	authorModuleCommits map[string]map[string]int
	// authorModuleSurvival accumulates per-(module, author) surviving gravity
	// mass across repos in this period — the input to the survival-weighted,
	// structure-neutral Breadth (Hill number). Keyed module → author → mass.
	authorModuleSurvival map[string]map[string]float64
	authorFirstDate      map[string]time.Time
	authorLastDate       map[string]time.Time
	repoCount            int
	changePressure       metric.ChangePressure
}

func newAccumulator() *accumulator {
	return &accumulator{
		raw:                  metric.NewRawScores(),
		debtCounts:           make(map[string]int),
		authorRepoCommits:    make(map[string]map[string]int),
		authorModuleCommits:  make(map[string]map[string]int),
		authorModuleSurvival: make(map[string]map[string]float64),
		authorFirstDate:      make(map[string]time.Time),
		authorLastDate:       make(map[string]time.Time),
		changePressure:       make(metric.ChangePressure),
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
		if cfg.IsExcludedAuthor(c.Author) {
			continue
		}
		c.CoAuthors = resolveCoAuthors(c.CoAuthors, c.Author, cfg)
		r = append(r, c)
	}
	return r
}

// resolveCoAuthors aliases each Co-authored-by name, drops excluded authors and
// any that coincide with the primary. Mirrors the internal/cli helper.
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

func filterFileStats(commits []git.Commit, patterns []string) []git.Commit {
	if len(patterns) == 0 {
		return commits
	}
	// Memoize by filename: a path recurs across thousands of commits and the glob
	// match is otherwise re-run every time (see metric.Excluder).
	excl := metric.NewExcluder(patterns)
	for i := range commits {
		var f []git.FileStat
		for _, fs := range commits[i].FileStats {
			if !excl.IsExcluded(fs.Filename) {
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
