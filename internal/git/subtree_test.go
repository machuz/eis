package git

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

// commitSubtreeSquash makes a commit carrying a git-subtree-split trailer under a
// chosen author — the shape `git subtree add/pull --squash` produces, where one
// integrator's commit contains an imported repo's whole tree.
func commitSubtreeSquash(t *testing.T, dir, subject, dirName, split, author, email string) {
	t.Helper()
	runIn(t, dir, "git", "add", "-A")
	runIn(t, dir, "git", "commit", "-q",
		"-m", subject,
		"-m", "git-subtree-dir: "+dirName+"\ngit-subtree-split: "+split,
		"--author", author+" <"+email+">")
}

// A subtree-squash commit is detected by its git-subtree-split trailer; ordinary
// commits (including ones that merely mention subtree in prose) are not.
func TestSubtreeSquashCommits_DetectsOnlyTrailer(t *testing.T) {
	dir := newTempRepo(t)

	writeFile(t, dir, "normal.go", "package p\nvar x = 1\n")
	commit(t, dir, "ordinary commit about git-subtree-split in prose")

	writeFile(t, dir, "vendored/lib.go", "package lib\nvar y = 2\n")
	commitSubtreeSquash(t, dir, "Squashed 'vendored/' changes from aaaa..bbbb",
		"vendored", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "Integrator", "int@example.com")

	set, err := SubtreeSquashCommits(context.Background(), dir, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if len(set) != 1 {
		t.Fatalf("expected exactly 1 subtree-squash commit, got %d: %v", len(set), set)
	}

	head := revParse(t, dir, "HEAD")
	if _, ok := set[head]; !ok {
		t.Errorf("HEAD (%s) should be the detected squash commit; set=%v", head, set)
	}
	firstParent := revParse(t, dir, "HEAD~1")
	if _, ok := set[firstParent]; ok {
		t.Errorf("ordinary commit %s must not be flagged as subtree-squash", firstParent)
	}
}

// End to end at the blame layer: an integrator imports code via subtree-squash
// (all lines blame to their one commit), then a second engineer edits one line.
// Suppression must drop only the untouched imported lines (which would otherwise
// inflate the integrator) and keep the edited line credited to its real author.
func TestDropSubtreeSquashBlame_KeepsLaterEdits(t *testing.T) {
	dir := newTempRepo(t)

	const imported = "line one AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\n" +
		"line two BBBBBBBBBBBBBBBBBBBBBBBBBBBBBB\n" +
		"line three CCCCCCCCCCCCCCCCCCCCCCCCCCCC\n"
	writeFile(t, dir, "svc/app.go", imported)
	commitSubtreeSquash(t, dir, "Squashed 'svc/' changes from 1111..2222",
		"svc", "2222222222222222222222222222222222222222", "Integrator", "int@example.com")

	// A different engineer rewrites the middle line only.
	const edited = "line one AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\n" +
		"line two EDITED BY REALAUTHOR DDDDDDDDDDDD\n" +
		"line three CCCCCCCCCCCCCCCCCCCCCCCCCCCC\n"
	writeFile(t, dir, "svc/app.go", edited)
	commitAs(t, dir, "fix middle line", "RealAuthor", "real@example.com")

	lines, err := BlameFile(context.Background(), dir, "svc/app.go")
	if err != nil {
		t.Fatal(err)
	}

	// Baseline: without suppression the integrator captures the 2 untouched
	// imported lines — the inflation this fix removes.
	byAuthor := map[string]int{}
	for _, bl := range lines {
		byAuthor[bl.Author]++
	}
	if byAuthor["Integrator"] != 2 {
		t.Fatalf("precondition: expected Integrator to blame-own 2 imported lines, got %d (%v)", byAuthor["Integrator"], byAuthor)
	}
	if byAuthor["RealAuthor"] != 1 {
		t.Fatalf("precondition: expected RealAuthor to own the edited line, got %d (%v)", byAuthor["RealAuthor"], byAuthor)
	}

	set, err := SubtreeSquashCommits(context.Background(), dir, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	kept, dropped := DropSubtreeSquashBlame(lines, set)
	if dropped != 2 {
		t.Errorf("expected 2 imported lines dropped, got %d", dropped)
	}
	keptByAuthor := map[string]int{}
	for _, bl := range kept {
		keptByAuthor[bl.Author]++
	}
	if keptByAuthor["Integrator"] != 0 {
		t.Errorf("integrator must keep no imported lines, got %d", keptByAuthor["Integrator"])
	}
	if keptByAuthor["RealAuthor"] != 1 {
		t.Errorf("real author's edited line must survive suppression, got %d", keptByAuthor["RealAuthor"])
	}
}

func TestDropSubtreeSquashBlame_EmptySetNoOp(t *testing.T) {
	lines := []BlameLine{{Author: "A", Commit: "x"}, {Author: "B", Commit: "y"}}
	kept, dropped := DropSubtreeSquashBlame(lines, nil)
	if dropped != 0 || len(kept) != 2 {
		t.Errorf("nil set must be a no-op, got dropped=%d kept=%d", dropped, len(kept))
	}
	kept, dropped = DropSubtreeSquashBlame(lines, map[string]struct{}{})
	if dropped != 0 || len(kept) != 2 {
		t.Errorf("empty set must be a no-op, got dropped=%d kept=%d", dropped, len(kept))
	}
}

func revParse(t *testing.T, dir, rev string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", rev)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git rev-parse %s: %v\n%s", rev, err, out)
	}
	return strings.TrimSpace(string(out))
}
