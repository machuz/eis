package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestAnalyzeStreamingMatchesMaterialized is the value-invariance gate for the
// streaming ingest: the same repo, analyzed once through the materialized path
// and once through the streaming path, must produce byte-identical JSON. Both
// feed the same CommitAggregator, so this locks the two feed strategies together.
func TestAnalyzeStreamingMatchesMaterialized(t *testing.T) {
	repo := buildRichFixture(t)

	// Pin the envelope clock so decay/classification is deterministic.
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	opts := AnalyzeOptions{
		Format:       "json",
		Workers:      4,
		PressureMode: "include",
		NoCache:      true,
		AnalysisTime: at,
	}

	run := func() string {
		dr, cfg, gotAt, err := RunAnalyzePipeline(opts, []string{repo})
		if err != nil {
			t.Fatalf("RunAnalyzePipeline: %v", err)
		}
		out, _ := captureStd(t, func() {
			if err := outputAnalyzeResults(dr, cfg, "json", gotAt); err != nil {
				t.Fatalf("output: %v", err)
			}
		})
		return out
	}

	origThresh := streamingCommitThreshold

	streamingCommitThreshold = 1 << 30 // force materialized
	materialized := run()

	streamingCommitThreshold = 0 // force streaming
	streaming := run()

	streamingCommitThreshold = origThresh

	if materialized != streaming {
		// Find the first differing line to make failures actionable.
		ml := splitLines(materialized)
		sl := splitLines(streaming)
		for i := 0; i < len(ml) && i < len(sl); i++ {
			if ml[i] != sl[i] {
				t.Fatalf("streaming diverges at line %d:\n  materialized: %s\n  streaming:    %s", i, ml[i], sl[i])
			}
		}
		t.Fatalf("streaming output differs in length: materialized=%d streaming=%d lines", len(ml), len(sl))
	}
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

// buildRichFixture creates a small but varied history: multiple authors, two
// display names sharing one email (identity collapse), a co-authored commit,
// several modules, comment-heavy code (comment filter), and a revert.
func buildRichFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git := func(env []string, args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), env...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	write := func(rel, content string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git(nil, "init", "-q")
	git(nil, "config", "user.name", "seed")
	git(nil, "config", "user.email", "seed@x.com")

	type who struct{ name, email string }
	env := func(w who, day int) []string {
		d := fmt.Sprintf("2025-01-%02dT12:00:00", day)
		return []string{
			"GIT_AUTHOR_NAME=" + w.name, "GIT_AUTHOR_EMAIL=" + w.email,
			"GIT_COMMITTER_NAME=" + w.name, "GIT_COMMITTER_EMAIL=" + w.email,
			"GIT_AUTHOR_DATE=" + d, "GIT_COMMITTER_DATE=" + d,
		}
	}
	alice := who{"Alice A", "alice@x.com"}
	aliceAlt := who{"alice", "alice@x.com"} // same email, different display name
	bob := who{"Bob B", "bob@x.com"}

	day := 1
	commitAs := func(w who, msg, trailer string) {
		git(nil, "add", "-A")
		args := []string{"commit", "-q", "-m", msg}
		if trailer != "" {
			args = append(args, "-m", trailer)
		}
		git(env(w, day), args...)
		day++
	}

	for i := 0; i < 12; i++ {
		write("backend/api.go", fmt.Sprintf("package backend\n// handler %d\nfunc H%d() int { return %d }\n", i, i, i))
		write("backend/db.go", fmt.Sprintf("package backend\nvar Q%d = %q\n", i, fmt.Sprintf("select %d", i)))
		write("frontend/app.ts", fmt.Sprintf("// comment %d\nexport const A%d = %d;\n", i, i, i))
		write("docs/readme.md", fmt.Sprintf("# Doc\n\nrev %d\n", i))
		w := alice
		if i%3 == 1 {
			w = bob
		} else if i%3 == 2 {
			w = aliceAlt
		}
		trailer := ""
		if i == 5 {
			trailer = "Co-authored-by: Bob B <bob@x.com>"
		}
		commitAs(w, fmt.Sprintf("change %d", i), trailer)
	}

	// A commit then its revert — both should be excluded from metrics.
	write("backend/temp.go", "package backend\nfunc Temp() {}\n")
	commitAs(bob, "add temp", "")
	git(env(bob, day), "revert", "--no-edit", "HEAD")
	day++

	return dir
}
