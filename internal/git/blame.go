package git

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type BlameLine struct {
	Author        string
	CommitterTime time.Time
	Filename      string
}

func ListFiles(ctx context.Context, repoPath string, patterns []string) ([]string, error) {
	args := []string{"ls-files", "--"}
	args = append(args, patterns...)

	lines, err := RunLines(ctx, repoPath, args...)
	if err != nil {
		return nil, err
	}
	return lines, nil
}

// ListAllFiles returns all tracked files in the repo (for domain auto-detection)
func ListAllFiles(ctx context.Context, repoPath string) ([]string, error) {
	lines, err := RunLines(ctx, repoPath, "ls-files")
	if err != nil {
		return nil, err
	}
	return lines, nil
}

func SampleFiles(files []string, maxFiles int) []string {
	if len(files) <= maxFiles {
		return files
	}

	rng := rand.New(rand.NewSource(42)) // deterministic sampling
	rng.Shuffle(len(files), func(i, j int) {
		files[i], files[j] = files[j], files[i]
	})
	return files[:maxFiles]
}

func BlameFile(ctx context.Context, repoPath, filepath string) ([]BlameLine, error) {
	lines, err := RunLines(ctx, repoPath,
		"blame", "--line-porcelain", "-w", filepath,
	)
	if err != nil {
		return nil, err
	}

	return parsePorcelainBlame(lines, filepath), nil
}

func BlameFileAtParent(ctx context.Context, repoPath, commitHash, filepath string) ([]string, error) {
	lines, err := RunLines(ctx, repoPath,
		"blame", "--line-porcelain", "-w", commitHash+"^", "--", filepath,
	)
	if err != nil {
		return nil, err
	}

	var authors []string
	for _, line := range lines {
		if strings.HasPrefix(line, "author ") {
			authors = append(authors, strings.TrimPrefix(line, "author "))
		}
	}
	return authors, nil
}

func DiffTreeFiles(ctx context.Context, repoPath, commitHash string) ([]string, error) {
	// Use raw format to distinguish blobs from submodule commits
	lines, err := RunLines(ctx, repoPath,
		"diff-tree", "--no-commit-id", "-r", commitHash,
	)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, line := range lines {
		// Raw format: ":oldmode newmode oldhash newhash status\tpath"
		// Submodules have mode 160000; skip them
		if strings.HasPrefix(line, ":") && strings.Contains(line, " 160000 ") {
			continue
		}
		// Extract path after tab
		if idx := strings.IndexByte(line, '\t'); idx >= 0 {
			files = append(files, line[idx+1:])
		}
	}
	return files, nil
}

func BlameFiles(ctx context.Context, repoPath string, files []string, maxFiles int, progressFn func(done, total int)) ([]BlameLine, error) {
	sampled := SampleFiles(files, maxFiles)
	total := len(sampled)

	var allLines []BlameLine
	for i, f := range sampled {
		blameLines, err := BlameFile(ctx, repoPath, f)
		if err != nil {
			continue // skip files that can't be blamed
		}
		allLines = append(allLines, blameLines...)

		if progressFn != nil && (i+1)%50 == 0 {
			progressFn(i+1, total)
		}
	}
	if progressFn != nil {
		progressFn(total, total)
	}

	return allLines, nil
}

func parsePorcelainBlame(lines []string, defaultFilename string) []BlameLine {
	var result []BlameLine
	var author string
	var committerTime time.Time
	filename := defaultFilename

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "author "):
			author = strings.TrimPrefix(line, "author ")
		case strings.HasPrefix(line, "committer-time "):
			ts, err := strconv.ParseInt(strings.TrimPrefix(line, "committer-time "), 10, 64)
			if err == nil {
				committerTime = time.Unix(ts, 0)
			}
		case strings.HasPrefix(line, "filename "):
			filename = strings.TrimPrefix(line, "filename ")
		case strings.HasPrefix(line, "\t"):
			// content line = end of block
			if author != "" {
				result = append(result, BlameLine{
					Author:        author,
					CommitterTime: committerTime,
					Filename:      filename,
				})
			}
			author = ""
		}
	}

	return result
}

