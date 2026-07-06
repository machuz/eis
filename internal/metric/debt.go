package metric

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

type DebtData struct {
	Generated map[string]float64
	Cleaned   map[string]float64
}

// ResolveFunc maps a git author name to its canonical name
type ResolveFunc func(string) string

// ProgressFunc reports debt analysis progress (done, total fix commits)
type ProgressFunc func(done, total int)

// VerboseFunc logs detailed per-operation info (message string)
type VerboseFunc func(msg string)

// CalcDebt calculates debt cleanup scores on a 0-100 scale.
// 50 = neutral (equal generation and cleanup, or insufficient data)
// >50 = net cleaner, <50 = net debt creator
// Formula: 50 + 50 * (cleaned - generated) / (cleaned + generated)
// coMap (commit SHA → contributor set) lets the "who generated this debt" side be
// split across the original lines' co-authors, matching the fixer side which
// splits across the fix commit's co-authors. Pass nil to disable co-author split.
func CalcDebt(ctx context.Context, repoPath string, fixCommits []git.Commit, maxSample int, debtThreshold int, blameTimeoutSec int, workers int, resolve ResolveFunc, coMap map[string][]string, progressFn ProgressFunc, verboseFn VerboseFunc) (map[string]float64, *DebtData) {
	generated := make(map[string]float64)
	cleaned := make(map[string]float64)

	if resolve == nil {
		resolve = func(s string) string { return s }
	}
	if workers <= 0 {
		workers = 4
	}

	// Sample fix commits
	sample := fixCommits
	if len(sample) > maxSample {
		sample = sample[:maxSample]
	}
	total := len(sample)

	timeout := time.Duration(blameTimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	// The git work — diff-tree per commit + blame-at-parent per changed file — is
	// the dominant cost of analysis on large repos and was fully serial. Run the
	// blames concurrently (each is an independent git subprocess), but ACCUMULATE
	// in the original (commit, file, line) order so the float sums are
	// bit-identical to the serial version. So: parallel gather → serial fold.
	type fileBlame struct {
		file  string
		lines []git.BlameLine
		ok    bool
	}
	perCommitFiles := make([][]fileBlame, total)

	// Phase 1: changed files per sampled commit (parallel diff-tree).
	parallelForEach(ctx, workers, total, func(i int) {
		files, err := git.DiffTreeFiles(ctx, repoPath, sample[i].Hash)
		if err != nil {
			if verboseFn != nil {
				verboseFn(fmt.Sprintf("  [debt] skip commit %s (diff-tree error: %v)", sample[i].Hash[:8], err))
			}
			return
		}
		fbs := make([]fileBlame, 0, len(files))
		for _, f := range files {
			if f != "" {
				fbs = append(fbs, fileBlame{file: f})
			}
		}
		perCommitFiles[i] = fbs
	})

	// Phase 2: blame each (commit, file) at the fix's parent (parallel worker pool
	// over the flattened job list; results written to their fixed slots).
	type job struct{ ci, fi int }
	var jobs []job
	for ci := range perCommitFiles {
		for fi := range perCommitFiles[ci] {
			jobs = append(jobs, job{ci, fi})
		}
	}
	parallelForEach(ctx, workers, len(jobs), func(k int) {
		j := jobs[k]
		if ctx.Err() != nil {
			return
		}
		fb := &perCommitFiles[j.ci][j.fi]
		fileCtx, fileCancel := context.WithTimeout(ctx, timeout)
		lines, err := git.BlameFileAtParent(fileCtx, repoPath, sample[j.ci].Hash, fb.file)
		timedOut := fileCtx.Err() != nil
		fileCancel()
		if err != nil || timedOut {
			if verboseFn != nil {
				if timedOut {
					verboseFn(fmt.Sprintf("    blame %s: TIMEOUT (>%ds, skipped)", fb.file, blameTimeoutSec))
				} else {
					verboseFn(fmt.Sprintf("    blame %s: error (%v)", fb.file, err))
				}
			}
			return
		}
		fb.lines = lines
		fb.ok = true
	})

	// Phase 3: fold into the debt maps in serial (commit, file, line) order — the
	// exact order the previous implementation used, so results are byte-identical.
	for i, fc := range sample {
		fixerSet := CommitAuthors(fc)
		fShare := 1.0 / float64(len(fixerSet))
		for _, fb := range perCommitFiles[i] {
			if !fb.ok {
				continue
			}
			for _, bl := range fb.lines {
				origSet := coMap[bl.Commit]
				if len(origSet) == 0 {
					origSet = []string{resolve(bl.Author)}
				}
				oShare := 1.0 / float64(len(origSet))
				for _, oa := range origSet {
					if oa == "" {
						continue
					}
					for _, fx := range fixerSet {
						fxr := resolve(fx)
						if fxr == "" || oa == fxr {
							continue
						}
						w := oShare * fShare
						generated[oa] += w
						cleaned[fxr] += w
					}
				}
			}
		}
		if progressFn != nil {
			progressFn(i+1, total)
		}
	}

	// Calculate scores on 0-100 scale
	result := make(map[string]float64)

	// Collect all authors
	allAuthors := make(map[string]bool)
	for a := range generated {
		allAuthors[a] = true
	}
	for a := range cleaned {
		allAuthors[a] = true
	}

	for author := range allAuthors {
		gen := generated[author]
		cln := cleaned[author]
		total := gen + cln

		if total < float64(debtThreshold) {
			result[author] = 50 // neutral: insufficient data
			continue
		}

		// Score: 50 + 50 * (cleaned - generated) / (cleaned + generated)
		// Range: 0 (pure debt creator) to 100 (pure cleaner), 50 = balanced
		score := 50.0 + 50.0*(cln-gen)/total
		result[author] = math.Max(0, math.Min(100, score))
	}

	return result, &DebtData{Generated: generated, Cleaned: cleaned}
}

// parallelForEach runs fn(i) for i in [0,n) across a bounded worker pool. fn must
// be safe to run concurrently for distinct i (the debt phases write only to slot
// i, so they are). Returns once every index has been processed.
func parallelForEach(ctx context.Context, workers, n int, fn func(i int)) {
	if n <= 0 {
		return
	}
	if workers < 1 {
		workers = 1
	}
	if workers > n {
		workers = n
	}
	jobs := make(chan int, n)
	for i := 0; i < n; i++ {
		jobs <- i
	}
	close(jobs)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				if ctx.Err() != nil {
					return
				}
				fn(i)
			}
		}()
	}
	wg.Wait()
}
