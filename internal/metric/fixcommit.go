package metric

import (
	"regexp"
	"strings"

	"github.com/machuz/eis/v2/internal/git"
)

// fix-commit detection. Once the basis of the (removed) Quality axis; now used
// only to feed Debt Cleanup, which needs to know which commits are fixes.
var fixPattern = regexp.MustCompile(`(?i)^[^\w]*(?:\[?\s*(?:fix|revert|hotfix)\s*\]?[:/\s])`)

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
