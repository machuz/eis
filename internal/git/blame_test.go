package git

import (
	"context"
	"strings"
	"testing"
	"time"
)

// Blame of a Go file that mixes code and comments must only yield BlameLines for
// real code lines. Verifies Survival/Indispensability cannot be inflated by
// comment-heavy files.
func TestBlameFile_FiltersComments(t *testing.T) {
	dir := newTempRepo(t)
	writeFile(t, dir, "m.go", `package m

// top-level comment
func A() {}

/*
 * block comment
 */
func B() {}
`)
	commit(t, dir, "init")

	lines, err := BlameFile(context.Background(), dir, "m.go")
	if err != nil {
		t.Fatal(err)
	}
	// Real code lines: "package m", "func A() {}", "func B() {}" = 3.
	if got, want := len(lines), 3; got != want {
		t.Errorf("blame code lines = %d, want %d", got, want)
	}
}

// Prose file blame should count every line (used for research/paper preservation).
func TestBlameFile_MarkdownUnfiltered(t *testing.T) {
	dir := newTempRepo(t)
	writeFile(t, dir, "doc.md", "# Title\n\nBody paragraph.\n\n## Section\n")
	commit(t, dir, "init")

	lines, err := BlameFile(context.Background(), dir, "doc.md")
	if err != nil {
		t.Fatal(err)
	}
	// 5 raw lines including blanks.
	if got, want := len(lines), 5; got != want {
		t.Errorf("markdown blame lines = %d, want %d", got, want)
	}
}

// A 5MB single-line SQL dump used to deadlock blame: bufio.Scanner aborts
// with bufio.ErrTooLong, the loop exits without draining stdout, git
// blocks writing into a full pipe, cmd.Wait blocks waiting for git, and
// the parent context (timeline boundary) had no timeout. The scanner
// buffer is now 64MB and stdout is drained before Wait, so the call
// must complete inside a short context-bounded window.
func TestBlameFileAtCommit_DoesNotHangOnHugeLine(t *testing.T) {
	dir := newTempRepo(t)
	// 5MB of 'a' on a single line (no newline) — the exact pathological
	// shape that triggered the original hang.
	writeFile(t, dir, "dump.sql", strings.Repeat("a", 5*1024*1024))
	commit(t, dir, "huge dump")

	head, err := HeadHash(context.Background(), dir)
	if err != nil {
		t.Fatalf("head hash: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		// We don't care about the result, only that this returns.
		_, _ = BlameFileAtCommit(ctx, dir, head, "dump.sql")
		close(done)
	}()

	select {
	case <-done:
		// good — returned without deadlock
	case <-time.After(10 * time.Second):
		t.Fatal("BlameFileAtCommit hung on 5MB single-line file")
	}
}

// FilterFilesBySize must drop blobs whose size exceeds maxBytes and keep
// everything else. maxBytes <= 0 disables filtering.
func TestFilterFilesBySize(t *testing.T) {
	dir := newTempRepo(t)
	writeFile(t, dir, "small.go", "package m\n")
	// 2MB single-line file
	writeFile(t, dir, "big.sql", strings.Repeat("x", 2*1024*1024))
	writeFile(t, dir, "tiny.txt", "ok\n")
	commit(t, dir, "init")

	head, err := HeadHash(context.Background(), dir)
	if err != nil {
		t.Fatalf("head hash: %v", err)
	}

	all := []string{"small.go", "big.sql", "tiny.txt"}

	var skipped []string
	verbose := func(msg string) { skipped = append(skipped, msg) }

	// 1MB cap: big.sql must be dropped, others kept.
	got, err := FilterFilesBySize(context.Background(), dir, head, all, 1024*1024, verbose)
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 kept files, got %d (%v)", len(got), got)
	}
	for _, f := range got {
		if f == "big.sql" {
			t.Errorf("big.sql should have been filtered out")
		}
	}
	if len(skipped) != 1 || !strings.Contains(skipped[0], "big.sql") {
		t.Errorf("expected one verbose skip notice for big.sql, got %v", skipped)
	}

	// Disabled filter (<=0): everything passes through unchanged.
	got, err = FilterFilesBySize(context.Background(), dir, head, all, 0, nil)
	if err != nil {
		t.Fatalf("filter (disabled): %v", err)
	}
	if len(got) != len(all) {
		t.Errorf("disabled filter dropped files: got %d want %d", len(got), len(all))
	}

	// Empty commitHash falls back to HEAD and behaves identically.
	got, err = FilterFilesBySize(context.Background(), dir, "", all, 1024*1024, nil)
	if err != nil {
		t.Fatalf("filter (HEAD fallback): %v", err)
	}
	if len(got) != 2 {
		t.Errorf("HEAD fallback gave %d kept files, want 2", len(got))
	}
}
