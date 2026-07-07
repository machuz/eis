package metric

import (
	"math"
	"sort"

	"github.com/machuz/eis/v2/internal/git"
)

// SubstantialAuthorLines is the minimum surviving-blame footprint for an author
// to count as a "substantial" contributor when deciding whether the robust/
// dormant pressure split is meaningful (see PressureThreshold).
const SubstantialAuthorLines = 1000

// MinSubstantiveLines is the minimum comment-excluded changed-line count
// (Insertions+Deletions on a single file) for that file to mark its module as
// "touched" for OTHERS-pressure — the contest signal behind RobustSurvival.
//
// git.FileStat.Insertions/Deletions are already comment- and blank-line-excluded
// (see internal/git/comment.go + log_test.go: a pure-comment or blank-only change
// yields 0/0), and a pure rename / mode-only change yields 0/0 from --numstat.
// So a file whose Insertions+Deletions is 0 changed NOTHING substantive — it is a
// cosmetic touch (rename, comment/doc edit, whitespace-only blank line). Counting
// such a touch as "another author contested this module" fakes contest: a solo
// author's surviving lines there would score as RobustSurvival and mint gravity
// where no one has actually stress-tested the code.
//
// The value is 1: ANY single real (non-comment, non-blank) inserted or deleted
// line is enough to count as a genuine touch. We intentionally do not raise the
// bar higher — a one-line fix by another engineer IS a real edit and a real
// contest; the goal is only to exclude ZERO-substance cosmetic commits, not to
// judge how large a change must be. (Note: git log here does not pass -w, so a
// pure re-indent/reformat that git records as changed line content still counts
// as substantive — this is the same signal Production/Design already trust, so
// others-pressure stays consistent with the rest of the metric path.)
const MinSubstantiveLines = 1

// PressureThreshold returns the change-pressure cutoff that the robust/dormant
// survival split compares against. It returns +Inf — forcing every surviving
// line into dormant — when the repo is solo-dominated, because "others'
// pressure" cannot be observed without at least two substantial authors.
//
// "Substantial" is an ABSOLUTE footprint (>= minLines surviving blame lines),
// deliberately NOT a share of the repo. The earlier share-based gate (>=2
// authors each owning >=10% of all blame) conflated two opposite situations:
// a solo-dominated repo (one author owns nearly everything) AND a healthy,
// highly distributed repo where ownership is spread so thin that no single
// author reaches 10% (React, Kubernetes, Rust, Rails). Both failed the gate,
// so robust survival was zeroed for everyone in large collaborative repos,
// collapsing survGate, Gravity, and Architect detection. Counting authors by
// absolute footprint passes distributed repos while still excluding true
// solo projects.
func PressureThreshold(pressure ChangePressure, blameByAuthor map[string]int, minLines int) float64 {
	substantial := 0
	for _, count := range blameByAuthor {
		if count >= minLines {
			substantial++
		}
	}
	if substantial < 2 {
		return math.Inf(1)
	}
	return pressure.MedianPressure()
}

// ChangePressure maps module path → pressure value (commits / blame lines).
type ChangePressure map[string]float64

// CalcChangePressure computes per-module change pressure.
// pressure = commits_touching_module / blame_lines_in_module
func CalcChangePressure(commits []git.Commit, blameLines []git.BlameLine, mr ModuleResolver) ChangePressure {
	// Count commits per module (a commit counts once per module it touches)
	moduleCommits := make(map[string]int)
	for _, c := range commits {
		touched := make(map[string]bool)
		for _, fs := range c.FileStats {
			if mod := mr.ModuleOf(fs.Filename); mod != "" {
				touched[mod] = true
			}
		}
		for mod := range touched {
			moduleCommits[mod]++
		}
	}
	return CalcChangePressureFrom(moduleCommits, blameLines, mr)
}

// CalcChangePressureFrom is CalcChangePressure with the per-module commit counts
// already tallied (e.g. by the streaming CommitAggregator — Cochange.ModuleCommits
// is exactly this map), so the full commit slice need not be retained. Only the
// blame-derived denominator is computed here. Value-identical to CalcChangePressure
// for the same moduleCommits.
func CalcChangePressureFrom(moduleCommits map[string]int, blameLines []git.BlameLine, mr ModuleResolver) ChangePressure {
	// Count blame lines per module
	moduleBlameLines := make(map[string]int)
	for _, bl := range blameLines {
		if mod := mr.ModuleOf(bl.Filename); mod != "" {
			moduleBlameLines[mod]++
		}
	}

	// Calculate pressure
	pressure := make(ChangePressure)
	for mod, commits := range moduleCommits {
		lines := moduleBlameLines[mod]
		if lines > 0 {
			pressure[mod] = float64(commits) / float64(lines)
		} else {
			// Module has commits but no surviving blame lines — high churn
			pressure[mod] = float64(commits)
		}
	}

	// Also include modules with blame lines but no recent commits (pressure = 0)
	for mod := range moduleBlameLines {
		if _, ok := pressure[mod]; !ok {
			pressure[mod] = 0
		}
	}

	return pressure
}

