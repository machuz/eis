package metric

import (
	"math"
	"time"

	"github.com/machuz/engineering-impact-score/internal/git"
)

// SurvivalResult holds both time-decayed and raw (non-decayed) blame counts,
// plus robust/dormant survival split based on change pressure.
type SurvivalResult struct {
	Decayed map[string]float64 // time-weighted survival (total)
	Raw     map[string]float64 // raw blame line count (no decay)
	Robust  map[string]float64 // survival in high-pressure modules
	Dormant map[string]float64 // survival in low-pressure modules
}

func CalcSurvival(blameLines []git.BlameLine, tau float64, now time.Time) SurvivalResult {
	decayed := make(map[string]float64)
	raw := make(map[string]float64)

	for _, bl := range blameLines {
		raw[bl.Author]++

		daysAlive := now.Sub(bl.CommitterTime).Hours() / 24
		if daysAlive < 0 {
			daysAlive = 0
		}
		weight := math.Exp(-daysAlive / tau)
		decayed[bl.Author] += weight
	}

	return SurvivalResult{Decayed: decayed, Raw: raw}
}

// CalcSurvivalWithPressure splits survival into robust (high-pressure modules)
// and dormant (low-pressure modules) based on change pressure threshold.
func CalcSurvivalWithPressure(blameLines []git.BlameLine, tau float64, now time.Time, pressure ChangePressure, threshold float64) SurvivalResult {
	decayed := make(map[string]float64)
	raw := make(map[string]float64)
	robust := make(map[string]float64)
	dormant := make(map[string]float64)

	for _, bl := range blameLines {
		raw[bl.Author]++

		daysAlive := now.Sub(bl.CommitterTime).Hours() / 24
		if daysAlive < 0 {
			daysAlive = 0
		}
		weight := math.Exp(-daysAlive / tau)
		decayed[bl.Author] += weight

		mod := ModuleOf(bl.Filename)
		if pressure[mod] >= threshold {
			robust[bl.Author] += weight
		} else {
			dormant[bl.Author] += weight
		}
	}

	return SurvivalResult{Decayed: decayed, Raw: raw, Robust: robust, Dormant: dormant}
}
