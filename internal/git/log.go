package git

import (
	"context"
	"strconv"
	"strings"
	"time"
)

type Commit struct {
	Hash      string
	Author    string
	Date      time.Time
	Subject   string
	IsMerge   bool
	FileStats []FileStat
}

type FileStat struct {
	Insertions int
	Deletions  int
	Filename   string
}

func ParseLog(ctx context.Context, repoPath string) ([]Commit, error) {
	// Include merge commits (no --no-merges) so fix detection works on merge subjects.
	// %P = parent hashes; merge commits have 2+ parents separated by spaces.
	lines, err := RunLines(ctx, repoPath,
		"log", "--all",
		"--format=COMMIT:%H|%P|%an|%ai|%s",
		"--numstat",
	)
	if err != nil {
		return nil, err
	}

	var commits []Commit
	var current *Commit

	for _, line := range lines {
		if strings.HasPrefix(line, "COMMIT:") {
			if current != nil {
				commits = append(commits, *current)
			}
			parts := strings.SplitN(line[7:], "|", 5)
			if len(parts) < 5 {
				continue
			}
			parents := parts[1]
			isMerge := strings.Contains(parents, " ")
			date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])
			current = &Commit{
				Hash:    parts[0],
				Author:  parts[2],
				Date:    date,
				Subject: parts[4],
				IsMerge: isMerge,
			}
			continue
		}

		if current == nil || strings.TrimSpace(line) == "" {
			continue
		}

		// numstat line: insertions\tdeletions\tfilename
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}

		ins, _ := strconv.Atoi(parts[0])
		del, _ := strconv.Atoi(parts[1])
		current.FileStats = append(current.FileStats, FileStat{
			Insertions: ins,
			Deletions:  del,
			Filename:   parts[2],
		})
	}

	if current != nil {
		commits = append(commits, *current)
	}

	return commits, nil
}
