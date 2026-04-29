package git

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/rand"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// defaultBlameTimeoutSec mirrors metric.CalcDebt's fallback when the caller
// passes 0 or a negative timeout — kept here so per-file blame timeouts in
// the worker pools have a sane upper bound even when cfg.BlameTimeout is
// not wired through.
const defaultBlameTimeoutSec = 120

// filterChunkSize bounds how many file paths get passed to a single
// `git ls-tree -l` call; argv length is OS-limited so we batch.
const filterChunkSize = 500

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

// SampleFiles performs stratified sampling by module to ensure every module
// is represented. Each module (first 3 directory components) gets at least
// minPerModule files, with remaining budget allocated proportionally.
func SampleFiles(files []string, maxFiles int) []string {
	if len(files) <= maxFiles {
		return files
	}

	rng := rand.New(rand.NewSource(42)) // deterministic sampling

	// Group files by module
	type group struct {
		module string
		files  []string
	}
	groupMap := make(map[string]*group)
	var moduleOrder []string
	for _, f := range files {
		mod := moduleOfPath(f)
		g, ok := groupMap[mod]
		if !ok {
			g = &group{module: mod}
			groupMap[mod] = g
			moduleOrder = append(moduleOrder, mod)
		}
		g.files = append(g.files, f)
	}

	// Sort modules for deterministic order
	sort.Strings(moduleOrder)

	// Shuffle files within each module
	for _, mod := range moduleOrder {
		g := groupMap[mod]
		rng.Shuffle(len(g.files), func(i, j int) {
			g.files[i], g.files[j] = g.files[j], g.files[i]
		})
	}

	const minPerModule = 2
	nModules := len(moduleOrder)

	// Phase 1: guaranteed minimum per module
	guaranteed := nModules * minPerModule
	if guaranteed > maxFiles {
		// More modules than budget allows at minPerModule; give 1 each then fill
		guaranteed = min(nModules, maxFiles)
	}

	result := make([]string, 0, maxFiles)
	taken := make(map[string]int, nModules)

	actualMin := minPerModule
	if nModules*minPerModule > maxFiles {
		actualMin = 1
	}

	for _, mod := range moduleOrder {
		g := groupMap[mod]
		n := min(actualMin, len(g.files))
		if len(result)+n > maxFiles {
			break
		}
		result = append(result, g.files[:n]...)
		taken[mod] = n
	}

	// Phase 2: proportional allocation of remaining budget
	remaining := maxFiles - len(result)
	if remaining > 0 {
		// Count files not yet taken
		totalUntaken := 0
		for _, mod := range moduleOrder {
			g := groupMap[mod]
			totalUntaken += len(g.files) - taken[mod]
		}

		// Allocate proportionally
		type allocation struct {
			mod   string
			alloc int
			frac  float64
		}
		allocs := make([]allocation, 0, nModules)
		allocated := 0
		for _, mod := range moduleOrder {
			g := groupMap[mod]
			untaken := len(g.files) - taken[mod]
			if untaken <= 0 {
				continue
			}
			proportion := float64(untaken) / float64(totalUntaken)
			share := proportion * float64(remaining)
			floor := int(share)
			if floor > untaken {
				floor = untaken
			}
			allocs = append(allocs, allocation{mod, floor, share - float64(floor)})
			allocated += floor
		}

		// Distribute leftover slots by largest fractional remainder
		leftover := remaining - allocated
		if leftover > 0 {
			sort.Slice(allocs, func(i, j int) bool {
				return allocs[i].frac > allocs[j].frac
			})
			for i := range allocs {
				if leftover <= 0 {
					break
				}
				g := groupMap[allocs[i].mod]
				untaken := len(g.files) - taken[allocs[i].mod]
				if allocs[i].alloc < untaken {
					allocs[i].alloc++
					leftover--
				}
			}
		}

		// Take allocated files from each module
		for _, a := range allocs {
			if a.alloc <= 0 {
				continue
			}
			g := groupMap[a.mod]
			start := taken[a.mod]
			end := start + a.alloc
			if end > len(g.files) {
				end = len(g.files)
			}
			result = append(result, g.files[start:end]...)
		}
	}

	return result
}

