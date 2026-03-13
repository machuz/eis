package team

import "github.com/machuz/engineering-impact-score/internal/scorer"

// TeamClassification holds the 5-axis team-level topology.
type TeamClassification struct {
	Character TeamLabel // composite: team personality (derived from Structure × Culture)
	Structure TeamLabel // from Role distribution: who is on the team
	Culture   TeamLabel // from Style distribution: how they work
	Phase     TeamLabel // from State distribution: lifecycle stage
	Risk      TeamLabel // from Health metrics: biggest concern
}

// TeamLabel is a named classification with confidence.
type TeamLabel struct {
	Name       string
	Confidence float64
}

// Classify assigns 5-axis labels to a team.
func Classify(tr TeamResult) TeamClassification {
	structure := classifyStructure(tr)
	culture := classifyCulture(tr)
	phase := classifyPhase(tr)
	risk := classifyRisk(tr)
	character := classifyCharacter(tr, structure, culture, phase)

	return TeamClassification{
		Character: character,
		Structure: structure,
		Culture:   culture,
		Phase:     phase,
		Risk:      risk,
	}
}

// weightedRatio computes the influence-weighted ratio for members matching pred.
// High-output members carry more weight: weight = max(Total/100, 0.1).
// This reflects the sociological observation that strong contributors
// shape team culture and structure more than low-output members.
func weightedRatio(members []scorer.Result, pred func(scorer.Result) bool) float64 {
	if len(members) == 0 {
		return 0
	}
	var matched, total float64
	for _, m := range members {
		w := m.Total / 100.0
		if w < 0.1 {
			w = 0.1
		}
		total += w
		if pred(m) {
			matched += w
		}
	}
	if total == 0 {
		return 0
	}
	return matched / total
}

// --- Structure: from Role distribution ---

func classifyStructure(tr TeamResult) TeamLabel {
	if tr.MemberCount == 0 {
		return TeamLabel{"—", 0}
	}

	n := float64(tr.MemberCount)
	architects := tr.RoleDist["Architect"]
	anchors := tr.RoleDist["Anchor"]
	producers := tr.RoleDist["Producer"]
	noneCount := tr.RoleDist["—"]
	noneRatio := float64(noneCount) / n
	aar := tr.Health.AAR

	wArchitect := weightedRatio(tr.Members, func(m scorer.Result) bool { return m.Role == "Architect" })
	wProducer := weightedRatio(tr.Members, func(m scorer.Result) bool { return m.Role == "Producer" })
	wAnchor := weightedRatio(tr.Members, func(m scorer.Result) bool { return m.Role == "Anchor" })

	// Count Architect/Builders — they design AND implement, so AAR overload doesn't apply
	architectBuilders := 0
	for _, m := range tr.Members {
		if m.Role == "Architect" && m.Style == "Builder" {
			architectBuilders++
		}
	}

	rules := []labelRule{
		// Architectural Engine: Architect≥1, Anchor≥2, AAR 0.3-0.8, low unstructured ratio
		{"Architectural Engine", func() float64 {
			if architects < 1 || anchors < 2 {
				return 0
			}
			if aar < 0.3 || aar > 0.8 {
				return 0
			}
			if noneRatio > 0.4 {
				return 0 // too many unstructured → Emerging Architecture instead
			}
			return clamp(minf(wArchitect*3, wAnchor*2), 0, 1)
		}()},

		// Architectural Team: Architect≥1, Anchor≥1, low unstructured ratio
		{"Architectural Team", func() float64 {
			if architects < 1 || anchors < 1 {
				return 0
			}
			if noneRatio > 0.4 {
				return 0
			}
			return clamp(minf(wArchitect*2.5, wAnchor*2), 0, 0.90)
		}()},

		// Architecture-Heavy: Architect≥1, AAR>2.0, but NOT if all Architects are also Builders
		// Architect/Builders design AND implement, so they don't cause the
		// "design outpaces implementation" problem.
		{"Architecture-Heavy", func() float64 {
			if architects < 1 || aar <= 2.0 {
				return 0
			}
			if architectBuilders == architects {
				return 0 // all Architects are Builders — no overload
			}
			return clamp(wArchitect*2, 0, 0.85)
		}()},

		// Emerging Architecture: Architect≥1, "—" > 40%
		{"Emerging Architecture", func() float64 {
			if architects < 1 || noneRatio <= 0.4 {
				return 0
			}
			return clamp(wArchitect*2, 0, 0.85)
		}()},

		// Delivery Team: Producer > 50%
		{"Delivery Team", func() float64 {
			if float64(producers)/n < 0.5 {
				return 0
			}
			return wProducer
		}()},

		// Maintenance Team: Anchor heavy, no Architect
		{"Maintenance Team", func() float64 {
			if architects > 0 || float64(anchors)/n < 0.4 {
				return 0
			}
			return wAnchor
		}()},

		// Unstructured: "—" > 50%
		{"Unstructured", func() float64 {
			if noneRatio <= 0.5 {
				return 0
			}
			return noneRatio
		}()},

		// Balanced: default
		{"Balanced", 0.30},
	}

	return pickBestLabel(rules)
}

