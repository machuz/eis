package cli

import (
	"strings"
	"testing"

	"github.com/machuz/eis/v2/internal/metric"
	"github.com/machuz/eis/v2/internal/output"
)

func gvModule(r output.GraveyardReport, name string) (output.GraveyardModule, bool) {
	for _, m := range r.Modules {
		if m.Module == name {
			return m, true
		}
	}
	return output.GraveyardModule{}, false
}

func TestBuildGraveyard_RepeatedDeathGateAndIntensity(t *testing.T) {
	graves := []metric.FileGrave{
		// hard.go: 3 repeated deaths -> hotspot. 30 others-dead of 100 adds.
		{File: "svc/x/hard.go", Module: "svc/x", Deaths: 3, OthersDeadLines: 30, TotalAdds: 100, Contributors: 4},
		// util.go in same module: 1 death -> healthy one-off, NOT a hotspot, but
		// its adds still count toward the module denominator.
		{File: "svc/x/util.go", Module: "svc/x", Deaths: 1, OthersDeadLines: 5, TotalAdds: 100, Contributors: 2},
		// calm module: only self-ish single deaths -> no hotspot -> not surfaced.
		{File: "svc/calm/a.go", Module: "svc/calm", Deaths: 1, OthersDeadLines: 3, TotalAdds: 500, Contributors: 2},
	}
	currentFiles := map[string]bool{"svc/x/hard.go": true, "svc/x/util.go": true, "svc/calm/a.go": true}

	r := buildGraveyard(graves, currentFiles, 5, false, false)

	// Only svc/x surfaces (svc/calm has no repeated-death hotspot).
	if len(r.Modules) != 1 {
		t.Fatalf("modules = %d, want 1 (only svc/x)", len(r.Modules))
	}
	m := r.Modules[0]
	if m.Module != "svc/x" {
		t.Fatalf("module = %s, want svc/x", m.Module)
	}
	if len(m.Hotspots) != 1 || m.Hotspots[0].File != "svc/x/hard.go" {
		t.Errorf("hotspots wrong: %+v", m.Hotspots)
	}
	if m.DeathEvents != 3 {
		t.Errorf("death_events = %d, want 3", m.DeathEvents)
	}
	// intensity = hotspot others-dead (30) / module total adds (100+100=200) = 0.15
	if m.DeathIntensity != 0.15 {
		t.Errorf("death_intensity = %v, want 0.15", m.DeathIntensity)
	}
	if m.Hotspots[0].ContestedByN != 4 {
		t.Errorf("contested_by_n = %d, want 4", m.Hotspots[0].ContestedByN)
	}
}

// A high-churn hard spot should out-score a calm module: this pins the
// "difficult vs healthy" separation the metric exists for.
func TestBuildGraveyard_HardSpotOutscoresCalm(t *testing.T) {
	graves := []metric.FileGrave{
		{File: "hard/x.go", Module: "hard", Deaths: 5, OthersDeadLines: 80, TotalAdds: 100, Contributors: 5},
		{File: "calm/y.go", Module: "calm", Deaths: 2, OthersDeadLines: 4, TotalAdds: 400, Contributors: 2},
	}
	r := buildGraveyard(graves, nil, 5, false, false) // nil currentFiles = keep all
	hard, ok := gvModule(r, "hard")
	if !ok {
		t.Fatal("hard module missing")
	}
	calm, ok := gvModule(r, "calm")
	if !ok {
		t.Fatal("calm module missing")
	}
	if !(hard.DeathIntensity > calm.DeathIntensity) {
		t.Errorf("hard (%.2f) should out-intensity calm (%.2f)", hard.DeathIntensity, calm.DeathIntensity)
	}
	if hard.DeathIntensity != 0.8 || calm.DeathIntensity != 0.01 {
		t.Errorf("intensities: hard=%.2f (want 0.80) calm=%.2f (want 0.01)", hard.DeathIntensity, calm.DeathIntensity)
	}
}

// Whole-file deletion (file not at HEAD) must be excluded.
func TestBuildGraveyard_ExcludesDeletedFiles(t *testing.T) {
	graves := []metric.FileGrave{
		{File: "svc/x/gone.go", Module: "svc/x", Deaths: 9, OthersDeadLines: 90, TotalAdds: 100, Contributors: 4},
	}
	currentFiles := map[string]bool{} // gone.go not present at HEAD
	r := buildGraveyard(graves, currentFiles, 5, false, false)
	if len(r.Modules) != 0 {
		t.Errorf("deleted file must be excluded, got %+v", r.Modules)
	}
}

// Non-source modules and test files are dropped when default excludes are on
// (graveyard shares anchors' code universe), and kept when off.
func TestBuildGraveyard_ExcludesNonSourceAndTests(t *testing.T) {
	graves := []metric.FileGrave{
		{File: "src/core.go", Module: "src", Deaths: 3, OthersDeadLines: 30, TotalAdds: 100, Contributors: 4},            // core -> keep
		{File: "src/core_test.go", Module: "src", Deaths: 4, OthersDeadLines: 40, TotalAdds: 50, Contributors: 4},        // test file -> drop
		{File: "docs/api/Store.md", Module: "docs/api", Deaths: 9, OthersDeadLines: 90, TotalAdds: 100, Contributors: 8}, // non-source -> drop
		{File: "examples/counter/app.js", Module: "examples/counter", Deaths: 5, OthersDeadLines: 50, TotalAdds: 80, Contributors: 6},
	}
	cur := map[string]bool{"src/core.go": true, "src/core_test.go": true, "docs/api/Store.md": true, "examples/counter/app.js": true}

	on := buildGraveyard(graves, cur, 5, true, true)
	if len(on.Modules) != 1 || on.Modules[0].Module != "src" {
		t.Fatalf("excludes on: want only src, got %+v", on.Modules)
	}
	// src's only hotspot is core.go (core_test.go dropped).
	if len(on.Modules[0].Hotspots) != 1 || on.Modules[0].Hotspots[0].File != "src/core.go" {
		t.Errorf("excludes on: test file not dropped: %+v", on.Modules[0].Hotspots)
	}

	off := buildGraveyard(graves, cur, 5, false, false)
	if len(off.Modules) != 3 { // src, docs/api, examples/counter
		t.Errorf("excludes off: want 3 modules, got %d", len(off.Modules))
	}
}

// Firewall: the serialized report carries no author identity.
func TestBuildGraveyard_NoOwnerNames(t *testing.T) {
	graves := []metric.FileGrave{
		{File: "svc/x/hard.go", Module: "svc/x", Deaths: 3, OthersDeadLines: 30, TotalAdds: 100, Contributors: 4},
	}
	r := buildGraveyard(graves, nil, 5, false, false)
	var buf strings.Builder
	if err := output.EncodeGraveyardJSON(&buf, r); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, tok := range []string{"author", "owner", "email", "\"Alice\"", "deleter", "top_author"} {
		if strings.Contains(s, tok) {
			t.Errorf("graveyard leaked identity token %q:\n%s", tok, s)
		}
	}
	if !strings.Contains(s, `"contested_by_n": 4`) || !strings.Contains(s, `"death_intensity"`) {
		t.Errorf("expected anonymous count + intensity:\n%s", s)
	}
}
