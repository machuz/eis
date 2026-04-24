package metric

import (
	"path/filepath"
	"strings"

	"github.com/machuz/eis/v2/internal/git"
)

// CalcDesign scores architecture contributions weighted by lines changed.
// A commit touching architecture files scores the sum of (insertions + deletions)
// for those architecture files, giving more credit to substantial design work.
func CalcDesign(commits []git.Commit, archPatterns []string) map[string]float64 {
	result := make(map[string]float64)

	for _, c := range commits {
		archLines := archLinesChanged(c.FileStats, archPatterns)
		if archLines > 0 {
			result[c.Author] += float64(archLines)
		}
	}

	return result
}

// archLinesChanged returns total lines changed in architecture files for a commit.
func archLinesChanged(files []git.FileStat, patterns []string) int {
	total := 0
	for _, fs := range files {
		normalized := filepath.ToSlash(fs.Filename)
		for _, pattern := range patterns {
			if matchArchPattern(normalized, pattern) {
				total += fs.Insertions + fs.Deletions
				break
			}
		}
	}
	return total
}

func matchArchPattern(filename, pattern string) bool {
	// Directory pattern: "*/domainservice/" or "di/*.go"
	// Strip leading */ for substring matching
	cleanPattern := strings.TrimPrefix(pattern, "*/")

	// Directory pattern ending with /
	if strings.HasSuffix(cleanPattern, "/") {
		dir := strings.TrimSuffix(cleanPattern, "/")
		return strings.Contains(filename, "/"+dir+"/") || strings.HasPrefix(filename, dir+"/")
	}

	// Glob pattern like "di/*.go" or "*interface*"
	// Try matching against each path segment and also the full path
	if strings.Contains(cleanPattern, "/") {
		// Pattern with directory: match against full path
		matched, _ := filepath.Match(cleanPattern, filename)
		if matched {
			return true
		}
		// Try suffix matching: "di/*.go" should match "app/di/container.go"
		parts := strings.Split(filename, "/")
		for i := range parts {
			suffix := strings.Join(parts[i:], "/")
			matched, _ := filepath.Match(cleanPattern, suffix)
			if matched {
				return true
			}
		}
		return false
	}

	// Simple pattern without directory: match against each path component
	// e.g., "*interface*" should match "repository/user_repository_interface.go"
	parts := strings.Split(filename, "/")
	for _, part := range parts {
		matched, _ := filepath.Match(cleanPattern, part)
		if matched {
			return true
		}
	}

	// Also check if the pattern matches as a substring in path components
	if strings.Contains(cleanPattern, "*") {
		return false // already tried glob matching above
	}
	return strings.Contains(filename, cleanPattern)
}