// BlameFilesStream processes blame with a scanner for memory efficiency on large repos
func BlameFileStream(ctx context.Context, repoPath, filepath string) ([]BlameLine, error) {
	stdout, cmd, err := RunStream(ctx, repoPath,
		"blame", "--line-porcelain", "-w", filepath,
	)
	if err != nil {
		return nil, err
	}
	defer stdout.Close()

	var result []BlameLine
	var author string
	var committerTime time.Time
	filename := filepath

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "author "):
			author = strings.TrimPrefix(line, "author ")
		case strings.HasPrefix(line, "committer-time "):
			ts, err := strconv.ParseInt(strings.TrimPrefix(line, "committer-time "), 10, 64)
			if err == nil {
				committerTime = time.Unix(ts, 0)
			}
		case strings.HasPrefix(line, "filename "):
			filename = strings.TrimPrefix(line, "filename ")
		case strings.HasPrefix(line, "\t"):
			if author != "" {
				result = append(result, BlameLine{
					Author:        author,
					CommitterTime: committerTime,
					Filename:      filename,
				})
			}
			author = ""
		}
	}

	_ = cmd.Wait()
	return result, scanner.Err()
}

// ConcurrentBlameFiles runs blame on files concurrently with a worker pool
func ConcurrentBlameFiles(ctx context.Context, repoPath string, files []string, maxFiles, workers int, progressFn func(done, total int), verboseFn func(string)) ([]BlameLine, error) {
	sampled := SampleFiles(files, maxFiles)
	total := len(sampled)

	if workers <= 0 {
		workers = 4
	}

	type result struct {
		file  string
		lines []BlameLine
		err   error
		dur   time.Duration
	}

	fileCh := make(chan string, total)
	resultCh := make(chan result, total)

	// Start workers
	for w := 0; w < workers; w++ {
		go func() {
			for f := range fileCh {
				start := time.Now()
				lines, err := BlameFileStream(ctx, repoPath, f)
				resultCh <- result{f, lines, err, time.Since(start)}
			}
		}()
	}

	// Send files
	for _, f := range sampled {
		fileCh <- f
	}
	close(fileCh)

	// Collect results
	var allLines []BlameLine
	for i := 0; i < total; i++ {
		r := <-resultCh
		if r.err == nil {
			allLines = append(allLines, r.lines...)
		}
		if verboseFn != nil && (r.dur > 2*time.Second || r.err != nil) {
			if r.err != nil {
				verboseFn(fmt.Sprintf("  [blame] %s: error (%v)", r.file, r.err))
			} else {
				verboseFn(fmt.Sprintf("  [blame] %s: %d lines (SLOW: %v)", r.file, len(r.lines), r.dur.Round(time.Millisecond)))
			}
		}
		if progressFn != nil && (i+1)%50 == 0 {
			progressFn(i+1, total)
		}
	}
	if progressFn != nil {
		progressFn(total, total)
	}

	return allLines, nil
}

// BlameFileAtCommit runs blame at a specific commit hash.
func BlameFileAtCommit(ctx context.Context, repoPath, commitHash, filepath string) ([]BlameLine, error) {
	stdout, cmd, err := RunStream(ctx, repoPath,
		"blame", "--line-porcelain", "-w", commitHash, "--", filepath,
	)
	if err != nil {
		return nil, err
	}
	defer stdout.Close()

	var result []BlameLine
	var author string
	var committerTime time.Time
	filename := filepath

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "author "):
			author = strings.TrimPrefix(line, "author ")
		case strings.HasPrefix(line, "committer-time "):
			ts, err := strconv.ParseInt(strings.TrimPrefix(line, "committer-time "), 10, 64)
			if err == nil {
				committerTime = time.Unix(ts, 0)
			}
		case strings.HasPrefix(line, "filename "):
			filename = strings.TrimPrefix(line, "filename ")
		case strings.HasPrefix(line, "\t"):
			if author != "" {
				result = append(result, BlameLine{
					Author:        author,
					CommitterTime: committerTime,
					Filename:      filename,
				})
			}
			author = ""
		}
	}

	_ = cmd.Wait()
	return result, scanner.Err()
}