// --- Culture: from Style distribution ---

func classifyCulture(tr TeamResult) TeamLabel {
	if tr.MemberCount == 0 {
		return TeamLabel{"—", 0}
	}

	wBuilder := weightedRatio(tr.Members, func(m scorer.Result) bool { return m.Style == "Builder" })
	wBalanced := weightedRatio(tr.Members, func(m scorer.Result) bool { return m.Style == "Balanced" })
	wResilient := weightedRatio(tr.Members, func(m scorer.Result) bool { return m.Style == "Resilient" })
	wMass := weightedRatio(tr.Members, func(m scorer.Result) bool { return m.Style == "Mass" })
	wRescue := weightedRatio(tr.Members, func(m scorer.Result) bool { return m.Style == "Rescue" })
	wChurn := weightedRatio(tr.Members, func(m scorer.Result) bool { return m.Style == "Churn" })
	wSpread := weightedRatio(tr.Members, func(m scorer.Result) bool { return m.Style == "Spread" })

	rules := []labelRule{
		// Builder Culture
		{"Builder", func() float64 {
			if wBuilder < 0.15 {
				return 0
			}
			return clamp(wBuilder*1.2, 0, 1)
		}()},

		// Stability Culture: Balanced + Resilient dominant
		{"Stability", func() float64 {
			combined := wBalanced + wResilient
			if combined < 0.5 {
				return 0
			}
			return clamp(combined*0.9, 0, 1)
		}()},

		// Mass Production Culture
		{"Mass Production", func() float64 {
			if wMass < 0.3 {
				return 0
			}
			return wMass
		}()},

		// Firefighting Culture: Rescue + Churn
		{"Firefighting", func() float64 {
			combined := wRescue + wChurn
			if combined < 0.2 {
				return 0
			}
			return combined
		}()},

		// Exploration Culture: Spread dominant
		{"Exploration", func() float64 {
			if wSpread < 0.2 {
				return 0
			}
			return wSpread
		}()},

		// Mixed: no dominant culture
		{"Mixed", 0.25},
	}

	return pickBestLabel(rules)
}

// --- Phase: from State distribution ---

