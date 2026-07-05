package metric

import (
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

func TestMatchArchPattern_SegmentPrecise(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		filename string
		want     bool
	}{
		// Regression: a single-level directory pattern must NOT match a
		// same-named directory nested deeper (the Next.js route collision that
		// inflated Design — "app/(site)/stores/..." is a shop route, not a
		// state store).
		{"single-level dir rejects nested same-name", "*/stores/", "app/(site)/stores/[storeId]/page.tsx", false},
		{"single-level dir matches real store", "*/stores/", "app/stores/cart.ts", true},
		{"single-level dir rejects top-level", "*/stores/", "stores/cart.ts", false},

		// Opt-in any-depth with **.
		{"double-wild dir matches any depth", "**/stores/", "app/(site)/stores/[storeId]/page.tsx", true},
		{"double-wild dir matches shallow", "**/stores/", "app/stores/cart.ts", true},

		// File patterns: ** = any depth, * = one segment.
		{"double-wild file any depth", "**/router.go", "internal/app/router.go", true},
		{"double-wild file shallow", "**/router.go", "router.go", true},
		{"double-wild file no false suffix", "**/router.go", "internal/router_helper.go", false},
		{"single-wild file one segment only", "*/router.go", "internal/app/router.go", false},
		{"single-wild file matches one parent", "*/router.go", "app/router.go", true},

		// Intra-segment globs preserved.
		{"di glob any depth", "**/di/*.go", "app/di/wire.go", true},
		{"di glob rejects non-go", "**/di/*.go", "app/di/notes.md", false},
		{"repository interface", "**/repository/*interface*", "internal/domain/repository/user_repository_interface.go", true},
		{"repository interface no match", "**/repository/*interface*", "internal/domain/repository/user_repository.go", false},

		// Anchored (no wildcard) patterns are exact-depth.
		{"anchored di top-level", "di/*.go", "di/container.go", true},
		{"anchored di rejects nested", "di/*.go", "app/di/container.go", false},

		// Directory pattern requires a file underneath (not the dir itself).
		{"dir pattern needs file under", "*/stores/", "app/stores", false},

		// Trailing whitespace in a pattern must be trimmed, not split into a
		// stray empty segment that breaks matching.
		{"trailing space dir pattern", "*/stores/ ", "app/stores/cart.ts", true},
		{"trailing space file pattern", "**/router.go ", "internal/app/router.go", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchArchPattern(tt.filename, tt.pattern); got != tt.want {
				t.Errorf("matchArchPattern(%q, %q) = %v, want %v", tt.filename, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestCalcDesignSurviving(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tau := 365.0
	patterns := []string{"core/*", "packages/react/src/*"}
	lines := []git.BlameLine{
		// Alice: 2 surviving lines in arch files, recent (~full weight).
		{Author: "alice", Filename: "core/engine.go", CommitterTime: now.AddDate(0, 0, -10)},
		{Author: "alice", Filename: "packages/react/src/fiber.js", CommitterTime: now.AddDate(0, 0, -10)},
		// Alice also has a surviving NON-arch line — must NOT count toward design.
		{Author: "alice", Filename: "docs/readme.md", CommitterTime: now.AddDate(0, 0, -10)},
		// Bob: 1 surviving arch line but OLD (~1 tau → weight ~e^-1), so churned-away
		// arch volume decays out and can't out-score fresh durable ownership.
		{Author: "bob", Filename: "core/legacy.go", CommitterTime: now.AddDate(-1, 0, 0)},
	}
	got := CalcDesignSurviving(lines, patterns, tau, now)

	if len(got) != 2 {
		t.Fatalf("want 2 authors with arch survival, got %d: %v", len(got), got)
	}
	// Non-arch file excluded: alice = 2 arch lines only.
	if got["alice"] < 1.9 || got["alice"] > 2.01 {
		t.Errorf("alice design = %v, want ~2 (two fresh arch lines, doc excluded)", got["alice"])
	}
	// Time decay applied: bob's year-old line is worth ~e^-1 ≈ 0.37, not 1.
	if got["bob"] > 0.5 || got["bob"] < 0.2 {
		t.Errorf("bob design = %v, want ~0.37 (one ~1-tau-old arch line)", got["bob"])
	}
	// Durable fresh ownership outranks decayed old volume.
	if got["alice"] <= got["bob"] {
		t.Errorf("fresh durable arch owner (%.2f) must outscore decayed (%.2f)", got["alice"], got["bob"])
	}
}
