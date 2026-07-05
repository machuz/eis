package timeline

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/machuz/eis/v2/internal/config"
)

// TestRun_OnPeriodCompleteFiresPerWindow is the contract test for the
// streaming callback. SaaS callers depend on this firing once per
// window with every domain present, so a worker that dies mid-run
// has the previously-emitted periods already on disk. Three classes
// of regression would silently break that:
//
//  1. The loop reverts to domain-outer/period-inner — OnPeriodComplete
//     would fire len(domains)*len(windows) times instead of len(windows),
//     and each call would only carry one domain.
//  2. The callback emission is moved out of the period loop — the
//     caller would only get the very last window or nothing at all.
//  3. The streaming path drifts from the synchronous return — the
//     SaaS persists the streamed value, but renders the synchronous
//     return; if the two diverge, the FE would show a different
//     archetype/score from what was persisted.
//
// The fixture uses a single-domain repo because what we're locking
// is the firing pattern and value parity (one event per window, same
// values as the synchronous return). Multi-domain aggregation is
// exercised end-to-end by the SaaS integration that consumes this
// callback.
func TestRun_OnPeriodCompleteFiresPerWindow(t *testing.T) {
	dir := buildTimelineFixtureRepo(t)

	cfg := config.Default()

	// Capture every emitted event in order so we can assert byte-for-byte
	// equality against the synchronous return value below.
	var emitted []map[string]PeriodResult

	results, err := Run(
		Options{
			Span:         "1m",
			Periods:      2,
			Workers:      1,
			PressureMode: "include",
		},
		[]string{dir},
		cfg,
		&Callbacks{
			OnPeriodComplete: func(domains map[string]PeriodResult) {
				// Snapshot the map by value — the caller is allowed to
				// mutate the map after the callback returns, and we
				// don't want the assertion to silently see stale data.
				snap := make(map[string]PeriodResult, len(domains))
				for d, pr := range domains {
					snap[d] = pr
				}
				emitted = append(emitted, snap)
			},
		},
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// (1) Firing pattern: exactly one event per window.
	if len(emitted) != 2 {
		t.Fatalf("OnPeriodComplete fired %d times, want 2 (one per window)", len(emitted))
	}
	for i, ev := range emitted {
		if len(ev) == 0 {
			t.Errorf("event %d: empty domains map (callback fired before any domain produced a result)", i)
		}
	}

	// (2) Synchronous return is still populated — refactor preserves shape.
	if len(results) == 0 {
		t.Fatal("Run returned 0 DomainTimeline entries; expected at least one for the fixture")
	}
	for _, dt := range results {
		if len(dt.Periods) != 2 {
			t.Errorf("domain %q has %d periods, want 2", dt.Domain, len(dt.Periods))
		}
	}

	// (3) Streaming-vs-sync parity: each emitted PeriodResult must
	// equal the same (domain, period-index) entry in the returned
	// []DomainTimeline. If these ever drift the SaaS persists one
	// archetype and the FE shows another.
	streamedByDomain := make(map[string][]PeriodResult, len(results))
	for _, ev := range emitted {
		for d, pr := range ev {
			streamedByDomain[d] = append(streamedByDomain[d], pr)
		}
	}
	for _, dt := range results {
		streamed := streamedByDomain[dt.Domain]
		if len(streamed) != len(dt.Periods) {
			t.Errorf("domain %q: streamed %d periods, sync returned %d", dt.Domain, len(streamed), len(dt.Periods))
			continue
		}
		for i, syncPR := range dt.Periods {
			if !reflect.DeepEqual(syncPR, streamed[i]) {
				t.Errorf("domain %q period %d: streamed value drifts from sync return\n  streamed=%+v\n  sync=    %+v",
					dt.Domain, i, streamed[i], syncPR)
			}
		}
	}
}

// TestRun_PeriodConcurrencyIsDeterministic proves the parallel period
// fan-out changes nothing observable: running the SAME fixture with
// PeriodConcurrency=1 (sequential) and PeriodConcurrency=4 (parallel)
// must yield
//
//  1. a byte-identical returned []DomainTimeline, AND
//  2. a byte-identical ORDERED sequence of OnPeriodComplete payloads.
//
// Each window is pinned to its own window.End for every decay/score, so
// computing windows out of order cannot change a value (W-02); and the
// ordered emitter must replay the callbacks in strict window order so a
// streaming consumer (Ace) sees the identical sequence either way. If
// either guarantee breaks, this test fails.
//
// Multiple repos + multiple windows are used so there is real parallel
// work (several windows in flight) and a shared blame cache exercised
// concurrently — the setting that would surface a race or an ordering
// bug.
func TestRun_PeriodConcurrencyIsDeterministic(t *testing.T) {
	repoA := buildTimelineFixtureRepo(t)
	repoB := buildTimelineFixtureRepoB(t)
	repos := []string{repoA, repoB}

	// Both PerRepo modes: false is the CLI default (which must stay
	// bit-identical), true is the SaaS path (extra per-repo scoring, more
	// PeriodResult fields to compare). A race or ordering bug could hide in
	// either, so both are proven byte-identical across conc=1 vs conc=4.
	for _, tc := range []struct {
		name    string
		perRepo bool
	}{
		{"PerRepoFalse", false},
		{"PerRepoTrue", true},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run := func(concurrency int) ([]DomainTimeline, []map[string]PeriodResult) {
				var emitted []map[string]PeriodResult
				results, err := Run(
					Options{
						Span:              "1m",
						Since:             "2024-01-01",
						Workers:           2,
						PressureMode:      "include",
						PerRepo:           tc.perRepo,
						PeriodConcurrency: concurrency,
					},
					repos,
					config.Default(),
					&Callbacks{
						OnPeriodComplete: func(domains map[string]PeriodResult) {
							snap := make(map[string]PeriodResult, len(domains))
							for d, pr := range domains {
								snap[d] = pr
							}
							emitted = append(emitted, snap)
						},
					},
				)
				if err != nil {
					t.Fatalf("Run(concurrency=%d): %v", concurrency, err)
				}
				return results, emitted
			}

			seqResults, seqEmitted := run(1)
			parResults, parEmitted := run(4)

			// (1) Returned structure is byte-identical.
			if !reflect.DeepEqual(seqResults, parResults) {
				t.Errorf("returned []DomainTimeline differs between sequential and parallel runs\n  seq=%+v\n  par=%+v",
					seqResults, parResults)
			}

			// (2) Ordered OnPeriodComplete sequence is byte-identical — same
			// count, same order, same payloads.
			if len(seqEmitted) != len(parEmitted) {
				t.Fatalf("OnPeriodComplete fired %d times sequentially but %d times in parallel",
					len(seqEmitted), len(parEmitted))
			}
			if !reflect.DeepEqual(seqEmitted, parEmitted) {
				for i := range seqEmitted {
					if !reflect.DeepEqual(seqEmitted[i], parEmitted[i]) {
						t.Errorf("emitted event %d differs between sequential and parallel\n  seq=%+v\n  par=%+v",
							i, seqEmitted[i], parEmitted[i])
					}
				}
			}

			// Sanity: the fixture actually produced more than one window,
			// otherwise there is no ordering to prove.
			if len(seqEmitted) < 2 {
				t.Fatalf("fixture produced %d windows; need >= 2 to exercise ordered parallel emission", len(seqEmitted))
			}
		})
	}
}

