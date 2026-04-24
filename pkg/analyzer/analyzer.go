// Package analyzer provides the core analysis pipeline for external use.
// This is a library-friendly version without CLI dependencies (no spinner, color, cache).
package analyzer

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/machuz/eis/v2/internal/cache"
	"github.com/machuz/eis/v2/internal/config"
	"github.com/machuz/eis/v2/internal/domain"
	"github.com/machuz/eis/v2/internal/git"
	"github.com/machuz/eis/v2/internal/metric"
	"github.com/machuz/eis/v2/internal/scorer"
)

// Options controls the analysis pipeline behavior.
type Options struct {
	Tau          float64
	SampleSize   int
	Workers      int
	PressureMode string // "include" or "ignore"
	ActiveDays   int
	DomainFilter string
	PerRepo      bool
	CacheEnabled bool
	CacheDir     string // custom cache dir (empty = default ~/.eis/cache)
}

// DomainResults holds scored results for a single domain.
type DomainResults struct {
	Domain    domain.Domain
	Results   []scorer.Result
	Risks     []metric.ModuleRisk
	RepoCount int
	PerRepo   []RepoResult

	// Module Topology (3-axis)
	Cochange       metric.CochangeResult
	Ownership      []metric.ModuleOwnership
	ModuleSurvival map[string]float64 // module → survival rate (0-1)
}

// RepoResult holds scored results for a single repository.
type RepoResult struct {
	RepoName       string
	Domain         domain.Domain
	Results        []scorer.Result
	Cochange       metric.CochangeResult
	Ownership      []metric.ModuleOwnership
	ModuleSurvival map[string]float64
}

// ProgressFunc is called with (done, total) during long-running operations.
type ProgressFunc func(done, total int)

// Callbacks allows callers to hook into pipeline progress.
type Callbacks struct {
	OnRepoStart     func(repoName string, d domain.Domain)
	OnRepoSkipped   func(repoName, reason string)
	OnBlameProgress ProgressFunc
	OnDebtProgress  ProgressFunc
	OnVerbose       func(msg string)
}

