package metric

import (
	"math"
	"sort"

	"github.com/machuz/eis/v2/internal/git"
)

type ModuleRisk struct {
	Module    string
	TopAuthor string
	Share     float64
	Level     string // "CRITICAL" or "HIGH"
}

func CalcIndispensability(blameLines []git.BlameLine, mr ModuleResolver, criticalThreshold, highThreshold float64) (map[string]float64, []ModuleRisk) {
	// Group blame lines by module via the shared resolver, so module_patterns
	// and ExcludeModules apply consistently with the other module metrics.
	moduleAuthors := make(map[string]map[string]float64) // module -> author -> owned lines (co-author-split)

	for _, bl := range blameLines {
		module := mr.ModuleOf(bl.Filename)
		if module == "" || module == PeripheralModule {
			// Skip root-level/excluded files, and the peripheral denoising
			// bucket: it aggregates unrelated low-liveness fallback modules, so
			// its ownership concentration is not a real key-person risk and must
			// not inflate any author's Indispensability.
			continue
		}

		if _, ok := moduleAuthors[module]; !ok {
			moduleAuthors[module] = make(map[string]float64)
		}
		blameShares(bl, func(a string, s float64) { moduleAuthors[module][a] += s })
	}

	// Calculate indispensability per author
	criticalModules := make(map[string]int)
	highModules := make(map[string]int)
	var risks []ModuleRisk

	for module, authors := range moduleAuthors {
		total := 0.0
		topAuthor := ""
		topCount := 0.0

		for author, count := range authors {
			total += count
			// Tie-break by author name so a module owned equally by two authors
			// picks the same top author on every run (map iteration order is
			// nondeterministic) — otherwise the credited owner, and the scores
			// derived from it, would vary run-to-run.
			if count > topCount || (count == topCount && (topAuthor == "" || author < topAuthor)) {
				topCount = count
				topAuthor = author
			}
		}

		if total == 0 {
			continue
		}

		share := topCount / total

		if share >= criticalThreshold {
			criticalModules[topAuthor]++
			risks = append(risks, ModuleRisk{
				Module:    module,
				TopAuthor: topAuthor,
				Share:     share,
				Level:     "CRITICAL",
			})
		} else if share >= highThreshold {
			highModules[topAuthor]++
			risks = append(risks, ModuleRisk{
				Module:    module,
				TopAuthor: topAuthor,
				Share:     share,
				Level:     "HIGH",
			})
		}
	}

	// Deterministic risk ordering: risks are appended in map-iteration order
	// above, so sort them (highest share first, then module name) before they
	// reach output.
	sort.Slice(risks, func(i, j int) bool {
		if risks[i].Share != risks[j].Share {
			return risks[i].Share > risks[j].Share
		}
		return risks[i].Module < risks[j].Module
	})

	result := make(map[string]float64)
	allAuthors := make(map[string]bool)
	for a := range criticalModules {
		allAuthors[a] = true
	}
	for a := range highModules {
		allAuthors[a] = true
	}

	for author := range allAuthors {
		result[author] = float64(criticalModules[author])*1.0 + float64(highModules[author])*0.5
	}

	return result, risks
}

// ModuleOwnership represents the ownership distribution of a single module.
// This is the structural inverse of Indispensability: instead of asking
// "how indispensable is this person?", it asks "how is this module's
// knowledge distributed?"
//
// A module with SOLE_OWNER or CONCENTRATED ownership is a structural risk —
// if that person leaves, the module enters Dead Zone.
// A module with HEALTHY ownership has distributed knowledge — resilient.
// A module with FRAGMENTED ownership has no clear owner — coordination risk.
type ModuleOwnership struct {
	Module      string
	TotalLines  int
	AuthorCount int
	TopAuthor   string
	TopShare    float64 // 0.0-1.0
	Entropy     float64 // Shannon entropy (higher = more distributed)
	Level       string  // "SOLE_OWNER", "CONCENTRATED", "HEALTHY", "FRAGMENTED"
}

// CalcOwnershipFragmentation analyzes blame-line distribution per module.
// Uses the convention-aware ModuleResolver for consistency with ChangePressure.
//
// This complements CalcIndispensability: Indispensability measures person-level
// risk ("this person owns too much"), while OwnershipFragmentation measures
// module-level risk ("this module's knowledge is too concentrated/scattered").
func CalcOwnershipFragmentation(blameLines []git.BlameLine, mr ModuleResolver) []ModuleOwnership {
	// Group blame lines by module → author → owned lines (co-author-split)
	moduleAuthors := make(map[string]map[string]float64)

	for _, bl := range blameLines {
		mod := mr.ModuleOf(bl.Filename)
		if mod == "" || mod == PeripheralModule {
			continue // excluded module (ExcludeModules) or peripheral bucket
		}
		if _, ok := moduleAuthors[mod]; !ok {
			moduleAuthors[mod] = make(map[string]float64)
		}
		blameShares(bl, func(a string, s float64) { moduleAuthors[mod][a] += s })
	}

	var results []ModuleOwnership
	for mod, authors := range moduleAuthors {
		total := 0.0
		topAuthor := ""
		topCount := 0.0

		for author, count := range authors {
			total += count
			// Tie-break by author name so a module owned equally by two authors
			// picks the same top author on every run (map iteration order is
			// nondeterministic) — otherwise the credited owner, and the scores
			// derived from it, would vary run-to-run.
			if count > topCount || (count == topCount && (topAuthor == "" || author < topAuthor)) {
				topCount = count
				topAuthor = author
			}
		}

		if total == 0 {
			continue
		}

		topShare := topCount / total

		// Shannon entropy: H = -Σ p·log₂(p)
		// Higher entropy = more distributed ownership
		entropy := 0.0
		for _, count := range authors {
			p := count / total
			if p > 0 {
				entropy -= p * math.Log2(p)
			}
		}

		level := classifyOwnership(topShare, len(authors))

		results = append(results, ModuleOwnership{
			Module:      mod,
			TotalLines:  int(math.Round(total)),
			AuthorCount: len(authors),
			TopAuthor:   topAuthor,
			TopShare:    topShare,
			Entropy:     entropy,
			Level:       level,
		})
	}

	// Sort by TopShare descending (most concentrated first = highest risk)
	sort.Slice(results, func(i, j int) bool {
		if results[i].TopShare != results[j].TopShare {
			return results[i].TopShare > results[j].TopShare
		}
		return results[i].Module < results[j].Module
	})

	return results
}

// classifyOwnership determines the ownership health level of a module.
//
//	SOLE_OWNER:    1 author — bus factor = 1, structural collapse risk
//	CONCENTRATED:  top author ≥ 80% — effectively sole owner
//	HEALTHY:       top author 40-80% — distributed with clear ownership
//	FRAGMENTED:    top author < 40% — no clear owner, coordination risk
func classifyOwnership(topShare float64, authorCount int) string {
	if authorCount == 1 {
		return "SOLE_OWNER"
	}
	if topShare >= 0.80 {
		return "CONCENTRATED"
	}
	if topShare >= 0.40 {
		return "HEALTHY"
	}
	return "FRAGMENTED"
}
