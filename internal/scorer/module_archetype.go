package scorer

import (
	"sort"
	"time"

	"github.com/machuz/engineering-impact-score/internal/metric"
)

// ModuleScore holds 4 module indicators + 3-axis classification.
type ModuleScore struct {
	Module                string
	BoundaryIntegrity     float64 // 0-100: how clean the module boundary is
	ChangeAbsorption      float64 // 0-100: how well changes survive
	KnowledgeDistribution float64 // 0-100: how well ownership is distributed
	Stability             float64 // 0-100: how infrequently the module changes

	// Raw data
	ChangePressure    float64
	AvgCoupling       float64
	MaxCoupling       float64
	CouplingPairCount int
	AuthorCount       int
	BlameLines        int
	ModuleCommits     int
	OwnershipLevel    string
	TopAuthorShare    float64
	OwnerActive       bool

	// 3-axis classification
	Coupling      string  // "Isolated" / "Independent" / "Linked" / "Hub"
	CouplingConf  float64
	Vitality      string  // "Stable" / "Warming" / "Turbulent" / "Critical" / "Dead"
	VitalityConf  float64
	Ownership     string  // "Distributed" / "Concentrated" / "Orphaned"
	OwnershipConf float64
}

// IsAnomaly returns true if this module has at least one axis flagged as a risk.
func (ms ModuleScore) IsAnomaly() bool {
	switch ms.Coupling {
	case "Hub":
		return true
	}
	switch ms.Vitality {
	case "Turbulent", "Critical", "Dead":
		return true
	}
	switch ms.Ownership {
	case "Orphaned":
		return true
	}
	return false
}

// ScoreModules computes 4 module indicators and classifies each module on 3 axes.
func ScoreModules(
	pressure metric.ChangePressure,
	cochangeResults []metric.CochangeResult,
	ownership []metric.ModuleOwnership,
	moduleSurvival map[string]float64,
	authorLastDate map[string]time.Time,
	activeDays int,
) []ModuleScore {
	// Collect all modules from all data sources
	allModules := make(map[string]bool)
	for mod := range pressure {
		allModules[mod] = true
	}
	for mod := range moduleSurvival {
		allModules[mod] = true
	}
	for _, o := range ownership {
		allModules[o.Module] = true
	}

	// Merge module commits from all cochange results
	moduleCommits := make(map[string]int)
	for _, cr := range cochangeResults {
		for mod, count := range cr.ModuleCommits {
			moduleCommits[mod] += count
		}
		for mod := range cr.ModuleCommits {
			allModules[mod] = true
		}
	}

	// Build per-module coupling stats (average coupling across all pairs involving this module)
	type couplingStats struct {
		sum   float64
		max   float64
		count int
	}
	moduleCoupling := make(map[string]*couplingStats)
	for _, cr := range cochangeResults {
		for _, p := range cr.Pairs {
			for _, mod := range []string{p.ModuleA, p.ModuleB} {
				cs, ok := moduleCoupling[mod]
				if !ok {
					cs = &couplingStats{}
					moduleCoupling[mod] = cs
				}
				cs.sum += p.Coupling
				cs.count++
				if p.Coupling > cs.max {
					cs.max = p.Coupling
				}
			}
		}
	}

	// Build ownership lookup
	ownershipMap := make(map[string]*metric.ModuleOwnership)
	for i := range ownership {
		ownershipMap[ownership[i].Module] = &ownership[i]
	}

	// Compute percentile ranks for change pressure
	pressureRanks := percentileRanks(pressure)

	// Build scores
	var scores []ModuleScore
	for mod := range allModules {
		ms := ModuleScore{
			Module:         mod,
			ChangePressure: pressure[mod],
			ModuleCommits:  moduleCommits[mod],
		}

		// BoundaryIntegrity: (1 - avgCoupling) * 100
		if cs, ok := moduleCoupling[mod]; ok && cs.count > 0 {
			ms.AvgCoupling = cs.sum / float64(cs.count)
			ms.MaxCoupling = cs.max
			ms.CouplingPairCount = cs.count
		}
		ms.BoundaryIntegrity = (1 - ms.AvgCoupling) * 100

		// ChangeAbsorption: moduleSurvival * 100
		if surv, ok := moduleSurvival[mod]; ok {
			ms.ChangeAbsorption = surv * 100
		}

		// Stability: (1 - percentileRank) * 100
		if rank, ok := pressureRanks[mod]; ok {
			ms.Stability = (1 - rank) * 100
		} else {
			ms.Stability = 100 // no pressure data = stable
		}

		// KnowledgeDistribution + ownership raw data
		if o, ok := ownershipMap[mod]; ok {
			ms.AuthorCount = o.AuthorCount
			ms.BlameLines = o.TotalLines
			ms.OwnershipLevel = o.Level
			ms.TopAuthorShare = o.TopShare
			ms.KnowledgeDistribution = knowledgeScore(o.Level, o.Entropy)

			// Check if top author is active
			if lastDate, ok := authorLastDate[o.TopAuthor]; ok {
				daysSince := time.Since(lastDate).Hours() / 24
				ms.OwnerActive = daysSince <= float64(activeDays)
			}
		}

		// 3-axis classification
		coupling, couplingConf := classifyCoupling(ms)
		ms.Coupling = coupling
		ms.CouplingConf = couplingConf

		vitality, vitalityConf := classifyVitality(ms)
		ms.Vitality = vitality
		ms.VitalityConf = vitalityConf

		own, ownConf := classifyOwnershipAxis(ms, ownershipMap[mod], authorLastDate, activeDays)
		ms.Ownership = own
		ms.OwnershipConf = ownConf

		scores = append(scores, ms)
	}

	// Sort by Stability ascending (most unstable first = most interesting)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Stability < scores[j].Stability
	})

	return scores
}

