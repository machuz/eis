package cli

import (
	"math"
	"testing"

	"github.com/machuz/eis/v2/internal/metric"
	"github.com/machuz/eis/v2/internal/scorer"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestComputeStructuralDebt(t *testing.T) {
	tests := []struct {
		name              string
		mods              []scorer.ModuleScore
		includePeripheral bool
		topN              int

		wantSDR         float64
		wantDead        float64
		wantOrphaned    float64
		wantAtRisk      float64
		wantClassified  int
		wantDebtMass    int
		wantModuleCount int
		wantTopCount    int
		wantConcModule  string
		wantConcShare   float64
	}{
		{
			name: "mixed dead orphaned concentrated distributed",
			mods: []scorer.ModuleScore{
				{Module: "a", BlameLines: 100, Vitality: "Dead", Ownership: "Concentrated"},
				{Module: "b", BlameLines: 300, Vitality: "Stable", Ownership: "Orphaned"},
				{Module: "c", BlameLines: 200, Vitality: "Stable", Ownership: "Concentrated", OwnerActive: true},
				{Module: "d", BlameLines: 400, Vitality: "Stable", Ownership: "Distributed"},
			},
			topN:            10,
			wantSDR:         float64(100+300) / 1000.0,
			wantDead:        100.0 / 1000.0,
			wantOrphaned:    300.0 / 1000.0,
			wantAtRisk:      200.0 / 1000.0,
			wantClassified:  1000,
			wantDebtMass:    400,
			wantModuleCount: 4,
			wantTopCount:    2,
			wantConcModule:  "b", // 300 is the largest debt module
			wantConcShare:   300.0 / 400.0,
		},
		{
			name: "dead takes precedence over orphaned (no double count)",
			mods: []scorer.ModuleScore{
				{Module: "x", BlameLines: 500, Vitality: "Dead", Ownership: "Orphaned"},
			},
			topN:            10,
			wantSDR:         1.0,
			wantDead:        1.0,
			wantOrphaned:    0.0,
			wantAtRisk:      0.0,
			wantClassified:  500,
			wantDebtMass:    500,
			wantModuleCount: 1,
			wantTopCount:    1,
			wantConcModule:  "x",
			wantConcShare:   1.0,
		},
		{
			name: "concentrated but owner inactive is NOT at-risk",
			mods: []scorer.ModuleScore{
				{Module: "a", BlameLines: 100, Vitality: "Stable", Ownership: "Concentrated", OwnerActive: false},
				{Module: "b", BlameLines: 100, Vitality: "Stable", Ownership: "Distributed"},
			},
			topN:            10,
			wantSDR:         0.0,
			wantAtRisk:      0.0,
			wantClassified:  200,
			wantDebtMass:    0,
			wantModuleCount: 2,
			wantTopCount:    0,
			wantConcModule:  "",
			wantConcShare:   0.0,
		},
		{
			name: "peripheral sentinel excluded from denominator by default",
			mods: []scorer.ModuleScore{
				{Module: metric.PeripheralModule, BlameLines: 900, Vitality: "Stable", Ownership: "Distributed"},
				{Module: "real", BlameLines: 100, Vitality: "Stable", Ownership: "Orphaned"},
			},
			topN:            10,
			wantSDR:         1.0, // 100 orphaned / 100 classified
			wantOrphaned:    1.0,
			wantClassified:  100,
			wantDebtMass:    100,
			wantModuleCount: 1,
			wantTopCount:    1,
			wantConcModule:  "real",
			wantConcShare:   1.0,
		},
		{
			name: "peripheral included in denominator when flag set",
			mods: []scorer.ModuleScore{
				{Module: metric.PeripheralModule, BlameLines: 900, Vitality: "Stable", Ownership: "Distributed"},
				{Module: "real", BlameLines: 100, Vitality: "Stable", Ownership: "Orphaned"},
			},
			includePeripheral: true,
			topN:              10,
			wantSDR:           100.0 / 1000.0,
			wantOrphaned:      100.0 / 1000.0,
			wantClassified:    1000,
			wantDebtMass:      100,
			wantModuleCount:   2,
			wantTopCount:      1,
			wantConcModule:    "real",
			wantConcShare:     1.0,
		},
		{
			name: "zero/negative mass modules are skipped",
			mods: []scorer.ModuleScore{
				{Module: "a", BlameLines: 0, Vitality: "Dead", Ownership: "Orphaned"},
				{Module: "b", BlameLines: 250, Vitality: "Dead", Ownership: "Concentrated"},
			},
			topN:            10,
			wantSDR:         1.0,
			wantDead:        1.0,
			wantClassified:  250,
			wantDebtMass:    250,
			wantModuleCount: 1,
			wantTopCount:    1,
			wantConcModule:  "b",
			wantConcShare:   1.0,
		},
		{
			name:            "empty input yields zeroed report",
			mods:            nil,
			topN:            10,
			wantSDR:         0.0,
			wantClassified:  0,
			wantDebtMass:    0,
			wantModuleCount: 0,
			wantTopCount:    0,
		},
		{
			name: "topN truncates the drill list but not the metrics",
			mods: []scorer.ModuleScore{
				{Module: "a", BlameLines: 400, Vitality: "Dead"},
				{Module: "b", BlameLines: 300, Ownership: "Orphaned"},
				{Module: "c", BlameLines: 200, Ownership: "Orphaned"},
			},
			topN:            1,
			wantSDR:         1.0,
			wantDead:        400.0 / 900.0,
			wantOrphaned:    500.0 / 900.0,
			wantClassified:  900,
			wantDebtMass:    900,
			wantModuleCount: 3,
			wantTopCount:    1,   // truncated
			wantConcModule:  "a", // still computed over the FULL set
			wantConcShare:   400.0 / 900.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeStructuralDebt("Backend", tt.mods, nil, tt.includePeripheral, tt.topN)

			if !approx(got.SDR, tt.wantSDR) {
				t.Errorf("SDR = %v, want %v", got.SDR, tt.wantSDR)
			}
			if !approx(got.DeadRatio, tt.wantDead) {
				t.Errorf("DeadRatio = %v, want %v", got.DeadRatio, tt.wantDead)
			}
			if !approx(got.OrphanedRatio, tt.wantOrphaned) {
				t.Errorf("OrphanedRatio = %v, want %v", got.OrphanedRatio, tt.wantOrphaned)
			}
			if !approx(got.AtRiskRatio, tt.wantAtRisk) {
				t.Errorf("AtRiskRatio = %v, want %v", got.AtRiskRatio, tt.wantAtRisk)
			}
			if got.ClassifiedMass != tt.wantClassified {
				t.Errorf("ClassifiedMass = %d, want %d", got.ClassifiedMass, tt.wantClassified)
			}
			if got.DebtMass != tt.wantDebtMass {
				t.Errorf("DebtMass = %d, want %d", got.DebtMass, tt.wantDebtMass)
			}
			if got.ModuleCount != tt.wantModuleCount {
				t.Errorf("ModuleCount = %d, want %d", got.ModuleCount, tt.wantModuleCount)
			}
			if len(got.TopDebtModules) != tt.wantTopCount {
				t.Errorf("len(TopDebtModules) = %d, want %d", len(got.TopDebtModules), tt.wantTopCount)
			}
			if got.Concentration.TopModule != tt.wantConcModule {
				t.Errorf("Concentration.TopModule = %q, want %q", got.Concentration.TopModule, tt.wantConcModule)
			}
			if !approx(got.Concentration.Share, tt.wantConcShare) {
				t.Errorf("Concentration.Share = %v, want %v", got.Concentration.Share, tt.wantConcShare)
			}
		})
	}
}

