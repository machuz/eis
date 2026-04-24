package timeline

import (
	"sort"

	"github.com/machuz/eis/v2/internal/scorer"
	"github.com/machuz/eis/v2/internal/team"
)

// PeriodResult holds scored results for a single time period within a domain.
type PeriodResult struct {
	Label   string
	Start   string // ISO date
	End     string // ISO date
	Members []scorer.Result
}

// AuthorTimeline tracks one author's scores across periods.
type AuthorTimeline struct {
	Author      string
	Periods     []AuthorPeriod
	Transitions []Transition
}

// AuthorPeriod holds an author's scores for one period.
type AuthorPeriod struct {
	Label            string
	Impact           float64
	Production       float64
	Quality          float64
	Survival         float64
	RobustSurvival   float64
	DormantSurvival  float64
	Design           float64
	Breadth          float64
	DebtCleanup      float64
	Indispensability float64
	Gravity          float64
	Role             string
	RoleConf         float64
	Style            string
	StyleConf        float64
	State            string
	StateConf        float64
	TotalCommits     int
	LinesAdded       int
	LinesDeleted     int
	Active           bool
}

// Transition represents a change in Role, Style, or State between periods.
type Transition struct {
	Axis     string // "Role", "Style", "State"
	From     string
	To       string
	AtPeriod string
}

// BuildTimeline takes per-period results and produces per-author timelines.
func BuildTimeline(periods []PeriodResult) []AuthorTimeline {
	// Collect all authors across all periods
	authorSet := make(map[string]bool)
	for _, p := range periods {
		for _, m := range p.Members {
			authorSet[m.Author] = true
		}
	}

	var timelines []AuthorTimeline
	for author := range authorSet {
		tl := AuthorTimeline{Author: author}

		for _, p := range periods {
			ap := AuthorPeriod{Label: p.Label}
			found := false
			for _, m := range p.Members {
				if m.Author == author {
					ap = AuthorPeriod{
						Label:            p.Label,
						Impact:           m.Impact,
						Production:       m.Production,
						Quality:          m.Quality,
						Survival:         m.Survival,
						RobustSurvival:   m.RobustSurvival,
						DormantSurvival:  m.DormantSurvival,
						Design:           m.Design,
						Breadth:          m.Breadth,
						DebtCleanup:      m.DebtCleanup,
						Indispensability: m.Indispensability,
						Gravity:          m.Gravity,
						Role:             m.Role,
						RoleConf:         m.RoleConf,
						Style:            m.Style,
						StyleConf:        m.StyleConf,
						State:            m.State,
						StateConf:        m.StateConf,
						TotalCommits:     m.TotalCommits,
						LinesAdded:       m.LinesAdded,
						LinesDeleted:     m.LinesDeleted,
						Active:           m.RecentlyActive,
					}
					found = true
					break
				}
			}
			if !found {
				ap.Label = p.Label
			}
			tl.Periods = append(tl.Periods, ap)
		}

		tl.Transitions = DetectTransitions(tl.Periods)
		timelines = append(timelines, tl)
	}

	// Sort by latest period's impact descending
	sort.Slice(timelines, func(i, j int) bool {
		iTotal := latestNonZeroImpact(timelines[i].Periods)
		jTotal := latestNonZeroImpact(timelines[j].Periods)
		return iTotal > jTotal
	})

	return timelines
}

func latestNonZeroImpact(periods []AuthorPeriod) float64 {
	for i := len(periods) - 1; i >= 0; i-- {
		if periods[i].Impact > 0 {
			return periods[i].Impact
		}
	}
	return 0
}

// DetectTransitions finds changes in Role, Style, and State across periods.
func DetectTransitions(periods []AuthorPeriod) []Transition {
	var transitions []Transition
	if len(periods) < 2 {
		return transitions
	}

	for i := 1; i < len(periods); i++ {
		prev := periods[i-1]
		curr := periods[i]

		if prev.Role != "" && curr.Role != "" && prev.Role != "—" && curr.Role != "—" && prev.Role != curr.Role {
			transitions = append(transitions, Transition{
				Axis:     "Role",
				From:     prev.Role,
				To:       curr.Role,
				AtPeriod: curr.Label,
			})
		}
		if prev.Style != "" && curr.Style != "" && prev.Style != "—" && curr.Style != "—" && prev.Style != curr.Style {
			transitions = append(transitions, Transition{
				Axis:     "Style",
				From:     prev.Style,
				To:       curr.Style,
				AtPeriod: curr.Label,
			})
		}
		if prev.State != "" && curr.State != "" && prev.State != "—" && curr.State != "—" && prev.State != curr.State {
			transitions = append(transitions, Transition{
				Axis:     "State",
				From:     prev.State,
				To:       curr.State,
				AtPeriod: curr.Label,
			})
		}
	}

	return transitions
}

