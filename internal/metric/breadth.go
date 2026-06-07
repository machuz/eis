package metric

import "math"

// ComputeBreadth derives the per-author raw Breadth score as the effective
// number of modules over which an author holds surviving gravity — a Hill
// number (q=1, exp of Shannon entropy) over the per-module shares of their
// time-decayed surviving blame mass.
//
// Why this definition: the old Breadth counted distinct repos (or modules)
// with ≥N commits. That had two independent flaws. The unit (repo) was a
// packaging decision, so reorganising the same code from one monorepo into
// many repos — a no-op — changed Breadth (gameable, and it collapsed to {0,1}
// inside a monorepo). And the measure (commit presence) ignored survival, the
// one axis you could inflate by spreading thin commits. Keying on the module
// topology makes Breadth packaging-invariant; weighting by surviving gravity
// makes a drive-by touch worth ~nothing; and the Hill number turns "how many
// modules" into "how many modules they *effectively* hold" — a specialist with
// all gravity in one module scores ≈1, someone spread evenly across five
// scores ≈5, and 80%-in-one-plus-scraps scores ≈1.3. See
// docs/eis/breadth-redefinition.md (in the orbit repo).
//
// Input is module → author → surviving gravity mass (the shape produced by
// CalcModuleSurvivalByAuthor). Authors with zero total gravity are omitted, so
// the result map matches the prior "only set when non-zero" contract.
//
// One shared function for both the analyzer and timeline pipelines: two copies
// of the breadth loop drift, and the same git history would then yield
// different Breadth depending on which pipeline observed it (W-02 violation).
func ComputeBreadth(moduleSurvivalByAuthor map[string]map[string]float64) map[string]float64 {
	// Transpose to author → []mass so each author's distribution is in hand.
	authorMasses := make(map[string][]float64)
	for _, authors := range moduleSurvivalByAuthor {
		for author, mass := range authors {
			if mass <= 0 {
				continue
			}
			authorMasses[author] = append(authorMasses[author], mass)
		}
	}

	result := make(map[string]float64, len(authorMasses))
	for author, masses := range authorMasses {
		var total float64
		for _, m := range masses {
			total += m
		}
		if total <= 0 {
			continue
		}
		var entropy float64
		for _, m := range masses {
			p := m / total
			entropy -= p * math.Log(p)
		}
		result[author] = math.Exp(entropy)
	}
	return result
}
