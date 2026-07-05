package metric

import (
	"context"
	"fmt"
	"math"
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
func CalcDebt(ctx context.Context, repoPath string, fixCommits []git.Commit, maxSample int, debtThreshold int, blameTimeoutSec int, resolve ResolveFunc, coMap map[string][]string, progressFn ProgressFunc, verboseFn VerboseFunc) (map[string]float64, *DebtData) {
	generated := make(map[string]float64)
	cleaned := make(map[string]float64)

	if resolve == nil {
		resolve = func(s string) string { return s }
	}

	// Sample fix commits
	sample := fixCommits
	if len(sample) > maxSample {
		sample = sample[:maxSample]
	}

	total := len(sample)
	for i, fc := range sample {
		// Check context cancellation
		if ctx.Err() != nil {
			break
		}

		fixer := resolve(fc.Author) // display / logging
		// The fix commit's contributor set shares the cleanup credit.
		fixerSet := CommitAuthors(fc)
		fShare := 1.0 / float64(len(fixerSet))
		commitStart := time.Now()

		// Get changed files
		files, err := git.DiffTreeFiles(ctx, repoPath, fc.Hash)
		if err != nil {
			if verboseFn != nil {
				verboseFn(fmt.Sprintf("  [debt] skip commit %s (diff-tree error: %v)", fc.Hash[:8], err))
			}
			if progressFn != nil {
				progressFn(i+1, total)
			}
			continue
		}

		if verboseFn != nil {
			verboseFn(fmt.Sprintf("  [debt] commit %d/%d %s by %s (%d files)", i+1, total, fc.Hash[:8], fixer, len(files)))
		}

		for _, f := range files {
			if f == "" {
				continue
			}
			if verboseFn != nil {
				verboseFn(fmt.Sprintf("    blaming %s ...", f))
			}
			// Blame at parent to find original authors (with configurable timeout per file)
			timeout := time.Duration(blameTimeoutSec) * time.Second
			if timeout <= 0 {
				timeout = 120 * time.Second
			}
			fileCtx, fileCancel := context.WithTimeout(ctx, timeout)
			fileStart := time.Now()
			blLines, err := git.BlameFileAtParent(fileCtx, repoPath, fc.Hash, f)
			timedOut := fileCtx.Err() != nil
			fileCancel()
			elapsed := time.Since(fileStart)
			if err != nil || timedOut {
				if verboseFn != nil {
					if timedOut {
						verboseFn(fmt.Sprintf("    blame %s: TIMEOUT (>%ds, skipped)", f, blameTimeoutSec))
					} else {
						verboseFn(fmt.Sprintf("    blame %s: error (%v)", f, err))
					}
				}
				continue
			}
			if verboseFn != nil {
				if elapsed > 2*time.Second {
					verboseFn(fmt.Sprintf("    blame %s: %d lines (SLOW: %v)", f, len(blLines), elapsed.Round(time.Millisecond)))
				} else {
					verboseFn(fmt.Sprintf("    blame %s: %d lines (%v)", f, len(blLines), elapsed.Round(time.Millisecond)))
				}
			}

			// Each original line's debt is split across its co-authors (via coMap),
			// and its cleanup credit across the fixer's co-authors — the cross-product
			// of shares. Self-cleaning (an author fixing their own line) is excluded,
			// as before. Per line the shares sum to ~1, so magnitudes are preserved.
			for _, bl := range blLines {
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

		if verboseFn != nil {
			verboseFn(fmt.Sprintf("  [debt] commit %s done in %v", fc.Hash[:8], time.Since(commitStart).Round(time.Millisecond)))
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
