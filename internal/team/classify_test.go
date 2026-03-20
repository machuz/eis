package team

import (
	"testing"

	"github.com/machuz/eis/internal/scorer"
)

func TestClassify_Structure_ArchitecturalEngine(t *testing.T) {
	tr := TeamResult{
		MemberCount:      5,
		TotalMemberCount: 5,
		Members: []scorer.Result{
			{Author: "a1", Role: "Architect", Impact: 80},
			{Author: "a2", Role: "Anchor", Impact: 70},
			{Author: "a3", Role: "Anchor", Impact: 65},
			{Author: "a4", Role: "Producer", Impact: 50},
			{Author: "a5", Role: "Producer", Impact: 45},
		},
		Health: TeamHealth{
			Sustainability:  90,
			Complementarity: 80,
			AAR:             0.5, // 1/2
		},
		RoleDist:  map[string]int{"Architect": 1, "Anchor": 2, "Producer": 2},
		StyleDist: map[string]int{"Builder": 1, "Balanced": 4},
		StateDist: map[string]int{"Active": 4, "Growing": 1},
	}

	c := Classify(tr)

	if c.Structure.Name != "Architectural Engine" {
		t.Errorf("Structure = %s, want Architectural Engine", c.Structure.Name)
	}
}

func TestClassify_Structure_EmergingArchitecture(t *testing.T) {
	// Architect 1, Anchor 2, "—" 4 → Emerging Architecture
	tr := TeamResult{
		MemberCount:      7,
		TotalMemberCount: 7,
		Members: []scorer.Result{
			{Author: "a1", Role: "Architect", Impact: 75},
			{Author: "a2", Role: "Anchor", Impact: 60},
			{Author: "a3", Role: "Anchor", Impact: 55},
			{Author: "a4", Role: "—", Impact: 25},
			{Author: "a5", Role: "—", Impact: 20},
			{Author: "a6", Role: "—", Impact: 15},
			{Author: "a7", Role: "—", Impact: 10},
		},
		Health: TeamHealth{
			AAR: 0.5,
		},
		RoleDist:  map[string]int{"Architect": 1, "Anchor": 2, "—": 4},
		StyleDist: map[string]int{"Balanced": 7},
		StateDist: map[string]int{"Active": 7},
	}

	c := Classify(tr)

	if c.Structure.Name != "Emerging Architecture" {
		t.Errorf("Structure = %s, want Emerging Architecture", c.Structure.Name)
	}
}

func TestClassify_Structure_DeliveryTeam(t *testing.T) {
	tr := TeamResult{
		MemberCount:      6,
		TotalMemberCount: 6,
		Members: []scorer.Result{
			{Author: "p1", Role: "Producer", Impact: 70},
			{Author: "p2", Role: "Producer", Impact: 65},
			{Author: "p3", Role: "Producer", Impact: 60},
			{Author: "p4", Role: "Producer", Impact: 55},
			{Author: "a1", Role: "—", Impact: 20},
			{Author: "a2", Role: "—", Impact: 15},
		},
		Health:    TeamHealth{},
		RoleDist:  map[string]int{"Producer": 4, "—": 2},
		StyleDist: map[string]int{"Mass": 3, "Balanced": 3},
		StateDist: map[string]int{"Active": 6},
	}

	c := Classify(tr)

	if c.Structure.Name != "Delivery Team" {
		t.Errorf("Structure = %s, want Delivery Team", c.Structure.Name)
	}
}

func TestClassify_Culture_Builder(t *testing.T) {
	tr := TeamResult{
		MemberCount:      5,
		TotalMemberCount: 5,
		Members: []scorer.Result{
			{Author: "b1", Style: "Builder", Impact: 80},
			{Author: "b2", Style: "Builder", Impact: 70},
			{Author: "b3", Style: "Balanced", Impact: 50},
			{Author: "b4", Style: "Balanced", Impact: 40},
			{Author: "b5", Style: "Builder", Impact: 60},
		},
		Health:    TeamHealth{},
		RoleDist:  map[string]int{"Architect": 1, "Anchor": 2, "Producer": 2},
		StyleDist: map[string]int{"Builder": 3, "Balanced": 2},
		StateDist: map[string]int{"Active": 5},
	}

	c := Classify(tr)

	if c.Culture.Name != "Builder" {
		t.Errorf("Culture = %s, want Builder", c.Culture.Name)
	}
}

