package metric

import "testing"

// TestIsExcludedDirectoryPatterns covers the trailing-"/" directory form added
// to exclude vendored/generated TREES (vendor/, third_party/, node_modules/) at
// any depth — the dominant source of score inflation and blame-RSS blow-up that
// filepath.Match ("*" never crosses "/") cannot express.
func TestIsExcludedDirectoryPatterns(t *testing.T) {
	pats := []string{"vendor/", "third_party/", "node_modules/"}

	excluded := []string{
		"vendor/github.com/pkg/errors/errors.go", // root vendor tree, nested deep
		"vendor/foo.go",                          // direct child
		"services/loki/vendor/x/y.go",            // vendor nested under a service
		"third_party/protobuf/x.cc",
		"web/node_modules/react/index.js",
	}
	for _, f := range excluded {
		if !IsExcluded(f, pats) {
			t.Errorf("expected %q to be excluded by a directory pattern", f)
		}
	}

	// Must NOT over-match: a file merely named vendor.go, or a directory whose
	// name only contains "vendor" as a substring, is real authored code.
	kept := []string{
		"vendor.go",     // a file, not the vendor tree
		"src/vendor.go", // file named vendor.go inside a real dir
		"myvendor/x.go", // dir name contains but isn't "vendor"
		"vendored/x.go", // "vendored" != "vendor"
		"internal/nodemodules.go",
		"cmd/main.go",
	}
	for _, f := range kept {
		if IsExcluded(f, pats) {
			t.Errorf("expected %q to be KEPT (false-positive exclusion)", f)
		}
	}
}

// TestIsExcludedGeneratedBasenames covers the generated-code / lockfile basename
// globs added alongside the directory patterns.
func TestIsExcludedGeneratedBasenames(t *testing.T) {
	pats := []string{"*.pb.go", "*_pb2.py", "*.min.js", "Cargo.lock", "Gemfile.lock"}

	excluded := []string{
		"api/user.pb.go", // generated protobuf, nested
		"gen/schema_pb2.py",
		"web/dist/app.min.js",
		"Cargo.lock",
		"crates/core/Cargo.lock", // lockfile nested in a workspace
		"Gemfile.lock",
	}
	for _, f := range excluded {
		if !IsExcluded(f, pats) {
			t.Errorf("expected %q to be excluded by a generated/lock pattern", f)
		}
	}

	kept := []string{
		"api/user.go",    // hand-written, not *.pb.go
		"web/app.js",     // not minified
		"internal/pb.go", // "pb.go" but not "*.pb.go" (basename is "pb.go", pattern needs <x>.pb.go)
		"Cargo.toml",     // manifest, not the lock
	}
	for _, f := range kept {
		if IsExcluded(f, pats) {
			t.Errorf("expected %q to be KEPT (false-positive exclusion)", f)
		}
	}
}
