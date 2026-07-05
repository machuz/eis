package metric

import (
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

func TestBlameShares(t *testing.T) {
	// solo → one call, share 1
	got := map[string]float64{}
	blameShares(git.BlameLine{Author: "alice"}, func(a string, s float64) { got[a] += s })
	if len(got) != 1 || got["alice"] != 1 {
		t.Errorf("solo = %v", got)
	}
	// co-authored → 1/N each
	got = map[string]float64{}
	blameShares(git.BlameLine{Author: "alice", Authors: []string{"alice", "bob", "carol"}}, func(a string, s float64) { got[a] += s })
	for _, a := range []string{"alice", "bob", "carol"} {
		if !coApprox(got[a], 1.0/3.0) {
			t.Errorf("%s = %v, want 1/3", a, got[a])
		}
	}
}

func TestCalcDesign_SplitsCoAuthors(t *testing.T) {
	commits := []git.Commit{
		{Author: "alice", CoAuthors: []string{"bob"}, FileStats: []git.FileStat{{Filename: "pkg/core.go", Insertions: 8}}},
	}
	got := CalcDesign(commits, []string{"pkg/*"})
	if !coApprox(got["alice"], 4) || !coApprox(got["bob"], 4) {
		t.Errorf("design split: alice=%v bob=%v, want 4/4", got["alice"], got["bob"])
	}
}

func TestCalcOwnershipFragmentation_SplitsCoAuthors(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	mr := NewModuleResolver([]string{"*"}) // first path component = module
	bl := []git.BlameLine{
		{Author: "alice", CommitterTime: now, Filename: "mod/a.go", Commit: "s1", Authors: []string{"alice", "bob"}},
		{Author: "alice", CommitterTime: now, Filename: "mod/b.go", Commit: "s1", Authors: []string{"alice", "bob"}},
	}
	res := CalcOwnershipFragmentation(bl, mr)
	if len(res) != 1 {
		t.Fatalf("modules=%d, want 1", len(res))
	}
	// alice + bob each own half → not a sole owner; topShare 0.5, 2 authors.
	if !coApprox(res[0].TopShare, 0.5) {
		t.Errorf("topShare=%v, want 0.5 (co-owned, not sole)", res[0].TopShare)
	}
	if res[0].AuthorCount != 2 {
		t.Errorf("authorCount=%d, want 2", res[0].AuthorCount)
	}
}

func TestCalcCatalysis_SplitsCoAuthorFoundation(t *testing.T) {
	t0 := time.Unix(1_600_000_000, 0)
	t1 := t0.Add(24 * time.Hour)
	now := t1.Add(24 * time.Hour)
	commits := []git.Commit{
		{Author: "alice", CoAuthors: []string{"bob"}, Date: t0, FileStats: []git.FileStat{{Filename: "f.go", Insertions: 1}}},
		{Author: "carol", Date: t1, FileStats: []git.FileStat{{Filename: "f.go", Insertions: 1}}},
	}
	bl := []git.BlameLine{
		{Author: "alice", CommitterTime: t0, Filename: "f.go", Commit: "s1", Authors: []string{"alice", "bob"}},
		{Author: "carol", CommitterTime: t1, Filename: "f.go", Commit: "s2"},
	}
	got := CalcCatalysis(commits, bl, 1e9, now) // huge tau → factor ~1
	// alice+bob are the co-authored foundation; each should get catalysis for
	// carol's later surviving mass, split — and equally to each other.
	if got["alice"] <= 0 || !coApprox(got["alice"], got["bob"]) {
		t.Errorf("catalysis foundation split: alice=%v bob=%v (want equal, >0)", got["alice"], got["bob"])
	}
	if got["carol"] != 0 {
		t.Errorf("carol arrived last, must not be a foundation: %v", got["carol"])
	}
}
