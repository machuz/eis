package output

import (
	"encoding/json"
	"os"

	"github.com/machuz/eis/v2/internal/team"
)

type teamJSONOutput struct {
	Teams []teamJSON `json:"teams"`
}

type teamJSON struct {
	Name             string           `json:"name"`
	Domain           string           `json:"domain"`
	CoreMemberCount  int              `json:"core_member_count"`
	EffectiveCount   int              `json:"effective_member_count"`
	TotalMemberCount int              `json:"total_member_count"`
	RepoCount        int              `json:"repo_count"`
	Classification   teamClassifyJSON `json:"classification"`
	Averages         teamAverages     `json:"averages"`
	Health           teamHealthJSON   `json:"health"`
	Structure        structureJSON    `json:"structure_metrics"`
	Warnings         []string         `json:"warnings,omitempty"`
	Roles            map[string]int   `json:"role_distribution"`
	Styles           map[string]int   `json:"style_distribution"`
	States           map[string]int   `json:"state_distribution"`
	Members          []jsonMember     `json:"members"`
}

type teamClassifyJSON struct {
	Character     string  `json:"character"`
	CharConf      float64 `json:"character_confidence"`
	Structure     string  `json:"structure"`
	StructureConf float64 `json:"structure_confidence"`
	Culture       string  `json:"culture"`
	CultureConf   float64 `json:"culture_confidence"`
	Phase         string  `json:"phase"`
	PhaseConf     float64 `json:"phase_confidence"`
	Risk          string  `json:"risk"`
	RiskConf      float64 `json:"risk_confidence"`
}

type teamAverages struct {
	Production       float64 `json:"production"`
	Quality          float64 `json:"quality"`
	Survival         float64 `json:"survival"`
	RobustSurvival   float64 `json:"robust_survival"`
	DormantSurvival  float64 `json:"dormant_survival"`
	Design           float64 `json:"design"`
	Breadth          float64 `json:"breadth"`
	DebtCleanup      float64 `json:"debt_cleanup"`
	Indispensability float64 `json:"indispensability"`
	Impact           float64 `json:"impact"`
}

type teamHealthJSON struct {
	Complementarity     float64 `json:"complementarity"`
	GrowthPotential     float64 `json:"growth_potential"`
	Sustainability      float64 `json:"sustainability"`
	DebtBalance         float64 `json:"debt_balance"`
	ProductivityDensity float64 `json:"productivity_density"`
	QualityConsistency  float64 `json:"quality_consistency"`
	RiskRatio           float64 `json:"risk_ratio"`
}

type structureJSON struct {
	AAR                  float64 `json:"aar"`
	AnchorDensity        float64 `json:"anchor_density"`
	ArchitectureCoverage float64 `json:"architecture_coverage"`
}

// PrintTeamJSON outputs team results as JSON.
func PrintTeamJSON(teams []team.TeamResult) error {
	out := teamJSONOutput{}

	for _, tr := range teams {
		tj := teamJSON{
			Name:             tr.Name,
			Domain:           tr.Domain,
			CoreMemberCount:  tr.CoreMemberCount,
			EffectiveCount:   tr.MemberCount,
			TotalMemberCount: tr.TotalMemberCount,
			RepoCount:        tr.RepoCount,
			Classification: teamClassifyJSON{
				Character:     tr.Classification.Character.Name,
				CharConf:      tr.Classification.Character.Confidence,
				Structure:     tr.Classification.Structure.Name,
				StructureConf: tr.Classification.Structure.Confidence,
				Culture:       tr.Classification.Culture.Name,
				CultureConf:   tr.Classification.Culture.Confidence,
				Phase:         tr.Classification.Phase.Name,
				PhaseConf:     tr.Classification.Phase.Confidence,
				Risk:          tr.Classification.Risk.Name,
				RiskConf:      tr.Classification.Risk.Confidence,
			},
			Averages: teamAverages{
				Production:       round1(tr.AvgProduction),
				Quality:          round1(tr.AvgQuality),
				Survival:         round1(tr.AvgSurvival),
				RobustSurvival:   round1(tr.AvgRobustSurvival),
				DormantSurvival:  round1(tr.AvgDormantSurvival),
				Design:           round1(tr.AvgDesign),
				Breadth:          round1(tr.AvgBreadth),
				DebtCleanup:      round1(tr.AvgDebtCleanup),
				Indispensability: round1(tr.AvgIndispensability),
				Impact:           round1(tr.AvgImpact),
			},
			Health: teamHealthJSON{
				Complementarity:     round1(tr.Health.Complementarity),
				GrowthPotential:     round1(tr.Health.GrowthPotential),
				Sustainability:      round1(tr.Health.Sustainability),
				DebtBalance:         round1(tr.Health.DebtBalance),
				ProductivityDensity: round1(tr.Health.ProductivityDensity),
				QualityConsistency:  round1(tr.Health.QualityConsistency),
				RiskRatio:           round1(tr.Health.RiskRatio),
			},
			Structure: structureJSON{
				AAR:                  round2(tr.Health.AAR),
				AnchorDensity:        round2(tr.Health.AnchorDensity),
				ArchitectureCoverage: round2(tr.Health.ArchitectureCoverage),
			},
			Warnings: tr.Warnings,
			Roles:    tr.RoleDist,
			Styles:   tr.StyleDist,
			States:   tr.StateDist,
		}

		for i, m := range tr.Members {
			tj.Members = append(tj.Members, jsonMember{
				Rank:             i + 1,
				Member:           m.Author,
				Active:           m.RecentlyActive,
				Commits:          m.TotalCommits,
				Production:       round1(m.Production),
				Quality:          round1(m.Quality),
				Survival:         round1(m.Survival),
				RobustSurvival:   round1(m.RobustSurvival),
				DormantSurvival:  round1(m.DormantSurvival),
				Design:           round1(m.Design),
				Breadth:          round1(m.Breadth),
				DebtCleanup:      round1(m.DebtCleanup),
				Indispensability: round1(m.Indispensability),
				Gravity:          round1(m.Gravity),
				Impact:           round1(m.Impact),
				Role:             m.Role,
				RoleConf:         m.RoleConf,
				Style:            m.Style,
				StyleConf:        m.StyleConf,
				State:            m.State,
				StateConf:        m.StateConf,
			})
		}

		out.Teams = append(out.Teams, tj)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
