package metric

import (
	"path/filepath"
	"strings"

	"github.com/machuz/eis/v2/internal/git"
)

func CalcProduction(commits []git.Commit, excludePatterns []string) map[string]float64 {
	result := make(map[string]float64)

	for _, c := range commits {
		authors := CommitAuthors(c)
		share := 1.0 / float64(len(authors))
		for _, fs := range c.FileStats {
			if IsExcluded(fs.Filename, excludePatterns) {
				continue
			}
			v := float64(fs.Insertions+fs.Deletions) * share
			for _, a := range authors {
				result[a] += v
			}
		}
	}

	return result
}

// CommitAuthors returns a commit's contributing authors — the Author plus any
// Co-authored-by names, deduped and Author-first (the Author sometimes lists
// itself as a co-author). Length ≥ 1. Callers split a commit's contribution
// equally across this set so squash-merges and pairing credit the real authors.
func CommitAuthors(c git.Commit) []string {
	if len(c.CoAuthors) == 0 {
		return []string{c.Author}
	}
	out := make([]string, 0, 1+len(c.CoAuthors))
	seen := make(map[string]bool, 1+len(c.CoAuthors))
	add := func(a string) {
		if a != "" && !seen[a] {
			seen[a] = true
			out = append(out, a)
		}
	}
	add(c.Author)
	for _, a := range c.CoAuthors {
		add(a)
	}
	if len(out) == 0 {
		return []string{c.Author}
	}
	return out
}

// blameShares invokes fn(author, share) for each contributor of a blame line's
// commit: 1/N for each of N co-authors, or 1.0 for the common solo author. Lets
// every blame-derived metric split co-authored lines the same way survival does.
func blameShares(bl git.BlameLine, fn func(author string, share float64)) {
	if len(bl.Authors) == 0 {
		fn(bl.Author, 1.0)
		return
	}
	share := 1.0 / float64(len(bl.Authors))
	for _, a := range bl.Authors {
		fn(a, share)
	}
}

// CoAuthorMap maps commit SHA → its full contributor set, for the commits that
// HAVE co-authors only (sparse — most commits are solo, so the map stays small
// even on huge repos). The pipeline attaches these sets to blame lines by SHA so
// survival is split across the real authors.
func CoAuthorMap(commits []git.Commit) map[string][]string {
	m := make(map[string][]string)
	for _, c := range commits {
		if len(c.CoAuthors) == 0 {
			continue
		}
		m[c.Hash] = CommitAuthors(c)
	}
	return m
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
		// Directory pattern (trailing "/"): exclude the whole tree under a
		// directory of this name AT ANY DEPTH. filepath.Match can't express this —
		// "*" never crosses "/", so "vendor/*" misses vendor/github.com/x/y.go and
		// "vendor/**" isn't a Match glob at all. Vendored/generated TREES
		// (vendor/, third_party/, node_modules/) are the dominant source of both
		// score inflation (someone gets credited for third-party code) and blame
		// RSS blow-up, so they need an explicit path-segment form. "vendor/"
		// matches vendor/… at the root AND services/x/vendor/… nested, but NOT a
		// file named vendor.go, nor a myvendor/ or vendored/ directory.
		if strings.HasSuffix(pattern, "/") {
			dir := strings.TrimSuffix(pattern, "/")
			if filename == dir ||
				strings.HasPrefix(filename, dir+"/") ||
				strings.Contains(filename, "/"+dir+"/") {
				return true
			}
			continue
		}
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
