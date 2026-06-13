package metric

import (
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

// TestIndispensability_SuppressesPeripheral verifies that the peripheral
// denoising bucket never surfaces as a real module: it must not appear as a
// bus-factor risk or ownership row, and owning only peripheral mass must not
// grant any Indispensability score. A real declared module is unaffected.
func TestIndispensability_SuppressesPeripheral(t *testing.T) {
	mr := NewModuleResolver(nil)
	// "20241030/a/..." is a single-month fallback module → folds into the
	// peripheral sentinel; "app/domain/star/..." is declared → survives.
	commits := []git.Commit{
		livenessCommit(2024, time.October, "20241030/a/migrate.sql"),
		livenessCommit(2024, time.October, "app/domain/star/star.go"),
	}
	folded := mr.WithFold(ComputeModuleFold(commits, mr, 2))

	// Sanity: the fallback path really does resolve to the sentinel.
	if got := folded.ModuleOf("20241030/a/migrate.sql"); got != PeripheralModule {
		t.Fatalf("setup: folded path resolved to %q, want %q", got, PeripheralModule)
	}

	blame := []git.BlameLine{
		{Filename: "20241030/a/migrate.sql", Author: "ghost"},
		{Filename: "20241030/a/migrate.sql", Author: "ghost"},
		{Filename: "20241030/a/migrate.sql", Author: "ghost"},
		{Filename: "app/domain/star/star.go", Author: "real"},
		{Filename: "app/domain/star/star.go", Author: "real"},
	}

	scores, risks := CalcIndispensability(blame, folded, 0.8, 0.6)

	for _, r := range risks {
		if r.Module == PeripheralModule {
			t.Errorf("bus-factor risk should not include peripheral, got %+v", r)
		}
	}
	if scores["ghost"] != 0 {
		t.Errorf("owning only peripheral mass must not grant Indispensability; ghost=%v, want 0", scores["ghost"])
	}
	if scores["real"] <= 0 {
		t.Errorf("real declared-module owner should score > 0; got %v", scores["real"])
	}

	for _, o := range CalcOwnershipFragmentation(blame, folded) {
		if o.Module == PeripheralModule {
			t.Errorf("ownership fragmentation should not include peripheral, got %+v", o)
		}
	}
}
