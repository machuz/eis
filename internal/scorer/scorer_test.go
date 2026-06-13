package scorer

import (
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/config"
	"github.com/machuz/eis/v2/internal/metric"
)

// TestScore_ExcludesMergeOnlyAuthors is the regression test for the bug where
// authors who only merged PRs (no actual code authorship) leaked into per-domain
// team rollups via the quality axis.
//
// Setup mirrors what pkg/analyzer accumulates:
//   - "alice" wrote real commits → TotalCommits=10, Production/Quality/etc populated.
//   - "merger" only merged PRs   → TotalCommits=0 (analyzer.go only increments it
//     for non-merge commits) but Quality is populated because metric.CalcQuality
//     is fed merge+non-merge commits.
//
// The bug: scorer.Score collected authors from the union of Production/Quality/...
// maps, so "merger" was emitted with non-zero Quality (and downstream Impact /
// Gravity) despite never authoring code in this domain.
//
// The fix: skip any author with raw.TotalCommits[author] == 0, matching the
// existing "code authorship" semantic of TotalCommits.
func TestScore_ExcludesMergeOnlyAuthors(t *testing.T) {
	raw := metric.NewRawScores()

	// alice — real author
	raw.Production["alice"] = 200 // per-day rate post-Run normalization
	raw.Quality["alice"] = 80
	raw.Survival["alice"] = 50
	raw.Design["alice"] = 40
	raw.Breadth["alice"] = 2
	raw.DebtCleanup["alice"] = 50
	raw.Indispensability["alice"] = 30
	raw.TotalCommits["alice"] = 10
	raw.LinesAdded["alice"] = 500
	raw.LinesDeleted["alice"] = 100

	// merger — merge-only. analyzer.go does NOT increment TotalCommits for merge
	// commits, so TotalCommits[merger] stays 0 even though Quality is non-zero
	// (CalcQuality is fed both merge and non-merge commits).
	raw.Quality["merger"] = 43.75
	// Production / Survival / Design / Indispensability all stay at 0 for merger.

	cfg := config.Default()
	authorLastDate := map[string]time.Time{
		"alice":  time.Now(),
		"merger": time.Now(),
	}

	results := Score(raw, cfg, authorLastDate)

	for _, r := range results {
		if r.Author == "merger" {
			t.Fatalf("merge-only author 'merger' should not appear in scorer results, got %+v", r)
		}
	}

	// Sanity: alice (real author) is still present.
	foundAlice := false
	for _, r := range results {
		if r.Author == "alice" {
			foundAlice = true
			if r.TotalCommits != 10 {
				t.Errorf("alice TotalCommits = %d, want 10", r.TotalCommits)
			}
			break
		}
	}
	if !foundAlice {
		t.Fatalf("real author 'alice' missing from results; got %d results", len(results))
	}
}

// TestScore_ZeroCommitAuthorWithImpact guards against authors that have non-zero
// computed metrics (Quality, Design, etc.) but TotalCommits=0. They are by
// definition merge-only / phantom contributors and must be filtered.
func TestScore_ZeroCommitAuthorWithImpact(t *testing.T) {
	raw := metric.NewRawScores()
	// Only Quality populated, like a merge-only author.
	raw.Quality["ghost"] = 50

	cfg := config.Default()
	results := Score(raw, cfg, map[string]time.Time{"ghost": time.Now()})

	if len(results) != 0 {
		t.Fatalf("expected 0 results for an author with TotalCommits=0, got %d: %+v", len(results), results)
	}
}

// TestScore_DormantSurvivalEscapesUnprovenPenalty guards the fix for the
// robust==0 "unproven" penalty over-deflating stable-codebase owners. The 0.8x
// Impact penalty must fire only when code survives NOWHERE (robust==0 AND
// dormant==0 — a pure churn/mass profile), NOT when code survives in quiet,
// low-pressure modules (robust==0 but dormant>0), which is the signature of a
// founder whose foundational code sits in stable modules.
func TestScore_DormantSurvivalEscapesUnprovenPenalty(t *testing.T) {
	raw := metric.NewRawScores()

	// Register both in Authors() (which unions only Production/Quality/Survival/
	// Design/Breadth/Debt/Indisp). Survival here is the legacy axis, unused in
	// the robust/dormant pressure path, so it does not affect Impact below.
	raw.Survival["stable"] = 50
	raw.Survival["churn"] = 50

	// stable: no robust survival, but its code lives on in low-pressure modules.
	raw.RobustSurvival["stable"] = 0
	raw.DormantSurvival["stable"] = 100
	raw.TotalCommits["stable"] = 10

	// churn: code survives nowhere — robust AND dormant are both zero.
	raw.RobustSurvival["churn"] = 0
	raw.DormantSurvival["churn"] = 0
	raw.TotalCommits["churn"] = 10

	cfg := config.Default()
	now := time.Now()
	results := Score(raw, cfg, map[string]time.Time{"stable": now, "churn": now})

	var stable, churn *Result
	for i := range results {
		switch results[i].Author {
		case "stable":
			stable = &results[i]
		case "churn":
			churn = &results[i]
		}
	}
	if stable == nil || churn == nil {
		t.Fatalf("missing results: stable=%v churn=%v", stable, churn)
	}

	// Both share weights: Survival=0.25 (dormant slice 0.20), DebtCleanup=0.15
	// with the neutral 50 default. Normalized dormant: stable=100, churn=0.
	//   stable Impact = 100*(0.25*0.20) + 50*0.15 = 5 + 7.5 = 12.5   (NOT penalized)
	//   churn  Impact = (0          + 50*0.15) * 0.80 = 7.5 * 0.80 = 6.0 (penalized)
	const tol = 0.01
	if got := stable.Impact; got < 12.5-tol || got > 12.5+tol {
		t.Errorf("stable (dormant survivor) Impact = %.4f, want 12.5 (no 0.8x penalty); "+
			"old robust==0 logic would have given 10.0", got)
	}
	if got := churn.Impact; got < 6.0-tol || got > 6.0+tol {
		t.Errorf("churn (survives nowhere) Impact = %.4f, want 6.0 (0.8x penalty applied)", got)
	}
}