func classifyPhase(tr TeamResult) TeamLabel {
	if tr.MemberCount == 0 {
		return TeamLabel{"—", 0}
	}

	n := float64(tr.MemberCount)
	growingRatio := float64(tr.StateDist["Growing"]) / n
	activeRatio := float64(tr.StateDist["Active"]) / n
	formerRatio := float64(tr.StateDist["Former"]) / n
	silentRatio := float64(tr.StateDist["Silent"]) / n
	fragileRatio := float64(tr.StateDist["Fragile"]) / n
	riskRatio := formerRatio + silentRatio + fragileRatio

	wActive := weightedRatio(tr.Members, func(m scorer.Result) bool { return m.State == "Active" })

	rules := []labelRule{
		// Emerging: Growing > 40%
		{"Emerging", func() float64 {
			if growingRatio < 0.4 {
				return 0
			}
			return presence(growingRatio)
		}()},

		// Scaling: Growing 20-40%
		{"Scaling", func() float64 {
			if growingRatio < 0.2 || growingRatio >= 0.4 {
				return 0
			}
			return minf(presence(growingRatio), healthHigh(tr.Health.GrowthPotential))
		}()},

		// Mature: Active > 80%, high sustainability
		{"Mature", func() float64 {
			if activeRatio < 0.8 {
				return 0
			}
			return minf(presence(wActive), healthHigh(tr.Health.Sustainability))
		}()},

		// Stable: Active > 60%
		{"Stable", func() float64 {
			if activeRatio < 0.6 {
				return 0
			}
			return 0.50 + wActive*0.3
		}()},

		// Legacy-Heavy: risk states high but core team is strong (AvgTotal >= 40, has Architect)
		// "Strong but carrying historical weight" — not truly declining
		{"Legacy-Heavy", func() float64 {
			if riskRatio < 0.3 || tr.AvgTotal < 40 {
				return 0
			}
			if tr.RoleDist["Architect"] == 0 {
				return 0
			}
			return minf(presence(riskRatio), axisHigh(tr.AvgTotal))
		}()},

		// Mature with Attrition: moderate risk (20-40%) but active core still dominant
		{"Mature with Attrition", func() float64 {
			if riskRatio < 0.2 || riskRatio >= 0.4 {
				return 0
			}
			if activeRatio < 0.4 {
				return 0
			}
			return minf(presence(riskRatio)*0.8, 0.50+wActive*0.3)
		}()},

		// Declining: risk states > 30% AND core team is weak (no strong leadership)
		{"Declining", func() float64 {
			if riskRatio < 0.3 {
				return 0
			}
			// If core team is strong with an Architect, Legacy-Heavy should win
			if tr.AvgTotal >= 40 && tr.RoleDist["Architect"] > 0 {
				return presence(riskRatio) * 0.5 // reduced confidence
			}
			return presence(riskRatio)
		}()},

		// Rebuilding: mix of growing + former/silent
		{"Rebuilding", func() float64 {
			if growingRatio < 0.1 || riskRatio < 0.1 {
				return 0
			}
			return minf(presence(growingRatio), presence(riskRatio))
		}()},
	}

	return pickBestLabel(rules)
}

// --- Risk: from Health metrics ---

func classifyRisk(tr TeamResult) TeamLabel {
	if tr.MemberCount == 0 {
		return TeamLabel{"—", 0}
	}

	rules := []labelRule{
		// Bus Factor: small team, high indispensability concentration
		{"Bus Factor", func() float64 {
			if tr.MemberCount > 5 {
				return 0
			}
			return minf(
				axisHigh(tr.AvgIndispensability),
				1.0-float64(tr.MemberCount)/10.0,
			)
		}()},

		// Design Vacuum: no Architect, low complementarity
		{"Design Vacuum", func() float64 {
			if tr.RoleDist["Architect"] > 0 {
				return 0
			}
			return 1.0 - tr.Health.Complementarity/100.0
		}()},

		// Quality Drift: low consistency
		{"Quality Drift", func() float64 {
			if tr.Health.QualityConsistency > 60 {
				return 0
			}
			return 1.0 - tr.Health.QualityConsistency/100.0
		}()},

		// Debt Spiral: accumulating debt
		{"Debt Spiral", func() float64 {
			if tr.Health.DebtBalance > 45 {
				return 0
			}
			return 1.0 - tr.Health.DebtBalance/100.0
		}()},

		// Talent Drain: high risk ratio
		{"Talent Drain", func() float64 {
			if tr.Health.RiskRatio < 25 {
				return 0
			}
			return tr.Health.RiskRatio / 100.0
		}()},

		// Healthy: no critical risks
		{"Healthy", 0.30},
	}

	return pickBestLabel(rules)
}

// --- Character: composite from Structure × Culture × metrics ---

