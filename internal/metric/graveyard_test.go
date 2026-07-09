package metric

import (
	"testing"

	"github.com/machuz/eis/v2/internal/git"
)

func fstats(fs ...git.FileStat) []git.FileStat { return fs }
func fstat(file string, ins, del int) git.FileStat {
	return git.FileStat{Filename: file, Insertions: ins, Deletions: del}
}

func TestCalcGraveyard_OthersContestedVsSelfRewrite(t *testing.T) {
	mr := NewModuleResolver([]string{"svc/*"})
	// chronological:
	//  1. A adds 10 to hard.go
	//  2. B deletes 8, adds 8  -> others-contested death #1 (B≠A, 8 of A's lines)
	//  3. C deletes 8, adds 8  -> others-contested death #2 (C≠B)
	//  4. C deletes 4, adds 4  -> SELF-rewrite (C=C) -> NOT a death
	commits := []git.Commit{
		{Author: "A", FileStats: fstats(fstat("svc/x/hard.go", 10, 0))},
		{Author: "B", FileStats: fstats(fstat("svc/x/hard.go", 8, 8))},
		{Author: "C", FileStats: fstats(fstat("svc/x/hard.go", 8, 8))},
		{Author: "C", FileStats: fstats(fstat("svc/x/hard.go", 4, 4))},
	}
	g := CalcGraveyard(commits, mr)["svc/x/hard.go"]
	if g == nil {
		t.Fatal("missing hard.go")
	}
	if g.Module != "svc/x" {
		t.Errorf("module = %q, want svc/x", g.Module)
	}
	if g.Deaths != 2 {
		t.Errorf("deaths = %d, want 2 (self-rewrite excluded)", g.Deaths)
	}
	if g.OthersDeadLines != 16 { // 8 + 8
		t.Errorf("others dead lines = %d, want 16", g.OthersDeadLines)
	}
	// Contributors: A (died), B (deleter+died), C (deleter) -> 3.
	if g.Contributors != 3 {
		t.Errorf("contributors = %d, want 3", g.Contributors)
	}
	if g.TotalAdds != 30 { // 10+8+8+4
		t.Errorf("total adds = %d, want 30", g.TotalAdds)
	}
}

func TestCalcGraveyard_PureSelfRewriteIsSilent(t *testing.T) {
	mr := NewModuleResolver(nil)
	// A keeps rewriting A's own code -> healthy, zero deaths.
	commits := []git.Commit{
		{Author: "A", FileStats: fstats(fstat("a/b/f.go", 20, 0))},
		{Author: "A", FileStats: fstats(fstat("a/b/f.go", 15, 15))},
		{Author: "A", FileStats: fstats(fstat("a/b/f.go", 10, 10))},
	}
	g := CalcGraveyard(commits, mr)["a/b/f.go"]
	if g.Deaths != 0 || g.OthersDeadLines != 0 {
		t.Errorf("self-rewrite should have 0 deaths, got deaths=%d dead=%d", g.Deaths, g.OthersDeadLines)
	}
}

func TestCalcGraveyard_MergesSkipped(t *testing.T) {
	mr := NewModuleResolver(nil)
	commits := []git.Commit{
		{Author: "A", FileStats: fstats(fstat("a/b/f.go", 10, 0))},
		{Author: "B", IsMerge: true, FileStats: fstats(fstat("a/b/f.go", 5, 5))}, // merge -> ignored
	}
	g := CalcGraveyard(commits, mr)["a/b/f.go"]
	if g.Deaths != 0 {
		t.Errorf("merge commit must be skipped, got deaths=%d", g.Deaths)
	}
}
