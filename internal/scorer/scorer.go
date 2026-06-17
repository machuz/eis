package scorer

import (
	"sort"
	"time"

	"github.com/machuz/eis/v2/internal/config"
	"github.com/machuz/eis/v2/internal/metric"
)

type Result struct {
	Author           string
	Production       float64
	Catalysis        float64
	Survival         float64
	RawSurvival      float64 // normalized raw blame (no decay), used for archetype detection
	RobustSurvival   float64 // survival in high change-pressure modules
	DormantSurvival  float64 // survival in low change-pressure modules
	RawRobustSurv    float64 // raw (pre-normalize) robust survival, for archetype detection
	RawDormantSurv   float64 // raw (pre-normalize) dormant survival, for archetype detection
	TestedSurvival   float64 // survival from test-guarded files (normalized, pre-α)
	UntestedSurvival float64 // survival from untested files (normalized, pre-α)
	RawTestedSurv    float64 // raw (pre-normalize) tested survival, for JSON/downstream observability
	RawUntestedSurv  float64 // raw (pre-normalize) untested survival
	Design           float64
	Breadth          float64
	DebtCleanup      float64
	Indispensability float64
	Gravity          float64 // structural influence: f(Catalysis, RobustSurvival, Design, Breadth, Indispensability)
	Impact           float64
	TotalCommits     int
	LinesAdded       int
	LinesDeleted     int
	RecentlyActive   bool    // true if author has commits within active_days (default 30)
	Role             string  // Role axis: what they contribute (Architect, Anchor, Cleaner, Producer, Specialist, —)
	RoleConf         float64 // Role confidence (0.0-1.0)
	Style            string  // Style axis: how they contribute (Builder, Resilient, Rescue, Churn, Mass, Balanced, Spread, —)
	StyleConf        float64 // Style confidence (0.0-1.0)
	State            string  // State axis: lifecycle phase (Active, Growing, Former, Silent, Fragile, —)
	StateConf        float64 // State confidence (0.0-1.0)
}

// gravityScore composes Gravity — how much the system's STRUCTURE depends on
// this engineer. All inputs are normalised 0..100, the weights sum to 1, so
// Gravity stays on a 0..100 scale.
//
// Gravity is relational — it cannot be observed from one person alone. So the
// formula LEADS with the axes that require other people and resist solo
// inflation, and only lightly credits the axes a lone author can max out alone:
//
//   - Catalysis (0.30): others build on your surviving code. Zero in a solo
//     project, and the hardest axis to fake — the strongest evidence that the
//     system genuinely leans on you.
//   - RobustSurvival (0.25): code that lasts UNDER change pressure.
//     DormantSurvival is excluded — code resting in quiet modules is durable but
//     exerts no pull on what others build, so it must not manufacture gravity.
//   - Design (0.20): architectural shaping of the system.
//   - Breadth (0.15): how widely the work is embedded across modules.
//   - Indispensability (0.10): sole-ownership / bus-factor. Down from the old
//     0.40 — it is the most solo-inflatable axis (own a repo alone and it pins
//     to 100), so a one-person project can no longer mint high gravity from
//     ownership the way it used to.
func gravityScore(r Result) float64 {
	return r.Catalysis*0.30 +
		r.RobustSurvival*0.25 +
		r.Design*0.20 +
		r.Breadth*0.15 +
		r.Indispensability*0.10
}

// ScoreAt is like Score but uses refTime as the "now" reference for RecentlyActive calculation.
// This is essential for timeline analysis where past periods must not compare against real "now".
func ScoreAt(raw *metric.RawScores, cfg *config.Config, authorLastDate map[string]time.Time, refTime time.Time) []Result {
	return scoreImpl(raw, cfg, authorLastDate, refTime)
}

func Score(raw *metric.RawScores, cfg *config.Config, authorLastDate map[string]time.Time) []Result {
	return scoreImpl(raw, cfg, authorLastDate, time.Now())
}