func TestClassify_Culture_Stability(t *testing.T) {
	tr := TeamResult{
		MemberCount:      5,
		TotalMemberCount: 5,
		Members: []scorer.Result{
			{Author: "b1", Style: "Balanced", Impact: 60},
			{Author: "b2", Style: "Balanced", Impact: 55},
			{Author: "b3", Style: "Resilient", Impact: 50},
			{Author: "b4", Style: "Balanced", Impact: 45},
			{Author: "b5", Style: "Balanced", Impact: 40},
		},
		Health:    TeamHealth{},
		RoleDist:  map[string]int{"Anchor": 3, "Producer": 2},
		StyleDist: map[string]int{"Balanced": 4, "Resilient": 1},
		StateDist: map[string]int{"Active": 5},
	}

	c := Classify(tr)

	if c.Culture.Name != "Stability" {
		t.Errorf("Culture = %s, want Stability", c.Culture.Name)
	}
}

func TestClassify_Phase_Mature(t *testing.T) {
	tr := TeamResult{
		MemberCount:      5,
		TotalMemberCount: 5,
		Members: []scorer.Result{
			{Author: "a1", State: "Active", Impact: 70},
			{Author: "a2", State: "Active", Impact: 65},
			{Author: "a3", State: "Active", Impact: 60},
			{Author: "a4", State: "Active", Impact: 55},
			{Author: "a5", State: "Active", Impact: 50},
		},
		Health: TeamHealth{
			Sustainability: 90,
		},
		RoleDist:  map[string]int{"Architect": 1, "Anchor": 2, "Producer": 2},
		StyleDist: map[string]int{"Builder": 1, "Balanced": 4},
		StateDist: map[string]int{"Active": 5},
	}

	c := Classify(tr)

	if c.Phase.Name != "Mature" {
		t.Errorf("Phase = %s, want Mature", c.Phase.Name)
	}
}

func TestClassify_Risk_DesignVacuum(t *testing.T) {
	tr := TeamResult{
		MemberCount:      5,
		TotalMemberCount: 5,
		Members: []scorer.Result{
			{Author: "p1", Role: "Producer", Impact: 30},
			{Author: "p2", Role: "Producer", Impact: 25},
			{Author: "p3", Role: "Producer", Impact: 20},
			{Author: "p4", Role: "—", Impact: 15},
			{Author: "p5", Role: "—", Impact: 10},
		},
		Health: TeamHealth{
			Complementarity:    20,
			QualityConsistency: 70,
			DebtBalance:        50,
		},
		RoleDist:  map[string]int{"Producer": 3, "—": 2},
		StyleDist: map[string]int{"Mass": 3, "Balanced": 2},
		StateDist: map[string]int{"Active": 5},
	}

	c := Classify(tr)

	if c.Risk.Name != "Design Vacuum" {
		t.Errorf("Risk = %s, want Design Vacuum", c.Risk.Name)
	}
}

func TestClassify_Risk_Healthy(t *testing.T) {
	tr := TeamResult{
		MemberCount:      5,
		TotalMemberCount: 5,
		Members: []scorer.Result{
			{Author: "a1", Role: "Architect", Impact: 70},
			{Author: "a2", Role: "Anchor", Impact: 60},
			{Author: "a3", Role: "Anchor", Impact: 55},
			{Author: "a4", Role: "Producer", Impact: 50},
			{Author: "a5", Role: "Producer", Impact: 45},
		},
		Health: TeamHealth{
			Complementarity:    80,
			QualityConsistency: 75,
			DebtBalance:        55,
		},
		RoleDist:  map[string]int{"Architect": 1, "Anchor": 2, "Producer": 2},
		StyleDist: map[string]int{"Builder": 1, "Balanced": 4},
		StateDist: map[string]int{"Active": 5},
	}

	c := Classify(tr)

	if c.Risk.Name != "Healthy" {
		t.Errorf("Risk = %s, want Healthy", c.Risk.Name)
	}
}

func TestClassify_Risk_TalentDrain(t *testing.T) {
	tr := TeamResult{
		MemberCount:      4,
		TotalMemberCount: 4,
		Members: []scorer.Result{
			{Author: "a1", Role: "Architect", State: "Active", Impact: 50},
			{Author: "a2", Role: "—", State: "Silent", Impact: 15},
			{Author: "a3", Role: "—", State: "Silent", Impact: 10},
			{Author: "a4", Role: "—", State: "Former", Impact: 5},
		},
		Health: TeamHealth{
			Complementarity:    50,
			QualityConsistency: 70,
			DebtBalance:        50,
			RiskRatio:          50,
		},
		RoleDist:  map[string]int{"Architect": 1, "—": 3},
		StyleDist: map[string]int{"Balanced": 4},
		StateDist: map[string]int{"Active": 1, "Silent": 2, "Former": 1},
	}

	c := Classify(tr)

	if c.Risk.Name != "Talent Drain" {
		t.Errorf("Risk = %s, want Talent Drain", c.Risk.Name)
	}
}

