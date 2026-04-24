package output

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/machuz/eis/v2/internal/timeline"
)

// PrintTeamTimelineCSV outputs team timeline data as CSV.
func PrintTeamTimelineCSV(timelines []timeline.TeamTimeline) {
	w := csv.NewWriter(os.Stdout)

	w.Write([]string{
		"team", "domain", "period",
		"core_members", "effective_members", "total_members",
		"character", "structure", "culture", "phase", "risk",
		"avg_production", "avg_quality", "avg_survival", "avg_design", "avg_debt_cleanup", "avg_impact",
		"complementarity", "growth_potential", "sustainability", "debt_balance",
		"productivity_density", "quality_consistency", "risk_ratio",
	})

	for _, tl := range timelines {
		for _, p := range tl.Periods {
			w.Write([]string{
				tl.TeamName,
				tl.Domain,
				p.Label,
				fmt.Sprintf("%d", p.CoreMembers),
				fmt.Sprintf("%d", p.EffectiveMembers),
				fmt.Sprintf("%d", p.TotalMembers),
				p.Character,
				p.Structure,
				p.Culture,
				p.Phase,
				p.Risk,
				fmt.Sprintf("%.1f", p.AvgProduction),
				fmt.Sprintf("%.1f", p.AvgQuality),
				fmt.Sprintf("%.1f", p.AvgSurvival),
				fmt.Sprintf("%.1f", p.AvgDesign),
				fmt.Sprintf("%.1f", p.AvgDebtCleanup),
				fmt.Sprintf("%.1f", p.AvgImpact),
				fmt.Sprintf("%.1f", p.Complementarity),
				fmt.Sprintf("%.1f", p.GrowthPotential),
				fmt.Sprintf("%.1f", p.Sustainability),
				fmt.Sprintf("%.1f", p.DebtBalance),
				fmt.Sprintf("%.1f", p.ProductivityDensity),
				fmt.Sprintf("%.1f", p.QualityConsistency),
				fmt.Sprintf("%.1f", p.RiskRatio),
			})
		}
	}

	w.Flush()
}
