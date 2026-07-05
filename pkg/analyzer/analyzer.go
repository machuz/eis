// Package analyzer provides the core analysis pipeline for external use.
// This is a library-friendly version without CLI dependencies (no spinner, color, cache).
package analyzer

import (
	"context"
	"fmt"
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

	// AnalysisTime is the W-02 envelope clock: the single decay/classification
	// reference threaded into every score-affecting computation (survival,
	// catalysis, module survival, scorer.ScoreAt). Pin it for reproducibility —
	// same (git_sha, AnalysisTime) ⇒ same scores. Zero value defaults to
	// time.Now().UTC(), preserving prior behavior for existing callers. This
	// mirrors the analyze CLI and timeline so the three pipelines cannot drift.
	AnalysisTime time.Time
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
	// ModuleSurvivalByAuthor decomposes surviving blame mass by module → author.
	// The spatial distribution of each author's surviving gravity; the primitive
	// for structure-neutral (module-unit) Breadth on the consuming side.
	ModuleSurvivalByAuthor map[string]map[string]float64
}

// RepoResult holds scored results for a single repository.
type RepoResult struct {
	RepoName               string
	Domain                 domain.Domain
	Results                []scorer.Result
	Cochange               metric.CochangeResult
	Ownership              []metric.ModuleOwnership
	ModuleSurvival         map[string]float64
	ModuleSurvivalByAuthor map[string]map[string]float64
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

	// analysisTime is the W-02 envelope clock — the SINGLE decay/classification
	// reference threaded through every score below (survival, catalysis, module
	// survival, scorer.ScoreAt). Pin it for reproducibility; defaults to now.
	// (This library twin has no perf timer, so there is no separate `start`.)
	// See the analyze CLI and pkg/timeline/run.go (window.End) for the same
	// threading — keeping the three pipelines in lockstep is a W-02 invariant.
	analysisTime := opts.AnalysisTime
	if analysisTime.IsZero() {
		analysisTime = time.Now().UTC()
	}

	workers := opts.Workers
	if workers == 0 {
		workers = 4
	}

	// Module resolvers are built per repo (see the loop below) so each
	// repo can honor its own RepoOverrides without leaking pattern sets
	// across repos. Resolution remains deterministic per repo (W-02/W-03)
	// because it depends only on the resolved pattern list.

	// Initialize cache store
	cacheStore := cache.NewWithDir(opts.CacheEnabled, opts.CacheDir)

	type accumulator struct {
		raw                 *metric.RawScores
		debtCounts          map[string]int
		authorRepoCommits   map[string]map[string]int
		authorModuleCommits map[string]map[string]int
		authorFirstDate     map[string]time.Time
		authorLastDate      map[string]time.Time
		repoCount           int
		risks               []metric.ModuleRisk
		changePressure      metric.ChangePressure
		// Module Topology accumulators
		allCommits    []git.Commit
		allBlameLines []git.BlameLine
	}
	newAcc := func() *accumulator {
		return &accumulator{
			raw:                 metric.NewRawScores(),
			debtCounts:          make(map[string]int),
			authorRepoCommits:   make(map[string]map[string]int),
			authorModuleCommits: make(map[string]map[string]int),
			authorFirstDate:     make(map[string]time.Time),
			authorLastDate:      make(map[string]time.Time),
			changePressure:      make(metric.ChangePressure),
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
		// resolver carries this repo's effective ModuleResolver so per-repo
		// rollups (Cochange / Ownership / ModuleSurvival) honor RepoOverrides
		// instead of falling back to the org-level resolver used for the
		// domain merge.
		resolver metric.ModuleResolver
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

		// Per-repo module resolver: honors RepoOverrides for this repo.
		// The lookup key here is repoName (filepath.Base) — match that in
		// the YAML repo_overrides map.
		moduleResolver := metric.NewModuleResolverWithExcludes(config.PatternsForRepo(cfg, repoName), config.ExcludesForRepo(cfg, repoName))

		// Get HEAD hash for cache keys
		headHash, _ := git.HeadHash(ctx, repoPath)

		// Parse git log (cached)
		var commits []git.Commit
		logCacheKey := cache.LogKey(repoPath, headHash, cfg.CommentFilterEnabled())
		if headHash != "" && cacheStore.Get(logCacheKey, &commits) {
			// cache hit
		} else {
			var err error
			git.ConfigureBotCoAuthorPatterns(cfg.BotCoAuthorPatterns)
			commits, err = git.ParseLogParallel(ctx, repoPath, workers, cfg.CommentFilterEnabled())
			if err != nil {
				return nil, fmt.Errorf("parse log %s: %w", repoName, err)
			}
			if headHash != "" {
				cacheStore.Set(logCacheKey, commits)
			}
		}
		// Collapse split identities (one person under several names sharing an
		// email) before alias resolution; reused to remap blame authors by name.
		idmap := git.BuildIdentityMap(commits)
		git.CanonicalizeAuthors(commits, nil, idmap)
		// config patterns + .gitattributes linguist-generated/vendored (so generated
		// /vendored files don't inflate gravity); used by every file filter below.
		excludes := effectiveExcludes(ctx, repoPath, cfg)
		commits = filterCommits(commits, cfg)
		commits = filterFileStats(commits, excludes)

		// Module liveness gate (ADR step 2): fold fallback-derived modules that
		// were not touched in >= ModuleLivenessMinMonths distinct calendar
		// months into metric.PeripheralModule. Computed from commit.Date
		// (deterministic, W-02). Must run BEFORE the resolver reaches any
		// module-topology metric so the bucketing reflects the fold.
		fold := metric.ComputeModuleFold(commits, moduleResolver, cfg.ModuleLivenessMinMonths)
		moduleResolver = moduleResolver.WithFold(fold)

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

		prod := metric.CalcProduction(commits, excludes)
		mergeMap(acc.raw.Production, prod)

		added, deleted := metric.CalcLines(commits, excludes)
		mergeMapInt(acc.raw.LinesAdded, added)
		mergeMapInt(acc.raw.LinesDeleted, deleted)

		// Catalysis is computed in the blame stage below (it needs the commit
		// history for file origin + the surviving blame lines).

		// Design is computed in the blame stage below (survival-weighted: surviving
		// arch-file blame lines, not arch lines changed — needs blame + analysisTime).

		for _, c := range commits {
			if _, ok := acc.authorRepoCommits[c.Author]; !ok {
				acc.authorRepoCommits[c.Author] = make(map[string]int)
			}
			acc.authorRepoCommits[c.Author][repoName]++
			// Per-module commit counts feed module-unit Breadth. A commit
			// counts once per distinct module it touches (W-02: dedup so
			// touching three files in one module isn't three "modules").
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
		files = filterFiles(files, excludes)

		var blameVerbose func(string)
		if cb.OnVerbose != nil {
			blameVerbose = cb.OnVerbose
		}

		// Blame analysis (cached)
		var blameLines []git.BlameLine
		git.ConfigureBlameMoveDetection(cfg.BlameMoveDetection)
		blameCacheKey := cache.BlameKey(repoPath, headHash, files, cfg.SampleSize, cfg.BlameMoveDetection)
		if headHash != "" && cacheStore.Get(blameCacheKey, &blameLines) {
			// cache hit
		} else {
			// Pre-skip oversized files so blame can't deadlock on a
			// single huge single-line dump.
			files, _ = git.FilterFilesBySize(ctx, repoPath, "HEAD", files, cfg.MaxBlameFileBytes, blameVerbose)
			blameLines, _ = git.ConcurrentBlameFiles(ctx, repoPath, files, cfg.SampleSize, workers, cfg.BlameTimeout, cb.OnBlameProgress, blameVerbose)
			if headHash != "" && len(blameLines) > 0 {
				cacheStore.Set(blameCacheKey, blameLines)
			}
		}
		git.CanonicalizeAuthors(nil, blameLines, idmap)
		coMap := metric.CoAuthorMap(commits)
		for i := range blameLines {
			blameLines[i].Author = cfg.ResolveAuthor(blameLines[i].Author)
			if set, ok := coMap[blameLines[i].Commit]; ok {
				blameLines[i].Authors = set
			}
		}
		blameLines = filterBlameLines(blameLines, cfg)

		// Accumulate for Module Topology
		acc.allCommits = append(acc.allCommits, commits...)
		acc.allBlameLines = append(acc.allBlameLines, blameLines...)

		// Design (survival-weighted): surviving, time-decayed arch-file blame lines
		// per author, on the same analysisTime + tau as survival. Arch churn later
		// rewritten has decayed out of the blame, so design measures durable structural
		// ownership rather than raw arch-edit volume.
		designSurv := metric.CalcDesignSurviving(blameLines, cfg.ArchitecturePatterns, cfg.Tau, analysisTime)
		mergeMap(acc.raw.Design, designSurv)

		var repoSurvDecayed, repoSurvRaw, repoSurvRobust, repoSurvDormant map[string]float64
		if opts.PressureMode != "ignore" {
			repoPressure := metric.CalcChangePressure(commits, blameLines, moduleResolver)
			for mod, p := range repoPressure {
				acc.changePressure[repoName+"/"+mod] = p
			}
			// blameLines authors are already alias-resolved in place above.
			blameByAuthor := make(map[string]int)
			for _, bl := range blameLines {
				blameByAuthor[bl.Author]++
			}
			threshold := metric.PressureThreshold(repoPressure, blameByAuthor, metric.SubstantialAuthorLines)
			repoOthers := metric.CalcOthersPressure(commits, blameLines, moduleResolver)
			sr := metric.CalcSurvivalWithPressure(blameLines, cfg.TauForDomain(string(repoDomain)), analysisTime, repoPressure, threshold, moduleResolver, repoOthers)
			repoSurvDecayed, repoSurvRaw, repoSurvRobust, repoSurvDormant = sr.Decayed, sr.Raw, sr.Robust, sr.Dormant
			mergeMap(acc.raw.Survival, repoSurvDecayed)
			mergeMap(acc.raw.RawSurvival, repoSurvRaw)
			mergeMap(acc.raw.RobustSurvival, repoSurvRobust)
			mergeMap(acc.raw.DormantSurvival, repoSurvDormant)
		} else {
			sr := metric.CalcSurvival(blameLines, cfg.TauForDomain(string(repoDomain)), analysisTime)
			repoSurvDecayed, repoSurvRaw = sr.Decayed, sr.Raw
			mergeMap(acc.raw.Survival, repoSurvDecayed)
			mergeMap(acc.raw.RawSurvival, repoSurvRaw)
		}

		indisp, risks := metric.CalcIndispensability(blameLines, moduleResolver, cfg.BusFactor.Critical, cfg.BusFactor.High)
		mergeMap(acc.raw.Indispensability, indisp)

		// Catalysis: surviving mass of others' work on files this author
		// originated (commit history → file origin; blame → surviving mass).
		catalysis := metric.CalcCatalysis(commits, blameLines, cfg.TauForDomain(string(repoDomain)), analysisTime)
		mergeMap(acc.raw.Catalysis, catalysis)
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
		debt, _ := metric.CalcDebt(ctx, repoPath, fixCommits, 50, cfg.DebtThreshold, cfg.BlameTimeout, cfg.ResolveAuthor, metric.CoAuthorMap(commits), debtProg, debtVerbose)
		mergeMapAvg(acc.raw.DebtCleanup, debt, acc.debtCounts)

		if opts.PerRepo {
			rr := metric.NewRawScores()
			mergeMap(rr.Production, prod)
			mergeMap(rr.Catalysis, catalysis)
			mergeMap(rr.Design, designSurv)
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
				resolver: moduleResolver,
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
		// Breadth = effective number of modules an author holds surviving
		// gravity in (Hill number over per-module gravity), structure-neutral
		// and survival-weighted. The same per-(author,module) survival feeds
		// DomainResults below, so compute it once with the org-level resolver.
		domainResolver := metric.NewModuleResolverWithExcludes(config.PatternsForRepo(cfg, ""), config.ExcludesForRepo(cfg, ""))
		// Module liveness gate (ADR step 2) at the domain rollup: fold from the
		// domain's concatenated per-repo commits (acc.allCommits — date +
		// filestats present), deterministic from commit.Date (W-02). Applied
		// before any module-topology metric below (Breadth, cochange,
		// ownership, module survival) so the fold is honored uniformly.
		domainFold := metric.ComputeModuleFold(acc.allCommits, domainResolver, cfg.ModuleLivenessMinMonths)
		domainResolver = domainResolver.WithFold(domainFold)
		moduleSurvivalByAuthor := metric.CalcModuleSurvivalByAuthor(acc.allBlameLines, cfg.TauForDomain(string(d)), analysisTime, domainResolver)
		breadth := metric.ComputeBreadth(moduleSurvivalByAuthor)
		for author, b := range breadth {
			acc.raw.Breadth[author] = b
		}
		for author, total := range acc.raw.Production {
			first, last := acc.authorFirstDate[author], acc.authorLastDate[author]
			days := last.Sub(first).Hours() / 24
			if days < 1 {
				days = 1
			}
			acc.raw.Production[author] = total / days
		}
		scored := scorer.ScoreAt(acc.raw, cfg, acc.authorLastDate, analysisTime)
		var filtered []scorer.Result
		for _, r := range scored {
			if !cfg.IsExcludedAuthor(r.Author) {
				filtered = append(filtered, r)
			}
		}
		if len(filtered) == 0 {
			continue
		}
		// Module Topology — compute from accumulated cross-repo commits using
		// the ORG-level resolver (domainResolver, built above for Breadth):
		// this rollup spans multiple repos in a domain, so per-repo overrides
		// don't apply uniformly. moduleSurvivalByAuthor is reused from the
		// Breadth computation above (same resolver, same inputs).
		cochange := metric.CalcCochange(acc.allCommits, domainResolver)
		ownership := metric.CalcOwnershipFragmentation(acc.allBlameLines, domainResolver)
		moduleSurvival := metric.CalcModuleSurvival(acc.allBlameLines, cfg.TauForDomain(string(d)), analysisTime, domainResolver)

		dr := DomainResults{
			Domain: d, Results: filtered, Risks: acc.risks, RepoCount: acc.repoCount,
			Cochange: cochange, Ownership: ownership, ModuleSurvival: moduleSurvival,
			ModuleSurvivalByAuthor: moduleSurvivalByAuthor,
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
				// Per-repo Breadth = effective number of modules the author
				// holds gravity in WITHIN this repo (Hill number over this
				// repo's per-module gravity), using this repo's resolver so
				// RepoOverrides apply. Reused for the RepoResult below.
				repoModuleSurvivalByAuthor := metric.CalcModuleSurvivalByAuthor(ra.blameLines, cfg.TauForDomain(string(ra.domain)), analysisTime, ra.resolver)
				repoBreadth := metric.ComputeBreadth(repoModuleSurvivalByAuthor)
				for a := range ra.acc.raw.TotalCommits {
					ra.acc.raw.Breadth[a] = repoBreadth[a]
				}
				s := scorer.ScoreAt(ra.acc.raw, cfg, ra.authorLastDate, analysisTime)
				var rf []scorer.Result
				for _, r := range s {
					if !cfg.IsExcludedAuthor(r.Author) {
						rf = append(rf, r)
					}
				}
				if len(rf) > 0 {
					// Per-repo module topology, using THIS repo's resolver
					// (which honors RepoOverrides).
					repoCochange := metric.CalcCochange(ra.commits, ra.resolver)
					repoOwnership := metric.CalcOwnershipFragmentation(ra.blameLines, ra.resolver)
					repoModuleSurvival := metric.CalcModuleSurvival(ra.blameLines, cfg.TauForDomain(string(ra.domain)), analysisTime, ra.resolver)
					dr.PerRepo = append(dr.PerRepo, RepoResult{
						RepoName: ra.repoName, Domain: ra.domain, Results: rf,
						Cochange: repoCochange, Ownership: repoOwnership, ModuleSurvival: repoModuleSurvival,
						ModuleSurvivalByAuthor: repoModuleSurvivalByAuthor,
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

// effectiveExcludes returns the config exclude patterns plus, when
// respect_gitattributes is on, the linguist-generated/vendored paths (escaped to
// exact-match globs) so generated/vendored files don't inflate gravity.
// Best-effort — a git error falls back to the config patterns. Mirrors the same
// helper in internal/cli so the CLI and library pipelines exclude identically.
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