// OthersPressure answers, per (module, author), how much authors OTHER than that
// author churn the module — the basis of "others-contested" robust survival.
//
// Total change pressure can be self-inflicted: an author who commits to their own
// module 50 times makes it high-pressure, so their surviving lines there read as
// "robust" even though no one else has ever stress-tested them. That lets a solo
// author of low-quality code that simply hasn't been touched yet show durable
// survival. Splitting robust/dormant on OTHERS' pressure instead fixes it: code
// is robust only if it survives where OTHER people are actively working. A
// founder whose foundational module others keep editing still passes; a dead
// corner or a self-churned module does not.
type OthersPressure struct {
	moduleCommits       map[string]int            // total commits touching each module
	authorModuleCommits map[string]map[string]int // module → author → commits touching it
	moduleBlame         map[string]int            // surviving blame lines per module
}

// CalcOthersPressure tallies, per module, the total commits and the per-author
// commit counts, plus the surviving blame-line count, so For(mod, author) can
// return the module's pressure from everyone EXCEPT author.
func CalcOthersPressure(commits []git.Commit, blameLines []git.BlameLine, mr ModuleResolver) *OthersPressure {
	moduleCommits := make(map[string]int)
	moduleAuthorCommits := make(map[string]map[string]int)
	for _, c := range commits {
		touched := make(map[string]bool)
		for _, fs := range c.FileStats {
			// Substance gate: a cosmetic touch (rename / comment / blank-only,
			// i.e. Insertions+Deletions == 0 after comment-exclusion) does NOT
			// mark the module contested. See MinSubstantiveLines.
			if fs.Insertions+fs.Deletions < MinSubstantiveLines {
				continue
			}
			if mod := mr.ModuleOf(fs.Filename); mod != "" {
				touched[mod] = true
			}
		}
		for mod := range touched {
			moduleCommits[mod]++
			am := moduleAuthorCommits[mod]
			if am == nil {
				am = make(map[string]int)
				moduleAuthorCommits[mod] = am
			}
			am[c.Author]++
		}
	}
	return CalcOthersPressureFrom(moduleCommits, moduleAuthorCommits, blameLines, mr)
}

// CalcOthersPressureFrom is CalcOthersPressure with the commit-derived tallies
// (total commits per module + module→author→commits) already computed — the
// streaming CommitAggregator supplies Cochange.ModuleCommits and
// ModuleAuthorCommits, so the commit slice need not be retained. Only the
// surviving blame count per module is computed here.
func CalcOthersPressureFrom(moduleCommits map[string]int, moduleAuthorCommits map[string]map[string]int, blameLines []git.BlameLine, mr ModuleResolver) *OthersPressure {
	o := &OthersPressure{
		moduleCommits:       moduleCommits,
		authorModuleCommits: moduleAuthorCommits,
		moduleBlame:         make(map[string]int),
	}
	for _, bl := range blameLines {
		if mod := mr.ModuleOf(bl.Filename); mod != "" {
			o.moduleBlame[mod]++
		}
	}
	return o
}

// For returns the change pressure on mod from authors other than `author`:
// (others' commits touching mod) / (surviving blame lines in mod). Mirrors
// CalcChangePressure's pressure = commits/lines, with the author's own commits
// removed; a module with commits but no surviving lines yields the raw others'
// commit count, the same high-churn convention as the total pressure.
func (o *OthersPressure) For(mod, author string) float64 {
	if o == nil {
		return 0
	}
	others := o.moduleCommits[mod] - o.authorModuleCommits[mod][author]
	if others < 0 {
		others = 0
	}
	lines := o.moduleBlame[mod]
	if lines <= 0 {
		return float64(others)
	}
	return float64(others) / float64(lines)
}

// MedianPressure returns the median pressure value across all modules.
func (cp ChangePressure) MedianPressure() float64 {
	if len(cp) == 0 {
		return 0
	}
	vals := make([]float64, 0, len(cp))
	for _, v := range cp {
		vals = append(vals, v)
	}
	sort.Float64s(vals)

	n := len(vals)
	if n%2 == 0 {
		return (vals[n/2-1] + vals[n/2]) / 2
	}
	return vals[n/2]
}
