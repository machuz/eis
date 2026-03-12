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
	// High design + indispensability + low survival = the hole hasn't been filled
	if high(r.Design) && high(r.Indispensability) && low(r.Survival) {
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

	// Drifter: high breadth but low production and design (wide but shallow)
	if high(r.Breadth) && low(r.Production) && low(r.Design) {
		return "Drifter"
	}

	// Specialist: high survival but narrow scope (deep in one area)
	if high(r.Survival) && low(r.Breadth) {
		return "Specialist"
	}

	// Growing: low production but high quality (learning, writing carefully)
	if low(r.Production) && high(r.Quality) {
		return "Growing"
	}

	// Fallback based on total score
	if r.Total >= 40 {
		return "Solid"
	}
	return "—"
}
