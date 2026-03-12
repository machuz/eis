package team

import (
	"github.com/machuz/engineering-impact-score/internal/scorer"
)

// MinContributionThreshold is the minimum Total score for a member to be
// counted as a "core" team member. Members below this threshold who are not
// in a risk state (Former/Silent/Fragile) are treated as peripheral
// (cross-functional helpers) and excluded from team metrics.
const MinContributionThreshold = 20.0

// TeamResult holds aggregated team-level metrics.
//
// Members are categorized into three tiers:
//   - Core: RecentlyActive && Total >= MinContributionThreshold → averages denominator
//   - Risk: State in {Former, Silent, Fragile} → distributions & risk metrics
//   - Peripheral: everyone else → TotalMemberCount only
//
// MemberCount = core + risk ("effective"). Averages use CoreMemberCount.
type TeamResult struct {
	Name             string
	Domain           string
	MemberCount      int // effective members (core + risk)
	CoreMemberCount  int // core active members (for averages)
	TotalMemberCount int // all members including peripheral
	RepoCount        int
	Members          []scorer.Result // effective members (core + risk) — for classification
	CoreMembers      []scorer.Result // core members only — for averages/quality

	// Member axis averages (0-100), computed from CoreMembers only
	AvgProduction       float64
	AvgQuality          float64
	AvgSurvival         float64
	AvgRobustSurvival   float64
	AvgDormantSurvival  float64
	AvgDesign           float64
	AvgBreadth          float64
	AvgDebtCleanup      float64
	AvgIndispensability float64
	AvgTotal            float64

	// Role/Style/State distribution counts (from effective members)
	RoleDist  map[string]int
	StyleDist map[string]int
	StateDist map[string]int

	// Team health scores (0-100)
	Health TeamHealth

	// Team 5-axis classification (Structure/Culture/Phase/Risk + composite Character)
	Classification TeamClassification

	// Warnings: dangerous patterns detected from combining metrics
	Warnings []string
}

// TeamHealth holds the health axis scores.
type TeamHealth struct {
	Complementarity     float64
	GrowthPotential     float64
	Sustainability      float64
	DebtBalance         float64
	ProductivityDensity float64
	QualityConsistency  float64
	RiskRatio           float64 // percentage of risk members (0-100)

	// Structure metrics
	AAR                  float64 // Architect-to-Anchor Ratio (raw ratio, not 0-100)
	AnchorDensity        float64 // Anchors / Active members (0-1)
	ArchitectureCoverage float64 // (Architects + Anchors) / Team size (0-1)
}

// Aggregate computes team-level metrics from individual results.
//
// Members are split into three tiers:
//   - Core: RecentlyActive && Total >= MinContributionThreshold
//   - Risk: State in {Former, Silent, Fragile} (always included for detection)
//   - Peripheral: everyone else (cross-functional helpers, excluded from metrics)
//
// Averages are computed from core members only. Distributions and classification
// use effective members (core + risk). This prevents drive-by contributors from
// diluting metrics while keeping risk states visible.
func Aggregate(name, domain string, repoCount int, results []scorer.Result, memberFilter []string) TeamResult {
	allMembers := filterMembers(results, memberFilter)
	if len(allMembers) == 0 {
		return TeamResult{Name: name, Domain: domain, RepoCount: repoCount}
	}

	// Categorize members into core / risk / peripheral
	var coreMembers []scorer.Result
	var riskMembers []scorer.Result
	for _, m := range allMembers {
		if isRiskState(m.State) {
			riskMembers = append(riskMembers, m)
		} else if m.RecentlyActive && m.Total >= MinContributionThreshold {
			coreMembers = append(coreMembers, m)
		}
		// else: peripheral — only counted in TotalMemberCount
	}

	// Effective members = core + risk
	effectiveMembers := make([]scorer.Result, 0, len(coreMembers)+len(riskMembers))
	effectiveMembers = append(effectiveMembers, coreMembers...)
	effectiveMembers = append(effectiveMembers, riskMembers...)

	if len(effectiveMembers) == 0 {
		return TeamResult{
			Name:             name,
			Domain:           domain,
			TotalMemberCount: len(allMembers),
			RepoCount:        repoCount,
		}
	}

	tr := TeamResult{
		Name:             name,
		Domain:           domain,
		MemberCount:      len(effectiveMembers),
		CoreMemberCount:  len(coreMembers),
		TotalMemberCount: len(allMembers),
		RepoCount:        repoCount,
		Members:          effectiveMembers,
		CoreMembers:      coreMembers,
		RoleDist:         make(map[string]int),
		StyleDist:        make(map[string]int),
		StateDist:        make(map[string]int),
	}

	// Averages from core members only
	if len(coreMembers) > 0 {
		var sumProd, sumQual, sumSurv, sumRobust, sumDormant float64
		var sumDesign, sumBreadth, sumDebt, sumIndisp, sumTotal float64
		for _, m := range coreMembers {
			sumProd += m.Production
			sumQual += m.Quality
			sumSurv += m.Survival
			sumRobust += m.RobustSurvival
			sumDormant += m.DormantSurvival
			sumDesign += m.Design
			sumBreadth += m.Breadth
			sumDebt += m.DebtCleanup
			sumIndisp += m.Indispensability
			sumTotal += m.Total
		}
		n := float64(len(coreMembers))
		tr.AvgProduction = sumProd / n
		tr.AvgQuality = sumQual / n
		tr.AvgSurvival = sumSurv / n
		tr.AvgRobustSurvival = sumRobust / n
		tr.AvgDormantSurvival = sumDormant / n
		tr.AvgDesign = sumDesign / n
		tr.AvgBreadth = sumBreadth / n
		tr.AvgDebtCleanup = sumDebt / n
		tr.AvgIndispensability = sumIndisp / n
		tr.AvgTotal = sumTotal / n
	}

	// Distributions from effective members (core + risk)
	for _, m := range effectiveMembers {
		tr.RoleDist[m.Role]++
		tr.StyleDist[m.Style]++
		tr.StateDist[m.State]++
	}

	tr.Health = CalcHealth(tr)
	tr.Classification = Classify(tr)
	tr.Warnings = detectWarnings(tr)

	return tr
}

// isRiskState returns true if the state indicates a risk condition
// that should always be visible in team metrics.
func isRiskState(state string) bool {
	return state == "Former" || state == "Silent" || state == "Fragile"
}

func filterMembers(results []scorer.Result, filter []string) []scorer.Result {
	if len(filter) == 0 {
		return results
	}
	allowed := make(map[string]bool, len(filter))
	for _, name := range filter {
		allowed[name] = true
	}
	var out []scorer.Result
	for _, r := range results {
		if allowed[r.Author] {
			out = append(out, r)
		}
	}
	return out
}