// knowledgeScore converts ownership level + entropy to a 0-100 score.
func knowledgeScore(level string, entropy float64) float64 {
	base := 50.0
	switch level {
	case "HEALTHY":
		base = 80
	case "FRAGMENTED":
		base = 50
	case "CONCENTRATED":
		base = 25
	case "SOLE_OWNER":
		base = 10
	}
	// Entropy micro-adjustment: ±5 based on entropy (higher = more distributed)
	adj := (entropy - 1.5) * 2.5
	if adj > 5 {
		adj = 5
	}
	if adj < -5 {
		adj = -5
	}
	result := base + adj
	if result > 100 {
		result = 100
	}
	if result < 0 {
		result = 0
	}
	return result
}

// percentileRanks returns percentile rank (0-1) for each module's pressure.
func percentileRanks(pressure metric.ChangePressure) map[string]float64 {
	if len(pressure) == 0 {
		return nil
	}
	type kv struct {
		mod      string
		pressure float64
	}
	var sorted []kv
	for mod, p := range pressure {
		sorted = append(sorted, kv{mod, p})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].pressure < sorted[j].pressure
	})

	n := float64(len(sorted))
	ranks := make(map[string]float64, len(sorted))
	for i, item := range sorted {
		ranks[item.mod] = float64(i) / (n - 1)
	}
	if n == 1 {
		for mod := range ranks {
			ranks[mod] = 0.5
		}
	}
	return ranks
}

