package output

import (
	"encoding/json"
	"io"
)

// WriteContext is the structured response of `eis get-write-context` — the core
// an MCP `get_write_context` tool exposes. Unlike the precheck hook (a Claude
// Code envelope, silent on healthy code), this is a QUERY API: it returns an
// entry for every path that resolves to an indexed module, clean or not.
//
// FIREWALL: like the index and the hook, it carries NO owner names — only
// structural facts and directives.
type WriteContext struct {
	Modules []WriteContextEntry `json:"modules"`
}

// WriteContextEntry is one path's resolved module and its structural facts.
type WriteContextEntry struct {
	Path                   string   `json:"path"`
	Module                 string   `json:"module"`
	DebtTier               string   `json:"debt_tier"` // "" | "Dead" | "Orphaned"
	AtRisk                 bool     `json:"at_risk"`
	OwnerActive            bool     `json:"owner_active"`
	OwnerLeftDays          int      `json:"owner_left_days"`
	UntouchedDays          int      `json:"untouched_days"`
	OwnershipConcentration float64  `json:"ownership_concentration"`
	Recommendation         string   `json:"recommendation"`
	Directives             []string `json:"directives"` // [] when clean

	// Reserved (v0 always null):
	//   - SurvivingIdiomDigest: Build2 idiom/anchor propagation.
	//   - per-module change_pressure is intentionally absent (cut by the lean
	//     pipeline the index is built with).
	SurvivingIdiomDigest *string `json:"surviving_idiom_digest"`
}

// EncodeWriteContext writes the structured context as indented JSON to w.
func EncodeWriteContext(w io.Writer, wc WriteContext) error {
	if wc.Modules == nil {
		wc.Modules = []WriteContextEntry{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(wc)
}
