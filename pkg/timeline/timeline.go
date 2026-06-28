package timeline

import (
	"github.com/machuz/eis/v2/internal/cache"
	"github.com/machuz/eis/v2/internal/timeline"
)

// BlameCacheEpoch mirrors internal/cache.Epoch — the version of the on-disk
// blame/log cache CONTENT (the gob shapes and the parsing that fills them).
// A consumer that PERSISTS eis's cache across versions (e.g. ace's S3 blame
// cache) MUST fold this into its own cache namespace, so an entry written
// under one epoch is never decoded under another. It changes ONLY when the
// cached output changes — not on every release — so the persistent cache
// stays warm across patch/minor bumps. See internal/cache.Epoch for the
// bump policy.
const BlameCacheEpoch = cache.Epoch

type PeriodResult = timeline.PeriodResult
type RepoPeriodResult = timeline.RepoPeriodResult
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
