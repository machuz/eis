package metric

import (
	"path/filepath"

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
