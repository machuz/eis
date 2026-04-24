package timeline

import "github.com/machuz/eis/v2/internal/timeline"

type PeriodResult = timeline.PeriodResult
type AuthorTimeline = timeline.AuthorTimeline
type AuthorPeriod = timeline.AuthorPeriod
type Transition = timeline.Transition
type TeamTimeline = timeline.TeamTimeline
type TeamPeriodResult = timeline.TeamPeriodResult
type TeamPeriodSnapshot = timeline.TeamPeriodSnapshot

var (
	BuildTimeline     = timeline.BuildTimeline
	DetectTransitions = timeline.DetectTransitions
	BuildTeamTimeline = timeline.BuildTeamTimeline
)
