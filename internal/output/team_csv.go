package output

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/machuz/eis/v2/internal/team"
)

// PrintTeamCSV outputs team results as CSV (one row per team).
func PrintTeamCSV(teams []team.TeamResult) {
	w := csv.NewWriter(os.Stdout)

	w.Write([]string{
		"team", "domain", "core_members", "effective_members", "total_members", "repos",
		"character", "structure", "culture", "phase", "risk",
		"avg_production", "avg_quality", "avg_survival", "avg_robust_survival", "avg_dormant_survival",
		"avg_design", "avg_breadth", "avg_debt_cleanup", "avg_indispensability", "avg_impact",
		"complementarity", "growth_potential", "sustainability", "debt_balance",
		"productivity_density", "quality_consistency", "risk_ratio",
		"aar", "anchor_density", "architecture_coverage",
	})

	for _, tr := range teams {
		h := tr.Health
		c := tr.Classification
		w.Write([]string{
			tr.Name,
			tr.Domain,
			fmt.Sprintf("%d", tr.CoreMemberCount),
			fmt.Sprintf("%d", tr.MemberCount),
			fmt.Sprintf("%d", tr.TotalMemberCount),
			fmt.Sprintf("%d", tr.RepoCount),
			c.Character.Name,
			c.Structure.Name,
			c.Culture.Name,
			c.Phase.Name,
			c.Risk.Name,
			fmt.Sprintf("%.1f", tr.AvgProduction),
			fmt.Sprintf("%.1f", tr.AvgQuality),
			fmt.Sprintf("%.1f", tr.AvgSurvival),
			fmt.Sprintf("%.1f", tr.AvgRobustSurvival),
			fmt.Sprintf("%.1f", tr.AvgDormantSurvival),
			fmt.Sprintf("%.1f", tr.AvgDesign),
			fmt.Sprintf("%.1f", tr.AvgBreadth),
			fmt.Sprintf("%.1f", tr.AvgDebtCleanup),
			fmt.Sprintf("%.1f", tr.AvgIndispensability),
			fmt.Sprintf("%.1f", tr.AvgImpact),
			fmt.Sprintf("%.1f", h.Complementarity),
			fmt.Sprintf("%.1f", h.GrowthPotential),
			fmt.Sprintf("%.1f", h.Sustainability),
			fmt.Sprintf("%.1f", h.DebtBalance),
			fmt.Sprintf("%.1f", h.ProductivityDensity),
			fmt.Sprintf("%.1f", h.QualityConsistency),
			fmt.Sprintf("%.1f", h.RiskRatio),
			fmt.Sprintf("%.2f", h.AAR),
			fmt.Sprintf("%.2f", h.AnchorDensity),
			fmt.Sprintf("%.2f", h.ArchitectureCoverage),
		})
	}

	w.Flush()
}