// classifyCoupling classifies a module's boundary quality.
// 4 levels: Isolated (no co-change pairs), Independent, Linked, Hub.
func classifyCoupling(ms ModuleScore) (string, float64) {
	// Hard gate: no co-change pairs at all → Isolated
	if ms.CouplingPairCount == 0 {
		return "Isolated", 0.95
	}

	avgPct := ms.AvgCoupling * 100 // 0-100 scale

	rules := []classifyRule{
		// Hub: high average coupling — central dependency point
		{"Hub", func() float64 {
			return highness(avgPct)
		}},
		// Linked: moderate coupling — not isolated but not a hub
		{"Linked", func() float64 {
			if avgPct >= 15 && avgPct < 60 {
				return 0.50 + (avgPct-15)/90
			}
			return 0
		}},
		// Independent: low coupling — clean boundary
		{"Independent", func() float64 {
			return highness(ms.BoundaryIntegrity)
		}},
	}
	best := pickBest(rules, 0.10)
	return best.Name, best.Confidence
}

// classifyVitality classifies a module's life force.
// 5 levels: Stable (healthy low-change), Warming (increasing pressure),
// Turbulent (high pressure + low survival), Critical (extreme), Dead.
// Requires blame data for survival-based classifications.
func classifyVitality(ms ModuleScore) (string, float64) {
	hasBlameData := ms.BlameLines > 0
	pressureLevel := 100 - ms.Stability // invert stability → pressure level (0-100)

	rules := []classifyRule{
		// Dead: hard gate — no commits AND no active owner
		{"Dead", func() float64 {
			if ms.ModuleCommits == 0 && !ms.OwnerActive {
				return 0.95
			}
			return 0
		}},
		// Critical: very high pressure + very low survival (requires blame data)
		{"Critical", func() float64 {
			if !hasBlameData {
				return 0
			}
			// Both pressure and survival must be extreme
			if pressureLevel >= 80 && ms.ChangeAbsorption < 20 {
				conf := 0.85 + (pressureLevel-80)/400 + (20-ms.ChangeAbsorption)/400
				if conf > 1.0 {
					conf = 1.0
				}
				return conf
			}
			return 0
		}},
		// Turbulent: high pressure + low survival (requires blame data)
		{"Turbulent", func() float64 {
			if !hasBlameData {
				return 0
			}
			return minf(highness(pressureLevel), lowness(ms.ChangeAbsorption))
		}},
		// Warming: moderate pressure + moderate-to-low survival (requires blame data)
		{"Warming", func() float64 {
			if !hasBlameData {
				return 0
			}
			// Pressure is rising but not yet extreme
			if pressureLevel >= 30 && pressureLevel < 70 && ms.ChangeAbsorption < 50 {
				return 0.40 + (pressureLevel-30)/100
			}
			return 0
		}},
		// Stable: low pressure (healthy equilibrium)
		{"Stable", func() float64 {
			return highness(ms.Stability)
		}},
	}
	best := pickBest(rules, 0.10)
	return best.Name, best.Confidence
}

// classifyOwnershipAxis classifies a module's knowledge distribution.
func classifyOwnershipAxis(ms ModuleScore, own *metric.ModuleOwnership, authorLastDate map[string]time.Time, activeDays int) (string, float64) {
	rules := []classifyRule{
		// Orphaned: owner not active
		{"Orphaned", func() float64 {
			if own == nil {
				return 0
			}
			if lastDate, ok := authorLastDate[own.TopAuthor]; ok {
				daysSince := time.Since(lastDate).Hours() / 24
				if daysSince > float64(activeDays) && own.TopShare >= 0.50 {
					return 0.70 + own.TopShare*0.25
				}
			}
			return 0
		}},
		// Concentrated: knowledge in few hands
		{"Concentrated", func() float64 {
			if own == nil {
				return 0
			}
			switch own.Level {
			case "SOLE_OWNER":
				return 0.95
			case "CONCENTRATED":
				return 0.70 + own.TopShare*0.25
			}
			return 0
		}},
		// Distributed: healthy spread
		{"Distributed", func() float64 {
			if own == nil {
				return 0
			}
			switch own.Level {
			case "HEALTHY":
				return 0.80
			case "FRAGMENTED":
				return 0.60
			}
			return 0
		}},
	}
	best := pickBest(rules, 0.10)
	return best.Name, best.Confidence
}
