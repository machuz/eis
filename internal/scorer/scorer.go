package scorer

import (
	"sort"
	"time"

	"github.com/machuz/engineering-impact-score/internal/config"
	"github.com/machuz/engineering-impact-score/internal/metric"
)

type Result struct {
	Author           string
	Production       float64
	Quality          float64
	Survival         float64
	RawSurvival      float64 // normalized raw blame (no decay), used for archetype detection
	RobustSurvival   float64 // survival in high change-pressure modules
	DormantSurvival  float64 // survival in low change-pressure modules
	RawRobustSurv    float64 // raw (pre-normalize) robust survival, for archetype detection
	RawDormantSurv   float64 // raw (pre-normalize) dormant survival, for archetype detection
	Design           float64
	Breadth          float64
	DebtCleanup      float64
	Indispensability float64
	Gravity          float64 // structural influence: f(Indispensability, Breadth, Design)
	Total            float64
	TotalCommits   int
	RecentlyActive bool    // true if author has commits within active_days (default 30)
	Role           string  // Role axis: what they contribute (Architect, Anchor, Cleaner, Producer, Specialist, —)
	RoleConf       float64 // Role confidence (0.0-1.0)
	Style          string  // Style axis: how they contribute (Builder, Resilient, Rescue, Churn, Mass, Balanced, Spread, —)
	StyleConf      float64 // Style confidence (0.0-1.0)
	State          string  // State axis: lifecycle phase (Active, Growing, Former, Silent, Fragile, —)
	StateConf      float64 // State confidence (0.0-1.0)
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
	normDesign := Normalize(raw.Design)
	normIndisp := Normalize(raw.Indispensability)
	normRawSurv := Normalize(raw.RawSurvival)

	// Quality is already on 0-100 absolute scale (100 - fix_ratio), use directly
	// This makes Quality scores comparable across organizations
	normQual := raw.Quality

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
		// Determine if author has been active in last 6 months
		recentlyActive := false
		if lastDate, ok := authorLastDate[author]; ok {
			recentlyActive = refTime.Sub(lastDate).Hours()/24 <= float64(cfg.ActiveDays)
		}

		r := Result{
			Author:           author,
			Production:       normProd[author],
			Quality:          normQual[author],
			Survival:         normSurv[author],
			RobustSurvival:   normRobustSurv[author],
			DormantSurvival:  normDormantSurv[author],
			RawRobustSurv:    raw.RobustSurvival[author],
			RawDormantSurv:   raw.DormantSurvival[author],
			Design:           normDesign[author],
			Breadth:          normBreadth[author],
			DebtCleanup:      normDebt[author],
			Indispensability: normIndisp[author],
			RawSurvival:      normRawSurv[author],
			TotalCommits:     raw.TotalCommits[author],
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
			productionFactor := r.Production/100*0.8 + 0.2  // 0.2 at Prod=0, 1.0 at Prod=100
			designDamping := maxf(robustFactor, productionFactor)
			effectiveDesign := r.Design * designDamping

			r.Total = r.Production*w.Production +
				r.Quality*w.Quality +
				r.RobustSurvival*robustWeight +
				r.DormantSurvival*dormantWeight +
				effectiveDesign*w.Design +
				r.Breadth*w.Breadth +
				r.DebtCleanup*w.DebtCleanup +
				r.Indispensability*w.Indispensability

			// Penalty: code that has never survived under change pressure
			// is fundamentally unproven. Apply 0.8x multiplier to Total.
			if r.RobustSurvival == 0 {
				r.Total *= 0.80
			}
		} else {
			r.Total = r.Production*w.Production +
				r.Quality*w.Quality +
				r.Survival*w.Survival +
				r.Design*w.Design +
				r.Breadth*w.Breadth +
				r.DebtCleanup*w.DebtCleanup +
				r.Indispensability*w.Indispensability
		}

		// Gravity: structural influence on the system.
		// Weighted combination of the three axes that determine how much
		// the system's shape depends on this engineer's work.
		r.Gravity = r.Indispensability*0.40 + r.Breadth*0.30 + r.Design*0.30

		role, style, state := classifyTopology(r)
		r.Role = role.Name
		r.RoleConf = role.Confidence
		r.Style = style.Name
		r.StyleConf = style.Confidence
		r.State = state.Name
		r.StateConf = state.Confidence

		results = append(results, r)
	}

	// Sort by total descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Total > results[j].Total
	})

	return results
}