// ListFilesAtCommit returns tracked files at a specific commit hash, filtered by patterns.
func ListFilesAtCommit(ctx context.Context, repoPath, commitHash string, patterns []string) ([]string, error) {
	lines, err := RunLines(ctx, repoPath,
		"ls-tree", "-r", "--name-only", commitHash,
	)
	if err != nil {
		return nil, err
	}

	if len(patterns) == 0 {
		return lines, nil
	}

	// Filter by extension patterns (e.g. "*.go", "*.ts")
	var filtered []string
	for _, f := range lines {
		for _, p := range patterns {
			// Simple glob: *.ext
			if strings.HasPrefix(p, "*.") {
				ext := p[1:] // ".go"
				if strings.HasSuffix(f, ext) {
					filtered = append(filtered, f)
					break
				}
			} else if strings.Contains(f, p) {
				filtered = append(filtered, f)
				break
			}
		}
	}
	return filtered, nil
}

// ConcurrentBlameFilesAtCommit runs blame at a specific commit on files concurrently.
func ConcurrentBlameFilesAtCommit(ctx context.Context, repoPath, commitHash string, files []string, maxFiles, workers int, progressFn func(done, total int), verboseFn func(string)) ([]BlameLine, error) {
	sampled := SampleFiles(files, maxFiles)
	total := len(sampled)

	if workers <= 0 {
		workers = 4
	}

	type result struct {
		file  string
		lines []BlameLine
		err   error
		dur   time.Duration
	}

	fileCh := make(chan string, total)
	resultCh := make(chan result, total)

	// Start workers
	for w := 0; w < workers; w++ {
		go func() {
			for f := range fileCh {
				start := time.Now()
				lines, err := BlameFileAtCommit(ctx, repoPath, commitHash, f)
				resultCh <- result{f, lines, err, time.Since(start)}
			}
		}()
	}

	// Send files
	for _, f := range sampled {
		fileCh <- f
	}
	close(fileCh)

	// Collect results
	var allLines []BlameLine
	for i := 0; i < total; i++ {
		r := <-resultCh
		if r.err == nil {
			allLines = append(allLines, r.lines...)
		}
		if verboseFn != nil && (r.dur > 2*time.Second || r.err != nil) {
			if r.err != nil {
				verboseFn(fmt.Sprintf("  [blame@%s] %s: error (%v)", commitHash[:8], r.file, r.err))
			} else {
				verboseFn(fmt.Sprintf("  [blame@%s] %s: %d lines (SLOW: %v)", commitHash[:8], r.file, len(r.lines), r.dur.Round(time.Millisecond)))
			}
		}
		if progressFn != nil && (i+1)%50 == 0 {
			progressFn(i+1, total)
		}
	}
	if progressFn != nil {
		progressFn(total, total)
	}

	return allLines, nil
}

// FindCommitAtDate returns the latest commit hash on or before the given date.
func FindCommitAtDate(ctx context.Context, repoPath string, before time.Time) (string, error) {
	lines, err := RunLines(ctx, repoPath,
		"log", "--all", "--format=%H", "-1", "--before="+before.Format("2006-01-02T15:04:05"),
	)
	if err != nil {
		return "", err
	}
	if len(lines) == 0 {
		return "", fmt.Errorf("no commits found before %s", before.Format("2006-01-02"))
	}
	return lines[0], nil
}

// IsShallowRepo checks if the repository is a shallow clone
func IsShallowRepo(ctx context.Context, repoPath string) bool {
	lines, err := RunLines(ctx, repoPath, "rev-parse", "--is-shallow-repository")
	if err != nil || len(lines) == 0 {
		return false
	}
	return lines[0] == "true"
}

func init() {
	// Suppress unused import warning
	_ = fmt.Sprint
}
