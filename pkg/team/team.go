package team

import "github.com/machuz/eis/v2/internal/team"

type TeamResult = team.TeamResult
type TeamHealth = team.TeamHealth
type TeamClassification = team.TeamClassification

var (
	Aggregate  = team.Aggregate
	Classify   = team.Classify
	CalcHealth = team.CalcHealth
)

const MinContributionThreshold = team.MinContributionThreshold
