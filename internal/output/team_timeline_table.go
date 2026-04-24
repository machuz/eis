package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/machuz/eis/v2/internal/timeline"
)

// PrintTeamTimelineTable renders team timeline data as a colored terminal table.
func PrintTeamTimelineTable(tl timeline.TeamTimeline) {
	fmt.Println()
	color.New(color.FgHiCyan, color.Bold).Printf("═══ %s / %s — Team Timeline ═══\n\n", tl.TeamName, tl.Domain)

	// Membership
	color.New(color.FgWhite, color.Bold).Println("Membership:")
	fmt.Printf("  %-16s", "Period")
	for _, p := range tl.Periods {
		fmt.Printf("  %-14s", p.Label)
	}
	fmt.Println()
	printTeamTimelineRow("Core", tl.Periods, func(p timeline.TeamPeriodSnapshot) string {
		return fmt.Sprintf("%d", p.CoreMembers)
	})
	printTeamTimelineRow("Effective", tl.Periods, func(p timeline.TeamPeriodSnapshot) string {
		return fmt.Sprintf("%d", p.EffectiveMembers)
	})
	printTeamTimelineRow("Total", tl.Periods, func(p timeline.TeamPeriodSnapshot) string {
		return fmt.Sprintf("%d", p.TotalMembers)
	})
	fmt.Println()

	// Classification
	color.New(color.FgWhite, color.Bold).Println("Classification:")
	fmt.Printf("  %-16s", "Period")
	for _, p := range tl.Periods {
		fmt.Printf("  %-14s", p.Label)
	}
	fmt.Println()
	printTeamTimelineClassRow("Character", tl.Periods, func(p timeline.TeamPeriodSnapshot) string { return p.Character })
	printTeamTimelineClassRow("Structure", tl.Periods, func(p timeline.TeamPeriodSnapshot) string { return p.Structure })
	printTeamTimelineClassRow("Culture", tl.Periods, func(p timeline.TeamPeriodSnapshot) string { return p.Culture })
	printTeamTimelineClassRow("Phase", tl.Periods, func(p timeline.TeamPeriodSnapshot) string { return p.Phase })
	printTeamTimelineClassRow("Risk", tl.Periods, func(p timeline.TeamPeriodSnapshot) string { return p.Risk })
	fmt.Println()

	// Health
	color.New(color.FgWhite, color.Bold).Println("Health:")
	fmt.Printf("  %-16s", "Period")
	for _, p := range tl.Periods {
		fmt.Printf("  %-14s", p.Label)
	}
	fmt.Println()
	printTeamTimelineFloatRow("Complement", tl.Periods, func(p timeline.TeamPeriodSnapshot) float64 { return p.Complementarity })
	printTeamTimelineFloatRow("Growth", tl.Periods, func(p timeline.TeamPeriodSnapshot) float64 { return p.GrowthPotential })
	printTeamTimelineFloatRow("Sustain", tl.Periods, func(p timeline.TeamPeriodSnapshot) float64 { return p.Sustainability })
	printTeamTimelineFloatRow("DebtBalance", tl.Periods, func(p timeline.TeamPeriodSnapshot) float64 { return p.DebtBalance })
	printTeamTimelineFloatRow("ProdDensity", tl.Periods, func(p timeline.TeamPeriodSnapshot) float64 { return p.ProductivityDensity })
	printTeamTimelineFloatRow("QualConsist", tl.Periods, func(p timeline.TeamPeriodSnapshot) float64 { return p.QualityConsistency })
	printTeamTimelineFloatRow("RiskRatio", tl.Periods, func(p timeline.TeamPeriodSnapshot) float64 { return p.RiskRatio })
	fmt.Println()

	// Averages
	color.New(color.FgWhite, color.Bold).Println("Score Averages:")
	fmt.Printf("  %-16s", "Period")
	for _, p := range tl.Periods {
		fmt.Printf("  %-14s", p.Label)
	}
	fmt.Println()
	printTeamTimelineFloatRow("Production", tl.Periods, func(p timeline.TeamPeriodSnapshot) float64 { return p.AvgProduction })
	printTeamTimelineFloatRow("Quality", tl.Periods, func(p timeline.TeamPeriodSnapshot) float64 { return p.AvgQuality })
	printTeamTimelineFloatRow("Survival", tl.Periods, func(p timeline.TeamPeriodSnapshot) float64 { return p.AvgSurvival })
	printTeamTimelineFloatRow("Design", tl.Periods, func(p timeline.TeamPeriodSnapshot) float64 { return p.AvgDesign })
	printTeamTimelineFloatRow("DebtCleanup", tl.Periods, func(p timeline.TeamPeriodSnapshot) float64 { return p.AvgDebtCleanup })
	printTeamTimelineFloatRow("Impact", tl.Periods, func(p timeline.TeamPeriodSnapshot) float64 { return p.AvgImpact })
	fmt.Println()

	// Transitions
	if len(tl.Transitions) > 0 {
		color.New(color.FgWhite, color.Bold).Println("Transitions:")
		for _, t := range tl.Transitions {
			arrow := color.New(color.FgHiYellow).Sprintf("→")
			fmt.Printf("  [%s] %s: %s %s %s\n", t.AtPeriod, t.Axis, t.From, arrow, t.To)
		}
		fmt.Println()
	}
}

func printTeamTimelineRow(label string, periods []timeline.TeamPeriodSnapshot, fn func(timeline.TeamPeriodSnapshot) string) {
	fmt.Printf("  %-16s", label)
	for _, p := range periods {
		fmt.Printf("  %-14s", fn(p))
	}
	fmt.Println()
}

func printTeamTimelineClassRow(label string, periods []timeline.TeamPeriodSnapshot, fn func(timeline.TeamPeriodSnapshot) string) {
	fmt.Printf("  %-16s", label)
	prev := ""
	for _, p := range periods {
		val := fn(p)
		if val == "" {
			val = "—"
		}
		// Highlight changes
		if prev != "" && prev != "—" && val != "—" && val != prev {
			fmt.Printf("  %s", padRight(color.New(color.FgHiYellow, color.Bold).Sprintf("%-14s", val), 16))
		} else {
			fmt.Printf("  %-14s", val)
		}
		prev = val
	}
	fmt.Println()
}

func printTeamTimelineFloatRow(label string, periods []timeline.TeamPeriodSnapshot, fn func(timeline.TeamPeriodSnapshot) float64) {
	fmt.Printf("  %-16s", label)
	prev := -1.0
	for _, p := range periods {
		val := fn(p)
		valStr := fmt.Sprintf("%.1f", val)

		// Show trend indicator
		if prev >= 0 && val > 0 {
			diff := val - prev
			if diff > 5 {
				valStr = color.New(color.FgHiGreen).Sprintf("%.1f ↑", val)
			} else if diff < -5 {
				valStr = color.New(color.FgHiRed).Sprintf("%.1f ↓", val)
			}
		}

		formatted := padRight(valStr, 14)
		fmt.Printf("  %s", formatted)
		prev = val
	}
	fmt.Println()
}

// PrintTeamTimelineSeparator prints a divider between team timelines.
func PrintTeamTimelineSeparator() {
	fmt.Println(strings.Repeat("─", 80))
}
