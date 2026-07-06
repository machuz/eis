package metric

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/machuz/eis/v2/internal/git"
)

func gitDebt(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// A fix commit's debt-cleanup credit is split across the fixer's human co-authors,
// while the cleaned debt is attributed to the original author.
func TestCalcDebt_SplitsFixCoAuthors(t *testing.T) {
	dir := t.TempDir()
	gitDebt(t, dir, "init", "-q", "-b", "main")
	gitDebt(t, dir, "config", "user.email", "t@x.com")
	gitDebt(t, dir, "config", "user.name", "T")
	gitDebt(t, dir, "config", "commit.gpgsign", "false")

	write := func(s string) {
		if err := os.WriteFile(filepath.Join(dir, "f.go"), []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Alice writes debt-laden code.
	write("package p\nvar aVar = 1\nvar bVar = 2\nvar cVar = 3\n")
	gitDebt(t, dir, "add", "-A")
	gitDebt(t, dir, "commit", "-q", "-m", "add", "--author", "Alice <alice@x.com>")
	// Bob, co-authored by Carol (human), fixes it (removes Alice's lines).
	write("package p\n")
	gitDebt(t, dir, "add", "-A")
	gitDebt(t, dir, "commit", "-q", "-m", "fix cleanup", "-m", "Co-authored-by: Carol <carol@x.com>", "--author", "Bob <bob@x.com>")

	commits, err := git.ParseLog(context.Background(), dir, false)
	if err != nil {
		t.Fatal(err)
	}
	fixCommits := GetFixCommits(commits)
	if len(fixCommits) == 0 {
		t.Fatal("no fix commits detected")
	}
	coMap := CoAuthorMap(commits)
	_, data := CalcDebt(context.Background(), dir, fixCommits, 50, 0, 120, 4, nil, coMap, nil, nil)

	if data.Cleaned["Bob"] <= 0 || data.Cleaned["Carol"] <= 0 {
		t.Errorf("cleanup credit not split to co-author: Bob=%v Carol=%v", data.Cleaned["Bob"], data.Cleaned["Carol"])
	}
	if !coApprox(data.Cleaned["Bob"], data.Cleaned["Carol"]) {
		t.Errorf("Bob/Carol cleanup should be equal: %v vs %v", data.Cleaned["Bob"], data.Cleaned["Carol"])
	}
	if data.Generated["Alice"] <= 0 {
		t.Errorf("Alice should be credited with generating the cleaned debt: %v", data.Generated["Alice"])
	}
}
