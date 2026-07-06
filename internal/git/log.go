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

// maxFilteredDiffLines caps how many diff lines of a single file in a single
// commit get per-line comment filtering. A commit that dumps a huge generated or
// vendored file produces a diff with hundreds of thousands of lines; filtering
// each one (the dominant cost of ParseLog on large repos) is pointless for such
// files. Past the cap we stop filtering that file and keep its raw numstat
// counts — comment ratio is negligible at that scale anyway.
const maxFilteredDiffLines = 50000

// parallelLogMinCommits is the commit count below which ParseLogParallel just
// runs the serial ParseLog — for small repos the rev-list + fan-out overhead
// isn't worth it (and the serial path is already sub-second). var, not const,
// so tests can lower it to exercise the parallel path on a tiny fixture repo.
var parallelLogMinCommits = 4000

type Commit struct {
	Hash      string
	Author    string
	Email     string
	Date      time.Time
	Subject   string
	IsMerge   bool
	FileStats []FileStat
	// CoAuthors holds the names from Co-authored-by trailers (canonicalized later,
	// like Author). Squash-merges and pairing land the real authors here; a
	// commit's contribution is split equally among {Author} ∪ CoAuthors.
	CoAuthors []string
}

// coAuthorSep matches the git `%(trailers:...separator=%x1f)` byte between
// multiple Co-authored-by values.
const coAuthorSep = "\x1f"

// parseCoAuthorNames extracts human author NAMES from a Co-authored-by trailer
// field (values are "Name <email>", joined by \x1f). The name is the text before
// the " <email>". Bot / AI-assistant co-authors are dropped here (see isBotIdentity):
// gravity measures HUMAN structural influence, so an AI-assisted commit must not
// dilute the human author. Empty in → nil, so a solo commit allocates nothing.
func parseCoAuthorNames(field string) []string {
	if field == "" {
		return nil
	}
	var out []string
	for _, v := range strings.Split(field, coAuthorSep) {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		name, email := v, ""
		if i := strings.LastIndex(v, " <"); i > 0 {
			name = strings.TrimSpace(v[:i])
			email = strings.TrimSuffix(strings.TrimSpace(v[i+2:]), ">")
		}
		if name == "" || isBotIdentity(name, email) {
			continue
		}
		out = append(out, name)
	}
	return out
}

// builtinBotEmailSuffixes are AI coding-assistant email domains used in
// Co-authored-by trailers by their CLIs. Most AGENTS on GitHub already commit as
// "…[bot]" accounts (caught by the [bot] rule), so this list targets the tool
// trailers. Email-anchored to avoid excluding a human who shares a first name with
// a tool (e.g. a person named "Devin").
var builtinBotEmailSuffixes = []string{
	"@anthropic.com",   // Claude
	"@cursor.com",      // Cursor
	"@cursor.sh",       //
	"@devin.ai",        // Devin (Cognition)
	"@codeium.com",     // Windsurf / Codeium
	"@sourcegraph.com", // Cody — vendor bot address (humans use personal emails)
}

// builtinBotSubstrings match in the name OR email — only for handles distinctive
// enough not to collide with a human name.
var builtinBotSubstrings = []string{"copilot"}

// extraBotCoAuthorPatterns are config-supplied substrings (lowercased) matched in
// a co-author's name OR email. Set once per run via ConfigureBotCoAuthorPatterns —
// a package var (like blameMoveArgs) because bot detection runs deep in the parser
// where config isn't threaded. Lets a repo tag its own AI tools as non-human.
var extraBotCoAuthorPatterns []string

// ConfigureBotCoAuthorPatterns sets the run's extra bot/AI co-author patterns
// (config bot_coauthor_patterns). Call once before parsing; empty entries ignored.
func ConfigureBotCoAuthorPatterns(patterns []string) {
	extraBotCoAuthorPatterns = extraBotCoAuthorPatterns[:0]
	for _, p := range patterns {
		if p = strings.ToLower(strings.TrimSpace(p)); p != "" {
			extraBotCoAuthorPatterns = append(extraBotCoAuthorPatterns, p)
		}
	}
}