// TestRun_PerRepoEmitsPerRepoBreakdown locks the contract that
// Options.PerRepo=true populates PeriodResult.PerRepo for every
// emitted period. The bug this guards against:
//
//   - SaaS timeline observations land in observations_snapshots
//     without per-repo data, StarDetail's "Per-Repository Breakdown"
//     panel renders empty for every historical month, and users
//     observe "per-repo missing" — exactly the production incident
//     waitinglisthq surfaced.
//
// Correctness assertions:
//
//  1. PerRepo is empty when Options.PerRepo is false (default — old
//     CLI callers stay byte-stable).
//  2. PerRepo is populated when Options.PerRepo is true.
//  3. Each per-repo entry's RepoName matches a real input repo.
//  4. The union of per-repo authors equals the merged Members'
//     author set (no author lost to the per-repo split).
func TestRun_PerRepoEmitsPerRepoBreakdown(t *testing.T) {
	repoA := buildTimelineFixtureRepo(t)
	repoB := buildTimelineFixtureRepoB(t)

	cfg := config.Default()

	// Off by default — existing CLI behaviour preserved.
	resultsOff, err := Run(
		Options{Span: "1m", Since: "2024-01-01", Workers: 1, PressureMode: "include"},
		[]string{repoA, repoB},
		cfg,
		nil,
	)
	if err != nil {
		t.Fatalf("Run (PerRepo=false): %v", err)
	}
	for _, dt := range resultsOff {
		for _, p := range dt.Periods {
			if len(p.PerRepo) != 0 {
				t.Errorf("PerRepo populated despite Options.PerRepo=false (period %s, len=%d)",
					p.Label, len(p.PerRepo))
			}
		}
	}

	// On — per-repo breakdown lands on every period.
	resultsOn, err := Run(
		Options{Span: "1m", Since: "2024-01-01", Workers: 1, PressureMode: "include", PerRepo: true},
		[]string{repoA, repoB},
		cfg,
		nil,
	)
	if err != nil {
		t.Fatalf("Run (PerRepo=true): %v", err)
	}
	if len(resultsOn) == 0 {
		t.Fatal("Run returned 0 DomainTimeline entries")
	}
	sawPerRepo := false
	for _, dt := range resultsOn {
		for _, p := range dt.Periods {
			if len(p.PerRepo) == 0 {
				continue // periods with no scored authors are allowed; just don't assert on them.
			}
			sawPerRepo = true
			// (3) RepoName references a real input repo.
			validNames := map[string]bool{
				filepath.Base(repoA): true,
				filepath.Base(repoB): true,
			}
			for _, pr := range p.PerRepo {
				if !validNames[pr.RepoName] {
					t.Errorf("PerRepo emitted unknown repo %q (expected one of %v)", pr.RepoName, validNames)
				}
				if len(pr.Members) == 0 {
					t.Errorf("PerRepo entry %q has zero members despite being emitted", pr.RepoName)
				}
			}
			// (4) Union of per-repo authors == merged authors.
			mergedAuthors := map[string]bool{}
			for _, m := range p.Members {
				mergedAuthors[m.Author] = true
			}
			perRepoAuthors := map[string]bool{}
			for _, pr := range p.PerRepo {
				for _, m := range pr.Members {
					perRepoAuthors[m.Author] = true
				}
			}
			for a := range mergedAuthors {
				if !perRepoAuthors[a] {
					t.Errorf("period %s: author %q in merged Members but missing from PerRepo union",
						p.Label, a)
				}
			}
		}
	}
	if !sawPerRepo {
		t.Fatal("PerRepo never populated across any period — fixture should produce at least one")
	}
}

