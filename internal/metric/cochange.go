package metric

import (
	"sort"

	"github.com/machuz/eis/v2/internal/git"
)

// ModulePair represents a co-change relationship between two modules.
// When two modules frequently appear in the same commit, they have
// implicit coupling — a structural dependency invisible in import graphs.
type ModulePair struct {
	ModuleA       string
	ModuleB       string
	CochangeCount int     // number of commits touching both modules
	Coupling      float64 // Jaccard coefficient: co-change / union
}

// CochangeResult holds the complete co-change analysis.
type CochangeResult struct {
	Pairs         []ModulePair   // sorted by coupling descending
	ModuleCommits map[string]int // module → total commits touching it
}

// CalcCochange computes co-change coupling between modules.
//
// For each commit, it identifies which modules are touched, then counts
// how often each pair of modules appears together. The coupling coefficient
// uses the Jaccard index: co-change / (commits_A + commits_B - co-change).
//
// This is effectively an auto-generated Design Structure Matrix (DSM),
// the modular design framework proposed by Baldwin & Clark.
// High coupling between modules that should be independent signals
// a leaky boundary — a structural risk invisible from code alone.
func CalcCochange(commits []git.Commit) CochangeResult {
	// Count commits per module (a commit counts once per module it touches)
	moduleCommits := make(map[string]int)

	// Count co-occurrences per module pair
	pairCount := make(map[[2]string]int)

	for _, c := range commits {
		// Unique modules touched by this commit
		touched := make(map[string]bool)
		for _, fs := range c.FileStats {
			mod := ModuleOf(fs.Filename)
			touched[mod] = true
		}

		mods := make([]string, 0, len(touched))
		for mod := range touched {
			mods = append(mods, mod)
			moduleCommits[mod]++
		}

		// Count all pairs (canonical order for deterministic keys)
		sort.Strings(mods)
		for i := 0; i < len(mods); i++ {
			for j := i + 1; j < len(mods); j++ {
				key := [2]string{mods[i], mods[j]}
				pairCount[key]++
			}
		}
	}

	// Calculate Jaccard coupling coefficient
	var pairs []ModulePair
	for key, count := range pairCount {
		// Jaccard: intersection / union = co-change / (A + B - co-change)
		union := moduleCommits[key[0]] + moduleCommits[key[1]] - count
		coupling := 0.0
		if union > 0 {
			coupling = float64(count) / float64(union)
		}

		pairs = append(pairs, ModulePair{
			ModuleA:       key[0],
			ModuleB:       key[1],
			CochangeCount: count,
			Coupling:      coupling,
		})
	}

	// Sort by coupling descending
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Coupling == pairs[j].Coupling {
			return pairs[i].CochangeCount > pairs[j].CochangeCount
		}
		return pairs[i].Coupling > pairs[j].Coupling
	})

	return CochangeResult{
		Pairs:         pairs,
		ModuleCommits: moduleCommits,
	}
}
