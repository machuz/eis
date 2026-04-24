package metric

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/machuz/eis/v2/internal/git"
)

// ChangePressure maps module path → pressure value (commits / blame lines).
type ChangePressure map[string]float64

// ModuleOf extracts a module identifier from a file path.
// Uses the first 3 directory components (e.g. "app/domain/payment").
// For shallower paths, uses the parent directory.
func ModuleOf(path string) string {
	dir := filepath.Dir(path)
	parts := strings.Split(filepath.ToSlash(dir), "/")
	if len(parts) > 3 {
		parts = parts[:3]
	}
	return strings.Join(parts, "/")
}

// CalcChangePressure computes per-module change pressure.
// pressure = commits_touching_module / blame_lines_in_module
func CalcChangePressure(commits []git.Commit, blameLines []git.BlameLine) ChangePressure {
	// Count commits per module (a commit counts once per module it touches)
	moduleCommits := make(map[string]int)
	for _, c := range commits {
		touched := make(map[string]bool)
		for _, fs := range c.FileStats {
			mod := ModuleOf(fs.Filename)
			touched[mod] = true
		}
		for mod := range touched {
			moduleCommits[mod]++
		}
	}

	// Count blame lines per module
	moduleBlameLines := make(map[string]int)
	for _, bl := range blameLines {
		mod := ModuleOf(bl.Filename)
		moduleBlameLines[mod]++
	}

	// Calculate pressure
	pressure := make(ChangePressure)
	for mod, commits := range moduleCommits {
		lines := moduleBlameLines[mod]
		if lines > 0 {
			pressure[mod] = float64(commits) / float64(lines)
		} else {
			// Module has commits but no surviving blame lines — high churn
			pressure[mod] = float64(commits)
		}
	}

	// Also include modules with blame lines but no recent commits (pressure = 0)
	for mod := range moduleBlameLines {
		if _, ok := pressure[mod]; !ok {
			pressure[mod] = 0
		}
	}

	return pressure
}

// MedianPressure returns the median pressure value across all modules.
func (cp ChangePressure) MedianPressure() float64 {
	if len(cp) == 0 {
		return 0
	}
	vals := make([]float64, 0, len(cp))
	for _, v := range cp {
		vals = append(vals, v)
	}
	sort.Float64s(vals)

	n := len(vals)
	if n%2 == 0 {
		return (vals[n/2-1] + vals[n/2]) / 2
	}
	return vals[n/2]
}
