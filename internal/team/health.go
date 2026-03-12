package team

import "math"

// CalcHealth computes team health metrics from a TeamResult.
func CalcHealth(tr TeamResult) TeamHealth {
	if tr.MemberCount == 0 {
		return TeamHealth{}
	}
	return TeamHealth{
		Complementarity:      calcComplementarity(tr),
		GrowthPotential:      calcGrowthPotential(tr),
		Sustainability:       calcSustainability(tr),
		DebtBalance:          calcDebtBalance(tr),
		ProductivityDensity:  calcProductivityDensity(tr),
		QualityConsistency:   calcQualityConsistency(tr),
		RiskRatio:            calcRiskRatio(tr),
		AAR:                  calcAAR(tr),
		AnchorDensity:        calcAnchorDensity(tr),
		ArchitectureCoverage: calcArchitectureCoverage(tr),
	}
}

// Complementarity: role diversity coverage.
// known roles = [Architect, Anchor, Cleaner, Producer, Specialist]
// coverage = unique / 5 * 80 + bonuses (clamp 0-100)
func calcComplementarity(tr TeamResult) float64 {
	knownRoles := []string{"Architect", "Anchor", "Cleaner", "Producer", "Specialist"}
	unique := 0
	for _, role := range knownRoles {
		if tr.RoleDist[role] > 0 {
			unique++
		}
	}
	coverage := float64(unique) / 5.0

	bonus := 0.0
	if tr.RoleDist["Architect"] > 0 {
		bonus += 10
	}
	if tr.RoleDist["Anchor"] > 0 {
		bonus += 5
	}
	if tr.RoleDist["Cleaner"] > 0 {
		bonus += 5
	}

	return clamp(coverage*80+bonus, 0, 100)
}

// GrowthPotential: growing ratio + mentoring capacity.
// growingRatio * 60 + Builder bonus + Cleaner bonus
func calcGrowthPotential(tr TeamResult) float64 {
	n := float64(tr.MemberCount)
	growingCount := float64(tr.StateDist["Growing"])
	growingRatio := growingCount / n

	score := growingRatio * 60
	if tr.StyleDist["Builder"] > 0 {
		score += 20
	}
	if tr.RoleDist["Cleaner"] > 0 {
		score += 20
	}
	return clamp(score, 0, 100)
}

// Sustainability: inverse of risk state ratio + Architect stability.
func calcSustainability(tr TeamResult) float64 {
	n := float64(tr.MemberCount)
	riskCount := float64(tr.StateDist["Former"] + tr.StateDist["Silent"] + tr.StateDist["Fragile"])
	riskRatio := riskCount / n

	score := (1 - riskRatio) * 80
	if tr.RoleDist["Architect"] > 0 {
		score += 20
	}
	return clamp(score, 0, 100)
}

// DebtBalance: average debt cleanup score. 50 = neutral.
func calcDebtBalance(tr TeamResult) float64 {
	return clamp(tr.AvgDebtCleanup, 0, 100)
}

// ProductivityDensity: how much output per core member.
// Uses average production score (from core members only).
// Small team bonus uses CoreMemberCount — the real contributors.
func calcProductivityDensity(tr TeamResult) float64 {
	if tr.CoreMemberCount == 0 {
		return 0
	}
	base := tr.AvgProduction

	// Small team bonus: few core members with high avg production is notable
	if tr.CoreMemberCount <= 3 && base >= 50 {
		base = clamp(base*1.2, 0, 100)
	} else if tr.CoreMemberCount <= 5 && base >= 50 {
		base = clamp(base*1.1, 0, 100)
	}

	return clamp(base, 0, 100)
}

// QualityConsistency: team quality level and consistency (low variance = good).
// Uses CoreMembers only — risk members' quality doesn't reflect active team consistency.
// score = avgQuality * 0.6 + (100 - stdev) * 0.4
func calcQualityConsistency(tr TeamResult) float64 {
	if tr.CoreMemberCount < 2 {
		return clamp(tr.AvgQuality, 0, 100)
	}

	// Calculate standard deviation of quality from core members
	var sumSq float64
	for _, m := range tr.CoreMembers {
		diff := m.Quality - tr.AvgQuality
		sumSq += diff * diff
	}
	stdev := math.Sqrt(sumSq / float64(tr.CoreMemberCount))

	// Higher avg quality + lower variance = better consistency
	score := tr.AvgQuality*0.6 + clamp(100-stdev*2, 0, 100)*0.4
	return clamp(score, 0, 100)
}

// calcRiskRatio: percentage of members in risk states (Former, Silent, Fragile).
func calcRiskRatio(tr TeamResult) float64 {
	n := float64(tr.MemberCount)
	riskCount := float64(tr.StateDist["Former"] + tr.StateDist["Silent"] + tr.StateDist["Fragile"])
	return clamp(riskCount/n*100, 0, 100)
}

// calcAAR: Architect-to-Anchor Ratio.
// Returns the raw ratio. Special cases: 0 architects → 0, 0 anchors → -1 (sentinel for "no anchors").
func calcAAR(tr TeamResult) float64 {
	architects := tr.RoleDist["Architect"]
	anchors := tr.RoleDist["Anchor"]
	if architects == 0 {
		return 0
	}
	if anchors == 0 {
		return -1 // sentinel: architect(s) present but no anchor
	}
	return float64(architects) / float64(anchors)
}

// calcAnchorDensity: Anchors / Active members (0-1).
func calcAnchorDensity(tr TeamResult) float64 {
	if tr.MemberCount == 0 {
		return 0
	}
	return float64(tr.RoleDist["Anchor"]) / float64(tr.MemberCount)
}

// calcArchitectureCoverage: (Architects + Anchors) / Team size (0-1).
func calcArchitectureCoverage(tr TeamResult) float64 {
	if tr.MemberCount == 0 {
		return 0
	}
	return float64(tr.RoleDist["Architect"]+tr.RoleDist["Anchor"]) / float64(tr.MemberCount)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
