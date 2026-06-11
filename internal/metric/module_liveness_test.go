package metric

import (
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

// commit is a tiny helper to build a git.Commit whose Date carries year/month
// and whose FileStats reference the given filenames. The gate only reads
// commit.Date + FileStat.Filename, so insertions/deletions are arbitrary.
func livenessCommit(year int, month time.Month, files ...string) git.Commit {
	fs := make([]git.FileStat, 0, len(files))
	for _, f := range files {
		fs = append(fs, git.FileStat{Insertions: 1, Deletions: 0, Filename: f})
	}
	return git.Commit{
		Hash:      "h",
		Author:    "alice",
		Date:      time.Date(year, month, 1, 12, 0, 0, 0, time.UTC),
		Subject:   "s",
		FileStats: fs,
	}
}

func TestComputeModuleFold(t *testing.T) {
	// Default resolver: "**/domain/*" is a declared pattern, the date-dir
	// "20241030" resolves via the 2-component fallback.
	mr := NewModuleResolver(nil)

	tests := []struct {
		name      string
		commits   []git.Commit
		mr        ModuleResolver
		minMonths int
		// wantFold are module ids expected in the fold set; wantNotFold are
		// module ids that MUST NOT be in the fold set (asserted only when the
		// fold set is non-nil — absence is the success condition).
		wantFold    []string
		wantNotFold []string
		wantNil     bool
	}{
		{
			name: "fallback touched in one month folds (K=2)",
			commits: []git.Commit{
				livenessCommit(2024, time.October, "20241030/a/migrate.sql"),
			},
			mr:        mr,
			minMonths: 2,
			wantFold:  []string{"20241030/a"},
		},
		{
			name: "fallback touched in two distinct months does not fold",
			commits: []git.Commit{
				livenessCommit(2024, time.October, "20241030/a/migrate.sql"),
				livenessCommit(2024, time.November, "20241030/a/migrate.sql"),
			},
			mr:          mr,
			minMonths:   2,
			wantNotFold: []string{"20241030/a"},
			wantNil:     true, // nothing else folds either
		},
		{
			name: "declared module in one month never folds",
			commits: []git.Commit{
				livenessCommit(2024, time.October, "app/domain/star/star.go"),
			},
			mr:          mr,
			minMonths:   2,
			wantNotFold: []string{"app/domain/star"},
			wantNil:     true,
		},
		{
			name: "declared wins even when same module appears as fallback elsewhere",
			commits: []git.Commit{
				// Pattern-declared via **/domain/* -> "app/domain/star".
				livenessCommit(2024, time.October, "app/domain/star/star.go"),
				// A hypothetical single-month fallback path that resolves to the
				// SAME id would still never fold because declared wins. We use a
				// distinct single-month fallback to confirm only it folds.
				livenessCommit(2024, time.October, "20241030/a/migrate.sql"),
			},
			mr:          mr,
			minMonths:   2,
			wantFold:    []string{"20241030/a"},
			wantNotFold: []string{"app/domain/star"},
		},
		{
			name: "excluded module never appears in fold set",
			commits: []git.Commit{
				livenessCommit(2024, time.October, "20241030/a/migrate.sql"),
			},
			mr:          NewModuleResolverWithExcludes(nil, []string{"20*"}),
			minMonths:   2,
			wantNotFold: []string{"20241030/a"},
			wantNil:     true, // excluded -> "" -> ignored -> nothing folds
		},
		{
			name: "minMonths <= 1 disables gate (returns nil)",
			commits: []git.Commit{
				livenessCommit(2024, time.October, "20241030/a/migrate.sql"),
			},
			mr:        mr,
			minMonths: 1,
			wantNil:   true,
		},
		{
			name: "minMonths == 0 disables gate (returns nil)",
			commits: []git.Commit{
				livenessCommit(2024, time.October, "20241030/a/migrate.sql"),
			},
			mr:        mr,
			minMonths: 0,
			wantNil:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fold := ComputeModuleFold(tc.commits, tc.mr, tc.minMonths)

			if tc.wantNil && len(fold) != 0 {
				t.Fatalf("expected nil/empty fold set, got %v", fold)
			}
			for _, m := range tc.wantFold {
				if !fold[m] {
					t.Errorf("expected %q in fold set, got %v", m, fold)
				}
			}
			for _, m := range tc.wantNotFold {
				if fold[m] {
					t.Errorf("did not expect %q in fold set, got %v", m, fold)
				}
			}
		})
	}
}

