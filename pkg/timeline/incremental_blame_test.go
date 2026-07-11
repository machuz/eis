package timeline

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/config"
)

// buildIncrementalBlameFixtureRepo builds a repo tailored to exercise the
// per-file incremental blame cache (cache.BlameFileAtCommitKey):
//
//   - stable.go is committed once and never touched again, so every window
//     after the first must REUSE its cached blame rather than recompute it.
//   - churn.go is rewritten in several windows, so it must be RE-BLAMED each
//     time its last-touch SHA changes.
//   - old.go is renamed to new.go partway through, so the grouping-by-input-path
//     logic is exercised on lines git reports under a historical filename.
//
// Two authors keep the resulting signals non-trivial. Dates are forced so the
// fixture is deterministic across machines (same pattern as the other fixtures).
func buildIncrementalBlameFixtureRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	mustGit := func(env []string, args ...string) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(), env...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	mustGit(nil, "init", "-q", "-b", "main")
	mustGit(nil, "config", "user.email", "a@test")
	mustGit(nil, "config", "user.name", "alice")

	write := func(file, content string) {
		full := filepath.Join(dir, file)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	dateEnv := func(date, name, email string) []string {
		return []string{
			"GIT_AUTHOR_DATE=" + date + "T10:00:00+00:00",
			"GIT_COMMITTER_DATE=" + date + "T10:00:00+00:00",
			"GIT_AUTHOR_NAME=" + name, "GIT_AUTHOR_EMAIL=" + email,
			"GIT_COMMITTER_NAME=" + name, "GIT_COMMITTER_EMAIL=" + email,
		}
	}

	// 2024-01: stable.go (never touched again), churn.go v1, old.go
	write("stable.go", "package p\n\nfunc Stable() int { return 1 }\n")
	write("churn.go", "package p\n\nfunc Churn() int { return 1 }\n")
	write("old.go", "package p\n\nfunc Old() int { return 1 }\n")
	mustGit(nil, "add", "-A")
	mustGit(dateEnv("2024-01-15", "alice", "a@test"), "commit", "-q", "-m", "jan")

	// 2024-03: churn.go v2 (bob). stable.go + old.go unchanged → reused.
	write("churn.go", "package p\n\nfunc Churn() int { return 2 }\nfunc Extra() {}\n")
	mustGit(nil, "add", "-A")
	mustGit(dateEnv("2024-03-15", "bob", "b@test"), "commit", "-q", "-m", "mar")

	// 2024-05: rename old.go -> new.go (content unchanged), churn.go v3 (alice).
	mustGit(dateEnv("2024-05-15", "bob", "b@test"), "mv", "old.go", "new.go")
	write("churn.go", "package p\n\nfunc Churn() int { return 3 }\nfunc Extra() {}\nfunc More() {}\n")
	mustGit(nil, "add", "-A")
	mustGit(dateEnv("2024-05-15", "alice", "a@test"), "commit", "-q", "-m", "may")

	// 2024-07: new.go modified (alice). stable.go STILL untouched since Jan.
	write("new.go", "package p\n\nfunc Old() int { return 2 }\n")
	mustGit(nil, "add", "-A")
	mustGit(dateEnv("2024-07-15", "alice", "a@test"), "commit", "-q", "-m", "jul")

	return dir
}

// TestRun_IncrementalBlameMatchesFullBlame is the correctness gate for the
// per-file incremental blame cache. It runs the SAME fixture two ways:
//
//	CacheEnabled=false → every window re-blames all its files fresh (no reuse) =
//	                     the full-blame reference.
//	CacheEnabled=true  → files unchanged since an earlier window reuse their
//	                     cached per-file blame (the optimization under test).
//
// Both must return a byte-identical []DomainTimeline. If the last-touch key ever
// unified two different file contents, or reuse leaked a stale blame, the two
// runs diverge and this fails. Observation values are the product's core
// guarantee (the un-gameable metric), so this parity is non-negotiable.
//
// Both PerRepo modes are checked: false is the CLI default, true is the SaaS
// path with extra per-repo scoring (more fields that a cache bug could corrupt).
func TestRun_IncrementalBlameMatchesFullBlame(t *testing.T) {
	// Isolate the on-disk blame cache under a temp HOME so the CacheEnabled=true
	// run starts cold and cannot be perturbed by a real ~/.eis/cache.
	t.Setenv("HOME", t.TempDir())

	cfg := config.Default()
	// Sanity: the fixture must exercise the incremental (-M) path, not the
	// whole-boundary fallback — otherwise this test proves nothing.
	if !incrementalBlame(cfg.BlameMoveDetection) {
		t.Fatalf("config.Default().BlameMoveDetection=%q is not on the incremental path; test would be vacuous", cfg.BlameMoveDetection)
	}

	repo := buildIncrementalBlameFixtureRepo(t)

	for _, perRepo := range []bool{false, true} {
		perRepo := perRepo
		name := "PerRepoFalse"
		if perRepo {
			name = "PerRepoTrue"
		}
		t.Run(name, func(t *testing.T) {
			// Pin the analysis instant so both runs (cache off, then on) share
			// the same trailing window. BuildPeriods clamps the still-open final
			// window's End to Now, and window.End is that window's decay
			// reference — so an unpinned time.Now() would score it against two
			// different instants and drift in the last float digits, unrelated
			// to the incremental-vs-full blame equality this test asserts.
			fixedNow := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)
			run := func(cacheEnabled bool) []DomainTimeline {
				results, err := Run(
					context.Background(),
					Options{
						Span:              "1m",
						Since:             "2024-01-01",
						Workers:           2,
						PressureMode:      "include",
						PerRepo:           perRepo,
						PeriodConcurrency: 3,
						CacheEnabled:      cacheEnabled,
						Now:               fixedNow,
					},
					[]string{repo},
					cfg,
					&Callbacks{},
				)
				if err != nil {
					t.Fatalf("Run(cache=%v): %v", cacheEnabled, err)
				}
				return results
			}

			full := run(false)       // reference: no reuse, every file blamed fresh
			incremental := run(true) // per-file last-touch reuse across windows

			if len(full) == 0 {
				t.Fatal("fixture produced 0 DomainTimeline entries")
			}
			if !reflect.DeepEqual(full, incremental) {
				t.Errorf("incremental per-file blame diverged from full blame\n  full=%+v\n  incremental=%+v", full, incremental)
			}
		})
	}
}
