package scorer

// AxisMatch holds a classification label and its confidence score (0.0-1.0).
type AxisMatch struct {
	Name       string
	Confidence float64
}

// classifyTopology returns 3-axis classification: Role, Style, State.
func classifyTopology(r Result) (role AxisMatch, style AxisMatch, state AxisMatch) {
	role = classifyRole(r)
	style = classifyStyle(r)
	state = classifyState(r)
	return
}

// --- Role: what they contribute to the team ---

func classifyRole(r Result) AxisMatch {
	rules := []classifyRule{
		// Architect: high design influence with durable code under pressure.
		// When pressure data exists, use robust survival — but high production
		// overrides to total survival (the architect's own changes replace their
		// code, creating a false low robust score).
		{"Architect", func() float64 {
			surv := r.Survival
			if r.DormantSurvival > 0 || r.RobustSurvival > 0 {
				if highness(r.Production) > 0 {
					// High production = active builder whose own changes create pressure.
					// Total survival is the fair measure.
					surv = r.Survival
				} else {
					surv = r.RobustSurvival
				}
			}
			return minf(highness(r.Design), highness(surv), notLow(r.Breadth))
		}},
		// Anchor: reliable quality contributor, not yet shaping design.
		{"Anchor", func() float64 {
			return minf(highness(r.Quality), notLow(r.Production))
		}},
		// Cleaner: high quality, high survival, high debt cleanup.
		{"Cleaner", func() float64 {
			return minf(highness(r.Quality), highness(r.Survival), highness(r.DebtCleanup))
		}},
		// Producer: meaningful production output.
		{"Producer", func() float64 {
			return notLow(r.Production)
		}},
		// Specialist: deep in narrow area, high survival but low breadth.
		{"Specialist", func() float64 {
			return minf(highness(r.Survival), lowness(r.Breadth))
		}},
	}

	return pickBest(rules, 0.10)
}

// --- Style: how they contribute ---

func classifyStyle(r Result) AxisMatch {
	rules := []classifyRule{
		// Builder: designs, builds heavily, AND cleans up. The full package.
		// Production gate filters solo-inflated non-builders.
		{"Builder", func() float64 {
			return minf(highness(r.Production), highness(r.Design), notLow(r.DebtCleanup))
		}},
		// Resilient: iterates heavily but what survives under pressure is durable.
		{"Resilient", func() float64 {
			if r.RobustSurvival == 0 {
				return 0
			}
			return minf(highness(r.Production), lowness(r.Survival), notLow(r.RobustSurvival))
		}},
		// Rescue: high output cleaning up others' legacy code.
		{"Rescue", func() float64 {
			return minf(highness(r.Production), lowness(r.Survival), highness(r.DebtCleanup))
		}},
		// Churn: high output but terrible quality, constant rework.
		{"Churn", func() float64 {
			if r.Production-r.Survival < 30 {
				return 0
			}
			return minf(notLow(r.Production), lowness(r.Quality), lowness(r.Survival))
		}},
		// Mass: high output but code doesn't survive.
		{"Mass", func() float64 {
			return minf(highness(r.Production), lowness(r.Survival))
		}},
		// Emergent: creating new structural gravity that hasn't been battle-tested.
		// High gravity (wide influence + ownership) + meaningful production + low robust survival.
		// This signals a future Architect candidate whose structures are still being
		// challenged and refined by the team — the creative friction before convergence.
		{"Emergent", func() float64 {
			if r.Gravity < 50 || r.Production < 30 {
				return 0
			}
			robustLow := lowness(r.RobustSurvival) // code doesn't survive pressure yet
			if robustLow == 0 {
				return 0 // robust is high → already proven, not "emergent"
			}
			return minf(highness(r.Gravity), notLow(r.Production), robustLow)
		}},
		// Balanced: steady contributor, no dominant pattern.
		{"Balanced", func() float64 {
			if r.Impact < 30 {
				return 0
			}
			return 0.30
		}},
		// Spread: wide presence, low depth everywhere.
		{"Spread", func() float64 {
			return minf(highness(r.Breadth), lowness(r.Production), lowness(r.Survival), lowness(r.Design))
		}},
	}

	return pickBest(rules, 0.10)
}

// --- State: lifecycle phase ---

