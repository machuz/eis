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
	Design           float64
	Breadth          float64
	DebtCleanup      float64
	Indispensability float64
	Total            float64
	TotalCommits     int
	RecentlyActive   bool   // true if author has commits within last 6 months
	Archetype        string // primary archetype name
	ArchetypeConf    float64        // primary archetype confidence (0.0-1.0)
	Secondary        ArchetypeMatch // second-best archetype match
}

func Score(raw *metric.RawScores, cfg *config.Config, authorLastDate map[string]time.Time) []Result {
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
	normDesign := Normalize(raw.Design)
	normIndisp := Normalize(raw.Indispensability)
	normRawSurv := Normalize(raw.RawSurvival)

	// Quality is already on 0-100 absolute scale (100 - fix_ratio), use directly
	// This makes Quality scores comparable across organizations
	normQual := raw.Quality

	// Breadth: relative scale — normalized within the group
	normBreadth := Normalize(raw.Breadth)

	// Debt is already on 0-100 scale, use directly
	normDebt := raw.DebtCleanup

	// Collect all authors
	authors := raw.Authors()
	w := cfg.Weights

	var results []Result
	for _, author := range authors {
		// Determine if author has been active in last 6 months
		recentlyActive := false
		if lastDate, ok := authorLastDate[author]; ok {
			recentlyActive = time.Since(lastDate).Hours()/24 <= 180
		}

		r := Result{
			Author:           author,
			Production:       normProd[author],
			Quality:          normQual[author],
			Survival:         normSurv[author],
			Design:           normDesign[author],
			Breadth:          normBreadth[author],
			DebtCleanup:      normDebt[author],
			Indispensability: normIndisp[author],
			RawSurvival:      normRawSurv[author],
			TotalCommits:     raw.TotalCommits[author],
			RecentlyActive:   recentlyActive,
		}

		r.Total = r.Production*w.Production +
			r.Quality*w.Quality +
			r.Survival*w.Survival +
			r.Design*w.Design +
			r.Breadth*w.Breadth +
			r.DebtCleanup*w.DebtCleanup +
			r.Indispensability*w.Indispensability

		primary, secondary := classifyArchetypeWithConfidence(r)
		r.Archetype = primary.Name
		r.ArchetypeConf = primary.Confidence
		r.Secondary = secondary

		results = append(results, r)
	}

	// Sort by total descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Total > results[j].Total
	})

	return results
}
