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
func CalcModuleSurvival(blameLines []git.BlameLine, tau float64, now time.Time) map[string]float64 {
	type modStats struct {
		decayedSum float64
		rawCount   float64
	}

	modules := make(map[string]*modStats)

	for _, bl := range blameLines {
		mod := ModuleOf(bl.Filename)

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