func TestComputeStructuralDebt_OwnerNameWired(t *testing.T) {
	mods := []scorer.ModuleScore{
		{Module: "svc/auth", BlameLines: 200, Ownership: "Orphaned"},
	}
	names := map[string]string{"svc/auth": "tanaka"}
	got := computeStructuralDebt("Backend", mods, names, false, 10)
	if len(got.TopDebtModules) != 1 {
		t.Fatalf("expected 1 debt module, got %d", len(got.TopDebtModules))
	}
	if got.TopDebtModules[0].LastOwner != "tanaka" {
		t.Errorf("LastOwner = %q, want tanaka", got.TopDebtModules[0].LastOwner)
	}
	if got.TopDebtModules[0].Tier != "Orphaned" {
		t.Errorf("Tier = %q, want Orphaned", got.TopDebtModules[0].Tier)
	}
}

func TestOwnerNames(t *testing.T) {
	own := []metric.ModuleOwnership{
		{Module: "a", TopAuthor: "alice"},
		{Module: "b", TopAuthor: "bob"},
		{Module: "a", TopAuthor: "carol"}, // later record wins (mirrors ScoreModules)
	}
	got := ownerNames(own)
	if got["a"] != "carol" {
		t.Errorf("ownerNames[a] = %q, want carol (last wins)", got["a"])
	}
	if got["b"] != "bob" {
		t.Errorf("ownerNames[b] = %q, want bob", got["b"])
	}
	if ownerNames(nil) != nil {
		t.Errorf("ownerNames(nil) should be nil")
	}
}
