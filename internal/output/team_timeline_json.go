package output

import (
	"encoding/json"
	"os"

	"github.com/machuz/eis/v2/internal/timeline"
)

type teamTimelineJSONOutput struct {
	TeamTimelines []teamTimelineJSON `json:"team_timelines"`
}

type teamTimelineJSON struct {
	TeamName    string                        `json:"team_name"`
	Domain      string                        `json:"domain"`
	Periods     []teamPeriodSnapshotJSON      `json:"periods"`
	Transitions []teamTimelineTransitionJSON      `json:"transitions,omitempty"`
}

type teamPeriodSnapshotJSON struct {
	Label            string         `json:"label"`
	CoreMembers      int            `json:"core_members"`
	EffectiveMembers int            `json:"effective_members"`
	TotalMembers     int            `json:"total_members"`
	Averages         teamAverages   `json:"averages"`
	Health           teamHealthJSON `json:"health"`
	Classification   teamClassifyJSON `json:"classification"`
	RoleDist         map[string]int `json:"role_distribution"`
	StyleDist        map[string]int `json:"style_distribution"`
	StateDist        map[string]int `json:"state_distribution"`
}

type teamTimelineTransitionJSON struct {
	Axis     string `json:"axis"`
	From     string `json:"from"`
	To       string `json:"to"`
	AtPeriod string `json:"at_period"`
}

// PrintTeamTimelineJSON outputs team timeline data as JSON.
func PrintTeamTimelineJSON(timelines []timeline.TeamTimeline) error {
	out := teamTimelineJSONOutput{}

	for _, tl := range timelines {
		tj := teamTimelineJSON{
			TeamName: tl.TeamName,
			Domain:   tl.Domain,
		}

		for _, p := range tl.Periods {
			snap := teamPeriodSnapshotJSON{
				Label:            p.Label,
				CoreMembers:      p.CoreMembers,
				EffectiveMembers: p.EffectiveMembers,
				TotalMembers:     p.TotalMembers,
				Averages: teamAverages{
					Production:  round1(p.AvgProduction),
					Quality:     round1(p.AvgQuality),
					Survival:    round1(p.AvgSurvival),
					Design:      round1(p.AvgDesign),
					DebtCleanup: round1(p.AvgDebtCleanup),
					Impact:      round1(p.AvgImpact),
				},
				Health: teamHealthJSON{
					Complementarity:     round1(p.Complementarity),
					GrowthPotential:     round1(p.GrowthPotential),
					Sustainability:      round1(p.Sustainability),
					DebtBalance:         round1(p.DebtBalance),
					ProductivityDensity: round1(p.ProductivityDensity),
					QualityConsistency:  round1(p.QualityConsistency),
					RiskRatio:           round1(p.RiskRatio),
				},
				Classification: teamClassifyJSON{
					Character: p.Character,
					Structure: p.Structure,
					Culture:   p.Culture,
					Phase:     p.Phase,
					Risk:      p.Risk,
				},
				RoleDist:  p.RoleDist,
				StyleDist: p.StyleDist,
				StateDist: p.StateDist,
			}
			tj.Periods = append(tj.Periods, snap)
		}

		for _, t := range tl.Transitions {
			tj.Transitions = append(tj.Transitions, teamTimelineTransitionJSON{
				Axis:     t.Axis,
				From:     t.From,
				To:       t.To,
				AtPeriod: t.AtPeriod,
			})
		}

		out.TeamTimelines = append(out.TeamTimelines, tj)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
