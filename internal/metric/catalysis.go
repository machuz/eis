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
// file-creator whose lines were all rewritten away built nothing lasting. It is
// surviving precedence:
//
//	Catalysis(X) = Σ over files where X's own code STILL SURVIVES, of the
//	               surviving mass contributed by authors who first touched that
//	               file AFTER X did.
//
// So whoever's living code others later built more living code upon is the
// catalyst. If X seeds a file and B rewrites it entirely, X's lines vanish (X
// gets nothing) and B — whose code now survives — becomes the foundation the
// next contributor builds on. Triple-gated against gaming: it needs (1) X's
// own foundation to survive, (2) other distinct humans arriving later, (3)
// whose work also survives. No line-level history is required: only per-file,
// per-author first-contribution dates (git log) and surviving mass (git blame).
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
			if d, ok := fa[c.Author]; !ok || c.Date.Before(d) {
				fa[c.Author] = c.Date
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
		fa[bl.Author] += factor
	}

	result := make(map[string]float64)
	for file, authorSurv := range survMass {
		fc := firstContrib[file]
		if fc == nil {
			continue // no commit history for this file — precedence unknown
		}
		for x, xSurv := range authorSurv {
			if xSurv <= 0 {
				continue // X's foundation must still survive to be a foundation
			}
			xFirst, ok := fc[x]
			if !ok {
				continue
			}
			// Credit X for the surviving mass of everyone who arrived later.
			for b, bSurv := range authorSurv {
				if b == x {
					continue
				}
				if bFirst, ok := fc[b]; ok && bFirst.After(xFirst) {
					result[x] += bSurv
				}
			}
		}
	}

	return result
}