// moduleOfPath extracts a module identifier from a file path.
// Uses the first 3 directory components (mirrors metric.ModuleOf).
func moduleOfPath(path string) string {
	dir := filepath.Dir(path)
	parts := strings.Split(filepath.ToSlash(dir), "/")
	if len(parts) > 3 {
		parts = parts[:3]
	}
	return strings.Join(parts, "/")
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

	filter := NewFileFilter(filepath)
	filename := filepath
	filterFor := filepath

	var authors []string
	var pendingAuthor string
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "author "):
			pendingAuthor = strings.TrimPrefix(line, "author ")
		case strings.HasPrefix(line, "filename "):
			filename = strings.TrimPrefix(line, "filename ")
			if filename != filterFor {
				filter = NewFileFilter(filename)
				filterFor = filename
			}
		case strings.HasPrefix(line, "\t"):
			content := line[1:]
			if pendingAuthor != "" && !filter.IsSkip(content) {
				authors = append(authors, pendingAuthor)
			}
			pendingAuthor = ""
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

	filter := NewFileFilter(defaultFilename)
	filterFor := defaultFilename

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
			if filename != filterFor {
				filter = NewFileFilter(filename)
				filterFor = filename
			}
		case strings.HasPrefix(line, "\t"):
			// content line = end of block
			content := line[1:]
			if author != "" && !filter.IsSkip(content) {
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
	filter := NewFileFilter(filepath)
	filterFor := filepath

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), scannerMaxBuf)

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
			if filename != filterFor {
				filter = NewFileFilter(filename)
				filterFor = filename
			}
		case strings.HasPrefix(line, "\t"):
			content := line[1:]
			if author != "" && !filter.IsSkip(content) {
				result = append(result, BlameLine{
					Author:        author,
					CommitterTime: committerTime,
					Filename:      filename,
				})
			}
			author = ""
		}
	}

	// Drain anything the scanner skipped (bufio.ErrTooLong path) so git's
	// pipe write side doesn't block cmd.Wait below.
	_, _ = io.Copy(io.Discard, stdout)

	_ = cmd.Wait()
	return result, scanner.Err()
}

