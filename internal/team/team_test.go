package team

import (
	"testing"

	"github.com/machuz/eis/v2/internal/scorer"
)

func TestAggregate_Basic(t *testing.T) {
	results := []scorer.Result{
		{Author: "alice", RecentlyActive: true, Production: 80, Quality: 90, Survival: 70, Design: 60, Breadth: 50, DebtCleanup: 55, Impact: 65, Role: "Architect", Style: "Builder", State: "Active"},
		{Author: "bob", RecentlyActive: true, Production: 40, Quality: 85, Survival: 50, Design: 30, Breadth: 40, DebtCleanup: 60, Impact: 45, Role: "Anchor", Style: "Balanced", State: "Growing"},
	}

	tr := Aggregate("backend-team", "Backend", 3, results, nil)

	// Both members are core (Active + Impact >= 20)
	if tr.CoreMemberCount != 2 {
		t.Errorf("CoreMemberCount = %d, want 2", tr.CoreMemberCount)
	}
	if tr.MemberCount != 2 {
		t.Errorf("MemberCount = %d, want 2", tr.MemberCount)
	}
	if tr.AvgProduction != 60 {
		t.Errorf("AvgProduction = %f, want 60", tr.AvgProduction)
	}
	if tr.AvgQuality != 87.5 {
		t.Errorf("AvgQuality = %f, want 87.5", tr.AvgQuality)
	}
	if tr.RoleDist["Architect"] != 1 {
		t.Errorf("RoleDist[Architect] = %d, want 1", tr.RoleDist["Architect"])
	}
}

func TestAggregate_WithFilter(t *testing.T) {
	results := []scorer.Result{
		{Author: "alice", RecentlyActive: true, Production: 80, Quality: 90, Impact: 65, Role: "Architect", Style: "Builder", State: "Active"},
		{Author: "bob", RecentlyActive: true, Production: 40, Quality: 85, Impact: 45, Role: "Anchor", Style: "Balanced", State: "Growing"},
		{Author: "charlie", RecentlyActive: true, Production: 60, Quality: 70, Impact: 55, Role: "Producer", Style: "Mass", State: "Active"},
	}

	tr := Aggregate("team", "Backend", 3, results, []string{"alice", "bob"})

	if tr.CoreMemberCount != 2 {
		t.Errorf("CoreMemberCount = %d, want 2", tr.CoreMemberCount)
	}
	if tr.AvgProduction != 60 {
		t.Errorf("AvgProduction = %f, want 60", tr.AvgProduction)
	}
}

func TestAggregate_CoreAndRiskSplit(t *testing.T) {
	results := []scorer.Result{
		{Author: "alice", RecentlyActive: true, Production: 80, Quality: 90, DebtCleanup: 70, Impact: 65, Role: "Architect", Style: "Builder", State: "Active"},
		{Author: "bob", RecentlyActive: false, Production: 10, Quality: 50, DebtCleanup: 30, Impact: 20, Role: "—", Style: "—", State: "Former"},
		{Author: "charlie", RecentlyActive: true, Production: 60, Quality: 80, DebtCleanup: 60, Impact: 55, Role: "Producer", Style: "Mass", State: "Active"},
	}

	tr := Aggregate("team", "Backend", 3, results, nil)

	// alice + charlie = core, bob = risk (Former)
	if tr.CoreMemberCount != 2 {
		t.Errorf("CoreMemberCount = %d, want 2", tr.CoreMemberCount)
	}
	// effective = core + risk = 3
	if tr.MemberCount != 3 {
		t.Errorf("MemberCount = %d, want 3 (core + risk)", tr.MemberCount)
	}
	if tr.TotalMemberCount != 3 {
		t.Errorf("TotalMemberCount = %d, want 3", tr.TotalMemberCount)
	}
	// Averages from core only (alice + charlie)
	if tr.AvgProduction != 70 {
		t.Errorf("AvgProduction = %f, want 70 (core only)", tr.AvgProduction)
	}
	if tr.AvgDebtCleanup != 65 {
		t.Errorf("AvgDebtCleanup = %f, want 65 (core only)", tr.AvgDebtCleanup)
	}
	// Bob (Former) should appear in state distribution
	if tr.StateDist["Former"] != 1 {
		t.Errorf("StateDist[Former] = %d, want 1", tr.StateDist["Former"])
	}
}

