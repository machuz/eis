package output

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/machuz/engineering-impact-score/internal/team"
)

const (
	colWidth = 46 // visible character width per column
	colGap   = 4  // gap between columns
	barWidth = 10 // bar chart width for 2-column mode
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func visibleLen(s string) int {
	return utf8.RuneCountInString(ansiRe.ReplaceAllString(s, ""))
}

// padRight pads a colored string to the given visible width.
func padRight(s string, width int) string {
	vl := visibleLen(s)
	if vl >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vl)
}

// sideBySide prints two columns of lines side by side.
func sideBySide(left, right []string) {
	n := len(left)
	if len(right) > n {
		n = len(right)
	}
	gap := strings.Repeat(" ", colGap)
	for i := 0; i < n; i++ {
		l := ""
		if i < len(left) {
			l = left[i]
		}
		r := ""
		if i < len(right) {
			r = right[i]
		}
		fmt.Printf("%s%s%s\n", padRight(l, colWidth), gap, r)
	}
}

// PrintTeamTable renders team results with bar charts to the terminal.
func PrintTeamTable(teams []team.TeamResult) {
	for _, tr := range teams {
		printTeamHeader(tr)
		printWarnings(tr)
		printClassification2Col(tr)
		printDistributions2Col(tr)
		printHealthAndAverages(tr)
		printStructureAndState(tr)
		fmt.Println()
	}
}

func printWarnings(tr team.TeamResult) {
	if len(tr.Warnings) == 0 {
		return
	}
	warnFmt := color.New(color.FgHiRed, color.Bold)
	dimFmt := color.New(color.FgHiYellow)
	warnFmt.Println("⚠ Warnings:")
	for _, w := range tr.Warnings {
		fmt.Printf("  %s\n", dimFmt.Sprint(w))
	}
	fmt.Println()
}

func printTeamHeader(tr team.TeamResult) {
	fmt.Println()
	riskCount := tr.MemberCount - tr.CoreMemberCount
	if riskCount > 0 || tr.TotalMemberCount > tr.MemberCount {
		color.New(color.FgHiCyan, color.Bold).Printf(
			"═══ %s / %s (%d core + %d risk / %d total, %d repos) ═══\n",
			tr.Name, tr.Domain, tr.CoreMemberCount, riskCount, tr.TotalMemberCount, tr.RepoCount,
		)
	} else {
		color.New(color.FgHiCyan, color.Bold).Printf(
			"═══ %s / %s (%d members, %d repos) ═══\n",
			tr.Name, tr.Domain, tr.MemberCount, tr.RepoCount,
		)
	}

	charStr := formatCharacterLabel(tr.Classification.Character)
	fmt.Printf("  %s\n", charStr)

	if riskCount > 0 || tr.TotalMemberCount > tr.MemberCount {
		peripheralCount := tr.TotalMemberCount - tr.MemberCount
		fmt.Printf("  %s\n",
			color.New(color.FgHiBlack).Sprintf("%d core + %d risk members · %d peripheral · %d repos",
				tr.CoreMemberCount, riskCount, peripheralCount, tr.RepoCount))
	}
	fmt.Println()
}

func printClassification2Col(tr team.TeamResult) {
	c := tr.Classification
	labelFmt := color.New(color.FgHiBlue, color.Bold).SprintfFunc()
	confFmt := color.New(color.FgHiBlack).SprintfFunc()
	dimFmt := color.New(color.FgWhite).SprintfFunc()

	structStr := formatLabel(c.Structure, labelFmt, confFmt)
	cultureStr := formatLabel(c.Culture, labelFmt, confFmt)
	phaseStr := formatLabel(c.Phase, labelFmt, confFmt)
	riskStr := formatRiskLabel(c.Risk)

	color.New(color.FgWhite, color.Bold).Println("Classification:")
	left := fmt.Sprintf("  %s %s", dimFmt("Structure:"), structStr)
	right := fmt.Sprintf("%s %s", dimFmt("Culture:"), cultureStr)
	fmt.Printf("%s%s%s\n", padRight(left, colWidth), strings.Repeat(" ", colGap), right)

	left = fmt.Sprintf("  %s %s", dimFmt("Phase:    "), phaseStr)
	right = fmt.Sprintf("%s %s", dimFmt("Risk:   "), riskStr)
	fmt.Printf("%s%s%s\n", padRight(left, colWidth), strings.Repeat(" ", colGap), right)
	fmt.Println()
}

