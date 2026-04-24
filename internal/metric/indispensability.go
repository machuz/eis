package metric

import (
	"math"
	"path/filepath"
	"sort"
	"strings"

	"github.com/machuz/eis/v2/internal/git"
)

type ModuleRisk struct {
	Module    string
	TopAuthor string
	Share     float64
	Level     string // "CRITICAL" or "HIGH"
}

func CalcIndispensability(blameLines []git.BlameLine, criticalThreshold, highThreshold float64) (map[string]float64, []ModuleRisk) {
	// Group blame lines by module (top-level directory)
	moduleAuthors := make(map[string]map[string]int) // module -> author -> count

	for _, bl := range blameLines {
		module := getModule(bl.Filename)
		if module == "" {
			continue
		}

		if _, ok := moduleAuthors[module]; !ok {
			moduleAuthors[module] = make(map[string]int)
		}
		moduleAuthors[module][bl.Author]++
	}

	// Calculate indispensability per author
	criticalModules := make(map[string]int)
	highModules := make(map[string]int)
	var risks []ModuleRisk

	for module, authors := range moduleAuthors {
		total := 0
		topAuthor := ""
		topCount := 0

		for author, count := range authors {
			total += count
			if count > topCount {
				topCount = count
				topAuthor = author
			}
		}

		if total == 0 {
			continue
		}

		share := float64(topCount) / float64(total)

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
// Uses ModuleOf() (3-level path) for consistency with ChangePressure.
//
// This complements CalcIndispensability: Indispensability measures person-level
// risk ("this person owns too much"), while OwnershipFragmentation measures
// module-level risk ("this module's knowledge is too concentrated/scattered").
func CalcOwnershipFragmentation(blameLines []git.BlameLine) []ModuleOwnership {
	// Group blame lines by module → author → count
	moduleAuthors := make(map[string]map[string]int)

	for _, bl := range blameLines {
		mod := ModuleOf(bl.Filename)
		if _, ok := moduleAuthors[mod]; !ok {
			moduleAuthors[mod] = make(map[string]int)
		}
		moduleAuthors[mod][bl.Author]++
	}

	var results []ModuleOwnership
	for mod, authors := range moduleAuthors {
		total := 0
		topAuthor := ""
		topCount := 0

		for author, count := range authors {
			total += count
			if count > topCount {
				topCount = count
				topAuthor = author
			}
		}

		if total == 0 {
			continue
		}

		topShare := float64(topCount) / float64(total)

		// Shannon entropy: H = -Σ p·log₂(p)
		// Higher entropy = more distributed ownership
		entropy := 0.0
		for _, count := range authors {
			p := float64(count) / float64(total)
			if p > 0 {
				entropy -= p * math.Log2(p)
			}
		}

		level := classifyOwnership(topShare, len(authors))

		results = append(results, ModuleOwnership{
			Module:      mod,
			TotalLines:  total,
			AuthorCount: len(authors),
			TopAuthor:   topAuthor,
			TopShare:    topShare,
			Entropy:     entropy,
			Level:       level,
		})
	}

	// Sort by TopShare descending (most concentrated first = highest risk)
	sort.Slice(results, func(i, j int) bool {
		return results[i].TopShare > results[j].TopShare
	})

	return results
}

// classifyOwnership determines the ownership health level of a module.
//
//   SOLE_OWNER:    1 author — bus factor = 1, structural collapse risk
//   CONCENTRATED:  top author ≥ 80% — effectively sole owner
//   HEALTHY:       top author 40-80% — distributed with clear ownership
//   FRAGMENTED:    top author < 40% — no clear owner, coordination risk
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

func getModule(filename string) string {
	parts := strings.Split(filepath.ToSlash(filename), "/")
	if len(parts) < 2 {
		return ""
	}
	// Use first two path components as module (e.g., "app/domain")
	if len(parts) >= 3 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}
