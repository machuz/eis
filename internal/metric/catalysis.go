package metric

import (
	"math"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

// CalcCatalysis measures how much an engineer *enabled others* — the surviving
// mass of code that later contributors built on top of this engineer's still-
// living foundation.
//
// It is the one axis that looks at what your code did FOR OTHERS rather than
// what you did to the code. The foundation is not "who created the file" — a
// file-creator whose lines were rewritten away built nothing lasting. It is
// surviving precedence, weighted by how much of your foundation actually remains:
//
//	Catalysis(X) = Σ over files where X's own code STILL SURVIVES, of the
//	               surviving mass contributed by authors who first touched that
//	               file AFTER X did, EACH weighted by X's surviving share of the
//	               file (xSurv / total surviving mass in the file).
//
// The share weight is what stops "created the file first, therefore credited for
// everything built in it later." If X seeds a file and B rewrites most of it,
// X's surviving share shrinks toward zero, so X is credited for almost none of
// B's mass — B is now the foundation, not X. Only when X's foundation still
// substantially survives does X collect the credit for what others built. A full
// rewrite (X's share == 0) yields nothing, as before; a partial rewrite no longer
// pays the file-creator B's whole surviving mass. Triple-gated against gaming: it
// needs (1) X's own foundation to survive, (2) other distinct humans arriving
// later, (3) whose work also survives. No line-level history is required: only
// per-file, per-author first-contribution dates (git log) and surviving mass
// (git blame).
//
// `now` is the decay reference (passed in, never time.Now() here) so the result
// is reproducible for a given (commits, blame, now).
func CalcCatalysis(commits []git.Commit, blameLines []git.BlameLine, tau float64, now time.Time) map[string]float64 {
	// file -> author -> earliest (non-merge) contribution date
	firstContrib := make(map[string]map[string]time.Time)
	for _, c := range commits {
		if c.IsMerge {
			continue
		}
		for _, fs := range c.FileStats {
			fa := firstContrib[fs.Filename]
			if fa == nil {
				fa = make(map[string]time.Time)
				firstContrib[fs.Filename] = fa
			}
			// Every contributor (author + co-authors) shares the origination date,
			// so a co-authored file-creating commit makes both foundations.
			for _, a := range CommitAuthors(c) {
				if d, ok := fa[a]; !ok || c.Date.Before(d) {
					fa[a] = c.Date
				}
			}
		}
	}

	// file -> author -> surviving, time-decayed mass
	survMass := make(map[string]map[string]float64)
	for _, bl := range blameLines {
		fa := survMass[bl.Filename]
		if fa == nil {
			fa = make(map[string]float64)
			survMass[bl.Filename] = fa
		}
		daysAlive := now.Sub(bl.CommitterTime).Hours() / 24
		if daysAlive < 0 {
			daysAlive = 0
		}
		factor := 1.0
		if tau > 0 {
			factor = math.Exp(-daysAlive / tau)
		}
		blameShares(bl, func(a string, s float64) { fa[a] += factor * s })
	}

	result := make(map[string]float64)
	for file, authorSurv := range survMass {
		fc := firstContrib[file]
		if fc == nil {
			continue // no commit history for this file — precedence unknown
		}
		// Total surviving mass in the file — the denominator of each author's
		// surviving share. Always > 0 here (survMass only holds factor>0 lines).
		fileTotal := 0.0
		for _, s := range authorSurv {
			fileTotal += s
		}
		if fileTotal <= 0 {
			continue
		}
		for x, xSurv := range authorSurv {
			if xSurv <= 0 {
				continue // X's foundation must still survive to be a foundation
			}
			xFirst, ok := fc[x]
			if !ok {
				continue
			}
			// X is the foundation only to the extent X's own code still survives
			// here: a file-creator rewritten down to a remnant has a tiny share
			// and is credited for almost none of the later mass.
			xShare := xSurv / fileTotal
			// Credit X for the surviving mass of everyone who arrived later,
			// scaled by X's surviving share of the file.
			for b, bSurv := range authorSurv {
				if b == x {
					continue
				}
				if bFirst, ok := fc[b]; ok && bFirst.After(xFirst) {
					result[x] += bSurv * xShare
				}
			}
		}
	}

	return result
}