// ConcurrentBlameFiles runs blame on files concurrently with a worker pool.
// blameTimeoutSec bounds each per-file blame call; <= 0 falls back to
// defaultBlameTimeoutSec. Without this guard a single pathological file
// (huge single-line dumps, repo corruption) can stall the whole pool
// indefinitely.
func ConcurrentBlameFiles(ctx context.Context, repoPath string, files []string, maxFiles, workers int, blameTimeoutSec int, progressFn func(done, total int), verboseFn func(string)) ([]BlameLine, error) {
	sampled := SampleFiles(files, maxFiles)
	total := len(sampled)

	if workers <= 0 {
		workers = 4
	}

	timeoutSec := blameTimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = defaultBlameTimeoutSec
	}
	timeout := time.Duration(timeoutSec) * time.Second

	type result struct {
		file    string
		lines   []BlameLine
		err     error
		dur     time.Duration
		timeout bool
	}

	fileCh := make(chan string, total)
	resultCh := make(chan result, total)

	// Start workers
	for w := 0; w < workers; w++ {
		go func() {
			for f := range fileCh {
				start := time.Now()
				fileCtx, cancel := context.WithTimeout(ctx, timeout)
				lines, err := BlameFileStream(fileCtx, repoPath, f)
				timedOut := fileCtx.Err() == context.DeadlineExceeded
				cancel()
				resultCh <- result{f, lines, err, time.Since(start), timedOut}
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
		if r.err == nil && !r.timeout {
			allLines = append(allLines, r.lines...)
		}
		if verboseFn != nil && (r.dur > 2*time.Second || r.err != nil || r.timeout) {
			switch {
			case r.timeout:
				verboseFn(fmt.Sprintf("  [blame] %s: TIMEOUT (>%ds, skipped)", r.file, timeoutSec))
			case r.err != nil:
				verboseFn(fmt.Sprintf("  [blame] %s: error (%v)", r.file, r.err))
			default:
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
	filter := NewFileFilter(filepath)
	filterFor := filepath

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), scannerMaxBuf)

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
			if filename != filterFor {
				filter = NewFileFilter(filename)
				filterFor = filename
			}
		case strings.HasPrefix(line, "\t"):
			content := line[1:]
			if author != "" && !filter.IsSkip(content) {
				result = append(result, BlameLine{
					Author:        author,
					CommitterTime: committerTime,
					Filename:      filename,
				})
			}
			author = ""
		}
	}

	// Same drain-before-Wait fail-safe as BlameFileStream.
	_, _ = io.Copy(io.Discard, stdout)

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
// blameTimeoutSec bounds each per-file blame call; see ConcurrentBlameFiles
// for the rationale. Timeline boundary blame in particular needed this
// because the boundary tree can include checked-in dumps absent from HEAD.
func ConcurrentBlameFilesAtCommit(ctx context.Context, repoPath, commitHash string, files []string, maxFiles, workers int, blameTimeoutSec int, progressFn func(done, total int), verboseFn func(string)) ([]BlameLine, error) {
	sampled := SampleFiles(files, maxFiles)
	total := len(sampled)

	if workers <= 0 {
		workers = 4
	}

	timeoutSec := blameTimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = defaultBlameTimeoutSec
	}
	timeout := time.Duration(timeoutSec) * time.Second

	type result struct {
		file    string
		lines   []BlameLine
		err     error
		dur     time.Duration
		timeout bool
	}

	fileCh := make(chan string, total)
	resultCh := make(chan result, total)

	// Start workers
	for w := 0; w < workers; w++ {
		go func() {
			for f := range fileCh {
				start := time.Now()
				fileCtx, cancel := context.WithTimeout(ctx, timeout)
				lines, err := BlameFileAtCommit(fileCtx, repoPath, commitHash, f)
				timedOut := fileCtx.Err() == context.DeadlineExceeded
				cancel()
				resultCh <- result{f, lines, err, time.Since(start), timedOut}
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
		if r.err == nil && !r.timeout {
			allLines = append(allLines, r.lines...)
		}
		if verboseFn != nil && (r.dur > 2*time.Second || r.err != nil || r.timeout) {
			short := commitHash
			if len(short) > 8 {
				short = short[:8]
			}
			switch {
			case r.timeout:
				verboseFn(fmt.Sprintf("  [blame@%s] %s: TIMEOUT (>%ds, skipped)", short, r.file, timeoutSec))
			case r.err != nil:
				verboseFn(fmt.Sprintf("  [blame@%s] %s: error (%v)", short, r.file, r.err))
			default:
				verboseFn(fmt.Sprintf("  [blame@%s] %s: %d lines (SLOW: %v)", short, r.file, len(r.lines), r.dur.Round(time.Millisecond)))
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

// FilterFilesBySize returns files whose blob size at commitHash is <= maxBytes.
// maxBytes <= 0 disables filtering and returns the input unchanged. Files
// whose size cannot be determined (e.g. submodules, missing objects, or git
// errors) are kept — the goal is to avoid known-huge files, not to be
// strict about every edge case. verboseFn, when non-nil, is invoked once
// per skipped file so callers can surface what was dropped.
//
// commitHash == "" or "HEAD" both resolve against HEAD via `git ls-tree -l`.
// We chunk into filterChunkSize-sized argv batches so very wide file lists
// don't bump into the OS argv length limit.
func FilterFilesBySize(ctx context.Context, repoPath, commitHash string, files []string, maxBytes int64, verboseFn func(string)) ([]string, error) {
	if maxBytes <= 0 || len(files) == 0 {
		return files, nil
	}

	ref := commitHash
	if ref == "" {
		ref = "HEAD"
	}

	// Build size lookup: path -> bytes. Missing entries mean "unknown size";
	// such files pass the filter unchanged so we don't accidentally drop
	// real code because of an ls-tree quirk.
	sizes := make(map[string]int64, len(files))
	for start := 0; start < len(files); start += filterChunkSize {
		end := min(start+filterChunkSize, len(files))
		args := []string{"ls-tree", "-l", "-z", ref, "--"}
		args = append(args, files[start:end]...)

		// We use -z (NUL-terminated) so paths with spaces or tabs work.
		// RunLines is line-oriented (newline split), so go direct: read
		// all stdout, split on NUL.
		stdout, cmd, err := RunStream(ctx, repoPath, args...)
		if err != nil {
			return files, err
		}
		raw, _ := io.ReadAll(stdout)
		_ = stdout.Close()
		_ = cmd.Wait()

		for rec := range strings.SplitSeq(string(raw), "\x00") {
			if rec == "" {
				continue
			}
			// Format: "<mode> <type> <object> <size>\t<path>"
			meta, path, ok := strings.Cut(rec, "\t")
			if !ok {
				continue
			}
			fields := strings.Fields(meta)
			if len(fields) < 4 {
				continue
			}
			// fields[3] is size or "-" for non-blobs (trees/submodules)
			n, err := strconv.ParseInt(fields[3], 10, 64)
			if err != nil {
				continue
			}
			sizes[path] = n
		}
	}

	kept := make([]string, 0, len(files))
	for _, f := range files {
		size, ok := sizes[f]
		if !ok {
			// Unknown size — keep, callers shouldn't lose files just
			// because ls-tree didn't list them.
			kept = append(kept, f)
			continue
		}
		if size > maxBytes {
			if verboseFn != nil {
				verboseFn(fmt.Sprintf("[blame] skip large file: %s (%d bytes > %d limit)", f, size, maxBytes))
			}
			continue
		}
		kept = append(kept, f)
	}
	return kept, nil
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

// HeadHash returns the current HEAD commit hash for a repo.
func HeadHash(ctx context.Context, repoPath string) (string, error) {
	lines, err := RunLines(ctx, repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	if len(lines) == 0 {
		return "", fmt.Errorf("no HEAD found")
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