func printDistributions2Col(tr team.TeamResult) {
	leftLines := buildDistLines("Role Distribution:", tr.RoleDist, tr.MemberCount)
	rightLines := buildDistLines("Style Distribution:", tr.StyleDist, tr.MemberCount)
	sideBySide(leftLines, rightLines)
	fmt.Println()
}

func buildDistLines(title string, dist map[string]int, total int) []string {
	if len(dist) == 0 {
		return nil
	}

	titleStr := color.New(color.FgWhite, color.Bold).Sprint(title)
	lines := []string{titleStr}

	type kv struct {
		key   string
		count int
	}
	var sorted []kv
	for k, v := range dist {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].count > sorted[j].count })

	for _, item := range sorted {
		pct := float64(item.count) / float64(total) * 100
		barLen := int(pct / 100 * float64(barWidth))
		if barLen < 1 && item.count > 0 {
			barLen = 1
		}
		bar := color.New(color.FgHiBlue).Sprint(strings.Repeat("█", barLen)) +
			color.New(color.FgHiBlack).Sprint(strings.Repeat("░", barWidth-barLen))
		lines = append(lines, fmt.Sprintf("  %-12s %s  %d (%.0f%%)", item.key, bar, item.count, pct))
	}
	return lines
}

func printHealthAndAverages(tr team.TeamResult) {
	leftLines := buildHealthLines(tr)
	rightLines := buildAverageLines(tr)
	sideBySide(leftLines, rightLines)
	fmt.Println()
}

func buildHealthLines(tr team.TeamResult) []string {
	titleStr := color.New(color.FgWhite, color.Bold).Sprint("Team Health:")
	lines := []string{titleStr}

	h := tr.Health
	axes := []struct {
		name string
		val  float64
	}{
		{"Complement", h.Complementarity},
		{"Growth", h.GrowthPotential},
		{"Sustain", h.Sustainability},
		{"DebtBalance", h.DebtBalance},
		{"ProdDensity", h.ProductivityDensity},
		{"QualConsist", h.QualityConsistency},
	}

	for _, a := range axes {
		bar := renderBarShort(a.val, barWidth)
		c := healthColor(a.val)
		lines = append(lines, fmt.Sprintf("  %-12s %s %s", a.name, bar, c.Sprintf("%.0f/100", a.val)))
	}

	// Risk ratio
	riskBar := renderBarInverseShort(h.RiskRatio, barWidth)
	riskColor := riskLevelColor(h.RiskRatio)
	lines = append(lines, fmt.Sprintf("  %-12s %s %s", "RiskRatio", riskBar, riskColor.Sprintf("%.0f%%", h.RiskRatio)))

	return lines
}

func buildAverageLines(tr team.TeamResult) []string {
	titleStr := color.New(color.FgWhite, color.Bold).Sprint("Averages:")
	lines := []string{titleStr}

	axes := []struct {
		name string
		val  float64
	}{
		{"Production", tr.AvgProduction},
		{"Quality", tr.AvgQuality},
		{"Survival", tr.AvgSurvival},
		{"Design", tr.AvgDesign},
		{"Breadth", tr.AvgBreadth},
		{"Debt Cleanup", tr.AvgDebtCleanup},
	}

	for _, a := range axes {
		c := scoreColor(a.val)
		lines = append(lines, fmt.Sprintf("  %-14s %s", a.name, c.Sprintf("%.1f", a.val)))
	}

	// Total as highlight
	totalColor := scoreColor(tr.AvgTotal)
	lines = append(lines, fmt.Sprintf("  %-14s %s", "Total", totalColor.Sprintf("%.1f", tr.AvgTotal)))

	return lines
}

func printStructureAndState(tr team.TeamResult) {
	leftLines := buildStructureMetricLines(tr)
	rightLines := buildDistLines("State Distribution:", tr.StateDist, tr.MemberCount)

	if len(leftLines) == 0 && len(rightLines) == 0 {
		return
	}
	sideBySide(leftLines, rightLines)
}

