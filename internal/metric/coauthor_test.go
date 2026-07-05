package metric

import (
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

func TestCommitAuthors(t *testing.T) {
	if a := CommitAuthors(git.Commit{Author: "alice"}); len(a) != 1 || a[0] != "alice" {
		t.Errorf("solo = %v", a)
	}
	// dedup the primary if it also appears as a co-author; Author-first order
	a := CommitAuthors(git.Commit{Author: "alice", CoAuthors: []string{"alice", "bob"}})
	if len(a) != 2 || a[0] != "alice" || a[1] != "bob" {
		t.Errorf("dedup = %v, want [alice bob]", a)
	}
}

func TestCalcProduction_SplitsCoAuthors(t *testing.T) {
	commits := []git.Commit{
		{Author: "alice", CoAuthors: []string{"bob"}, FileStats: []git.FileStat{{Filename: "a.go", Insertions: 10}}},
		{Author: "alice", FileStats: []git.FileStat{{Filename: "b.go", Insertions: 4}}},
	}
	got := CalcProduction(commits, nil)
	if got["alice"] != 9 { // 5 (split) + 4 (solo)
		t.Errorf("alice = %v, want 9", got["alice"])
	}
	if got["bob"] != 5 {
		t.Errorf("bob = %v, want 5", got["bob"])
	}
}

// A co-authored line's survival mass is split equally across the contributor set;
// a solo line credits its single author fully. Mass is conserved.
func TestCalcSurvival_SplitsCoAuthors(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	bl := []git.BlameLine{
		{Author: "alice", CommitterTime: now, Filename: "a.go", Commit: "sha1", Authors: []string{"alice", "bob"}},
		{Author: "carol", CommitterTime: now, Filename: "b.go"}, // solo (Authors empty)
	}
	r := CalcSurvival(bl, 180, now) // daysAlive 0 → weight 1
	if !coApprox(r.Decayed["alice"], 0.5) || !coApprox(r.Decayed["bob"], 0.5) {
		t.Errorf("decayed split: alice=%v bob=%v, want 0.5/0.5", r.Decayed["alice"], r.Decayed["bob"])
	}
	if !coApprox(r.Decayed["carol"], 1.0) {
		t.Errorf("carol decayed=%v, want 1.0", r.Decayed["carol"])
	}
	if !coApprox(r.Raw["alice"], 0.5) || !coApprox(r.Raw["carol"], 1.0) {
		t.Errorf("raw split: %v", r.Raw)
	}
	// total decayed mass conserved: 0.5+0.5+1.0 = 2 lines.
	total := r.Decayed["alice"] + r.Decayed["bob"] + r.Decayed["carol"]
	if !coApprox(total, 2.0) {
		t.Errorf("total decayed=%v, want 2 (mass conserved)", total)
	}
}

func TestCoAuthorMap(t *testing.T) {
	commits := []git.Commit{
		{Hash: "h1", Author: "alice", CoAuthors: []string{"bob"}},
		{Hash: "h2", Author: "carol"}, // solo → not in map
	}
	m := CoAuthorMap(commits)
	if len(m) != 1 {
		t.Fatalf("map size=%d, want 1 (sparse)", len(m))
	}
	if got := m["h1"]; len(got) != 2 || got[0] != "alice" || got[1] != "bob" {
		t.Errorf("h1 set=%v", got)
	}
	if _, ok := m["h2"]; ok {
		t.Error("solo commit must not be in the map")
	}
}

func coApprox(a, b float64) bool { return a-b < 1e-9 && b-a < 1e-9 }
