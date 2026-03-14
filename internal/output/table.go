package output

import (
	"fmt"
	"os"
	"regexp"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/machuz/engineering-impact-score/internal/metric"
	"github.com/machuz/engineering-impact-score/internal/scorer"
	"github.com/rodaine/table"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripAnsiWidth(s string) int {
	return utf8.RuneCountInString(ansiRegexp.ReplaceAllString(s, ""))
}

func PrintRankings(results []scorer.Result) {
	headerFmt := color.New(color.FgCyan, color.Bold).SprintfFunc()
	columnFmt := color.New(color.FgWhite).SprintfFunc()

	// Detect if pressure data is available
	hasPressure := false
	for _, r := range results {
		if r.RobustSurvival > 0 || r.DormantSurvival > 0 {
			hasPressure = true
			break
		}
	}

	var tbl table.Table
	if hasPressure {
		tbl = table.New("#", "Member", "Active", "Prod", "Qual", "Robust", "Dormant", "Design", "Breadth", "Debt", "Indisp", "Grav", "Total", "Role", "Style", "State")
	} else {
		tbl = table.New("#", "Member", "Active", "Prod", "Qual", "Surv", "Design", "Breadth", "Debt", "Indisp", "Grav", "Total", "Role", "Style", "State")
	}
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithWidthFunc(stripAnsiWidth).WithWriter(os.Stdout)

	nameFmt := color.New(color.FgHiYellow).SprintfFunc()
	labelFmt := color.New(color.FgHiBlue).SprintfFunc()
	activeFmt := color.New(color.FgHiGreen).SprintfFunc()
	inactiveFmt := color.New(color.FgHiBlack).SprintfFunc()
	confFmt := color.New(color.FgHiBlack).SprintfFunc()

	for i, r := range results {
		totalStr := formatTotal(r.Total)

		roleStr := formatAxis(r.Role, r.RoleConf, labelFmt, confFmt)
		styleStr := formatAxis(r.Style, r.StyleConf, labelFmt, confFmt)
		stateStr := formatAxis(r.State, r.StateConf, labelFmt, confFmt)

		activeStr := inactiveFmt("—")
		if r.RecentlyActive {
			activeStr = activeFmt("✓")
		}
		gravStr := formatGravity(r)

		if hasPressure {
			tbl.AddRow(
				i+1,
				nameFmt("%s", r.Author),
				activeStr,
				fmt.Sprintf("%.0f", r.Production),
				fmt.Sprintf("%.0f", r.Quality),
				fmt.Sprintf("%.0f", r.RobustSurvival),
				fmt.Sprintf("%.0f", r.DormantSurvival),
				fmt.Sprintf("%.0f", r.Design),
				fmt.Sprintf("%.0f", r.Breadth),
				fmt.Sprintf("%.0f", r.DebtCleanup),
				fmt.Sprintf("%.0f", r.Indispensability),
				gravStr,
				totalStr,
				roleStr,
				styleStr,
				stateStr,
			)
		} else {
			tbl.AddRow(
				i+1,
				nameFmt("%s", r.Author),
				activeStr,
				fmt.Sprintf("%.0f", r.Production),
				fmt.Sprintf("%.0f", r.Quality),
				fmt.Sprintf("%.0f", r.Survival),
				fmt.Sprintf("%.0f", r.Design),
				fmt.Sprintf("%.0f", r.Breadth),
				fmt.Sprintf("%.0f", r.DebtCleanup),
				fmt.Sprintf("%.0f", r.Indispensability),
				gravStr,
				totalStr,
				roleStr,
				styleStr,
				stateStr,
			)
		}
	}

	fmt.Println()
	tbl.Print()
	fmt.Println()
}

func PrintBusFactorRisks(risks []metric.ModuleRisk) {
	if len(risks) == 0 {
		return
	}

	headerFmt := color.New(color.FgRed, color.Bold).SprintfFunc()
	columnFmt := color.New(color.FgWhite).SprintfFunc()

	fmt.Println()
	color.New(color.FgRed, color.Bold).Println("Bus Factor Risks:")

	tbl := table.New("Level", "Module", "Owner", "Share")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithWriter(os.Stdout)

	for _, r := range risks {
		tbl.AddRow(r.Level, r.Module, r.TopAuthor, fmt.Sprintf("%.0f%%", r.Share*100))
	}

	tbl.Print()
	fmt.Println()
}

func formatAxis(name string, conf float64, labelFmt, confFmt func(string, ...interface{}) string) string {
	if name == "" || name == "—" {
		return "—"
	}
	return fmt.Sprintf("%s %s", labelFmt("%s", name), confFmt("(%.2f)", conf))
}

// formatGravity colors gravity by its health quality.
// High gravity + high quality/survival = green (healthy structural influence).
// High gravity + low quality/survival = red (fragile structural dependency).
// Low gravity = dim (not enough influence to matter).
func formatGravity(r scorer.Result) string {
	g := r.Gravity
	if g < 20 {
		return color.New(color.FgHiBlack).Sprintf("%.0f", g)
	}

	// Gravity health: how durable is this structural influence?
	// Quality and RobustSurvival indicate whether the gravity is sustainable.
	health := r.Quality*0.6 + r.RobustSurvival*0.4
	if r.RobustSurvival == 0 && r.DormantSurvival == 0 {
		// No pressure data available — fall back to Quality + Survival
		health = r.Quality*0.6 + r.Survival*0.4
	}

	switch {
	case health >= 60:
		// Healthy gravity: durable structural influence
		return color.New(color.FgHiGreen, color.Bold).Sprintf("%.0f", g)
	case health >= 40:
		// Moderate gravity quality
		return color.New(color.FgHiYellow).Sprintf("%.0f", g)
	default:
		// Fragile gravity: high influence but poor durability
		return color.New(color.FgHiRed, color.Bold).Sprintf("%.0f", g)
	}
}

func formatTotal(total float64) string {
	switch {
	case total >= 80:
		return color.New(color.FgHiMagenta, color.Bold).Sprintf("%.1f", total)
	case total >= 60:
		return color.New(color.FgHiGreen, color.Bold).Sprintf("%.1f", total)
	case total >= 40:
		return color.New(color.FgHiYellow).Sprintf("%.1f", total)
	case total >= 20:
		return color.New(color.FgWhite).Sprintf("%.1f", total)
	default:
		return color.New(color.FgHiBlack).Sprintf("%.1f", total)
	}
}

// PerRepoData holds per-repo scored results for cross-repo comparison output.
type PerRepoData struct {
	RepoName string
	Results  []scorer.Result
}

// PrintPerRepoComparison prints a cross-repo comparison table showing each author's
// Role, Style, State and Total score per repository, with a Pattern column.
func PrintPerRepoComparison(domainName string, perRepo []PerRepoData, aggregated []scorer.Result) {
	if len(perRepo) == 0 {
		return
	}

	fmt.Println()
	color.New(color.FgHiCyan, color.Bold).Printf("─── %s Per-Repository Breakdown ───\n\n", domainName)

	// Collect all authors from aggregated results (sorted by total desc)
	authors := make([]string, 0, len(aggregated))
	for _, r := range aggregated {
		authors = append(authors, r.Author)
	}

	// Build repo name list
	repoNames := make([]string, 0, len(perRepo))
	for _, rr := range perRepo {
		repoNames = append(repoNames, rr.RepoName)
	}

	// Build lookup: author -> repoName -> Result
	lookup := make(map[string]map[string]*scorer.Result)
	for _, rr := range perRepo {
		for i := range rr.Results {
			r := &rr.Results[i]
			if lookup[r.Author] == nil {
				lookup[r.Author] = make(map[string]*scorer.Result)
			}
			lookup[r.Author][rr.RepoName] = r
		}
	}

	// Build table header
	headerFmt := color.New(color.FgCyan, color.Bold).SprintfFunc()
	columnFmt := color.New(color.FgWhite).SprintfFunc()

	headers := []interface{}{"Author"}
	for _, rn := range repoNames {
		headers = append(headers, rn)
	}
	headers = append(headers, "Pattern")

	tbl := table.New(headers...)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithWidthFunc(stripAnsiWidth).WithWriter(os.Stdout)

	nameFmt := color.New(color.FgHiYellow).SprintfFunc()
	dimFmt := color.New(color.FgHiBlack).SprintfFunc()
	patternFmt := color.New(color.FgHiGreen, color.Bold).SprintfFunc()

	for _, author := range authors {
		row := []interface{}{nameFmt("%s", author)}
		roles := make([]string, 0)
		for _, rn := range repoNames {
			r, ok := lookup[author][rn]
			if !ok {
				row = append(row, dimFmt("—"))
			} else {
				var cell string
				if r.Role != "" && r.Role != "—" {
					cell = fmt.Sprintf("%s %.0f", r.Role, r.Total)
				} else {
					cell = dimFmt("%.0f", r.Total)
				}
				row = append(row, cell)
				roles = append(roles, r.Role)
			}
		}
		pattern := derivePattern(roles)
		row = append(row, patternFmt("%s", pattern))
		tbl.AddRow(row...)
	}

	tbl.Print()
	fmt.Println()
}

// derivePattern determines an author's cross-repo pattern from their roles.
func derivePattern(roles []string) string {
	if len(roles) == 0 {
		return "—"
	}
	if len(roles) == 1 {
		return "Single Repo"
	}

	// Check if all roles are the same
	allSame := true
	first := roles[0]
	for _, r := range roles[1:] {
		if r != first {
			allSame = false
			break
		}
	}
	if allSame {
		if first == "Architect" {
			return "Reproducible"
		}
		if first == "" || first == "—" {
			return "Emerging"
		}
		return "Consistently " + first
	}

	// Check if any repo has Architect
	hasArchitect := false
	for _, r := range roles {
		if r == "Architect" {
			hasArchitect = true
			break
		}
	}
	if hasArchitect {
		return "Context-dependent"
	}

	return "Variable"
}

func PrintSummary(results []scorer.Result, repoCount int) {
	fmt.Printf("Analyzed %d repo(s), %d engineers\n", repoCount, len(results))
	fmt.Println()

	legend := []struct {
		min   float64
		max   float64
		label string
	}{
		{80, 100, "Irreplaceable core member"},
		{60, 79, "Near-core. Strong"},
		{40, 59, "Senior-level (40+ is genuinely strong)"},
		{30, 39, "Mid-level"},
		{20, 29, "Junior-Mid"},
		{0, 19, "Junior"},
	}

	for _, l := range legend {
		count := 0
		for _, r := range results {
			if r.Total >= l.min && r.Total <= l.max {
				count++
			}
		}
		if count > 0 {
			fmt.Printf("  %3.0f-%3.0f  %s: %d\n", l.min, l.max, l.label, count)
		}
	}
	fmt.Println()
}