// TeamPeriodResult holds a team's aggregated result for one period.
type TeamPeriodResult struct {
	Label      string
	Start      string
	End        string
	TeamResult team.TeamResult
}

// TeamTimeline tracks a team's metrics across periods.
type TeamTimeline struct {
	TeamName    string
	Domain      string
	Periods     []TeamPeriodSnapshot
	Transitions []Transition
}

// TeamPeriodSnapshot holds the key metrics for one period.
type TeamPeriodSnapshot struct {
	Label           string
	CoreMembers     int
	EffectiveMembers int
	TotalMembers    int

	// Averages
	AvgImpact      float64
	AvgProduction  float64
	AvgQuality     float64
	AvgSurvival    float64
	AvgDesign      float64
	AvgDebtCleanup float64

	// Health
	Complementarity     float64
	GrowthPotential     float64
	Sustainability      float64
	DebtBalance         float64
	ProductivityDensity float64
	QualityConsistency  float64
	RiskRatio           float64

	// Classification
	Character string
	Structure string
	Culture   string
	Phase     string
	Risk      string

	// Distribution (top entries)
	RoleDist  map[string]int
	StyleDist map[string]int
	StateDist map[string]int
}

// BuildTeamTimeline builds a team timeline from per-period team results.
func BuildTeamTimeline(teamName, domain string, periods []TeamPeriodResult) TeamTimeline {
	tl := TeamTimeline{
		TeamName: teamName,
		Domain:   domain,
	}

	for _, p := range periods {
		tr := p.TeamResult
		snap := TeamPeriodSnapshot{
			Label:            p.Label,
			CoreMembers:      tr.CoreMemberCount,
			EffectiveMembers: tr.MemberCount,
			TotalMembers:     tr.TotalMemberCount,

			AvgImpact:      tr.AvgImpact,
			AvgProduction:  tr.AvgProduction,
			AvgQuality:     tr.AvgQuality,
			AvgSurvival:    tr.AvgSurvival,
			AvgDesign:      tr.AvgDesign,
			AvgDebtCleanup: tr.AvgDebtCleanup,

			Complementarity:     tr.Health.Complementarity,
			GrowthPotential:     tr.Health.GrowthPotential,
			Sustainability:      tr.Health.Sustainability,
			DebtBalance:         tr.Health.DebtBalance,
			ProductivityDensity: tr.Health.ProductivityDensity,
			QualityConsistency:  tr.Health.QualityConsistency,
			RiskRatio:           tr.Health.RiskRatio,

			Character: tr.Classification.Character.Name,
			Structure: tr.Classification.Structure.Name,
			Culture:   tr.Classification.Culture.Name,
			Phase:     tr.Classification.Phase.Name,
			Risk:      tr.Classification.Risk.Name,

			RoleDist:  tr.RoleDist,
			StyleDist: tr.StyleDist,
			StateDist: tr.StateDist,
		}
		tl.Periods = append(tl.Periods, snap)
	}

	tl.Transitions = detectTeamTransitions(tl.Periods)
	return tl
}

func detectTeamTransitions(periods []TeamPeriodSnapshot) []Transition {
	var transitions []Transition
	if len(periods) < 2 {
		return transitions
	}

	axes := []struct {
		name string
		get  func(TeamPeriodSnapshot) string
	}{
		{"Character", func(p TeamPeriodSnapshot) string { return p.Character }},
		{"Structure", func(p TeamPeriodSnapshot) string { return p.Structure }},
		{"Culture", func(p TeamPeriodSnapshot) string { return p.Culture }},
		{"Phase", func(p TeamPeriodSnapshot) string { return p.Phase }},
		{"Risk", func(p TeamPeriodSnapshot) string { return p.Risk }},
	}

	for i := 1; i < len(periods); i++ {
		for _, ax := range axes {
			prev := ax.get(periods[i-1])
			curr := ax.get(periods[i])
			if prev != "" && curr != "" && prev != "—" && curr != "—" && prev != curr {
				transitions = append(transitions, Transition{
					Axis:     ax.name,
					From:     prev,
					To:       curr,
					AtPeriod: periods[i].Label,
				})
			}
		}
	}

	return transitions
}