// Run executes the full analysis pipeline on the given repository paths.
func Run(opts Options, repoPaths []string, cfg *config.Config, cb *Callbacks) ([]DomainResults, error) {
	if cb == nil {
		cb = &Callbacks{}
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

	ctx := context.Background()
	start := time.Now()
	workers := opts.Workers
	if workers == 0 {
		workers = 4
	}

	// Initialize cache store
	cacheStore := cache.NewWithDir(opts.CacheEnabled, opts.CacheDir)

	type accumulator struct {
		raw               *metric.RawScores
		qualityCounts     map[string]int
		debtCounts        map[string]int
		authorRepoCommits map[string]map[string]int
		authorFirstDate   map[string]time.Time
		authorLastDate    map[string]time.Time
		repoCount         int
		risks             []metric.ModuleRisk
		changePressure    metric.ChangePressure
		// Module Topology accumulators
		allCommits  []git.Commit
		allBlameLines []git.BlameLine
	}
	newAcc := func() *accumulator {
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
	type repoAccState struct {
		acc             *accumulator
		repoName        string
		domain          domain.Domain
		authorFirstDate map[string]time.Time
		authorLastDate  map[string]time.Time
		commits         []git.Commit
		blameLines      []git.BlameLine
	}

	accumulators := make(map[domain.Domain]*accumulator)
	var repoAccumulators []repoAccState

	// Deduplicate
	seen := make(map[string]bool)
	var deduped []string
	for _, p := range repoPaths {
		real, err := filepath.EvalSymlinks(p)
		if err != nil {
			real = p
		}
		if !seen[real] {
			seen[real] = true
			deduped = append(deduped, p)
		}
	}
	repoPaths = deduped

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
		repoDomain := ResolveRepoDomain(ctx, repoPath, repoName, cfg)
		if opts.DomainFilter != "" && !strings.EqualFold(string(repoDomain), opts.DomainFilter) {
			continue
		}
		if cb.OnRepoStart != nil {
			cb.OnRepoStart(repoName, repoDomain)
		}

		acc, ok := accumulators[repoDomain]
		if !ok {
			acc = newAcc()
			accumulators[repoDomain] = acc
		}
		acc.repoCount++

		// Get HEAD hash for cache keys
		headHash, _ := git.HeadHash(ctx, repoPath)

		// Parse git log (cached)
		var commits []git.Commit
		logCacheKey := cache.LogKey(repoPath, headHash)
		if headHash != "" && cacheStore.Get(logCacheKey, &commits) {
			// cache hit
		} else {
			var err error
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

		// Parse merge commits (cached)
		var mergeCommits []git.Commit
		mergeCacheKey := cache.MergeLogKey(repoPath, headHash)
		if headHash != "" && cacheStore.Get(mergeCacheKey, &mergeCommits) {
			// cache hit
		} else {
			mergeCommits, _ = git.ParseMergeCommits(ctx, repoPath)
			if headHash != "" {
				cacheStore.Set(mergeCacheKey, mergeCommits)
			}
		}
		mergeCommits = filterCommits(mergeCommits, cfg)

		prod := metric.CalcProduction(commits, cfg.ExcludeFilePatterns)
		mergeMap(acc.raw.Production, prod)

		added, deleted := metric.CalcLines(commits, cfg.ExcludeFilePatterns)
		mergeMapInt(acc.raw.LinesAdded, added)
		mergeMapInt(acc.raw.LinesDeleted, deleted)

		allCommits := make([]git.Commit, len(commits), len(commits)+len(mergeCommits))
		copy(allCommits, commits)
		allCommits = append(allCommits, mergeCommits...)
		qual := metric.CalcQuality(allCommits)
		mergeMapAvg(acc.raw.Quality, qual, acc.qualityCounts)

		design := metric.CalcDesign(commits, cfg.ArchitecturePatterns)
		mergeMap(acc.raw.Design, design)

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
		for _, c := range mergeCommits {
			if first, ok := acc.authorFirstDate[c.Author]; !ok || c.Date.Before(first) {
				acc.authorFirstDate[c.Author] = c.Date
			}
			if last, ok := acc.authorLastDate[c.Author]; !ok || c.Date.After(last) {
				acc.authorLastDate[c.Author] = c.Date
			}
		}

		files, err := git.ListFiles(ctx, repoPath, cfg.BlameExtensions)
		if err != nil {
			continue
		}
		files = filterFiles(files, cfg.ExcludeFilePatterns)

		var blameVerbose func(string)
		if cb.OnVerbose != nil {
			blameVerbose = cb.OnVerbose
		}

		// Blame analysis (cached)
		var blameLines []git.BlameLine
		blameCacheKey := cache.BlameKey(repoPath, headHash, files, cfg.SampleSize)
		if headHash != "" && cacheStore.Get(blameCacheKey, &blameLines) {
			// cache hit
		} else {
			blameLines, _ = git.ConcurrentBlameFiles(ctx, repoPath, files, cfg.SampleSize, workers, cb.OnBlameProgress, blameVerbose)
			if headHash != "" && len(blameLines) > 0 {
				cacheStore.Set(blameCacheKey, blameLines)
			}
		}
		for i := range blameLines {
			blameLines[i].Author = cfg.ResolveAuthor(blameLines[i].Author)
		}
		blameLines = filterBlameLines(blameLines, cfg)

		// Accumulate for Module Topology
		acc.allCommits = append(acc.allCommits, commits...)
		acc.allBlameLines = append(acc.allBlameLines, blameLines...)

		var repoSurvDecayed, repoSurvRaw, repoSurvRobust, repoSurvDormant map[string]float64
		if opts.PressureMode != "ignore" {
			repoPressure := metric.CalcChangePressure(commits, blameLines)
			for mod, p := range repoPressure {
				acc.changePressure[repoName+"/"+mod] = p
			}
			blameByAuthor := make(map[string]int)
			for _, bl := range blameLines {
				blameByAuthor[cfg.ResolveAuthor(bl.Author)]++
			}
			minShare := float64(len(blameLines)) * 0.10
			substantial := 0
			for _, count := range blameByAuthor {
				if float64(count) >= minShare && count >= 1000 {
					substantial++
				}
			}
			threshold := repoPressure.MedianPressure()
			if substantial < 2 {
				threshold = math.Inf(1)
			}
			sr := metric.CalcSurvivalWithPressure(blameLines, cfg.Tau, start, repoPressure, threshold)
			repoSurvDecayed, repoSurvRaw, repoSurvRobust, repoSurvDormant = sr.Decayed, sr.Raw, sr.Robust, sr.Dormant
			mergeMap(acc.raw.Survival, repoSurvDecayed)
			mergeMap(acc.raw.RawSurvival, repoSurvRaw)
			mergeMap(acc.raw.RobustSurvival, repoSurvRobust)
			mergeMap(acc.raw.DormantSurvival, repoSurvDormant)
		} else {
			sr := metric.CalcSurvival(blameLines, cfg.Tau, start)
			repoSurvDecayed, repoSurvRaw = sr.Decayed, sr.Raw
			mergeMap(acc.raw.Survival, repoSurvDecayed)
			mergeMap(acc.raw.RawSurvival, repoSurvRaw)
		}

		indisp, risks := metric.CalcIndispensability(blameLines, cfg.BusFactor.Critical, cfg.BusFactor.High)
		mergeMap(acc.raw.Indispensability, indisp)
		acc.risks = append(acc.risks, risks...)

		fixCommits := metric.GetFixCommits(commits)
		var debtVerbose metric.VerboseFunc
		if cb.OnVerbose != nil {
			debtVerbose = cb.OnVerbose
		}
		var debtProg metric.ProgressFunc
	if cb.OnDebtProgress != nil {
		debtProg = metric.ProgressFunc(cb.OnDebtProgress)
	}
	debt, _ := metric.CalcDebt(ctx, repoPath, fixCommits, 50, cfg.DebtThreshold, cfg.BlameTimeout, cfg.ResolveAuthor, debtProg, debtVerbose)
		mergeMapAvg(acc.raw.DebtCleanup, debt, acc.debtCounts)

		if opts.PerRepo {
			rr := metric.NewRawScores()
			mergeMap(rr.Production, prod)
			mergeMap(rr.Quality, qual)
			mergeMap(rr.Design, design)
			mergeMap(rr.Indispensability, indisp)
			mergeMap(rr.DebtCleanup, debt)
			mergeMap(rr.Survival, repoSurvDecayed)
			mergeMap(rr.RawSurvival, repoSurvRaw)
			if repoSurvRobust != nil {
				mergeMap(rr.RobustSurvival, repoSurvRobust)
			}
			if repoSurvDormant != nil {
				mergeMap(rr.DormantSurvival, repoSurvDormant)
			}
			rf, rl := make(map[string]time.Time), make(map[string]time.Time)
			for _, c := range commits {
				rr.TotalCommits[c.Author]++
				if f, ok := rf[c.Author]; !ok || c.Date.Before(f) {
					rf[c.Author] = c.Date
				}
				if l, ok := rl[c.Author]; !ok || c.Date.After(l) {
					rl[c.Author] = c.Date
				}
			}
			repoAccumulators = append(repoAccumulators, repoAccState{
				acc: &accumulator{raw: rr}, repoName: repoName, domain: repoDomain,
				authorFirstDate: rf, authorLastDate: rl,
				commits: commits, blameLines: blameLines,
			})
		}
	}

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
		for author, repos := range acc.authorRepoCommits {
			count := 0
			for _, n := range repos {
				if n >= 3 {
					count++
				}
			}
			if count > 0 {
				acc.raw.Breadth[author] = float64(count)
			}
		}
		for author, total := range acc.raw.Production {
			first, last := acc.authorFirstDate[author], acc.authorLastDate[author]
			days := last.Sub(first).Hours() / 24
			if days < 1 {
				days = 1
			}
			acc.raw.Production[author] = total / days
		}
		scored := scorer.Score(acc.raw, cfg, acc.authorLastDate)
		var filtered []scorer.Result
		for _, r := range scored {
			if !cfg.IsExcludedAuthor(r.Author) {
				filtered = append(filtered, r)
			}
		}
		if len(filtered) == 0 {
			continue
		}
		// Module Topology — compute from accumulated commits/blameLines
		cochange := metric.CalcCochange(acc.allCommits)
		ownership := metric.CalcOwnershipFragmentation(acc.allBlameLines)
		moduleSurvival := metric.CalcModuleSurvival(acc.allBlameLines, cfg.Tau, start)

		dr := DomainResults{
			Domain: d, Results: filtered, Risks: acc.risks, RepoCount: acc.repoCount,
			Cochange: cochange, Ownership: ownership, ModuleSurvival: moduleSurvival,
		}

		if opts.PerRepo {
			for _, ra := range repoAccumulators {
				if ra.domain != d {
					continue
				}
				for a, t := range ra.acc.raw.Production {
					days := ra.authorLastDate[a].Sub(ra.authorFirstDate[a]).Hours() / 24
					if days < 1 {
						days = 1
					}
					ra.acc.raw.Production[a] = t / days
				}
				for a := range ra.acc.raw.TotalCommits {
					ra.acc.raw.Breadth[a] = 1
				}
				s := scorer.Score(ra.acc.raw, cfg, ra.authorLastDate)
				var rf []scorer.Result
				for _, r := range s {
					if !cfg.IsExcludedAuthor(r.Author) {
						rf = append(rf, r)
					}
				}
				if len(rf) > 0 {
					// Per-repo module topology
					repoCochange := metric.CalcCochange(ra.commits)
					repoOwnership := metric.CalcOwnershipFragmentation(ra.blameLines)
					repoModuleSurvival := metric.CalcModuleSurvival(ra.blameLines, cfg.Tau, start)
					dr.PerRepo = append(dr.PerRepo, RepoResult{
						RepoName: ra.repoName, Domain: ra.domain, Results: rf,
						Cochange: repoCochange, Ownership: repoOwnership, ModuleSurvival: repoModuleSurvival,
					})
				}
			}
		}
		results = append(results, dr)
	}
	return results, nil
}

// ResolveRepoDomain determines the domain for a repo.
func ResolveRepoDomain(ctx context.Context, repoPath, repoName string, cfg *config.Config) domain.Domain {
	for name, entry := range cfg.Domains {
		if len(entry.Repos) > 0 && domain.MatchRepoPattern(repoName, entry.Repos) {
			return domain.NormalizeName(name)
		}
	}
	files, err := git.ListAllFiles(ctx, repoPath)
	if err != nil || len(files) == 0 {
		return domain.Unknown
	}
	extMap := domain.BuildExtMap(cfg.CustomExtensions(), cfg.UseDefaultDomains())
	return domain.DetectFromFiles(files, extMap)
}

// FindGitRepos walks a directory tree up to maxDepth and returns paths containing .git.
func FindGitRepos(root string, maxDepth int) ([]string, error) {
	var repos []string
	rootDepth := len(strings.Split(filepath.ToSlash(root), "/"))
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		if len(strings.Split(filepath.ToSlash(path), "/"))-rootDepth > maxDepth {
			return filepath.SkipDir
		}
		if fi, err := os.Stat(filepath.Join(path, ".git")); err == nil {
			if fi.IsDir() {
				repos = append(repos, path)
			}
			return filepath.SkipDir
		}
		return nil
	})
	return repos, err
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
