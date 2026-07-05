package metric

import (
	"math"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

// CalcModuleSurvival computes per-module survival rate.
// For each module, it calculates the ratio of time-decayed blame sum
// to raw blame count, yielding a 0-1 survival rate.
// This measures how well a module's code endures over time,
// independent of who wrote it.
func CalcModuleSurvival(blameLines []git.BlameLine, tau float64, now time.Time, mr ModuleResolver) map[string]float64 {
	type modStats struct {
		decayedSum float64
		rawCount   float64
	}

	modules := make(map[string]*modStats)

	for _, bl := range blameLines {
		mod := mr.ModuleOf(bl.Filename)
		if mod == "" {
			continue // excluded module (ExcludeModules)
		}

		ms, ok := modules[mod]
		if !ok {
			ms = &modStats{}
			modules[mod] = ms
		}

		ms.rawCount++

		daysAlive := now.Sub(bl.CommitterTime).Hours() / 24
		if daysAlive < 0 {
			daysAlive = 0
		}
		weight := math.Exp(-daysAlive / tau)
		ms.decayedSum += weight
	}

	result := make(map[string]float64, len(modules))
	for mod, ms := range modules {
		if ms.rawCount > 0 {
			result[mod] = ms.decayedSum / ms.rawCount
		}
	}

	return result
}

// CalcModuleSurvivalByAuthor decomposes surviving blame by (module, author).
// Where CalcModuleSurvival returns a per-module durability RATE
// (decayed/raw), this returns the per-author time-decayed surviving blame
// MASS within each module — the sum of line decay weights an author still
// holds there. Summed over modules it approximates an author's total decayed
// survival; per module it is the spatial distribution of their surviving
// gravity.
//
// This is the primitive the consuming side needs to make Breadth structure-
// neutral: the effective number of modules over which an author holds
// surviving gravity (a Hill number over these per-module masses) is invariant
// to whether the code is packaged as one monorepo or many repos. The decay
// math mirrors CalcModuleSurvival exactly so the two stay consistent (W-02).
func CalcModuleSurvivalByAuthor(blameLines []git.BlameLine, tau float64, now time.Time, mr ModuleResolver) map[string]map[string]float64 {
	result := make(map[string]map[string]float64)

	for _, bl := range blameLines {
		mod := mr.ModuleOf(bl.Filename)
		if mod == "" {
			continue // excluded module (ExcludeModules) — keeps Breadth clean
		}

		daysAlive := now.Sub(bl.CommitterTime).Hours() / 24
		if daysAlive < 0 {
			daysAlive = 0
		}
		weight := math.Exp(-daysAlive / tau)

		authors, ok := result[mod]
		if !ok {
			authors = make(map[string]float64)
			result[mod] = authors
		}
		blameShares(bl, func(a string, s float64) { authors[a] += weight * s })
	}

	return result
}