func TestClassify_Character_Elliptical(t *testing.T) {
	tr := TeamResult{
		MemberCount:      5,
		TotalMemberCount: 5,
		Members: []scorer.Result{
			{Author: "a1", Role: "Architect", Style: "Balanced", State: "Active", Impact: 80},
			{Author: "a2", Role: "Anchor", Style: "Resilient", State: "Active", Impact: 70},
			{Author: "a3", Role: "Anchor", Style: "Balanced", State: "Active", Impact: 65},
			{Author: "a4", Role: "Producer", Style: "Balanced", State: "Active", Impact: 55},
			{Author: "a5", Role: "Producer", Style: "Balanced", State: "Active", Impact: 50},
		},
		Health: TeamHealth{
			Sustainability:  90,
			Complementarity: 80,
			AAR:             0.5,
		},
		RoleDist:  map[string]int{"Architect": 1, "Anchor": 2, "Producer": 2},
		StyleDist: map[string]int{"Balanced": 4, "Resilient": 1},
		StateDist: map[string]int{"Active": 5},
	}

	c := Classify(tr)

	if c.Character.Name != "Elliptical" {
		t.Errorf("Character = %s, want Elliptical", c.Character.Name)
	}
}

func TestClassify_Character_Starburst(t *testing.T) {
	tr := TeamResult{
		MemberCount:      5,
		TotalMemberCount: 5,
		Members: []scorer.Result{
			{Author: "a1", Role: "Architect", Style: "Builder", State: "Active", Impact: 85},
			{Author: "a2", Role: "Anchor", Style: "Builder", State: "Active", Impact: 70},
			{Author: "a3", Role: "Producer", Style: "Builder", State: "Active", Impact: 65},
			{Author: "a4", Role: "Producer", Style: "Balanced", State: "Active", Impact: 50},
			{Author: "a5", Role: "—", Style: "Balanced", State: "Growing", Impact: 30},
		},
		Health: TeamHealth{
			Sustainability:  80,
			Complementarity: 70,
			AAR:             1.0,
		},
		RoleDist:  map[string]int{"Architect": 1, "Anchor": 1, "Producer": 2, "—": 1},
		StyleDist: map[string]int{"Builder": 3, "Balanced": 2},
		StateDist: map[string]int{"Active": 4, "Growing": 1},
	}

	c := Classify(tr)

	if c.Character.Name != "Starburst" {
		t.Errorf("Character = %s, want Starburst", c.Character.Name)
	}
}

func TestClassify_Empty(t *testing.T) {
	tr := TeamResult{MemberCount: 0}
	c := Classify(tr)

	if c.Structure.Name != "—" || c.Culture.Name != "—" || c.Phase.Name != "—" || c.Risk.Name != "—" {
		t.Errorf("Empty should be all —, got %s/%s/%s/%s",
			c.Structure.Name, c.Culture.Name, c.Phase.Name, c.Risk.Name)
	}
}

func TestClassify_WeightedInfluence(t *testing.T) {
	// A team with 1 high-output Builder and 4 low-output Balanced members.
	// The Builder's influence should dominate due to weighting.
	tr := TeamResult{
		MemberCount:      5,
		TotalMemberCount: 5,
		Members: []scorer.Result{
			{Author: "star", Style: "Builder", Impact: 90},
			{Author: "b1", Style: "Balanced", Impact: 15},
			{Author: "b2", Style: "Balanced", Impact: 15},
			{Author: "b3", Style: "Balanced", Impact: 15},
			{Author: "b4", Style: "Balanced", Impact: 15},
		},
		Health:    TeamHealth{},
		RoleDist:  map[string]int{"Producer": 5},
		StyleDist: map[string]int{"Builder": 1, "Balanced": 4},
		StateDist: map[string]int{"Active": 5},
	}

	c := Classify(tr)

	// Without weighting, Balanced (4) would dominate.
	// With weighting, the star Builder (Impact=90, weight=0.9) outweighs
	// 4 × Balanced (Impact=15, weight=0.15 each = 0.6 total).
	if c.Culture.Name != "Builder" {
		t.Errorf("Culture = %s, want Builder (star player should dominate)", c.Culture.Name)
	}
}
