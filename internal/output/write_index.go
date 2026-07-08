package output

import (
	"encoding/json"
	"io"
)

// WriteIndex is the per-module "read before you write" index consumed by AI
// coding agents (and the pre-write hook that resolves module → structural facts
// locally, without a remote round-trip). Unlike the structural-debt drill (top-N,
// human-facing, owner names OK), this covers EVERY module and is AGENT-FACING:
//
//	FIREWALL: it carries NO person names — only structural facts (owner_active,
//	owner_left_days, untouched_days, ownership_concentration). Owner identity is
//	deliberately dropped here.
type WriteIndex struct {
	GeneratedAt string                      `json:"generated_at"` // RFC3339
	Commit      string                      `json:"commit"`       // HEAD sha
	Modules     map[string]WriteIndexModule `json:"modules"`
	// ModulePatterns are the config architecture path→module rules so the hook
	// can resolve a brand-new file's module id. When the user declared none,
	// these are the built-in defaults and ModulePatternsSource says so.
	ModulePatterns       []string `json:"module_patterns"`
	ModulePatternsSource string   `json:"module_patterns_source"` // "config" | "builtin_default"

	// Reserved for later increments (intentionally absent in v0):
	//   - per-module change_pressure: cut by the lean debt pipeline (#355).
	//   - surviving_idiom_digest / anchors: Build2 idiom propagation.
}

// WriteIndexModule is one module's structural facts. No owner name (firewall).
type WriteIndexModule struct {
	DebtTier               string  `json:"debt_tier"`               // "" | "Dead" | "Orphaned"
	AtRisk                 bool    `json:"at_risk"`                 // concentrated + owner active (bus-factor 1)
	OwnerActive            bool    `json:"owner_active"`            //
	OwnerLeftDays          int     `json:"owner_left_days"`         // days since owner last committed
	UntouchedDays          int     `json:"untouched_days"`          // days since any commit touched the module
	OwnershipConcentration float64 `json:"ownership_concentration"` // top-author blame share, 0-1
	Recommendation         string  `json:"recommendation"`          // orphaned_module|dead_module|bus_factor_1|""
}

// EncodeWriteIndex writes the index as indented JSON to w.
func EncodeWriteIndex(w io.Writer, idx WriteIndex) error {
	if idx.Modules == nil {
		idx.Modules = map[string]WriteIndexModule{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(idx)
}
