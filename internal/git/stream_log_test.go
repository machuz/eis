package git

import (
	"context"
	"fmt"
	"testing"
)

// TestStreamLogParallel_MatchesSerial verifies the streaming walk delivers the
// SAME commits, IN THE SAME ORDER, with the same per-file filtered counts, as
// serial ParseLog. Order matters: the analyzer folds Production (a non-associative
// float sum) in delivery order, so a reorder would change the last ULP. The
// min-commit threshold is lowered so a tiny fixture exercises the real chunked
// fan-out instead of the serial fallback.
func TestStreamLogParallel_MatchesSerial(t *testing.T) {
	dir := newTempRepo(t)
	for i := 0; i < 40; i++ {
		writeFile(t, dir, "a.go", fmt.Sprintf("package a\n// note %d\nfunc F%d() int { return %d }\n", i, i, i))
		writeFile(t, dir, "sub/b.go", fmt.Sprintf("package sub\nvar N%d = %d\n", i, i))
		writeFile(t, dir, "README.md", fmt.Sprintf("# Title\n\nrev %d\n\n", i))
		commit(t, dir, fmt.Sprintf("c%d", i))
	}

	serial, err := ParseLog(context.Background(), dir, true)
	if err != nil {
		t.Fatalf("ParseLog: %v", err)
	}

	old := parallelLogMinCommits
	parallelLogMinCommits = 1 // force the parallel streaming path
	defer func() { parallelLogMinCommits = old }()

	var streamed []Commit
	if err := StreamLogParallel(context.Background(), dir, 4, true, func(c *Commit) {
		// Copy: the pointer is only valid during the callback.
		cc := *c
		cc.FileStats = append([]FileStat(nil), c.FileStats...)
		streamed = append(streamed, cc)
	}); err != nil {
		t.Fatalf("StreamLogParallel: %v", err)
	}

	if len(serial) != len(streamed) {
		t.Fatalf("commit count: serial=%d streamed=%d", len(serial), len(streamed))
	}
	for i := range serial {
		s, st := serial[i], streamed[i]
		if s.Hash != st.Hash {
			t.Fatalf("order mismatch at %d: serial=%s streamed=%s", i, s.Hash, st.Hash)
		}
		if len(s.FileStats) != len(st.FileStats) {
			t.Fatalf("%s filestat count: serial=%d streamed=%d", s.Hash, len(s.FileStats), len(st.FileStats))
		}
		for _, sf := range s.FileStats {
			stf, found := findFileStat(st, sf.Filename)
			if !found || stf != sf {
				t.Fatalf("%s file %s mismatch: serial=%+v streamed=%+v", s.Hash, sf.Filename, sf, stf)
			}
		}
	}
}

// TestStreamIdentityMap_MatchesBuild verifies the format-only identity stream
// yields the identical map to BuildIdentityMap over ParseLog — the analyzer uses
// it to canonicalize authors, so any divergence would shift attribution.
func TestStreamIdentityMap_MatchesBuild(t *testing.T) {
	dir := newTempRepo(t)
	// Two display names sharing one email (should collapse to the top name),
	// plus a distinct author, across several commits.
	authors := []struct{ name, email string }{
		{"Alice A", "alice@x.com"},
		{"alice", "alice@x.com"},
		{"Alice A", "alice@x.com"},
		{"Bob B", "bob@x.com"},
	}
	for i := 0; i < 8; i++ {
		a := authors[i%len(authors)]
		writeFile(t, dir, "f.go", fmt.Sprintf("package f\nvar V%d = %d\n", i, i))
		runIn(t, dir, "git", "add", "-A")
		runIn(t, dir, "git", "-c", "user.name="+a.name, "-c", "user.email="+a.email,
			"commit", "-q", "-m", fmt.Sprintf("c%d", i))
	}

	commits, err := ParseLog(context.Background(), dir, true)
	if err != nil {
		t.Fatalf("ParseLog: %v", err)
	}
	built := BuildIdentityMap(commits)
	streamed, err := StreamIdentityMap(context.Background(), dir)
	if err != nil {
		t.Fatalf("StreamIdentityMap: %v", err)
	}
	if len(built) != len(streamed) {
		t.Fatalf("size: built=%d streamed=%d (%v vs %v)", len(built), len(streamed), built, streamed)
	}
	for k, v := range built {
		if streamed[k] != v {
			t.Fatalf("built[%q]=%q streamed=%q", k, v, streamed[k])
		}
	}
}
