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
		tbl = table.New("#", "Member", "Active", "Prod", "Qual", "Robust", "Dormant", "Design", "Breadth", "Debt", "Indisp", "Total", "Type", "2nd")
	} else {
		tbl = table.New("#", "Member", "Active", "Prod", "Qual", "Surv", "Design", "Breadth", "Debt", "Indisp", "Total", "Type", "2nd")
	}
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithWidthFunc(stripAnsiWidth).WithWriter(os.Stdout)

	nameFmt := color.New(color.FgHiYellow).SprintfFunc()
	typeFmt := color.New(color.FgHiBlue).SprintfFunc()
	activeFmt := color.New(color.FgHiGreen).SprintfFunc()
	inactiveFmt := color.New(color.FgHiBlack).SprintfFunc()
	confFmt := color.New(color.FgHiBlack).SprintfFunc()

	for i, r := range results {
		totalStr := formatTotal(r.Total)
		typeStr := r.Archetype
		if typeStr != "" && typeStr != "—" {
			typeStr = fmt.Sprintf("%s %s", typeFmt("%s", r.Archetype), confFmt("(%.2f)", r.ArchetypeConf))
		}
		activeStr := inactiveFmt("—")
		if r.RecentlyActive {
			activeStr = activeFmt("✓")
		}
		secondaryStr := ""
		if r.Secondary.Name != "" && r.Secondary.Confidence > 0 {
			secondaryStr = confFmt("%s (%.2f)", r.Secondary.Name, r.Secondary.Confidence)
		}
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
				totalStr,
				typeStr,
				secondaryStr,
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
				totalStr,
				typeStr,
				secondaryStr,
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
