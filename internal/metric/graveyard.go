package metric

import "github.com/machuz/eis/v2/internal/git"

// FileGrave holds one file's "death" genealogy: how often another author (B≠A)
// deleted/rewrote lines someone else had written. It is the negative of survival
// — code that people wrote here and others tore out.
//
// FIREWALL: it keeps only a COUNT of the distinct people who fought over the file
// (Contributors), never identities.
type FileGrave struct {
	File            string
	Module          string
	Deaths          int // others-contested deletion EVENTS (commit×file, B≠A)
	OthersDeadLines int // lines deleted that belonged to a different author
	TotalAdds       int // all lines added to this file (production denominator)
	Contributors    int // distinct people involved (deleted authors ∪ deleters)

	// contributors is the working set collapsed to Contributors at the end so no
	// identity ever leaves CalcGraveyard.
	contributors map[string]struct{}
}

// CalcGraveyard reconstructs a cheap line genealogy from numstat alone (no blame)
// to find others-contested deaths. commits MUST be in CHRONOLOGICAL order
// (oldest first) and already author-canonicalized / alias-resolved, so B≠A is
// accurate.
//
// Model: each file keeps a LIFO inventory of author-tagged line runs. A commit's
// deletions pop the most-recent runs first (newest lines rewritten first — the
// rewrite-of-a-recent-attempt signal); any popped run authored by someone other
// than the deleter is an others-contested death. Additions push a new run. Self-
// rewrite (deleter == author of the popped run) is NOT a death — that's healthy
// refactoring.
func CalcGraveyard(commits []git.Commit, mr ModuleResolver) map[string]*FileGrave {
	type run struct {
		author string
		n      int
	}
	inv := make(map[string][]run)
	graves := make(map[string]*FileGrave)

	get := func(f string) *FileGrave {
		g := graves[f]
		if g == nil {
			g = &FileGrave{File: f, Module: mr.ModuleOf(f), contributors: make(map[string]struct{})}
			graves[f] = g
		}
		return g
	}

	for _, c := range commits {
		if c.IsMerge {
			continue
		}
		for _, fsStat := range c.FileStats {
			f := fsStat.Filename
			g := get(f)

			// Deletions first, against the pre-commit inventory (LIFO).
			del := fsStat.Deletions
			othersDeleted := 0
			for del > 0 && len(inv[f]) > 0 {
				last := &inv[f][len(inv[f])-1]
				take := del
				if take > last.n {
					take = last.n
				}
				if last.author != c.Author {
					othersDeleted += take
					g.contributors[last.author] = struct{}{} // the A whose line died
				}
				last.n -= take
				del -= take
				if last.n == 0 {
					inv[f] = inv[f][:len(inv[f])-1]
				}
			}
			if othersDeleted > 0 {
				g.Deaths++
				g.OthersDeadLines += othersDeleted
				g.contributors[c.Author] = struct{}{} // the B who deleted
			}

			// Additions push a new author-tagged run.
			if fsStat.Insertions > 0 {
				inv[f] = append(inv[f], run{author: c.Author, n: fsStat.Insertions})
				g.TotalAdds += fsStat.Insertions
			}
		}
	}

	for _, g := range graves {
		g.Contributors = len(g.contributors)
		g.contributors = nil
	}
	return graves
}
