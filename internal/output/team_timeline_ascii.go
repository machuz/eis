package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/guptarohit/asciigraph"
	"github.com/machuz/eis/v2/internal/timeline"
)

// PrintTeamTimelineASCII renders team timeline data as ASCII line charts.
func PrintTeamTimelineASCII(tl timeline.TeamTimeline) {
	fmt.Println()
	color.New(color.FgHiCyan, color.Bold).Printf("=== %s / %s -- Team Timeline ===\n", tl.TeamName, tl.Domain)

	// Build score average series
	avgImpactData := make([]float64, 0, len(tl.Periods))
	avgProdData := make([]float64, 0, len(tl.Periods))
	avgQualData := make([]float64, 0, len(tl.Periods))
	avgSurvData := make([]float64, 0, len(tl.Periods))
	avgDesignData := make([]float64, 0, len(tl.Periods))
	var labels []string

	for _, p := range tl.Periods {
		avgImpactData = append(avgImpactData, p.AvgImpact)
		avgProdData = append(avgProdData, p.AvgProduction)
		avgQualData = append(avgQualData, p.AvgQuality)
		avgSurvData = append(avgSurvData, p.AvgSurvival)
		avgDesignData = append(avgDesignData, p.AvgDesign)
		labels = append(labels, p.Label)
	}

	// Score averages chart
	color.New(color.FgWhite, color.Bold).Println("\nScore Averages:")

	data := [][]float64{avgImpactData, avgProdData, avgQualData, avgSurvData, avgDesignData}

	// Calculate chart width to match label axis spacing
	chartWidth := 60
	if len(labels) > 1 {
		spacing := chartWidth / (len(labels) - 1)
		if spacing < len(labels[0])+2 {
			spacing = len(labels[0]) + 2
			chartWidth = spacing * (len(labels) - 1)
		}
	}

	chart := asciigraph.PlotMany(data,
		asciigraph.Height(15),
		asciigraph.Width(chartWidth),
		asciigraph.SeriesColors(
			asciigraph.Blue,
			asciigraph.Green,
			asciigraph.Yellow,
			asciigraph.Red,
			asciigraph.Cyan,
		),
		asciigraph.Caption(buildLabelAxis(labels)),
	)
	fmt.Println(chart)

	// Legend
	fmt.Printf("  %s  %s  %s  %s  %s\n",
		color.BlueString("AvgImpact"),
		color.GreenString("AvgProduction"),
		color.YellowString("AvgQuality"),
		color.RedString("AvgSurvival"),
		color.CyanString("AvgDesign"),
	)

	// Classification summary
	fmt.Println()
	color.New(color.FgWhite, color.Bold).Println("Classification:")
	classAxes := []struct {
		name string
		get  func(timeline.TeamPeriodSnapshot) string
	}{
		{"Character", func(p timeline.TeamPeriodSnapshot) string { return p.Character }},
		{"Structure", func(p timeline.TeamPeriodSnapshot) string { return p.Structure }},
		{"Culture", func(p timeline.TeamPeriodSnapshot) string { return p.Culture }},
		{"Phase", func(p timeline.TeamPeriodSnapshot) string { return p.Phase }},
		{"Risk", func(p timeline.TeamPeriodSnapshot) string { return p.Risk }},
	}
	for _, ax := range classAxes {
		var vals []string
		for _, p := range tl.Periods {
			v := ax.get(p)
			if v == "" {
				v = "-"
			}
			vals = append(vals, v)
		}
		fmt.Printf("  %-12s %s\n", ax.name+":", strings.Join(vals, " -> "))
	}

	// Transitions
	if len(tl.Transitions) > 0 {
		fmt.Println()
		color.New(color.FgWhite, color.Bold).Println("Transitions:")
		for _, t := range tl.Transitions {
			fmt.Printf("  [%s] %s: %s %s %s\n",
				t.AtPeriod, t.Axis, t.From,
				color.HiYellowString("->"),
				t.To,
			)
		}
	}

	fmt.Println()
}