// isBotIdentity reports whether a co-author is a bot or AI assistant rather than a
// human engineer, so it can be excluded from the contributor split (gravity is a
// measure of HUMAN durable contribution). Checks the GitHub "[bot]" convention, a
// built-in list of AI-tool email domains + handles, and any configured extras.
func isBotIdentity(name, email string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	e := strings.ToLower(strings.TrimSpace(email))
	if strings.HasSuffix(n, "[bot]") || strings.Contains(e, "[bot]") {
		return true
	}
	for _, s := range builtinBotEmailSuffixes {
		if strings.HasSuffix(e, s) {
			return true
		}
	}
	for _, s := range builtinBotSubstrings {
		if strings.Contains(n, s) || (e != "" && strings.Contains(e, s)) {
			return true
		}
	}
	for _, s := range extraBotCoAuthorPatterns {
		if (n != "" && strings.Contains(n, s)) || (e != "" && strings.Contains(e, s)) {
			return true
		}
	}
	return false
}

// FileStat holds contribution-eligible line counts for a file in a commit.
// Code files exclude comment-only and blank lines (gaming protection);
// prose files (.md/.txt/etc.) and unknown types count every line.
type FileStat struct {
	Insertions int
	Deletions  int
	Filename   string
}

// logArgs builds the `git log` arguments for a full-history walk. With
// commentFilter it adds `-p` so parseLogStream can strip comment/blank diff
// lines; without it, only `--numstat` is requested — far less output and the
// dominant ParseLog cost on large repos — at the price of comment-inclusive
// insertion counts.
func logArgs(commentFilter bool, extra ...string) []string {
	args := []string{
		"log", "--all", "--no-merges", "--no-color",
		"--format=COMMIT:%H|%an|%ae|%ai|%(trailers:key=Co-authored-by,valueonly,separator=%x1f)|%s",
		"--numstat",
	}
	if commentFilter {
		args = append(args, "-p")
	}
	return append(args, extra...)
}

// ParseLog returns non-merge commits with per-file line stats. With
// commentFilter it uses `-p --numstat` (diff hunks let comment/blank lines be
// filtered out so comment spam cannot inflate Production, Design, or Debt);
// otherwise it uses `--numstat` only for speed.
func ParseLog(ctx context.Context, repoPath string, commentFilter bool) ([]Commit, error) {
	stdout, cmd, err := RunStream(ctx, repoPath, logArgs(commentFilter)...)
	if err != nil {
		return nil, err
	}
	defer stdout.Close()
	commits, scanErr := parseLogStream(stdout)
	if scanErr != nil && cmd.Process != nil {
		// A hard read error left git with unconsumed stdout; kill it so it
		// can't block on a full pipe and wedge cmd.Wait() forever.
		_ = cmd.Process.Kill()
	}
	waitErr := cmd.Wait()
	if scanErr != nil {
		return commits, scanErr
	}
	return commits, waitErr
}

// maxLineKeep bounds how many bytes of a single physical line parseLogStream
// holds in memory. A diff body can contain a generated/minified file committed
// as one multi-megabyte line; we must still CONSUME the whole physical line so
// git can finish writing and never blocks on a full stdout pipe — an unconsumed
// over-long line was the cause of a ParseLog deadlock on huge repos (ClickHouse)
// where bufio.Scanner's 16MB token cap made the parser bail mid-stream, leaving
// git stuck on write and cmd.Wait() blocked forever. We only need a bounded
// prefix to read the line's leading byte and run comment filtering; the rest is
// drained and discarded. Real source lines are far below this, so only
// pathological blobs are truncated — and their numstat counts come from the
// numstat region, not the diff body, so the metrics are unaffected.
const maxLineKeep = 1 << 20 // 1 MiB

