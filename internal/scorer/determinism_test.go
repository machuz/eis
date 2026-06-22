package scorer

import (
	"reflect"
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/config"
	"github.com/machuz/eis/v2/internal/git"
	"github.com/machuz/eis/v2/internal/metric"
)

// W-02 determinism regression: the same (git_sha, analysis_time) MUST always
// produce the same scores. EIS is a telescope — it observes; it does not
// generate. A scoring path that reaches for wall-clock time.Now() as its
// decay/classification reference drifts run-to-run for an unchanged git_sha,
// which is exactly the violation this test pins shut.
//
// These tests use a fixed synthetic blame/commit corpus (anonymized authors)
// and a PINNED analysisTime, then assert:
//  1. same pinned time, run twice → byte-identical Results;
//  2. same pinned time across a wall-clock gap → identical (no hidden now());
//  3. a different pinned time DOES move the time-sensitive outputs, proving the
//     envelope clock genuinely flows into decay + classification.

// synthAuthors / synthBlame / synthCommits build a small deterministic corpus
// anchored relative to a reference instant. Two authors so normalization has a
// real pool; survival is set up so it decays measurably as the reference moves.
func synthCommits(ref time.Time) []git.Commit {
	return []git.Commit{
		{Author: "author-a", Date: ref.AddDate(0, -6, 0), FileStats: []git.FileStat{{Filename: "core.go", Insertions: 100}}},
		{Author: "author-b", Date: ref.AddDate(0, -2, 0), FileStats: []git.FileStat{{Filename: "core.go", Insertions: 40}}},
		{Author: "author-b", Date: ref.AddDate(0, -1, 0), FileStats: []git.FileStat{{Filename: "util.go", Insertions: 30}}},
	}
}

func synthBlame(ref time.Time) []git.BlameLine {
	var lines []git.BlameLine
	// author-a's foundation still lives in core.go (older, decays more).
	for i := 0; i < 20; i++ {
		lines = append(lines, git.BlameLine{Filename: "core.go", Author: "author-a", CommitterTime: ref.AddDate(0, -5, 0)})
	}
	// author-b's lines, more recent.
	for i := 0; i < 12; i++ {
		lines = append(lines, git.BlameLine{Filename: "core.go", Author: "author-b", CommitterTime: ref.AddDate(0, -2, 0)})
	}
	for i := 0; i < 8; i++ {
		lines = append(lines, git.BlameLine{Filename: "util.go", Author: "author-b", CommitterTime: ref.AddDate(0, -1, 0)})
	}
	return lines
}

// scoreCorpus runs the full score-affecting path (survival + catalysis +
// ScoreAt) against an explicit analysisTime, exactly as the analyze CLI now
// does. corpusRef anchors WHERE the commits/blame sit on the calendar; refTime
// is the W-02 envelope clock the scores are computed against. Keeping them
// separate lets a test move the clock without moving the data.
func scoreCorpus(cfg *config.Config, corpusRef, refTime time.Time) []Result {
	commits := synthCommits(corpusRef)
	blame := synthBlame(corpusRef)

	raw := metric.NewRawScores()
	mr := metric.NewModuleResolver(nil)

	surv := metric.CalcSurvivalFull(blame, cfg.Tau, refTime, nil, 0, mr, nil, cfg.UntestedSurvivalWeight, nil)
	for a, v := range surv.Decayed {
		raw.Survival[a] = v
	}
	for a, v := range surv.Raw {
		raw.RawSurvival[a] = v
	}
	for a, v := range surv.Tested {
		raw.TestedSurvival[a] = v
	}
	for a, v := range surv.Untested {
		raw.UntestedSurvival[a] = v
	}

	cat := metric.CalcCatalysis(commits, blame, cfg.Tau, refTime)
	for a, v := range cat {
		raw.Catalysis[a] = v
	}

	// Minimal supporting raw fields so ScoreAt produces a full Result set.
	authorLastDate := map[string]time.Time{}
	for _, c := range commits {
		raw.TotalCommits[c.Author]++
		for _, fs := range c.FileStats {
			raw.LinesAdded[c.Author] += fs.Insertions
			raw.Production[c.Author] += float64(fs.Insertions)
		}
		if c.Date.After(authorLastDate[c.Author]) {
			authorLastDate[c.Author] = c.Date
		}
	}

	return ScoreAt(raw, cfg, authorLastDate, refTime)
}

