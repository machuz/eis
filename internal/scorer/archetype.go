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
		// Balanced: steady contributor, no dominant pattern.
		{"Balanced", func() float64 {
			if r.Total < 30 {
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
		// Fragile: code survives only because it's not under change pressure.
		{"Fragile", func() float64 {
			if r.Quality >= 70 {
				return 0
			}
			if r.DormantSurvival > 0 || r.RobustSurvival > 0 {
				return minf(highness(r.DormantSurvival), lowness(r.RobustSurvival), lowness(r.Production))
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