func classifyState(r Result) AxisMatch {
	rules := []classifyRule{
		// Former: code still in codebase (high raw) but no longer active (low decayed).
		{"Former", func() float64 {
			return minf(highness(r.RawSurvival), lowness(r.Survival),
				maxf(highness(r.Design), highness(r.Indispensability)))
		}},
		// Silent: neither builds nor cleans — net drain. Requires ≥100 commits.
		{"Silent", func() float64 {
			if r.TotalCommits < 100 {
				return 0
			}
			return minf(lowness(r.Production), lowness(r.Survival), lowness(r.DebtCleanup))
		}},
		// Fragile: code survives only because no one touches it.
		// Stacks two independent "untouched" signals:
		//  - dormant by change pressure (nothing edits the module)
		//  - untested by test coverage (nothing guards it)
		// Both high → highest confidence. Either alone → weaker Fragile.
		// Solo ownership (Indispensability ≥ 60) and low Production remain
		// gating conditions so active-but-dormant builders aren't flagged.
		{"Fragile", func() float64 {
			if r.Indispensability < 60 || r.Production >= 40 {
				// Not a solo fossil: fall through to the fallback block.
			} else {
				dormantRatio, hasPressure := computeDormantRatio(r)
				untestedRatio, hasCoverage := computeUntestedRatio(r)

				if hasPressure && hasCoverage && dormantRatio >= 80 && untestedRatio >= 50 {
					// Both signals align: fossil + unguarded. Boost above the
					// dormant-only ceiling to distinguish in downstream display.
					// Cap at 0.97 so the boosted band stays distinct from
					// certainty-claiming 1.00 values elsewhere in the codebase.
					boost := minf((dormantRatio-80)/200, (untestedRatio-50)/500)
					if boost > 0.07 {
						boost = 0.07
					}
					return 0.90 + boost // 0.90–0.97
				}
				if hasPressure && dormantRatio >= 80 {
					return 0.85 + (dormantRatio-80)/200 // 0.85–0.95
				}
				// Untested-only fallback: only when pressure data is truly
				// unavailable. Active-but-untested code in a busy module is
				// NOT Fragile — the reviewer flagged this false-positive path.
				if !hasPressure && hasCoverage && untestedRatio >= 70 && r.Survival >= 50 {
					return 0.80 + (untestedRatio-70)/300 // 0.80–0.90
				}
			}
			// Final fallback: no pressure, no coverage, or gating conditions failed.
			if r.Quality >= 70 {
				return 0
			}
			return minf(highness(r.Survival), lowness(r.Production))
		}},
		// Growing: low volume, high quality — on a growth trajectory.
		{"Growing", func() float64 {
			return minf(lowness(r.Production), highness(r.Quality))
		}},
		// Active: recently contributing.
		{"Active", func() float64 {
			if r.RecentlyActive {
				return 0.80
			}
			return 0
		}},
	}

	return pickBest(rules, 0.10)
}

// --- shared helpers ---

type classifyRule struct {
	name  string
	score func() float64
}

func pickBest(rules []classifyRule, minConf float64) AxisMatch {
	type match struct {
		AxisMatch
		priority int
	}

	var matches []match
	for i, rule := range rules {
		score := rule.score()
		if score >= minConf {
			matches = append(matches, match{AxisMatch{rule.name, score}, i})
		}
	}

	if len(matches) == 0 {
		return AxisMatch{Name: "—", Confidence: 0}
	}

	// Sort: confidence descending, priority as tiebreaker within margin.
	const priorityMargin = 0.15
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			swap := false
			diff := matches[j].Confidence - matches[i].Confidence
			if diff > priorityMargin {
				swap = true
			} else if abs(diff) <= priorityMargin && matches[j].priority < matches[i].priority {
				swap = true
			}
			if swap {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	best := matches[0].AxisMatch
	best.Confidence = roundConf(best.Confidence)
	return best
}

// Helpers: how strongly a value matches "high" (>=60) or "low" (<30)
// Returns 0.0-1.0 as a soft match
func highness(v float64) float64 {
	if v >= 80 {
		return 1.0
	}
	if v >= 60 {
		return 0.5 + (v-60)/40
	}
	if v >= 40 {
		return (v - 40) / 40 * 0.3
	}
	return 0
}

func lowness(v float64) float64 {
	if v < 10 {
		return 1.0
	}
	if v < 30 {
		return 0.5 + (30-v)/40
	}
	if v < 50 {
		return (50 - v) / 40 * 0.3
	}
	return 0
}

// notLow: 1.0 if >= 50, ramps from 0 at 10 to 1.0 at 50
func notLow(v float64) float64 {
	if v >= 50 {
		return 1.0
	}
	if v >= 30 {
		return 0.5 + (v-30)/40
	}
	if v >= 10 {
		return (v - 10) / 40 * 0.3
	}
	return 0
}

// min of values
func minf(vals ...float64) float64 {
	m := vals[0]
	for _, v := range vals[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func roundConf(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

// computeDormantRatio returns the percentage of raw survival that falls in
// low-pressure modules. Returns (ratio, true) when pressure data exists,
// otherwise (0, false).
func computeDormantRatio(r Result) (float64, bool) {
	total := r.RawDormantSurv + r.RawRobustSurv
	if total <= 0 {
		return 0, false
	}
	return r.RawDormantSurv / total * 100, true
}

// computeUntestedRatio returns the percentage of raw survival that comes
// from files without test coverage. Returns (ratio, true) when we have
// any survival data to compute from, otherwise (0, false).
func computeUntestedRatio(r Result) (float64, bool) {
	total := r.RawTestedSurv + r.RawUntestedSurv
	if total <= 0 {
		return 0, false
	}
	return r.RawUntestedSurv / total * 100, true
}
