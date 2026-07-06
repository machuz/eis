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
	_, data := CalcDebt(context.Background(), dir, fixCommits, 50, 0, 120, 4, nil, coMap, nil, nil, nil)

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

// Debt honors the file-exclusion patterns: a fix that only touches an excluded
// (generated / lockfile) path contributes nothing — consistent with Survival and
// Production, and avoiding the pathologically slow -M blame of churny lockfiles.
func TestCalcDebt_ExcludesFiles(t *testing.T) {
	dir := t.TempDir()
	gitDebt(t, dir, "init", "-q", "-b", "main")
	gitDebt(t, dir, "config", "user.email", "t@x.com")
	gitDebt(t, dir, "config", "user.name", "T")
	gitDebt(t, dir, "config", "commit.gpgsign", "false")

	writeF := func(name, s string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Alice adds lines to a generated lockfile; Bob removes them (a "fix").
	writeF("package-lock.json", "{\n\"a\":1,\n\"b\":2,\n\"c\":3\n}\n")
	gitDebt(t, dir, "add", "-A")
	gitDebt(t, dir, "commit", "-q", "-m", "add lock", "--author", "Alice <alice@x.com>")
	writeF("package-lock.json", "{\n}\n")
	gitDebt(t, dir, "add", "-A")
	gitDebt(t, dir, "commit", "-q", "-m", "fix lock", "--author", "Bob <bob@x.com>")

	commits, err := git.ParseLog(context.Background(), dir, false)
	if err != nil {
		t.Fatal(err)
	}
	fixCommits := GetFixCommits(commits)
	coMap := CoAuthorMap(commits)

	// Without excludes, the lockfile churn IS counted.
	_, incl := CalcDebt(context.Background(), dir, fixCommits, 50, 0, 120, 4, nil, coMap, nil, nil, nil)
	if incl.Generated["Alice"] <= 0 {
		t.Fatalf("precondition: lockfile debt should count without excludes, got %v", incl.Generated["Alice"])
	}
	// With the lockfile excluded, it contributes nothing.
	_, excl := CalcDebt(context.Background(), dir, fixCommits, 50, 0, 120, 4, nil, coMap, []string{"package-lock.json"}, nil, nil)
	if len(excl.Generated) != 0 || len(excl.Cleaned) != 0 {
		t.Errorf("excluded lockfile should contribute no debt, got generated=%v cleaned=%v", excl.Generated, excl.Cleaned)
	}
}
