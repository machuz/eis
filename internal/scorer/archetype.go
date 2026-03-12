package scorer

// classifyArchetype determines an engineer's archetype based on their 7-axis scores.
// Thresholds: high >= 60, mid >= 30, low < 30
// Rules are ordered by specificity (most specific first).
func classifyArchetype(r Result) string {
	high := func(v float64) bool { return v >= 60 }
	low := func(v float64) bool { return v < 30 }

	// Architect: high production + survival + design (the backbone)
	if high(r.Production) && high(r.Survival) && high(r.Design) {
		return "Architect"
	}

	// Former Architect: designed the system but no longer active
	// High raw survival (code still in codebase) + low decayed survival (inactive) = departed
	// Combined with high design or indispensability = was an architect
	if high(r.RawSurvival) && low(r.Survival) && (high(r.Design) || high(r.Indispensability)) {
		return "Former Architect"
	}

	// Mass Producer: high production but code doesn't survive (debt factory)
	if high(r.Production) && low(r.Survival) {
		return "Mass Producer"
	}

	// Solid Cleaner: high quality + survival + debt cleanup (the quiet hero)
	if high(r.Quality) && high(r.Survival) && high(r.DebtCleanup) {
		return "Solid Cleaner"
	}

	// Spreader: high breadth but low production and design (wide but shallow)
	if high(r.Breadth) && low(r.Production) && low(r.Design) {
		return "Spreader"
	}

	// Silent Killer: low production, low survival, low debt cleanup (net drain on team)
	if low(r.Production) && low(r.Survival) && low(r.DebtCleanup) {
		return "Silent Killer"
	}

	// Specialist: high survival but narrow scope (deep in one area)
	if high(r.Survival) && low(r.Breadth) {
		return "Specialist"
	}

	// Quality Anchor: high quality + mid production — stable and reliable, but not yet shaping design
	if high(r.Quality) && !low(r.Production) {
		return "Quality Anchor"
	}

	// Growing: low production but high quality (learning, writing carefully)
	if low(r.Production) && high(r.Quality) {
		return "Growing"
	}

	// Fallback based on total score
	if r.Total >= 40 {
		return "Solid"
	}
	if r.Total >= 30 {
		return "Balanced"
	}
	return "—"
}
