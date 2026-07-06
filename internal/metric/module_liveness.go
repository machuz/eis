package metric

import (
	"github.com/machuz/eis/v2/internal/git"
)

// ComputeModuleFold returns the set of fallback-derived modules that have NOT
// earned module-hood — touched in fewer than minMonths distinct calendar
// months — which the caller folds into PeripheralModule via WithFold.
// Pattern-declared modules never fold. Excluded modules ("") are ignored.
// Deterministic: the month set is derived from commit.Date, not wall-clock.
// minMonths <= 1 disables the gate (returns nil).
//
// This is the EIS module liveness gate (ADR step 2): pattern-declared modules
// carry authored intent and are admitted immediately, while fallback-derived
// candidates (the conservative 2-component default — where date-dir / DML /
// dump pollution enters) must EARN module-hood through observed temporal
// persistence across distinct calendar months. Non-graduated fallback modules
// fold into PeripheralModule (observation preserved, not dropped — D-07).
func ComputeModuleFold(commits []git.Commit, mr ModuleResolver, minMonths int) map[string]bool {
	acc := NewModuleFoldAccumulator(mr, minMonths)
	for i := range commits {
		acc.Add(&commits[i])
	}
	return acc.Result()
}

// ModuleFoldAccumulator computes the module-liveness fold incrementally, one
// commit at a time, so a streaming log walk can derive it without materializing
// []Commit. Feed each (already alias-resolved / exclusion-filtered) commit via
// Add, then call Result. The result is identical to ComputeModuleFold over the
// same commit set regardless of Add order (month sets + declared flags are
// order-independent).
type ModuleFoldAccumulator struct {
	mr        ModuleResolver
	minMonths int
	// months[module] is the set of distinct calendar months (year*12+month) in
	// which a fallback-derived path of that module was touched.
	months map[string]map[int]struct{}
	// declared marks modules that were pattern-declared on ANY path — intent
	// wins, so such a module never folds.
	declared map[string]bool
	// cache memoizes filename → resolution (pure for a fixed resolver); the same
	// path recurs across thousands of commits.
	cache map[string]foldResolved
}

type foldResolved struct {
	module     string
	isDeclared bool
}

// NewModuleFoldAccumulator returns an accumulator. minMonths <= 1 disables the
// gate (Result returns nil), matching ComputeModuleFold.
func NewModuleFoldAccumulator(mr ModuleResolver, minMonths int) *ModuleFoldAccumulator {
	return &ModuleFoldAccumulator{
		mr:        mr,
		minMonths: minMonths,
		months:    make(map[string]map[int]struct{}),
		declared:  make(map[string]bool),
		cache:     make(map[string]foldResolved),
	}
}

// Add folds one commit's file touches into the month/declared evidence.
func (a *ModuleFoldAccumulator) Add(c *git.Commit) {
	if a.minMonths <= 1 {
		return
	}
	ym := int(c.Date.Year())*12 + int(c.Date.Month())
	for _, fs := range c.FileStats {
		res, ok := a.cache[fs.Filename]
		if !ok {
			m, d := a.mr.ResolveWithOrigin(fs.Filename)
			res = foldResolved{module: m, isDeclared: d}
			a.cache[fs.Filename] = res
		}
		module, isDeclared := res.module, res.isDeclared
		if module == "" {
			// Excluded — outside module topology entirely.
			continue
		}
		if isDeclared {
			a.declared[module] = true
			continue
		}
		set := a.months[module]
		if set == nil {
			set = make(map[int]struct{})
			a.months[module] = set
		}
		set[ym] = struct{}{}
	}
}

// Result finalizes the fold set (fallback modules touched in < minMonths distinct
// months that were never pattern-declared). Returns nil when the gate is off.
func (a *ModuleFoldAccumulator) Result() map[string]bool {
	if a.minMonths <= 1 {
		return nil
	}
	var fold map[string]bool
	for module, set := range a.months {
		if a.declared[module] {
			continue
		}
		if len(set) < a.minMonths {
			if fold == nil {
				fold = make(map[string]bool)
			}
			fold[module] = true
		}
	}
	return fold
}
