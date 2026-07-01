package main

import "testing"

func TestSplitCSV(t *testing.T) {
	got := splitCSV(" react, , kubernetes ,vite")
	want := []string{"react", "kubernetes", "vite"}
	if len(got) != len(want) {
		t.Fatalf("len = %d (%v), want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if len(splitCSV("  ,  ")) != 0 {
		t.Errorf("blank-only CSV must yield no dirs")
	}
}

func TestManifestFilter(t *testing.T) {
	m := &repoManifest{Repos: []manifestEntry{
		{Dir: "react", FullName: "react/react", LanguageFamily: "javascript"},
		{Dir: "vite", FullName: "vitejs/vite", LanguageFamily: "typescript"},
		{Dir: "go", FullName: "golang/go", LanguageFamily: "go"},
	}}
	m.filter([]string{"vite", "go"})
	if len(m.Repos) != 2 || m.Repos[0].Dir != "vite" || m.Repos[1].Dir != "go" {
		t.Fatalf("filter kept %+v, want [vite, go]", m.Repos)
	}
}
