package metric

import "testing"

func TestEscapeGlob(t *testing.T) {
	// A plain path is unchanged and matches only itself.
	if got := EscapeGlob("a/b/c.pb.go"); got != "a/b/c.pb.go" {
		t.Fatalf("plain path changed: %q", got)
	}
	if !IsExcluded("a/b/c.pb.go", []string{EscapeGlob("a/b/c.pb.go")}) {
		t.Fatal("escaped plain path should match itself")
	}
	// A path with glob metacharacters must become literal-matching: it matches its
	// own path but NOT a different path the raw glob would have captured.
	raw := "weird/[gen]*.go"
	esc := EscapeGlob(raw)
	if !IsExcluded(raw, []string{esc}) {
		t.Errorf("escaped %q should match its own path", esc)
	}
	if IsExcluded("weird/g.go", []string{esc}) {
		t.Error("escaped pattern over-matched a different path (weird/g.go)")
	}
}
