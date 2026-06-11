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
	if minMonths <= 1 {
		return nil
	}

	// months[module] is the set of distinct calendar months (year*12+month) in
	// which a fallback-derived path of that module was touched.
	months := make(map[string]map[int]struct{})
	// declared marks modules that were pattern-declared on ANY path. A module
	// declared even once is intent-bearing and never folds, even if other
	// paths resolve it as a fallback (declared wins).
	declared := make(map[string]bool)

	for _, c := range commits {
		ym := int(c.Date.Year())*12 + int(c.Date.Month())
		for _, fs := range c.FileStats {
			module, isDeclared := mr.ResolveWithOrigin(fs.Filename)
			if module == "" {
				// Excluded — outside module topology entirely.
				continue
			}
			if isDeclared {
				declared[module] = true
				continue
			}
			set := months[module]
			if set == nil {
				set = make(map[int]struct{})
				months[module] = set
			}
			set[ym] = struct{}{}
		}
	}

	var fold map[string]bool
	for module, set := range months {
		if declared[module] {
			// Declared somewhere else — intent wins, never fold.
			continue
		}
		if len(set) < minMonths {
			if fold == nil {
				fold = make(map[string]bool)
			}
			fold[module] = true
		}
	}
	return fold
}
