package metric

import "testing"

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
