package team

import (
	"fmt"
	"sort"

	"github.com/machuz/engineering-impact-score/internal/scorer"
)

// detectWarnings generates human-readable warnings from dangerous metric combinations.
func detectWarnings(tr TeamResult) []string {
	var warnings []string

	if tr.CoreMemberCount == 0 {
		warnings = append(warnings, "No core contributors — team has no active members above contribution threshold")
		return warnings
	}

	// High bus factor: few core members carrying many repos
	if tr.CoreMemberCount <= 3 && tr.RepoCount >= 5 {
		warnings = append(warnings, fmt.Sprintf(
			"%d core members carrying %d repos — high bus factor risk",
			tr.CoreMemberCount, tr.RepoCount))
	}

	// Risk ratio warning
	if tr.Health.RiskRatio >= 40 {
		riskCount := tr.MemberCount - tr.CoreMemberCount
		warnings = append(warnings, fmt.Sprintf(
			"%.0f%% risk ratio — %d of %d effective members are Former/Silent/Fragile",
			tr.Health.RiskRatio, riskCount, tr.MemberCount))
	} else if tr.Health.RiskRatio >= 25 {
		warnings = append(warnings, fmt.Sprintf(
			"%.0f%% risk ratio — approaching danger zone",
			tr.Health.RiskRatio))
	}

	// Top contributor concentration
	if tr.CoreMemberCount >= 2 {
		topProd, topName := findTopContributor(tr.CoreMembers)
		totalProd := tr.AvgProduction * float64(tr.CoreMemberCount)
		if totalProd > 0 {
			share := topProd / totalProd * 100
			if share >= 45 {
				// Simulate removal
				remaining := (totalProd - topProd) / float64(tr.CoreMemberCount-1)
				warnings = append(warnings, fmt.Sprintf(
					"Top contributor (%s) accounts for %.0f%% of core production — ProdDensity drops to %.0f without them",
					topName, share, remaining))
			}
		}
	}

	// Silent accumulation
	silentCount := tr.StateDist["Silent"]
	if silentCount >= 2 {
		warnings = append(warnings, fmt.Sprintf(
			"%d Silent members — headcount says %d but effective contributors are %d",
			silentCount, tr.TotalMemberCount, tr.CoreMemberCount))
	}

	// No architect with multiple repos
	if tr.RoleDist["Architect"] == 0 && tr.RepoCount >= 3 {
		warnings = append(warnings, fmt.Sprintf(
			"No Architect across %d repos — design decisions are implicit and unguarded",
			tr.RepoCount))
	}

	// Gravity-based warnings
	warnings = append(warnings, detectGravityWarnings(tr)...)

	return warnings
}

// detectGravityWarnings generates warnings based on structural influence (Gravity).
func detectGravityWarnings(tr TeamResult) []string {
	var warnings []string
	if len(tr.Members) == 0 {
		return nil
	}

	// Find top gravity contributor
	sorted := make([]scorer.Result, len(tr.Members))
	copy(sorted, tr.Members)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Gravity > sorted[j].Gravity
	})

	top := sorted[0]

	// Single gravity center: top gravity > 80, significant gap, AND low robust survival
	// If robust survival is high, the concentration is durable — not a warning
	if top.Gravity >= 80 && top.RobustSurvival < 50 {
		gap := top.Gravity
		if len(sorted) > 1 {
			gap = top.Gravity - sorted[1].Gravity
		}
		if gap >= 30 {
			warnings = append(warnings, fmt.Sprintf(
				"Fragile gravity center — %s (Grav %.0f, RobustSurv %.0f) concentrates structural influence without durability",
				top.Author, top.Gravity, top.RobustSurvival))
		}
	}

	// Fragile gravity: high gravity but zero/low robust survival
	for _, m := range tr.Members {
		if m.Gravity >= 50 && m.RobustSurvival < 20 {
			warnings = append(warnings, fmt.Sprintf(
				"Fragile gravity — %s (Grav %.0f) has high influence but low robust survival (%.0f)",
				m.Author, m.Gravity, m.RobustSurvival))
		}
	}

	// Architect gravity exists but structural coverage remains low
	if tr.RoleDist["Architect"] > 0 && tr.Health.ArchitectureCoverage < 0.25 {
		warnings = append(warnings, fmt.Sprintf(
			"Architect gravity exists but structural coverage remains low (%.0f%%)",
			tr.Health.ArchitectureCoverage*100))
	}

	return warnings
}

func findTopContributor(members []scorer.Result) (float64, string) {
	if len(members) == 0 {
		return 0, ""
	}
	sorted := make([]scorer.Result, len(members))
	copy(sorted, members)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Production > sorted[j].Production
	})
	return sorted[0].Production, sorted[0].Author
}
