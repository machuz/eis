package cli

import (
	"reflect"
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/output"
)

// TestStructuralDebtLeanMatchesFull is the correctness invariant for the lean
// debt pipeline: the structural-debt report (SDR, ratios, masses, module count,
// top_debt_modules, concentration, insufficient_data) must be IDENTICAL whether
// produced by the lean path or the full pipeline. Lean only skips inputs debt
// never reads (modulePressure/moduleSurvival feed non-Dead Vitality & Coupling),
// so any divergence here means the cut reached a debt dependency = a bug.
func TestStructuralDebtLeanMatchesFull(t *testing.T) {
	repo := buildRichFixture(t)

	// Pin the envelope clock for pipeline determinism. (ScoreModules' owner/
	// untouched day counts still use wall clock, but both runs share the same
	// author/module dates, and the report only exposes those as ints far from any
	// boundary in this fixture — so the compared report is stable.)
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	base := AnalyzeOptions{
		Format:       "json",
		Workers:      4,
		PressureMode: "include",
		NoCache:      true,
		AnalysisTime: at,
		// Match the structural-debt command so lean and full agree on blame
		// attribution (the command forces move detection off in both modes).
		BlameMoveDetection: "off",
	}

	reports := func(lean bool) []output.StructuralDebtReport {
		o := base
		o.LeanDebt = lean
		drs, cfg, _, err := RunAnalyzePipeline(o, []string{repo})
		if err != nil {
			t.Fatalf("RunAnalyzePipeline(lean=%v): %v", lean, err)
		}
		applyExcl := !architectureDeclared(cfg)
		var rs []output.StructuralDebtReport
		for _, d := range drs {
			rs = append(rs, computeStructuralDebt(
				string(d.Domain), d.ModuleScores, ownerNames(d.Ownership),
				false, applyExcl, debtOwnerGoneDaysDefault, debtStaleDaysDefault, 10))
		}
		return rs
	}

	full := reports(false)
	lean := reports(true)

	if len(full) == 0 {
		t.Fatal("fixture produced no domains — cannot validate equivalence")
	}
	if !reflect.DeepEqual(full, lean) {
		t.Fatalf("lean debt pipeline diverges from full:\n full: %+v\n lean: %+v", full, lean)
	}
}