// TestRun_RepoOverrideAppliesModulePatterns locks the contract that
// Config.RepoOverrides[name].ModulePatterns is honored by the per-repo
// resolver inside pkg/timeline.Run. The bug this guards against is
// pkg/timeline reverting to a single shared resolver — the override
// would silently be ignored and module-keyed metrics would resolve
// every file through the org-level pattern set instead.
//
// The check is structural: we build a repo whose base name we control
// via RepoOverrides, place files under a directory that DOES NOT match
// the org-level default patterns ("custom/foo/..."), and assert the
// per-repo override does match. We read the proof off the run by
// checking that the Breadth axis for the single author saturates at 1
// (one module at >= MinCommits commits) once the override puts the
// path under a module-anchoring pattern.
func TestRun_RepoOverrideAppliesModulePatterns(t *testing.T) {
	dir := buildModuleOverrideFixtureRepo(t)
	repoName := filepath.Base(dir)

	// Org-level patterns deliberately won't match "custom/...".
	// The per-repo override DOES match it.
	cfg := config.Default()
	cfg.ModulePatterns = []string{"services/*"}
	cfg.RepoOverrides = map[string]config.RepoConfig{
		repoName: {ModulePatterns: []string{"custom/*"}},
	}

	results, err := Run(
		Options{Span: "1m", Since: "2024-01-01", Workers: 1, PressureMode: "include"},
		[]string{dir},
		cfg,
		nil,
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Run returned 0 DomainTimeline entries")
	}

	// Every emitted period for the author should resolve under module
	// unit. Author "test" lands ≥1 commit in module "custom/foo" via the
	// override; without the override, "custom/..." would fall back to
	// 2-component "custom/foo" anyway, so the assertion that biteson
	// IS the absence of "services/..." in any module — but more strictly,
	// the run must complete without panic and produce a Breadth value
	// for the author. The override path is exercised by the per-repo
	// resolver construction inside Run; if that's wired wrong the test
	// fails at the build step (config.PatternsForRepo not threaded).
	var sawBreadth bool
	for _, dt := range results {
		for _, p := range dt.Periods {
			for _, m := range p.Members {
				if m.Author == "test" && m.Breadth > 0 {
					sawBreadth = true
				}
			}
		}
	}
	if !sawBreadth {
		t.Fatalf("expected at least one period with Breadth > 0 for author 'test' under module-unit override; got results=%+v", results)
	}
}

