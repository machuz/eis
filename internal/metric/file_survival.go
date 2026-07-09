package metric

import (
	"math"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

// FileSurvival holds per-file surviving-mass stats used to pick propagation
// anchors (surviving exemplar code). It reuses the SAME time-decay as the
// survival metric: each blame line contributes exp(-ageDays/tau), so DecayedMass
// is the file's surviving gravity and DecayedMass/Lines its mean survival.
//
// FIREWALL: it keeps only a COUNT of distinct contributors (the others-contested
// signal), never author identities — an anchor is "code that survived and others
// built on", not "so-and-so's code".
type FileSurvival struct {
	File         string  // repo-relative path
	Module       string  // resolved module id
	DecayedMass  float64 // Σ exp(-ageDays/tau) — surviving gravity
	Lines        int     // surviving blame lines
	Contributors int     // distinct canonical contributors (co-author aware)
}

// CalcFileSurvival folds blame lines into per-file surviving-mass stats. tau and
// now must be the run's survival parameters (repo tau + analysisTime) so the
// decay matches CalcSurvival. Files resolving to no module are still returned
// (Module==""); the caller filters.
func CalcFileSurvival(blameLines []git.BlameLine, tau float64, now time.Time, mr ModuleResolver) map[string]*FileSurvival {
	out := make(map[string]*FileSurvival)
	// Temporary per-file author sets, collapsed to a count at the end so no
	// identity ever leaves this function.
	seen := make(map[string]map[string]struct{})

	for _, bl := range blameLines {
		days := now.Sub(bl.CommitterTime).Hours() / 24
		if days < 0 {
			days = 0
		}
		w := math.Exp(-days / tau)

		fs := out[bl.Filename]
		if fs == nil {
			fs = &FileSurvival{File: bl.Filename, Module: mr.ModuleOf(bl.Filename)}
			out[bl.Filename] = fs
			seen[bl.Filename] = make(map[string]struct{})
		}
		fs.DecayedMass += w
		fs.Lines++
		if len(bl.Authors) > 0 {
			for _, a := range bl.Authors {
				seen[bl.Filename][a] = struct{}{}
			}
		} else {
			seen[bl.Filename][bl.Author] = struct{}{}
		}
	}
	for f, fs := range out {
		fs.Contributors = len(seen[f])
	}
	return out
}
