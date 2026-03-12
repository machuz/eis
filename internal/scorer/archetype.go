package scorer

// ArchetypeMatch holds an archetype name and its confidence score (0.0-1.0).
type ArchetypeMatch struct {
	Name       string
	Confidence float64
}

// classifyArchetypeWithConfidence scores all archetypes and returns primary + secondary.
func classifyArchetypeWithConfidence(r Result) (primary ArchetypeMatch, secondary ArchetypeMatch) {
	type rule struct {
		name  string
		score func() float64
	}

	// Helpers: how strongly a value matches "high" (>=60) or "low" (<30)
	// Returns 0.0-1.0 as a soft match
	highness := func(v float64) float64 {
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
	lowness := func(v float64) float64 {
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
	notLow := func(v float64) float64 {
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
	minf := func(vals ...float64) float64 {
		m := vals[0]
		for _, v := range vals[1:] {
			if v < m {
				m = v
			}
		}
		return m
	}
	// average of values
	avgf := func(vals ...float64) float64 {
		sum := 0.0
		for _, v := range vals {
			sum += v
		}
		return sum / float64(len(vals))
	}

	rules := []rule{
		{"Architect", func() float64 {
			return minf(highness(r.Production), highness(r.Survival), highness(r.Design))
		}},
		{"Former Architect", func() float64 {
			return minf(highness(r.RawSurvival), lowness(r.Survival),
				maxf(highness(r.Design), highness(r.Indispensability)))
		}},
		{"Churn Producer", func() float64 {
			if r.Production-r.Survival < 30 {
				return 0
			}
			return minf(notLow(r.Production), lowness(r.Quality), lowness(r.Survival))
		}},
		{"Rescue Producer", func() float64 {
			return minf(highness(r.Production), lowness(r.Survival), highness(r.DebtCleanup))
		}},
		{"Mass Producer", func() float64 {
			return minf(highness(r.Production), lowness(r.Survival))
		}},
		{"Solid Cleaner", func() float64 {
			return minf(highness(r.Quality), highness(r.Survival), highness(r.DebtCleanup))
		}},
		{"Spreader", func() float64 {
			return minf(highness(r.Breadth), lowness(r.Production), lowness(r.Survival), lowness(r.Design))
		}},
		{"Silent Killer", func() float64 {
			if r.TotalCommits < 100 {
				return 0
			}
			return minf(lowness(r.Production), lowness(r.Survival), lowness(r.DebtCleanup))
		}},
		{"Specialist", func() float64 {
			return minf(highness(r.Survival), lowness(r.Breadth))
		}},
		{"Quality Anchor", func() float64 {
			return minf(highness(r.Quality), notLow(r.Production))
		}},
		{"Growing", func() float64 {
			return minf(lowness(r.Production), highness(r.Quality))
		}},
	}

	// Score all rules, tracking rule index for priority
	type match struct {
		ArchetypeMatch
		priority int // lower = higher priority (rule definition order)
	}
	var matches []match
	for i, rule := range rules {
		score := rule.score()
		if score >= 0.10 {
			matches = append(matches, match{ArchetypeMatch{rule.name, score}, i})
		}
	}

	// Sort by confidence descending, but use rule priority as tiebreaker
	// when scores are within a margin (0.15). This ensures more specific
	// archetypes (defined earlier) win over generic ones at similar confidence.
	const priorityMargin = 0.15
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			swap := false
			diff := matches[j].Confidence - matches[i].Confidence
			if diff > priorityMargin {
				// j is clearly better → swap
				swap = true
			} else if abs(diff) <= priorityMargin && matches[j].priority < matches[i].priority {
				// similar confidence → prefer higher priority (lower index)
				swap = true
			}
			if swap {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	if len(matches) == 0 {
		// Fallback
		name := "—"
		if r.Total >= 40 {
			name = "Solid"
		} else if r.Total >= 30 {
			name = "Balanced"
		}
		return ArchetypeMatch{Name: name, Confidence: avgf(0.3)}, ArchetypeMatch{}
	}

	primary = matches[0].ArchetypeMatch

	// Round confidence
	primary.Confidence = roundConf(primary.Confidence)

	if len(matches) > 1 {
		secondary = matches[1].ArchetypeMatch
		secondary.Confidence = roundConf(secondary.Confidence)
	}

	return primary, secondary
}

// classifyArchetype returns just the archetype name (backward compatible).
func classifyArchetype(r Result) string {
	p, _ := classifyArchetypeWithConfidence(r)
	return p.Name
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
