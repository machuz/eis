package metric

import (
	"math"
	"time"

	"github.com/machuz/eis/internal/git"
)

// SurvivalResult holds multiple views of per-author blame survival:
//   - Decayed: time-weighted survival with untested lines scaled by α (drives Impact)
//   - Raw: line count, no decay, no weighting (for archetype / volume heuristics)
//   - Robust / Dormant: Decayed split by module change pressure (Impact weighting)
//   - Tested / Untested: time-decayed survival split by test coverage,
//     BEFORE α is applied — exposed for SaaS observability so the weighting
//     can be analysed or overridden downstream.
type SurvivalResult struct {
	Decayed  map[string]float64
	Raw      map[string]float64
	Robust   map[string]float64
	Dormant  map[string]float64
	Tested   map[string]float64
	Untested map[string]float64
}

// DefaultUntestedWeight is the multiplier applied to untested blame lines
// when folding them into Decayed / Robust / Dormant. A value of 0.5 means
// untested survival contributes half as much as tested survival to Impact.
const DefaultUntestedWeight = 0.5

func CalcSurvival(blameLines []git.BlameLine, tau float64, now time.Time) SurvivalResult {
	return calcSurvivalImpl(blameLines, tau, now, nil, 0, nil, 1.0)
}

// CalcSurvivalWithPressure splits survival into robust (high-pressure modules)
// and dormant (low-pressure modules) based on change pressure threshold.
func CalcSurvivalWithPressure(blameLines []git.BlameLine, tau float64, now time.Time, pressure ChangePressure, threshold float64) SurvivalResult {
	return calcSurvivalImpl(blameLines, tau, now, pressure, threshold, nil, 1.0)
}

// CalcSurvivalFull is the comprehensive survival calculator. It produces the
// pressure-based split (Robust / Dormant) AND the test-coverage split
// (Tested / Untested) in a single pass, applying untestedWeight to the
// effective Decayed/Robust/Dormant values so that untested code contributes
// proportionally less to Impact.
//
// Pass nil `tested` (or untestedWeight=1.0) to skip the test-coverage weighting.
func CalcSurvivalFull(blameLines []git.BlameLine, tau float64, now time.Time, pressure ChangePressure, threshold float64, tested *TestedSet, untestedWeight float64) SurvivalResult {
	return calcSurvivalImpl(blameLines, tau, now, pressure, threshold, tested, untestedWeight)
}

func calcSurvivalImpl(blameLines []git.BlameLine, tau float64, now time.Time, pressure ChangePressure, threshold float64, tested *TestedSet, untestedWeight float64) SurvivalResult {
	decayed := make(map[string]float64)
	raw := make(map[string]float64)
	robust := make(map[string]float64)
	dormant := make(map[string]float64)
	testedSurv := make(map[string]float64)
	untestedSurv := make(map[string]float64)

	usePressure := pressure != nil
	useTested := tested != nil && untestedWeight != 1.0
	if untestedWeight < 0 {
		untestedWeight = 0
	}

	for _, bl := range blameLines {
		raw[bl.Author]++

		daysAlive := now.Sub(bl.CommitterTime).Hours() / 24
		if daysAlive < 0 {
			daysAlive = 0
		}
		weight := math.Exp(-daysAlive / tau)

		// Determine per-line test-coverage multiplier. When tests aren't being
		// considered, every line weighs fully.
		isTested := false
		if tested != nil {
			isTested = tested.IsTested(bl.Filename)
			if isTested {
				testedSurv[bl.Author] += weight
			} else {
				untestedSurv[bl.Author] += weight
			}
		}
		coverageMult := 1.0
		if useTested && !isTested {
			coverageMult = untestedWeight
		}

		effective := weight * coverageMult
		decayed[bl.Author] += effective

		if usePressure {
			mod := ModuleOf(bl.Filename)
			if pressure[mod] >= threshold {
				robust[bl.Author] += effective
			} else {
				dormant[bl.Author] += effective
			}
		}
	}

	result := SurvivalResult{Decayed: decayed, Raw: raw}
	if usePressure {
		result.Robust = robust
		result.Dormant = dormant
	}
	if tested != nil {
		result.Tested = testedSurv
		result.Untested = untestedSurv
	}
	return result
}
