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
			authors := CommitAuthors(c)
			v := float64(archLines) / float64(len(authors))
			for _, a := range authors {
				result[a] += v
			}
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

// matchArchPattern reports whether a file path counts as an architecture file
// under the given glob pattern. It uses the same `*` / `**` convention as the
// module resolver (compilePattern in module.go): `*` matches exactly ONE path
// segment, `**` matches any depth (zero or more consecutive segments), and a
// literal segment may carry intra-segment globs like "*.go" or "*interface*"
// (resolved with filepath.Match).
//
// A pattern ending in "/" is a DIRECTORY pattern: a file matches when the
// pattern matches a leading prefix of the file's path (the file lives somewhere
// under that directory). Otherwise it is a FILE pattern that must match the
// whole path.
//
// This deliberately replaces the old substring matching, under which "*/stores/"
// matched a `stores` directory at ANY depth and so miscounted framework route
// directories (e.g. Next.js "app/(site)/stores/[id]/page.tsx") as architecture.
// Segment-precise, "*/stores/" matches "app/stores/..." but not
// "app/(site)/stores/..."; use "**/stores/" to opt back into any-depth.
func matchArchPattern(filename, pattern string) bool {
	pattern = strings.TrimSpace(pattern)
	isDir := strings.HasSuffix(pattern, "/")
	pat, ok := compilePattern(pattern)
	if !ok {
		return false
	}
	parts := strings.Split(filepath.ToSlash(filename), "/")
	if isDir {
		// File lives under the matched directory: the pattern matches a leading
		// prefix and at least one segment remains (the file itself).
		n, matched := matchArchPrefix(pat.components, parts)
		return matched && n >= 1 && n < len(parts)
	}
	return matchArchFull(pat.components, parts)
}

// archSegMatch matches one non-`**` pattern component against a single path
// segment. Wildcard (`*`) matches any one segment; a literal is matched with
// filepath.Match so intra-segment globs ("*.go", "*interface*") work, falling
// back to exact equality if the pattern is malformed.
func archSegMatch(c patternComponent, part string) bool {
	if c.Wildcard {
		return true
	}
	if ok, err := filepath.Match(c.Literal, part); err == nil {
		return ok
	}
	return c.Literal == part
}

// matchArchFull reports whether comps match the entire parts slice. `**`
// backtracks over the shortest run that lets the rest of the pattern match.
func matchArchFull(comps []patternComponent, parts []string) bool {
	if len(comps) == 0 {
		return len(parts) == 0
	}
	if comps[0].DoubleWild {
		for skip := 0; skip <= len(parts); skip++ {
			if matchArchFull(comps[1:], parts[skip:]) {
				return true
			}
		}
		return false
	}
	if len(parts) == 0 {
		return false
	}
	if !archSegMatch(comps[0], parts[0]) {
		return false
	}
	return matchArchFull(comps[1:], parts[1:])
}

// matchArchPrefix reports how many leading parts comps consume (a prefix match),
// mirroring matchPrefix in module.go but with intra-segment globbing.
func matchArchPrefix(comps []patternComponent, parts []string) (int, bool) {
	if len(comps) == 0 {
		return 0, true
	}
	if comps[0].DoubleWild {
		for skip := 0; skip <= len(parts); skip++ {
			if n, ok := matchArchPrefix(comps[1:], parts[skip:]); ok {
				return skip + n, true
			}
		}
		return 0, false
	}
	if len(parts) == 0 {
		return 0, false
	}
	if !archSegMatch(comps[0], parts[0]) {
		return 0, false
	}
	n, ok := matchArchPrefix(comps[1:], parts[1:])
	if !ok {
		return 0, false
	}
	return 1 + n, true
}