func scoreImpl(raw *metric.RawScores, cfg *config.Config, authorLastDate map[string]time.Time, refTime time.Time) []Result {
	// Production: absolute scale — raw.Production is already per-day rate
	// Score = min(per_day / production_daily_ref * 100, 100)
	normProd := make(map[string]float64)
	for author, perDay := range raw.Production {
		score := perDay / cfg.ProductionDailyRef * 100
		if score > 100 {
			score = 100
		}
		normProd[author] = score
	}
	normSurv := Normalize(raw.Survival)
	normRobustSurv := Normalize(raw.RobustSurvival)
	normDormantSurv := Normalize(raw.DormantSurvival)
	normTestedSurv := Normalize(raw.TestedSurvival)
	normUntestedSurv := Normalize(raw.UntestedSurvival)
	normDesign := Normalize(raw.Design)
	normIndisp := Normalize(raw.Indispensability)
	normRawSurv := Normalize(raw.RawSurvival)

	// Catalysis is a relative surviving-mass score (others' work on files you
	// originated), normalized within the group like Survival.
	normCatalysis := Normalize(raw.Catalysis)

	// Breadth: relative scale — normalized within the group
	normBreadth := Normalize(raw.Breadth)

	// Debt is already on 0-100 scale, use directly
	// Authors not in the debt map get 50 (neutral / insufficient data)
	normDebt := make(map[string]float64)
	for _, a := range raw.Authors() {
		if v, ok := raw.DebtCleanup[a]; ok {
			normDebt[a] = v
		} else {
			normDebt[a] = 50
		}
	}

	// Collect all authors
	authors := raw.Authors()
	w := cfg.Weights

	var results []Result
	for _, author := range authors {
		// Observation, not evaluation: a domain's team consists of people who
		// authored code (non-merge commits) to that domain's repos. Merge-only
		// authors (e.g. PR-mergers, automation accounts) have TotalCommits == 0
		// because TotalCommits is incremented only for non-merge commits.
		// Emitting them in the team rollup conflates "merged PRs" with "wrote
		// code". Reviewer / merger contribution is a dark-matter concern (D-06
		// prose), not a member-list concern.
		if raw.TotalCommits[author] == 0 {
			continue
		}

		// Determine if author has been active in last 6 months
		recentlyActive := false
		if lastDate, ok := authorLastDate[author]; ok {
			recentlyActive = refTime.Sub(lastDate).Hours()/24 <= float64(cfg.ActiveDays)
		}

		r := Result{
			Author:           author,
			Production:       normProd[author],
			Catalysis:        normCatalysis[author],
			Survival:         normSurv[author],
			RobustSurvival:   normRobustSurv[author],
			DormantSurvival:  normDormantSurv[author],
			RawRobustSurv:    raw.RobustSurvival[author],
			RawDormantSurv:   raw.DormantSurvival[author],
			TestedSurvival:   normTestedSurv[author],
			UntestedSurvival: normUntestedSurv[author],
			RawTestedSurv:    raw.TestedSurvival[author],
			RawUntestedSurv:  raw.UntestedSurvival[author],
			Design:           normDesign[author],
			Breadth:          normBreadth[author],
			DebtCleanup:      normDebt[author],
			Indispensability: normIndisp[author],
			RawSurvival:      normRawSurv[author],
			TotalCommits:     raw.TotalCommits[author],
			LinesAdded:       raw.LinesAdded[author],
			LinesDeleted:     raw.LinesDeleted[author],
			RecentlyActive:   recentlyActive,
		}

		// When robust/dormant data is available, split survival weight 80/20.
		// Otherwise fall back to classic single survival.
		hasPressureData := len(raw.RobustSurvival) > 0 || len(raw.DormantSurvival) > 0
		if hasPressureData {
			robustWeight := w.Survival * 0.80
			dormantWeight := w.Survival * 0.20

			// Design is only proven when code survives under change pressure
			// OR when the author actively builds (high production proves design through action).
			// Low production + high design = likely inflated by solo ownership.
			robustFactor := r.RobustSurvival/100*0.8 + 0.2 // 0.2 at Robust=0, 1.0 at Robust=100
			productionFactor := r.Production/100*0.8 + 0.2 // 0.2 at Prod=0, 1.0 at Prod=100
			designDamping := maxf(robustFactor, productionFactor)
			effectiveDesign := r.Design * designDamping

			r.Impact = r.Production*w.Production +
				r.Catalysis*w.Catalysis +
				r.RobustSurvival*robustWeight +
				r.DormantSurvival*dormantWeight +
				effectiveDesign*w.Design +
				r.Breadth*w.Breadth +
				r.DebtCleanup*w.DebtCleanup +
				r.Indispensability*w.Indispensability

			// Penalty: code that has never survived ANYWHERE — neither under
			// change pressure (robust) nor in stable, low-pressure modules
			// (dormant) — is fundamentally unproven. Apply a 0.8x multiplier.
			// Code that survives dormantly (RobustSurvival==0 but
			// DormantSurvival>0) HAS lasted, just not under churn, so it is
			// exempt: this hits pure churn/mass profiles, not stable-codebase
			// owners and founders whose foundational code lives in quiet modules.
			// Gated on raw survival mass so it is immune to normalization.
			if r.RawRobustSurv == 0 && r.RawDormantSurv == 0 {
				r.Impact *= 0.80
			}
		} else {
			r.Impact = r.Production*w.Production +
				r.Catalysis*w.Catalysis +
				r.Survival*w.Survival +
				r.Design*w.Design +
				r.Breadth*w.Breadth +
				r.DebtCleanup*w.DebtCleanup +
				r.Indispensability*w.Indispensability
		}

		r.Gravity = gravityScore(r)

		role, style, state := classifyTopology(r)
		r.Role = role.Name
		r.RoleConf = role.Confidence
		r.Style = style.Name
		r.StyleConf = style.Confidence
		r.State = state.Name
		r.StateConf = state.Confidence

		results = append(results, r)
	}

	// Sort by impact descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Impact > results[j].Impact
	})

	return results
}
