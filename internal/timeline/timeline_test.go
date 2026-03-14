package timeline

import (
	"testing"

	"github.com/machuz/engineering-impact-score/internal/scorer"
)

func TestBuildTimeline_BasicFlow(t *testing.T) {
	periods := []PeriodResult{
		{
			Label: "2025-Q2",
			Members: []scorer.Result{
				{Author: "alice", Total: 42.3, Production: 35, Quality: 78, Role: "Producer", Style: "Mass", State: "Active"},
				{Author: "bob", Total: 30.0, Production: 20, Quality: 60, Role: "Producer", Style: "Balanced", State: "Active"},
			},
		},
		{
			Label: "2025-Q3",
			Members: []scorer.Result{
				{Author: "alice", Total: 55.7, Production: 52, Quality: 80, Role: "Anchor", Style: "Builder", State: "Active"},
				{Author: "bob", Total: 25.0, Production: 15, Quality: 55, Role: "Producer", Style: "Balanced", State: "Growing"},
			},
		},
	}

	timelines := BuildTimeline(periods)

	if len(timelines) != 2 {
		t.Fatalf("expected 2 timelines, got %d", len(timelines))
	}

	// Should be sorted by latest total descending
	if timelines[0].Author != "alice" {
		t.Errorf("expected alice first (higher total), got %s", timelines[0].Author)
	}

	if len(timelines[0].Periods) != 2 {
		t.Fatalf("expected 2 periods for alice, got %d", len(timelines[0].Periods))
	}

	// Check alice's periods
	if timelines[0].Periods[0].Total != 42.3 {
		t.Errorf("expected alice Q2 total 42.3, got %.1f", timelines[0].Periods[0].Total)
	}
	if timelines[0].Periods[1].Total != 55.7 {
		t.Errorf("expected alice Q3 total 55.7, got %.1f", timelines[0].Periods[1].Total)
	}
}

func TestDetectTransitions(t *testing.T) {
	periods := []AuthorPeriod{
		{Label: "Q1", Role: "Producer", Style: "Mass", State: "Active"},
		{Label: "Q2", Role: "Producer", Style: "Builder", State: "Active"},
		{Label: "Q3", Role: "Anchor", Style: "Builder", State: "Active"},
	}

	transitions := DetectTransitions(periods)

	if len(transitions) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(transitions))
	}

	// Style transition at Q2
	found := false
	for _, tr := range transitions {
		if tr.Axis == "Style" && tr.From == "Mass" && tr.To == "Builder" && tr.AtPeriod == "Q2" {
			found = true
		}
	}
	if !found {
		t.Error("expected Style transition Mass→Builder at Q2")
	}

	// Role transition at Q3
	found = false
	for _, tr := range transitions {
		if tr.Axis == "Role" && tr.From == "Producer" && tr.To == "Anchor" && tr.AtPeriod == "Q3" {
			found = true
		}
	}
	if !found {
		t.Error("expected Role transition Producer→Anchor at Q3")
	}
}

func TestDetectTransitions_SkipsDash(t *testing.T) {
	periods := []AuthorPeriod{
		{Label: "Q1", Role: "—", Style: "Mass", State: "Active"},
		{Label: "Q2", Role: "Producer", Style: "Mass", State: "Active"},
	}

	transitions := DetectTransitions(periods)

	for _, tr := range transitions {
		if tr.Axis == "Role" {
			t.Error("should not detect transition from — to Producer")
		}
	}
}

func TestDetectTransitions_Empty(t *testing.T) {
	transitions := DetectTransitions(nil)
	if len(transitions) != 0 {
		t.Error("expected 0 transitions for nil input")
	}

	transitions = DetectTransitions([]AuthorPeriod{{Label: "Q1", Role: "Producer"}})
	if len(transitions) != 0 {
		t.Error("expected 0 transitions for single period")
	}
}

func TestBuildTimeline_AuthorMissingFromPeriod(t *testing.T) {
	periods := []PeriodResult{
		{
			Label:   "Q1",
			Members: []scorer.Result{{Author: "alice", Total: 40}},
		},
		{
			Label:   "Q2",
			Members: []scorer.Result{{Author: "alice", Total: 50}, {Author: "bob", Total: 30}},
		},
	}

	timelines := BuildTimeline(periods)

	// bob should have a zero-value Q1 period
	var bob *AuthorTimeline
	for i := range timelines {
		if timelines[i].Author == "bob" {
			bob = &timelines[i]
		}
	}
	if bob == nil {
		t.Fatal("expected bob in timelines")
	}
	if bob.Periods[0].Total != 0 {
		t.Errorf("expected bob Q1 total 0, got %.1f", bob.Periods[0].Total)
	}
	if bob.Periods[1].Total != 30 {
		t.Errorf("expected bob Q2 total 30, got %.1f", bob.Periods[1].Total)
	}
}
