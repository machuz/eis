package git

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// parallelLogMinCommits is the commit count below which ParseLogParallel just
// runs the serial ParseLog — for small repos the rev-list + fan-out overhead
// isn't worth it (and the serial path is already sub-second). var, not const,
// so tests can lower it to exercise the parallel path on a tiny fixture repo.
var parallelLogMinCommits = 4000

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
	commits, scanErr := parseLogStream(stdout)
	waitErr := cmd.Wait()
	if scanErr != nil {
		return commits, scanErr
	}
	return commits, waitErr
}

// parseLogStream parses the stdout of a `git log -p --numstat` invocation into
// commits with comment-filtered per-file line counts. Shared by the serial
// ParseLog and the per-chunk workers of ParseLogParallel.
func parseLogStream(r io.Reader) ([]Commit, error) {
	scanner := bufio.NewScanner(r)
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
	return commits, scanner.Err()
}

// ParseLogParallel is a drop-in faster ParseLog for large repos. The cost of
// ParseLog is git generating `-p` (full patch) output for every commit; on a
// 200k-commit repo that single stream is the dominant phase of analysis. This
// splits history into `workers` contiguous chunks (cheap rev-list first, which
// emits no diffs) and parses each chunk's `git log -p` concurrently, then
// concatenates in history order — so the patch generation and the Go-side
// comment filtering both fan out across cores. Output is identical to ParseLog:
// same commit set (--all --no-merges), same order, same per-file filtered counts.
//
// Small repos (or workers < 2) fall back to serial ParseLog.
func ParseLogParallel(ctx context.Context, repoPath string, workers int) ([]Commit, error) {
	if workers < 2 {
		return ParseLog(ctx, repoPath)
	}
	shas, err := revListAll(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	if len(shas) < parallelLogMinCommits {
		return ParseLog(ctx, repoPath)
	}

	chunks := chunkStrings(shas, workers)
	results := make([][]Commit, len(chunks))
	errs := make([]error, len(chunks))
	var wg sync.WaitGroup
	for i := range chunks {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = parseLogChunk(ctx, repoPath, chunks[i])
		}(i)
	}
	wg.Wait()

	total := 0
	for i, e := range errs {
		if e != nil {
			return nil, e
		}
		total += len(results[i])
	}
	out := make([]Commit, 0, total)
	for _, r := range results {
		out = append(out, r...)
	}
	return out, nil
}

// revListAll returns every non-merge commit SHA reachable from any ref, in git's
// default (reverse-chronological) order — the same set and order ParseLog walks,
// but with no patch generation, so it returns in ~a second even on huge repos.
func revListAll(ctx context.Context, repoPath string) ([]string, error) {
	stdout, cmd, err := RunStream(ctx, repoPath, "rev-list", "--all", "--no-merges")
	if err != nil {
		return nil, err
	}
	defer stdout.Close()
	var shas []string
	sc := bufio.NewScanner(stdout)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		if s := sc.Text(); s != "" {
			shas = append(shas, s)
		}
	}
	scanErr := sc.Err()
	waitErr := cmd.Wait()
	if scanErr != nil {
		return shas, scanErr
	}
	return shas, waitErr
}

// parseLogChunk runs `git log -p --numstat` over exactly the given commits
// (fed on stdin, shown in input order via --no-walk) and parses the stream. Each
// commit is still diffed against its first parent, identical to the serial walk.
func parseLogChunk(ctx context.Context, repoPath string, shas []string) ([]Commit, error) {
	cmd := exec.CommandContext(ctx, "git",
		"log", "--no-walk=unsorted", "--no-merges", "--no-color",
		"--format=COMMIT:%H|%an|%ai|%s",
		"--numstat", "-p", "--stdin",
	)
	cmd.Dir = repoPath
	cmd.Stdin = strings.NewReader(strings.Join(shas, "\n") + "\n")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	commits, scanErr := parseLogStream(stdout)
	waitErr := cmd.Wait()
	if scanErr != nil {
		return commits, scanErr
	}
	return commits, waitErr
}

// chunkStrings splits s into n contiguous, near-equal slices (preserving order).
func chunkStrings(s []string, n int) [][]string {
	if n < 1 {
		n = 1
	}
	if n > len(s) {
		n = len(s)
	}
	out := make([][]string, 0, n)
	base := len(s) / n
	rem := len(s) % n
	i := 0
	for c := 0; c < n; c++ {
		size := base
		if c < rem {
			size++
		}
		out = append(out, s[i:i+size])
		i += size
	}
	return out
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
