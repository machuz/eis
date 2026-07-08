package cli

// eis get-write-context — the structured-JSON sibling of precheck-hook and the
// core an MCP `get_write_context` tool exposes. Same path→module resolution +
// index lookup + directive vocabulary as the hook (shared via resolveIndexModule
// and recommendationDirectives, so they can't drift), but it returns a query
// response instead of a Claude Code hook envelope.
//
// FIREWALL: no owner names — structural facts + directives only.

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/machuz/eis/v2/internal/output"
)

func runGetWriteContext(args []string) error {
	fs := flag.NewFlagSet("get-write-context", flag.ExitOnError)
	pathsFlag := fs.String("paths", "", "Comma-separated file paths the agent intends to write")
	// --intent is accepted but not yet used: reserved for future anchor/idiom
	// selection (Build2). Passed through so callers can start sending it now.
	_ = fs.String("intent", "", "What the agent intends to do (reserved; currently ignored)")
	indexPath := fs.String("index", "", "Path to write-index.json (default: <cwd>/.eis/write-index.json)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var paths []string
	for _, p := range strings.Split(*pathsFlag, ",") {
		if p = strings.TrimSpace(p); p != "" {
			paths = append(paths, p)
		}
	}

	cwd, _ := os.Getwd()

	resolvedIndex := *indexPath
	if resolvedIndex == "" {
		resolvedIndex = filepath.Join(cwd, defaultWriteIndexPath)
	}
	// Fail-open: a missing/broken index yields an empty modules list, never an
	// error (this is a best-effort context provider).
	idx, _ := loadWriteIndex(resolvedIndex)

	wc := buildWriteContext(paths, cwd, idx)
	return output.EncodeWriteContext(os.Stdout, wc)
}

// buildWriteContext is the pure, testable core: for each path, resolve → look up
// → assemble a structured entry. Paths that don't resolve to an indexed module
// are omitted (fail-open); indexed modules are always returned, clean or not
// (clean ⇒ empty debt_tier + empty directives). Reuses the precheck core so the
// two commands share one resolver and one directive vocabulary.
func buildWriteContext(paths []string, cwd string, idx output.WriteIndex) output.WriteContext {
	entries := make([]output.WriteContextEntry, 0, len(paths))
	for _, p := range paths {
		mod, entry, ok := resolveIndexModule(p, cwd, idx)
		if !ok {
			continue // unresolvable or not-in-index — omit (fail-open).
		}
		entries = append(entries, output.WriteContextEntry{
			Path:                   p,
			Module:                 mod,
			DebtTier:               entry.DebtTier,
			AtRisk:                 entry.AtRisk,
			OwnerActive:            entry.OwnerActive,
			OwnerLeftDays:          entry.OwnerLeftDays,
			UntouchedDays:          entry.UntouchedDays,
			OwnershipConcentration: entry.OwnershipConcentration,
			Recommendation:         entry.Recommendation,
			Directives:             recommendationDirectives(entry.Recommendation),
			SurvivingIdiomDigest:   nil, // reserved (Build2).
			// NOTE (firewall): entry carries no owner name; none is emitted.
		})
	}
	return output.WriteContext{Modules: entries}
}