// readLogLine reads one '\n'-terminated line from r, consuming the entire
// physical line but retaining at most maxLineKeep bytes. Unlike bufio.Scanner it
// never fails on an over-long line. It returns (line, nil) for each line — the
// trailing newline stripped — and (nil, io.EOF) once the stream is exhausted. A
// final line without a trailing newline is returned with a nil error, and the
// following call reports io.EOF. The returned slice is valid only until the next
// call.
func readLogLine(r *bufio.Reader) ([]byte, error) {
	frag, err := r.ReadSlice('\n')
	if err != bufio.ErrBufferFull {
		if len(frag) == 0 && err == io.EOF {
			return nil, io.EOF
		}
		if err == nil || err == io.EOF {
			return trimLineEnd(frag), nil
		}
		return trimLineEnd(frag), err
	}
	// Over-long line: keep a bounded prefix, then drain the remainder so the
	// writer (git) is never blocked on a full pipe.
	kept := make([]byte, len(frag))
	copy(kept, frag)
	for err == bufio.ErrBufferFull {
		frag, err = r.ReadSlice('\n')
		if room := maxLineKeep - len(kept); room > 0 {
			if room > len(frag) {
				room = len(frag)
			}
			kept = append(kept, frag[:room]...)
		}
	}
	if err != nil && err != io.EOF {
		return trimLineEnd(kept), err
	}
	return trimLineEnd(kept), nil
}

// trimLineEnd drops a trailing "\n" or "\r\n" from a raw line.
func trimLineEnd(b []byte) []byte {
	if n := len(b); n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
		if n := len(b); n > 0 && b[n-1] == '\r' {
			b = b[:n-1]
		}
	}
	return b
}

// parseLogStream parses the stdout of a `git log -p --numstat` invocation into
// commits with comment-filtered per-file line counts. Shared by the serial
// ParseLog and the per-chunk workers of ParseLogParallel. It always reads to EOF
// (or a hard read error) so the upstream git process can drain its stdout — see
// maxLineKeep for why bailing mid-stream deadlocks.
func parseLogStream(r io.Reader) ([]Commit, error) {
	br := bufio.NewReaderSize(r, maxLineKeep)

	// intern deduplicates FileStat.Filename backing strings. The same path recurs
	// across thousands of commits, and each numstat line would otherwise allocate
	// its own copy — millions of identical strings on a large history. Collapsing
	// them to one backing string per distinct path is the dominant filename-memory
	// win and is value-neutral (an interned string is byte-equal to the original).
	intern := make(map[string]string)
	internStr := func(s string) string {
		if got, ok := intern[s]; ok {
			return got
		}
		intern[s] = s
		return s
	}

	var commits []Commit
	var current *Commit

	var inDiff, sawHunk, capped bool
	var curFileName string
	var filter *FileFilter
	var fIns, fDel, fLines int

	flushFile := func() {
		defer func() {
			inDiff, sawHunk, capped = false, false, false
			curFileName = ""
			filter = nil
			fIns, fDel, fLines = 0, 0, 0
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

	for {
		lineBytes, readErr := readLogLine(br)
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			// Hard read error (e.g. the pipe broke). Return what we have; the
			// caller kills/waits the git process so it can't deadlock.
			flushFile()
			if current != nil {
				commits = append(commits, *current)
			}
			return commits, readErr
		}
		line := string(lineBytes)

		if strings.HasPrefix(line, "COMMIT:") {
			flushFile()
			if current != nil {
				commits = append(commits, *current)
			}
			// Fields: hash|an|ae|ai|<co-authors>|subject. The Co-authored-by trailer
			// field sits BEFORE the subject so the subject stays the LAST field and
			// may still contain '|'. Co-authors are \x1f-separated (never a newline).
			parts := strings.SplitN(line[7:], "|", 6)
			if len(parts) < 6 {
				current = nil
				continue
			}
			date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])
			current = &Commit{
				Hash:      parts[0],
				Author:    parts[1],
				Email:     parts[2],
				Date:      date,
				CoAuthors: parseCoAuthorNames(parts[4]),
				Subject:   parts[5],
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
			if capped {
				// Pathologically large file diff — skip the rest of its hunk
				// lines without per-line comment filtering (the costly part).
				// flushFile keeps the raw numstat counts (filter == nil path).
				continue
			}
			if strings.HasPrefix(line, "@@") {
				filter = NewFileFilter(curFileName)
				continue
			}
			if line == "" {
				continue
			}
			fLines++
			if fLines > maxFilteredDiffLines {
				// One commit dumped a giant generated/vendored file. Filtering
				// its diff line-by-line is what makes ParseLog crawl on huge
				// repos; cap it and fall back to numstat for this file.
				capped = true
				filter = nil
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
			Filename:   internStr(resolveRenamePath(parts[2])),
		})
	}

	flushFile()
	if current != nil {
		commits = append(commits, *current)
	}
	return commits, nil
}

