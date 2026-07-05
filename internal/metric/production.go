package metric

import (
	"path/filepath"
	"strings"

	"github.com/machuz/eis/v2/internal/git"
)

func CalcProduction(commits []git.Commit, excludePatterns []string) map[string]float64 {
	result := make(map[string]float64)

	for _, c := range commits {
		for _, fs := range c.FileStats {
			if IsExcluded(fs.Filename, excludePatterns) {
				continue
			}
			result[c.Author] += float64(fs.Insertions + fs.Deletions)
		}
	}

	return result
}

// CalcLines returns per-author total lines added and deleted (excluding excluded patterns).
func CalcLines(commits []git.Commit, excludePatterns []string) (added map[string]int, deleted map[string]int) {
	added = make(map[string]int)
	deleted = make(map[string]int)

	for _, c := range commits {
		for _, fs := range c.FileStats {
			if IsExcluded(fs.Filename, excludePatterns) {
				continue
			}
			added[c.Author] += fs.Insertions
			deleted[c.Author] += fs.Deletions
		}
	}

	return added, deleted
}

// IsExcluded checks if a filename matches any of the exclude patterns.
func IsExcluded(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, _ := filepath.Match(pattern, filename)
		if matched {
			return true
		}
		// Also check basename
		base := filepath.Base(filename)
		matched, _ = filepath.Match(pattern, base)
		if matched {
			return true
		}
	}
	return false
}

// EscapeGlob quotes filepath.Match metacharacters so a literal path becomes a
// pattern matching only itself. Used to fold exact .gitattributes-excluded paths
// (linguist-generated / linguist-vendored) into the pattern-based exclusion that
// IsExcluded drives, without a path that happens to contain *, ?, [ or \ turning
// into an over-broad glob.
func EscapeGlob(path string) string {
	if !strings.ContainsAny(path, `*?[\`) {
		return path
	}
	var b strings.Builder
	b.Grow(len(path) + 4)
	for _, r := range path {
		switch r {
		case '*', '?', '[', '\\':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}
