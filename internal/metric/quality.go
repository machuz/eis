package metric

import (
	"regexp"
	"strings"

	"github.com/machuz/eis/v2/internal/git"
)

var fixPattern = regexp.MustCompile(`(?i)^[^\w]*(?:\[?\s*(?:fix|revert|hotfix)\s*\]?[:/\s])`)

func CalcQuality(commits []git.Commit) map[string]float64 {
	type counts struct {
		total int
		fixes int
	}

	authorCounts := make(map[string]*counts)

	for _, c := range commits {
		ac, ok := authorCounts[c.Author]
		if !ok {
			ac = &counts{}
			authorCounts[c.Author] = ac
		}
		// Merge commits contribute fixes but NOT to total count,
		// so they don't inflate Quality by diluting fix ratio.
		if !c.IsMerge {
			ac.total++
		}

		if isFixCommit(c.Subject) {
			ac.fixes++
		}
	}

	result := make(map[string]float64)
	for author, ac := range authorCounts {
		if ac.total == 0 {
			result[author] = 100
			continue
		}
		fixRatio := float64(ac.fixes) / float64(ac.total) * 100
		result[author] = 100 - fixRatio
	}

	return result
}

func isFixCommit(subject string) bool {
	if fixPattern.MatchString(subject) {
		return true
	}
	if strings.Contains(subject, "修正") {
		return true
	}
	return false
}

func GetFixCommits(commits []git.Commit) []git.Commit {
	var fixes []git.Commit
	for _, c := range commits {
		if isFixCommit(c.Subject) {
			fixes = append(fixes, c)
		}
	}
	return fixes
}
