package git

import (
	"bufio"
	"context"
	"strings"
)

// subtreeFieldSep separates the SHA from the git-subtree-split trailer value in
// the SubtreeSquashCommits log format. \x1f (unit separator) never appears in a
// SHA or trailer value.
const subtreeFieldSep = "\x1f"

// SubtreeSquashCommits returns the SHAs of `git subtree ... --squash` commits
// reachable from rev (use "HEAD", or a boundary commit for a point-in-time walk).
//
// A subtree-squash commit collapses an imported repo's entire history into a
// single commit authored by whoever ran the command — the monorepo integrator.
// git blame then attributes every untouched imported line to that one commit, so
// all the original authors' surviving mass is credited to the integrator. This
// is the same authorship-collapse the Co-authored-by split handles for
// squash-merges, but subtree-squash leaves no co-author trailers to split on.
//
// Detection keys on the machine-written `git-subtree-split` trailer, present on
// both `subtree add` and `subtree pull --squash`. Normal commits have an empty
// trailer, so the match is exact (zero false positives) and cheap — one log pass
// (~0.2s over 24k commits). This is observation, not inference: we read a marker
// git itself wrote.
func SubtreeSquashCommits(ctx context.Context, repoPath, rev string) (map[string]struct{}, error) {
	if rev == "" {
		rev = "HEAD"
	}
	stdout, cmd, err := RunStream(ctx, repoPath,
		"log", rev, "--no-color",
		"--format=%H"+subtreeFieldSep+"%(trailers:key=git-subtree-split,valueonly,separator=,)")
	if err != nil {
		return nil, err
	}
	defer stdout.Close()

	set := make(map[string]struct{})
	sc := bufio.NewScanner(stdout)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		i := strings.Index(line, subtreeFieldSep)
		if i < 0 {
			continue
		}
		if strings.TrimSpace(line[i+len(subtreeFieldSep):]) == "" {
			continue // ordinary commit: no git-subtree-split trailer
		}
		set[line[:i]] = struct{}{}
	}
	scanErr := sc.Err()
	if scanErr != nil && cmd.Process != nil {
		// Unconsumed stdout would wedge Wait on a full pipe; kill first.
		_ = cmd.Process.Kill()
	}
	waitErr := cmd.Wait()
	if scanErr != nil {
		return set, scanErr
	}
	return set, waitErr
}

// DropSubtreeSquashBlame removes blame lines whose commit is a subtree-squash
// import (see SubtreeSquashCommits). Those lines credit the integrator, not the
// real author — and the real author is unrecoverable from this repo because
// --squash discarded the original commits from its history. A line an engineer
// edited AFTER the import blames to that engineer's own commit and is kept, so
// this precisely targets untouched imported code. Returns the kept slice and the
// number dropped (for W-07 logging). A nil/empty set is a no-op.
func DropSubtreeSquashBlame(lines []BlameLine, squash map[string]struct{}) ([]BlameLine, int) {
	if len(squash) == 0 || len(lines) == 0 {
		return lines, 0
	}
	kept := make([]BlameLine, 0, len(lines))
	for _, bl := range lines {
		if _, ok := squash[bl.Commit]; ok {
			continue
		}
		kept = append(kept, bl)
	}
	return kept, len(lines) - len(kept)
}