func TestAggregate_PeripheralExcluded(t *testing.T) {
	results := []scorer.Result{
		{Author: "alice", RecentlyActive: true, Production: 80, Quality: 90, Impact: 65, Role: "Architect", Style: "Builder", State: "Active"},
		{Author: "helper", RecentlyActive: true, Production: 5, Quality: 60, Impact: 12, Role: "—", Style: "—", State: "Active"},
	}

	tr := Aggregate("team", "Backend", 2, results, nil)

	// alice = core, helper = peripheral (Active but Impact < 20)
	if tr.CoreMemberCount != 1 {
		t.Errorf("CoreMemberCount = %d, want 1", tr.CoreMemberCount)
	}
	if tr.MemberCount != 1 {
		t.Errorf("MemberCount = %d, want 1 (no risk members)", tr.MemberCount)
	}
	if tr.TotalMemberCount != 2 {
		t.Errorf("TotalMemberCount = %d, want 2", tr.TotalMemberCount)
	}
	// Average from alice only
	if tr.AvgProduction != 80 {
		t.Errorf("AvgProduction = %f, want 80 (alice only)", tr.AvgProduction)
	}
}

func TestAggregate_SilentDetection(t *testing.T) {
	// The "busy team covering for silent" scenario
	results := []scorer.Result{
		{Author: "alice", RecentlyActive: true, Production: 100, Quality: 70, Impact: 85, Role: "Architect", Style: "Builder", State: "Active"},
		{Author: "bob", RecentlyActive: true, Production: 50, Quality: 80, Impact: 40, Role: "Anchor", Style: "Balanced", State: "Active"},
		{Author: "ghost", RecentlyActive: false, Production: 10, Quality: 95, Impact: 25, Role: "—", Style: "Spread", State: "Silent"},
	}

	tr := Aggregate("team", "Backend", 3, results, nil)

	// alice + bob = core, ghost = risk (Silent)
	if tr.CoreMemberCount != 2 {
		t.Errorf("CoreMemberCount = %d, want 2", tr.CoreMemberCount)
	}
	if tr.MemberCount != 3 {
		t.Errorf("MemberCount = %d, want 3 (2 core + 1 risk)", tr.MemberCount)
	}
	// Averages from core only — not dragged down by ghost
	if tr.AvgProduction != 75 {
		t.Errorf("AvgProduction = %f, want 75 (core only)", tr.AvgProduction)
	}
	// Silent shows in distribution → detectable
	if tr.StateDist["Silent"] != 1 {
		t.Errorf("StateDist[Silent] = %d, want 1", tr.StateDist["Silent"])
	}
	// RiskRatio should reflect 1 risk / 3 effective = 33%
	if tr.Health.RiskRatio < 30 || tr.Health.RiskRatio > 35 {
		t.Errorf("RiskRatio = %f, want ~33", tr.Health.RiskRatio)
	}
}

func TestAggregate_AllInactive_WithRisk(t *testing.T) {
	results := []scorer.Result{
		{Author: "alice", RecentlyActive: false, Production: 80, Role: "Architect", Style: "Builder", State: "Former", Impact: 60},
	}

	tr := Aggregate("team", "Backend", 1, results, nil)

	// alice is risk (Former), no core
	if tr.CoreMemberCount != 0 {
		t.Errorf("CoreMemberCount = %d, want 0", tr.CoreMemberCount)
	}
	if tr.MemberCount != 1 {
		t.Errorf("MemberCount = %d, want 1 (risk member)", tr.MemberCount)
	}
	if tr.TotalMemberCount != 1 {
		t.Errorf("TotalMemberCount = %d, want 1", tr.TotalMemberCount)
	}
	// Former should appear in distributions
	if tr.StateDist["Former"] != 1 {
		t.Errorf("StateDist[Former] = %d, want 1", tr.StateDist["Former"])
	}
	// Averages should be 0 (no core members)
	if tr.AvgProduction != 0 {
		t.Errorf("AvgProduction = %f, want 0 (no core)", tr.AvgProduction)
	}
}

func TestAggregate_Empty(t *testing.T) {
	tr := Aggregate("empty", "Backend", 0, nil, nil)

	if tr.MemberCount != 0 {
		t.Errorf("MemberCount = %d, want 0", tr.MemberCount)
	}
}

func TestAggregate_AllFilteredOut(t *testing.T) {
	results := []scorer.Result{
		{Author: "alice", RecentlyActive: true, Production: 80, Role: "Architect", Style: "Builder", State: "Active", Impact: 65},
	}

	tr := Aggregate("team", "Backend", 1, results, []string{"nobody"})

	if tr.MemberCount != 0 {
		t.Errorf("MemberCount = %d, want 0", tr.MemberCount)
	}
}

func TestAggregate_AllPeripheral(t *testing.T) {
	// All members are active but below threshold — no core, no risk
	results := []scorer.Result{
		{Author: "helper1", RecentlyActive: true, Production: 5, Impact: 10, State: "Active"},
		{Author: "helper2", RecentlyActive: true, Production: 3, Impact: 8, State: "Active"},
	}

	tr := Aggregate("team", "Backend", 2, results, nil)

	if tr.MemberCount != 0 {
		t.Errorf("MemberCount = %d, want 0 (all peripheral)", tr.MemberCount)
	}
	if tr.TotalMemberCount != 2 {
		t.Errorf("TotalMemberCount = %d, want 2", tr.TotalMemberCount)
	}
}
