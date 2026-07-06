package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// buildBlameStabilityRepo builds a repo where f.go is created in c1 and never
// touched again, while other.go changes in c2 and c3. It returns the repo path
// and the three commit SHAs in order.
func buildBlameStabilityRepo(t *testing.T) (dir, c1, c2, c3 string) {
	t.Helper()
	dir = t.TempDir()

	run := func(env []string, args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(), env...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	write := func(file, content string) {
		if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	env := func(date, name string) []string {
		return []string{
			"GIT_AUTHOR_DATE=" + date + "T10:00:00+00:00",
			"GIT_COMMITTER_DATE=" + date + "T10:00:00+00:00",
			"GIT_AUTHOR_NAME=" + name, "GIT_AUTHOR_EMAIL=" + name + "@test",
			"GIT_COMMITTER_NAME=" + name, "GIT_COMMITTER_EMAIL=" + name + "@test",
		}
	}
	head := func() string {
		cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("rev-parse: %v\n%s", err, out)
		}
		return string(out[:40])
	}

	run(nil, "init", "-q", "-b", "main")
	run(nil, "config", "user.email", "x@test")
	run(nil, "config", "user.name", "x")

	write("f.go", "package p\n\nfunc F() int {\n\treturn 1\n}\n")
	write("other.go", "package p\n\nfunc Other() int { return 1 }\n")
	run(nil, "add", "-A")
	run(env("2024-01-15", "alice"), "commit", "-q", "-m", "c1")
	c1 = head()

	write("other.go", "package p\n\nfunc Other() int { return 2 }\n")
	run(nil, "add", "-A")
	run(env("2024-02-15", "bob"), "commit", "-q", "-m", "c2")
	c2 = head()

	write("other.go", "package p\n\nfunc Other() int { return 3 }\nfunc Extra() {}\n")
	run(nil, "add", "-A")
	run(env("2024-03-15", "carol"), "commit", "-q", "-m", "c3")
	c3 = head()

	return dir, c1, c2, c3
}

// TestBlameFileAtCommit_UnchangedFileIsStable is the git-level premise the
// per-file incremental blame cache stands on: a file that is byte-identical
// between two commits has identical blame at both. f.go is untouched after c1,
// so its blame at c1, c2 and c3 must be equal — which is exactly why keying the
// cache on a file's last-touch SHA (not the moving window boundary) is sound.
func TestBlameFileAtCommit_UnchangedFileIsStable(t *testing.T) {
	dir, c1, c2, c3 := buildBlameStabilityRepo(t)
	ctx := context.Background()

	at := func(commit string) []BlameLine {
		lines, err := BlameFileAtCommit(ctx, dir, commit, "f.go")
		if err != nil {
			t.Fatalf("blame f.go @ %s: %v", commit, err)
		}
		if len(lines) == 0 {
			t.Fatalf("blame f.go @ %s: 0 lines", commit)
		}
		return lines
	}

	base := at(c1)
	if !reflect.DeepEqual(base, at(c2)) {
		t.Error("blame of unchanged f.go differs between c1 and c2")
	}
	if !reflect.DeepEqual(base, at(c3)) {
		t.Error("blame of unchanged f.go differs between c1 and c3")
	}
}

// TestConcurrentBlameFilesAtCommitByFile_MatchesFlat proves the grouped-by-input
// -path variant returns the same lines as the flat wrapper — so routing the
// incremental cache through ByFile changes nothing about WHAT is blamed, only
// how it is grouped for caching. Compared as multisets because both flatten in
// input order but the flat wrapper is defined in terms of ByFile anyway; the
// multiset guards against any future divergence (e.g. dropped/duplicated files).
func TestConcurrentBlameFilesAtCommitByFile_MatchesFlat(t *testing.T) {
	dir, _, _, c3 := buildBlameStabilityRepo(t)
	ctx := context.Background()
	files := []string{"f.go", "other.go"}

	flat, err := ConcurrentBlameFilesAtCommit(ctx, dir, c3, files, 1<<30, 2, 0, nil, nil)
	if err != nil {
		t.Fatalf("flat blame: %v", err)
	}
	byFile := ConcurrentBlameFilesAtCommitByFile(ctx, dir, c3, files, 2, 0, nil, nil)

	var grouped []BlameLine
	for _, f := range files {
		grouped = append(grouped, byFile[f]...)
	}

	if len(flat) == 0 || len(grouped) == 0 {
		t.Fatalf("empty blame: flat=%d grouped=%d", len(flat), len(grouped))
	}
	if !sameLineMultiset(flat, grouped) {
		t.Errorf("ByFile grouped blame differs from flat blame: flat=%d lines, grouped=%d lines", len(flat), len(grouped))
	}
}

// sameLineMultiset compares two blame slices ignoring order.
func sameLineMultiset(a, b []BlameLine) bool {
	if len(a) != len(b) {
		return false
	}
	key := func(l BlameLine) string {
		return l.Filename + "\x00" + l.Commit + "\x00" + l.Author
	}
	ka := make([]string, len(a))
	kb := make([]string, len(b))
	for i := range a {
		ka[i] = key(a[i])
	}
	for i := range b {
		kb[i] = key(b[i])
	}
	sort.Strings(ka)
	sort.Strings(kb)
	return reflect.DeepEqual(ka, kb)
}