func buildStructureMetricLines(tr team.TeamResult) []string {
	h := tr.Health
	if h.AAR == 0 && h.AnchorDensity == 0 && h.ArchitectureCoverage == 0 {
		return nil
	}

	titleStr := color.New(color.FgWhite, color.Bold).Sprint("Structure Metrics:")
	lines := []string{titleStr}

	// AAR
	architects := tr.RoleDist["Architect"]
	anchors := tr.RoleDist["Anchor"]
	aarLabel := formatAAR(h.AAR)
	aarColor := aarScoreColor(h.AAR)
	if h.AAR > 0 {
		// Normal ratio — show counts
		lines = append(lines, fmt.Sprintf("  %-12s %s  (%d Arch / %d Anch)",
			"AAR", aarColor.Sprint(aarLabel), architects, anchors))
	} else {
		// N/A or ∞ — label is self-explanatory
		lines = append(lines, fmt.Sprintf("  %-12s %s",
			"AAR", aarColor.Sprint(aarLabel)))
	}

	// Anchor Density
	densityPct := h.AnchorDensity * 100
	bar := renderBarShort(densityPct, barWidth)
	fmt2 := color.New(color.FgWhite)
	lines = append(lines, fmt.Sprintf("  %-12s %s %s", "AnchorDens", bar, fmt2.Sprintf("%.0f%%", densityPct)))

	// Architecture Coverage
	covPct := h.ArchitectureCoverage * 100
	bar2 := renderBarShort(covPct, barWidth)
	lines = append(lines, fmt.Sprintf("  %-12s %s %s", "ArchCover", bar2, fmt2.Sprintf("%.0f%%", covPct)))

	return lines
}

func formatCharacterLabel(l team.TeamLabel) string {
	if l.Name == "" || l.Name == "—" {
		return "—"
	}
	nameFmt := color.New(color.FgHiYellow, color.Bold)
	confFmt := color.New(color.FgHiBlack)
	return fmt.Sprintf("★ %s %s", nameFmt.Sprint(l.Name), confFmt.Sprintf("(%.2f)", l.Confidence))
}

func formatLabel(l team.TeamLabel, labelFmt, confFmt func(string, ...interface{}) string) string {
	if l.Name == "" || l.Name == "—" {
		return "—"
	}
	return fmt.Sprintf("%s %s", labelFmt("%s", l.Name), confFmt("(%.2f)", l.Confidence))
}

func formatRiskLabel(l team.TeamLabel) string {
	if l.Name == "" || l.Name == "—" {
		return "—"
	}
	var c *color.Color
	switch l.Name {
	case "Healthy":
		c = color.New(color.FgHiGreen, color.Bold)
	default:
		c = color.New(color.FgHiRed, color.Bold)
	}
	confFmt := color.New(color.FgHiBlack).SprintfFunc()
	return fmt.Sprintf("%s %s", c.Sprint(l.Name), confFmt("(%.2f)", l.Confidence))
}

func formatAAR(aar float64) string {
	if aar == 0 {
		return "N/A (no Architect)"
	}
	if aar < 0 {
		return "∞ (no Anchor)"
	}
	return fmt.Sprintf("%.2f", aar)
}

func aarScoreColor(aar float64) *color.Color {
	switch {
	case aar < 0:
		return color.New(color.FgHiRed, color.Bold)
	case aar >= 0.3 && aar <= 0.8:
		return color.New(color.FgHiGreen, color.Bold)
	case aar > 0 && aar < 0.3:
		return color.New(color.FgHiYellow)
	case aar > 0.8 && aar <= 2.0:
		return color.New(color.FgHiYellow)
	case aar > 2.0:
		return color.New(color.FgHiRed)
	default:
		return color.New(color.FgWhite)
	}
}

func renderBarShort(val float64, width int) string {
	filled := int(val / 100 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return color.New(color.FgHiGreen).Sprint(strings.Repeat("█", filled)) +
		color.New(color.FgHiBlack).Sprint(strings.Repeat("░", width-filled))
}

func renderBarInverseShort(val float64, width int) string {
	filled := int(val / 100 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return color.New(color.FgHiRed).Sprint(strings.Repeat("█", filled)) +
		color.New(color.FgHiBlack).Sprint(strings.Repeat("░", width-filled))
}

func scoreColor(v float64) *color.Color {
	switch {
	case v >= 80:
		return color.New(color.FgHiMagenta, color.Bold)
	case v >= 60:
		return color.New(color.FgHiGreen, color.Bold)
	case v >= 40:
		return color.New(color.FgHiYellow)
	default:
		return color.New(color.FgWhite)
	}
}

func healthColor(v float64) *color.Color {
	switch {
	case v >= 70:
		return color.New(color.FgHiGreen, color.Bold)
	case v >= 50:
		return color.New(color.FgHiYellow)
	case v >= 30:
		return color.New(color.FgYellow)
	default:
		return color.New(color.FgHiRed)
	}
}

func riskLevelColor(v float64) *color.Color {
	switch {
	case v >= 50:
		return color.New(color.FgHiRed, color.Bold)
	case v >= 25:
		return color.New(color.FgYellow)
	default:
		return color.New(color.FgHiGreen)
	}
}