func classifyCharacter(tr TeamResult, structure, culture, phase TeamLabel) TeamLabel {
	s := structure.Name
	cu := culture.Name
	sc := tr.Health.ArchitectureCoverage
	aar := tr.Health.AAR
	pd := tr.Health.ProductivityDensity

	// Count Architect/Builders for AAR relaxation
	archBuilders := 0
	totalArchitects := 0
	for _, m := range tr.Members {
		if m.Role == "Architect" {
			totalArchitects++
			if m.Style == "Builder" {
				archBuilders++
			}
		}
	}
	allArchitectsAreBuilders := totalArchitects > 0 && archBuilders == totalArchitects

	rules := []labelRule{
		// Elite: SC>0.4, AAR 0.3-0.8, PD>35 — balanced architecture + protection + output
		// AAR constraint relaxed if all Architects are also Builders (they design AND implement)
		{"Elite", func() float64 {
			if sc <= 0.4 || pd <= 35 {
				return 0
			}
			if aar < 0.3 || (aar > 0.8 && !allArchitectsAreBuilders) {
				return 0
			}
			return clamp(sc+pd/100, 0, 1)
		}()},

		// Fortress: Architectural structure + Stability culture
		{"Fortress", func() float64 {
			if (s == "Architectural Engine" || s == "Architectural Team") && cu == "Stability" {
				return minf(structure.Confidence, culture.Confidence)
			}
			return 0
		}()},

		// Pioneer: Architectural + Builder culture
		{"Pioneer", func() float64 {
			if (s == "Architectural Engine" || s == "Architectural Team" || s == "Emerging Architecture") && cu == "Builder" {
				return minf(structure.Confidence, culture.Confidence)
			}
			return 0
		}()},

		// Academy: Builder culture + growing phase
		{"Academy", func() float64 {
			if cu != "Builder" {
				return 0
			}
			if phase.Name == "Scaling" || phase.Name == "Emerging" {
				return minf(culture.Confidence, phase.Confidence)
			}
			return culture.Confidence * 0.5
		}()},

		// Feature Factory: High PD but weak structure — ships fast, erodes architecture
		{"Feature Factory", func() float64 {
			if pd <= 40 || sc >= 0.2 {
				return 0
			}
			return clamp(pd/100+0.3, 0, 0.85)
		}()},

		// Factory: Delivery + Mass Production
		{"Factory", func() float64 {
			if s == "Delivery Team" && cu == "Mass Production" {
				return minf(structure.Confidence, culture.Confidence)
			}
			if s == "Delivery Team" {
				return structure.Confidence * 0.6
			}
			return 0
		}()},

		// Firefighting: Firefighting culture dominant
		{"Firefighting", func() float64 {
			if cu != "Firefighting" {
				return 0
			}
			return culture.Confidence
		}()},

		// Guardian: Maintenance + Stability, or high SC + low PD (Legacy Maintenance)
		{"Guardian", func() float64 {
			if s == "Maintenance Team" && (cu == "Stability" || cu == "Mixed") {
				return minf(structure.Confidence, 0.70)
			}
			// Legacy Maintenance: SC>0.5 but PD<20
			if sc > 0.5 && pd < 20 {
				return 0.65
			}
			return 0
		}()},

		// Explorer: Exploration culture
		{"Explorer", func() float64 {
			if cu != "Exploration" {
				return 0
			}
			return culture.Confidence * 0.8
		}()},

		// Balanced: default
		{"Balanced", 0.25},
	}

	return pickBestLabel(rules)
}

// --- helpers ---

type labelRule struct {
	name  string
	score float64
}

func pickBestLabel(rules []labelRule) TeamLabel {
	best := TeamLabel{"—", 0}
	for _, r := range rules {
		if r.score > best.Confidence {
			best = TeamLabel{r.name, r.score}
		}
	}
	best.Confidence = float64(int(best.Confidence*100+0.5)) / 100
	return best
}

// presence: 0.0 at ratio=0, ramps to 1.0 at ratio>=0.3
func presence(ratio float64) float64 {
	if ratio >= 0.3 {
		return 1.0
	}
	if ratio > 0 {
		return 0.3 + ratio/0.3*0.7
	}
	return 0
}

// healthHigh: 1.0 if health >= 70, ramp from 0 at 30
func healthHigh(v float64) float64 {
	if v >= 70 {
		return 1.0
	}
	if v >= 50 {
		return 0.5 + (v-50)/40
	}
	if v >= 30 {
		return (v - 30) / 40 * 0.3
	}
	return 0
}

// axisHigh: 1.0 if axis avg >= 60, ramp from 0 at 20
func axisHigh(v float64) float64 {
	if v >= 60 {
		return 1.0
	}
	if v >= 40 {
		return 0.5 + (v-40)/40
	}
	if v >= 20 {
		return (v - 20) / 40 * 0.3
	}
	return 0
}

// axisLow: 1.0 if axis avg < 20, ramp from 0 at 50
func axisLow(v float64) float64 {
	if v < 20 {
		return 1.0
	}
	if v < 35 {
		return 0.5 + (35-v)/30
	}
	if v < 50 {
		return (50 - v) / 30 * 0.3
	}
	return 0
}

func minf(vals ...float64) float64 {
	m := vals[0]
	for _, v := range vals[1:] {
		if v < m {
			m = v
		}
	}
	return m
}
