package team

import (
	"math"
	"testing"

	"github.com/machuz/eis/v2/internal/scorer"
)

func TestCalcHealth_FullTeam(t *testing.T) {
	members := []scorer.Result{
		{Catalysis: 90, Role: "Architect", Style: "Builder", State: "Active"},
		{Catalysis: 85, Role: "Anchor", Style: "Balanced", State: "Growing"},
		{Catalysis: 80, Role: "Cleaner", Style: "Rescue", State: "Active"},
		{Catalysis: 75, Role: "Producer", Style: "Mass", State: "Active"},
		{Catalysis: 70, Role: "Specialist", Style: "Balanced", State: "Active"},
	}
	tr := TeamResult{
		MemberCount:     5,
		CoreMemberCount: 5,
		Members:         members,
		CoreMembers:     members,
		AvgProduction:   60,
		AvgCatalysis:    80,
		AvgDebtCleanup:  55,
		RoleDist: map[string]int{
			"Architect": 1, "Anchor": 1, "Cleaner": 1, "Producer": 1, "Specialist": 1,
		},
		StyleDist: map[string]int{
			"Builder": 1, "Balanced": 2, "Rescue": 1, "Mass": 1,
		},
		StateDist: map[string]int{
			"Active": 4, "Growing": 1,
		},
	}

	h := CalcHealth(tr)

	// All 5 roles covered: 5/5 * 80 + 10 + 5 + 5 = 100
	if h.Complementarity != 100 {
		t.Errorf("Complementarity = %f, want 100", h.Complementarity)
	}

	// Growing: 1/5 * 60 + Builder(20) + Cleaner(20) = 12 + 40 = 52
	if h.GrowthPotential != 52 {
		t.Errorf("GrowthPotential = %f, want 52", h.GrowthPotential)
	}

	// No risk states: (1-0)*80 + Architect(20) = 100
	if h.Sustainability != 100 {
		t.Errorf("Sustainability = %f, want 100", h.Sustainability)
	}

	// Debt = avg 55
	if h.DebtBalance != 55 {
		t.Errorf("DebtBalance = %f, want 55", h.DebtBalance)
	}

	// Risk ratio = 0%
	if h.RiskRatio != 0 {
		t.Errorf("RiskRatio = %f, want 0", h.RiskRatio)
	}
}

func TestCalcHealth_RiskyTeam(t *testing.T) {
	// 1 core (Active) + 3 risk (Former, Silent, Fragile) = 4 effective
	members := []scorer.Result{
		{Catalysis: 90, Role: "Producer", Style: "Mass", State: "Former"},
		{Catalysis: 30, Role: "Producer", Style: "Churn", State: "Silent"},
		{Catalysis: 50, Role: "—", Style: "Spread", State: "Fragile"},
		{Catalysis: 80, Role: "Producer", Style: "Mass", State: "Active"},
	}
	coreMembers := []scorer.Result{members[3]} // only Active member
	tr := TeamResult{
		MemberCount:     4,
		CoreMemberCount: 1,
		Members:         members,
		CoreMembers:     coreMembers,
		AvgProduction:   40,
		AvgCatalysis:    80, // from core only
		AvgDebtCleanup:  35,
		RoleDist: map[string]int{
			"Producer": 3, "—": 1,
		},
		StyleDist: map[string]int{
			"Mass": 2, "Churn": 1, "Spread": 1,
		},
		StateDist: map[string]int{
			"Former": 1, "Silent": 1, "Fragile": 1, "Active": 1,
		},
	}

	h := CalcHealth(tr)

	// Only Producer: 1/5 * 80 + 0 = 16
	if h.Complementarity != 16 {
		t.Errorf("Complementarity = %f, want 16", h.Complementarity)
	}

	// 3/4 risk states: (1-0.75)*80 + 0 = 20
	if h.Sustainability != 20 {
		t.Errorf("Sustainability = %f, want 20", h.Sustainability)
	}

	// Risk ratio = 75%
	if h.RiskRatio != 75 {
		t.Errorf("RiskRatio = %f, want 75", h.RiskRatio)
	}

	// Debt = 35
	if h.DebtBalance != 35 {
		t.Errorf("DebtBalance = %f, want 35", h.DebtBalance)
	}
}

func TestCalcHealth_Empty(t *testing.T) {
	tr := TeamResult{MemberCount: 0}
	h := CalcHealth(tr)

	if h.Complementarity != 0 || h.Sustainability != 0 || h.RiskRatio != 0 {
		t.Errorf("Empty team should have all zeros, got %+v", h)
	}
}

func TestCalcCatalysisConsistency_LowVariance(t *testing.T) {
	members := []scorer.Result{
		{Catalysis: 88},
		{Catalysis: 91},
		{Catalysis: 91},
	}
	tr := TeamResult{
		MemberCount:     3,
		CoreMemberCount: 3,
		AvgCatalysis:    90,
		Members:         members,
		CoreMembers:     members,
	}

	h := CalcHealth(tr)

	// High avg, low variance → should be near 90
	if h.CatalysisConsistency < 80 {
		t.Errorf("CatalysisConsistency = %f, want >= 80 for consistent high quality", h.CatalysisConsistency)
	}
}

func TestCalcCatalysisConsistency_HighVariance(t *testing.T) {
	members := []scorer.Result{
		{Catalysis: 95},
		{Catalysis: 60},
		{Catalysis: 25},
	}
	tr := TeamResult{
		MemberCount:     3,
		CoreMemberCount: 3,
		AvgCatalysis:    60,
		Members:         members,
		CoreMembers:     members,
	}

	h := CalcHealth(tr)

	// Medium avg, high variance → lower score
	if h.CatalysisConsistency > 70 {
		t.Errorf("CatalysisConsistency = %f, want < 70 for inconsistent quality", h.CatalysisConsistency)
	}
}

func TestCalcProductivityDensity_SmallHighOutput(t *testing.T) {
	tr := TeamResult{
		MemberCount:     3,
		CoreMemberCount: 3,
		AvgProduction:   70,
	}

	h := CalcHealth(tr)

	// Small team with high production gets 1.2x bonus: 70*1.2 = 84
	if math.Abs(h.ProductivityDensity-84) > 0.1 {
		t.Errorf("ProductivityDensity = %f, want 84 for small high-output team", h.ProductivityDensity)
	}
}

func TestCalcProductivityDensity_LargeTeam(t *testing.T) {
	tr := TeamResult{
		MemberCount:     10,
		CoreMemberCount: 10,
		AvgProduction:   70,
	}

	h := CalcHealth(tr)

	// Large team: no bonus
	if math.Abs(h.ProductivityDensity-70) > 0.1 {
		t.Errorf("ProductivityDensity = %f, want 70 for large team (no bonus)", h.ProductivityDensity)
	}
}

func TestCalcProductivityDensity_CoreVsEffective(t *testing.T) {
	// 3 core + 2 risk = 5 effective, but ProdDensity bonus uses CoreMemberCount
	tr := TeamResult{
		MemberCount:     5,
		CoreMemberCount: 3,
		AvgProduction:   70, // from core only
	}

	h := CalcHealth(tr)

	// CoreMemberCount=3, so small team bonus applies: 70*1.2 = 84
	if math.Abs(h.ProductivityDensity-84) > 0.1 {
		t.Errorf("ProductivityDensity = %f, want 84 (bonus based on core count)", h.ProductivityDensity)
	}
}
