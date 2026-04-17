package git

import (
	"bufio"
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

// FileStat holds contribution-eligible line counts for a file in a commit.
// Code files exclude comment-only and blank lines (gaming protection);
// prose files (.md/.txt/etc.) and unknown types count every line.
type FileStat struct {
	Insertions int
	Deletions  int
	Filename   string
}

// ParseLog returns non-merge commits with per-file line stats. Uses
// `-p --numstat` to get both a filename manifest and diff hunks; comment/blank
// lines in code files are filtered out via FileFilter so that comment spam
// cannot inflate Production, Design, or Debt metrics.
func ParseLog(ctx context.Context, repoPath string) ([]Commit, error) {
	stdout, cmd, err := RunStream(ctx, repoPath,
		"log", "--all", "--no-merges", "--no-color",
		"--format=COMMIT:%H|%an|%ai|%s",
		"--numstat", "-p",
	)
	if err != nil {
		return nil, err
	}
	defer stdout.Close()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)

	var commits []Commit
	var current *Commit

	var inDiff, sawHunk bool
	var curFileName string
	var filter *FileFilter
	var fIns, fDel int

	flushFile := func() {
		defer func() {
			inDiff, sawHunk = false, false
			curFileName = ""
			filter = nil
			fIns, fDel = 0, 0
		}()
		if current == nil || !inDiff {
			return
		}
		// Prose/unknown files (filter == nil) keep numstat counts untouched.
		// Code files with at least one hunk get their counts replaced by filtered counts.
		if filter == nil || !sawHunk {
			return
		}
		for i := range current.FileStats {
			if current.FileStats[i].Filename == curFileName {
				current.FileStats[i].Insertions = fIns
				current.FileStats[i].Deletions = fDel
				break
			}
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "COMMIT:") {
			flushFile()
			if current != nil {
				commits = append(commits, *current)
			}
			parts := strings.SplitN(line[7:], "|", 4)
			if len(parts) < 4 {
				current = nil
				continue
			}
			date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[2])
			current = &Commit{
				Hash:    parts[0],
				Author:  parts[1],
				Date:    date,
				Subject: parts[3],
			}
			continue
		}

		if current == nil {
			continue
		}

		if strings.HasPrefix(line, "diff --git ") {
			flushFile()
			// `diff --git a/<path> b/<path>` — a filename may itself contain " b/",
			// so anchor on the LAST occurrence to find the new-side separator.
			if idx := strings.LastIndex(line, " b/"); idx > 0 {
				curFileName = line[idx+3:]
				filter = NewFileFilter(curFileName)
				inDiff = true
			}
			continue
		}

		if inDiff {
			if !sawHunk {
				if strings.HasPrefix(line, "+++ b/") {
					newName := strings.TrimPrefix(line, "+++ b/")
					if newName != "" && newName != "/dev/null" {
						curFileName = newName
						filter = NewFileFilter(curFileName)
					}
					continue
				}
				if strings.HasPrefix(line, "@@") {
					sawHunk = true
					// Block-comment state doesn't carry across hunks.
					filter = NewFileFilter(curFileName)
					continue
				}
				continue
			}
			if strings.HasPrefix(line, "@@") {
				filter = NewFileFilter(curFileName)
				continue
			}
			if line == "" {
				continue
			}
			switch line[0] {
			case '+':
				if !filter.IsSkip(line[1:]) {
					fIns++
				}
			case '-':
				if !filter.IsSkip(line[1:]) {
					fDel++
				}
			}
			continue
		}

		// Numstat region: insertions\tdeletions\tfilename
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}
		ins, _ := strconv.Atoi(parts[0])
		del, _ := strconv.Atoi(parts[1])
		current.FileStats = append(current.FileStats, FileStat{
			Insertions: ins,
			Deletions:  del,
			Filename:   resolveRenamePath(parts[2]),
		})
	}

	flushFile()
	if current != nil {
		commits = append(commits, *current)
	}

	scanErr := scanner.Err()
	waitErr := cmd.Wait()
	if scanErr != nil {
		return commits, scanErr
	}
	if waitErr != nil {
		return commits, waitErr
	}
	return commits, nil
}

// resolveRenamePath converts a git-numstat path that may embed rename syntax
// into the new-side path used by the diff's `+++ b/<path>` header.
//
//	"dir/{old.go => new.go}"       → "dir/new.go"
//	"{old_dir => new_dir}/file.go" → "new_dir/file.go"
//	"old/path.go => new/path.go"   → "new/path.go"
//	"regular/file.go"              → "regular/file.go"
//
// Keeping FileStat.Filename normalized to the new path lets downstream
// matching (comment filter, exclude patterns, arch-file detection) work
// without each caller re-parsing the rename syntax.
func resolveRenamePath(p string) string {
	if !strings.Contains(p, " => ") {
		return p
	}
	if i := strings.IndexByte(p, '{'); i >= 0 {
		if j := strings.IndexByte(p[i:], '}'); j > 0 {
			inner := p[i+1 : i+j]
			if sep := strings.Index(inner, " => "); sep >= 0 {
				joined := p[:i] + inner[sep+4:] + p[i+j+1:]
				// `src/{foo => }/bar` → `src//bar`; collapse doubled slashes.
				for strings.Contains(joined, "//") {
					joined = strings.ReplaceAll(joined, "//", "/")
				}
				return strings.TrimSuffix(joined, "/")
			}
		}
	}
	if sep := strings.Index(p, " => "); sep >= 0 {
		return p[sep+4:]
	}
	return p
}

// ParseMergeCommits returns merge-only commits (no file stats).
// Used to detect fix/revert subjects in merge commit messages.
func ParseMergeCommits(ctx context.Context, repoPath string) ([]Commit, error) {
	lines, err := RunLines(ctx, repoPath,
		"log", "--all", "--merges",
		"--format=COMMIT:%H|%an|%ai|%s",
	)
	if err != nil {
		return nil, err
	}

	var commits []Commit
	for _, line := range lines {
		if !strings.HasPrefix(line, "COMMIT:") {
			continue
		}
		parts := strings.SplitN(line[7:], "|", 4)
		if len(parts) < 4 {
			continue
		}
		date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[2])
		commits = append(commits, Commit{
			Hash:    parts[0],
			Author:  parts[1],
			Date:    date,
			Subject: parts[3],
			IsMerge: true,
		})
	}

	return commits, nil
}