// parallelLogChunksPerWorker oversamples chunks relative to workers. Diff volume
// per commit is wildly uneven (a repo can have bursts of huge commits), so equal
// commit-count chunks are load-imbalanced and the slowest chunk dominates
// wall-time. Cutting history into many more, smaller chunks than there are
// workers — drained by a pool — shrinks the slowest single chunk (≈ 1/(workers·K)
// of history) and lets idle workers steal the remaining work.
const parallelLogChunksPerWorker = 8

// ParseLogParallel is a drop-in faster ParseLog for large repos. The cost of
// ParseLog is git generating `-p` (full patch) output for every commit AND the
// Go side comment-filtering every diff line; on a 200k-commit repo that single
// serial stream is the dominant phase of analysis. This lists the commits cheaply
// first (rev-list emits no diffs), cuts them into many contiguous chunks, and
// parses each chunk's `git log -p` concurrently via a worker pool, then
// concatenates in history order. Output is identical to ParseLog: same commit set
// (--all --no-merges), same order, same per-file filtered counts.
//
// Small repos (or workers < 2) fall back to serial ParseLog.
func ParseLogParallel(ctx context.Context, repoPath string, workers int, commentFilter bool) ([]Commit, error) {
	if workers < 2 {
		return ParseLog(ctx, repoPath, commentFilter)
	}
	shas, err := revListAll(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	if len(shas) < parallelLogMinCommits {
		return ParseLog(ctx, repoPath, commentFilter)
	}

	chunks := chunkStrings(shas, workers*parallelLogChunksPerWorker)
	results := make([][]Commit, len(chunks))
	errs := make([]error, len(chunks))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i], errs[i] = parseLogChunk(ctx, repoPath, chunks[i], commentFilter)
			}
		}()
	}
	for i := range chunks {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	total := 0
	for i, e := range errs {
		if e != nil {
			return nil, e
		}
		total += len(results[i])
	}
	out := make([]Commit, 0, total)
	for i := range results {
		out = append(out, results[i]...)
		// Release each chunk as it is copied so the concatenation doesn't hold
		// BOTH the per-chunk slices and the merged slice at once — that doubled
		// peak resident commits on large histories. Output is unchanged.
		results[i] = nil
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

// parseLogChunk runs `git log --numstat [-p]` over exactly the given commits
// (fed on stdin, shown in input order via --no-walk) and parses the stream. Each
// commit is still diffed against its first parent, identical to the serial walk.
func parseLogChunk(ctx context.Context, repoPath string, shas []string, commentFilter bool) ([]Commit, error) {
	// Derive a cancelable context so a hard parse error can kill this chunk's
	// git before cmd.Wait(); otherwise a git left mid-write on a full stdout
	// pipe would wedge Wait() forever and, through wg.Wait(), hang the whole run.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	args := []string{
		"log", "--no-walk=unsorted", "--no-merges", "--no-color",
		"--format=COMMIT:%H|%an|%ae|%ai|%(trailers:key=Co-authored-by,valueonly,separator=%x1f)|%s",
		"--numstat",
	}
	if commentFilter {
		args = append(args, "-p")
	}
	args = append(args, "--stdin")
	cmd := exec.CommandContext(ctx, "git", args...)
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
	if scanErr != nil {
		cancel()
	}
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
		// No co-author trailer here: this path only detects fix/revert subjects in
		// merge messages (no author aggregation), and its parser expects 5 fields.
		"--format=COMMIT:%H|%an|%ae|%ai|%s",
	)
	if err != nil {
		return nil, err
	}

	var commits []Commit
	for _, line := range lines {
		if !strings.HasPrefix(line, "COMMIT:") {
			continue
		}
		parts := strings.SplitN(line[7:], "|", 5)
		if len(parts) < 5 {
			continue
		}
		date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])
		commits = append(commits, Commit{
			Hash:    parts[0],
			Author:  parts[1],
			Email:   parts[2],
			Date:    date,
			Subject: parts[4],
			IsMerge: true,
		})
	}

	return commits, nil
}