func TestModuleOfWithFold(t *testing.T) {
	mr := NewModuleResolver(nil)

	commits := []git.Commit{
		// Single-month fallback -> folds.
		livenessCommit(2024, time.October, "20241030/a/migrate.sql"),
		// Two-month fallback -> survives.
		livenessCommit(2024, time.October, "tools/scripts/run.go"),
		livenessCommit(2024, time.November, "tools/scripts/run.go"),
		// Declared -> survives.
		livenessCommit(2024, time.October, "app/domain/star/star.go"),
	}
	fold := ComputeModuleFold(commits, mr, 2)
	folded := mr.WithFold(fold)

	// A folded path returns the sentinel.
	if got := folded.ModuleOf("20241030/a/migrate.sql"); got != PeripheralModule {
		t.Errorf("folded path: got %q, want %q", got, PeripheralModule)
	}
	// A surviving fallback path returns its own module id.
	if got := folded.ModuleOf("tools/scripts/run.go"); got != "tools/scripts" {
		t.Errorf("surviving fallback: got %q, want %q", got, "tools/scripts")
	}
	// A declared path returns its real module id (never folded).
	if got := folded.ModuleOf("app/domain/star/star.go"); got != "app/domain/star" {
		t.Errorf("declared path: got %q, want %q", got, "app/domain/star")
	}

	// WithFold(nil) disables folding: the previously-folded path returns its
	// own id again (value semantics — original resolver is untouched).
	if got := mr.ModuleOf("20241030/a/migrate.sql"); got != "20241030/a" {
		t.Errorf("unfolded resolver: got %q, want %q", got, "20241030/a")
	}

	// ResolveWithOrigin never applies fold and exposes declared-ness.
	if m, declared := folded.ResolveWithOrigin("app/domain/star/star.go"); m != "app/domain/star" || !declared {
		t.Errorf("ResolveWithOrigin declared: got (%q,%v), want (app/domain/star,true)", m, declared)
	}
	if m, declared := folded.ResolveWithOrigin("20241030/a/migrate.sql"); m != "20241030/a" || declared {
		t.Errorf("ResolveWithOrigin fallback: got (%q,%v), want (20241030/a,false)", m, declared)
	}
}

func TestModuleFoldIntegration_SurvivalLandsUnderPeripheral(t *testing.T) {
	mr := NewModuleResolver(nil)

	now := time.Date(2024, time.December, 1, 0, 0, 0, 0, time.UTC)

	// "20241030/a" is touched in a single month -> folds into peripheral.
	commits := []git.Commit{
		livenessCommit(2024, time.October, "20241030/a/migrate.sql"),
	}
	fold := ComputeModuleFold(commits, mr, 2)
	mr = mr.WithFold(fold)

	blameLines := []git.BlameLine{
		{Author: "alice", CommitterTime: now.AddDate(0, 0, -10), Filename: "20241030/a/migrate.sql"},
		{Author: "bob", CommitterTime: now.AddDate(0, 0, -20), Filename: "20241030/a/seed.sql"},
	}

	got := CalcModuleSurvivalByAuthor(blameLines, 180, now, mr)

	if _, ok := got["20241030/a"]; ok {
		t.Errorf("folded module %q should not appear as its own bucket; got %v", "20241030/a", got)
	}
	peripheral, ok := got[PeripheralModule]
	if !ok {
		t.Fatalf("expected mass under %q, got %v", PeripheralModule, got)
	}
	if peripheral["alice"] <= 0 || peripheral["bob"] <= 0 {
		t.Errorf("expected both authors' survival mass under %q, got %v", PeripheralModule, peripheral)
	}
}
