package metric

import (
	"math"
	"time"

	"github.com/machuz/eis/v2/internal/git"
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
	// No pressure split → ModuleResolver is unused; pass the zero value.
	return calcSurvivalImpl(blameLines, tau, now, nil, 0, ModuleResolver{}, nil, 1.0, nil)
}

// CalcSurvivalWithPressure splits survival into robust (high-pressure modules)
// and dormant (low-pressure modules) based on change pressure threshold.
// mr must use the same convention set as the resolver that built `pressure`,
// so the module keys line up.
func CalcSurvivalWithPressure(blameLines []git.BlameLine, tau float64, now time.Time, pressure ChangePressure, threshold float64, mr ModuleResolver, others *OthersPressure) SurvivalResult {
	return calcSurvivalImpl(blameLines, tau, now, pressure, threshold, mr, nil, 1.0, others)
}

// CalcSurvivalFull is the comprehensive survival calculator. It produces the
// pressure-based split (Robust / Dormant) AND the test-coverage split
// (Tested / Untested) in a single pass, applying untestedWeight to the
// effective Decayed/Robust/Dormant values so that untested code contributes
// proportionally less to Impact.
//
// Pass nil `tested` (or untestedWeight=1.0) to skip the test-coverage weighting.
// mr must use the same convention set as the resolver that built `pressure`.
// When `others` is non-nil the robust/dormant split uses OTHERS-contested
// pressure (others.For(mod, author)) instead of the module's total pressure, so
// self-churn and untouched dead corners do not manufacture robust survival.
func CalcSurvivalFull(blameLines []git.BlameLine, tau float64, now time.Time, pressure ChangePressure, threshold float64, mr ModuleResolver, tested *TestedSet, untestedWeight float64, others *OthersPressure) SurvivalResult {
	return calcSurvivalImpl(blameLines, tau, now, pressure, threshold, mr, tested, untestedWeight, others)
}

func calcSurvivalImpl(blameLines []git.BlameLine, tau float64, now time.Time, pressure ChangePressure, threshold float64, mr ModuleResolver, tested *TestedSet, untestedWeight float64, others *OthersPressure) SurvivalResult {
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

	// accrue applies one line's contribution to a single author, scaled by share.
	// Defined once (captures only the stable maps/flags) so co-author splitting adds
	// no per-line allocation. For a co-authored line share = 1/N; for a solo line
	// share = 1. The robust/dormant split reads OTHERS-pressure per author, so each
	// co-author's robust survival correctly excludes their own churn.
	accrue := func(a string, share, weight, effective float64, isTested bool, mod string) {
		raw[a] += share
		if tested != nil {
			if isTested {
				testedSurv[a] += weight * share
			} else {
				untestedSurv[a] += weight * share
			}
		}
		decayed[a] += effective * share
		if usePressure {
			p := pressure[mod]
			if others != nil {
				p = others.For(mod, a)
			}
			if p >= threshold {
				robust[a] += effective * share
			} else {
				dormant[a] += effective * share
			}
		}
	}

	for _, bl := range blameLines {
		daysAlive := now.Sub(bl.CommitterTime).Hours() / 24
		if daysAlive < 0 {
			daysAlive = 0
		}
		weight := math.Exp(-daysAlive / tau)

		isTested := false
		if tested != nil {
			isTested = tested.IsTested(bl.Filename)
		}
		coverageMult := 1.0
		if useTested && !isTested {
			coverageMult = untestedWeight
		}
		effective := weight * coverageMult

		mod := ""
		if usePressure {
			mod = mr.ModuleOf(bl.Filename)
		}

		// Split the line equally among the commit's contributor set (co-authors),
		// falling back to the single blame author for the common solo line.
		if len(bl.Authors) == 0 {
			accrue(bl.Author, 1.0, weight, effective, isTested, mod)
			continue
		}
		share := 1.0 / float64(len(bl.Authors))
		for _, a := range bl.Authors {
			accrue(a, share, weight, effective, isTested, mod)
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
