package git

import (
	"context"
	"testing"
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