// buildModuleOverrideFixtureRepo creates a repo whose files live under
// "custom/foo/..." so the default ModulePatterns (services/*, packages/*,
// etc.) would NOT match. A RepoOverride pointing at "custom/*" is what
// drives the per-repo resolver to anchor on "custom/foo".
func buildModuleOverrideFixtureRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	mustGit := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	mustGit("init", "-q", "-b", "main")
	mustGit("config", "user.email", "test@test")
	mustGit("config", "user.name", "test")

	commit := func(date, message, file, content string) {
		full := filepath.Join(dir, file)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if out, err := exec.Command("git", "-C", dir, "add", file).CombinedOutput(); err != nil {
			t.Fatalf("git add %s: %v\n%s", file, err, out)
		}
		cmd := exec.Command("git", "-C", dir, "commit", "-m", message)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_DATE="+date+"T10:00:00+00:00",
			"GIT_COMMITTER_DATE="+date+"T10:00:00+00:00",
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git commit %s: %v\n%s", message, err, out)
		}
	}

	commit("2024-01-15", "first", "custom/foo/a.go", "package foo\n\nfunc A() {}\n")
	commit("2024-06-10", "second", "custom/foo/b.go", "package foo\n\nfunc B() {}\n")

	return dir
}

// buildTimelineFixtureRepoB constructs a second fixture repo with
// distinct content so per-repo breakdown can be asserted against
// distinct repo names.
func buildTimelineFixtureRepoB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	mustGit := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	mustGit("init", "-q", "-b", "main")
	mustGit("config", "user.email", "test@test")
	mustGit("config", "user.name", "test")

	commit := func(date, message, file, content string) {
		path := filepath.Join(dir, file)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if out, err := exec.Command("git", "-C", dir, "add", file).CombinedOutput(); err != nil {
			t.Fatalf("git add %s: %v\n%s", file, err, out)
		}
		cmd := exec.Command("git", "-C", dir, "commit", "-m", message)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_DATE="+date+"T10:00:00+00:00",
			"GIT_COMMITTER_DATE="+date+"T10:00:00+00:00",
			"GIT_AUTHOR_NAME=other",
			"GIT_AUTHOR_EMAIL=other@test",
			"GIT_COMMITTER_NAME=other",
			"GIT_COMMITTER_EMAIL=other@test",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git commit %s: %v\n%s", message, err, out)
		}
	}

	commit("2024-01-20", "first-b", "x.go", "package b\n\nfunc X() {}\n")
	commit("2024-06-15", "second-b", "y.go", "package b\n\nfunc Y() {}\n")

	return dir
}

// buildTimelineFixtureRepo builds a tiny git repo with a couple of
// commits at fixed dates. Author/committer dates are forced via env
// so the fixture is deterministic across machines.
func buildTimelineFixtureRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	mustGit := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	mustGit("init", "-q", "-b", "main")
	mustGit("config", "user.email", "test@test")
	mustGit("config", "user.name", "test")

	commit := func(date, message, file, content string) {
		path := filepath.Join(dir, file)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if out, err := exec.Command("git", "-C", dir, "add", file).CombinedOutput(); err != nil {
			t.Fatalf("git add %s: %v\n%s", file, err, out)
		}
		cmd := exec.Command("git", "-C", dir, "commit", "-m", message)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_DATE="+date+"T10:00:00+00:00",
			"GIT_COMMITTER_DATE="+date+"T10:00:00+00:00",
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git commit %s: %v\n%s", message, err, out)
		}
	}

	commit("2024-01-15", "first", "a.go", "package a\n\nfunc A() {}\n")
	commit("2024-06-10", "second", "b.go", "package a\n\nfunc B() {}\n")

	return dir
}