// TestScoreAt_DeterministicForPinnedAnalysisTime proves that pinning the
// envelope clock makes the WHOLE scoring path reproducible: the same inputs and
// the same analysisTime yield deep-equal Results, with the time-load-bearing
// fields (Survival, RobustSurvival, Catalysis, Gravity, State, RecentlyActive,
// LastActiveDate) identical.
func TestScoreAt_DeterministicForPinnedAnalysisTime(t *testing.T) {
	cfg := config.Default()
	corpusRef := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	pinned := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)

	first := scoreCorpus(cfg, corpusRef, pinned)
	second := scoreCorpus(cfg, corpusRef, pinned)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same inputs + same pinned analysisTime must be byte-identical:\n first=%+v\nsecond=%+v", first, second)
	}

	// Spell out the time-sensitive fields too, so a regression that only moves
	// one of them is reported precisely rather than as an opaque DeepEqual fail.
	for i := range first {
		a, b := first[i], second[i]
		if a.Survival != b.Survival || a.RobustSurvival != b.RobustSurvival ||
			a.Catalysis != b.Catalysis || a.Gravity != b.Gravity ||
			a.State != b.State || a.RecentlyActive != b.RecentlyActive ||
			!a.LastActiveDate.Equal(b.LastActiveDate) {
			t.Fatalf("time-sensitive field drift for %q:\n a=%+v\n b=%+v", a.Author, a, b)
		}
	}
}

// TestScoreAt_IndependentOfWallClock proves the pinned time, not the real
// clock, is what the scores depend on: two calls separated by a real-time gap
// but given the SAME pinned analysisTime produce identical output. (Before the
// fix, the path reached for time.Now() internally, so this could differ.)
func TestScoreAt_IndependentOfWallClock(t *testing.T) {
	cfg := config.Default()
	corpusRef := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	pinned := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)

	before := scoreCorpus(cfg, corpusRef, pinned)
	// Advance the real wall clock between the two runs.
	time.Sleep(5 * time.Millisecond)
	after := scoreCorpus(cfg, corpusRef, pinned)

	if !reflect.DeepEqual(before, after) {
		t.Fatalf("pinned analysisTime must make output independent of wall clock:\n before=%+v\n after=%+v", before, after)
	}
}

// TestScoreAt_AnalysisTimeActuallyFlowsThrough is the negative control: if the
// envelope clock were ignored (e.g. a stray time.Now() shadowing it), moving
// analysisTime would NOT change the output. RecentlyActive is the cleanest
// witness because it is a HARD time threshold (gap vs active_days), immune to
// the relative normalization that can leave decayed Survival flat when all
// authors decay together. We anchor the corpus so its latest commit sits just
// inside active_days of the "near" clock; a large forward jump must then flip
// everyone to not-recently-active — confirming the time genuinely threads
// through classification, not decoration.
func TestScoreAt_AnalysisTimeActuallyFlowsThrough(t *testing.T) {
	cfg := config.Default()
	// Place the corpus so the latest commit (corpusRef - 1 month) is only a few
	// days before the near clock.
	corpusRef := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	nearClock := corpusRef.AddDate(0, -1, 0).AddDate(0, 0, 5) // 5 days after last commit
	farClock := corpusRef.AddDate(3, 0, 0)                    // 3 years later

	near := indexByAuthor(scoreCorpus(cfg, corpusRef, nearClock))
	far := indexByAuthor(scoreCorpus(cfg, corpusRef, farClock))

	// Sanity: at the near clock, recent committers are RecentlyActive.
	anyRecentNear := false
	for _, n := range near {
		if n.RecentlyActive {
			anyRecentNear = true
		}
	}
	if !anyRecentNear {
		t.Fatal("test setup: expected at least one RecentlyActive author at the near clock")
	}

	// Moving the envelope clock years forward MUST flip everyone off
	// RecentlyActive — proof the clock flows into classification.
	for a, f := range far {
		if f.RecentlyActive {
			t.Fatalf("author %q still RecentlyActive 3 years after last commit (active_days=%d) — analysisTime not flowing into classification (W-02 wiring broken)", a, cfg.ActiveDays)
		}
	}
}

func indexByAuthor(rs []Result) map[string]Result {
	m := make(map[string]Result, len(rs))
	for _, r := range rs {
		m[r.Author] = r
	}
	return m
}
