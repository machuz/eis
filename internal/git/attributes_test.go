package git

import (
	"context"
	"testing"
)

// LinguistExcluded returns exactly the paths marked linguist-generated /
// linguist-vendored (bare or =true), and never real source or .gitattributes.
func TestLinguistExcluded(t *testing.T) {
	dir := newTempRepo(t)
	writeFile(t, dir, ".gitattributes",
		"gen/*.go linguist-generated\n"+
			"vend/** linguist-vendored\n"+
			"docs/*.md linguist-generated=true\n")
	writeFile(t, dir, "real.go", "package p\nvar x = 1\n")
	writeFile(t, dir, "gen/auto.go", "package p\nvar y = 2\n")
	writeFile(t, dir, "vend/dep.go", "package p\nvar z = 3\n")
	writeFile(t, dir, "docs/readme.md", "# hi\n")
	commit(t, dir, "init")

	got, err := LinguistExcluded(context.Background(), dir)
	if err != nil {
		t.Fatalf("LinguistExcluded: %v", err)
	}
	set := map[string]bool{}
	for _, p := range got {
		set[p] = true
	}
	for _, want := range []string{"gen/auto.go", "vend/dep.go", "docs/readme.md"} {
		if !set[want] {
			t.Errorf("expected %q excluded, got %v", want, got)
		}
	}
	if set["real.go"] {
		t.Error("real.go must NOT be excluded")
	}
	if set[".gitattributes"] {
		t.Error(".gitattributes must NOT be excluded")
	}
}

// A repo with no .gitattributes yields an empty set (and no error).
func TestLinguistExcluded_NoAttributes(t *testing.T) {
	dir := newTempRepo(t)
	writeFile(t, dir, "a.go", "package p\n")
	commit(t, dir, "init")
	got, err := LinguistExcluded(context.Background(), dir)
	if err != nil {
		t.Fatalf("LinguistExcluded: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no exclusions, got %v", got)
	}
}
