package metric

import (
	"sort"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

// CommitAggregator folds each commit's contribution into compact per-author /
// per-module accumulators AS IT STREAMS from the log, so the analyzer never has
// to hold the full []Commit in memory — the ceiling for a cold first observation
// of a giant repo. Feed each already canonicalized, alias-resolved,
// exclusion-filtered, filestat-filtered, non-reverted commit via Fold IN HISTORY
// ORDER on a single goroutine, then call Finalize.
//
// Value-invariance: every field is the same per-commit reduction over the same
// filtered commit set the current materialized pipeline computes. Fold runs in
// history order on a single consumer, so the non-associative float sum in
// Production is bit-identical to the serial walk; all other reductions (int
// sums, date min/max, month/pair sets) are order-independent anyway.
//
// The resolver must already be FOLDED (module liveness gate applied) so ModuleOf
// buckets exactly as the materialized path does.
type CommitAggregator struct {
	mr ModuleResolver

	// Production is order-sensitive (float sum) — folded in history order.
	production   map[string]float64
	linesAdded   map[string]int
	linesDeleted map[string]int

	totalCommits        map[string]int
	authorModuleCommits map[string]map[string]int
	firstDate           map[string]time.Time
	lastDate            map[string]time.Time

	// Others-pressure accumulators — SUBSTANCE-GATED, kept SEPARATE from the
	// activity-based moduleCommits/authorModuleCommits above. A commit counts
	// toward a module here only if it made a substantive (non-cosmetic) change to
	// that module's files (see MinSubstantiveLines), so a rename/comment/blank-only
	// touch by another author no longer fakes "contested". Change-pressure,
	// co-change, and Breadth deliberately keep the un-gated tallies (they measure
	// activity/coupling, not contest). Mirrors the batch CalcOthersPressure gate so
	// the two ingest paths stay value-identical.
	substModuleCommits       map[string]int            // module → substantive commits touching it
	substModuleAuthorCommits map[string]map[string]int // module → author → substantive commits

	// co-change accumulators (Jaccard computed once at Finalize)
	moduleCommits map[string]int
	pairCount     map[[2]string]int

	// moduleLastDate: per (folded) module → the latest commit date that touched
	// it. Date max is order-independent, so it's value-identical across the
	// streaming and materialized ingest paths. Feeds the "untouched N days"
	// figure in the structural-debt drill rows.
	moduleLastDate map[string]time.Time

	// catalysis precedence: file → author → earliest contribution date. History-
	// sublinear (bounded by distinct files × authors-per-file), retained past the
	// stream so the blame stage can combine it with surviving mass.
	firstContrib map[string]map[string]time.Time

	// sparse SHA → contributor set, only for co-authored commits.
	coMap map[string][]string

	// fix commits, in history order, with FileStats dropped (CalcDebt never reads
	// them) — the order matches GetFixCommits so DebtKey and the debt sample are
	// identical.
	fixCommits []git.Commit
}

// NewCommitAggregator returns an aggregator over the given FOLDED resolver.
func NewCommitAggregator(mr ModuleResolver) *CommitAggregator {
	return &CommitAggregator{
		mr:                       mr,
		production:               make(map[string]float64),
		linesAdded:               make(map[string]int),
		linesDeleted:             make(map[string]int),
		totalCommits:             make(map[string]int),
		authorModuleCommits:      make(map[string]map[string]int),
		firstDate:                make(map[string]time.Time),
		lastDate:                 make(map[string]time.Time),
		substModuleCommits:       make(map[string]int),
		substModuleAuthorCommits: make(map[string]map[string]int),
		moduleCommits:            make(map[string]int),
		pairCount:                make(map[[2]string]int),
		moduleLastDate:           make(map[string]time.Time),
		firstContrib:             make(map[string]map[string]time.Time),
		coMap:                    make(map[string][]string),
	}
}

// Fold ingests one filtered commit (see the type doc for the required
// pre-filtering) and updates every accumulator.
func (a *CommitAggregator) Fold(c *git.Commit) {
	authors := CommitAuthors(*c)
	share := 1.0 / float64(len(authors))

	// Production (split across contributors) + Lines (keyed by primary author
	// only, matching CalcLines). Mirrors CalcProduction/CalcLines exactly; the
	// commits are already filestat-filtered, so no exclude re-check is needed.
	for _, fs := range c.FileStats {
		v := float64(fs.Insertions+fs.Deletions) * share
		for _, au := range authors {
			a.production[au] += v
		}
		a.linesAdded[c.Author] += fs.Insertions
		a.linesDeleted[c.Author] += fs.Deletions
	}

	a.totalCommits[c.Author]++

	// Unique folded modules this commit touches — shared by per-module commit
	// counts (Breadth input) and co-change, exactly like the materialized path's
	// two separate passes (same resolver ⇒ same set).
	var touched map[string]bool
	for _, fs := range c.FileStats {
		if mod := a.mr.ModuleOf(fs.Filename); mod != "" {
			if touched == nil {
				touched = make(map[string]bool)
			}
			touched[mod] = true
		}
	}
	if len(touched) > 0 {
		m := a.authorModuleCommits[c.Author]
		if m == nil {
			m = make(map[string]int)
			a.authorModuleCommits[c.Author] = m
		}
		mods := make([]string, 0, len(touched))
		for mod := range touched {
			m[mod]++
			a.moduleCommits[mod]++
			if d, ok := a.moduleLastDate[mod]; !ok || c.Date.After(d) {
				a.moduleLastDate[mod] = c.Date
			}
			mods = append(mods, mod)
		}
		// Canonical order for deterministic pair keys (matches CalcCochange).
		sort.Strings(mods)
		for i := 0; i < len(mods); i++ {
			for j := i + 1; j < len(mods); j++ {
				a.pairCount[[2]string{mods[i], mods[j]}]++
			}
		}
	}

	// Others-pressure tallies (SUBSTANCE-GATED, separate from the activity set
	// above). A module is contested by this commit only if at least one of its
	// files had a substantive (non-cosmetic) change — mirrors the batch
	// CalcOthersPressure gate so streaming and materialized stay value-identical.
	var substTouched map[string]bool
	for _, fs := range c.FileStats {
		if fs.Insertions+fs.Deletions < MinSubstantiveLines {
			continue
		}
		if mod := a.mr.ModuleOf(fs.Filename); mod != "" {
			if substTouched == nil {
				substTouched = make(map[string]bool)
			}
			substTouched[mod] = true
		}
	}
	for mod := range substTouched {
		a.substModuleCommits[mod]++
		am := a.substModuleAuthorCommits[mod]
		if am == nil {
			am = make(map[string]int)
			a.substModuleAuthorCommits[mod] = am
		}
		am[c.Author]++
	}

	if d, ok := a.firstDate[c.Author]; !ok || c.Date.Before(d) {
		a.firstDate[c.Author] = c.Date
	}
	if d, ok := a.lastDate[c.Author]; !ok || c.Date.After(d) {
		a.lastDate[c.Author] = c.Date
	}

	// Catalysis precedence (non-merge only — the stream is --no-merges, but keep
	// the guard for parity with BuildFirstContrib).
	if !c.IsMerge {
		for _, fs := range c.FileStats {
			fa := a.firstContrib[fs.Filename]
			if fa == nil {
				fa = make(map[string]time.Time)
				a.firstContrib[fs.Filename] = fa
			}
			for _, au := range authors {
				if d, ok := fa[au]; !ok || c.Date.Before(d) {
					fa[au] = c.Date
				}
			}
		}
	}

	if len(c.CoAuthors) > 0 {
		a.coMap[c.Hash] = CommitAuthors(*c)
	}

	if isFixCommit(c.Subject) {
		// Copy the co-authors slice — the streamed commit's backing array is
		// reused/freed after Fold returns.
		var co []string
		if len(c.CoAuthors) > 0 {
			co = append(co, c.CoAuthors...)
		}
		a.fixCommits = append(a.fixCommits, git.Commit{
			Hash:      c.Hash,
			Author:    c.Author,
			Email:     c.Email,
			Date:      c.Date,
			Subject:   c.Subject,
			CoAuthors: co,
		})
	}
}

// Aggregate is the finalized output of a CommitAggregator — exactly the
// commit-derived maps the analyzer's per-repo loop builds today.
type Aggregate struct {
	Production          map[string]float64
	LinesAdded          map[string]int
	LinesDeleted        map[string]int
	TotalCommits        map[string]int
	AuthorModuleCommits map[string]map[string]int
	FirstDate           map[string]time.Time
	LastDate            map[string]time.Time
	Cochange            CochangeResult
	FirstContrib        map[string]map[string]time.Time
	CoMap               map[string][]string
	FixCommits          []git.Commit
	// OthersModuleCommits and ModuleAuthorCommits are the SUBSTANCE-GATED
	// others-pressure inputs: a commit is counted (once per module) only if it made
	// a substantive change to that module's files (MinSubstantiveLines) — cosmetic
	// touches (rename/comment/blank-only) are excluded so they can't fake
	// "contested". OthersModuleCommits is the total per module; ModuleAuthorCommits
	// is module → author → substantive commits. These are the exact
	// (moduleCommits, moduleAuthorCommits) pair CalcOthersPressureFrom needs, and
	// they must be used TOGETHER (never mixed with the un-gated Cochange.ModuleCommits,
	// or the others = total − self subtraction would over-count cosmetic commits).
	OthersModuleCommits map[string]int
	ModuleAuthorCommits map[string]map[string]int
	// ModuleLastDate maps each folded module → the latest commit date that
	// touched it (for the structural-debt "untouched N days" figure).
	ModuleLastDate map[string]time.Time
}

// Finalize returns the accumulated maps. Cochange's Jaccard + sort runs once here
// (deterministic regardless of fold order), matching CalcCochange.
func (a *CommitAggregator) Finalize() Aggregate {
	return Aggregate{
		Production:          a.production,
		LinesAdded:          a.linesAdded,
		LinesDeleted:        a.linesDeleted,
		TotalCommits:        a.totalCommits,
		AuthorModuleCommits: a.authorModuleCommits,
		FirstDate:           a.firstDate,
		LastDate:            a.lastDate,
		Cochange:            buildCochangeResult(a.moduleCommits, a.pairCount),
		FirstContrib:        a.firstContrib,
		CoMap:               a.coMap,
		FixCommits:          a.fixCommits,
		OthersModuleCommits: a.substModuleCommits,
		ModuleAuthorCommits: a.substModuleAuthorCommits,
		ModuleLastDate:      a.moduleLastDate,
	}
}
