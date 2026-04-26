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
