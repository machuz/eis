package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// newTempRepo initialises an empty git repo in a t.TempDir() and returns its path.
// A deterministic user.email/name is configured so commit authors are predictable.
func newTempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runIn(t, dir, "git", "init", "-q", "-b", "main")
	runIn(t, dir, "git", "config", "user.email", "test@example.com")
	runIn(t, dir, "git", "config", "user.name", "Tester")
	runIn(t, dir, "git", "config", "commit.gpgsign", "false")
	return dir
}

func runIn(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func commit(t *testing.T, dir, msg string) {
	t.Helper()
	runIn(t, dir, "git", "add", "-A")
	runIn(t, dir, "git", "commit", "-q", "-m", msg)
}

func findFileStat(c Commit, name string) (FileStat, bool) {
	for _, fs := range c.FileStats {
		if fs.Filename == name {
			return fs, true
		}
	}
	return FileStat{}, false
}

// A commit that adds only comments to a Go file should count zero code lines.
func TestParseLog_GoCommentsExcluded(t *testing.T) {
	dir := newTempRepo(t)

	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	commit(t, dir, "init")

	writeFile(t, dir, "main.go", `package main

// This is a new comment.
// Another comment line.
// Third comment line.
func main() {}
`)
	commit(t, dir, "add comments")

	commits, err := ParseLog(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) < 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}
	// Newest first.
	latest := commits[0]
	if latest.Subject != "add comments" {
		t.Fatalf("expected latest subject 'add comments', got %q", latest.Subject)
	}
	fs, ok := findFileStat(latest, "main.go")
	if !ok {
		t.Fatal("main.go not in FileStats")
	}
	// Added 3 comment lines + 1 blank line; all should be filtered.
	if fs.Insertions != 0 || fs.Deletions != 0 {
		t.Errorf("expected 0/0 for pure-comment commit, got +%d/-%d", fs.Insertions, fs.Deletions)
	}
}

// A commit to a markdown file should count every line, including blank ones,
// because prose is preserved verbatim for research/paper use.
func TestParseLog_MarkdownUnfiltered(t *testing.T) {
	dir := newTempRepo(t)
	writeFile(t, dir, "README.md", "# Title\n")
	commit(t, dir, "init")

	writeFile(t, dir, "README.md", "# Title\n\n## Section\n\nSome prose.\n")
	commit(t, dir, "extend readme")

	commits, err := ParseLog(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	latest := commits[0]
	fs, ok := findFileStat(latest, "README.md")
	if !ok {
		t.Fatal("README.md not in FileStats")
	}
	if fs.Insertions == 0 {
		t.Errorf("expected >0 insertions for markdown commit, got %d", fs.Insertions)
	}
}

// Mixed commit: some code, some comments. Only real code counts.
func TestParseLog_MixedCommit(t *testing.T) {
	dir := newTempRepo(t)
	writeFile(t, dir, "app.go", "package app\n")
	commit(t, dir, "init")

	writeFile(t, dir, "app.go", `package app

// Added comment 1
// Added comment 2
func NewFunc() int {
	return 42
}
`)
	commit(t, dir, "add func + comments")

	commits, err := ParseLog(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	latest := commits[0]
	fs, ok := findFileStat(latest, "app.go")
	if !ok {
		t.Fatal("app.go not in FileStats")
	}
	// Code lines added: func NewFunc() int { ... return 42 ... } = 3 code lines.
	// Comments (2) and blank lines should be filtered out.
	if fs.Insertions != 3 {
		t.Errorf("expected 3 code-line insertions, got %d", fs.Insertions)
	}
}

// Block comment (/* ... */) across multiple lines must not inflate counts.
func TestParseLog_GoBlockComment(t *testing.T) {
	dir := newTempRepo(t)
	writeFile(t, dir, "pkg.go", "package pkg\n")
	commit(t, dir, "init")

	writeFile(t, dir, "pkg.go", `package pkg

/*
 * Multi-line block comment.
 * Second line.
 * Third line.
 */
func Foo() {}
`)
	commit(t, dir, "add block comment and func")

	commits, err := ParseLog(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	latest := commits[0]
	fs, _ := findFileStat(latest, "pkg.go")
	// Only "func Foo() {}" counts — 1 code line.
	if fs.Insertions != 1 {
		t.Errorf("expected 1 code insertion (block comment excluded), got %d", fs.Insertions)
	}
}

// Renaming a Go file and adding only comments must still produce zero
// filtered line counts. Prior to rename-syntax resolution the numstat filename
// was `dir/{old.go => new.go}` which never matched the diff's `dir/new.go`,
// leaving raw counts (gaming hole).
func TestParseLog_RenameWithCommentOnlyChange(t *testing.T) {
	dir := newTempRepo(t)
	_ = os.MkdirAll(filepath.Join(dir, "pkg"), 0o755)
	writeFile(t, dir, "pkg/old.go", `package pkg

func A() {}
func B() {}
func C() {}
func D() {}
func E() {}
func F() {}
func G() {}
func H() {}
`)
	commit(t, dir, "init")

	// git mv preserves enough similarity that git detects a rename.
	runIn(t, dir, "git", "mv", "pkg/old.go", "pkg/new.go")
	writeFile(t, dir, "pkg/new.go", `package pkg

// New comment added during rename.
func A() {}
func B() {}
func C() {}
func D() {}
func E() {}
func F() {}
func G() {}
func H() {}
`)
	commit(t, dir, "rename + comment")

	commits, err := ParseLog(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	latest := commits[0]
	// Filename should be resolved to the new path.
	fs, ok := findFileStat(latest, "pkg/new.go")
	if !ok {
		names := []string{}
		for _, f := range latest.FileStats {
			names = append(names, f.Filename)
		}
		t.Fatalf("expected FileStat for pkg/new.go, got %v", names)
	}
	// Only a comment was added during the rename — filtered counts must be zero.
	if fs.Insertions != 0 || fs.Deletions != 0 {
		t.Errorf("rename+comment commit: expected 0/0 filtered counts, got +%d/-%d", fs.Insertions, fs.Deletions)
	}
}

func TestResolveRenamePath(t *testing.T) {
	cases := []struct{ in, want string }{
		{"regular/file.go", "regular/file.go"},
		{"dir/{old.go => new.go}", "dir/new.go"},
		{"{old_dir => new_dir}/file.go", "new_dir/file.go"},
		{"old/path.go => new/path.go", "new/path.go"},
		{"src/{subdir => }/file.go", "src/file.go"},
		{"src/{ => subdir}/file.go", "src/subdir/file.go"},
	}
	for _, c := range cases {
		if got := resolveRenamePath(c.in); got != c.want {
			t.Errorf("resolveRenamePath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// Python docstring handling.
func TestParseLog_PythonDocstring(t *testing.T) {
	dir := newTempRepo(t)
	writeFile(t, dir, "foo.py", "x = 1\n")
	commit(t, dir, "init")

	writeFile(t, dir, "foo.py", `x = 1

def greet():
    """A docstring.

    With multiple lines.
    """
    return "hi"
`)
	commit(t, dir, "add greet")

	commits, err := ParseLog(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	latest := commits[0]
	fs, _ := findFileStat(latest, "foo.py")
	// Code lines: def greet():, return "hi"  = 2.
	if fs.Insertions != 2 {
		t.Errorf("expected 2 python code insertions (docstring excluded), got %d", fs.Insertions)
	}
}
