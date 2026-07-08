package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/color"
)

// StructuralDebtReport is the Tier-1 Structural Debt Meter result for a single
// domain. Mass is blame-derived surviving lines (the SAME source that decides
// ownership), so mass and orphan status can never disagree. See
// internal/cli/structural_debt.go for how it is computed.
//
// This is the AI-agnostic headline: what fraction of the classified source
// mass is already structural debt (Dead/Orphaned), and how much more is one
// departure away (bus-factor 1, at-risk).
type StructuralDebtReport struct {
	Domain         string        `json:"domain"`
	SDR            float64       `json:"sdr"`            // (dead + orphaned) / classified mass
	DeadRatio      float64       `json:"dead_ratio"`     // dead / classified mass
	OrphanedRatio  float64       `json:"orphaned_ratio"` // orphaned / classified mass
	AtRiskRatio    float64       `json:"at_risk_ratio"`  // leading indicator, NOT in SDR
	ClassifiedMass int           `json:"classified_mass"`
	DebtMass       int           `json:"debt_mass"` // dead + orphaned mass
	ModuleCount    int           `json:"module_count"`
	TopDebtModules []DebtModule  `json:"top_debt_modules"`
	Concentration  Concentration `json:"concentration"`
}

// DebtModule is one drill row for the "worst debt (by mass)" list. LastOwner /
// OwnerLeftDays / UntouchedDays make the debt visceral ("tanaka · left 8mo ago").
type DebtModule struct {
	Module string `json:"module"`
	Mass   int    `json:"mass"`
	Tier   string `json:"tier"` // "Dead" | "Orphaned"
	// LastOwner is the module's top blame author (from metric.ModuleOwnership).
	// TODO(owner-story): OwnerLeftDays / UntouchedDays need per-author last-active
	// dates (authorLastDate) and per-module last-commit dates, which the shared
	// pipeline does not yet surface on DomainResults — they stay 0 until wired.
	LastOwner     string `json:"last_owner"`
	OwnerLeftDays int    `json:"owner_left_days"`
	UntouchedDays int    `json:"untouched_days"`
}

// Concentration surfaces (rather than hides) when a single module dominates the
// debt mass — a 41% share in one file is a very different story from evenly
// spread rot.
type Concentration struct {
	TopModule string  `json:"top_module"`
	Share     float64 `json:"share"` // topModule.Mass / debtMass
}

// PrintStructuralDebtTable renders the human/CTO-facing report to stdout.
func PrintStructuralDebtTable(reports []StructuralDebtReport) {
	title := color.New(color.FgHiCyan, color.Bold)
	head := color.New(color.FgHiRed, color.Bold)
	sub := color.New(color.FgYellow)
	dim := color.New(color.FgHiBlack)

	for _, r := range reports {
		fmt.Println()
		title.Printf("Structural Debt — %s\n", r.Domain)

		head.Printf("  SDR  %s   ", pct(r.SDR))
		dim.Printf("debt / classified source mass\n")

		sub.Printf("        Dead      %s   %s lines\n", pct(r.DeadRatio), commafy(int(float64(r.ClassifiedMass)*r.DeadRatio+0.5)))
		sub.Printf("        Orphaned  %s   %s lines\n", pct(r.OrphanedRatio), commafy(int(float64(r.ClassifiedMass)*r.OrphanedRatio+0.5)))

		atRiskMass := int(float64(r.ClassifiedMass)*r.AtRiskRatio + 0.5)
		fmt.Printf("  At-risk %s  %s lines  ", pct(r.AtRiskRatio), commafy(atRiskMass))
		dim.Printf("(bus-factor 1, owner active — next to fall)\n")

		fmt.Printf("  Base   %s lines / %d modules\n", commafy(r.ClassifiedMass), r.ModuleCount)

		if len(r.TopDebtModules) > 0 {
			dim.Printf("  Worst debt (by mass):  MASS / TIER / MODULE / OWNER·SINCE\n")
			for _, m := range r.TopDebtModules {
				owner := m.LastOwner
				if owner == "" {
					owner = "—"
				}
				since := ""
				if m.OwnerLeftDays > 0 {
					since = fmt.Sprintf(" · left %dd ago", m.OwnerLeftDays)
				}
				fmt.Printf("    %10s  %-8s  %s  %s%s\n",
					commafy(m.Mass), m.Tier, m.Module, owner, since)
			}
		}

		if r.Concentration.TopModule != "" && r.Concentration.Share > 0 {
			warn := color.New(color.FgHiYellow)
			warn.Printf("  ⚠ concentration: %s of debt mass in one module %s\n",
				pct(r.Concentration.Share), r.Concentration.TopModule)
		}
		fmt.Println()
	}
	// TODO(trend): SaaS stores per-run snapshots for time-series; the CLI stays
	// stateless and AI-agnostic (no authorship lens) in v0.
}

// PrintStructuralDebtJSON emits the machine contract consumed by the SaaS / MCP
// layer. A stable top-level array holds one strict report object per domain.
func PrintStructuralDebtJSON(reports []StructuralDebtReport) error {
	if reports == nil {
		reports = []StructuralDebtReport{}
	}
	// Ensure top_debt_modules serializes as [] (not null) for every report.
	for i := range reports {
		if reports[i].TopDebtModules == nil {
			reports[i].TopDebtModules = []DebtModule{}
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(reports)
}

// PrintStructuralDebtCSV emits one summary row per domain (mirrors PrintTeamCSV).
func PrintStructuralDebtCSV(reports []StructuralDebtReport) {
	w := csv.NewWriter(os.Stdout)
	w.Write([]string{
		"domain", "sdr", "dead_ratio", "orphaned_ratio", "at_risk_ratio",
		"classified_mass", "debt_mass", "module_count",
		"concentration_top_module", "concentration_share",
	})
	for _, r := range reports {
		w.Write([]string{
			r.Domain,
			strconv.FormatFloat(r.SDR, 'f', 4, 64),
			strconv.FormatFloat(r.DeadRatio, 'f', 4, 64),
			strconv.FormatFloat(r.OrphanedRatio, 'f', 4, 64),
			strconv.FormatFloat(r.AtRiskRatio, 'f', 4, 64),
			strconv.Itoa(r.ClassifiedMass),
			strconv.Itoa(r.DebtMass),
			strconv.Itoa(r.ModuleCount),
			r.Concentration.TopModule,
			strconv.FormatFloat(r.Concentration.Share, 'f', 4, 64),
		})
	}
	w.Flush()
}

// pct formats a 0-1 ratio as a percentage string, e.g. 0.382 -> "38.2%".
func pct(r float64) string {
	return fmt.Sprintf("%.1f%%", r*100)
}

// commafy renders an int with thousands separators, e.g. 42180 -> "42,180".
func commafy(n int) string {
	s := strconv.Itoa(n)
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}
