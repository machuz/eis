package metric

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/machuz/eis/v2/internal/git"
)

type DebtData struct {
	Generated map[string]int
	Cleaned   map[string]int
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
func CalcDebt(ctx context.Context, repoPath string, fixCommits []git.Commit, maxSample int, debtThreshold int, blameTimeoutSec int, resolve ResolveFunc, progressFn ProgressFunc, verboseFn VerboseFunc) (map[string]float64, *DebtData) {
	generated := make(map[string]int)
	cleaned := make(map[string]int)

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

		fixer := resolve(fc.Author)
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
			authors, err := git.BlameFileAtParent(fileCtx, repoPath, fc.Hash, f)
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
					verboseFn(fmt.Sprintf("    blame %s: %d authors (SLOW: %v)", f, len(authors), elapsed.Round(time.Millisecond)))
				} else {
					verboseFn(fmt.Sprintf("    blame %s: %d authors (%v)", f, len(authors), elapsed.Round(time.Millisecond)))
				}
			}

			for _, origAuthor := range authors {
				origAuthor = resolve(origAuthor)
				if origAuthor != fixer && origAuthor != "" {
					generated[origAuthor]++
					cleaned[fixer]++
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

		if total < debtThreshold {
			result[author] = 50 // neutral: insufficient data
			continue
		}

		// Score: 50 + 50 * (cleaned - generated) / (cleaned + generated)
		// Range: 0 (pure debt creator) to 100 (pure cleaner), 50 = balanced
		score := 50.0 + 50.0*float64(cln-gen)/float64(total)
		result[author] = math.Max(0, math.Min(100, score))
	}

	return result, &DebtData{Generated: generated, Cleaned: cleaned}
}
